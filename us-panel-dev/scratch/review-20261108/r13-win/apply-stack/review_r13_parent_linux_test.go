//go:build linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestR13ReviewHistoryParentSwap(t *testing.T) {
	p := t.TempDir()
	parent := filepath.Join(p, "parent")
	old := filepath.Join(p, "old")
	out := filepath.Join(p, "outside")
	root := filepath.Join(parent, "root")
	os.MkdirAll(root, 0700)
	os.MkdirAll(filepath.Join(out, "root"), 0700)
	os.WriteFile(filepath.Join(root, "victim"), []byte("inside"), 0600)
	atomicWriteTestHook = func(string) error {
		if e := os.Rename(parent, old); e != nil {
			return e
		}
		return os.Symlink(out, parent)
	}
	defer func() { atomicWriteTestHook = nil }()
	_, e := recordChange(root, "victim", []byte("next"))
	if e != nil {
		t.Fatal(e)
	}
	entries, e := os.ReadDir(filepath.Join(old, "root", changeHistoryDir))
	if e != nil || len(entries) != 1 {
		t.Fatalf("pinned record missing %v %v", entries, e)
	}
	if _, e := os.Stat(filepath.Join(out, "root", changeHistoryDir)); !os.IsNotExist(e) {
		t.Fatal("history escaped")
	}
}
func TestR13ReviewDeskHistoryAncestorSwap(t *testing.T) {
	p := t.TempDir()
	parent := filepath.Join(p, "parent")
	old := filepath.Join(p, "old")
	out := filepath.Join(p, "outside")
	root := filepath.Join(parent, "root")
	os.MkdirAll(root, 0700)
	os.MkdirAll(filepath.Join(out, "root"), 0700)
	os.WriteFile(filepath.Join(root, "victim"), []byte("inside"), 0600)
	os.WriteFile(filepath.Join(out, "root", "victim"), []byte("OUTSIDE_SENTINEL"), 0600)
	var stop atomic.Bool
	var swaps atomic.Int64
	done := make(chan struct{})
	go func() {
		defer close(done)
		for !stop.Load() {
			if os.Rename(parent, old) != nil {
				continue
			}
			if os.Symlink(out, parent) == nil {
				swaps.Add(1)
				time.Sleep(time.Microsecond)
				os.Remove(parent)
			}
			os.Rename(old, parent)
		}
	}()
	s := newView("q8")
	s.root = root
	success := 0
	for end := time.Now().Add(2 * time.Second); time.Now().Before(end); {
		text, ok, _ := s.localText("victim")
		if ok {
			success++
			if strings.Contains(text, "OUTSIDE_SENTINEL") {
				stop.Store(true)
				<-done
				t.Fatal("desk read outside during ancestor swap")
			}
		}
	}
	stop.Store(true)
	<-done
	t.Logf("desk success=%d swaps=%d", success, swaps.Load())
	if swaps.Load() == 0 {
		t.Fatal("vacuous")
	}
}
