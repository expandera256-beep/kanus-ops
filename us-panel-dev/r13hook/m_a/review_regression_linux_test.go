//go:build linux

package main

// Regression tests for PANEL DEV §2 review (06.11.2026), adapted from the
// reviewer's probes (panel-evidence/review/zz_review_probe_test.go.txt).

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func revAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	raw, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	return err == nil && !strings.Contains(string(raw), ") Z ") && syscall.Kill(pid, 0) == nil
}

// HIGH-3 (P1): a file merely named "go" at an arbitrary absolute path must not run.
func TestReviewH3AbsoluteFakeBinaryRejected(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(t.TempDir(), "go")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\necho PWNED_NOT_GO\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if sc, err := newSafeCmd(context.Background(), root, root, []string{bin}, 5*time.Second); err == nil {
		sc.Run()
		t.Fatalf("sahte mutlak 'go' kabul edildi; stdout=%q", sc.Stdout.String())
	}
	inRoot := filepath.Join(root, "python3")
	_ = os.WriteFile(inRoot, []byte("#!/bin/sh\necho PWNED\n"), 0o755)
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
	if sc, err := newSafeCmd(context.Background(), root, root, []string{"python3", "-c", "print(1)"}, 5*time.Second); err == nil {
		sc.Run()
		if strings.Contains(sc.Stdout.String(), "PWNED") {
			t.Fatal("PATH üzerinden çalışma alanı içindeki çalıştırıcı seçildi")
		}
	}
}

// HIGH-2 (P2): Terminal must refuse -toolexec/-exec and friends.
func TestReviewH2TerminalRejectsExecFlags(t *testing.T) {
	root := t.TempDir()
	for _, raw := range []string{
		`["go","build","-toolexec=sh -c 'touch PWNED_SHELL'"]`,
		`["go","build","-toolexec","sh"]`,
		`["go","test","-exec","bash","./..."]`,
		`["go","run","-exec=sh","main.go"]`,
		`["go","build","-overlay=o.json"]`,
		`["go","build","-modfile=x.mod"]`,
		`["go","build","-ldflags=-X main.a=b"]`,
		`["go","build","-gcflags=all=-N"]`,
		`["go","test","-vet=off","./..."]`,
		`["go","build","-a b"]`,
		`["go","vet","-vettool=sh"]`,
	} {
		if argv, err := terminalArgv(raw, root); err == nil {
			t.Fatalf("tehlikeli bayrak kabul edildi: %q", argv)
		}
	}
	for _, raw := range []string{`["go","test","-v","-race","-count=1","-run","TestX","./..."]`, `["go","build","./..."]`, `["go","vet","./..."]`, `["go","run","."]`, `["go","version"]`} {
		if _, err := terminalArgv(raw, root); err != nil {
			t.Fatalf("güvenli komut reddedildi %s: %v", raw, err)
		}
	}
	t.Setenv("GOFLAGS", "-toolexec=/bin/false")
	t.Setenv("CC", "/tmp/evil-cc")
	for _, kv := range safeExecEnv() {
		if strings.HasPrefix(kv, "GOFLAGS=") || strings.HasPrefix(kv, "CC=") {
			t.Fatalf("derleme davranışı env'i geçti: %s", kv)
		}
	}
}

