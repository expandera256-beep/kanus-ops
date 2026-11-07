//go:build linux

package main

import (
	"encoding/json"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestR13ParentSwapPathTakingWriter(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	parent := filepath.Join(root, "nested")
	old := filepath.Join(root, "old")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(out, "victim"), []byte("outside"), 0600)
	atomicWriteTestHook = func(string) error {
		if err := os.Rename(parent, old); err != nil {
			return err
		}
		return os.Symlink(out, parent)
	}
	defer func() { atomicWriteTestHook = nil }()
	if err := writeModelFile(root, filepath.Join(parent, "victim"), []byte("inside")); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(out, "victim")); err != nil || string(b) != "outside" {
		t.Fatal("parent-swap escaped")
	}
	if b, err := os.ReadFile(filepath.Join(old, "victim")); err != nil || string(b) != "inside" {
		t.Fatalf("pinned parent write %q %v", b, err)
	}
}
func TestR13HistoryPrivateModesAndNestedNewFile(t *testing.T) {
	root := t.TempDir()
	name, err := recordChange(root, "nested/new.txt", []byte("new"))
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(name)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("history file mode %v %v", info, err)
	}
	info, err = os.Stat(filepath.Join(root, changeHistoryDir))
	if err != nil || info.Mode().Perm() != 0700 {
		t.Fatalf("history dir mode %v %v", info, err)
	}
	if err := writeModelFile(root, filepath.Join(root, "nested/new.txt"), []byte("new")); err != nil {
		t.Fatal(err)
	}
	if _, err := undoLastChange(root); err != nil {
		t.Fatal(err)
	}
}

// The old desk/history helpers used checked paths and then ordinary opens.
// Exercise their actual entry points while a parent becomes an outside link.
func TestR13DeskHistoryParentSwap(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	parent, link := filepath.Join(root, "nested"), filepath.Join(root, "link")
	os.Mkdir(parent, 0700)
	os.WriteFile(filepath.Join(parent, "file.txt"), []byte("inside"), 0600)
	os.WriteFile(filepath.Join(out, "file.txt"), []byte("OUTSIDE_SENTINEL"), 0600)
	os.Symlink(out, link)
	var stop atomic.Bool
	var swaps atomic.Int64
	done := make(chan struct{})
	go func() {
		defer close(done)
		for !stop.Load() {
			if unix.Renameat2(unix.AT_FDCWD, parent, unix.AT_FDCWD, link, unix.RENAME_EXCHANGE) == nil {
				swaps.Add(1)
			}
		}
	}()
	defer func() { stop.Store(true); <-done }()
	s := newView("q8")
	s.root = root
	reads, snapshots := 0, 0
	for i := 0; i < 200; i++ {
		text, ok, _ := s.localText("nested/file.txt")
		if ok {
			reads++
			if text != "inside" {
				t.Fatalf("desk read escaped: %q", text)
			}
		}
		name, err := recordChange(root, "nested/file.txt", []byte("next"))
		if err == nil {
			snapshots++
			body, e := os.ReadFile(name)
			if e != nil {
				t.Fatal(e)
			}
			var rec changeRecord
			if e = json.Unmarshal(body, &rec); e != nil {
				t.Fatal(e)
			}
			if string(rec.Old) != "inside" {
				t.Fatalf("history snapshot escaped: %q", rec.Old)
			}
		}
	}
	if swaps.Load() == 0 || reads == 0 || snapshots == 0 {
		t.Fatalf("vacuous swaps=%d reads=%d snapshots=%d", swaps.Load(), reads, snapshots)
	}
	t.Logf("swaps=%d reads=%d snapshots=%d", swaps.Load(), reads, snapshots)
}
