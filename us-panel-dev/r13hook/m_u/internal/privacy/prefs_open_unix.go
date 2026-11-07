//go:build linux

package privacy

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// Walk from / using directory descriptors, never following a symlink in any
// component. Each directory remains pinned while opening its child. The configured
// pathname is trusted configuration, not HTTP input. No pathname-based writes.
func prefsParent(path string, create bool) (*os.File, string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, "", err
	}
	name := filepath.Base(abs)
	if name == "." || name == string(filepath.Separator) {
		return nil, "", errors.New("privacy prefs: invalid filename")
	}
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, "", err
	}
	for _, part := range strings.Split(strings.TrimPrefix(filepath.Dir(abs), "/"), "/") {
		if part == "" {
			continue
		}
		next, e := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
		if e == unix.ENOENT && create {
			e = unix.Mkdirat(fd, part, 0700)
			if e == nil || e == unix.EEXIST {
				// Persist the newly-created directory entry before moving to its child.
				if e = unix.Fsync(fd); e == nil {
					next, e = unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
				}
			}
		}
		unix.Close(fd)
		if e != nil {
			return nil, "", e
		}
		fd = next
	}
	return os.NewFile(uintptr(fd), filepath.Dir(abs)), name, nil
}

func openPrefsNoFollow(path string) (*os.File, error) {
	dir, name, err := prefsParent(path, false)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	fd, err := unix.Openat(int(dir.Fd()), name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), path)
	fi, err := f.Stat() // validate the opened fd, not a prior Lstat snapshot
	if err != nil {
		f.Close()
		return nil, err
	}
	if !fi.Mode().IsRegular() || fi.Size() > prefsMaxBytes {
		f.Close()
		return nil, errors.New("privacy prefs: regular file <= 1MiB required")
	}
	return f, nil
}

func writePrefsAtomic(path string, raw []byte, fault func(string) error) error {
	dir, name, err := prefsParent(path, true)
	if err != nil {
		return err
	}
	defer dir.Close()
	return writePrefsAt(dir, name, raw, fault)
}

func writePrefsAt(dir *os.File, name string, raw []byte, fault func(string) error) error {
	dfd := int(dir.Fd())
	var entropy [16]byte
	if _, err := rand.Read(entropy[:]); err != nil {
		return err
	}
	tmpName := ".privacy-prefs-" + hex.EncodeToString(entropy[:])
	fd, err := unix.Openat(dfd, tmpName, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(fd), tmpName)
	defer f.Close()
	renamed := false
	defer func() {
		if !renamed {
			_ = unix.Unlinkat(dfd, tmpName, 0)
		}
	}()
	check := func(stage string) error {
		if fault != nil {
			return fault(stage)
		}
		return nil
	}
	if err := check("write"); err != nil {
		return err
	}
	if _, err := f.Write(raw); err != nil {
		return err
	}
	if err := check("sync"); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := check("rename"); err != nil {
		return err
	}
	if err := unix.Renameat(dfd, tmpName, dfd, name); err != nil {
		return err
	}
	renamed = true
	// A failure here means commit durability is uncertain (rename already happened),
	// never silently reported as a successful durable save.
	if err := check("dirsync"); err != nil {
		return &prefsCommittedError{err}
	}
	if err := dir.Sync(); err != nil {
		return &prefsCommittedError{err}
	}
	return nil
}
