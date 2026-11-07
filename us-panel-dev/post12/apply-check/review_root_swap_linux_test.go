//go:build linux

package main

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// R12 (PANEL DEV R11-01): the workspace root itself is swapped for a symlink
// to an outside directory with RENAME_EXCHANGE in a tight loop while
// Write/Delete run. The root fd pinned in newFileTools must keep every
// operation inside the original directory.
func TestPanelRootSwapWriteDelete(t *testing.T) {
	for _, op := range []string{"write", "delete"} {
		t.Run(op, func(t *testing.T) {
			p := t.TempDir()
			root, outside, link := filepath.Join(p, "root"), filepath.Join(p, "outside"), filepath.Join(p, "link")
			os.Mkdir(root, 0o700)
			os.Mkdir(outside, 0o700)
			os.Symlink(outside, link)
			os.WriteFile(filepath.Join(root, "victim"), []byte("inside"), 0o600)
			os.WriteFile(filepath.Join(outside, "victim"), []byte("outside"), 0o600)
			ft, err := newFileTools(root)
			if err != nil {
				t.Fatal(err)
			}
			var stop atomic.Bool
			var swaps atomic.Int64
			done := make(chan struct{})
			go func() {
				defer close(done)
				for !stop.Load() {
					if unix.Renameat2(unix.AT_FDCWD, root, unix.AT_FDCWD, link, unix.RENAME_EXCHANGE) == nil {
						swaps.Add(1)
					}
				}
			}()
			attempts, success := 0, 0
			for end := time.Now().Add(4 * time.Second); time.Now().Before(end); {
				attempts++
				var e error
				if op == "write" {
					e = ft.Write("written", []byte("probe"))
				} else {
					os.WriteFile(filepath.Join(outside, "victim"), []byte("outside"), 0o600)
					e = ft.Delete("victim")
				}
				if e == nil {
					success++
				}
				if op == "write" {
					if _, e := os.Lstat(filepath.Join(outside, "written")); e == nil {
						stop.Store(true)
						<-done
						t.Fatalf("root-swap write escaped after %d attempts", attempts)
					}
				} else if b, e := os.ReadFile(filepath.Join(outside, "victim")); e != nil || string(b) != "outside" {
					// outside may be reached via the swapped name; check by content of the real outside dir
					if _, e2 := os.Lstat(filepath.Join(outside, "victim")); os.IsNotExist(e2) {
						stop.Store(true)
						<-done
						t.Fatalf("root-swap delete escaped after %d attempts", attempts)
					}
				}
			}
			stop.Store(true)
			<-done
			t.Logf("attempts=%d success=%d swaps=%d", attempts, success, swaps.Load())
			if swaps.Load() == 0 {
				t.Fatal("swap never happened; test is not exercising the race")
			}
		})
	}
}
