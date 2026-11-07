//go:build windows

package main

// r13b native Windows cases (PANEL DEV R13-01/02/03 and missing-case list).
// Rules: a setup step that does not happen is a FAIL (barrier), never a
// silent PASS; a case that needs admin/privilege/extra accounts is SKIPPED
// with a "SKIP-<REASON>:" line, never reported as PASS.

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Set at build time: go test -c -ldflags "-X kiyas.r13bPatchSHA=<sha256>".
var r13bPatchSHA = "unset"

var (
	r13bKernel32            = windows.NewLazySystemDLL("kernel32.dll")
	r13bGetProcessHandleCnt = r13bKernel32.NewProc("GetProcessHandleCount")
	r13bSetFileShortNameW   = r13bKernel32.NewProc("SetFileShortNameW")
)

func r13bElevated() bool { return windows.GetCurrentProcessToken().IsElevated() }

// Runs first in this file (go runs tests in file order): ties the log to the
// exact binary and patch.
func TestR13bAAMetadata(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(exe)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	_, err = io.Copy(h, f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	host, _ := os.Hostname()
	v := windows.RtlGetVersion()
	user := "?"
	if u, e := windows.GetCurrentProcessToken().GetTokenUser(); e == nil {
		user = u.User.Sid.String()
	}
	t.Logf("R13B_META binary_sha256=%s", hex.EncodeToString(h.Sum(nil)))
	t.Logf("R13B_META binary_name=%s", filepath.Base(exe))
	t.Logf("R13B_META patch_sha256=%s", r13bPatchSHA)
	t.Logf("R13B_META hostname=%s", host)
	t.Logf("R13B_META go_toolchain=%s goos=%s goarch=%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	t.Logf("R13B_META elevated=%v user_sid=%s", r13bElevated(), user)
	t.Logf("R13B_META os_version=%d.%d.%d sp=%d product_type=%d", v.MajorVersion, v.MinorVersion, v.BuildNumber, v.ServicePackMajor, v.ProductType)
	if r13bPatchSHA == "unset" || len(r13bPatchSHA) != 64 {
		t.Fatal("binary built without -X r13bPatchSHA; cannot be tied to the patch")
	}
}

// ---------- helpers ----------

func r13bErrno(err error) (syscall.Errno, bool) {
	var e syscall.Errno
	if errors.As(err, &e) {
		return e, true
	}
	return 0, false
}

func r13bNTStatus(err error) (windows.NTStatus, bool) {
	var s windows.NTStatus
	if errors.As(err, &s) {
		return s, true
	}
	return 0, false
}

func r13bDescribe(err error) string {
	if err == nil {
		return "nil"
	}
	if e, ok := r13bErrno(err); ok {
		return fmt.Sprintf("errno=%d(0x%X) %v", uint32(e), uint32(e), err)
	}
	if s, ok := r13bNTStatus(err); ok {
		return fmt.Sprintf("ntstatus=0x%08X %v", uint32(s), err)
	}
	return fmt.Sprintf("go-error=%T %v", err, err)
}

func r13bWrite(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o600); err != nil {
		t.Fatal(err)
	}
}

func r13bRead(p string) string { b, _ := os.ReadFile(p); return string(b) }

// r13bSwapDir renames dir away and puts a junction to target in its place.
// Any failure is a test FAILURE (barrier), and the swap is proven by reading
// probeRel through the ambient path afterwards.
func r13bSwapDir(t *testing.T, dir, moved, target, probeRel, probeWant string) {
	t.Helper()
	if err := os.Rename(dir, moved); err != nil {
		t.Fatalf("BARRIER: swap rename failed (swap NOT performed): %s", r13bDescribe(err))
	}
	r13Junction(t, dir, target)
	t.Cleanup(func() { os.Remove(dir) })
	if got := r13bRead(filepath.Join(dir, probeRel)); got != probeWant {
		t.Fatalf("BARRIER: swap not effective, ambient read %q", got)
	}
	t.Logf("BARRIER: swap performed %s -> junction(%s)", dir, target)
}

func r13bNoTemp(t *testing.T, dir string) {
	t.Helper()
	ents, _ := os.ReadDir(dir)
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), ".us-write-") {
			t.Fatalf("temp residue left: %s", e.Name())
		}
	}
}

func r13bSession(t *testing.T, root string) *session {
	t.Helper()
	s := &session{}
	s.setRoot(root)
	if s.rootPin == nil {
		t.Fatal("session root not pinned")
	}
	t.Cleanup(func() { s.rootPin.Close() })
	return s
}

// ---------- R13-03: parent swap helper, rewritten ----------

