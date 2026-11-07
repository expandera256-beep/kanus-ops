//go:build windows

package main

import (
	"errors"
	"os"
	"path/filepath"
)

// openBeneath on Windows: components are checked by noSymlinkPath (reparse
// points refused) before open; a race-free handle-relative open is NOT
// implemented here (NOT FIXED on Windows, documented).
// rootHandle on Windows only records the root path. Handle-relative
// (NtCreateFile RootDirectory) operations are NOT implemented this round:
// Windows Write/Delete remain path based (NOT DONE, documented in report).
type rootHandle struct{ root string }

func pinRoot(root string) (*rootHandle, error) { return &rootHandle{root: root}, nil }
func (h *rootHandle) Close() error             { return nil }

func openBeneathAt(h *rootHandle, root, rel string) (*os.File, error) {
	return openBeneath(h.root, rel)
}
func writeBeneathAt(h *rootHandle, root, rel string, data []byte) error {
	return writeBeneath(h.root, rel, data)
}
func deleteBeneathAt(h *rootHandle, rel string) error { return deleteBeneath(h.root, rel) }

func openBeneath(root, rel string) (*os.File, error) {
	full := filepath.Join(root, rel)
	if err := noSymlinkPath(root, full); err != nil {
		return nil, errFileToolEscape
	}
	f, err := os.Open(full)
	if err != nil {
		return nil, err
	}
	if info, err := f.Stat(); err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, errors.New("normal dosya değil")
	}
	return f, nil
}

// writeBeneath/deleteBeneath on Windows: path based (component check, then
// temp + rename). Not race-free against a concurrent junction swap: NOT FIXED.
func writeBeneath(root, rel string, data []byte) error {
	return atomicWriteFilePath(root, filepath.Join(root, rel), data)
}

func deleteBeneath(root, rel string) error {
	full := filepath.Join(root, rel)
	if err := noSymlinkPath(root, full); err != nil {
		return errFileToolEscape
	}
	info, err := os.Lstat(full)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("yalnız normal dosya silinir")
	}
	return os.Remove(full)
}

func atomicWriteFilePath(root, full string, data []byte) error {
	if int64(len(data)) > fileToolMaxWrite {
		return errors.New("yazma boyut sınırı aşıldı")
	}
	if err := noSymlinkPath(root, full); err != nil {
		return err
	}
	dir := filepath.Dir(full)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := noSymlinkPath(root, full); err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if info, err := os.Lstat(full); err == nil {
		mode = info.Mode().Perm()
	}
	f, err := os.CreateTemp(dir, ".us-write-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(mode); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err == nil && atomicWriteTestHook != nil {
		err = atomicWriteTestHook(tmp)
	}
	if err != nil {
		return err
	}
	if err = os.Rename(tmp, full); err != nil {
		return err
	}
	syncDir(dir)
	return nil
}
