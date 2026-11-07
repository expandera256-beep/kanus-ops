package licensesrv

// POST12-A: retry idempotency (X-US-Request-ID). Demir onayı 01:11 TSİ
// (KAN.US üzerinden). Bu testler düzeltmeden önce KIRMIZI, sonra YEŞİL olmalı.
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const idemID = "0123456789abcdef0123456789abcdef"

func idemChat(p *ModelProxy, key, id, input string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "/v1/chat", strings.NewReader(input)).WithContext(context.Background())
	r.Header.Set("Authorization", "Bearer "+key)
	if id != "" {
		r.Header.Set("X-US-Request-ID", id)
	}
	w := httptest.NewRecorder()
	p.ServeHTTP(w, r)
	return w
}

func idemReserve(api *CreditHandler, key, header, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "/v1/credits/reserve", strings.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+key)
	if header != "" {
		r.Header.Set("X-US-Request-ID", header)
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, r)
	return w
}

func idemCode(w *httptest.ResponseRecorder) string {
	var v struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &v)
	return v.Error
}

// Repro 1: aynı mantıksal chat isteği iki kez gelirse yalnız bir kez ücretlenir.
func TestPost12AChatRetryMustSettleOnce(t *testing.T) {
	var hits atomic.Int32
	p, st, key := proxyFixture(t, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, goodEvents)
	})
	first := idemChat(p, key, idemID, proxyInput)
	if !strings.Contains(first.Body.String(), "[DONE]") || first.Header().Get("X-US-Idempotency") != "1" {
		t.Fatal(first.Body.String())
	}
	second := idemChat(p, key, idemID, proxyInput)
	if second.Code != 409 || idemCode(second) != "request_terminal" {
		t.Fatalf("replay of committed request: code=%d body=%q", second.Code, second.Body.String())
	}
	l, _ := st.License(key)
	if l.Credits != 499 || l.Reserved != 0 || hits.Load() != 1 {
		t.Fatalf("same request charged twice: credits=%d reserved=%d upstream=%d want=499/0/1", l.Credits, l.Reserved, hits.Load())
	}
}

// Repro 2: reserve ACK'i kaybolursa aynı kimlikli tekrar ikinci hold açmaz.
func TestPost12AReserveLostResponseMustNotLeak(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "post12a")
	key := h.Store.st.Orders["555"][0]
	api := &CreditHandler{Store: h.Store}
	var ids []string
	for i := 0; i < 2; i++ {
		w := idemReserve(api, key, idemID, `{"idempotency_id":"`+idemID+`"}`)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		var out struct{ ID string }
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		ids = append(ids, out.ID)
		if (i == 1) != (w.Header().Get("X-US-Replayed") == "true") {
			t.Fatalf("attempt %d X-US-Replayed=%q", i, w.Header().Get("X-US-Replayed"))
		}
	}
	l, _ := h.Store.License(key)
	if l.Reserved != 1 || ids[0] == "" || ids[0] != ids[1] {
		t.Fatalf("lost reserve response leaks hold: reserved=%d ids=%v want=1 same", l.Reserved, ids)
	}
	// Settled reservation is terminal: same id never opens a new paid hold.
	if err := h.Store.settleFor(key, ids[0], true, time.Now()); err != nil {
		t.Fatal(err)
	}
	w := idemReserve(api, key, idemID, "")
	if w.Code != 409 || idemCode(w) != "request_terminal" {
		t.Fatalf("terminal reserve replay: %d %s", w.Code, w.Body.String())
	}
	if l, _ = h.Store.License(key); l.Credits != 499 || l.Reserved != 0 {
		t.Fatal(l)
	}
}

