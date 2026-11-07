//go:build windows

package i18n

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

type rootHandle = windows.Handle

// Directory/file access rights (winnt.h).
const (
	fileListDirectory   = 0x0001
	fileAddFile         = 0x0002
	fileAddSubdirectory = 0x0004
	fileTraverse        = 0x0020
	fileReadAttributes  = 0x0080
	dirAccess           = fileListDirectory | fileTraverse | fileAddFile | fileAddSubdirectory | fileReadAttributes
)

const shareAll = windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE

// ntOpen opens name relative to dir (or an absolute NT path when dir is 0)
// with NtCreateFile. FILE_OPEN_REPARSE_POINT makes the open land on a
// symlink/junction itself instead of its target, so the handle can be
// verified with GetFileInformationByHandle before any data is used.
func ntOpen(dir windows.Handle, name string, access, disposition, options uint32) (windows.Handle, error) {
	uname, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return 0, err
	}
	oa := &windows.OBJECT_ATTRIBUTES{RootDirectory: dir, ObjectName: uname, Attributes: windows.OBJ_CASE_INSENSITIVE}
	oa.Length = uint32(unsafe.Sizeof(*oa))
	var iosb windows.IO_STATUS_BLOCK
	var h windows.Handle
	err = windows.NtCreateFile(&h, access|windows.SYNCHRONIZE, oa, &iosb, nil, windows.FILE_ATTRIBUTE_NORMAL, shareAll,
		disposition, options|windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_OPEN_REPARSE_POINT, 0, 0)
	if err != nil {
		return 0, err
	}
	return h, nil
}

// verify checks the opened handle (not a path): no reparse point, and the
// expected kind (directory or regular file).
func verify(h windows.Handle, wantDir bool) (*windows.ByHandleFileInformation, error) {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return nil, err
	}
	isDir := info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || isDir != wantDir {
		return nil, errUnsafePath
	}
	return &info, nil
}

// openRoot pins the already-resolved trusted root as a directory handle.
func openRoot(resolved string) (rootHandle, error) {
	h, err := ntOpen(0, `\??\`+resolved, dirAccess, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE)
	if err != nil {
		return 0, err
	}
	if _, err := verify(h, true); err != nil {
		windows.CloseHandle(h)
		return 0, err
	}
	return h, nil
}

func closeBelow(root rootHandle, h windows.Handle) {
	if h != root && h != 0 {
		windows.CloseHandle(h)
	}
}

func walkBelow(root rootHandle, parts []string, create bool) (windows.Handle, error) {
	h := root
	disp := uint32(windows.FILE_OPEN)
	if create {
		disp = windows.FILE_OPEN_IF
	}
	for _, part := range parts[:len(parts)-1] {
		next, err := ntOpen(h, part, dirAccess, disp, windows.FILE_DIRECTORY_FILE)
		closeBelow(root, h)
		if err != nil {
			return 0, err
		}
		if _, err := verify(next, true); err != nil {
			windows.CloseHandle(next)
			return 0, err
		}
		h = next
	}
	return h, nil
}

func readPrefFile(root rootHandle, parts []string) ([]byte, error) {
	dir, err := walkBelow(root, parts, false)
	if err != nil {
		return nil, err
	}
	defer closeBelow(root, dir)
	h, err := ntOpen(dir, parts[len(parts)-1], windows.FILE_GENERIC_READ, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(h), parts[len(parts)-1])
	defer f.Close()
	info, err := verify(h, false)
	if err != nil {
		return nil, err
	}
	if info.FileSizeHigh != 0 || info.FileSizeLow > maxPrefSize {
		return nil, errUnsafePath
	}
	return readLimited(f)
}

type fileRenameInformation struct {
	ReplaceIfExists uint32
	RootDirectory   windows.Handle
	FileNameLength  uint32
	FileName        [1]uint16
}

// renameByHandle renames the open file h to name inside dir, relative to
// the dir handle (NtSetInformationFile/FileRenameInformation), replacing a
// regular file atomically.
func renameByHandle(h, dir windows.Handle, name string) error {
	u16, err := windows.UTF16FromString(name)
	if err != nil {
		return err
	}
	n := len(u16) - 1
	var probe fileRenameInformation
	size := int(unsafe.Offsetof(probe.FileName)) + n*2
	if size < int(unsafe.Sizeof(probe)) {
		size = int(unsafe.Sizeof(probe))
	}
	buf := make([]byte, size)
	ri := (*fileRenameInformation)(unsafe.Pointer(&buf[0]))
	ri.ReplaceIfExists = 1
	ri.RootDirectory = dir
	ri.FileNameLength = uint32(n * 2)
	copy(unsafe.Slice(&ri.FileName[0], n), u16[:n])
	var iosb windows.IO_STATUS_BLOCK
	return windows.NtSetInformationFile(h, &iosb, &buf[0], uint32(size), windows.FileRenameInformation)
}

func deleteByHandle(h windows.Handle) {
	var del uint8 = 1
	_ = windows.SetFileInformationByHandle(h, windows.FileDispositionInfo, &del, 1)
}

func writePrefFile(root rootHandle, parts []string, raw []byte) error {
	dir, err := walkBelow(root, parts, true)
	if err != nil {
		return err
	}
	defer closeBelow(root, dir)
	base := parts[len(parts)-1]
	// An existing target must be a regular file (checked on its handle).
	if old, err := ntOpen(dir, base, fileReadAttributes, windows.FILE_OPEN, 0); err == nil {
		_, verr := verify(old, false)
		windows.CloseHandle(old)
		if verr != nil {
			return verr
		}
	}
	var rnd [8]byte
	if _, err := rand.Read(rnd[:]); err != nil {
		return err
	}
	tmp := ".lang-" + hex.EncodeToString(rnd[:]) + ".tmp"
	h, err := ntOpen(dir, tmp, windows.FILE_GENERIC_WRITE|windows.DELETE|fileReadAttributes, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	var done uint32
	err = windows.WriteFile(h, raw, &done, nil)
	if err == nil && int(done) != len(raw) {
		err = errUnsafePath
	}
	if err == nil {
		err = windows.FlushFileBuffers(h)
	}
	if err == nil {
		err = renameByHandle(h, dir, base)
	}
	if err != nil {
		deleteByHandle(h)
	}
	return err
}