// HIGH-1 (P4/P9): project .venv interpreter must not be executed with the
// panel env (or at all without explicit opt-in) by Run, Tests or syntax check.
func TestReviewH1VenvProbeDoesNotLeak(t *testing.T) {
	t.Setenv("REVIEW_FAKE_SECRET", "s3cr3t-review")
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.py"), []byte("print(1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".venv", "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	dump := filepath.Join(t.TempDir(), "envdump.txt")
	_ = os.WriteFile(filepath.Join(root, ".venv", "bin", "python"), []byte("#!/bin/sh\nenv >> "+dump+"\necho 'Python 3.12.0'\n"), 0o755)
	_, _, _ = findPython(root, root)
	_, _ = detectLaunch(root)
	_ = runProjectTestsWithContext(context.Background(), root, 10*time.Second)
	_, _ = pyParseCheck(root, filepath.Join(root, "main.py"))
	s := newView("q-h1")
	_ = s.startProgram(root, "")
	s.stopProgram()
	b, _ := os.ReadFile(dump)
	if strings.Contains(string(b), "s3cr3t-review") {
		t.Fatalf("proje .venv yorumlayıcısı panel sırlarını gördü (%d kez)", strings.Count(string(b), "s3cr3t-review"))
	}
	if len(b) > 0 {
		t.Fatalf("proje .venv yorumlayıcısı onaysız çalıştırıldı")
	}
}

// MEDIUM-1 (P5): a setsid grandchild must die on timeout and cancel.
func TestReviewM1SetsidGrandchildKilled(t *testing.T) {
	py := needPython(t)
	for _, mode := range []string{"timeout", "cancel"} {
		root := t.TempDir()
		s := "import os,subprocess,time\np=subprocess.Popen(['sleep','37'])\nq=subprocess.Popen(['sleep','38'],start_new_session=True)\nopen('pids.txt','w').write('%d %d'%(p.pid,q.pid))\ntime.sleep(60)\n"
		_ = os.WriteFile(filepath.Join(root, "tree.py"), []byte(s), 0o644)
		sc, err := newSafeCmd(context.Background(), root, root, []string{py, "tree.py"}, 1500*time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		if err := sc.Start(); err != nil {
			t.Fatal(err)
		}
		var pids []int
		for i := 0; i < 100 && len(pids) < 2; i++ {
			time.Sleep(30 * time.Millisecond)
			raw, _ := os.ReadFile(filepath.Join(root, "pids.txt"))
			pids = nil
			for _, f := range strings.Fields(string(raw)) {
				n, _ := strconv.Atoi(f)
				pids = append(pids, n)
			}
		}
		if len(pids) < 2 {
			t.Fatal("pid yok")
		}
		if mode == "cancel" {
			sc.Cancel()
		}
		sc.Wait()
		dead := false
		for i := 0; i < 60 && !dead; i++ {
			dead = !revAlive(pids[0]) && !revAlive(pids[1])
			time.Sleep(50 * time.Millisecond)
		}
		for _, p := range pids {
			_ = syscall.Kill(p, syscall.SIGKILL)
		}
		if !dead {
			t.Fatalf("%s: setsid torunu hayatta kaldı", mode)
		}
	}
}

// MEDIUM-4 (P7): a hard link to an outside file must not be readable.
func TestReviewM4HardlinkReadRejected(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(t.TempDir(), "outside-secret.txt")
	_ = os.WriteFile(out, []byte("OUTSIDE-SECRET"), 0o600)
	if err := os.Link(out, filepath.Join(root, "hl.txt")); err != nil {
		t.Skip(err)
	}
	ft, _ := newFileTools(root)
	if b, err := ft.Read("hl.txt"); err == nil {
		t.Fatalf("hard link ile kök dışı okundu: %q", b)
	}
	if err := ft.Write("hl.txt", []byte("x")); err == nil {
		t.Fatal("hard link üzerine yazım kabul edildi")
	}
	if raw, _ := os.ReadFile(out); string(raw) != "OUTSIDE-SECRET" {
		t.Fatal("dış dosya değişti")
	}
}

func TestReviewTestsPathHasLimits(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go yok")
	}
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "go.mod"), []byte("module p\n\ngo 1.22\n"), 0o644)
	src := "package p\n\nimport (\n\t\"os\"\n\t\"testing\"\n\n\t\"syscall\"\n)\n\nfunc TestLimit(t *testing.T) {\n\tvar r syscall.Rlimit\n\t_ = syscall.Getrlimit(syscall.RLIMIT_CPU, &r)\n\tos.WriteFile(\"lim.txt\", []byte(func() string { if r.Cur == 600 { return \"LIMITED\" }; return \"UNLIMITED\" }()), 0o644)\n}\n"
	_ = os.WriteFile(filepath.Join(root, "p_test.go"), []byte(src), 0o644)
	_ = runProjectTestsWithContext(context.Background(), root, 2*time.Minute)
	raw, _ := os.ReadFile(filepath.Join(root, "lim.txt"))
	if string(raw) != "LIMITED" {
		t.Fatalf("Testler yolu kaynak sınırı uygulamadı: %q", raw)
	}
}

func TestReviewEditorRejectsHardlink(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(t.TempDir(), "o.txt")
	_ = os.WriteFile(out, []byte("OUTSIDE-SECRET"), 0o600)
	if err := os.Link(out, filepath.Join(root, "hl.txt")); err != nil {
		t.Skip(err)
	}
	s := newView("q-hl")
	s.root = root
	if text, ok, _ := s.localTextWithLimit("hl.txt", 8000); ok || strings.Contains(text, "OUTSIDE") {
		t.Fatalf("editör hard link ile dış dosyayı okudu: %q", text)
	}
}

// New finding (1): symlink swap race on an intermediate directory.
func TestReviewSymlinkSwapRaceNoEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	_ = os.WriteFile(filepath.Join(outside, "f.txt"), []byte("OUTSIDE"), 0o644)
	real := filepath.Join(root, "real")
	_ = os.MkdirAll(real, 0o755)
	_ = os.WriteFile(filepath.Join(real, "f.txt"), []byte("INSIDE"), 0o644)
	d := filepath.Join(root, "d")
	_ = os.Rename(real, d)
	ft, _ := newFileTools(root)
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
			}
			_ = os.Rename(d, real)
			_ = os.Symlink(outside, d)
			_ = os.Remove(d)
			_ = os.Rename(real, d)
		}
	}()
	escaped, ok := 0, 0
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if b, err := ft.Read("d/f.txt"); err == nil {
			ok++
			if string(b) == "OUTSIDE" {
				escaped++
			}
		}
	}
	close(stop)
	<-done
	if escaped != 0 {
		t.Fatalf("yarışta %d/%d okuma kök dışından geldi", escaped, ok)
	}
}