// Control: no hook, then a no-op hook; the exact existing nested target must
// be written, read back, and undo must restore the previous content.
func TestR13bWindowsParentHelperControl(t *testing.T) {
	root := t.TempDir()
	victim := filepath.Join(root, "nested", "victim")
	r13bWrite(t, victim, "before")
	if err := writeModelFile(root, victim, []byte("plain")); err != nil {
		t.Fatalf("control without hook: %s", r13bDescribe(err))
	}
	if got := r13bRead(victim); got != "plain" {
		t.Fatalf("control readback %q", got)
	}
	called := false
	atomicWriteTestHook = func(string) error { called = true; return nil }
	defer func() { atomicWriteTestHook = nil }()
	if _, err := recordChange(root, "nested/victim", []byte("hooked")); err != nil {
		t.Fatal(err)
	}
	called = false
	if err := writeModelFile(root, victim, []byte("hooked")); err != nil {
		t.Fatalf("control with no-op hook: %s", r13bDescribe(err))
	}
	if !called {
		t.Fatal("no-op hook was not entered")
	}
	if got := r13bRead(victim); got != "hooked" {
		t.Fatalf("no-op hook readback %q", got)
	}
	atomicWriteTestHook = nil
	if _, err := undoLastChange(root); err != nil {
		t.Fatalf("undo: %s", r13bDescribe(err))
	}
	if got := r13bRead(victim); got != "plain" {
		t.Fatalf("undo restored %q", got)
	}
}

// Swap with NO handles open inside the parent (between calls, root already
// pinned): rename must succeed (barrier), junction must be live (barrier),
// then the pinned write/read/delete must be refused with errFileToolEscape.
func TestR13bWindowsParentSwapPinnedRoot(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	parent := filepath.Join(root, "nested")
	r13bWrite(t, filepath.Join(parent, "victim"), "inside")
	r13bWrite(t, filepath.Join(out, "victim"), "outside")
	ft := r13Tools(t, root)
	r13bSwapDir(t, parent, filepath.Join(root, "old"), out, "victim", "outside")
	err := ft.Write("nested/victim", []byte("bad"))
	if !errors.Is(err, errFileToolEscape) {
		t.Fatalf("write through swapped parent: want errFileToolEscape, got %s", r13bDescribe(err))
	}
	if _, err := ft.Read("nested/victim"); !errors.Is(err, errFileToolEscape) {
		t.Fatalf("read: want errFileToolEscape, got %s", r13bDescribe(err))
	}
	if err := ft.Delete("nested/victim"); !errors.Is(err, errFileToolEscape) {
		t.Fatalf("delete: want errFileToolEscape, got %s", r13bDescribe(err))
	}
	if r13bRead(filepath.Join(out, "victim")) != "outside" || r13bRead(filepath.Join(root, "old", "victim")) != "inside" {
		t.Fatal("content changed")
	}
	r13bNoTemp(t, out)
}

// In-window swap (temp handle open inside the parent). Two explicit outcomes:
//
//	SWAPPED: rename+junction succeeded -> write must land in the pinned dir.
//	REFUSED: rename failed with ERROR_ACCESS_DENIED(5) or
//	         ERROR_SHARING_VIOLATION(32) -> control proves the same rename
//	         succeeds once the helper has returned (refusal caused by the
//	         open handle, not by a broken test).
//
// Any other error, or a hook that never ran, is a FAIL.
func TestR13WindowsParentSwapAtomicHelper(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	parent := filepath.Join(root, "nested")
	old := filepath.Join(root, "old")
	r13bWrite(t, filepath.Join(parent, "victim"), "before")
	r13bWrite(t, filepath.Join(out, "victim"), "outside")
	var entered, renamed, junctioned bool
	var renameErr error
	atomicWriteTestHook = func(string) error {
		entered = true
		if renameErr = os.Rename(parent, old); renameErr != nil {
			return renameErr
		}
		renamed = true
		if o, err := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", parent, out).CombinedOutput(); err != nil {
			return fmt.Errorf("junction: %v %s", err, o)
		}
		junctioned = true
		return nil
	}
	err := writeModelFile(root, filepath.Join(parent, "victim"), []byte("inside"))
	atomicWriteTestHook = nil
	t.Cleanup(func() {
		if junctioned {
			os.Remove(parent)
		}
	})
	t.Logf("hook entered=%v renamed=%v junction=%v rename_err=%s helper_err=%s", entered, renamed, junctioned, r13bDescribe(renameErr), r13bDescribe(err))
	if !entered {
		t.Fatal("BARRIER: hook never entered")
	}
	if r13bRead(filepath.Join(out, "victim")) != "outside" {
		t.Fatal("parent swap escaped")
	}
	switch {
	case junctioned:
		if err != nil {
			t.Fatalf("SWAPPED: pinned write failed: %s", r13bDescribe(err))
		}
		if got := r13bRead(filepath.Join(old, "victim")); got != "inside" {
			t.Fatalf("SWAPPED: pinned dir content %q", got)
		}
		t.Log("OUTCOME=SWAPPED (swap performed in window; write stayed pinned)")
	case !renamed:
		e, ok := r13bErrno(renameErr)
		if !ok || (e != windows.ERROR_ACCESS_DENIED && e != windows.ERROR_SHARING_VIOLATION) {
			t.Fatalf("REFUSED with unexpected error: %s", r13bDescribe(renameErr))
		}
		if !errors.Is(err, e) {
			t.Fatalf("helper did not surface the hook error: %s", r13bDescribe(err))
		}
		if got := r13bRead(filepath.Join(parent, "victim")); got != "before" {
			t.Fatalf("failed write changed target: %q", got)
		}
		r13bNoTemp(t, parent)
		// Control: with no handle open, the identical rename must work.
		if cerr := os.Rename(parent, old); cerr != nil {
			t.Fatalf("CONTROL: rename without open handle failed too (test broken): %s", r13bDescribe(cerr))
		}
		os.Rename(old, parent)
		t.Logf("OUTCOME=REFUSED errno=%d: Windows refuses renaming a directory with an open child handle; swap in this window is impossible (not a swap test, see TestR13bWindowsParentSwapPinnedRoot)", uint32(e))
	default:
		t.Fatalf("rename succeeded but junction failed: %s", r13bDescribe(err))
	}
}

