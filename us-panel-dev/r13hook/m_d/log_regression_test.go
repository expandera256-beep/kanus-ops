package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLargeIndexDoesNotSummarizeShortRepairOrCrowdOutFiles(t *testing.T) {
	t.Setenv("US_CONTEXT_TOKENS", "")
	root := t.TempDir()
	for i := 0; i < 80; i++ {
		name := fmt.Sprintf("file_%03d_long_name_for_project.py", i)
		if err := os.WriteFile(filepath.Join(root, name), []byte("print('original')\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	name := "file_000_long_name_for_project.py"
	content := strings.Repeat("a", 2900) + "\n"
	if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	s := newView("q8")
	s.root, s.key = root, "test-key"
	s.modelCh = make(chan modelResult, 1)
	calls := 0
	s.sendModel = func(_ workerID, _, wire string) modelResult {
		calls++
		if strings.Contains(wire, "--- PARÇA ---") {
			t.Fatal("short task was summarized")
		}
		if calls == 1 {
			if !strings.Contains(wire, name) {
				t.Fatal("index missing file")
			}
			return modelResult{ok: true, body: "OKU " + name}
		}
		if !strings.Contains(wire, "OKU "+name+"\n"+content) {
			t.Fatal("file was not fully included")
		}
		return modelResult{ok: true, body: "DOSYA " + name + "\nupdated\nBİTTİ\n"}
	}
	s.draft = "projeyi tamamla ve onar"
	s.submit()
	got, err := os.ReadFile(filepath.Join(root, name))
	if err != nil || string(got) != "updated\n" || calls != 2 {
		t.Fatalf("calls=%d content=%q err=%v", calls, got, err)
	}
}

func TestRunIntentDoesNotMatchNounOrNegation(t *testing.T) {
	for _, text := range []string{"çalıştırma komutunu düzelt", "programı çalıştırmayın", "başlatma ayarlarını anlat"} {
		if wantsRun(text) {
			t.Fatalf("unexpected run: %q", text)
		}
	}
	for _, text := range []string{"programı çalıştır", "projeyi başlatın", "run"} {
		if !wantsRun(text) {
			t.Fatalf("missing run: %q", text)
		}
	}
}

func TestStoppedProgramShowsLastExceptionAndWriterAcceptsFullOutput(t *testing.T) {
	var output cappedBuf
	text := strings.Repeat("log line\n", 3000) + "Traceback (most recent call last):\n  File x\nModuleNotFoundError: missing_module\n"
	if n, err := output.Write([]byte(text)); err != nil || n != len(text) {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if got := formatRun("server", output.String(), "durdu"); !strings.Contains(got, "ModuleNotFoundError: missing_module") {
		t.Fatal(got)
	}
}
