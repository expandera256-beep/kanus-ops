package licensesrv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenStoreEmptyFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "licenses.json")
	if err := os.WriteFile(p, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := OpenStore(p)
	if err != nil || st == nil {
		t.Fatalf("empty store must be usable: %v", err)
	}
	defer st.Close()
}

func TestOpenStoreWhitespaceAndInvalidFile(t *testing.T) {
	for _, input := range []string{"", "  \n\t"} {
		p := filepath.Join(t.TempDir(), "store.json")
		if err := os.WriteFile(p, []byte(input), 0600); err != nil {
			t.Fatal(err)
		}
		st, err := OpenStore(p)
		if err != nil || st == nil {
			t.Fatalf("store=%v err=%v", st, err)
		}
		st.mu.Lock()
		err = st.saveLocked()
		st.mu.Unlock()
		if err != nil {
			t.Fatal(err)
		}
		st.Close()
		if again, err := OpenStore(p); err != nil || again == nil {
			t.Fatal("restart failed", err)
		} else {
			again.Close()
		}
	}
	p := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(p, []byte("{"), 0600)
	if st, err := OpenStore(p); err == nil || st != nil {
		t.Fatal("invalid JSON accepted")
	}
}
