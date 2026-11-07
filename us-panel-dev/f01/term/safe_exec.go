package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Shared safe execution used by both the Terminal and the Run paths.
// It is a defence-in-depth layer for trusted single-user projects, NOT an
// container boundary. Linux uses Landlock; Windows uses a Low IL restricted
// token (no Windows read isolation). See sandbox_*.go.

// safeRunBins is the allowlist of executables (by base name) that any panel
// path may start. Terminal applies the stricter terminalArgv sub-command policy
// on top of this.
var safeRunBins = map[string]bool{"go": true, "python": true, "python3": true, "py": true, "node": true, "npm": true, "npx": true, "rea": true, "uvicorn": true, "flask": true, "gunicorn": true, "php": true, "git": true, "pyinstaller": true}

// safeEnvKeys is the environment allowlist. GOFLAGS, CC/CXX/CGO_* and
// similar build-behaviour variables are intentionally absent (they can inject
// -toolexec or a compiler binary). Everything else (tokens, license
// secrets, cloud credentials) is dropped before a child process starts.
var safeEnvKeys = map[string]bool{"PATH": true, "HOME": true, "USERPROFILE": true, "SYSTEMROOT": true, "WINDIR": true, "TEMP": true, "TMP": true, "TMPDIR": true, "LANG": true, "LC_ALL": true, "PATHEXT": true, "COMSPEC": true, "GOPATH": true, "GOCACHE": true, "GOMODCACHE": true, "GOPROXY": true, "LOCALAPPDATA": true, "APPDATA": true}

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

// trustedBin resolves an allowlisted executable by base name over the PATH
// entries that are absolute and NOT inside the workspace root, plus
// GOROOT/bin for go. Both the candidate and a caller-supplied absolute path
// are compared after EvalSymlinks, so a file merely named "go" or "python"
// elsewhere (e.g. a project .venv) is refused and the real path is pinned.
func trustedBin(bin, root string) (string, error) {
	if !safeBinAllowed(bin) {
		return "", errors.New("komut izinli değil")
	}
	name := filepath.Base(bin)
	realRoot := ""
	if root != "" {
		if r, err := filepath.EvalSymlinks(root); err == nil {
			realRoot = r
		}
	}
	inRoot := func(p string) bool {
		if realRoot == "" {
			return false
		}
		rp, err := filepath.EvalSymlinks(p)
		if err != nil {
			rp = p
		}
		return rp == realRoot || insideRoot(realRoot, rp)
	}
	var dirs []string
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" || !filepath.IsAbs(dir) || inRoot(dir) {
			continue
		}
		dirs = append(dirs, dir)
	}
	if strings.TrimSuffix(strings.ToLower(name), ".exe") == "go" {
		dirs = append(dirs, filepath.Join(runtime.GOROOT(), "bin"))
	}
	found := ""
	for _, dir := range dirs {
		for _, cand := range binCandidates(dir, name) {
			info, err := os.Stat(cand)
			if err == nil && info.Mode().IsRegular() && executableMode(info) {
				found = cand
				break
			}
		}
		if found != "" {
			break
		}
	}
	if filepath.IsAbs(bin) && found == "" || filepath.IsAbs(bin) && !sameReal(bin, found) {
		// Operator-configured REA binary (US_REA, set in the panel's own
		// environment, never by a project): exact path, outside the root.
		if rea := strings.TrimSpace(os.Getenv("US_REA")); rea != "" && filepath.IsAbs(rea) && strings.TrimSuffix(strings.ToLower(name), ".exe") == "rea" && sameReal(bin, rea) && !inRoot(rea) {
			if info, err := os.Stat(rea); err == nil && info.Mode().IsRegular() {
				return rea, nil
			}
		}
		// Operator-selected PyInstaller is pinned exactly, like US_REA.
		if configured := strings.TrimSpace(os.Getenv("US_PYINSTALLER")); configured != "" && filepath.IsAbs(configured) && strings.TrimSuffix(strings.ToLower(name), ".exe") == "pyinstaller" && sameReal(bin, configured) && !inRoot(configured) {
			if info, err := os.Stat(configured); err == nil && info.Mode().IsRegular() && executableMode(info) {
				return configured, nil
			}
		}
		// Panel-discovered standard installs (e.g. %LOCALAPPDATA%\Programs\Python).
		for _, p := range installedPythons() {
			if sameReal(bin, p) && !inRoot(p) {
				return p, nil
			}
		}
	}
	if found == "" {
		return "", fmt.Errorf("güvenilir PATH içinde %s bulunamadı", name)
	}
	real, err := filepath.EvalSymlinks(found)
	if err != nil || inRoot(real) {
		return "", errors.New("çalışma alanı içindeki çalıştırıcı reddedildi")
	}
	if filepath.IsAbs(bin) {
		a, errA := filepath.EvalSymlinks(bin)
		if errA != nil || a != real {
			return "", errors.New("çalıştırıcı güvenilir PATH çözümüyle eşleşmiyor")
		}
	}
	return found, nil
}