func TestPost12AReserveHeaderBodyMismatchAndBadID(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "post12a")
	key := h.Store.st.Orders["555"][0]
	api := &CreditHandler{Store: h.Store}
	if w := idemReserve(api, key, idemID, `{"idempotency_id":"ffffffffffffffffffffffffffffffff"}`); w.Code != 400 || idemCode(w) != "idempotency_header_mismatch" {
		t.Fatalf("header/body disagree: %d %s", w.Code, w.Body.String())
	}
	for _, bad := range []string{"short", strings.Repeat("a", 65), "abc def ghi jkl mno pqr", "ğğğğğğğğğğğğğğğğ"} {
		if w := idemReserve(api, key, bad, ""); w.Code != 400 {
			t.Fatalf("bad id %q accepted: %d", bad, w.Code)
		}
	}
	if l, _ := h.Store.License(key); l.Reserved != 0 {
		t.Fatal(l)
	}
	// Legacy client without an id keeps the old one-shot behaviour.
	if w := idemReserve(api, key, "", `{}`); w.Code != 200 {
		t.Fatal(w.Code)
	}
}

func TestPost12AChatSameIDDifferentBodyOrPrincipalRejected(t *testing.T) {
	var hits atomic.Int32
	p, st, key := proxyFixture(t, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Error(w, "boom", 500)
	})
	if w := idemChat(p, key, idemID, proxyInput); w.Code == 200 && !strings.Contains(w.Body.String(), "error") {
		t.Fatal("upstream failure reported as success")
	}
	other := strings.Replace(proxyInput, "merhaba", "selam", 1)
	if w := idemChat(p, key, idemID, other); w.Code != 409 || idemCode(w) != "idempotency_mismatch" {
		t.Fatalf("different body: %d %s", w.Code, w.Body.String())
	}
	// Second license (different principal) reusing the id.
	h := &Handler{Store: st, Cfg: Config{WebhookSecret: "whsec", LicenseSecret: "lic", Plans: map[int64]Plan{42: {Kind: "sub", Days: 30, Prompts: 500, Credits: 500}}, Now: time.Now}}
	body2 := strings.Replace(payload, "555", "556", 1)
	post(h, body2, sign("whsec", body2), "p2")
	key2 := st.st.Orders["556"][0]
	if w := idemChat(p, key2, idemID, proxyInput); w.Code != 409 || idemCode(w) != "idempotency_mismatch" {
		t.Fatalf("different principal: %d %s", w.Code, w.Body.String())
	}
	if hits.Load() != 1 {
		t.Fatalf("rejected replays reached upstream: %d", hits.Load())
	}
	if l, _ := st.License(key2); l.Reserved != 0 || l.Credits != 500 {
		t.Fatal(l)
	}
}

// Transient failure + retry: same hold reused, exactly one charge, no extra hold.
func TestPost12AChatTransientRetrySharesHold(t *testing.T) {
	var hits atomic.Int32
	var st *Store
	var key string
	p, store, k := proxyFixture(t, func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if l, _ := st.License(key); l.Reserved != 1 {
			t.Errorf("attempt %d: reserved=%d want 1 (no extra hold)", n, l.Reserved)
		}
		if n == 1 {
			http.Error(w, "temporary", 502)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, goodEvents)
	})
	st, key = store, k
	idemChat(p, key, idemID, proxyInput)
	if l, _ := st.License(key); l.Reserved != 1 || l.Credits != 500 {
		t.Fatalf("failed attempt must park (not charge) the hold: reserved=%d credits=%d", l.Reserved, l.Credits)
	}
	w := idemChat(p, key, idemID, proxyInput)
	if !strings.Contains(w.Body.String(), "[DONE]") {
		t.Fatal(w.Body.String())
	}
	l, _ := st.License(key)
	status, _ := st.CreditStatus(key, time.Now())
	if l.Credits != 499 || l.Reserved != 0 || status.Committed != 1 || hits.Load() != 2 {
		t.Fatalf("credits=%d reserved=%d commits=%d upstream=%d", l.Credits, l.Reserved, status.Committed, hits.Load())
	}
}

