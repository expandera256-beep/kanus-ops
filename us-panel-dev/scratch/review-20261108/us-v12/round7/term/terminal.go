package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Explicit user action only. No shell or model-generated commands are evaluated.
// Project programs are trusted-user code, NOT an OS/container security sandbox.
type terminalState struct {
	State    string   `json:"state"`
	Root     string   `json:"root"`
	Argv     []string `json:"argv"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	ExitCode int      `json:"exitCode"`
}
type panelTerminal struct {
	mu             sync.Mutex
	state          terminalState
	stdout, stderr cappedBuf
	cancel         context.CancelFunc
}

func terminalArgv(raw, root string) ([]string, error) {
	var argv []string
	if len(raw) > 8192 || json.Unmarshal([]byte(raw), &argv) != nil || len(argv) < 1 || len(argv) > 64 {
		return nil, fmt.Errorf("Komut JSON dizisi olmalı; örnek: [\"go\",\"test\",\"./...\"]")
	}
	for _, arg := range argv {
		if arg == "" || strings.ContainsAny(arg, "\x00\r\n;|&`$<>") {
			return nil, fmt.Errorf("Kabuk işleçleri veya boş argüman yasak.")
		}
		value := arg
		if _, v, ok := strings.Cut(arg, "="); ok {
			value = v
		}
		clean := strings.ReplaceAll(value, `\`, "/")
		if filepath.IsAbs(value) || strings.Contains(clean, ":") || strings.HasPrefix(clean, "/") {
			return nil, fmt.Errorf("Yalnız proje içi göreli yollar kullanılabilir.")
		}
		for _, part := range strings.Split(clean, "/") {
			if part == ".." {
				return nil, fmt.Errorf("Proje dışı yol yasak.")
			}
		}
		if strings.Contains(clean, "/") && clean != "./..." && !strings.HasPrefix(arg, "-") {
			if _, err := (&session{root: root}).safePath(value); err != nil {
				return nil, fmt.Errorf("Yol güvenle doğrulanamadı.")
			}
		}
	}
	allowed := false
	if len(argv) > 1 {
		switch argv[0] {
		case "go":
			allowed = map[string]bool{"version": true, "test": true, "build": true, "run": true, "vet": true}[argv[1]]
		case "python", "python3":
			allowed = strings.HasSuffix(argv[1], ".py") || (argv[1] == "-m" && len(argv) > 2 && map[string]bool{"unittest": true, "py_compile": true, "compileall": true}[argv[2]]) || argv[1] == "--version"
		case "node":
			allowed = argv[1] == "--test" || argv[1] == "--check" || argv[1] == "--version" || strings.HasSuffix(argv[1], ".js") || strings.HasSuffix(argv[1], ".mjs")
		}
	}
	for _, arg := range argv[1:] {
		if arg == "-C" || strings.HasPrefix(arg, "-C=") || arg == "-c" || arg == "-e" || arg == "--eval" || strings.HasPrefix(arg, "--eval=") || arg == "--require" || arg == "--import" || arg == "--loader" {
			allowed = false
		}
	}
	if !allowed {
		return nil, fmt.Errorf("Komut izinli değil. Go test/build/run/vet/version; Python dosyası/unittest/py_compile; Node dosyası/test/check desteklenir. Bağımlılıklar otomatik kurulmaz.")
	}
	return argv, nil
}
func (t *panelTerminal) start(root, raw string, timeout time.Duration) error {
	if timeout <= 0 || timeout > 10*time.Minute {
		return fmt.Errorf("Terminal süresi 1–600 saniye olmalı.")
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() || refusedFolder(root) {
		return fmt.Errorf("Geçerli proje klasörü yok.")
	}
	argv, err := terminalArgv(raw, root)
	if err != nil {
		return err
	}
	bin, err := exec.LookPath(argv[0])
	if err != nil {
		return fmt.Errorf("Çalıştırıcı bulunamadı: %s", argv[0])
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state.State == "running" {
		return fmt.Errorf("Terminal işlemi sürüyor; önce iptal edin.")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	cmd := exec.CommandContext(ctx, bin, argv[1:]...)
	configureProcessTree(cmd)
	cmd.Cancel = func() error { return killProcessTree(cmd.Process) }
	cmd.WaitDelay = 2 * time.Second
	cmd.Dir = root
	cmd.Env = projectTestEnv()
	// A new bounded buffer belongs to this generation; old output is discarded.
	t.stdout = cappedBuf{}
	t.stderr = cappedBuf{}
	cmd.Stdout = &t.stdout
	cmd.Stderr = &t.stderr
	if err = cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("Terminal işlemi başlamadı: %s", argv[0])
	}
	t.state = terminalState{State: "running", Root: root, Argv: argv, ExitCode: -1}
	t.cancel = cancel
	go func() {
		err := cmd.Wait()
		cause := ctx.Err()
		cancel()
		t.mu.Lock()
		defer t.mu.Unlock()
		t.cancel = nil
		t.state.ExitCode = cmd.ProcessState.ExitCode()
		switch {
		case cause == context.DeadlineExceeded:
			t.state.State = "timeout"
		case cause != nil:
			t.state.State = "cancelled"
		case err != nil:
			t.state.State = "failed"
		default:
			t.state.State = "success"
		}
	}()
	return nil
}
func (t *panelTerminal) stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cancel != nil {
		t.cancel()
	}
}
func (t *panelTerminal) snapshot() terminalState {
	t.mu.Lock()
	defer t.mu.Unlock()
	v := t.state
	if v.State == "" {
		v.State = "idle"
		v.ExitCode = -1
	}
	v.Argv = append([]string(nil), v.Argv...)
	v.Stdout = redactLine(t.stdout.String())
	v.Stderr = redactLine(t.stderr.String())
	return v
}