func binCandidates(dir, name string) []string {
	if runtime.GOOS != "windows" || strings.Contains(name, ".") {
		return []string{filepath.Join(dir, name)}
	}
	return []string{filepath.Join(dir, name+".exe"), filepath.Join(dir, name+".cmd"), filepath.Join(dir, name+".bat")}
}

func executableMode(info os.FileInfo) bool {
	return runtime.GOOS == "windows" || info.Mode().Perm()&0o111 != 0
}

// safeRun runs argv through newSafeCmd; output is stdout then stderr.
func safeRun(parent context.Context, root, dir string, argv []string, timeout time.Duration, stdin []byte, extraEnv ...string) (string, string, int, error) {
	sc, err := newSafeCmd(parent, root, dir, argv, timeout, extraEnv...)
	if err != nil {
		return "", "start_failed", -1, err
	}
	if stdin != nil {
		sc.Cmd.Stdin = bytes.NewReader(stdin)
	}
	state, code, err := sc.Run()
	out := sc.Stdout.String()
	if e := sc.Stderr.String(); e != "" {
		if out != "" && !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		out += e
	}
	return out, state, code, err
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
	root   string
}

// newSafeCmd builds a command with the allowlisted binary, locked cwd,
// filtered env, separate capped stdout/stderr, a timeout and process-group
// cancellation. Start must be called through (*safeCmd).Start.
func newSafeCmd(parent context.Context, root, dir string, argv []string, timeout time.Duration, extraEnv ...string) (*safeCmd, error) {
	if len(argv) == 0 {
		return nil, errors.New("komut izinli değil")
	}
	bin, err := trustedBin(argv[0], root)
	if err != nil {
		return nil, err
	}
	argv = append([]string{bin}, argv[1:]...)
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
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		cancel()
		return nil, err
	}
	sc := &safeCmd{root: realRoot, Cmd: cmd, Stdout: &cappedBuf{}, Stderr: &cappedBuf{}, ctx: ctx, cancel: cancel}
	cmd.Stdout = sc.Stdout
	cmd.Stderr = sc.Stderr
	return sc, nil
}

// Start starts the process and applies OS resource limits. On failure the
// context is released and an error is returned; callers must not report the
// program as started.
func (c *safeCmd) Start() error {
	cleanup, ready, err := prepareChildSandbox(c.Cmd, c.root)
	if err != nil {
		c.cancel()
		return err
	}
	defer cleanup()
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
	if err := ready(); err != nil {
		_ = killProcessTree(c.Cmd.Process)
		_ = c.Cmd.Wait()
		release()
		c.cancel()
		return fmt.Errorf("işlem başlamadı: %w", err)
	}
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

var (
	neutralOnce sync.Once
	neutralPath string
)

// neutralDir is an empty private working directory for probes that are not
// tied to a project (interpreter --version, stdin-only syntax checks).
func neutralDir() string {
	neutralOnce.Do(func() {
		dir, err := os.MkdirTemp("", "us-safe-cwd-")
		if err != nil {
			dir = os.TempDir()
		}
		neutralPath = dir
	})
	return neutralPath
}

func sameReal(a, b string) bool {
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	return errA == nil && errB == nil && ra == rb
}
