package privacy

import (
	"path/filepath"
	"testing"
	"time"
)

func TestPrefStoreRoundTripAndDeleteError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "p.json")
	ps, err := OpenPrefStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := ps.Put(Preferences{AccountID: "a1", MarketingConsent: true, Source: "t"}, time.Now()); err != nil {
		t.Fatal(err)
	}
	ps2, err := OpenPrefStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if !ps2.Get("a1").MarketingConsent {
		t.Fatal("missing")
	}
	if err := ps2.Delete("a1"); err != nil {
		t.Fatal(err)
	}
	broken := &PrefStore{path: "", byID: map[string]Preferences{}}
	if err := broken.Delete("x"); err == nil {
		t.Fatal("expected delete error on empty path")
	}
}
