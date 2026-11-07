//go:build linux

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const landlockChildArg = "--us-landlock-child-v1"

var probeLandlockABI = landlockABI

func landlockABI() (int, error) {
	abi, _, e := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET, 0, 0, unix.LANDLOCK_CREATE_RULESET_VERSION)
	if e != 0 {
		return 0, e
	}
	if abi < 1 {
		return 0, fmt.Errorf("Landlock ABI < 1")
	}
	return int(abi), nil
}

func sandboxWarning(err error) error {
	e := fmt.Errorf("UYARI: Landlock dosya izolasyonu uygulanamadı; güvenlik için komut çalıştırılmadı (sessiz geri dönüş yok): %w", err)
	log.Print(e)
	return e
}

// Only the re-exec child is restricted. LockOSThread is never undone: the
// restricted thread execs the tool, which inherits the irrevocable policy.
func init() {
	if len(os.Args) < 2 || os.Args[1] != landlockChildArg {
		return
	}
	if len(os.Args) < 5 {
		fmt.Fprintln(os.Stderr, sandboxWarning(fmt.Errorf("geçersiz çocuk argümanları")))
		os.Exit(126)
	}
	fd, err := strconv.Atoi(os.Args[3])
	if err != nil || fd < 3 {
		os.Exit(126)
	}
	status := os.NewFile(uintptr(fd), "landlock-exec-status")
	unix.CloseOnExec(fd)
	fail := func(err error) {
		e := sandboxWarning(err)
		fmt.Fprintln(status, e)
		fmt.Fprintln(os.Stderr, e)
		os.Exit(126)
	}
	runtime.LockOSThread()
	if err := applyLandlock(os.Args[2], os.Args[4]); err != nil {
		fail(err)
	}
	if _, err := io.WriteString(status, "READY\n"); err != nil {
		os.Exit(126)
	}
	if err := syscall.Exec(os.Args[4], os.Args[4:], os.Environ()); err != nil {
		fail(err)
	}
	os.Exit(126)
}

func sandboxPreflight() error {
	if abi, err := probeLandlockABI(); err != nil {
		return sandboxWarning(err)
	} else if abi < 3 {
		return sandboxWarning(fmt.Errorf("Landlock ABI %d: TRUNCATE koruması için ABI >= 3 gerekli", abi))
	}
	return nil
}

func prepareChildSandbox(cmd *exec.Cmd, root string) (func(), func() error, error) {
	if err := sandboxPreflight(); err != nil {
		return nil, nil, err
	}
	// Keep all writable runtime data inside this workspace, never the user's
	// HOME or shared caches. Refuse symlinked runtime directories.
	base := filepath.Join(root, ".us-exec")
	for _, rel := range []string{"", "home", "tmp", "go-cache", "go-path", "go-mod", "npm-cache"} {
		p := filepath.Join(base, rel)
		if err := os.Mkdir(p, 0700); err != nil && !os.IsExist(err) {
			return nil, nil, sandboxWarning(err)
		}
		st, err := os.Lstat(p)
		if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
			return nil, nil, sandboxWarning(fmt.Errorf("çalıştırma önbelleği güvenli klasör değil"))
		}
	}
	set := map[string]string{"HOME": filepath.Join(base, "home"), "TMPDIR": filepath.Join(base, "tmp"), "TMP": filepath.Join(base, "tmp"), "TEMP": filepath.Join(base, "tmp"), "GOCACHE": filepath.Join(base, "go-cache"), "GOPATH": filepath.Join(base, "go-path"), "GOMODCACHE": filepath.Join(base, "go-mod"), "npm_config_cache": filepath.Join(base, "npm-cache")}
	cmd.Env = replaceSandboxEnv(cmd.Env, set)
	self, err := os.Executable()
	if err != nil {
		return nil, nil, sandboxWarning(err)
	}
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, sandboxWarning(err)
	}
	fd := 3 + len(cmd.ExtraFiles)
	cmd.ExtraFiles = append(cmd.ExtraFiles, w)
	args := []string{self, landlockChildArg, root, strconv.Itoa(fd), cmd.Path}
	args = append(args, cmd.Args[1:]...)
	cmd.Path = self
	cmd.Args = args
	cleanup := func() { r.Close(); w.Close() }
	ready := func() error {
		w.Close()
		b, err := io.ReadAll(io.LimitReader(r, 4096))
		r.Close()
		if err != nil {
			return err
		}
		if string(b) != "READY\n" {
			return fmt.Errorf("UYARI: Landlock çocuk exec başlamadı: %s", b)
		}
		return nil
	}
	return cleanup, ready, nil
}

