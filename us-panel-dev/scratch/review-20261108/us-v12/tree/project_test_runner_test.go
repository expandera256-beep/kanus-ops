package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProjectRunnerDoesNotExposeInheritedSecrets(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "private-api-key")
	t.Setenv("US_ACCOUNT_SESSION", "private-session")
	t.Setenv("SHOPIFY_API_SECRET", "private-shopify-secret")
	for _, item := range projectTestEnv() {
		if strings.Contains(item, "private-") {
			t.Fatalf("credential inherited: %s", strings.SplitN(item, "=", 2)[0])
		}
	}
}

func TestProjectRunnerReportsInvalidAndEmptyWorkspace(t *testing.T) {
	root := t.TempDir()
	if got := runProjectTests(filepath.Join(root, "missing")); !strings.HasPrefix(got, "Test çalıştırılamadı") {
		t.Fatal(got)
	}
	if _, _, ok := findPython(root, root); !ok {
		t.Skip("Python unavailable")
	}
	if got := runProjectTests(root); !strings.HasPrefix(got, "Test bulunamadı") {
		t.Fatal(got)
	}
}

func TestProjectRunnerRunsConfiguredNodeTest(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node unavailable")
	}
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm unavailable")
	}
	root := t.TempDir()
	for code, want := range map[int]string{0: "Test doğrulanmadı", 7: "Test başarısız"} {
		pkg := `{"scripts":{"test":"node -e \"process.exit(` + string(rune('0'+code)) + `)\""}}`
		if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(pkg), 0600); err != nil {
			t.Fatal(err)
		}
		got := runProjectTests(root)
		if !strings.HasPrefix(got, want) || !strings.Contains(got, "--ignore-scripts") {
			t.Fatal(got)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"scripts":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if got := runProjectTests(root); !strings.HasPrefix(got, "Test bulunamadı") {
		t.Fatal(got)
	}
}

func writeSlowPythonTest(t *testing.T, root string) {
	t.Helper()
	if _, _, ok := findPython(root, root); !ok {
		t.Skip("Python unavailable")
	}
	// A surviving descendant would create escaped.txt after the parent is killed.
	code := "import unittest,subprocess,sys,time,pathlib\nclass Slow(unittest.TestCase):\n def test_slow(self):\n  subprocess.Popen([sys.executable,'-c',\"import time,pathlib; time.sleep(1.2); pathlib.Path('escaped.txt').write_text('survived')\"])\n  pathlib.Path('started.txt').write_text('started')\n  time.sleep(30)\n"
	if err := os.WriteFile(filepath.Join(root, "test_slow.py"), []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestProjectTimeoutKillsDescendantProcesses(t *testing.T) {
	root := t.TempDir()
	writeSlowPythonTest(t, root)
	start := time.Now()
	report := runProjectTestsWithContext(context.Background(), root, 500*time.Millisecond)
	if !strings.HasPrefix(report, "Test zaman aşımı") || time.Since(start) > 5*time.Second {
		t.Fatal(report)
	}
	if _, err := os.Stat(filepath.Join(root, "started.txt")); err != nil {
		t.Fatal("test process never started", err)
	}
	time.Sleep(1300 * time.Millisecond)
	if _, err := os.Stat(filepath.Join(root, "escaped.txt")); !os.IsNotExist(err) {
		t.Fatal("descendant survived timeout")
	}
}

func TestPanelCancelsActualTestAndPersistsCancelledState(t *testing.T) {
	t.Setenv("US_DESK", filepath.Join(t.TempDir(), "desk.json"))
	s := newView("q8")
	s.root = t.TempDir()
	writeSlowPythonTest(t, s.root)
	h := guiMux(&s)
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "http://127.0.0.1:4477/api/act", bytes.NewBufferString(`{"op":"project-tests"}`)))
		done <- w
	}()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(s.root, "started.txt")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("test did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "http://127.0.0.1:4477/api/act", bytes.NewBufferString(`{"op":"goto","path":".."}`)))
	if w.Code != http.StatusConflict {
		t.Fatalf("workspace mutation during test: %d", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "http://127.0.0.1:4477/api/act", bytes.NewBufferString(`{"op":"tests-stop"}`)))
	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case result := <-done:
		var snap guiSnap
		if err := json.Unmarshal(result.Body.Bytes(), &snap); err != nil {
			t.Fatal(err)
		}
		if snap.Testing || !strings.HasPrefix(snap.TestReport, "Test iptal edildi") {
			t.Fatal(result.Body.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("test cancellation blocked")
	}
	records := listTaskRecords(s.root)
	if len(records) != 1 || records[0].Status != "cancelled" {
		t.Fatalf("history=%+v", records)
	}
	time.Sleep(1300 * time.Millisecond)
	if _, err := os.Stat(filepath.Join(s.root, "escaped.txt")); !os.IsNotExist(err) {
		t.Fatal("descendant survived cancellation")
	}
}

func TestProjectRunnerRequiresRealGoAndNodeTestEvidence(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "go.mod"), []byte("module offlinecheck\ngo 1.22\n"), 0600)
	os.WriteFile(filepath.Join(root, "answer.go"), []byte("package answer\nfunc Value()int{return 42}\n"), 0600)
	if got := runProjectTests(root); !strings.HasPrefix(got, "Test bulunamadı") {
		t.Fatal(got)
	}
	os.WriteFile(filepath.Join(root, "answer_test.go"), []byte("package answer\nimport \"testing\"\nfunc TestValue(t *testing.T){if Value()!=42{t.Fatal(Value())}}\n"), 0600)
	if got := runProjectTests(root); !strings.HasPrefix(got, "Test komutu başarılı") {
		t.Fatal(got)
	}
	node := t.TempDir()
	os.WriteFile(filepath.Join(node, "package.json"), []byte(`{"scripts":{"test":"node --test test.cjs"}}`), 0600)
	os.WriteFile(filepath.Join(node, "test.cjs"), []byte("const {test}=require('node:test');const assert=require('node:assert/strict');test('proof',()=>assert.equal(1,1));\n"), 0600)
	if got := runProjectTests(node); !strings.HasPrefix(got, "Test komutu başarılı") {
		t.Fatal(got)
	}
	for _, text := range []string{"Ran 0 tests\nOK", "Ran 3 tests\nOK (skipped=3)", "# pass 0", "echo done"} {
		if projectTestEvidence(text) {
			t.Fatal("fake test pass", text)
		}
	}
}
