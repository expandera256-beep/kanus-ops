//go:build !windows

package main

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	pathpkg "path"
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
	h, err := pinRoot(root)
	if err != nil {
		return nil, err
	}
	defer h.Close()
	return openBeneathAt(h, root, rel)
}

// rootHandle is a workspace root opened once (O_DIRECTORY|O_NOFOLLOW) and
// kept as an fd. Every later operation is openat/mkdirat/renameat/unlinkat
// relative to a dup of this fd; the root path is never re-resolved, so a
// later swap of the root (or any ancestor) for a symlink cannot redirect it.
type rootHandle struct{ f *os.File }

func pinRoot(root string) (*rootHandle, error) {
	fd, err := unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, errFileToolEscape
	}
	return &rootHandle{f: os.NewFile(uintptr(fd), root)}, nil
}

func (h *rootHandle) Close() error {
	if h == nil || h.f == nil {
		return nil
	}
	return h.f.Close()
}

// dupRoot returns a private dup of the pinned root fd (caller closes it).
func (h *rootHandle) dupRoot() (int, error) {
	if h == nil || h.f == nil {
		return -1, errFileToolEscape
	}
	raw, err := h.f.SyscallConn()
	if err != nil {
		return -1, err
	}
	nfd := -1
	var derr error
	if cerr := raw.Control(func(fd uintptr) { nfd, derr = unix.FcntlInt(fd, unix.F_DUPFD_CLOEXEC, 0) }); cerr != nil {
		return -1, cerr
	}
	if derr != nil {
		return -1, derr
	}
	return nfd, nil
}

func openBeneathAt(h *rootHandle, root, rel string) (*os.File, error) {
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "." || strings.HasPrefix(rel, "/") || rel == ".." || strings.HasPrefix(rel, "../") {
		return nil, errFileToolEscape
	}
	parts := strings.Split(rel, "/")
	dirfd, err := h.dupRoot()
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
			if err == unix.ENOENT {
				return nil, os.ErrNotExist
			}
			return nil, errFileToolEscape
		}
		dirfd = next
	}
	fd, err := unix.Openat(dirfd, parts[len(parts)-1], unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	unix.Close(dirfd)
	if err != nil {
		if err == unix.ENOENT {
			return nil, os.ErrNotExist
		}
		return nil, errFileToolEscape
	}
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil || st.Mode&unix.S_IFMT != unix.S_IFREG || st.Nlink != 1 {
		unix.Close(fd)
		return nil, errors.New("normal dosya değil veya birden çok bağlantılı (hard link)")
	}
	return os.NewFile(uintptr(fd), filepath.Join(root, rel)), nil
}

