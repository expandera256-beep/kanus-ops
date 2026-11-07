//go:build linux

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestPyCompileRunsIsolated(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 yok")
	}
	_ = py
	root := t.TempDir()
	marker := filepath.Join(t.TempDir(), "shadow.txt")
	os.WriteFile(filepath.Join(root, "py_compile.py"), []byte("open("+quotePy(marker)+",'w').write('x')\n"), 0o644)
	os.WriteFile(filepath.Join(root, "compileall.py"), []byte("open("+quotePy(marker)+",'w').write('x')\n"), 0o644)
	os.WriteFile(filepath.Join(root, "a.py"), []byte("x = 1\n"), 0o644)
	for _, raw := range []string{`["python3","-m","py_compile","a.py"]`, `["python3","-m","compileall","a.py"]`} {
		argv, err := terminalArgv(raw, root)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Contains(argv, "-I") {
			t.Fatalf("-I yok: %q", argv)
		}
		safeRun(context.Background(), root, root, argv, 30*time.Second, nil)
		if _, err := os.Stat(marker); err == nil {
			t.Fatalf("projedeki sahte modül çalıştı: %q", argv)
		}
	}
}

func quotePy(s string) string { return "'" + s + "'" }
