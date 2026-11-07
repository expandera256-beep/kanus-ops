package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepairRequestListsActualPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.py"), []byte("print(1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wire := chatWire("projeyi tamamla ve eksikleri onar", root)
	if !strings.Contains(wire, "main.py") || !strings.Contains(wire, "OKU") || strings.Contains(wire, "print(1)") {
		t.Fatal(wire)
	}
}

func TestRepairAcceptsOnlyListedBareFilePaths(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"README.md", "main.py", ".env"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("hello"), 0o600); err != nil { t.Fatal(err) }
	}
	got := projectReadPaths(root, "README.md\nmain.py")
	if len(got) != 2 || got[0] != "README.md" || got[1] != "main.py" { t.Fatalf("%q", got) }
	for _, body := range []string{"README.md\nnot-real.txt", "README.md\n.env", "../outside", "Here is README.md"} {
		if len(projectReadPaths(root, body)) != 0 { t.Fatalf("unexpected paths: %q", body) }
	}
}

func TestWorkspaceCountDoesNotCountTruncationMarker(t *testing.T) {
	count, limited := workspaceNameCount([]string{"a.py", "b.py", "ve daha"})
	if count != 2 || !limited {
		t.Fatalf("count=%d limited=%v", count, limited)
	}
	if count, limited = workspaceNameCount([]string{"dizinde dosya yok"}); count != 0 || limited {
		t.Fatalf("empty: count=%d limited=%v", count, limited)
	}
}

func TestRepairDoesNotApplyAnUnreadProject(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "mevcut.txt"), []byte("eski\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newView("q8")
	s.root, s.key = root, "local-test"
	s.modelCh = make(chan modelResult, 1)
	calls := 0
	s.sendModel = func(workerID, string, string) modelResult {
		calls++
		return modelResult{ok: true, body: "DOSYA proje/altyapi_tamamla.py\nprint(1)\nBİTTİ\n"}
	}
	s.draft = "projeyi tamamla ve eksikleri onar"
	s.submit()
	if calls != 2 {
		t.Fatalf("calls: %d", calls)
	}
	if _, err := os.Stat(filepath.Join(root, "proje", "altyapi_tamamla.py")); !os.IsNotExist(err) {
		t.Fatalf("unread project was changed: %v", err)
	}
	var journal string
	for _, item := range s.journal {
		journal += item.Text + "\n"
	}
	if !strings.Contains(journal, "dosyaları okunamadı; değişiklik uygulanmadı") || !strings.Contains(journal, "Okunan: yok") {
		t.Fatal(journal)
	}
}

func TestRepairFollowupResetsProgress(t *testing.T) {
	s := newView("q8")
	s.root, s.key = t.TempDir(), "local-test"
	s.modelCh = make(chan modelResult, 1)
	calls := 0
	s.sendModel = func(workerID, string, string) modelResult {
		calls++
		if calls == 1 {
			s.noteLive(21035)
			return modelResult{ok: true, body: "Önce açıklama."}
		}
		if s.liveChars != 0 { t.Fatalf("önceki yanıt sayacı kaldı: %d", s.liveChars) }
		s.noteLive(55)
		if s.liveChars != 55 { t.Fatalf("ikinci çağrı ilerlemedi: %d", s.liveChars) }
		return modelResult{ok: true, body: "Dosya bulunamadı."}
	}
	s.draft = "projeyi tamamla ve eksikleri onar"
	s.submit()
	if calls != 2 { t.Fatalf("calls=%d", calls) }
}

func TestRepairDoesNotOverwriteUnreadFiles(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"okunan.txt", "diger.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("eski\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s := newView("q8")
	s.root, s.key = root, "local-test"
	s.modelCh = make(chan modelResult, 1)
	calls := 0
	s.sendModel = func(_ workerID, _, wire string) modelResult {
		calls++
		if calls == 1 {
			return modelResult{ok: true, body: "OKU okunan.txt"}
		}
		if !strings.Contains(wire, "OKU okunan.txt\neski") {
			t.Fatalf("file was not included: %s", wire)
		}
		return modelResult{ok: true, body: "DOSYA okunan.txt\nyeni\nDOSYA diger.txt\nyeni\nBİTTİ\n"}
	}
	s.draft = "projeyi tamamla ve eksikleri onar"
	s.submit()
	readBody, err := os.ReadFile(filepath.Join(root, "okunan.txt"))
	if err != nil || string(readBody) != "yeni\n" {
		t.Fatalf("read file: %q %v", readBody, err)
	}
	unreadBody, err := os.ReadFile(filepath.Join(root, "diger.txt"))
	if err != nil || string(unreadBody) != "eski\n" {
		t.Fatalf("unread file: %q %v", unreadBody, err)
	}
	var journal string
	for _, item := range s.journal {
		journal += item.Text + "\n"
	}
	if !strings.Contains(journal, "Tam okunmamış dosyalar değiştirilmedi: diger.txt") {
		t.Fatal(journal)
	}
}

func TestMissingPythonIsUnverifiedNotASyntaxError(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PATH", root)
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv("LOCALAPPDATA", root)
	t.Setenv("PROGRAMFILES", root)
	t.Setenv("PROGRAMFILES(X86)", root)
	path := filepath.Join(root, "app.py")
	if err := os.WriteFile(path, []byte("print(1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, why := pyParseCheck(root, path)
	if status != "denetlenemedi" || !strings.Contains(why, "Python bulunamadı") {
		t.Fatalf("status=%q reason=%q", status, why)
	}
}

func TestPartialReadDoesNotAuthorizeOverwrite(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "uzun.txt")
	body := strings.Repeat("x", 5000)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newView("q8")
	s.root = root
	_, ok, complete := s.localText("uzun.txt")
	if !ok || complete {
		t.Fatalf("read=%v complete=%v", ok, complete)
	}
	note, blocked := applyReplyFilesChecked(root, "DOSYA uzun.txt\nyeni\n", map[string]bool{})
	if note != "" || len(blocked) != 1 || blocked[0] != "uzun.txt" {
		t.Fatalf("note=%q blocked=%v", note, blocked)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != body {
		t.Fatalf("file changed: %v", err)
	}
}

func TestCompletionWordDoesNotMakeNewProjectARepair(t *testing.T) {
	if projectRepairRequest("Görev tamamlama özelliği olan yeni bir uygulama oluştur") {
		t.Fatal("yeni proje onarım sayıldı")
	}
	if !projectRepairRequest("Projeyi tamamla ve eksikleri onar") {
		t.Fatal("onarım isteği tanınmadı")
	}
	if !projectRepairRequest("Mevcut projeyi tamamla") {
		t.Fatal("mevcut proje tamamlama tanınmadı")
	}
}
