package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func awaitTerminal(t *testing.T, tr *panelTerminal) terminalState {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		v := tr.snapshot()
		if v.State != "running" {
			return v
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("terminal did not finish")
	return terminalState{}
}
func TestTerminalRealOutputExitTimeoutCancelAndIsolation(t *testing.T) {
	for _, mode := range []string{"exit", "timeout", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			script := "import os,sys\nprint(os.getcwd(),flush=True)\nprint('stderr-proof',file=sys.stderr,flush=True)\nassert 'US_KEY' not in os.environ\n"
			if mode == "exit" {
				script += "sys.exit(7)\n"
			} else {
				script += "import time; time.sleep(20)\n"
			}
			os.WriteFile(filepath.Join(root, "main.py"), []byte(script), 0600)
			t.Setenv("US_KEY", "not-for-child")
			tr := &panelTerminal{}
			timeout := 2 * time.Second
			if mode == "timeout" {
				timeout = 100 * time.Millisecond
			}
			if err := tr.start(root, `["python3","main.py"]`, timeout); err != nil {
				t.Fatal(err)
			}
			if mode == "cancel" {
				time.Sleep(50 * time.Millisecond)
				tr.stop()
			}
			v := awaitTerminal(t, tr)
			if mode == "exit" {
				if v.ExitCode != 7 || v.State != "failed" || !strings.Contains(v.Stdout, root) || !strings.Contains(v.Stderr, "stderr-proof") {
					t.Fatal(v)
				}
			} else if v.State != map[string]string{"timeout": "timeout", "cancel": "cancelled"}[mode] {
				t.Fatal(v)
			}
			if new(panelTerminal).snapshot().Stdout != "" {
				t.Fatal("cross-terminal output")
			}
		})
	}
}
func TestTerminalRejectsShellEscapeAndMissingRuntime(t *testing.T) {
	root := t.TempDir()
	for _, raw := range []string{`["sh","-c","echo bad"]`, `["python3","../secret.py"]`, `["node","--eval","1"]`, `["go","test","-C","../other"]`, `["go","test",";whoami"]`, `["python3","/tmp/a.py"]`} {
		if _, err := terminalArgv(raw, root); err == nil {
			t.Fatal(raw)
		}
	}
}
func TestGoLaunchDetection(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "go.mod"), []byte("module offline\ngo 1.20\n"), 0600)
	os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main(){}\n"), 0600)
	found, ok := detectLaunch(root)
	if !ok || found.Label != "go run ." || found.Dir != root {
		t.Fatal(found, ok)
	}
}
