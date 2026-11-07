//go:build !windows

package i18n

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// openDirNoFollow opens dir by walking every component from "/" with
// openat(O_NOFOLLOW|O_DIRECTORY); a symlink anywhere fails. With create,
// missing components are made with mode 0700 via mkdirat.
func openDirNoFollow(dir string, create bool) (int, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return -1, err
	}
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(filepath.Clean(abs), "/"), "/") {
		if part == "" {
			continue
		}
		if part == "." || part == ".." {
			unix.Close(fd)
			return -1, errUnsafePath
		}
		next, err := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if errors.Is(err, unix.ENOENT) && create {
			if merr := unix.Mkdirat(fd, part, 0o700); merr != nil && !errors.Is(merr, unix.EEXIST) {
				unix.Close(fd)
				return -1, merr
			}
			next, err = unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		}
		unix.Close(fd)
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

func readPrefFile(path string) ([]byte, error) {
	dirfd, err := openDirNoFollow(filepath.Dir(path), false)
	if err != nil {
		return nil, err
	}
	defer unix.Close(dirfd)
	fd, err := unix.Openat(dirfd, filepath.Base(path), unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), path)
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

func writePrefFile(path string, raw []byte) error {
	dirfd, err := openDirNoFollow(filepath.Dir(path), true)
	if err != nil {
		return err
	}
	defer unix.Close(dirfd)
	base := filepath.Base(path)
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
