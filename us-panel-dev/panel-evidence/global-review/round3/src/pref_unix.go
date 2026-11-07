//go:build !windows

package i18n

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

type rootHandle = int

// openRoot pins the already-resolved trusted root as a directory fd.
func openRoot(resolved string) (rootHandle, error) {
	return unix.Open(resolved, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
}

// walkBelow opens the intermediate directories of parts below root with
// openat(O_NOFOLLOW|O_DIRECTORY); a symlink below the root is rejected.
// The returned fd must be closed by the caller unless it equals root.
func walkBelow(root rootHandle, parts []string, create bool) (int, error) {
	fd := root
	for _, part := range parts[:len(parts)-1] {
		next, err := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if errors.Is(err, unix.ENOENT) && create {
			if merr := unix.Mkdirat(fd, part, 0o700); merr != nil && !errors.Is(merr, unix.EEXIST) {
				closeBelow(root, fd)
				return -1, merr
			}
			next, err = unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		}
		closeBelow(root, fd)
		if err != nil {
			if errors.Is(err, unix.ELOOP) || errors.Is(err, unix.ENOTDIR) {
				return -1, errUnsafePath
			}
			return -1, err
		}
		fd = next
	}
	return fd, nil
}

func closeBelow(root rootHandle, fd int) {
	if fd != root && fd >= 0 {
		unix.Close(fd)
	}
}

func readPrefFile(root rootHandle, parts []string) ([]byte, error) {
	dirfd, err := walkBelow(root, parts, false)
	if err != nil {
		return nil, err
	}
	defer closeBelow(root, dirfd)
	base := parts[len(parts)-1]
	fd, err := unix.Openat(dirfd, base, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), base)
	defer f.Close()
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil {
		return nil, err
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG || st.Size > maxPrefSize {
		return nil, errUnsafePath
	}
	return readLimited(f)
}

func writePrefFile(root rootHandle, parts []string, raw []byte) error {
	dirfd, err := walkBelow(root, parts, true)
	if err != nil {
		return err
	}
	defer closeBelow(root, dirfd)
	base := parts[len(parts)-1]
	var st unix.Stat_t
	if err := unix.Fstatat(dirfd, base, &st, unix.AT_SYMLINK_NOFOLLOW); err == nil && st.Mode&unix.S_IFMT != unix.S_IFREG {
		return errUnsafePath
	}
	var rnd [8]byte
	if _, err := rand.Read(rnd[:]); err != nil {
		return err
	}
	tmp := ".lang-" + hex.EncodeToString(rnd[:]) + ".tmp"
	fd, err := unix.Openat(dirfd, tmp, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0o600)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(fd), tmp)
	_, err = f.Write(raw)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = unix.Renameat(dirfd, tmp, dirfd, base)
	}
	if err != nil {
		_ = unix.Unlinkat(dirfd, tmp, 0)
		return err
	}
	_ = unix.Fsync(dirfd)
	return nil
}
