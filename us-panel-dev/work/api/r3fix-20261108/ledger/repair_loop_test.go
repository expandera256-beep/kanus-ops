package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const brokenMain = "package main\n\nimport \"fmt\"\n\nfunc add(a, b int) int {\n\treturn a + c\n}\n\nfunc main() { fmt.Println(add(1, 2)) }\n"
const fixedMain = "package main\n\nimport \"fmt\"\n\nfunc add(a, b int) int {\n\treturn a + b\n}\n\nfunc main() { fmt.Println(add(1, 2)) }\n"
const addTest = "package main\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n\tif add(2, 3) != 5 {\n\t\tt.Fatal(add(2, 3))\n\t}\n}\n"

func brokenGoProject(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go yok")
	}
	root := t.TempDir()
	for name, body := range map[string]string{"go.mod": "module broken\n\ngo 1.22\n", "main.go": brokenMain, "main_test.go": addTest} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// mockModel is an OpenAI-compatible /v1/chat/completions mock.
func mockModel(t *testing.T, answer func(prompt string, n int) string) (*httptest.Server, *int32) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "bad", http.StatusBadRequest)
			return
		}
		var req struct {
			Messages []struct{ Content string } `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		n := int(atomic.AddInt32(&calls, 1))
		content := answer(req.Messages[len(req.Messages)-1].Content, n)
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"role": "assistant", "content": content}}}})
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func fileJSON(path, content string) string {
	raw, _ := json.Marshal(map[string]any{"files": []repairEdit{{Path: path, Content: content}}})
	return string(raw)
}

func TestRepairLoopFixesBrokenGoProjectE2E(t *testing.T) {
	root := brokenGoProject(t)
	srv, calls := mockModel(t, func(prompt string, n int) string {
		if !strings.Contains(prompt, "undefined: c") || !strings.Contains(prompt, "--- main.go") {
			return `{"files":[]}`
		}
		return "Düzeltme:\n```json\n" + fileJSON("main.go", fixedMain) + "\n```"
	})
	approved := 0
	loop := &repairLoop{Root: root, Endpoint: srv.URL, Key: "test-key", Model: "mock", Budget: repairBudget{MaxIterations: 3, MaxModelCalls: 3, MaxDuration: 3 * time.Minute, StepTimeout: 2 * time.Minute, MaxEditBytes: 1 << 16},
		Approve: func(int, []repairEdit) bool { approved++; return true }}
	res, err := loop.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "fixed" || res.Iterations != 1 || *calls != 1 || approved != 1 {
		t.Fatalf("beklenmeyen sonuç: %+v calls=%d", res, *calls)
	}
	if got, _ := os.ReadFile(filepath.Join(root, "main.go")); string(got) != fixedMain {
		t.Fatal("düzeltme yazılmadı")
	}
	first, last := res.Steps[0], res.Steps[len(res.Steps)-1]
	if first.State != "failed" || first.ExitCode == 0 || !strings.Contains(first.Stderr, "undefined: c") {
		t.Fatalf("ilk derleme hatası yakalanmadı: %+v", first)
	}
	if last.State != "success" || last.ExitCode != 0 || strings.Join(last.Argv, " ") != "go test ./..." {
		t.Fatalf("yeniden test geçmedi: %+v", last)
	}
	// Independent retest outside the loop.
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("bağımsız test: %v %s", err, out)
	}
}

func TestRepairLoopStopsAtBudget(t *testing.T) {
	root := brokenGoProject(t)
	srv, calls := mockModel(t, func(string, int) string { return fileJSON("main.go", brokenMain) })
	loop := &repairLoop{Root: root, Endpoint: srv.URL, Key: "test-key", Budget: repairBudget{MaxIterations: 2, MaxModelCalls: 2, MaxDuration: 3 * time.Minute, StepTimeout: time.Minute, MaxEditBytes: 1 << 16},
		Approve: func(int, []repairEdit) bool { return true }}
	res, err := loop.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "budget_exhausted" || *calls != 2 || res.ModelCalls != 2 {
		t.Fatalf("bütçe uygulanmadı: %s calls=%d", res.Status, *calls)
	}
}

func TestRepairLoopRefusesEscapeUnapprovedAndBadJSON(t *testing.T) {
	root := brokenGoProject(t)
	outside := t.TempDir()
	cases := map[string]struct {
		answer  string
		approve bool
	}{
		"escape":     {fileJSON("../"+filepath.Base(outside)+"/pwn.go", "x"), true},
		"unapproved": {fileJSON("main.go", fixedMain), false},
		"badjson":    {"düzeltemedim", true},
	}
	for name, c := range cases {
		srv, _ := mockModel(t, func(string, int) string { return c.answer })
		approve := c.approve
		loop := &repairLoop{Root: root, Endpoint: srv.URL, Key: "test-key", Budget: repairBudget{MaxIterations: 2, MaxModelCalls: 2, MaxDuration: time.Minute, StepTimeout: time.Minute, MaxEditBytes: 1 << 16},
			Approve: func(int, []repairEdit) bool { return approve }}
		res, err := loop.Run(context.Background())
		if err != nil {
			t.Fatal(name, err)
		}
		want := map[string]string{"escape": "rejected", "unapproved": "rejected", "badjson": "model_error"}[name]
		if res.Status != want {
			t.Fatalf("%s: %s", name, res.Status)
		}
		if got, _ := os.ReadFile(filepath.Join(root, "main.go")); string(got) != brokenMain {
			t.Fatalf("%s: dosya değişti", name)
		}
	}
	if entries, _ := os.ReadDir(outside); len(entries) != 0 {
		t.Fatal("kök dışına yazıldı")
	}
}

func TestRepairLoopCancellation(t *testing.T) {
	root := brokenGoProject(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	loop := &repairLoop{Root: root, Endpoint: srv.URL, Budget: repairBudget{MaxIterations: 2, MaxModelCalls: 2, MaxDuration: time.Minute, StepTimeout: time.Minute, MaxEditBytes: 1 << 16}, Approve: func(int, []repairEdit) bool { return true }}
	go func() { time.Sleep(1500 * time.Millisecond); cancel() }()
	res, err := loop.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "cancelled" {
		t.Fatalf("iptal edilmedi: %s", res.Status)
	}
}
