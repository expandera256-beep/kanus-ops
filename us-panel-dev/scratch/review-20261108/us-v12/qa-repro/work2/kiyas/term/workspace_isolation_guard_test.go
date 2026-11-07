package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FINDING-01 cheap-hardening guards. These do NOT close the finding: project
// code under the same OS user can still open sibling paths directly.

func TestFileToolsRefuseSiblingWorkspace(t *testing.T) {
	parent := t.TempDir()
	a, b := filepath.Join(parent, "user-a"), filepath.Join(parent, "user-b")
	os.Mkdir(a, 0o700)
	os.Mkdir(b, 0o700)
	os.WriteFile(filepath.Join(b, "s.txt"), []byte("SYNTH"), 0o600)
	os.Symlink(b, filepath.Join(a, "link"))
	ft, err := newFileTools(a)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"../user-b/s.txt", filepath.Join(b, "s.txt"), "link/s.txt"} {
		if body, err := ft.Read(p); err == nil {
			t.Fatalf("sibling read via %q allowed: %q", p, body)
		}
		if err := ft.Write(p, []byte("x")); err == nil {
			t.Fatalf("sibling write via %q allowed", p)
		}
	}
}

func TestSafeExecEnvDoesNotExposeWorkspaceRoots(t *testing.T) {
	for _, k := range []string{"US_PROJECT_ROOT", "US_WORKSPACE", "US_ROOT", "US_DATA_DIR", "PWD", "OLDPWD"} {
		t.Setenv(k, "/synthetic/other-workspace")
	}
	for _, kv := range safeExecEnv() {
		if strings.Contains(kv, "/synthetic/other-workspace") {
			t.Fatalf("child env exposes workspace path: %s", kv)
		}
	}
}
