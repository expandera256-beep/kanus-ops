package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MEDIUM-5 (P8): a hand-edited task state must not apply edits without approval.
func TestReviewM5TamperedTaskStateNeedsApproval(t *testing.T) {
	t.Setenv("US_TASK_STATE_KEY_FILE", filepath.Join(t.TempDir(), "k"))
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "go.mod"), []byte("module p\n\ngo 1.21\n"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main(){}\n"), 0o644)
	evil := "package main\n\nfunc main() {}\n\n// INJECTED_BY_TAMPERED_STATE\n"
	path, _ := taskStatePath(root, "t1")
	_ = os.MkdirAll(filepath.Dir(path), 0o700)
	forged := `{"version":1,"id":"t1","root":` + jsonString(root) + `,"phase":"applying","result":{"status":"running"},"pendingEdits":[{"path":"main.go","content":` + jsonString(evil) + `,"hash":"` + hashContent([]byte(evil)) + `"}]}`
	_ = os.WriteFile(path, []byte(forged), 0o600)
	l := &repairLoop{Root: root, Endpoint: "http://127.0.0.1:1", Approve: func(int, []repairEdit) bool { return false }}
	_, _ = runPersistentRepair(context.Background(), l, "t1", false)
	b, _ := os.ReadFile(filepath.Join(root, "main.go"))
	if strings.Contains(string(b), "INJECTED") {
		t.Fatal("kurcalanmış görev durumu onaysız düzeltme uyguladı")
	}
	// A validly signed state with pending edits still needs approval on resume.
	st := &taskState{ID: "t2", Root: root, Phase: "applying", PendingEdits: []pendingEdit{{Path: "main.go", Content: evil, Hash: hashContent([]byte(evil))}}}
	if err := saveTaskState(st); err != nil {
		t.Fatal(err)
	}
	asked := false
	l2 := &repairLoop{Root: root, Endpoint: "http://127.0.0.1:1", Approve: func(int, []repairEdit) bool { asked = true; return false }}
	_, _ = runPersistentRepair(context.Background(), l2, "t2", false)
	b, _ = os.ReadFile(filepath.Join(root, "main.go"))
	if !asked || strings.Contains(string(b), "INJECTED") {
		t.Fatalf("devamda bekleyen düzeltme yeniden onaya sunulmadı (asked=%v)", asked)
	}
}

// LOW-1 (P7b): Windows ADS, drive-relative, reserved names and \\?\ prefixes.
func TestReviewWindowsPathFormsRejected(t *testing.T) {
	ft, _ := newFileTools(t.TempDir())
	for _, p := range []string{"a.txt:stream", `C:evil.txt`, "sub/x.txt::$DATA", "CON", "nul.txt", "sub/COM1", "LPT9.log", `\\?\C:\x`, `\\server\share\x`, "trail.", "sub /x.txt"} {
		if _, err := ft.resolve(p); err == nil {
			t.Fatalf("Windows yol biçimi kabul edildi: %q", p)
		}
	}
	if _, err := ft.resolve("sub/normal.txt"); err != nil {
		t.Fatal(err)
	}
}

func TestReviewNodeArgAllowlist(t *testing.T) {
	root := t.TempDir()
	for _, raw := range []string{`["node","--test","-p","1"]`, `["node","--test","--import=./a.js"]`, `["node","--test","--import","./a.js"]`, `["node","--check","a.js","--require=x"]`, `["node","--version","-e","1"]`, `["node","--print","1"]`, `["node","-r","x","a.js"]`} {
		if argv, err := terminalArgv(raw, root); err == nil {
			t.Fatalf("Node seçeneği kabul edildi: %q", argv)
		}
	}
	for _, raw := range []string{`["node","--test"]`, `["node","--test","t.test.js"]`, `["node","--check","a.js"]`, `["node","a.js","--port=1"]`, `["node","--version"]`} {
		if _, err := terminalArgv(raw, root); err != nil {
			t.Fatalf("güvenli Node komutu reddedildi %s: %v", raw, err)
		}
	}
}