// ---------- R13-01/02 desk + history on Windows ----------

func TestR13bWindowsDeskSubdirSwap(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	r13bWrite(t, filepath.Join(root, "sub", "victim"), "inside")
	r13bWrite(t, filepath.Join(out, "victim"), "outside")
	s := r13bSession(t, root)
	s.readFile("sub/victim")
	if s.openText != "inside" {
		t.Fatalf("initial desk read %q %q", s.openText, s.notice)
	}
	r13bSwapDir(t, filepath.Join(root, "sub"), filepath.Join(root, "sub-old"), out, "victim", "outside")
	if txt, ok, _ := s.localText("sub/victim"); ok || strings.Contains(txt, "outside") {
		t.Fatalf("desk read through junction: %q", txt)
	}
	s.saveFile("sub/victim", "saved")
	s.appendFile("sub/victim", "+tail")
	if r13bRead(filepath.Join(out, "victim")) != "outside" {
		t.Fatal("desk wrote through junction")
	}
	if _, err := os.Stat(filepath.Join(out, changeHistoryDir)); err == nil {
		t.Fatal("history escaped")
	}
	if r13bRead(filepath.Join(root, "sub-old", "victim")) != "inside" {
		t.Fatal("moved dir changed")
	}
}

// Session root itself renamed and replaced by a junction after setRoot.
func TestR13bWindowsDeskSessionRootSwap(t *testing.T) {
	base, out := t.TempDir(), t.TempDir()
	ws := filepath.Join(base, "ws")
	r13bWrite(t, filepath.Join(ws, "victim"), "inside")
	r13bWrite(t, filepath.Join(out, "victim"), "outside")
	s := r13bSession(t, ws)
	r13bSwapDir(t, ws, filepath.Join(base, "ws-moved"), out, "victim", "outside")
	s.appendFile("victim", "+x")
	if r13bRead(filepath.Join(out, "victim")) != "outside" {
		t.Fatal("append escaped through swapped root")
	}
	if got := r13bRead(filepath.Join(base, "ws-moved", "victim")); got != "inside+x" {
		t.Fatalf("pinned append %q notice %q", got, s.notice)
	}
	if _, err := s.undoLast(); err != nil {
		t.Fatalf("undo on pinned root: %s", r13bDescribe(err))
	}
	if got := r13bRead(filepath.Join(base, "ws-moved", "victim")); got != "inside" {
		t.Fatalf("undo %q", got)
	}
}