// Replay records and their reservations survive a restart.
func TestPost12AReplayRecordsPersistAcrossRestart(t *testing.T) {
	h, path := setup(t)
	post(h, payload, sign("whsec", payload), "post12a")
	key := h.Store.st.Orders["555"][0]
	w := idemReserve(&CreditHandler{Store: h.Store}, key, idemID, "")
	var a struct{ ID string }
	_ = json.Unmarshal(w.Body.Bytes(), &a)
	if err := h.Store.Close(); err != nil {
		t.Fatal(err)
	}
	st, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	h.Store = st
	w = idemReserve(&CreditHandler{Store: st}, key, idemID, "")
	var b struct{ ID string }
	_ = json.Unmarshal(w.Body.Bytes(), &b)
	l, _ := st.License(key)
	if w.Code != 200 || a.ID == "" || a.ID != b.ID || l.Reserved != 1 {
		t.Fatalf("after restart: code=%d ids=%s/%s reserved=%d", w.Code, a.ID, b.ID, l.Reserved)
	}
}

// Orphan cleanup: parked/abandoned holds are released after their TTL,
// committed records are never turned into releases, old tombstones pruned.
func TestPost12AOrphanSweep(t *testing.T) {
	var hits atomic.Int32
	p, st, key := proxyFixture(t, func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			http.Error(w, "x", 502)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, goodEvents)
	})
	idemChat(p, key, idemID, proxyInput) // parked hold
	done := strings.Replace(idemID, "0", "1", -1)
	idemChat(p, key, done, proxyInput) // committed
	if l, _ := st.License(key); l.Reserved != 1 || l.Credits != 499 {
		t.Fatal(l)
	}
	now := time.Now()
	if n := st.SweepIdempotency(now); n != 0 {
		t.Fatalf("live hold swept early: %d", n)
	}
	later := now.Add(IdempotencyParkTTL + time.Second)
	st.SweepIdempotency(later)
	l, _ := st.License(key)
	status, _ := st.CreditStatus(key, later)
	if l.Reserved != 0 || l.Credits != 499 || status.Committed != 1 {
		t.Fatalf("after sweep: reserved=%d credits=%d committed=%d", l.Reserved, l.Credits, status.Committed)
	}
	st.mu.Lock()
	recDone := st.st.Idempotency[done]
	recPark := st.st.Idempotency[idemID]
	st.mu.Unlock()
	if recDone == nil || recDone.State != idemCommitted || recPark == nil || recPark.State != idemReleased {
		t.Fatalf("states: done=%+v park=%+v", recDone, recPark)
	}
	if w := idemChat(p, key, idemID, proxyInput); w.Code != 409 || idemCode(w) != "reservation_released" {
		t.Fatalf("released chat replay: %d %s", w.Code, w.Body.String())
	}
	if l, _ := st.License(key); l.Reserved != 0 {
		t.Fatal("released replay locked new credit")
	}
	st.SweepIdempotency(later.Add(IdempotencyRetention + time.Minute))
	st.mu.Lock()
	left := len(st.st.Idempotency)
	st.mu.Unlock()
	if left != 0 {
		t.Fatalf("tombstones not pruned: %d", left)
	}
}

// A committed (terminal) replay must not touch the GPU path at all.
func TestPost12ATerminalReplayDoesNotWake(t *testing.T) {
	var chats atomic.Int32
	p, _, key := proxyFixture(t, func(w http.ResponseWriter, r *http.Request) {
		chats.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, goodEvents)
	})
	idemChat(p, key, idemID, proxyInput)
	p.Upstream = "http://127.0.0.1:1" // any probe/wake would fail loudly
	w := idemChat(p, key, idemID, proxyInput)
	if w.Code != 409 || w.Header().Get("X-US-GPU-State") != "" || chats.Load() != 1 {
		t.Fatalf("terminal replay reached wake/upstream: %d %v", w.Code, w.Header())
	}
}

