//go:build linux

package privacy

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestV4PrefsFIFOOpenedFDRejectedWithoutBlocking(t *testing.T) {
	p := filepath.Join(t.TempDir(), "fifo")
	if err := unix.Mkfifo(p, 0600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		f, e := openPrefsNoFollow(p)
		if f != nil {
			f.Close()
		}
		done <- e
	}()
	select {
	case e := <-done:
		if e == nil {
			t.Fatal("FIFO accepted")
		}
	case <-time.After(time.Second):
		t.Fatal("FIFO open blocked")
	}
}

func TestV4PrefsParentSymlinkReadAndWriteRejected(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(out, "p.json"), []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(out, link); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenPrefStore(filepath.Join(link, "p.json")); err == nil {
		t.Fatal("parent read escaped")
	}
	// Constructed in-memory store models a swap after startup as well.
	ps := &PrefStore{path: filepath.Join(link, "new.json"), byID: map[string]Preferences{}}
	if err := ps.Put(Preferences{AccountID: "a", Source: "sentinel"}, time.Now()); err == nil {
		t.Fatal("parent write escaped")
	}
	if _, err := os.Stat(filepath.Join(out, "new.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("outside write: %v", err)
	}
}

func TestV4PrefsPinnedDirectorySurvivesSwap(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	dir, name, err := prefsParent(filepath.Join(parent, "p.json"), false)
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()
	moved := filepath.Join(root, "moved")
	if err := os.Rename(parent, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(out, parent); err != nil {
		t.Fatal(err)
	}
	if err := writePrefsAt(dir, name, []byte(`{}`), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, name)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("escaped pinned fd")
	}
	if _, err := os.Stat(filepath.Join(moved, name)); err != nil {
		t.Fatal(err)
	}
}

func TestV4PrefsLeafSwapRaceNeverReadsOutside(t *testing.T) {
	dir, out := t.TempDir(), t.TempDir()
	path := filepath.Join(dir, "p.json")
	external := filepath.Join(out, "p.json")
	if err := os.WriteFile(external, []byte(`outside`), 0600); err != nil {
		t.Fatal(err)
	}
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			_ = os.Remove(path)
			_ = os.Symlink(external, path)
			_ = os.Remove(path)
			_ = os.WriteFile(filepath.Join(dir, "next"), []byte(`inside`), 0600)
			_ = os.Rename(filepath.Join(dir, "next"), path)
		}
	}()
	defer func() { close(stop); wg.Wait() }()
	for i := 0; i < 500; i++ {
		f, err := openPrefsNoFollow(path)
		if err != nil {
			continue
		}
		b := make([]byte, 32)
		n, _ := f.Read(b)
		f.Close()
		if string(b[:n]) == "outside" {
			t.Fatal("followed swapped leaf")
		}
	}
}

func TestV4PrefsFaultsCleanTempsAndRollback(t *testing.T) {
	for _, stage := range []string{"write", "sync", "rename", "actual-rename"} {
		t.Run(stage, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "p.json")
			ps, err := OpenPrefStore(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := ps.Put(Preferences{AccountID: "a", Source: "old"}, time.Now()); err != nil {
				t.Fatal(err)
			}
			if stage == "actual-rename" {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			} else {
				ps.writeFault = func(at string) error {
					if at == stage {
						return errors.New("injected")
					}
					return nil
				}
			}
			if err := ps.Put(Preferences{AccountID: "a", Source: "sentinel"}, time.Now()); err == nil {
				t.Fatal("fault not returned")
			}
			if ps.Get("a").Source != "old" {
				t.Fatal("memory rollback failed")
			}
			names, err := filepath.Glob(filepath.Join(dir, ".privacy-prefs-*"))
			if err != nil || len(names) != 0 {
				t.Fatalf("temp leftovers: %v %v", names, err)
			}
			if stage != "actual-rename" {
				reopen, err := OpenPrefStore(path)
				if err != nil || reopen.Get("a").Source != "old" {
					t.Fatal("disk modified before commit")
				}
			}
		})
	}
}

func TestV4PrefsDirectorySyncFailureIsExplicitCommittedState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "p.json")
	ps, err := OpenPrefStore(path)
	if err != nil {
		t.Fatal(err)
	}
	ps.writeFault = func(stage string) error {
		if stage == "dirsync" {
			return errors.New("injected directory sync failure")
		}
		return nil
	}
	err = ps.Put(Preferences{AccountID: "a", Source: "committed"}, time.Now())
	var committed *prefsCommittedError
	if !errors.As(err, &committed) {
		t.Fatal("directory sync failure not reported")
	}
	reopened, err := OpenPrefStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if ps.Get("a").Source != "committed" || reopened.Get("a").Source != "committed" {
		t.Fatal("visible disk/memory diverged")
	}
	names, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".privacy-prefs-*"))
	if len(names) != 0 {
		t.Fatal("temp left")
	}
	ps.writeFault = nil
	if err := ps.Delete("a"); err != nil {
		t.Fatal(err)
	}
}
