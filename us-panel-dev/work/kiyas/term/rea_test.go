package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestREAStaysInsideTheFolder(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "rea")
	script := "#!/bin/sh\nif [ -n \"$US_KEY\" ] || [ -n \"$USTA_KEY\" ]; then echo LEAK; exit 1; fi\nprintf '%s\\n' \"$@\"\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("US_REA", bin)
	t.Setenv("US_KEY", "usta_testsecretvalue")
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.js"), []byte("function toplam(a, b) { return a + b; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	note := runREA(root, "app.js", "inspect-artifact")
	if strings.Contains(note, "LEAK") || strings.Contains(note, "usta_testsecretvalue") {
		t.Fatal(note)
	}
	if !strings.Contains(note, "inspect-artifact") || !strings.Contains(note, "app.js") || strings.Contains(note, "hopper") {
		t.Fatal(note)
	}
	if blocked := runREA(root, "../disari.js", "inspect-artifact"); blocked != "Dosya klasörün dışında." {
		t.Fatal(blocked)
	}
	if refused := runREA(root, "app.js", "decompile"); !strings.Contains(refused, "REA komutu yok") {
		t.Fatal(refused)
	}
	if hopper := runREA(root, "app.js", "inspect-artifact --install-hopper"); !strings.Contains(hopper, "REA komutu yok") {
		t.Fatal(hopper)
	}
	js := runREA(root, "", "rea js app.js")
	if !strings.Contains(js, "analyze-javascript-application") || !strings.Contains(js, "app.js") {
		t.Fatal(js)
	}
	page := runREA(root, "", "https://rea.tools/tr/")
	if !strings.Contains(page, "inspect-artifact") || strings.Contains(page, "rea.tools") || strings.Contains(page, "function toplam") {
		t.Fatal(page)
	}
}

func TestPanelRunsREAWithoutTheModel(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "rea")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\"\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("US_REA", bin)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "toplam.js"), []byte("function toplam(a, b) { return a + b; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newView("q8")
	s.root = root
	s.browse = root
	srv := httptest.NewServer(guiMux(&s))
	defer srv.Close()
	response, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if !strings.Contains(string(raw), "Tersten incele") {
		t.Fatal("rea control missing")
	}
	asked := guiPost(t, srv.URL+"/api/act", `{"op":"task","text":"rea toplam.js"}`)
	if !strings.Contains(asked.Rea, "inspect-artifact") || !strings.Contains(asked.Rea, "toplam.js") || strings.Contains(asked.Rea, "function toplam") {
		t.Fatal(asked.Rea)
	}
	if asked.Notice == "Model çağrılıyor." {
		t.Fatal(asked.Notice)
	}
}

func TestOfficialREAInspectsOwnedFile(t *testing.T) {
	if _, err := exec.LookPath("npx"); err != nil {
		t.Skip("npx yok")
	}
	t.Setenv("US_REA", "")
	root := t.TempDir()
	body := "function toplam(a, b) { return a + b; }\nmodule.exports = { toplam };\n"
	if err := os.WriteFile(filepath.Join(root, "toplam.js"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	note := runREA(root, "toplam.js", "inspect-artifact")
	if strings.Contains(note, "REA yok") || strings.Contains(note, "bitmedi") || strings.Contains(note, "Dosya yok") {
		t.Fatal(note)
	}
	lower := strings.ToLower(note)
	if !strings.Contains(lower, "javascript") && !strings.Contains(note, "toplam.js") && !strings.Contains(lower, "evidence") {
		t.Fatal(note)
	}
}