// API uyumu (a): Idempotency-Key alias; X-US-Request-ID ile farklıysa 400.
func TestPost12AIdempotencyKeyAlias(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "post12a")
	key := h.Store.st.Orders["555"][0]
	api := &CreditHandler{Store: h.Store}
	do := func(xus, alias string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/v1/credits/reserve", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer "+key)
		if xus != "" {
			r.Header.Set("X-US-Request-ID", xus)
		}
		if alias != "" {
			r.Header.Set("Idempotency-Key", alias)
		}
		w := httptest.NewRecorder()
		api.ServeHTTP(w, r)
		return w
	}
	if w := do(idemID, "ffffffffffffffffffffffffffffffff"); w.Code != 400 || idemCode(w) != "idempotency_header_mismatch" {
		t.Fatal(w.Code, w.Body.String())
	}
	a, b := do("", idemID), do(idemID, idemID)
	if a.Code != 200 || b.Code != 200 || a.Body.String() != b.Body.String() || b.Header().Get("X-US-Replayed") != "true" {
		t.Fatal(a.Body.String(), b.Body.String())
	}
	if l, _ := h.Store.License(key); l.Reserved != 1 {
		t.Fatal(l.Reserved)
	}
}

// API uyumu (d): serbest bırakılmış rezervasyonun tekrarı yeni hold açmaz.
func TestPost12AReleasedReplayLocksNoCredit(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "post12a")
	key := h.Store.st.Orders["555"][0]
	api := &CreditHandler{Store: h.Store}
	w := idemReserve(api, key, idemID, "")
	var out struct{ ID string }
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if err := h.Store.settleFor(key, out.ID, false, time.Now()); err != nil {
		t.Fatal(err)
	}
	w = idemReserve(api, key, idemID, "")
	if w.Code != 409 || idemCode(w) != "reservation_released" || w.Header().Get("X-US-Replayed") != "true" {
		t.Fatal(w.Code, w.Body.String())
	}
	// TTL expiry path (sweep) behaves the same.
	id2 := strings.Replace(idemID, "0", "2", -1)
	idemReserve(api, key, id2, "")
	h.Store.SweepIdempotency(time.Now().Add(ReservationTTL + time.Minute))
	if w = idemReserve(api, key, id2, ""); w.Code != 409 || idemCode(w) != "reservation_released" {
		t.Fatal(w.Code, w.Body.String())
	}
	if l, _ := h.Store.License(key); l.Reserved != 0 || l.Credits != 500 {
		t.Fatalf("released replay locked credit: reserved=%d credits=%d", l.Reserved, l.Credits)
	}
}

// API uyumu (c): replay günlük kotayı yeniden tüketmez. api-0006 bu ağaçta
// yok; kota sayacı API'nin auth katmanını taklit eden test sarmalayıcısıdır ve
// sözleşme IsIdempotentReplay'dir.
func TestPost12AReplayDoesNotConsumeDailyQuota(t *testing.T) {
	var day, minute int
	p, st, key := proxyFixture(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, goodEvents)
	})
	gate := func(id, body string) {
		minute++
		if !st.IsIdempotentReplay(key, id, "/v1/chat", []byte(body)) {
			day++
		}
	}
	for i := 0; i < 3; i++ {
		gate(idemID, proxyInput)
		idemChat(p, key, idemID, proxyInput)
	}
	gate(strings.Replace(idemID, "0", "3", -1), proxyInput) // new logical request
	other := strings.Replace(proxyInput, "merhaba", "selam", 1)
	gate(idemID, other) // mismatch is not a replay
	if day != 3 || minute != 5 {
		t.Fatalf("daily=%d minute=%d want 3/5", day, minute)
	}
	if st.IsIdempotentReplay(key, idemID, "/v1/credits/reserve", nil) {
		t.Fatal("chat record treated as reserve replay")
	}
}
