package licensesrv

import (
	"testing"
	"time"
)

// API-0012: grace içinde sweep rezervasyonu serbest bırakmaz (Reserved kilitli),
// grace sonrası sweep debit etmeden release eder, geç commit reddedilir.
func TestTTLGraceKeepsReservedLockedThenReleases(t *testing.T) {
	h, _ := setup(t)
	h.Store.commitGrace = 10 * time.Minute
	post(h, payload, sign("whsec", payload), "ttlLock")
	key := h.Store.st.Orders["555"][0]
	now := h.now()
	id, err := h.Store.Reserve(key, 1, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	inGrace := now.Add(time.Minute + 30*time.Second)
	id2, err := h.Store.Reserve(key, 1, time.Hour, inGrace) // triggers sweep
	if err != nil {
		t.Fatal(err)
	}
	if l, _ := h.Store.License(key); l.Reserved != 2 || l.Credits != 500 {
		t.Fatalf("in grace: want reserved=2 credits=500, got reserved=%d credits=%d", l.Reserved, l.Credits)
	}
	afterGrace := now.Add(time.Minute + 10*time.Minute + time.Second)
	id3, err := h.Store.Reserve(key, 1, time.Hour, afterGrace) // sweep releases id
	if err != nil {
		t.Fatal(err)
	}
	if l, _ := h.Store.License(key); l.Reserved != 2 || l.Credits != 500 {
		t.Fatalf("after grace: want reserved=2 (id2,id3) credits=500, got reserved=%d credits=%d", l.Reserved, l.Credits)
	}
	if err := h.Store.settleFor(key, id, true, afterGrace); err == nil {
		t.Fatal("late commit after grace must fail")
	}
	for _, x := range []string{id2, id3} {
		if err := h.Store.settleFor(key, x, true, afterGrace); err != nil {
			t.Fatal(err)
		}
	}
	if l, _ := h.Store.License(key); l.Reserved != 0 || l.Credits != 498 {
		t.Fatalf("final: want reserved=0 credits=498, got reserved=%d credits=%d", l.Reserved, l.Credits)
	}
}
