package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJournalKeepsThePromptAndDropsTheFileBody(t *testing.T) {
	if got := journalText("plan\nDOSYA app.py\nprint('gizli govde')\n"); strings.Contains(got, "gizli govde") || !strings.Contains(got, "DOSYA app.py") || !strings.Contains(got, "dosya gövdesi loglanmadı") {
		t.Fatal(got)
	}
	long := strings.Repeat("garanti-bbva-api/app/api/routes.py, ", 80)
	cut := trimLog(long, 80)
	if strings.Contains(cut, "rout kesildi") || !strings.HasSuffix(cut, "kesildi.") {
		t.Fatal(cut)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "not.txt"), []byte("gizli govde satiri\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newView("q8")
	s.root = root
	s.key = "local-test"
	s.sendModel = func(worker workerID, key, body string) modelResult {
		if strings.Contains(body, "gizli govde") || strings.Contains(body, "local-test") {
			t.Fatal(body)
		}
		return modelResult{ok: true, body: "DOSYA out.py\nprint('gizli govde')\n"}
	}
	handler := guiMux(&s)
	// Verify the journal entry produced when changes are applied immediately.
	s.reviewBeforeApply = false
	srv := httptest.NewServer(handler)
	defer srv.Close()
	page, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(page.Body)
	page.Body.Close()
	if !strings.Contains(string(raw), "Günlük") || !strings.Contains(string(raw), "çalışma alanı") {
		t.Fatal("journal chrome")
	}
	desk := guiPost(t, srv.URL+"/api/act", `{"op":"task","text":"renkleri say"}`)
	var sawUser, sawMachine, sawThought, sawSummary bool
	var last int
	for _, item := range desk.Journal {
		if item.N <= last {
			t.Fatal(item.N)
		}
		last = item.N
		if strings.Contains(item.Text, "gizli govde") || strings.Contains(item.Text, "local-test") {
			t.Fatal(item.Text)
		}
		if item.Kind == "istem" && strings.Contains(item.Text, "kullanıcı: renkleri say") {
			sawUser = true
		}
		if item.Kind == "istem" && strings.Contains(item.Text, "renkleri say") && strings.Contains(item.Text, "İş modele gitti") && !strings.Contains(item.Text, "dosya adları:") {
			sawMachine = true
		}
		if item.Kind == "işlem" && strings.Contains(item.Text, "Yazıldı: out.py") && strings.Contains(item.Text, "kayda geçmedi") {
			sawThought = true
		}
		if item.Kind == "özet" && strings.Contains(item.Text, "renkleri say") && strings.Contains(item.Text, "out.py") && strings.Contains(item.Text, "Sonuç:") {
			sawSummary = true
		}
	}
	if !sawUser || !sawMachine || !sawThought || !sawSummary {
		t.Fatalf("%+v", desk.Journal)
	}
	rawDesk, err := os.ReadFile(os.Getenv("US_DESK"))
	if err != nil || !strings.Contains(string(rawDesk), "Özet.") || strings.Contains(string(rawDesk), "gizli govde") {
		t.Fatalf("saved journal %v %s", err, rawDesk)
	}
}
