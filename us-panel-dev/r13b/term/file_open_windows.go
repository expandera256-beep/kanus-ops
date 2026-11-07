//go:build windows

package main

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// rootHandle owns the initial directory handle. Paths are only labels after
// pinRoot; every descendant is opened relative to an already verified handle.
type rootHandle struct{ f *os.File }

func pinRoot(root string) (*rootHandle, error) {
	p, err := windows.UTF16PtrFromString(root)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(p, windows.FILE_GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	if err = checkHandle(h, true, false); err != nil {
		windows.CloseHandle(h)
		return nil, err
	}
	return &rootHandle{os.NewFile(uintptr(h), root)}, nil
}
func (h *rootHandle) Close() error {
	if h == nil || h.f == nil {
		return nil
	}
	return h.f.Close()
}
func (h *rootHandle) dupRoot() (windows.Handle, error) {
	if h == nil || h.f == nil {
		return 0, errFileToolEscape
	}
	c, err := h.f.SyscallConn()
	if err != nil {
		return 0, err
	}
	var dup windows.Handle
	var e error
	err = c.Control(func(fd uintptr) {
		e = windows.DuplicateHandle(windows.CurrentProcess(), windows.Handle(fd), windows.CurrentProcess(), &dup, 0, false, windows.DUPLICATE_SAME_ACCESS)
	})
	if err != nil {
		return 0, err
	}
	return dup, e
}

func checkHandle(h windows.Handle, directory, singleLink bool) error {
	var tag struct{ Attributes, Tag uint32 }
	if err := windows.GetFileInformationByHandleEx(h, windows.FileAttributeTagInfo, (*byte)(unsafe.Pointer(&tag)), uint32(unsafe.Sizeof(tag))); err != nil {
		return err
	}
	if tag.Attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || (tag.Attributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0) != directory {
		return errFileToolEscape
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return err
	}
	if singleLink && info.NumberOfLinks != 1 {
		return errors.New("normal dosya değil veya hard link")
	}
	return nil
}

// ntOpen accepts exactly one validated component, never an ambient path.
func ntOpen(parent windows.Handle, name string, directory, create bool, access uint32) (windows.Handle, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\\`) || !windowsSafeRel(name) {
		return 0, errFileToolEscape
	}
	u, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return 0, err
	}
	oa := windows.OBJECT_ATTRIBUTES{RootDirectory: parent, ObjectName: u, Attributes: windows.OBJ_CASE_INSENSITIVE}
	oa.Length = uint32(unsafe.Sizeof(oa))
	opts := uint32(windows.FILE_OPEN_REPARSE_POINT | windows.FILE_SYNCHRONOUS_IO_NONALERT)
	if directory {
		opts |= windows.FILE_DIRECTORY_FILE
	} else {
		opts |= windows.FILE_NON_DIRECTORY_FILE
	}
	disposition := uint32(windows.FILE_OPEN)
	if create {
		disposition = windows.FILE_CREATE
		if directory {
			disposition = windows.FILE_OPEN_IF
		}
	}
	var h windows.Handle
	var ios windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&h, access|windows.SYNCHRONIZE, &oa, &ios, nil, windows.FILE_ATTRIBUTE_NORMAL, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, disposition, opts, 0, 0)
	runtime.KeepAlive(u)
	if err != nil {
		if err == windows.STATUS_OBJECT_NAME_NOT_FOUND || err == windows.STATUS_OBJECT_PATH_NOT_FOUND {
			return 0, os.ErrNotExist
		}
		return 0, err
	}
	if err = checkHandle(h, directory, !directory); err == nil {
		// Reject DOS aliases even on volumes whose 8.3 alias has no tilde.
		// Query the opened object, never use the returned path for IO.
		buf := make([]uint16, 32768)
		var n uint32
		n, err = windows.GetFinalPathNameByHandle(h, &buf[0], uint32(len(buf)), 0)
		if err == nil && (n == 0 || n >= uint32(len(buf)) || !strings.EqualFold(filepath.Base(windows.UTF16ToString(buf[:n])), name)) {
			err = errFileToolEscape
		}
	}
	if err != nil {
		if create && !directory {
			_ = disposeHandle(h)
		}
		windows.CloseHandle(h)
		return 0, err
	}
	return h, nil
}
func splitWindowsRel(rel string) ([]string, error) {
	if !windowsSafeRel(rel) || filepath.IsAbs(rel) || strings.ContainsRune(rel, 0) {
		return nil, errFileToolEscape
	}
	parts := strings.Split(strings.ReplaceAll(rel, `\`, "/"), "/")
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			return nil, errFileToolEscape
		}
	}
	return parts, nil
}
func openDirBeneath(h *rootHandle, rel string, create bool) (windows.Handle, error) {
	cur, err := h.dupRoot()
	if err != nil {
		return 0, err
	}
	if rel == "." || rel == "" {
		return cur, nil
	}
	parts, err := splitWindowsRel(rel)
	if err != nil {
		windows.CloseHandle(cur)
		return 0, err
	}
	for _, p := range parts {
		next, e := ntOpen(cur, p, true, create, windows.FILE_GENERIC_READ|windows.FILE_TRAVERSE)
		windows.CloseHandle(cur)
		if e != nil {
			return 0, e
		}
		cur = next
	}
	return cur, nil
}
func windowsParent(h *rootHandle, rel string, create bool) (windows.Handle, string, error) {
	parts, err := splitWindowsRel(rel)
	if err != nil {
		return 0, "", err
	}
	dir := "."
	if len(parts) > 1 {
		dir = strings.Join(parts[:len(parts)-1], "/")
	}
	parent, err := openDirBeneath(h, dir, create)
	return parent, parts[len(parts)-1], err
}
func openBeneathAt(h *rootHandle, root, rel string) (*os.File, error) {
	parent, name, err := windowsParent(h, rel, false)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(parent)
	fd, err := ntOpen(parent, name, false, false, windows.FILE_GENERIC_READ)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), filepath.Join(root, rel)), nil
}
func disposeHandle(h windows.Handle) error {
	deleteFile := byte(1)
	return windows.SetFileInformationByHandle(h, windows.FileDispositionInfo, &deleteFile, 1)
}

// Native FILE_RENAME_INFORMATION has the same layout as FILE_RENAME_INFO.
// Use NtSetInformationFile: the Win32 wrapper canonicalizes relative names
// against cwd on affected Windows versions, defeating RootDirectory semantics.
// There is deliberately NO absolute-path fallback.
func renameHandle(h, parent windows.Handle, name string) error {
	u, err := windows.UTF16FromString(name)
	if err != nil {
		return err
	}
	var layout struct {
		Replace uint32
		Root    windows.Handle
		Length  uint32
		Name    [1]uint16
	}
	off := unsafe.Offsetof(layout.Name)
	buf := make([]byte, int(unsafe.Sizeof(layout))+2*len(u))
	p := (*struct {
		Replace uint32
		Root    windows.Handle
		Length  uint32
		Name    [1]uint16
	})(unsafe.Pointer(&buf[0]))
	p.Replace = 1
	p.Root = parent
	p.Length = uint32(2 * (len(u) - 1))
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(&buf[off])), len(u)), u)
	var ios windows.IO_STATUS_BLOCK
	return windows.NtSetInformationFile(h, &ios, &buf[0], uint32(len(buf)), windows.FileRenameInformation)
}
func writeBeneathAt(h *rootHandle, root, rel string, data []byte) error {
	limit := int64(fileToolMaxWrite)
	if strings.HasPrefix(filepath.ToSlash(rel), changeHistoryDir+"/") {
		limit = maxBackupBytes * 2
	}
	if int64(len(data)) > limit {
		return errors.New("yazma boyut sınırı aşıldı")
	}
	parent, name, err := windowsParent(h, rel, true)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(parent)
	// Existing targets must be regular, non-reparse, single-link files. Replacing
	// a concurrently substituted leaf cannot write through it: rename replaces
	// the directory entry rather than following it.
	old, err := ntOpen(parent, name, false, false, windows.FILE_GENERIC_READ)
	if err == nil {
		windows.CloseHandle(old)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var tmp string
	var fd windows.Handle
	for i := 0; i < 32; i++ {
		tmp = fmt.Sprintf(".us-write-%016x", rand.Uint64())
		fd, err = ntOpen(parent, tmp, false, true, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE)
		if err == nil {
			break
		}
		if err != windows.STATUS_OBJECT_NAME_COLLISION {
			return err
		}
	}
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(fd), tmp)
	defer f.Close()
	renamed := false
	defer func() {
		if !renamed {
			_ = disposeHandle(fd)
		}
	}()
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	if err == nil && atomicWriteTestHook != nil {
		err = atomicWriteTestHook(filepath.Join(root, filepath.Dir(rel), tmp))
	}
	if err != nil {
		return err
	}
	if err = renameHandle(fd, parent, name); err != nil {
		return err
	}
	renamed = true
	return nil
}
func deleteBeneathAt(h *rootHandle, rel string) error {
	parent, name, err := windowsParent(h, rel, false)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(parent)
	fd, err := ntOpen(parent, name, false, false, windows.FILE_GENERIC_READ|windows.DELETE)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(fd)
	return disposeHandle(fd)
}
func openBeneath(root, rel string) (*os.File, error) {
	h, err := pinRoot(root)
	if err != nil {
		return nil, err
	}
	defer h.Close()
	return openBeneathAt(h, root, rel)
}
func writeBeneath(root, rel string, data []byte) error {
	h, err := pinRoot(root)
	if err != nil {
		return err
	}
	defer h.Close()
	return writeBeneathAt(h, root, rel, data)
}
func deleteBeneath(root, rel string) error {
	h, err := pinRoot(root)
	if err != nil {
		return err
	}
	defer h.Close()
	return deleteBeneathAt(h, rel)
}
func readDirBeneathAt(h *rootHandle, rel string) ([]os.DirEntry, error) {
	fd, err := openDirBeneath(h, rel, false)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), rel)
	defer f.Close()
	return f.ReadDir(-1)
}
