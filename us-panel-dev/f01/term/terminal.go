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
	stdout, stderr *cappedBuf
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
	if allowed && argv[0] == "node" {
		if err := nodeArgsAllowed(argv[1:]); err != nil {
			return nil, err
		}
	}
	if allowed && argv[0] == "go" {
		if err := goFlagsAllowed(argv[1], argv[2:]); err != nil {
			return nil, err
		}
	}
	if !allowed {
		return nil, fmt.Errorf("Komut izinli değil. Go test/build/run/vet/version; Python dosyası/unittest/py_compile; Node dosyası/test/check desteklenir. Bağımlılıklar otomatik kurulmaz.")
	}
	// py_compile/compileall: isolated mode (-I) so a project-local
	// py_compile.py / compileall.py cannot shadow the stdlib module.
	if len(argv) > 2 && argv[1] == "-m" && (argv[2] == "py_compile" || argv[2] == "compileall") {
		argv = append([]string{argv[0], "-I"}, argv[1:]...)
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
	sc, err := newSafeCmd(context.Background(), root, root, append([]string{bin}, argv[1:]...), timeout, "PYTHONIOENCODING=utf-8")
	if err != nil {
		return fmt.Errorf("Terminal işlemi reddedildi: %v", err)
	}
	// A new bounded buffer belongs to this generation; old output is discarded.
	t.stdout = sc.Stdout
	t.stderr = sc.Stderr
	if err = sc.Start(); err != nil {
		return fmt.Errorf("Terminal işlemi başlamadı: %s: %v", argv[0], err)
	}
	t.state = terminalState{State: "running", Root: root, Argv: argv, ExitCode: -1}
	t.cancel = sc.Cancel
	go func() {
		state, code, _ := sc.Wait()
		t.mu.Lock()
		defer t.mu.Unlock()
		t.cancel = nil
		t.state.ExitCode = code
		t.state.State = state
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
	if t.stdout != nil {
		v.Stdout = redactLine(t.stdout.String())
	}
	if t.stderr != nil {
		v.Stderr = redactLine(t.stderr.String())
	}
	return v
}

// goFlagAllow is the per-subcommand flag allowlist. Anything else
// (-toolexec, -exec, -overlay, -modfile, -ldflags, -gcflags, -vettool, -vet,
// -args, -C, ...) is refused because it can run arbitrary programs or change
// what is built. The bool marks flags that take a separate value argument.
var goFlagAllow = map[string]map[string]bool{
	"build":   {"-v": false, "-race": false, "-trimpath": false, "-o": true, "-tags": true},
	"test":    {"-v": false, "-race": false, "-short": false, "-failfast": false, "-json": false, "-cover": false, "-count": true, "-run": true, "-skip": true, "-timeout": true, "-bench": true, "-benchtime": true, "-cpu": true, "-shuffle": true, "-list": true},
	"vet":     {},
	"run":     {"-race": false},
	"version": {},
}

func goFlagsAllowed(sub string, args []string) error {
	allow := goFlagAllow[sub]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.ContainsAny(arg, " \t") {
			return fmt.Errorf("Go argümanında boşluk yasak.")
		}
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		name, _, hasValue := strings.Cut(arg, "=")
		name = "-" + strings.TrimLeft(name, "-")
		takesValue, ok := allow[name]
		if !ok {
			return fmt.Errorf("Go bayrağı izinli değil: %s", name)
		}
		if takesValue && !hasValue {
			i++
			if i >= len(args) || strings.HasPrefix(args[i], "-") || strings.ContainsAny(args[i], " \t") {
				return fmt.Errorf("Go bayrağı değeri geçersiz: %s", name)
			}
		}
	}
	return nil
}

// nodeArgsAllowed: node --version (alone), node --test|--check <paths...>
// (no further options: -p/--print, --import, --require, --loader, -e ...
// are refused in every form), or node <file.js|.mjs> [program args...]
// (arguments after the script are passed to the script, not parsed by node).
func nodeArgsAllowed(args []string) error {
	switch args[0] {
	case "--version":
		if len(args) != 1 {
			return fmt.Errorf("Node --version ek argüman almaz.")
		}
		return nil
	case "--test", "--check":
		for _, a := range args[1:] {
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("Node seçeneği izinli değil: %s", a)
			}
		}
		if args[0] == "--check" && len(args) != 2 {
			return fmt.Errorf("Node --check tek dosya alır.")
		}
		return nil
	}
	if strings.HasPrefix(args[0], "-") {
		return fmt.Errorf("Node seçeneği izinli değil: %s", args[0])
	}
	return nil
}
