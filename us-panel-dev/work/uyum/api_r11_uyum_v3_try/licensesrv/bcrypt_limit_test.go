package licensesrv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var loginIP atomic.Int32

func loginReq(h *AccountHandler, ctx context.Context, email, pw string) (*httptest.ResponseRecorder, time.Duration) {
	raw, _ := json.Marshal(map[string]string{"email": email, "password": pw})
	r := httptest.NewRequest("POST", "/account/login", strings.NewReader(string(raw))).WithContext(ctx)
	n := loginIP.Add(1)
	r.RemoteAddr = fmt.Sprintf("198.51.%d.%d:1234", (n/250)%250, n%250+1)
	w := httptest.NewRecorder()
	start := time.Now()
	h.ServeHTTP(w, r)
	return w, time.Since(start)
}

// Production cost-12 account with a known password.
func cost12Handler(t *testing.T) (*AccountHandler, string) {
	t.Helper()
	h, _ := setup(t)
	a := NewAccountHandler(h.Store, testSessionSecret, time.Hour)
	accountRegister(t, a, "known@example.test", "")
	hash, err := bcrypt.GenerateFromPassword([]byte(accountPassword), 12)
	if err != nil {
		t.Fatal(err)
	}
	h.Store.mu.Lock()
	h.Store.st.Accounts["known@example.test"].PasswordHash = string(hash)
	h.Store.mu.Unlock()
	return a, "known@example.test"
}

func fillBcryptSlots(a *AccountHandler, n int) func() {
	for i := 0; i < n; i++ {
		a.work <- struct{}{}
	}
	return func() {
		for i := 0; i < n; i++ {
			<-a.work
		}
	}
}

func TestBcryptThirdLoginWaitsThen503(t *testing.T) {
	t.Setenv("US_BCRYPT_CONCURRENCY", "2")
	t.Setenv("US_BCRYPT_WAIT_SECONDS", "1")
	a, email := cost12Handler(t)
	release := fillBcryptSlots(a, 2) // two logins in progress
	defer release()
	w, d := loginReq(a, context.Background(), email, accountPassword)
	if w.Code != 503 || w.Header().Get("Retry-After") != "5" || !strings.Contains(w.Body.String(), `"status":"busy"`) {
		t.Fatalf("üçüncü giriş: %d %q %s", w.Code, w.Header().Get("Retry-After"), w.Body.String())
	}
	if d < 900*time.Millisecond || d > 3*time.Second {
		t.Fatalf("bekleme süresi %v (≈1s olmalı)", d)
	}
	// A slot freeing up within the wait lets the third login proceed.
	go func() { time.Sleep(200 * time.Millisecond); <-a.work }()
	w, _ = loginReq(a, context.Background(), email, accountPassword)
	a.work <- struct{}{}
	if w.Code != 200 {
		t.Fatalf("slot boşalınca giriş: %d", w.Code)
	}
}

func TestBcryptCancelledContextDoesNoWork(t *testing.T) {
	t.Setenv("US_BCRYPT_WAIT_SECONDS", "10")
	a, email := cost12Handler(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Free slots, cancelled client: must return without a cost-12 compare (~200ms+).
	for i := 0; i < 5; i++ {
		if _, d := loginReq(a, ctx, email, accountPassword); d > 60*time.Millisecond {
			t.Fatalf("iptal edilmiş istekte bcrypt yapıldı (%v)", d)
		}
	}
	// Busy slots, client disconnects while waiting: gives up before the wait limit.
	release := fillBcryptSlots(a, cap(a.work))
	defer release()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel2()
	w, d := loginReq(a, ctx2, email, accountPassword)
	if d > 2*time.Second || w.Code == 200 {
		t.Fatalf("beklerken iptal: %d %v", w.Code, d)
	}
}

func median(ds []time.Duration) time.Duration {
	sort.Slice(ds, func(i, j int) bool { return ds[i] < ds[j] })
	return ds[len(ds)/2]
}

func TestBcryptUnknownUserTimingAndBodyMatchKnownWrongPassword(t *testing.T) {
	t.Setenv("US_BCRYPT_CONCURRENCY", "2")
	a, email := cost12Handler(t)
	var unknown, wrong []time.Duration
	var bodyU, bodyW string
	for i := 0; i < 7; i++ {
		w1, d1 := loginReq(a, context.Background(), "nobody@example.test", "wrong-password-xyz")
		w2, d2 := loginReq(a, context.Background(), email, "wrong-password-xyz")
		if w1.Code != 401 || w2.Code != 401 {
			t.Fatalf("kodlar %d %d", w1.Code, w2.Code)
		}
		unknown, wrong = append(unknown, d1), append(wrong, d2)
		bodyU, bodyW = w1.Body.String(), w2.Body.String()
	}
	if bodyU != bodyW {
		t.Fatalf("hata gövdeleri farklı: %q vs %q", bodyU, bodyW)
	}
	mu, mw := median(unknown), median(wrong)
	// Both must be a real cost-12 compare: same range, ratio within [0.5, 2].
	if mu*2 < mw || mw*2 < mu {
		t.Fatalf("medyanlar farklı aralıkta: bilinmeyen %v, yanlış şifre %v", mu, mw)
	}
	if hashCost(a.dummy) != 12 {
		t.Fatalf("dummy cost %d", hashCost(a.dummy))
	}
}

func hashCost(h []byte) int { c, _ := bcrypt.Cost(h); return c }

func TestBcryptCost12HashesStillVerifyNoRehash(t *testing.T) {
	a, email := cost12Handler(t)
	a.Store.mu.Lock()
	before := a.Store.st.Accounts[email].PasswordHash
	a.Store.mu.Unlock()
	if w, _ := loginReq(a, context.Background(), email, accountPassword); w.Code != 200 {
		t.Fatalf("cost-12 giriş: %d %s", w.Code, w.Body.String())
	}
	a.Store.mu.Lock()
	after := a.Store.st.Accounts[email].PasswordHash
	a.Store.mu.Unlock()
	if before != after || hashCost([]byte(after)) != 12 || a.cost != 12 {
		t.Fatal("girişte yeniden hash veya cost değişti")
	}
}
