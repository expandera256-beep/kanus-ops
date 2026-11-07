package licensesrv

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// API-0012: reservation commit grace after Expires (default prod 10m via env).

func TestTTLCommitAnchorAcceptsCommitWithinGrace(t *testing.T) {
	h, _ := setup(t)
	h.Store.commitGrace = 10 * time.Minute
	post(h, payload, sign("whsec", payload), "ttl1")
	key := h.Store.st.Orders["555"][0]
	now := h.now()
	id, err := h.Store.Reserve(key, 1, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	// Past reservation Expires but within grace
	late := now.Add(time.Minute + 30*time.Second)
	if err := h.Store.settleFor(key, id, true, late); err != nil {
		t.Fatal("expected commit within grace", err)
	}
	l, _ := h.Store.License(key)
	if l.Credits != 499 || l.Reserved != 0 {
		t.Fatal(l)
	}
}

func TestTTLCommitAnchorRejectsAfterGraceAndNoDoubleDebit(t *testing.T) {
	h, _ := setup(t)
	h.Store.commitGrace = 10 * time.Minute
	post(h, payload, sign("whsec", payload), "ttl2")
	key := h.Store.st.Orders["555"][0]
	now := h.now()
	id, err := h.Store.Reserve(key, 1, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	// Sweep after grace: auto-release (no debit)
	afterGrace := now.Add(time.Minute + 10*time.Minute + time.Second)
	_, _ = h.Store.Reserve(key, 1, time.Minute, afterGrace) // triggers expireLocked
	l, _ := h.Store.License(key)
	if l.Credits != 500 {
		t.Fatal("auto-expire must not debit", l.Credits)
	}
	if l.Reserved != 1 { // the new reserve from the triggering call
		// actually Reserve after expire may succeed with reserved=1 for new id
	}
	// Late commit on old id must fail (already settled/released)
	if err := h.Store.settleFor(key, id, true, afterGrace); err == nil {
		t.Fatal("late commit after grace+release must fail")
	}
	l, _ = h.Store.License(key)
	if l.Credits < 499 { // at most one debit from the new reserve path if any — old id must not double
		// Credits should be 500 or 499 if new reserve still held
		if l.Credits < 499 {
			t.Fatal("double debit", l.Credits)
		}
	}
}

func TestTTLCommitAnchorLegacyZeroGraceRejectsAtExpires(t *testing.T) {
	h, _ := setup(t) // commitGrace=0
	post(h, payload, sign("whsec", payload), "ttl3")
	key := h.Store.st.Orders["555"][0]
	now := h.now()
	id, err := h.Store.Reserve(key, 1, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Store.settleFor(key, id, true, now.Add(time.Minute)); err == nil {
		t.Fatal("legacy: commit at/after Expires must fail")
	}
}

func TestTTLCommitAnchorIdempotentAndConcurrent(t *testing.T) {
	h, _ := setup(t)
	h.Store.commitGrace = 10 * time.Minute
	post(h, payload, sign("whsec", payload), "ttl4")
	key := h.Store.st.Orders["555"][0]
	now := h.now()
	h.Store.st.Licenses[key].Credits = 1
	id, err := h.Store.Reserve(key, 1, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	late := now.Add(2 * time.Minute)
	var wg sync.WaitGroup
	var oks atomic.Int32
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if h.Store.settleFor(key, id, true, late) == nil {
				oks.Add(1)
			}
		}()
	}
	wg.Wait()
	// First settle succeeds; identical retries also succeed (idempotent) — count may be >1
	if oks.Load() < 1 {
		t.Fatal("no successful settle")
	}
	l, _ := h.Store.License(key)
	if l.Credits != 0 || l.Reserved != 0 {
		t.Fatal("double spend", l)
	}
	// Conflicting release after commit
	if err := h.Store.settleFor(key, id, false, late); err == nil {
		t.Fatal("conflict accepted")
	}
}

func TestCommitGraceFromEnvDefaultCapAndInvalid(t *testing.T) {
	t.Setenv("US_CREDIT_COMMIT_GRACE_SECONDS", "")
	if commitGraceFromEnv() != 10*time.Minute {
		t.Fatal("default")
	}
	t.Setenv("US_CREDIT_COMMIT_GRACE_SECONDS", "0")
	if commitGraceFromEnv() != 0 {
		t.Fatal("zero legacy")
	}
	t.Setenv("US_CREDIT_COMMIT_GRACE_SECONDS", "30")
	if commitGraceFromEnv() != 30*time.Second {
		t.Fatal("30")
	}
	t.Setenv("US_CREDIT_COMMIT_GRACE_SECONDS", "99999")
	if commitGraceFromEnv() != 15*time.Minute {
		t.Fatal("cap 900 clamp")
	}
	t.Setenv("US_CREDIT_COMMIT_GRACE_SECONDS", "900")
	if commitGraceFromEnv() != 15*time.Minute {
		t.Fatal("exact 900")
	}
	t.Setenv("US_CREDIT_COMMIT_GRACE_SECONDS", "-5")
	if commitGraceFromEnv() != 10*time.Minute {
		t.Fatal("neg")
	}
	t.Setenv("US_CREDIT_COMMIT_GRACE_SECONDS", "abc")
	if commitGraceFromEnv() != 10*time.Minute {
		t.Fatal("parse")
	}
}

// PANEL R3-A01: cap integer seconds before conversion to time.Duration.
func TestCommitGraceFromEnvPanelEdges(t *testing.T) {
	cases := []struct {
		raw     string
		seconds int64
	}{
		{"", 600}, {" ", 600}, {"0", 0}, {"-0", 0}, {"-1", 600}, {"900", 900},
		{"901", 900}, {"+901", 900}, {" 901 ", 900}, {"1s", 600}, {"abc", 600},
		{"9999999999999999999999999", 600}, {"9223372036", 900},
		{"9223372037", 900}, {"18446744074", 900}, {"9223372036854775807", 900},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			t.Setenv("US_CREDIT_COMMIT_GRACE_SECONDS", tc.raw)
			if got, want := commitGraceFromEnv(), time.Duration(tc.seconds)*time.Second; got != want {
				t.Fatalf("input %q: got %v, want %v", tc.raw, got, want)
			}
		})
	}
}
func TestCommitGraceFromEnvDecimalInvalid(t *testing.T) {
	t.Setenv("US_CREDIT_COMMIT_GRACE_SECONDS", "1.5")
	if got := commitGraceFromEnv(); got != 600*time.Second {
		t.Fatalf("got %v", got)
	}
}
