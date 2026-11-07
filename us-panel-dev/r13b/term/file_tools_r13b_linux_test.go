//go:build linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// r13b: session-level root identity. Layout:
//
//	base/parent/ws  (session root, pinned by the first desk operation)
//	out/ws          (attacker tree with the same relative names)
//
// After the pin, base/parent is renamed to base/old and replaced with a
// symlink to out. A path re-resolution of s.root would land in out/ws.
type r13bSwap struct {
	base, ws, oldWS, outWS string
}

func newR13bSwap(t *testing.T) *r13bSwap {
	t.Helper()
	base, out := t.TempDir(), t.TempDir()
	ws := filepath.Join(base, "parent", "ws")
	outWS := filepath.Join(out, "ws")
	for _, d := range []string{filepath.Join(ws, "sub"), filepath.Join(outWS, "sub")} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite(t, filepath.Join(ws, "sub", "victim"), "inside")
	mustWrite(t, filepath.Join(outWS, "sub", "victim"), "outside")
	return &r13bSwap{base: base, ws: ws, oldWS: filepath.Join(base, "old", "ws"), outWS: outWS}
}

func (r *r13bSwap) swap(t *testing.T) {
	t.Helper()
	if err := os.Rename(filepath.Join(r.base, "parent"), filepath.Join(r.base, "old")); err != nil {
		t.Fatalf("barrier: rename failed: %v", err)
	}
	if err := os.Symlink(filepath.Dir(r.outWS), filepath.Join(r.base, "parent")); err != nil {
		t.Fatalf("barrier: symlink failed: %v", err)
	}
	// Barrier: prove the swap is live; the ambient path now resolves outside.
	if b, err := os.ReadFile(filepath.Join(r.ws, "sub", "victim")); err != nil || string(b) != "outside" {
		t.Fatalf("barrier: swap not effective: %q %v", b, err)
	}
}

func mustWrite(t *testing.T, p, s string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(s), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readStr(p string) string { b, _ := os.ReadFile(p); return string(b) }

func (r *r13bSwap) assertOutsideUntouched(t *testing.T) {
	t.Helper()
	if got := readStr(filepath.Join(r.outWS, "sub", "victim")); got != "outside" {
		t.Fatalf("outside sentinel changed: %q", got)
	}
	if _, err := os.Lstat(filepath.Join(r.outWS, changeHistoryDir)); err == nil {
		t.Fatal("history created outside the pinned session root")
	}
}

func r13bSession(t *testing.T, ws string) *session {
	t.Helper()
	s := &session{root: ws}
	// First desk operation: this is the session-start pin.
	if txt, ok, _ := s.localText("sub/victim"); !ok || txt != "inside" {
		t.Fatalf("initial read %q %v", txt, ok)
	}
	return s
}

func TestR13bSessionAncestorSwapRead(t *testing.T) {
	r := newR13bSwap(t)
	s := r13bSession(t, r.ws)
	r.swap(t)
	if txt, ok, _ := s.localText("sub/victim"); ok && strings.Contains(txt, "outside") {
		t.Fatalf("desk read outside sentinel after ancestor swap: %q", txt)
	}
}

func TestR13bSessionAncestorSwapSave(t *testing.T) {
	r := newR13bSwap(t)
	s := r13bSession(t, r.ws)
	s.readFile("sub/victim")
	r.swap(t)
	s.saveFile("sub/victim", "saved")
	r.assertOutsideUntouched(t)
	if got := readStr(filepath.Join(r.oldWS, "sub", "victim")); got != "saved" {
		t.Fatalf("pinned save: got %q notice %q", got, s.notice)
	}
}

func TestR13bSessionAncestorSwapAppend(t *testing.T) {
	r := newR13bSwap(t)
	s := r13bSession(t, r.ws)
	r.swap(t)
	s.appendFile("sub/victim", "+tail")
	r.assertOutsideUntouched(t)
	if got := readStr(filepath.Join(r.oldWS, "sub", "victim")); got != "inside+tail" {
		t.Fatalf("pinned append: got %q notice %q", got, s.notice)
	}
}

func TestR13bSessionAncestorSwapUndo(t *testing.T) {
	r := newR13bSwap(t)
	s := r13bSession(t, r.ws)
	s.readFile("sub/victim")
	s.saveFile("sub/victim", "saved")
	r.swap(t)
	mustWrite(t, filepath.Join(r.outWS, "sub", "victim"), "saved")
	if _, err := s.undoLast(); err != nil {
		t.Fatalf("undo on pinned root: %v", err)
	}
	if got := readStr(filepath.Join(r.oldWS, "sub", "victim")); got != "inside" {
		t.Fatalf("undo restored %q", got)
	}
	if got := readStr(filepath.Join(r.outWS, "sub", "victim")); got != "saved" {
		t.Fatalf("undo touched outside: %q", got)
	}
}

// R13-02 cleanup path: the write fails after the history record exists; the
// cleanup must delete the pinned record, never a same-named outside entry.
func TestR13bSaveCleanupRace(t *testing.T) {
	r := newR13bSwap(t)
	s := r13bSession(t, r.ws)
	s.readFile("sub/victim")
	var planted string
	atomicWriteTestHook = func(tmp string) error {
		if filepath.Base(filepath.Dir(tmp)) != "sub" {
			return nil // history record write; let it land
		}
		ents, err := os.ReadDir(filepath.Join(r.ws, changeHistoryDir))
		if err != nil || len(ents) != 1 {
			t.Errorf("barrier: expected one pinned record: %v %d", err, len(ents))
			return os.ErrInvalid
		}
		r.swap(t)
		planted = filepath.Join(r.outWS, changeHistoryDir, ents[0].Name())
		os.MkdirAll(filepath.Dir(planted), 0o700)
		mustWrite(t, planted, "outside-record")
		return os.ErrPermission
	}
	defer func() { atomicWriteTestHook = nil }()
	s.saveFile("sub/victim", "saved")
	if planted == "" {
		t.Fatal("barrier: hook never ran")
	}
	if got := readStr(planted); got != "outside-record" {
		t.Fatal("cleanup removed an outside entry")
	}
	if ents, _ := os.ReadDir(filepath.Join(r.oldWS, changeHistoryDir)); len(ents) != 0 {
		t.Fatalf("pinned record not cleaned: %d", len(ents))
	}
	if got := readStr(filepath.Join(r.oldWS, "sub", "victim")); got != "inside" {
		t.Fatalf("failed save changed file: %q", got)
	}
}

// Session root itself replaced (not an ancestor) after the pin.
func TestR13bSessionRootSwap(t *testing.T) {
	r := newR13bSwap(t)
	s := r13bSession(t, r.ws)
	moved := filepath.Join(r.base, "parent", "ws-moved")
	if err := os.Rename(r.ws, moved); err != nil {
		t.Fatalf("barrier: %v", err)
	}
	if err := os.Symlink(r.outWS, r.ws); err != nil {
		t.Fatalf("barrier: %v", err)
	}
	s.appendFile("sub/victim", "+x")
	if got := readStr(filepath.Join(r.outWS, "sub", "victim")); got != "outside" {
		t.Fatalf("root swap escaped: %q", got)
	}
	if got := readStr(filepath.Join(moved, "sub", "victim")); got != "inside+x" {
		t.Fatalf("pinned root append %q notice %q", got, s.notice)
	}
}
