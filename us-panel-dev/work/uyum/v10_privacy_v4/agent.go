package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func wantsStore(task string) bool {
	folded := foldText(task)
	for _, word := range []string{"e-ticaret", "eticaret", "e ticaret", "magaza", "vitrin"} {
		if strings.Contains(folded, word) {
			return true
		}
	}
	return false
}

func wantsSite(task string) bool {
	for _, word := range splitWords(task) {
		switch foldText(word) {
		case "site", "sitesi", "sitesini", "website":
			return true
		}
	}
	folded := foldText(task)
	return strings.Contains(folded, "web sitesi") || strings.Contains(folded, "web sayfa")
}

func (s *session) carryOut(task string) {
	started := time.Now()
	dir := filepath.Join(s.root, "us-site")
	if !insideRoot(s.root, dir) && dir != s.root {
		s.finishReply(task, "İstek okundu: site kurulacak.", "Site klasörü dışarı çıkıyor.")
		return
	}
	plan := []string{
		"Plan: sayfa yaz, yerelde yayınla, adresi göster.",
		"1. us-site/index.html ve site.css yazılacak.",
		"2. Sayfa 127.0.0.1 üzerinde dinlenecek.",
		"3. Sayfa açılırsa adres sonuçta duracak.",
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		s.finishCarry(task, dir, plan, nil, "", "Klasör açılmadı.", started)
		return
	}
	files := map[string]string{
		"index.html": sitePage,
		"site.css":   siteStyle,
	}
	var written []string
	for name, body := range files {
		path := filepath.Join(dir, name)
		if !insideRoot(dir, path) {
			s.finishCarry(task, dir, plan, nil, "", "Dosya yolu dışarı çıkıyor.", started)
			return
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			s.finishCarry(task, dir, plan, written, "", "Dosya yazılamadı.", started)
			return
		}
		written = append(written, name)
	}
	url, err := s.serveSite(dir)
	if err != nil {
		s.finishCarry(task, dir, plan, written, "", "Yayın açılmadı.", started)
		return
	}
	if err := pageUp(url); err != nil {
		s.finishCarry(task, dir, plan, written, url, "Sayfa açılmadı.", started)
		return
	}
	s.finishCarry(task, dir, plan, written, url, "Site ayakta.", started)
}

func (s *session) finishCarry(task, dir string, plan, files []string, url, note string, started time.Time) {
	lines := make([]string, 0, len(files))
	for _, name := range files {
		lines = append(lines, name)
	}
	if len(lines) == 0 {
		lines = []string{"dizinde dosya yok"}
	}
	runLine := note
	if url != "" {
		runLine = note + "\n" + url
	}
	run := projectRun{
		Task: task, Root: dir, Reading: "İstek okundu: site kurulacak.",
		Steps:      strings.Join(plan, "\n"),
		OrderTitle: "sonuç",
		OrderBody:  strings.Join(plan, "\n") + "\n" + note,
		Files:      lines,
		RunLine:    runLine,
		Lint:       "php aranmadı",
		OK:         note == "Site ayakta.",
		Elapsed:    time.Since(started),
	}
	s.root = dir
	s.browse = dir
	if run.OK && url != "" {
		s.review = "İnceleme geçti. " + url
	} else if note != "" {
		s.review = "İnceleme kaldı. " + note
	}
	s.project = run
	s.ran = true
	s.selected = 2
	s.scroll = 0
	s.notice = ""
	s.runStatus = "durum  ayakta"
	if !run.OK {
		s.runStatus = "durum  durdu"
	}
	s.flow = taskFlow(run, false)
	s.shown = 1
	s.flowing = true
	s.flowAt = time.Now()
	s.follow = true
	s.lintReady = trialResult{Ran: true, OK: run.OK, Line: run.Lint}
	s.draft = ""
	s.cursor = 0
}

func (s *session) serveSite(dir string) (string, error) {
	s.stopSite()
	ln, err := listenSite()
	if err != nil {
		return "", err
	}
	srv := &http.Server{
		Handler:           http.FileServer(http.Dir(dir)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	s.siteSrv = srv
	go func() { _ = srv.Serve(ln) }()
	return "http://" + ln.Addr().String() + "/", nil
}

func (s *session) stopSite() {
	if s.siteSrv == nil {
		return
	}
	_ = s.siteSrv.Close()
	s.siteSrv = nil
}

func listenSite() (net.Listener, error) {
	for _, port := range []int{4491, 4493, 4507, 4519, 4531} {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return ln, nil
		}
	}
	return net.Listen("tcp", "127.0.0.1:0")
}

func pageUp(url string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	var last error
	for try := 0; try < 8; try++ {
		resp, err := client.Get(url)
		if err != nil {
			last = err
			time.Sleep(30 * time.Millisecond)
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK && strings.Contains(string(body), "US deneme sitesi ayakta") {
			return nil
		}
		last = fmt.Errorf("sayfa %d", resp.StatusCode)
		time.Sleep(30 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("sayfa yok")
	}
	return last
}

const sitePage = `<!doctype html>
<html lang="tr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>US Atölye</title>
<link rel="stylesheet" href="site.css">
</head>
<body>
  <p class="mark">US</p>
  <h1>Atölye denemesi</h1>
  <p>Bu sayfa US tarafından yazıldı, yayınlandı ve açık bırakıldı.</p>
  <p id="durum">US deneme sitesi ayakta</p>
</body>
</html>
`

const siteStyle = `:root { color-scheme: dark; }
body { margin: 0; min-height: 100vh; background: #10110e; color: #f4f1ea; font: 18px/1.5 "Segoe UI", sans-serif; padding: 48px 24px; }
.mark { letter-spacing: .18em; color: #d6f25c; font-size: 12px; }
h1 { margin: 8px 0 12px; font-size: 40px; }
#durum { color: #d6f25c; }
`
