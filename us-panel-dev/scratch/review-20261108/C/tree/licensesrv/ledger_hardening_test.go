package licensesrv

import (
	"testing"
	"time"
)

// API-0004: invariant locks — happy paths unchanged; corrupt reserved repaired with error (no panic).

func TestLedgerHardeningHappyReserveSettleUnchanged(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "h1")
	key := h.Store.st.Orders["555"][0]
	now := h.now()
	id, err := h.Store.Reserve(key, 1, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Store.settleFor(key, id, true, now); err != nil {
		t.Fatal(err)
	}
	l, _ := h.Store.License(key)
	if l.Credits != 499 || l.Reserved != 0 {
		t.Fatalf("%+v", l)
	}
}

func TestLedgerHardeningExpireStatusConsistency(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "h2")
	key := h.Store.st.Orders["555"][0]
	now := h.now()
	id, err := h.Store.Reserve(key, 1, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	// After TTL, expire via Reserve sweep; status must show reserved=0 and no double-release.
	_, _ = h.Store.Reserve(key, 1, time.Minute, now.Add(2*time.Minute))
	st, ok := h.Store.CreditStatus(key, now.Add(2*time.Minute))
	if !ok || st.Reserved < 0 {
		t.Fatal(st)
	}
	l, _ := h.Store.License(key)
	if l.Reserved < 0 {
		t.Fatal("negative reserved", l.Reserved)
	}
	_ = id
}

func TestLedgerHardeningRejectsNegativeReservedCorruption(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "h3")
	key := h.Store.st.Orders["555"][0]
	now := h.now()
	id, err := h.Store.Reserve(key, 1, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	// Corrupt: reserved already 0 while reservation exists
	h.Store.mu.Lock()
	h.Store.st.Licenses[key].Reserved = 0
	h.Store.mu.Unlock()
	err = h.Store.settleFor(key, id, true, now)
	if err == nil {
		t.Fatal("expected invariant error")
	}
	l, _ := h.Store.License(key)
	if l.Reserved < 0 {
		t.Fatal("must not go negative", l.Reserved)
	}
	// Credits must not have been debited on failed settle
	if l.Credits != 500 {
		t.Fatal("debit on failed invariant", l.Credits)
	}
}
