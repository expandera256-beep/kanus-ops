//go:build unix

package privacy

import (
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestPrefStoreRefusesFIFO(t *testing.T) {
	dir := t.TempDir()
	fifo := filepath.Join(dir, "fifo.json")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := OpenPrefStore(fifo); done <- err }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("FIFO should be refused without blocking")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("OpenPrefStore blocked on FIFO")
	}
}
