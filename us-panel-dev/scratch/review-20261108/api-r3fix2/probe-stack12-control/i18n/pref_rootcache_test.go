package i18n

import (
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

func liveRoots() int64 { return rootOpened.Load() - rootClosed.Load() }

// cycleRoots runs Set+Get on n distinct roots below base with one store.
func cycleRoots(t *testing.T, p *PrefStore, set func(string), base string, from, n int) {
	t.Helper()
	for i := from; i < from+n; i++ {
		set(filepath.Join(base, "r"+strconv.Itoa(i)))
		if err := p.Set("en"); err != nil {
			t.Fatal(err)
		}
		if p.Get() != "en" {
			t.Fatal("round trip", i)
		}
	}
}

func newSwitchStore() (*PrefStore, func(string)) {
	var mu sync.Mutex
	cur := ""
	p := &PrefStore{Root: func() string { mu.Lock(); defer mu.Unlock(); return cur }}
	return p, func(s string) { mu.Lock(); cur = s; mu.Unlock() }
}

// TestPrefRootCacheBounded reproduces PANEL DEV R3-01 (64 roots: 7 -> 71
// fds). Direct measurement: the internal open/close counters must show at
// most maxRoots live root handles, and zero after Close.
func TestPrefRootCacheBounded(t *testing.T) {
	base := t.TempDir()
	p, set := newSwitchStore()
	live0 := liveRoots()
	cycleRoots(t, p, set, base, 0, 64)
	if d := liveRoots() - live0; d > maxRoots {
		t.Fatalf("live root handles after 64 roots: %d (> %d)", d, maxRoots)
	}
	p.mu.Lock()
	n := len(p.roots)
	p.mu.Unlock()
	if n > maxRoots {
		t.Fatalf("cache size %d > %d", n, maxRoots)
	}
	p.Close()
	if d := liveRoots() - live0; d != 0 {
		t.Fatalf("Close left %d root handles open", d)
	}
}

// TestPrefRootHandleScaling: OS handle growth must not scale with the
// number of roots. Absolute counts are noisy (runtime threads, IOCP,
// timers, AV/EDR); a leak of one handle per root would add >= 448 between
// the 64- and 512-root runs, so a 32-handle slack cleanly separates them.
func TestPrefRootHandleScaling(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	base := t.TempDir()
	p, set := newSwitchStore()
	defer p.Close()
	cycleRoots(t, p, set, base, 0, 32) // warm-up: lazy runtime/OS handles
	measure := func(from, n int) int {
		runtime.GC()
		b := openHandleCount(t)
		cycleRoots(t, p, set, base, from, n)
		runtime.GC()
		return openHandleCount(t) - b
	}
	d64 := measure(1000, 64)
	d512 := measure(2000, 512)
	t.Logf("handle delta: 64 roots %+d, 512 roots %+d", d64, d512)
	if d512-d64 > 32 {
		t.Fatalf("handle growth scales with roots: 64 -> %+d, 512 -> %+d", d64, d512)
	}
}

// TestPrefRootClosedHandleInvalid: every evicted root handle is really
// closed at the OS level right after close (fcntl/GetHandleInformation).
func TestPrefRootClosedHandleInvalid(t *testing.T) {
	var closed []rootHandle
	var bad []string
	rootCloseHook = func(h rootHandle) {
		closed = append(closed, h)
		if handleValid(h) {
			bad = append(bad, "still valid after close")
		}
	}
	defer func() { rootCloseHook = nil }()
	base := t.TempDir()
	p, set := newSwitchStore()
	cycleRoots(t, p, set, base, 0, 16)
	p.Close()
	if len(closed) != 16 {
		t.Fatalf("closed %d root handles, want 16", len(closed))
	}
	if len(bad) > 0 {
		t.Fatal(bad)
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
	if !handleValid(ra.h) {
		t.Fatal("in-use root handle invalid")
	}
	p.release(ra)
	if !ra.closed {
		t.Fatal("evicted root not closed after release")
	}
	cur = a
	if err := p.Set("en"); err != nil || p.Get() != "en" {
		t.Fatal("re-open after eviction failed", err)
	}
	p.Close()
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
	live0 := liveRoots()
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
	if d := liveRoots() - live0; d != 0 {
		t.Fatalf("concurrent switching left %d root handles open", d)
	}
}
