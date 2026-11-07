//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

// Directory junctions work without Developer Mode or elevated privileges on
// local NTFS. A failed junction operation is a FAIL, not a silent skip.
func r13Junction(t *testing.T, link, target string) {
	t.Helper()
	if out, err := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
		t.Fatalf("junction: %v: %s", err, out)
	}
}
func r13Tools(t *testing.T, root string) *fileTools {
	t.Helper()
	ft, err := newFileTools(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ft.Close() })
	return ft
}
func TestR13WindowsRelativeLifecycle(t *testing.T) {
	root := t.TempDir()
	ft := r13Tools(t, root)
	for _, body := range []string{"first", "replacement"} {
		if err := ft.Write("nested/file.txt", []byte(body)); err != nil {
			t.Fatal(err)
		}
		got, err := ft.Read("nested/file.txt")
		if err != nil || string(got) != body {
			t.Fatalf("read %q %v", got, err)
		}
	}
	if err := ft.Delete("nested/file.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "nested", "file.txt")); !os.IsNotExist(err) {
		t.Fatalf("delete: %v", err)
	}
}
func TestR13WindowsUnsafeNames(t *testing.T) {
	ft := r13Tools(t, t.TempDir())
	for _, name := range []string{"x:stream", "CON.txt", "nul", "com1.txt", "LPT9", "COM¹.txt", "LPT²", "dir/PRN", "a.", "a ", "LONGFI~1.TXT", "../outside", `\\?\C:\x`} {
		t.Run(name, func(t *testing.T) {
			if err := ft.Write(name, []byte("bad")); err == nil {
				t.Fatal("unsafe write accepted")
			}
			if _, err := ft.Read(name); err == nil {
				t.Fatal("unsafe read accepted")
			}
			if err := ft.Delete(name); err == nil {
				t.Fatal("unsafe delete accepted")
			}
		})
	}
}
func TestR13WindowsHardLink(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	target := filepath.Join(out, "sentinel")
	if err := os.WriteFile(target, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(target, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	ft := r13Tools(t, root)
	if err := ft.Write("link", []byte("bad")); err == nil {
		t.Fatal("hard link write accepted")
	}
	if _, err := ft.Read("link"); err == nil {
		t.Fatal("hard link read accepted")
	}
	if err := ft.Delete("link"); err == nil {
		t.Fatal("hard link delete accepted")
	}
	b, err := os.ReadFile(target)
	if err != nil || string(b) != "outside" {
		t.Fatalf("outside %q %v", b, err)
	}
}
func TestR13WindowsJunctionRejected(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	link := filepath.Join(root, "link")
	r13Junction(t, link, out)
	t.Cleanup(func() { os.Remove(link) })
	ft := r13Tools(t, root)
	if err := os.WriteFile(filepath.Join(out, "sentinel"), []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := pinRoot(link); err == nil {
		t.Fatal("junction root accepted")
	}
	if _, err := ft.Read("link/sentinel"); err == nil {
		t.Fatal("junction read accepted")
	}
	if err := ft.Write("link/sentinel", []byte("bad")); err == nil {
		t.Fatal("junction write accepted")
	}
	if err := ft.Delete("link/sentinel"); err == nil {
		t.Fatal("junction delete accepted")
	}
	b, err := os.ReadFile(filepath.Join(out, "sentinel"))
	if err != nil || string(b) != "outside" {
		t.Fatalf("outside %q %v", b, err)
	}
}
func TestR13WindowsRootSwapPinned(t *testing.T) {
	p := t.TempDir()
	root, old, out := filepath.Join(p, "root"), filepath.Join(p, "old"), filepath.Join(p, "outside")
	for _, dir := range []string{root, out} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(root, "victim"), []byte("inside"), 0600)
	os.WriteFile(filepath.Join(out, "victim"), []byte("outside"), 0600)
	ft := r13Tools(t, root)
	if err := os.Rename(root, old); err != nil {
		t.Fatal(err)
	}
	r13Junction(t, root, out)
	t.Cleanup(func() { ft.Close(); os.Remove(root) })
	b, err := ft.Read("victim")
	if err != nil || string(b) != "inside" {
		t.Fatalf("pinned read %q %v", b, err)
	}
	if err := ft.Write("written", []byte("inside")); err != nil {
		t.Fatal(err)
	}
	if err := ft.Delete("victim"); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(old, "written")); err != nil || string(b) != "inside" {
		t.Fatalf("pinned write %q %v", b, err)
	}
	if _, err := os.Stat(filepath.Join(out, "written")); !os.IsNotExist(err) {
		t.Fatal("write escaped")
	}
	if b, err := os.ReadFile(filepath.Join(out, "victim")); err != nil || string(b) != "outside" {
		t.Fatal("delete escaped")
	}
}
func TestR13WindowsRootSwapStress(t *testing.T) {
	for _, op := range []string{"write", "delete"} {
		t.Run(op, func(t *testing.T) {
			p := t.TempDir()
			root, old, out, link := filepath.Join(p, "root"), filepath.Join(p, "old"), filepath.Join(p, "out"), filepath.Join(p, "junction")
			os.Mkdir(root, 0700)
			os.Mkdir(out, 0700)
			os.WriteFile(filepath.Join(out, "victim"), []byte("outside"), 0600)
			r13Junction(t, link, out)
			ft := r13Tools(t, root)
			var stop atomic.Bool
			var swaps atomic.Int64
			done := make(chan struct{})
			go func() {
				defer close(done)
				for !stop.Load() {
					if os.Rename(root, old) != nil {
						continue
					}
					if os.Rename(link, root) != nil {
						os.Rename(old, root)
						continue
					}
					swaps.Add(1)
					time.Sleep(time.Millisecond)
					deadline := time.Now().Add(2 * time.Second)
					for os.Rename(root, link) != nil {
						if time.Now().After(deadline) {
							return
						}
						time.Sleep(time.Millisecond)
					}
					for os.Rename(old, root) != nil {
						if time.Now().After(deadline) {
							return
						}
						time.Sleep(time.Millisecond)
					}
				}
			}()
			success := 0
			for end := time.Now().Add(3 * time.Second); time.Now().Before(end); {
				if op == "write" {
					if ft.Write("written", []byte("probe")) == nil {
						success++
					}
				} else {
					if ft.Write("victim", []byte("inside")) == nil && ft.Delete("victim") == nil {
						success++
					}
				}
			}
			stop.Store(true)
			<-done
			if info, err := os.Lstat(root); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				t.Fatal("swap loop did not restore the original root")
			}
			ft.Close()
			os.Remove(link)
			if swaps.Load() == 0 || success == 0 {
				t.Fatalf("vacuous test swaps=%d success=%d", swaps.Load(), success)
			}
			if _, err := os.Stat(filepath.Join(out, "written")); !os.IsNotExist(err) {
				t.Fatal("write escaped")
			}
			if b, err := os.ReadFile(filepath.Join(out, "victim")); err != nil || string(b) != "outside" {
				t.Fatal("delete escaped")
			}
			t.Logf("swaps=%d successful_operations=%d", swaps.Load(), success)
		})
	}
}
// TestR13WindowsParentSwapAtomicHelper moved to file_tools_r13b_windows_test.go
// (r13b): the r13 version accepted any error as PASS (PANEL DEV R13-03).
func TestR13WindowsShortAlias(t *testing.T) {
	root := t.TempDir()
	long := filepath.Join(root, "A quite long filename for alias.txt")
	os.WriteFile(long, []byte("inside"), 0600)
	p, _ := windows.UTF16PtrFromString(long)
	buf := make([]uint16, 32768)
	n, err := windows.GetShortPathName(p, &buf[0], uint32(len(buf)))
	if err != nil {
		t.Fatal(err)
	}
	short := filepath.Base(windows.UTF16ToString(buf[:n]))
	t.Logf("short alias=%q tilde=%v", short, strings.Contains(short, "~"))
	if strings.EqualFold(short, filepath.Base(long)) {
		t.Skip("8.3 alias üretimi bu diskte kapalı")
	}
	ft := r13Tools(t, root)
	if _, err := ft.Read(short); err == nil {
		t.Fatal("8.3 alias accepted")
	}
	if err := ft.Write(short, []byte("bad")); err == nil {
		t.Fatal("8.3 alias write accepted")
	}
}

func TestR13WindowsDeskAndHistory(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "nested", "file.txt")
	if _, err := recordChange(root, "nested/file.txt", []byte("new")); err != nil {
		t.Fatal(err)
	}
	if err := writeModelFile(root, target, []byte("new")); err != nil {
		t.Fatal(err)
	}
	s := newView("q8")
	s.setRoot(root)
	defer s.rootPin.Close()
	if text, ok, complete := s.localText("nested/file.txt"); !ok || !complete || text != "new" {
		t.Fatalf("desk: %q %v %v", text, ok, complete)
	}
	if _, err := undoLastChange(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("new file undo failed")
	}
}
