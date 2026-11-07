//go:build linux

package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestSafeExecTimeoutKillsProcessGroup(t *testing.T) {
	py := needPython(t)
	root := t.TempDir()
	pidFile := filepath.Join(root, "child.pid")
	script := "import subprocess,sys,time\np=subprocess.Popen([sys.executable,'-c','import time; time.sleep(60)'])\nopen('child.pid','w').write(str(p.pid))\nprint('ready',flush=True)\ntime.sleep(60)"
	for _, mode := range []string{"timeout", "cancel"} {
		_ = os.Remove(pidFile)
		timeout := 30 * time.Second
		if mode == "timeout" {
			timeout = 1500 * time.Millisecond
		}
		sc, err := newSafeCmd(context.Background(), root, root, []string{py, "-c", script}, timeout)
		if err != nil {
			t.Fatal(err)
		}
		if err := sc.Start(); err != nil {
			t.Fatal(err)
		}
		deadline := time.Now().Add(10 * time.Second)
		var childPid int
		for time.Now().Before(deadline) {
			if raw, err := os.ReadFile(pidFile); err == nil && len(raw) > 0 {
				childPid, _ = strconv.Atoi(string(raw))
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		if childPid == 0 {
			t.Fatal("alt süreç pid yok")
		}
		if mode == "cancel" {
			sc.Cancel()
		}
		start := time.Now()
		state, _, _ := sc.Wait()
		want := map[string]string{"timeout": "timeout", "cancel": "cancelled"}[mode]
		if state != want {
			t.Fatalf("%s: durum %s", mode, state)
		}
		if time.Since(start) > 8*time.Second {
			t.Fatalf("%s: iptal gecikti", mode)
		}
		alive := true
		for i := 0; i < 100 && alive; i++ {
			// Reaped by init or zombie: both mean killed; check /proc state.
			raw, err := os.ReadFile("/proc/" + strconv.Itoa(childPid) + "/stat")
			alive = err == nil && !strings.Contains(string(raw), ") Z ") && syscall.Kill(childPid, 0) == nil
			if alive {
				time.Sleep(50 * time.Millisecond)
			}
		}
		if alive {
			_ = syscall.Kill(childPid, syscall.SIGKILL)
			t.Fatalf("%s: torun süreç öldürülmedi (pid %d)", mode, childPid)
		}
	}
}
