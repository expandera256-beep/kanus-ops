package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUndoModelChangeRestoresOriginalAndRefusesLaterEdits(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("old\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if note, _ := applyReplyFilesChecked(root, "DOSYA app.txt\nnew\n", map[string]bool{"app.txt": true}); note == "" {
		t.Fatal("write failed")
	}
	if got, err := undoLastChange(root); err != nil || got != "app.txt" {
		t.Fatalf("%s %v", got, err)
	}
	if got, _ := os.ReadFile(target); string(got) != "old\n" {
		t.Fatalf("%q", got)
	}
	applyReplyFilesChecked(root, "DOSYA app.txt\nnew\n", map[string]bool{"app.txt": true})
	if err := os.WriteFile(target, []byte("human\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := undoLastChange(root); err == nil {
		t.Fatal("modified file must be preserved")
	}
	if got, _ := os.ReadFile(target); string(got) != "human\n" {
		t.Fatalf("%q", got)
	}
}

func TestUndoNewFileAndProtectBackupDirectory(t *testing.T) {
	root := t.TempDir()
	if note := applyReplyFiles(root, "DOSYA new.txt\nhello\n"); note == "" {
		t.Fatal("no file")
	}
	if _, err := undoLastChange(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "new.txt")); !os.IsNotExist(err) {
		t.Fatal("new file remains")
	}
	if note := applyReplyFiles(root, "DOSYA .us-history/evil.json\n{}\n"); note != "" {
		t.Fatal(note)
	}
}

func TestModelCannotFollowSymlinkOutsideWorkspace(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	target := filepath.Join(outside, "target.txt")
	if err := os.WriteFile(target, []byte("safe"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skip("symlink unavailable")
	}
	if note := applyReplyFiles(root, "DOSYA link/target.txt\nchanged\n"); note != "" {
		t.Fatal(note)
	}
	if got, _ := os.ReadFile(target); string(got) != "safe" {
		t.Fatalf("%q", got)
	}
	s := newView("q8")
	s.root = root
	if _, ok, _ := s.localText("link/target.txt"); ok {
		t.Fatal("outside file read")
	}
	if names := strings.Join(workspaceNames(root), "\n"); strings.Contains(names, ".us-history") {
		t.Fatal(names)
	}
}

func TestChangedFileSinceModelReadIsNotOverwritten(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "data.txt")
	if err := os.WriteFile(target, []byte("old\n"), 0600); err != nil {
		t.Fatal(err)
	}
	s := newView("q8")
	s.root, s.key = root, "test"
	s.modelCh = make(chan modelResult, 1)
	calls := 0
	s.sendModel = func(_ workerID, _, wire string) modelResult {
		calls++
		if calls == 1 {
			return modelResult{ok: true, body: "OKU data.txt"}
		}
		if !strings.Contains(wire, "OKU data.txt\nold") {
			t.Fatal("file not read")
		}
		if err := os.WriteFile(target, []byte("human change\n"), 0600); err != nil {
			t.Fatal(err)
		}
		return modelResult{ok: true, body: "DOSYA data.txt\nmodel change\nBİTTİ\n"}
	}
	s.draft = "projeyi onar"
	s.submit()
	if got, _ := os.ReadFile(target); string(got) != "human change\n" {
		t.Fatalf("user edit lost: %q", got)
	}
}
