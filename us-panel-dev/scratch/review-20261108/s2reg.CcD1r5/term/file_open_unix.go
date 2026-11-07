//go:build !windows

package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// openBeneath opens root/rel for reading without ever following a symlink
// in any component: each directory is opened with openat(O_NOFOLLOW|
// O_DIRECTORY) relative to the previous fd, so swapping a component for a
// symlink between check and use cannot redirect the open (no stat-then-open
// TOCTOU). The final fd is fstat'ed: regular file with exactly one link.
func openBeneath(root, rel string) (*os.File, error) {
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "." || strings.HasPrefix(rel, "/") || rel == ".." || strings.HasPrefix(rel, "../") {
		return nil, errFileToolEscape
	}
	parts := strings.Split(rel, "/")
	dirfd, err := unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	for _, p := range parts[:len(parts)-1] {
		if p == "" || p == "." || p == ".." {
			unix.Close(dirfd)
			return nil, errFileToolEscape
		}
		next, err := unix.Openat(dirfd, p, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		unix.Close(dirfd)
		if err != nil {
			return nil, errFileToolEscape
		}
		dirfd = next
	}
	fd, err := unix.Openat(dirfd, parts[len(parts)-1], unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	unix.Close(dirfd)
	if err != nil {
		return nil, errFileToolEscape
	}
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil || st.Mode&unix.S_IFMT != unix.S_IFREG || st.Nlink != 1 {
		unix.Close(fd)
		return nil, errors.New("normal dosya değil veya birden çok bağlantılı (hard link)")
	}
	return os.NewFile(uintptr(fd), filepath.Join(root, rel)), nil
}
