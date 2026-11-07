package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyntaxFailurePreventsEntireMultiFileCommit(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("Python parser unavailable")
	}
	root := t.TempDir()
	old := "old\n"
	if err := os.WriteFile(filepath.Join(root, "keep.txt"), []byte(old), 0600); err != nil {
		t.Fatal(err)
	}
	files := []proposedFile{{Path: "keep.txt", Existed: true, Old: old, OldHash: hashContent([]byte(old)), New: "new\n"}, {Path: "server.py", New: "self.wfile.write(data)):\n"}}
	_, err := applyProposalBatch(root, files, writeModelFile)
	if err == nil || !strings.Contains(err.Error(), "sözdizimi") {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(root, "keep.txt"))
	if string(got) != old {
		t.Fatal("first file committed before syntax validation")
	}
	if _, err = os.Stat(filepath.Join(root, "server.py")); !os.IsNotExist(err) {
		t.Fatal("invalid Python file written")
	}
	if pendingTransaction(root) {
		t.Fatal("syntax error left a transaction journal")
	}
}
