//go:build windows

package i18n

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// checkDirNoLink walks dir from its volume root rejecting symlinks and
// reparse points (Lstat), creating missing components when create is set.
// Windows lacks openat; this Lstat-based check is the documented fallback.
func checkDirNoLink(dir string, create bool) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	vol := filepath.VolumeName(abs)
	cur := vol + string(filepath.Separator)
	for _, part := range strings.Split(strings.TrimPrefix(abs[len(vol):], string(filepath.Separator)), string(filepath.Separator)) {
		if part == "" {
			continue
		}
		cur = filepath.Join(cur, part)
		fi, err := os.Lstat(cur)
		if errors.Is(err, fs.ErrNotExist) && create {
			if err := os.Mkdir(cur, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
				return err
			}
			fi, err = os.Lstat(cur)
		}
		if err != nil {
			return err
		}
		if fi.Mode()&(fs.ModeSymlink|fs.ModeIrregular) != 0 || !fi.IsDir() {
			return errUnsafePath
		}
	}
	return nil
}

func readPrefFile(path string) ([]byte, error) {
	if err := checkDirNoLink(filepath.Dir(path), false); err != nil {
		return nil, err
	}
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || before.Size() > maxPrefSize {
		return nil, errUnsafePath
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, errUnsafePath
	}
	return readLimited(f)
}

func writePrefFile(path string, raw []byte) error {
	dir := filepath.Dir(path)
	if err := checkDirNoLink(dir, true); err != nil {
		return err
	}
	if fi, err := os.Lstat(path); err == nil && !fi.Mode().IsRegular() {
		return errUnsafePath
	}
	tmp, err := os.CreateTemp(dir, ".lang-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_, err = tmp.Write(raw)
	if err == nil {
		err = tmp.Sync()
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = checkDirNoLink(dir, false)
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	if err != nil {
		_ = os.Remove(name)
	}
	return err
}
