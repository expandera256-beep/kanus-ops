//go:build linux

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestQAR13LinuxLandlockSiblingRead(t *testing.T) {
	if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" && runtime.GOARCH != "riscv64" {
		t.Skip("Landlock syscall mapping not validated for this architecture")
	}
	abi, _, errno := syscall.Syscall6(444, 0, 0, 1, 0, 0, 0)
	if errno != 0 || abi < 1 {
		t.Skipf("Landlock ABI unavailable: errno=%d", errno)
	}
	if _, e := exec.LookPath("python3"); e != nil {
		t.Skip("python3 unavailable: cannot run positive-control fixture")
	}
	parent := t.TempDir()
	root := filepath.Join(parent, "workspace")
	sibling := filepath.Join(parent, "sibling")
	if os.Mkdir(root, 0700) != nil || os.Mkdir(sibling, 0700) != nil {
		t.Fatal("fixture directories failed")
	}
	inside := filepath.Join(root, "inside.txt")
	outside := filepath.Join(sibling, "outside.txt")
	if os.WriteFile(inside, []byte("QA_R13_INSIDE_ONLY"), 0600) != nil || os.WriteFile(outside, []byte("QA_R13_SYNTHETIC_OUTSIDE"), 0600) != nil {
		t.Fatal("fixture files failed")
	}
	script := filepath.Join(root, "main.py")
	py := `import pathlib,sys
if pathlib.Path(sys.argv[1]).read_text() != "QA_R13_INSIDE_ONLY": raise RuntimeError("inside fixture mismatch")
print("QA_INSIDE_CONTROL_OK",flush=True)
try:
 pathlib.Path(sys.argv[2]).read_text()
 print("QA_SIBLING_READ_ALLOWED",flush=True)
except PermissionError:
 print("QA_SIBLING_READ_DENIED",flush=True)
`
	if os.WriteFile(script, []byte(py), 0600) != nil {
		t.Fatal("fixture script failed")
	}
	out, state, _, err := safeRun(context.Background(), root, root, []string{"python3", script, inside, outside}, 10*time.Second, nil)
	control := strings.Contains(out, "QA_INSIDE_CONTROL_OK")
	denied := strings.Contains(out, "QA_SIBLING_READ_DENIED")
	allowed := strings.Contains(out, "QA_SIBLING_READ_ALLOWED")
	t.Logf("kernel_landlock_abi=%d inside_control=%t sibling_read_denied=%t sibling_read_allowed=%t state=%s", abi, control, denied, allowed, state)
	if err != nil || !control {
		t.Fatal("positive control did not execute; denial cannot be counted as PASS")
	}
	if !denied || allowed {
		t.Fatal("F01 Linux OPEN: child read synthetic sibling; .13 enforcement required")
	}
}
