//go:build !windows

package i18n

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func store(path string) *PrefStore { return &PrefStore{Path: func() string { return path }} }

// G-01: a FIFO must not block Get (nor leave the mutex held for Set).
func TestPrefGetFIFODoesNotBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lang.json")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Skip("mkfifo:", err)
	}
	t.Cleanup(func() { // unblock a stuck reader of the old implementation
		if f, err := os.OpenFile(path, os.O_WRONLY|syscall.O_NONBLOCK, 0); err == nil {
			f.Close()
		}
	})
	p := store(path)
	done := make(chan string, 1)
	go func() { done <- p.Get() }()
	select {
	case got := <-done:
		if got != "" {
			t.Fatalf("FIFO must yield default, got %q", got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Get blocked on FIFO")
	}
	setDone := make(chan error, 1)
	go func() { setDone <- store(filepath.Join(t.TempDir(), "lang.json")).Set("en") }()
	select {
	case <-setDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Set blocked")
	}
}

// G-01: a symlinked preference file must be ignored.
func TestPrefGetRejectsSymlinkFile(t *testing.T) {
	dir, out := t.TempDir(), t.TempDir()
	target := filepath.Join(out, "target.json")
	os.WriteFile(target, []byte(`{"lang":"en"}`), 0o600)
	path := filepath.Join(dir, "lang.json")
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if got := store(path).Get(); got != "" {
		t.Fatalf("symlink followed: %q", got)
	}
}

// G-01: huge (sparse) file must be rejected without reading it all.
func TestPrefGetRejectsLargeFileQuickly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lang.json")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"lang":"en"}`)
	f.Truncate(1 << 30) // 1 GiB sparse
	f.Close()
	start := time.Now()
	got := store(path).Get()
	if got != "" {
		t.Fatalf("large file accepted: %q", got)
	}
	if d := time.Since(start); d > 150*time.Millisecond {
		t.Fatalf("large file read took %v (unbounded read)", d)
	}
}

// G-02: a symlinked parent directory must not redirect the write.
func TestPrefSetRejectsSymlinkParent(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "parent")); err != nil {
		t.Fatal(err)
	}
	p := store(filepath.Join(root, "parent", "lang.json"))
	if err := p.Set("en"); err == nil {
		t.Error("Set through symlinked parent succeeded")
	}
	if _, err := os.Lstat(filepath.Join(outside, "lang.json")); err == nil {
		t.Fatal("write escaped to symlink target directory")
	}
	ents, _ := os.ReadDir(outside)
	if len(ents) != 0 {
		t.Fatalf("files created outside root: %v", ents)
	}
}

// G-02: a symlinked final component must be rejected, target untouched.
func TestPrefSetRejectsSymlinkFile(t *testing.T) {
	dir, out := t.TempDir(), t.TempDir()
	victim := filepath.Join(out, "victim")
	os.WriteFile(victim, []byte("keep"), 0o600)
	path := filepath.Join(dir, "lang.json")
	os.Symlink(victim, path)
	if err := store(path).Set("en"); err == nil {
		t.Error("Set over symlink succeeded")
	}
	if b, _ := os.ReadFile(victim); string(b) != "keep" {
		t.Fatal("victim modified")
	}
}

// Regression: normal write stays inside its directory, 0600, round-trips.
func TestPrefSetStaysInsideRoot(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "us", "lang.json")
	p := store(path)
	if err := p.Set("en"); err != nil {
		t.Fatal(err)
	}
	if p.Get() != "en" {
		t.Fatal("roundtrip")
	}
	fi, err := os.Lstat(path)
	if err != nil || !fi.Mode().IsRegular() || fi.Mode().Perm() != 0o600 {
		t.Fatalf("bad file %v %v", fi, err)
	}
	ents, _ := os.ReadDir(filepath.Dir(path))
	if len(ents) != 1 {
		t.Fatalf("temp files left: %v", ents)
	}
}