// Ancestor of the session root. Two explicit outcomes as in the helper test.
func TestR13bWindowsDeskAncestorSwap(t *testing.T) {
	base, outBase := t.TempDir(), t.TempDir()
	ws := filepath.Join(base, "parent", "ws")
	r13bWrite(t, filepath.Join(ws, "victim"), "inside")
	r13bWrite(t, filepath.Join(outBase, "ws", "victim"), "outside")
	s := r13bSession(t, ws)
	parent, old := filepath.Join(base, "parent"), filepath.Join(base, "old")
	err := os.Rename(parent, old)
	if err == nil {
		r13Junction(t, parent, outBase)
		t.Cleanup(func() { os.Remove(parent) })
		if r13bRead(filepath.Join(ws, "victim")) != "outside" {
			t.Fatal("BARRIER: swap not effective")
		}
		if txt, ok, _ := s.localText("victim"); !ok || txt != "inside" {
			t.Fatalf("SWAPPED: pinned read %q %v", txt, ok)
		}
		s.appendFile("victim", "+a")
		if r13bRead(filepath.Join(outBase, "ws", "victim")) != "outside" {
			t.Fatal("SWAPPED: append escaped")
		}
		if _, e := os.Stat(filepath.Join(outBase, "ws", changeHistoryDir)); e == nil {
			t.Fatal("SWAPPED: history escaped")
		}
		if got := r13bRead(filepath.Join(old, "ws", "victim")); got != "inside+a" {
			t.Fatalf("SWAPPED: pinned append %q", got)
		}
		t.Log("OUTCOME=SWAPPED")
		return
	}
	e, ok := r13bErrno(err)
	if !ok || (e != windows.ERROR_ACCESS_DENIED && e != windows.ERROR_SHARING_VIOLATION) {
		t.Fatalf("REFUSED with unexpected error: %s", r13bDescribe(err))
	}
	s.rootPin.Close()
	s.rootPin = nil
	if cerr := os.Rename(parent, old); cerr != nil {
		t.Fatalf("CONTROL: rename without pinned root failed too: %s", r13bDescribe(cerr))
	}
	t.Logf("OUTCOME=REFUSED errno=%d: ancestor of a pinned root cannot be renamed while the pin is held", uint32(e))
}

// Save, append and cleanup error paths: an injected write failure must
// remove the pinned history record and leave the target and dir clean.
func TestR13bWindowsDeskCleanupRace(t *testing.T) {
	root := t.TempDir()
	r13bWrite(t, filepath.Join(root, "sub", "victim"), "inside")
	s := r13bSession(t, root)
	s.readFile("sub/victim")
	injected := errors.New("injected write failure")
	seen := 0
	atomicWriteTestHook = func(tmp string) error {
		if filepath.Base(filepath.Dir(tmp)) != "sub" {
			return nil
		}
		ents, _ := os.ReadDir(filepath.Join(root, changeHistoryDir))
		if len(ents) != 1 {
			t.Errorf("BARRIER: expected 1 record, got %d", len(ents))
		}
		seen++
		return injected
	}
	defer func() { atomicWriteTestHook = nil }()
	s.saveFile("sub/victim", "saved")
	s.appendFile("sub/victim", "+tail")
	if seen != 2 {
		t.Fatalf("BARRIER: hook ran %d times", seen)
	}
	if ents, _ := os.ReadDir(filepath.Join(root, changeHistoryDir)); len(ents) != 0 {
		t.Fatalf("history record not cleaned: %d", len(ents))
	}
	if r13bRead(filepath.Join(root, "sub", "victim")) != "inside" {
		t.Fatal("target changed on failed write")
	}
	r13bNoTemp(t, filepath.Join(root, "sub"))
}

// ---------- reparse variants ----------

func TestR13bWindowsSymlinkRejected(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	r13bWrite(t, filepath.Join(out, "victim"), "outside")
	if err := os.Symlink(out, filepath.Join(root, "dlink")); err != nil {
		if e, ok := r13bErrno(err); ok && e == windows.ERROR_PRIVILEGE_NOT_HELD {
			t.Skipf("SKIP-PRIVILEGE: symlink creation needs SeCreateSymbolicLinkPrivilege (admin or Developer Mode): %s", r13bDescribe(err))
		}
		t.Fatalf("symlink setup: %s", r13bDescribe(err))
	}
	if err := os.Symlink(filepath.Join(out, "victim"), filepath.Join(root, "flink")); err != nil {
		t.Fatalf("file symlink setup: %s", r13bDescribe(err))
	}
	ft := r13Tools(t, root)
	for _, rel := range []string{"dlink/victim", "flink"} {
		if _, err := ft.Read(rel); err == nil {
			t.Fatalf("symlink read accepted: %s", rel)
		}
		if err := ft.Write(rel, []byte("bad")); err == nil {
			t.Fatalf("symlink write accepted: %s", rel)
		}
	}
	if err := ft.Delete("flink"); err == nil {
		t.Fatal("symlink delete accepted")
	}
	if r13bRead(filepath.Join(out, "victim")) != "outside" {
		t.Fatal("escaped")
	}
}

func TestR13bWindowsVolumeMountPoint(t *testing.T) {
	t.Skip("SKIP-ADMIN: a real volume mount point (SetVolumeMountPointW/mountvol) needs admin and changes the system mount table; same IO_REPARSE_TAG_MOUNT_POINT tag is covered by JunctionRejected, VolumeRootJunction and Retag")
}

