//go:build !windows

package i18n

import (
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func store(path string) *PrefStore {
	return &PrefStore{Root: func() string { return filepath.Dir(path) }, Name: filepath.Base(path)}
}

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
	// root is the trusted config root; "parent" is a symlink BELOW it.
	p := &PrefStore{Root: func() string { return root }, Name: "parent/lang.json"}
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

// R2-01: a legitimately symlinked config root (/home -> /var/home) works.
func TestPrefSymlinkedRootPersists(t *testing.T) {
	base := t.TempDir()
	os.MkdirAll(filepath.Join(base, "var", "home", "u", ".config", "us"), 0o700)
	if err := os.Symlink(filepath.Join(base, "var", "home"), filepath.Join(base, "home")); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(base, "home", "u", ".config", "us")
	p := &PrefStore{Root: func() string { return root }, Name: "lang.json"}
	if err := p.Set("en"); err != nil {
		t.Fatalf("save through symlinked root failed: %v", err)
	}
	if p.Get() != "en" {
		t.Fatal("not read back")
	}
	q := &PrefStore{Root: func() string { return root }, Name: "lang.json"} // new process
	if q.Get() != "en" {
		t.Fatal("not persisted")
	}
	if _, err := os.Lstat(filepath.Join(base, "var", "home", "u", ".config", "us", "lang.json")); err != nil {
		t.Fatal("file not in resolved root:", err)
	}
}

// R2-01: symlinks BELOW the (symlinked) root stay rejected.
func TestPrefSymlinkBelowSymlinkedRootRejected(t *testing.T) {
	base, outside := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(base, "real"), 0o700)
	os.Symlink(filepath.Join(base, "real"), filepath.Join(base, "cfg"))
	os.Symlink(outside, filepath.Join(base, "real", "sub"))
	p := &PrefStore{Root: func() string { return filepath.Join(base, "cfg") }, Name: "sub/lang.json"}
	if err := p.Set("en"); err == nil {
		t.Fatal("symlink below root accepted")
	}
	os.Symlink(filepath.Join(outside, "x.json"), filepath.Join(base, "real", "lang.json"))
	p2 := &PrefStore{Root: func() string { return filepath.Join(base, "cfg") }, Name: "lang.json"}
	if err := p2.Set("en"); err == nil {
		t.Fatal("symlinked file below root accepted")
	}
	if ents, _ := os.ReadDir(outside); len(ents) != 0 {
		t.Fatalf("wrote outside: %v", ents)
	}
}

// R2-01: re-pointing the root symlink after first use must not move writes
// (the root is pinned as a directory fd).
func TestPrefRootRebindPinned(t *testing.T) {
	base := t.TempDir()
	a, b := filepath.Join(base, "a"), filepath.Join(base, "b")
	os.MkdirAll(a, 0o700)
	os.MkdirAll(b, 0o700)
	link := filepath.Join(base, "cfg")
	os.Symlink(a, link)
	p := &PrefStore{Root: func() string { return link }, Name: "lang.json"}
	if err := p.Set("en"); err != nil {
		t.Fatal(err)
	}
	os.Remove(link)
	os.Symlink(b, link)
	if err := p.Set("tr"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(b, "lang.json")); err == nil {
		t.Fatal("write followed rebound root")
	}
	if raw, _ := os.ReadFile(filepath.Join(a, "lang.json")); string(raw) != `{"lang":"tr"}` {
		t.Fatalf("pinned root not updated: %q", raw)
	}
}

// R2-01: an unopenable root keeps the choice in memory, Set reports it.
func TestPrefUnavailableRootMemoryOnly(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file")
	os.WriteFile(f, []byte("x"), 0o600)
	p := &PrefStore{Root: func() string { return filepath.Join(f, "sub") }, Name: "lang.json"} // parent is a file
	err := p.Set("en")
	if !errors.Is(err, ErrNotPersisted) {
		t.Fatalf("want ErrNotPersisted, got %v", err)
	}
	if p.Get() != "en" {
		t.Fatal("memory fallback missing")
	}
}

// R2-01 HTTP: symlinked root -> 200 and persisted; unavailable root -> 503.
func TestPrefHandlerSymlinkedRootAndUnavailable(t *testing.T) {
	base := t.TempDir()
	os.MkdirAll(filepath.Join(base, "var", "home"), 0o700)
	os.Symlink(filepath.Join(base, "var", "home"), filepath.Join(base, "home"))
	ok := &PrefStore{Root: func() string { return filepath.Join(base, "home") }, Name: "lang.json"}
	post := func(p *PrefStore) int {
		r := httptest.NewRequest("POST", "/api/lang", strings.NewReader(`{"lang":"en"}`))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		PrefHandler(p, nil).ServeHTTP(w, r)
		return w.Code
	}
	if c := post(ok); c != 200 || ok.Get() != "en" {
		t.Fatalf("symlinked root: %d %q", c, ok.Get())
	}
	f := filepath.Join(base, "file")
	os.WriteFile(f, nil, 0o600)
	bad := &PrefStore{Root: func() string { return filepath.Join(f, "x") }, Name: "lang.json"}
	if c := post(bad); c != 503 {
		t.Fatalf("unavailable root status %d", c)
	}
}
