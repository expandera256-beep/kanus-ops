package main

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLimitedStreamKeepsPartialOnlyForRecovery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ilk bölüm  \"}}]}\n\ndata: {\"choices\":[{\"delta\":{},\"finish_reason\":\"length\"}]}\n\n")
	}))
	defer srv.Close()
	got := postCompletion(srv.URL, "a100", "test", "test", "system", 24)
	if got.ok || !got.limited || got.partial != "ilk bölüm  " || got.body != outputLimitError {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestNonStreamLengthIsNotAccepted(t *testing.T) {
	content, _, err := readCompletion(`{"choices":[{"message":{"content":"yarım"},"finish_reason":"length"}]}`)
	if content != "yarım" || err != outputLimitError { t.Fatalf("content=%q err=%q", content, err) }
}

func TestLengthWinsOverJobMarker(t *testing.T) {
	in := "data: {\"choices\":[{\"delta\":{\"content\":\"yarım\\nBİTTİ\\n\"},\"finish_reason\":\"length\"}]}\n\n"
	_, _, err := readModelStream(bufio.NewReader(strings.NewReader(in)), nil)
	if err != outputLimitError { t.Fatalf("unexpected error %q", err) }
}

func TestContinuationJoinsExactAnchorAcrossTwoLimits(t *testing.T) {
	t.Setenv("US_DESK", filepath.Join(t.TempDir(), "desk.json"))
	s := newView("a100")
	initial := "DOSYA main.txt\n" + strings.Repeat("şeffaf içerik ", 20) + "ilk "
	called := 0
	send := func(wire string, round int) modelResult {
		called++
		if round != -3 || len(wire)+len(continuationSystem) > promptWireBudget { t.Fatal("invalid continuation") }
		_, anchor, ok := strings.Cut(wire, "\nKÖPRÜ:\n")
		if !ok { t.Fatal("missing anchor") }
		if called == 1 { return modelResult{limited:true, partial:anchor+"ikinci "} }
		return modelResult{ok:true, raw:anchor+"son", body:anchor+"son"}
	}
	got := s.continueLimitedResponse("görev", modelResult{limited:true, partial:initial}, send)
	if !got.ok || got.raw != initial+"ikinci son" || called != 2 { t.Fatalf("got=%+v calls=%d", got, called) }
}

func TestContinuationRejectsMismatchedAnchor(t *testing.T) {
	s := newView("a100")
	got := s.continueLimitedResponse("görev", modelResult{limited:true, partial:"DOSYA main.txt\nyarım"}, func(string,int) modelResult {
		return modelResult{ok:true, raw:"Farklı başlangıç ve dosya içeriği"}
	})
	if got.ok || !strings.Contains(got.body, "dosyalar uygulanmadı") { t.Fatalf("%+v", got) }
}

func TestRepairRetriesOneFullyReadFileAfterLimit(t *testing.T) {
	t.Setenv("US_DESK", filepath.Join(t.TempDir(), "desk.json"))
	root := t.TempDir()
	path := filepath.Join(root, "main.txt")
	if err := os.WriteFile(path, []byte("eski\n"), 0600); err != nil { t.Fatal(err) }
	s := newView("a100")
	s.root, s.key = root, "test"
	s.modelCh = make(chan modelResult, 1)
	calls := 0
	s.sendModel = func(_ workerID, _, wire string) modelResult {
		calls++
		switch calls {
		case 1: return modelResult{ok:true, body:"OKU main.txt"}
		case 2: return modelResult{limited:true, partial:"DOSYA main.txt\nyeni bir "}
		default:
			if !strings.Contains(wire, "TEK DOSYA HEDEFİ:\nmain.txt") { t.Fatal("missing target") }
			return modelResult{ok:true, body:"DOSYA main.txt\nyeni bir içerik\n"}
		}
	}
	s.draft = "projeyi tamamla ve onar"
	s.submit()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "yeni bir içerik\n" || calls != 3 { t.Fatalf("got=%q calls=%d err=%v", got, calls, err) }
}

func TestRepairNeverWritesAnotherTargetAfterLimit(t *testing.T) {
	t.Setenv("US_DESK", filepath.Join(t.TempDir(), "desk.json"))
	root := t.TempDir()
	path := filepath.Join(root, "main.txt")
	if err := os.WriteFile(path, []byte("eski\n"), 0600); err != nil { t.Fatal(err) }
	s := newView("a100")
	s.root, s.key = root, "test"
	s.modelCh = make(chan modelResult, 1)
	calls := 0
	s.sendModel = func(_ workerID, _, wire string) modelResult {
		calls++
		switch calls {
		case 1: return modelResult{ok:true, body:"OKU main.txt"}
		case 2: return modelResult{limited:true, partial:"DOSYA main.txt\nyarım"}
		default: return modelResult{ok:true, body:"DOSYA other.txt\nuyumsuz"}
		}
	}
	s.draft = "projeyi tamamla ve onar"
	s.submit()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "eski\n" || calls != 3 { t.Fatalf("got=%q calls=%d err=%v", got, calls, err) }
	if _, err := os.Stat(filepath.Join(root, "other.txt")); !os.IsNotExist(err) { t.Fatalf("other file created: %v", err) }
}

func TestRepairRetryDoesNotTruncateReadContext(t *testing.T) {
	s := newView("a100")
	called := false
	got := s.retrySingleRepair(strings.Repeat("x", promptWireBudget), map[string]bool{"main.txt":true}, func(string,int) modelResult {
		called = true
		return modelResult{ok:true}
	})
	if got.ok || called || !strings.Contains(got.body, "sığmadı") { t.Fatalf("got=%+v called=%v", got, called) }
}
