//go:build windows

package i18n

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func winStore(root, name string) *PrefStore {
	return &PrefStore{Root: func() string { return root }, Name: name}
}

// junction creates a directory junction (no admin rights needed).
func junction(t *testing.T, link, target string) {
	t.Helper()
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
		t.Skipf("mklink /J unavailable: %v %s", err, out)
	}
}

func TestWinPrefNormalSave(t *testing.T) {
	root := t.TempDir()
	p := winStore(root, "lang.json")
	if err := p.Set("en"); err != nil {
		t.Fatal(err)
	}
	if winStore(root, "lang.json").Get() != "en" {
		t.Fatal("not persisted")
	}
	ents, _ := os.ReadDir(root)
	if len(ents) != 1 {
		t.Fatalf("temp files left: %v", ents)
	}
}

func TestWinPrefJunctionBelowRootRejected(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	junction(t, filepath.Join(root, "sub"), outside)
	if err := winStore(root, "sub/lang.json").Set("en"); err == nil {
		t.Fatal("junction below root accepted")
	}
	if winStore(root, "sub/lang.json").Get() != "" {
		t.Fatal("read through junction")
	}
	if ents, _ := os.ReadDir(outside); len(ents) != 0 {
		t.Fatalf("wrote outside root: %v", ents)
	}
}

func TestWinPrefSymlinkFileRejected(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	victim := filepath.Join(outside, "victim.json")
	os.WriteFile(victim, []byte(`{"lang":"en"}`), 0o600)
	if err := os.Symlink(victim, filepath.Join(root, "lang.json")); err != nil {
		t.Skip("symlink privilege missing:", err)
	}
	if winStore(root, "lang.json").Get() != "" {
		t.Fatal("symlink followed on read")
	}
	if err := winStore(root, "lang.json").Set("tr"); err == nil {
		t.Fatal("symlink replaced/followed on write")
	}
	if b, _ := os.ReadFile(victim); string(b) != `{"lang":"en"}` {
		t.Fatal("victim modified")
	}
}

func TestWinPrefJunctionRootAllowedAndPinned(t *testing.T) {
	base := t.TempDir()
	a, b := filepath.Join(base, "a"), filepath.Join(base, "b")
	os.MkdirAll(a, 0o700)
	os.MkdirAll(b, 0o700)
	link := filepath.Join(base, "cfg")
	junction(t, link, a)
	p := winStore(link, "lang.json")
	if err := p.Set("en"); err != nil {
		t.Fatalf("junctioned root rejected: %v", err)
	}
	os.Remove(link)
	junction(t, link, b)
	if err := p.Set("tr"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(b, "lang.json")); err == nil {
		t.Fatal("write followed re-pointed root")
	}
}
