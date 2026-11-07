package i18n

import (
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

// TestPrefRootCacheBounded reproduces PANEL DEV R3-01: 64 distinct roots
// on one store must not grow the open handle count (was 7 -> 71).
func TestPrefRootCacheBounded(t *testing.T) {
	base := t.TempDir()
	cur := ""
	var mu sync.Mutex
	p := &PrefStore{Root: func() string { mu.Lock(); defer mu.Unlock(); return cur }}
	set := func(i int) {
		mu.Lock()
		cur = filepath.Join(base, "r"+strconv.Itoa(i))
		mu.Unlock()
	}
	set(0)
	if err := p.Set("en"); err != nil {
		t.Fatal(err)
	}
	runtime.GC()
	before := openHandleCount(t)
	for i := 1; i <= 64; i++ {
		set(i)
		if err := p.Set("en"); err != nil {
			t.Fatal(err)
		}
		if p.Get() != "en" {
			t.Fatal("round trip", i)
		}
	}
	runtime.GC()
	after := openHandleCount(t)
	if after-before > 16 { // leak would add ~64; slack for runtime threads
		t.Fatalf("handle growth %d -> %d after 64 roots", before, after)
	}
	p.mu.Lock()
	n := len(p.roots)
	p.mu.Unlock()
	if n > maxRoots {
		t.Fatalf("cache size %d > %d", n, maxRoots)
	}
	p.Close()
	if c := openHandleCount(t); c > before {
		t.Fatalf("Close left handles: %d > %d", c, before)
	}
}

// TestPrefRootEvictionWhileInUse: a root evicted during an in-flight call
// stays open until that call releases it, then is closed exactly once.
func TestPrefRootEvictionWhileInUse(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	cur := a
	p := &PrefStore{Root: func() string { return cur }}
	_, ra := p.root() // simulate in-flight use of a
	if ra.err != nil {
		t.Fatal(ra.err)
	}
	cur = b
	_, rb := p.root() // evicts a
	p.release(rb)
	if !ra.evicted || ra.closed {
		t.Fatal("in-use root closed early")
	}
	if _, err := readPrefFile(ra.h, []string{"lang.json"}); err == nil {
		// file absent: an error is expected, but the handle must be valid
	}
	p.release(ra)
	if !ra.closed {
		t.Fatal("evicted root not closed after release")
	}
	cur = a
	if err := p.Set("en"); err != nil || p.Get() != "en" {
		t.Fatal("re-open after eviction failed", err)
	}
}

// TestPrefRootConcurrentSwitch: concurrent Get/Set while Root flips.
func TestPrefRootConcurrentSwitch(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	var mu sync.Mutex
	flip := false
	p := &PrefStore{Root: func() string {
		mu.Lock()
		defer mu.Unlock()
		flip = !flip
		if flip {
			return a
		}
		return b
	}}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				_ = p.Set("en")
				_ = p.Get()
			}
		}()
	}
	wg.Wait()
	p.Close()
}
