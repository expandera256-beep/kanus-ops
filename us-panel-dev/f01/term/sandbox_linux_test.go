//go:build linux

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestF01LandlockSiblingHomeAndWrite(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "project")
	sibling := filepath.Join(base, "sibling")
	home := filepath.Join(base, "home")
	for _, d := range []string{root, sibling, home} {
		if err := os.Mkdir(d, 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range []string{sibling, home} {
		if err := os.WriteFile(filepath.Join(d, "synthetic"), []byte("F01-FAKE"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", home)
	py, err := trustedBin("python3", root)
	if err != nil {
		t.Fatal(err)
	}
	script := `import errno,os,sys
for path in sys.argv[1:3]:
 try:
  open(path).read()
  raise AssertionError("OUTSIDE_READ_ALLOWED")
 except OSError as e:
  assert e.errno == errno.EACCES,repr(e)
  print("EACCES",flush=True)
try:
 open(sys.argv[3],"w").write("bad")
 raise AssertionError("OUTSIDE_WRITE_ALLOWED")
except OSError as e:
 assert e.errno == errno.EACCES,repr(e)
open("inside.txt","w").write("ok")
print("F01_OK",flush=True)
`
	out, state, _, err := safeRun(context.Background(), root, root, []string{py, "-c", script, filepath.Join(sibling, "synthetic"), filepath.Join(home, "synthetic"), filepath.Join(sibling, "write")}, 30*time.Second, nil)
	if err != nil || state != "success" || strings.Count(out, "EACCES") != 2 || !strings.Contains(out, "F01_OK") {
		t.Fatalf("state=%s err=%v out=%s", state, err, out)
	}
	if b, err := os.ReadFile(filepath.Join(sibling, "synthetic")); err != nil || string(b) != "F01-FAKE" {
		t.Fatal("ana süreç kısıtlandı", err)
	}
	t.Log("sibling=EACCES home=EACCES outside_write=EACCES inside_write=OK")
}

func TestF01LandlockGoTest(t *testing.T) {
	root := t.TempDir()
	for n, s := range map[string]string{"go.mod": "module f01fixture\n\ngo 1.22\n", "sum_test.go": "package fixture\nimport \"testing\"\nfunc TestSum(t *testing.T) { if 1+1!=2 { t.Fatal(\"sum\") } }\n"} {
		if err := os.WriteFile(filepath.Join(root, n), []byte(s), 0600); err != nil {
			t.Fatal(err)
		}
	}
	out, state, _, err := safeRun(context.Background(), root, root, []string{"go", "test", "./..."}, 2*time.Minute, nil, "GOPROXY=off", "GOSUMDB=off")
	if err != nil || state != "success" {
		t.Fatalf("state=%s err=%v out=%s", state, err, out)
	}
	t.Log(out)
}

func TestF01LandlockABIAndUnavailableVisible(t *testing.T) {
	abi, err := landlockABI()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("kernel Landlock ABI=%d", abi)
	old := probeLandlockABI
	defer func() { probeLandlockABI = old }()
	for _, e := range []error{syscall.ENOSYS, syscall.EOPNOTSUPP} {
		probeLandlockABI = func() (int, error) { return 0, e }
		root := t.TempDir()
		sc, err := newSafeCmd(context.Background(), root, root, []string{"go", "version"}, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		err = sc.Start()
		if err == nil || !strings.Contains(err.Error(), "UYARI") || !strings.Contains(err.Error(), "komut çalıştırılmadı") || !errors.Is(err, e) || sc.Cmd.Process != nil {
			t.Fatalf("sessiz fallback/uyarısız: %v", err)
		}
	}
	probeLandlockABI = func() (int, error) { return 0, nil }
	root := t.TempDir()
	sc, err := newSafeCmd(context.Background(), root, root, []string{"go", "version"}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := sc.Start(); err == nil || !strings.Contains(err.Error(), "UYARI") {
		t.Fatal("ABI<1 uyarısız fallback", err)
	}
}

func TestF01PackUsesSandbox(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "synthetic.txt")
	if err := os.WriteFile(outside, []byte("F01_BYPASS_READ"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.spec"), []byte("# synthetic"), 0600); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "pyinstaller")
	script := "#!/bin/sh\ncat '" + outside + "'\nmkdir -p dist\nprintf MZ > dist/app.exe\n"
	if err := os.WriteFile(bin, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("US_PYINSTALLER", bin)
	got := buildSingleExe(root)
	if strings.Contains(got, "F01_BYPASS_READ") || !strings.Contains(got, "Tek exe yazıldı") {
		t.Fatalf("pack izolasyon atlaması: %s", got)
	}
}

func TestF01ToolDirectoryNotGranted(t *testing.T) {
	root := t.TempDir()
	toolHome := t.TempDir()
	t.Setenv("HOME", toolHome)
	secret := filepath.Join(toolHome, "synthetic.txt")
	if err := os.WriteFile(secret, []byte("F01_FAKE_ADJACENT"), 0600); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(toolHome, "rea")
	script := "#!/bin/sh\nif cat '" + secret + "'; then echo ADJACENT_READ_ALLOWED; exit 1; else echo TOOL_FILE_ONLY_OK; fi\n"
	if err := os.WriteFile(bin, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("US_REA", bin)
	out, state, _, err := safeRun(context.Background(), root, root, []string{bin}, time.Second, nil)
	if err != nil || state != "success" || !strings.Contains(out, "TOOL_FILE_ONLY_OK") || strings.Contains(out, "F01_FAKE_ADJACENT") {
		t.Fatalf("bin üst dizini açıldı state=%s err=%v out=%s", state, err, out)
	}
}

func TestF01UnsupportedWarningReachesTerminal(t *testing.T) {
	old := probeLandlockABI
	defer func() { probeLandlockABI = old }()
	probeLandlockABI = func() (int, error) { return 0, syscall.ENOSYS }
	var term panelTerminal
	err := term.start(t.TempDir(), `["go","version"]`, time.Second)
	if err == nil || !strings.Contains(err.Error(), "UYARI") || !strings.Contains(err.Error(), "Landlock") {
		t.Fatalf("panel uyarıyı gizliyor: %v", err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.py"), []byte("print('MUST_NOT_RUN')"), 0600); err != nil {
		t.Fatal(err)
	}
	s := &session{}
	if out := s.startProgram(root, ""); !strings.Contains(out, "UYARI") || !strings.Contains(out, "Landlock") {
		t.Fatalf("Çalıştır uyarıyı gizliyor: %s", out)
	}
	if out := runProjectTests(root); !strings.Contains(out, "UYARI") || !strings.Contains(out, "Landlock") {
		t.Fatalf("Proje testleri uyarıyı gizliyor: %s", out)
	}
}
