package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Shared safe execution used by both the Terminal and the Run paths.
// It is a defence-in-depth layer for trusted single-user projects, NOT an
// OS/container sandbox: allowed interpreters still run project code.

// safeRunBins is the allowlist of executables (by base name) that any panel
// path may start. Terminal applies the stricter terminalArgv sub-command policy
// on top of this.
var safeRunBins = map[string]bool{"go": true, "python": true, "python3": true, "py": true, "node": true, "npm": true, "uvicorn": true, "flask": true, "gunicorn": true, "php": true}

// safeEnvKeys is the environment allowlist. Everything else (tokens, license
// secrets, cloud credentials) is dropped before a child process starts.
var safeEnvKeys = map[string]bool{"PATH": true, "HOME": true, "USERPROFILE": true, "SYSTEMROOT": true, "WINDIR": true, "TEMP": true, "TMP": true, "TMPDIR": true, "LANG": true, "LC_ALL": true, "PATHEXT": true, "COMSPEC": true, "GOPATH": true, "GOCACHE": true, "GOMODCACHE": true, "GOFLAGS": true, "GOPROXY": true, "LOCALAPPDATA": true, "APPDATA": true}

const (
	safeExecMaxOutput    = 16 << 10
	safeExecDefaultLimit = 10 * time.Minute
)

// safeLimits are per-process resource limits. Applied with setrlimit on
// Linux and a Job Object on Windows (see safe_limits_*.go).
type safeLimits struct {
	CPUSeconds  uint64
	MemoryBytes uint64
	Processes   uint64
	FileBytes   uint64
}

var defaultSafeLimits = safeLimits{CPUSeconds: 600, MemoryBytes: 4 << 30, Processes: 256, FileBytes: 512 << 20}

func safeExecEnv(extra ...string) []string {
	var env []string
	for _, value := range os.Environ() {
		key, _, _ := strings.Cut(value, "=")
		if safeEnvKeys[strings.ToUpper(key)] {
			env = append(env, value)
		}
	}
	return append(env, extra...)
}

func safeBinAllowed(bin string) bool {
	base := strings.ToLower(filepath.Base(bin))
	base = strings.TrimSuffix(base, ".exe")
	return safeRunBins[base]
}

// lockCwd resolves root and dir (following symlinks) and refuses any dir that
// is not root itself or below it. It returns the resolved dir.
func lockCwd(root, dir string) (string, error) {
	if strings.TrimSpace(root) == "" || refusedFolder(root) {
		return "", errors.New("geçerli proje klasörü yok")
	}
	if dir == "" {
		dir = root
	}
	for _, part := range strings.Split(filepath.ToSlash(dir), "/") {
		if part == ".." {
			return "", errors.New("proje dışı çalışma dizini")
		}
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", errors.New("proje klasörü çözülemedi")
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", errors.New("çalışma dizini çözülemedi")
	}
	info, err := os.Stat(realDir)
	if err != nil || !info.IsDir() {
		return "", errors.New("çalışma dizini klasör değil")
	}
	if realDir != realRoot && !insideRoot(realRoot, realDir) {
		return "", errors.New("çalışma dizini proje dışına çıkıyor")
	}
	return realDir, nil
}

type safeCmd struct {
	Cmd    *exec.Cmd
	Stdout *cappedBuf
	Stderr *cappedBuf
	ctx    context.Context
	cancel context.CancelFunc
	limit  func()
}

// newSafeCmd builds a command with the allowlisted binary, locked cwd,
// filtered env, separate capped stdout/stderr, a timeout and process-group
// cancellation. Start must be called through (*safeCmd).Start.
func newSafeCmd(parent context.Context, root, dir string, argv []string, timeout time.Duration, extraEnv ...string) (*safeCmd, error) {
	if len(argv) == 0 || !safeBinAllowed(argv[0]) {
		return nil, errors.New("komut izinli değil")
	}
	for _, arg := range argv {
		if strings.ContainsRune(arg, 0) {
			return nil, errors.New("geçersiz argüman")
		}
	}
	cwd, err := lockCwd(root, dir)
	if err != nil {
		return nil, err
	}
	if timeout <= 0 || timeout > safeExecDefaultLimit*6 {
		timeout = safeExecDefaultLimit
	}
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	configureProcessTree(cmd)
	cmd.Cancel = func() error { return killProcessTree(cmd.Process) }
	cmd.WaitDelay = 2 * time.Second
	cmd.Dir = cwd
	cmd.Env = safeExecEnv(extraEnv...)
	cmd.Stdin = strings.NewReader("")
	sc := &safeCmd{Cmd: cmd, Stdout: &cappedBuf{}, Stderr: &cappedBuf{}, ctx: ctx, cancel: cancel}
	cmd.Stdout = sc.Stdout
	cmd.Stderr = sc.Stderr
	return sc, nil
}

// Start starts the process and applies OS resource limits. On failure the
// context is released and an error is returned; callers must not report the
// program as started.
func (c *safeCmd) Start() error {
	if err := c.Cmd.Start(); err != nil {
		c.cancel()
		return fmt.Errorf("işlem başlamadı: %w", err)
	}
	release, err := applySafeLimits(c.Cmd.Process, defaultSafeLimits)
	if err != nil {
		_ = killProcessTree(c.Cmd.Process)
		_ = c.Cmd.Wait()
		c.cancel()
		return fmt.Errorf("kaynak sınırı uygulanamadı: %w", err)
	}
	c.limit = release
	return nil
}

// Wait waits and returns the state: success, failed, timeout or cancelled.
func (c *safeCmd) Wait() (string, int, error) {
	err := c.Cmd.Wait()
	cause := c.ctx.Err()
	c.cancel()
	if c.limit != nil {
		c.limit()
	}
	code := -1
	if c.Cmd.ProcessState != nil {
		code = c.Cmd.ProcessState.ExitCode()
	}
	switch {
	case errors.Is(cause, context.DeadlineExceeded):
		return "timeout", code, err
	case cause != nil:
		return "cancelled", code, err
	case err != nil:
		return "failed", code, err
	}
	return "success", code, nil
}

func (c *safeCmd) Cancel() { c.cancel() }

// Run is Start+Wait.
func (c *safeCmd) Run() (string, int, error) {
	if err := c.Start(); err != nil {
		return "start_failed", -1, err
	}
	return c.Wait()
}
