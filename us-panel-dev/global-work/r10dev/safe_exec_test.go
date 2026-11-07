package main

import (
	"context"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func needPython(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 yok")
	}
	return bin
}

func TestSafeExecRejectsNonAllowlistedBinary(t *testing.T) {
	root := t.TempDir()
	for _, argv := range [][]string{{"sh", "-c", "id"}, {"/bin/bash"}, {"curl", "x"}, {"rm", "-rf", "."}} {
		if _, err := newSafeCmd(context.Background(), root, root, argv, time.Second); err == nil {
			t.Fatalf("izinsiz komut kabul edildi: %v", argv)
		}
	}
	for _, raw := range []string{`["sh","-c","id"]`, `["go","env","-w","X=1"]`, `["python3","-c","print(1)"]`, `["node","-e","1"]`, `["go","test","../x"]`} {
		if _, err := terminalArgv(raw, root); err == nil {
			t.Fatalf("terminal izinsiz komutu kabul etti: %s", raw)
		}
	}
}

func TestSafeExecCwdLockBlocksDotDotAndSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "kacis")); err != nil {
		t.Skip("symlink yok")
	}
	if _, err := lockCwd(root, "sub"); err != nil {
		t.Fatalf("proje içi alt dizin reddedildi: %v", err)
	}
	for _, dir := range []string{"..", "sub/../..", filepath.Join(root, ".."), "kacis", filepath.Join(root, "kacis"), outside} {
		if _, err := lockCwd(root, dir); err == nil {
			t.Fatalf("kaçış kabul edildi: %s", dir)
		}
	}
	if _, err := newSafeCmd(context.Background(), root, "kacis", []string{"go", "version"}, time.Second); err == nil {
		t.Fatal("symlink cwd ile komut hazırlandı")
	}
}

func TestSafeExecEnvDoesNotLeakSecrets(t *testing.T) {
	py := needPython(t)
	t.Setenv("US_LICENSE_SECRET", "sizinti-lisans-123")
	t.Setenv("SHOPIFY_WEBHOOK_SECRET", "sizinti-shopify-456")
	t.Setenv("US_MODEL_UPSTREAM_TOKEN", "sizinti-token-789")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "sizinti-aws")
	for _, item := range safeExecEnv() {
		if strings.Contains(item, "sizinti") {
			t.Fatalf("env allowlist sızdırdı: %s", item)
		}
	}
	root := t.TempDir()
	sc, err := newSafeCmd(context.Background(), root, root, []string{py, "-c", "import os\nfor k,v in os.environ.items(): print(k+'='+v)"}, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	state, code, _ := sc.Run()
	if state != "success" || code != 0 {
		t.Fatalf("%s %d %s", state, code, sc.Stderr.String())
	}
	if strings.Contains(sc.Stdout.String(), "sizinti") {
		t.Fatalf("çocuk süreç gizli değeri gördü:\n%s", sc.Stdout.String())
	}
	// Run path uses the same filtered env.
	for _, item := range pythonEnv() {
		if strings.Contains(item, "sizinti") {
			t.Fatalf("Run env sızdırdı: %s", item)
		}
	}
}

func TestRunPathDoesNotLeakSecretsToProgram(t *testing.T) {
	needPython(t)
	t.Setenv("US_LICENSE_SECRET", "sizinti-run-path")
	root := t.TempDir()
	script := "import os,sys\nprint('err-line', file=sys.stderr)\nprint('LEAK' if any('sizinti' in v for v in os.environ.values()) else 'CLEAN')\n"
	if err := os.WriteFile(filepath.Join(root, "main.py"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newView("q-safe")
	defer s.stopProgram()
	line := s.startProgram(root, "")
	if !s.runStarted {
		t.Fatalf("program başlamadı: %s", line)
	}
	if !strings.Contains(line, "CLEAN") || strings.Contains(line, "LEAK") || strings.Contains(line, "US_LICENSE_SECRET") {
		t.Fatalf("Run yolu gizli env sızdırdı:\n%s", line)
	}
	if !strings.Contains(line, "[stderr]") || !strings.Contains(line, "err-line") {
		t.Fatalf("stderr ayrı gösterilmedi:\n%s", line)
	}
}

func TestSafeExecSeparateStreamsAndCap(t *testing.T) {
	py := needPython(t)
	root := t.TempDir()
	sc, err := newSafeCmd(context.Background(), root, root, []string{py, "-c", "import sys\nsys.stdout.write('o'*100000)\nsys.stderr.write('E-ONLY')\nsys.exit(7)"}, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	state, code, _ := sc.Run()
	if state != "failed" || code != 7 {
		t.Fatalf("durum/kod yanlış: %s %d", state, code)
	}
	if got := len(sc.Stdout.String()); got > safeExecMaxOutput {
		t.Fatalf("stdout sınırı aşıldı: %d", got)
	}
	if strings.Contains(sc.Stdout.String(), "E-ONLY") || sc.Stderr.String() != "E-ONLY" {
		t.Fatal("stdout/stderr karıştı")
	}
}

func TestSafeExecStartFailureIsReported(t *testing.T) {
	root := t.TempDir()
	fake := filepath.Join(root, "go")
	if err := os.WriteFile(fake, []byte{0, 1, 2, 3}, 0o755); err != nil {
		t.Fatal(err)
	}
	sc, err := newSafeCmd(context.Background(), root, root, []string{fake}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if state, _, err := sc.Run(); state != "start_failed" || err == nil {
		t.Fatalf("başlatma hatası raporlanmadı: %s %v", state, err)
	}
}

func TestRunLogNotWrittenWhenProgramDidNotStart(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.py"), []byte("print(1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir()) // no python3/go/node: launch cannot start
	s := newView("q-nostart")
	s.root = root
	s.browse = root
	srv := httptest.NewServer(guiMux(&s))
	defer srv.Close()
	defer s.stopProgram()
	snap := guiPost(t, srv.URL+"/api/act", `{"op":"run"}`)
	for _, item := range snap.Journal {
		if strings.Contains(item.Text, "Program çalıştırıldı") {
			t.Fatalf("başlamayan program için başarı günlüğü yazıldı: %+v", snap.Journal)
		}
	}
	found := false
	for _, item := range snap.Journal {
		found = found || strings.Contains(item.Text, "Program başlatılamadı")
	}
	if !found {
		t.Fatalf("başarısızlık günlüğü yok: %+v", snap.Journal)
	}
}
