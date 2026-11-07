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

// Adapted from PANEL DEV r9 PoC (zz_r9_bypass_test.go B2/B3): directory
// "d" flips between a real dir and a symlink to outside.
func startDirSwap(root, outside string) func() {
	real, d := filepath.Join(root, "realdir"), filepath.Join(root, "d")
	var stop atomic.Bool
	done := make(chan struct{})
	go func() {
		defer close(done)
		for !stop.Load() {
			os.Remove(d)
			os.Rename(real, d)
			os.Rename(d, real)
			os.Symlink(outside, d)
		}
	}()
	return func() { stop.Store(true); <-done }
}

func outsideNames(dir string) []string {
	ents, _ := os.ReadDir(dir)
	var n []string
	for _, e := range ents {
		n = append(n, e.Name())
	}
	return n
}

func TestReviewWriteRaceNoEscapeNoTempLeak(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(root, "realdir"), 0o755)
	ft, _ := newFileTools(root)
	stop := startDirSwap(root, outside)
	ok := 0
	for end := time.Now().Add(10 * time.Second); time.Now().Before(end); {
		if ft.Write("d/w.txt", []byte("WRITTEN")) == nil {
			ok++
		}
		if len(outsideNames(outside)) > 0 {
			break
		}
	}
	stop()
	if n := outsideNames(outside); len(n) > 0 {
		t.Fatalf("kök dışına yazıldı/geçici dosya sızdı (başarılı=%d): %v", ok, n)
	}
	// No leftover temp files inside the root either.
	filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err == nil && strings.HasPrefix(filepath.Base(p), ".us-write-") {
			t.Errorf("geçici dosya kaldı: %s", p)
		}
		return nil
	})
}

func TestReviewDeleteRaceNoEscape(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(root, "realdir"), 0o755)
	ft, _ := newFileTools(root)
	stop := startDirSwap(root, outside)
	deleted := false
	for end := time.Now().Add(3 * time.Second); time.Now().Before(end) && !deleted; {
		os.WriteFile(filepath.Join(outside, "v.txt"), []byte("victim"), 0o600)
		os.WriteFile(filepath.Join(root, "realdir", "v.txt"), []byte("x"), 0o644)
		_ = ft.Delete("d/v.txt")
		if _, err := os.Stat(filepath.Join(outside, "v.txt")); err != nil {
			deleted = true
		}
	}
	stop()
	if deleted {
		t.Fatal("kök dışındaki dosya silindi")
	}
}

func TestReviewWriteFailureCleansTemp(t *testing.T) {
	root := t.TempDir()
	ft, _ := newFileTools(root)
	atomicWriteTestHook = func(string) error { return os.ErrDeadlineExceeded }
	defer func() { atomicWriteTestHook = nil }()
	if ft.Write("sub/x.txt", []byte("a")) == nil {
		t.Fatal("hook hatası yutuldu")
	}
	ents, _ := os.ReadDir(filepath.Join(root, "sub"))
	if len(ents) != 0 {
		t.Fatalf("hata sonrası geçici dosya kaldı: %v", outsideNames(filepath.Join(root, "sub")))
	}
}