func applyLandlock(root, bin string) error {
	abi, err := landlockABI()
	if err != nil {
		return err
	}
	// ABI 1 covers bits 0..12; REFER (2), TRUNCATE (3), device ioctl (5).
	handled := uint64((1 << 13) - 1)
	if abi >= 2 {
		handled |= unix.LANDLOCK_ACCESS_FS_REFER
	}
	if abi >= 3 {
		handled |= unix.LANDLOCK_ACCESS_FS_TRUNCATE
	}
	if abi >= 5 {
		handled |= unix.LANDLOCK_ACCESS_FS_IOCTL_DEV
	}
	attr := unix.LandlockRulesetAttr{Access_fs: handled}
	fd, _, e := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET, uintptr(unsafe.Pointer(&attr)), 8, 0)
	if e != 0 {
		return e
	}
	defer unix.Close(int(fd))
	read := uint64(unix.LANDLOCK_ACCESS_FS_READ_FILE | unix.LANDLOCK_ACCESS_FS_READ_DIR | unix.LANDLOCK_ACCESS_FS_EXECUTE)
	fileRead := uint64(unix.LANDLOCK_ACCESS_FS_READ_FILE)
	add := func(path string, access uint64, optional bool) error {
		p, err := unix.Open(path, unix.O_PATH|unix.O_CLOEXEC, 0)
		if optional && os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Landlock yol kuralı: %s: %w", path, err)
		}
		defer unix.Close(p)
		var st unix.Stat_t
		if err := unix.Fstat(p, &st); err != nil {
			return err
		}
		if st.Mode&unix.S_IFMT != unix.S_IFDIR {
			access &= unix.LANDLOCK_ACCESS_FS_EXECUTE | unix.LANDLOCK_ACCESS_FS_READ_FILE | unix.LANDLOCK_ACCESS_FS_WRITE_FILE | unix.LANDLOCK_ACCESS_FS_TRUNCATE | unix.LANDLOCK_ACCESS_FS_IOCTL_DEV
		}
		a := unix.LandlockPathBeneathAttr{Allowed_access: access & handled, Parent_fd: int32(p)}
		_, _, e := unix.Syscall6(unix.SYS_LANDLOCK_ADD_RULE, fd, unix.LANDLOCK_RULE_PATH_BENEATH, uintptr(unsafe.Pointer(&a)), 0, 0, 0)
		if e != 0 {
			return fmt.Errorf("Landlock kuralı: %s: %w", path, e)
		}
		return nil
	}
	// Full project rights, including local cache/temp; never allow its parent.
	if err := add(root, handled, false); err != nil {
		return err
	}
	// Trusted system/toolchain directories are read/exec only, not HOME.
	for _, p := range []string{"/usr", "/bin", "/lib", "/lib64", runtime.GOROOT()} {
		if err := add(p, read, true); err != nil {
			return err
		}
	}
	// Authorize only the selected executable, never its arbitrary PATH parent
	// (which could itself be HOME or a sibling-workspace parent).
	realBin, err := filepath.EvalSymlinks(bin)
	if err != nil {
		return err
	}
	if err := add(realBin, read, false); err != nil {
		return err
	}
	// Minimal libc/DNS/TLS files; no /etc tree, /proc, /tmp or home grants.
	for _, p := range []string{"/etc/ld.so.cache", "/etc/ld.so.conf", "/etc/nsswitch.conf", "/etc/hosts", "/etc/resolv.conf", "/etc/localtime", "/etc/ssl/certs", "/etc/ssl/openssl.cnf", "/dev/urandom", "/dev/random"} {
		if err := add(p, fileRead|unix.LANDLOCK_ACCESS_FS_READ_DIR, true); err != nil {
			return err
		}
	}
	if err := add("/dev/null", fileRead|unix.LANDLOCK_ACCESS_FS_WRITE_FILE, true); err != nil {
		return err
	}
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return err
	}
	_, _, e = unix.Syscall(unix.SYS_LANDLOCK_RESTRICT_SELF, fd, 0, 0)
	if e != 0 {
		return e
	}
	return nil
}
