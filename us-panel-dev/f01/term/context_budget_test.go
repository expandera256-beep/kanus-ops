package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestContextBudgetKeepsPartialFileUnmodified(t *testing.T) {
	root := t.TempDir()
	original := strings.Repeat("a", 3900)
	path := filepath.Join(root, "main.txt")
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	s := newView("q8")
	s.root, s.key = root, "local-test"
	s.modelCh = make(chan modelResult, 1)
	calls := 0
	s.sendModel = func(_ workerID, _, wire string) modelResult {
		calls++
		if len(wire) > promptWireBudget {
			t.Fatalf("oversized wire: %d", len(wire))
		}
		if calls == 1 {
			return modelResult{ok: true, body: "OKU main.txt"}
		}
		if !strings.Contains(wire, "Bağlam sınırında kesildi") {
			t.Fatal("missing truncation notice")
		}
		return modelResult{ok: true, body: "DOSYA main.txt\ndeğişti\nBİTTİ\n"}
	}
	s.draft = "projeyi tamamla ve eksikleri onar " + strings.Repeat("uzun ", 250)
	s.submit()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != original {
		t.Fatalf("partial file changed: %v", err)
	}
}

func TestClipUTF8KeepsValidCharacters(t *testing.T) {
	out, clipped := clipUTF8("açığ", 4)
	if !clipped || !utf8.ValidString(out) || len(out) > 4 {
		t.Fatalf("%q", out)
	}
}

func TestOversizedWireFailsBeforeSending(t *testing.T) {
	result := postCompletion("http://127.0.0.1:1/v1/chat/completions", "a100", "test", strings.Repeat("x", 5001), "system", 32)
	if result.ok || !strings.Contains(result.body, "bağlam") {
		t.Fatalf("%+v", result)
	}
}
