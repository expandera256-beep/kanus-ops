package licensesrv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func loginFromIP(h *AccountHandler, ip, email, pw string) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(map[string]string{"email": email, "password": pw})
	r := httptest.NewRequest("POST", "/account/login", strings.NewReader(string(raw)))
	r.RemoteAddr = "127.0.0.1:5555" // behind nginx: X-Real-IP is trusted from loopback
	r.Header.Set("X-Real-IP", ip)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// Policy update: distributed failures use a high global threshold, while the
// new TestLoginPolicy suite covers pair/backoff/trusted-session semantics.
func TestAccountLimiterRotatingIPs(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	h, _ := setup(t)
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "victim@example.test", "")
	for _, email := range []string{"victim@example.test", "ghost@example.test"} {
		for i := 1; i <= 100; i++ {
			w := loginFromIP(a, fmt.Sprintf("203.%d.%d.%d", len(email), i/250, i%250+1), "  "+strings.ToUpper(email)+" ", "wrong-password-xyz")
			if w.Code != 401 {
				t.Fatalf("failure %d HTTP%d", i, w.Code)
			}
		}
		w := loginFromIP(a, "198.18.0.1", email, accountPassword)
		if w.Code != 429 || w.Header().Get("Retry-After") == "" || w.Header().Get(hdrQueueDepth) == "" {
			t.Fatalf("global gate HTTP%d", w.Code)
		}
	}
	if strings.Contains(buf.String(), "victim@example.test") || strings.Contains(buf.String(), "ghost@example.test") {
		t.Fatal("unmasked email in logs")
	}
}
func TestAccountLimiterSuccessResetsAndWindow(t *testing.T) {
	t.Setenv("US_LOGIN_PAIR_THRESHOLD", "3")
	h, _ := setup(t)
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "u@example.test", "")
	for i := 0; i < 2; i++ {
		loginFromIP(a, "203.0.113.1", "u@example.test", "wrong-password-xyz")
	}
	if w := loginFromIP(a, "203.0.113.1", "u@example.test", accountPassword); w.Code != 200 {
		t.Fatal(w.Code)
	}
	for i := 0; i < 2; i++ {
		if w := loginFromIP(a, "203.0.113.1", "u@example.test", "wrong-password-xyz"); w.Code != 401 {
			t.Fatal("pair reset", w.Code)
		}
	}
	l := a.acctLimit
	now := time.Now()
	key := loginPairKey("u@example.test", "203.0.113.1", false)
	l.fail(key, now)
	if b, _ := l.blocked(key, now); !b {
		t.Fatal("threshold")
	}
	if b, _ := l.blocked(key, now.Add(l.window+time.Second)); b {
		t.Fatal("quiet period expiry")
	}
}

func TestAccountLimiterMemoryBound(t *testing.T) {
	l := newAccountLimiter()
	l.maxKeys = 100
	now := time.Now()
	for i := 0; i < 1000; i++ {
		l.fail(fmt.Sprintf("u%d@example.test", i), now.Add(time.Duration(i)*time.Millisecond))
	}
	if n := l.size(); n > 100 {
		t.Fatalf("anahtar sınırı aşıldı: %d", n)
	}
}