// openDirBeneath returns an fd for root/relDir, walking every component
// with openat(O_DIRECTORY|O_NOFOLLOW). With create, missing components are
// made with mkdirat relative to the parent fd.
func openDirBeneath(h *rootHandle, relDir string, create bool) (int, error) {
	relDir = filepath.ToSlash(filepath.Clean(relDir))
	if strings.HasPrefix(relDir, "/") || relDir == ".." || strings.HasPrefix(relDir, "../") {
		return -1, errFileToolEscape
	}
	dirfd, err := h.dupRoot()
	if err != nil {
		return -1, err
	}
	if relDir == "." {
		return dirfd, nil
	}
	for _, p := range strings.Split(relDir, "/") {
		if p == "" || p == "." || p == ".." {
			unix.Close(dirfd)
			return -1, errFileToolEscape
		}
		next, err := unix.Openat(dirfd, p, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if err == unix.ENOENT && create {
			mode := uint32(0o755)
			if p == changeHistoryDir {
				mode = 0o700
			}
			if e := unix.Mkdirat(dirfd, p, mode); e != nil && e != unix.EEXIST {
				unix.Close(dirfd)
				return -1, e
			}
			next, err = unix.Openat(dirfd, p, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		}
		unix.Close(dirfd)
		if err != nil {
			return -1, errFileToolEscape
		}
		dirfd = next
	}
	return dirfd, nil
}

func splitBeneath(rel string) (string, string, error) {
	rel = filepath.ToSlash(filepath.Clean(rel))
	dir, name := pathpkg.Split(rel)
	if name == "" || name == "." || name == ".." {
		return "", "", errFileToolEscape
	}
	if dir == "" {
		dir = "."
	}
	return dir, name, nil
}

// writeBeneath: parent dir via openDirBeneath, temp via
// openat(O_CREAT|O_EXCL|O_NOFOLLOW) in that dir fd, fsync, renameat within
// the same dir fd, fsync dir. On any failure the temp is unlinkat'ed from
// the same dir fd, so it can never be left outside (or inside) the root.
func writeBeneath(root, rel string, data []byte) error {
	h, err := pinRoot(root)
	if err != nil {
		return err
	}
	defer h.Close()
	return writeBeneathAt(h, root, rel, data)
}

func writeBeneathAt(h *rootHandle, root, rel string, data []byte) error {
	dir, name, err := splitBeneath(rel)
	if err != nil {
		return err
	}
	dirfd, err := openDirBeneath(h, dir, true)
	if err != nil {
		return err
	}
	defer unix.Close(dirfd)
	mode := uint32(0o644)
	if strings.HasPrefix(filepath.ToSlash(rel), changeHistoryDir+"/") {
		mode = 0o600
	}
	var st unix.Stat_t
	if err := unix.Fstatat(dirfd, name, &st, unix.AT_SYMLINK_NOFOLLOW); err == nil {
		if st.Mode&unix.S_IFMT != unix.S_IFREG || st.Nlink > 1 {
			return errors.New("normal dosya değil veya hard link")
		}
		mode = st.Mode & 0o777
	}
	var tmp string
	var fd int
	for i := 0; ; i++ {
		tmp = fmt.Sprintf(".us-write-%d", rand.Uint32())
		fd, err = unix.Openat(dirfd, tmp, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, mode)
		if err != unix.EEXIST || i > 16 {
			break
		}
	}
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		if !ok {
			_ = unix.Unlinkat(dirfd, tmp, 0)
		}
	}()
	f := os.NewFile(uintptr(fd), tmp)
	_ = f.Chmod(os.FileMode(mode))
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && atomicWriteTestHook != nil {
		err = atomicWriteTestHook(filepath.Join(root, dir, tmp))
	}
	if err != nil {
		return err
	}
	if err := unix.Renameat(dirfd, tmp, dirfd, name); err != nil {
		return err
	}
	ok = true
	_ = unix.Fsync(dirfd)
	return nil
}

// deleteBeneath: unlinkat on the parent dir fd; the target must be a
// regular file (checked with fstatat AT_SYMLINK_NOFOLLOW on the same fd).
func deleteBeneath(root, rel string) error {
	h, err := pinRoot(root)
	if err != nil {
		return err
	}
	defer h.Close()
	return deleteBeneathAt(h, rel)
}

func deleteBeneathAt(h *rootHandle, rel string) error {
	dir, name, err := splitBeneath(rel)
	if err != nil {
		return err
	}
	dirfd, err := openDirBeneath(h, dir, false)
	if err != nil {
		return err
	}
	defer unix.Close(dirfd)
	var st unix.Stat_t
	if err := unix.Fstatat(dirfd, name, &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return err
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG {
		return errors.New("yalnız normal dosya silinir")
	}
	if err := unix.Unlinkat(dirfd, name, 0); err != nil {
		return err
	}
	_ = unix.Fsync(dirfd)
	return nil
}

// Directory enumeration uses the same pinned root as reads/writes.
func readDirBeneathAt(h *rootHandle, rel string) ([]os.DirEntry, error) {
	fd, err := openDirBeneath(h, rel, false)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), rel)
	defer f.Close()
	return f.ReadDir(-1)
}
