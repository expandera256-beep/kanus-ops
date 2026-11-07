package licensesrv
import("testing";"time")
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
