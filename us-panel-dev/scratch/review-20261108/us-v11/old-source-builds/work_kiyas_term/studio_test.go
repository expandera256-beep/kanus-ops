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

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "us-desk")
	if err != nil {
		panic(err)
	}
	os.Setenv("US_DESK", filepath.Join(dir, "desk.json"))
	os.Setenv("US_DESK_BASE", os.Getenv("US_DESK"))
	os.Exit(m.Run())
}

func TestStudioTracksProjectsRepoAndPlugins(t *testing.T) {
	t.Setenv("US_DESK", filepath.Join(t.TempDir(), "desk.json"))
	bin := filepath.Join(t.TempDir(), "gh")
	script := "#!/bin/sh\nif [ -n \"$US_KEY\" ]; then echo LEAK; exit 1; fi\necho acme/app https://github.com/acme/app\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("US_GH", bin)
	t.Setenv("US_KEY", "usta_testsecretvalue")
	root := t.TempDir()
	other := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "not.txt"), []byte("gizli govde satiri\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "board.svg"), []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"><text>yerel</text></svg>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath("git"); err == nil {
		cmd := exec.Command("git", "init")
		cmd.Dir = root
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	}
	s := newView("q8")
	s.root = root
	s.browse = root
	s.enterPlace(root, false)
	s.addTrack("paketle", 1)
	s.enterPlace(other, false)
	if len(s.tracks) != 0 {
		t.Fatal(s.tracks)
	}
	s.useProject(0)
	if len(s.tracks) != 1 || s.tracks[0].Title != "paketle" || s.root != root {
		t.Fatalf("%s %+v", s.root, s.tracks)
	}
	srv := httptest.NewServer(guiMux(&s))
	defer srv.Close()
	page, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(page.Body)
	page.Body.Close()
	html := string(raw)
	for _, want := range []string{"Stüdyo", "İliştir", "Tuval", "GitHub", "Bakış"} {
		if !strings.Contains(html, want) {
			t.Fatal(want)
		}
	}
	if !strings.Contains(html, `id="accesskey"`) {
		t.Fatal("key input missing")
	}
	attached := guiPost(t, srv.URL+"/api/act", `{"op":"attach","path":"not.txt"}`)
	if len(attached.Attached) != 1 || attached.Attached[0] != "not.txt" || strings.Contains(mustJSON(attached), "gizli govde") {
		t.Fatal(mustJSON(attached))
	}
	blocked := guiPost(t, srv.URL+"/api/act", `{"op":"attach","path":"../disari.txt"}`)
	if blocked.Notice == "" || len(blocked.Attached) != 1 {
		t.Fatal(blocked.Notice)
	}
	var wire string
	s.sendModel = func(worker workerID, key, body string) modelResult {
		wire = body
		return modelResult{ok: true, body: "oldu"}
	}
	s.key = "local-test"
	asked := guiPost(t, srv.URL+"/api/act", `{"op":"task","text":"renkleri say"}`)
	if strings.Contains(wire, "gizli govde") || !strings.Contains(wire, "ek: not.txt") {
		t.Fatal(wire)
	}
	if strings.Contains(mustJSON(asked), "gizli govde") {
		t.Fatal("body leaked")
	}
	plugin := guiPost(t, srv.URL+"/api/act", `{"op":"plugin","text":"mcp dosyalar git"}`)
	found := false
	for _, item := range plugin.Plugins {
		if item.Kind == "mcp" && item.Name == "dosyalar" {
			found = true
			if !strings.Contains(item.Text, "Çalıştırılmadı") {
				t.Fatal(item.Text)
			}
		}
	}
	if !found {
		t.Fatal(plugin.Plugins)
	}
	skill := guiPost(t, srv.URL+"/api/act", `{"op":"plugin","text":"skill not kisa satir"}`)
	noted := guiPost(t, srv.URL+"/api/act", `{"op":"skill","text":"not"}`)
	if len(noted.Notes) != 1 || noted.Notes[0] != "kisa satir" {
		t.Fatalf("%+v %s", skill.Plugins, noted.Notice)
	}
	repo := guiPost(t, srv.URL+"/api/act", `{"op":"repo"}`)
	if repo.Repo == "" {
		t.Fatal("repo empty")
	}
	hub := guiPost(t, srv.URL+"/api/act", `{"op":"github"}`)
	if !strings.Contains(hub.GitHub, "acme/app") || strings.Contains(hub.GitHub, "LEAK") || strings.Contains(hub.GitHub, "usta_testsecretvalue") {
		t.Fatal(hub.GitHub)
	}
	board := guiPost(t, srv.URL+"/api/act", `{"op":"render","path":"board.svg"}`)
	if board.Visual != "/view?path=board.svg" {
		t.Fatal(board.Visual)
	}
	view, err := http.Get(srv.URL + "/view?engine=tuval")
	if err != nil {
		t.Fatal(err)
	}
	svg, _ := io.ReadAll(view.Body)
	view.Body.Close()
	if !strings.Contains(string(svg), "<svg") || !strings.Contains(string(svg), "paketle") {
		t.Fatal(string(svg))
	}
	file, err := http.Get(srv.URL + board.Visual)
	if err != nil {
		t.Fatal(err)
	}
	seen, _ := io.ReadAll(file.Body)
	file.Body.Close()
	if !strings.Contains(string(seen), "yerel") {
		t.Fatal(string(seen))
	}
	auto := guiPost(t, srv.URL+"/api/act", `{"op":"auto","text":"bakış"}`)
	if auto.Visual != "/view?engine=tuval" || auto.Repo == "" || !strings.Contains(auto.GitHub, "acme/app") {
		t.Fatalf("%s %s %s", auto.Visual, auto.Repo, auto.GitHub)
	}
}