// Junction to a volume root carries IO_REPARSE_TAG_MOUNT_POINT like a mount point.
func TestR13bWindowsVolumeRootJunction(t *testing.T) {
	root := t.TempDir()
	vol := filepath.VolumeName(root) + `\`
	link := filepath.Join(root, "vol")
	r13Junction(t, link, vol)
	t.Cleanup(func() { os.Remove(link) })
	tag := r13bReparseTag(t, link)
	t.Logf("reparse tag=0x%08X", tag)
	if tag != windows.IO_REPARSE_TAG_MOUNT_POINT {
		t.Fatalf("BARRIER: unexpected tag 0x%08X", tag)
	}
	ft := r13Tools(t, root)
	if _, err := ft.Read("vol/Windows/win.ini"); !errors.Is(err, errFileToolEscape) {
		t.Fatalf("volume-root junction read: %s", r13bDescribe(err))
	}
	if err := ft.Write("vol/r13b-should-not-exist.txt", []byte("bad")); !errors.Is(err, errFileToolEscape) {
		t.Fatalf("volume-root junction write: %s", r13bDescribe(err))
	}
	if _, err := os.Stat(filepath.Join(vol, "r13b-should-not-exist.txt")); err == nil {
		t.Fatal("escaped to volume root")
	}
}

func r13bReparseTag(t *testing.T, p string) uint32 {
	t.Helper()
	u, _ := windows.UTF16PtrFromString(p)
	h, err := windows.CreateFile(u, windows.FILE_READ_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		t.Fatalf("open for tag: %s", r13bDescribe(err))
	}
	defer windows.CloseHandle(h)
	var tag struct{ Attributes, Tag uint32 }
	if err := windows.GetFileInformationByHandleEx(h, windows.FileAttributeTagInfo, (*byte)(unsafe.Pointer(&tag)), uint32(unsafe.Sizeof(tag))); err != nil {
		t.Fatal(err)
	}
	return tag.Tag
}

// r13bRetag turns an existing EMPTY directory into a junction to target with
// FSCTL_SET_REPARSE_POINT (owner may do this without admin for mount points).
func r13bRetag(dir, target string) error {
	u, _ := windows.UTF16PtrFromString(dir)
	h, err := windows.CreateFile(u, windows.GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	sub := utf16.Encode([]rune(`\??\` + target))
	prn := utf16.Encode([]rune(target))
	var path []uint16
	path = append(path, sub...)
	path = append(path, 0)
	path = append(path, prn...)
	path = append(path, 0)
	dataLen := 8 + 2*len(path)
	buf := make([]byte, 8+dataLen)
	le := func(off int, v uint32, n int) {
		for i := 0; i < n; i++ {
			buf[off+i] = byte(v >> (8 * i))
		}
	}
	le(0, windows.IO_REPARSE_TAG_MOUNT_POINT, 4)
	le(4, uint32(dataLen), 2)
	le(8, 0, 2)
	le(10, uint32(2*len(sub)), 2)
	le(12, uint32(2*(len(sub)+1)), 2)
	le(14, uint32(2*len(prn)), 2)
	for i, c := range path {
		le(16+2*i, uint32(c), 2)
	}
	var n uint32
	return windows.DeviceIoControl(h, windows.FSCTL_SET_REPARSE_POINT, &buf[0], uint32(len(buf)), nil, 0, &n, nil)
}

// Retag: an existing (validated) plain directory becomes a junction later.
func TestR13bWindowsRetagDirectory(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	r13bWrite(t, filepath.Join(out, "victim"), "outside")
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	ft := r13Tools(t, root)
	if err := ft.Write("sub/probe", []byte("ok")); err != nil {
		t.Fatalf("pre-retag write: %s", r13bDescribe(err))
	}
	if err := ft.Delete("sub/probe"); err != nil {
		t.Fatal(err)
	}
	if err := r13bRetag(sub, out); err != nil {
		t.Fatalf("BARRIER: retag failed: %s", r13bDescribe(err))
	}
	t.Cleanup(func() { os.Remove(sub) })
	if r13bRead(filepath.Join(sub, "victim")) != "outside" {
		t.Fatal("BARRIER: retag not effective")
	}
	if _, err := ft.Read("sub/victim"); !errors.Is(err, errFileToolEscape) {
		t.Fatalf("read after retag: %s", r13bDescribe(err))
	}
	if err := ft.Write("sub/victim", []byte("bad")); !errors.Is(err, errFileToolEscape) {
		t.Fatalf("write after retag: %s", r13bDescribe(err))
	}
	if r13bRead(filepath.Join(out, "victim")) != "outside" {
		t.Fatal("escaped")
	}
}

// Retag of the pinned (empty) root itself after the pin.
func TestR13bWindowsRetagPinnedRoot(t *testing.T) {
	base, out := t.TempDir(), t.TempDir()
	root := filepath.Join(base, "ws")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	r13bWrite(t, filepath.Join(out, "victim"), "outside")
	ft := r13Tools(t, root)
	if err := r13bRetag(root, out); err != nil {
		t.Fatalf("BARRIER: retag root failed: %s", r13bDescribe(err))
	}
	t.Cleanup(func() { os.Remove(root) })
	werr := ft.Write("victim", []byte("bad"))
	_, rerr := ft.Read("victim")
	t.Logf("after root retag: write=%s read=%s", r13bDescribe(werr), r13bDescribe(rerr))
	if r13bRead(filepath.Join(out, "victim")) != "outside" {
		t.Fatal("pinned root retag escaped (write)")
	}
	if rerr == nil {
		t.Fatal("read through retagged root returned data")
	}
}

// ---------- hard links ----------

func TestR13bWindowsHardlinkAfterOpen(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	r13bWrite(t, filepath.Join(root, "victim"), "inside")
	ft := r13Tools(t, root)
	f, err := openBeneathAt(ft.root, ft.Root, "victim")
	if err != nil {
		t.Fatalf("open: %s", r13bDescribe(err))
	}
	if err := os.Link(filepath.Join(root, "victim"), filepath.Join(out, "alias")); err != nil {
		f.Close()
		t.Fatalf("BARRIER: link after open failed: %s", r13bDescribe(err))
	}
	f.Close()
	if _, err := ft.Read("victim"); err == nil {
		t.Fatal("read of now multi-link file accepted")
	}
	if err := ft.Write("victim", []byte("bad")); err == nil {
		t.Fatal("write of multi-link file accepted")
	}
	if err := ft.Delete("victim"); err == nil {
		t.Fatal("delete of multi-link file accepted")
	}
	if r13bRead(filepath.Join(out, "alias")) != "inside" {
		t.Fatal("outside alias changed")
	}
}

// A hard link added inside the write window must not be written through:
// rename replaces the directory entry, the outside alias keeps old content.
func TestR13bWindowsHardlinkInWindow(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	r13bWrite(t, filepath.Join(root, "victim"), "old")
	ft := r13Tools(t, root)
	linked := false
	atomicWriteTestHook = func(string) error {
		if err := os.Link(filepath.Join(root, "victim"), filepath.Join(out, "alias")); err != nil {
			return fmt.Errorf("link: %w", err)
		}
		linked = true
		return nil
	}
	defer func() { atomicWriteTestHook = nil }()
	err := ft.Write("victim", []byte("new"))
	if !linked {
		t.Fatalf("BARRIER: link in window failed: %s", r13bDescribe(err))
	}
	if err != nil {
		t.Fatalf("write: %s", r13bDescribe(err))
	}
	if got := r13bRead(filepath.Join(out, "alias")); got != "old" {
		t.Fatalf("written through hard link: alias=%q", got)
	}
	if got := r13bRead(filepath.Join(root, "victim")); got != "new" {
		t.Fatalf("target %q", got)
	}
}

// ---------- 8.3 without tilde ----------

func r13bEnablePrivilege(name string) error {
	var tok windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &tok); err != nil {
		return err
	}
	defer tok.Close()
	var luid windows.LUID
	n, _ := windows.UTF16PtrFromString(name)
	if err := windows.LookupPrivilegeValue(nil, n, &luid); err != nil {
		return err
	}
	tp := windows.Tokenprivileges{PrivilegeCount: 1}
	tp.Privileges[0] = windows.LUIDAndAttributes{Luid: luid, Attributes: windows.SE_PRIVILEGE_ENABLED}
	if err := windows.AdjustTokenPrivileges(tok, false, &tp, 0, nil, nil); err != nil {
		return err
	}
	if e := windows.GetLastError(); e == windows.ERROR_NOT_ALL_ASSIGNED {
		return e
	}
	return nil
}

func TestR13bWindowsShortNameWithoutTilde(t *testing.T) {
	if !r13bElevated() {
		t.Skip("SKIP-ADMIN: assigning a tilde-free 8.3 name (SetFileShortNameW) requires SeRestorePrivilege (elevated); tilde aliases are covered by TestR13WindowsShortAlias and the GetFinalPathNameByHandle check")
	}
	if err := r13bEnablePrivilege("SeRestorePrivilege"); err != nil {
		t.Skipf("SKIP-PRIVILEGE: SeRestorePrivilege not available: %s", r13bDescribe(err))
	}
	root := t.TempDir()
	long := filepath.Join(root, "A quite long filename without tilde.txt")
	r13bWrite(t, long, "inside")
	u, _ := windows.UTF16PtrFromString(long)
	h, err := windows.CreateFile(u, windows.GENERIC_ALL|windows.DELETE, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		t.Fatal(err)
	}
	sn, _ := windows.UTF16PtrFromString("NOTILDE.TXT")
	r1, _, e := r13bSetFileShortNameW.Call(uintptr(h), uintptr(unsafe.Pointer(sn)))
	windows.CloseHandle(h)
	if r1 == 0 {
		t.Skipf("SKIP-VOLUME: SetFileShortNameW refused (8.3 names disabled on this volume?): %s", r13bDescribe(e))
	}
	buf := make([]uint16, 32768)
	n, err := windows.GetShortPathName(u, &buf[0], uint32(len(buf)))
	if err != nil {
		t.Fatal(err)
	}
	short := filepath.Base(windows.UTF16ToString(buf[:n]))
	t.Logf("short alias=%q", short)
	if !strings.EqualFold(short, "NOTILDE.TXT") || strings.Contains(short, "~") {
		t.Fatalf("BARRIER: alias not applied: %q", short)
	}
	ft := r13Tools(t, root)
	if _, err := ft.Read(short); err == nil {
		t.Fatal("tilde-free alias read accepted")
	}
	if err := ft.Write(short, []byte("bad")); err == nil {
		t.Fatal("tilde-free alias write accepted")
	}
	if err := ft.Delete(short); err == nil {
		t.Fatal("tilde-free alias delete accepted")
	}
	if r13bRead(long) != "inside" {
		t.Fatal("long file changed through alias")
	}
}

// ---------- NUL / UNC / drive variants ----------

func TestR13bWindowsPathVariants(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	r13bWrite(t, filepath.Join(out, "x"), "outside")
	ft := r13Tools(t, root)
	s := r13bSession(t, root)
	vol := filepath.VolumeName(root)
	names := []string{
		"NUL", "nul", "Nul.txt", "a/nul", "a/NUL.tar.gz", "nul:", "CONIN$", "conout$.txt", "COM1 ", "a./b",
		`\\.\NUL`, `\\.\C:\x`, `\\?\GLOBALROOT\Device\Null`, `\\?\C:\Windows\win.ini`,
		`\\localhost\C$\Windows\win.ini`, `//localhost/c$/x`, `\\?\UNC\localhost\c$\x`, `\??\C:\x`,
		"C:x", "c:x.txt", vol + "x", `C:\Windows\win.ini`, "c:/windows/win.ini",
		"x::$DATA", "x:$DATA", "x:s:$DATA", `sub\..\..\x`, filepath.Join(out, "x"),
	}
	for _, name := range names {
		t.Run(strings.NewReplacer(`\`, "_", "/", "_", ":", "_").Replace(name), func(t *testing.T) {
			if err := ft.Write(name, []byte("bad")); err == nil {
				t.Fatalf("write accepted: %q", name)
			}
			if _, err := ft.Read(name); err == nil {
				t.Fatalf("read accepted: %q", name)
			}
			if err := ft.Delete(name); err == nil {
				t.Fatalf("delete accepted: %q", name)
			}
			s.saveFile(name, "bad")
			if s.notice == "" {
				t.Fatalf("desk save accepted: %q", name)
			}
		})
	}
	if r13bRead(filepath.Join(out, "x")) != "outside" {
		t.Fatal("outside changed")
	}
	ents, _ := os.ReadDir(root)
	for _, e := range ents {
		if e.Name() != changeHistoryDir {
			t.Fatalf("unexpected entry created in root: %q", e.Name())
		}
	}
}

func TestR13bWindowsCaseSensitiveDir(t *testing.T) {
	t.Skip("SKIP-ENV: per-directory case sensitivity (fsutil file setCaseSensitiveInfo) needs the optional WSL component and changes system state; NOT TESTED")
}

// ---------- DACL / error paths / handles ----------

func r13bIcacls(t *testing.T, args ...string) {
	t.Helper()
	if out, err := exec.Command("icacls.exe", args...).CombinedOutput(); err != nil {
		t.Fatalf("BARRIER: icacls %v: %v %s", args, err, out)
	}
}

func r13bSID(t *testing.T) string {
	u, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	return "*" + u.User.Sid.String()
}

func TestR13bWindowsDACLDenied(t *testing.T) {
	if r13bElevated() {
		t.Log("note: elevated run; deny ACEs still apply to the owner SID")
	}
	root := t.TempDir()
	ft := r13Tools(t, root)
	sid := r13bSID(t)
	locked := filepath.Join(root, "locked")
	if err := os.Mkdir(locked, 0o700); err != nil {
		t.Fatal(err)
	}
	r13bWrite(t, filepath.Join(root, "secret"), "secret")
	r13bIcacls(t, locked, "/deny", sid+":(WD,AD)")
	r13bIcacls(t, filepath.Join(root, "secret"), "/deny", sid+":(R)")
	t.Cleanup(func() {
		exec.Command("icacls.exe", locked, "/remove:d", sid).Run()
		exec.Command("icacls.exe", filepath.Join(root, "secret"), "/remove:d", sid).Run()
	})
	err := ft.Write("locked/x", []byte("bad"))
	st, ok := r13bNTStatus(err)
	if !ok || st != windows.STATUS_ACCESS_DENIED {
		t.Fatalf("write into deny-dir: want NTSTATUS 0xC0000022, got %s", r13bDescribe(err))
	}
	t.Logf("deny-dir write: %s", r13bDescribe(err))
	_, err = ft.Read("secret")
	st, ok = r13bNTStatus(err)
	if !ok || st != windows.STATUS_ACCESS_DENIED {
		t.Fatalf("read of deny-read file: want NTSTATUS 0xC0000022, got %s", r13bDescribe(err))
	}
	exec.Command("icacls.exe", locked, "/remove:d", sid).Run()
	r13bNoTemp(t, locked)
}

func TestR13bWindowsDACLOtherAccount(t *testing.T) {
	t.Skip("SKIP-ACCOUNT: history privacy against a second low-privilege local account needs creating that account (admin); NOT TESTED")
}

func TestR13bWindowsReadonlyErrors(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "ro")
	r13bWrite(t, p, "keep")
	u, _ := windows.UTF16PtrFromString(p)
	if err := windows.SetFileAttributes(u, windows.FILE_ATTRIBUTE_READONLY); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { windows.SetFileAttributes(u, windows.FILE_ATTRIBUTE_NORMAL) })
	ft := r13Tools(t, root)
	err := ft.Delete("ro")
	if e, ok := r13bErrno(err); !ok || e != windows.ERROR_ACCESS_DENIED {
		t.Fatalf("delete readonly: want errno 5, got %s", r13bDescribe(err))
	}
	err = ft.Write("ro", []byte("bad"))
	if st, ok := r13bNTStatus(err); !ok || st != windows.STATUS_ACCESS_DENIED {
		t.Fatalf("replace readonly: want NTSTATUS 0xC0000022, got %s", r13bDescribe(err))
	}
	if r13bRead(p) != "keep" {
		t.Fatal("readonly content changed")
	}
	r13bNoTemp(t, root)
}

func TestR13bWindowsTempErrorPath(t *testing.T) {
	root := t.TempDir()
	r13bWrite(t, filepath.Join(root, "d", "v"), "keep")
	ft := r13Tools(t, root)
	injected := errors.New("injected")
	atomicWriteTestHook = func(string) error { return injected }
	defer func() { atomicWriteTestHook = nil }()
	if err := ft.Write("d/v", []byte("bad")); !errors.Is(err, injected) {
		t.Fatalf("want injected error, got %s", r13bDescribe(err))
	}
	if err := ft.Write("d/new", []byte("bad")); !errors.Is(err, injected) {
		t.Fatalf("want injected error, got %s", r13bDescribe(err))
	}
	if r13bRead(filepath.Join(root, "d", "v")) != "keep" {
		t.Fatal("target changed")
	}
	if _, err := os.Stat(filepath.Join(root, "d", "new")); err == nil {
		t.Fatal("new file created on failure")
	}
	r13bNoTemp(t, filepath.Join(root, "d"))
}

func r13bHandleCount(t *testing.T) uint32 {
	var n uint32
	r, _, e := r13bGetProcessHandleCnt.Call(uintptr(windows.CurrentProcess()), uintptr(unsafe.Pointer(&n)))
	if r == 0 {
		t.Fatalf("GetProcessHandleCount: %v", e)
	}
	return n
}

func TestR13bWindowsHandleLeak(t *testing.T) {
	root := t.TempDir()
	ft := r13Tools(t, root)
	s := r13bSession(t, root)
	for i := 0; i < 20; i++ { // warm up
		ft.Write("a/b", []byte("x"))
	}
	before := r13bHandleCount(t)
	for i := 0; i < 500; i++ {
		if err := ft.Write("a/b", []byte(fmt.Sprint(i))); err != nil {
			t.Fatal(err)
		}
		if _, err := ft.Read("a/b"); err != nil {
			t.Fatal(err)
		}
		ft.Read("a/missing")
		ft.Write("a/nul", nil)
		s.localText("a/b")
	}
	if err := ft.Delete("a/b"); err != nil {
		t.Fatal(err)
	}
	after := r13bHandleCount(t)
	t.Logf("handles before=%d after=%d", before, after)
	if after > before+32 {
		t.Fatalf("handle leak: %d -> %d", before, after)
	}
	r13bNoTemp(t, filepath.Join(root, "a"))
}
