package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const usTheory = "Plan: istem maddelere ayrılır ve her maddenin bitişi yazılır.\nAraç: listele, oku, yaz, ekle, ara, çalıştır, incele, takip, not, rea.\nDosya: yazma yalnız seçili klasörün içindedir. Gövde modele gitmez.\nÇalıştır: program veya sayfa yerelde açılır. Adres sonuçta durur.\nİnceleme: dosya ve çıktı beklenenle karşılaştırılır. Tutmuyorsa iş açık kalır.\nREA: seçili dosya resmi rea CLI ile statik incelenir. Hopper kurulmaz. Gövde modele gitmez.\nÇok adım: bir adım bitmeden sıradaki başlamaz. Aynı sonuç ikinci kez gelirse durulur."

var usTools = []string{"plan", "listele", "oku", "yaz", "ekle", "ara", "çalıştır", "incele", "takip", "not", "rea"}

type trackItem struct {
	Title string `json:"title"`
	State string `json:"state"`
}

func (s *session) readFile(rel string) {
	path, err := s.safePath(rel)
	if err != nil {
		s.notice = "Dosya klasörün dışında."
		return
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		s.notice = "Dosya yok."
		return
	}
	if info.Size() > 256*1024 {
		s.notice = "Dosya büyük. Açılmadı."
		return
	}
	body, err := os.ReadFile(path)
	if err != nil {
		s.notice = "Dosya okunamadı."
		return
	}
	text := redactLine(string(body))
	if len(text) > 8000 {
		text = text[:8000] + "\nkesildi."
	}
	s.openPath = relPath(s.root, path)
	s.openText = text
	s.notice = ""
}

func (s *session) saveFile(rel, text string) {
	if strings.TrimSpace(rel) == "" {
		rel = s.openPath
	}
	if strings.TrimSpace(text) == "" {
		s.notice = "Metin boş. Dosya duruyor."
		return
	}
	path, err := s.safePath(rel)
	if err != nil {
		s.notice = "Dosya klasörün dışında."
		return
	}
	if len(text) > 256*1024 {
		s.notice = "Metin büyük. Yazılmadı."
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		s.notice = "Klasör açılmadı."
		return
	}
	if !insideRoot(s.root, path) {
		s.notice = "Dosya klasörün dışında."
		return
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		s.notice = "Dosya yazılamadı."
		return
	}
	s.openPath = relPath(s.root, path)
	s.openText = redactLine(text)
	s.notice = ""
}

func (s *session) appendFile(rel, text string) {
	if strings.TrimSpace(rel) == "" {
		rel = s.openPath
	}
	if strings.TrimSpace(text) == "" {
		s.notice = "Metin boş. Dosya duruyor."
		return
	}
	path, err := s.safePath(rel)
	if err != nil {
		s.notice = "Dosya klasörün dışında."
		return
	}
	prev, _ := os.ReadFile(path)
	if len(prev)+len(text) > 256*1024 {
		s.notice = "Metin büyük. Eklenmedi."
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		s.notice = "Klasör açılmadı."
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		s.notice = "Dosya yazılamadı."
		return
	}
	_, err = f.WriteString(text)
	_ = f.Close()
	if err != nil {
		s.notice = "Dosya yazılamadı."
		return
	}
	s.readFile(relPath(s.root, path))
}

func (s *session) findInFolder(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		s.notice = "Arama boş."
		return
	}
	needle := strings.ToLower(query)
	var hits []string
	if refusedFolder(s.root) {
		s.hits = []string{"bu klasör taranmaz"}
		return
	}
	visited := 0
	_ = filepath.WalkDir(s.root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || len(hits) >= 20 {
			return nil
		}
		name := entry.Name()
		if entry.IsDir() {
			if path != s.root && (skipTreeDir(name) || depth(s.root, path) > workspaceDepth || isRootPath(s.root)) {
				return filepath.SkipDir
			}
			return nil
		}
		visited++
		if visited > 400 {
			return filepath.SkipAll
		}
		rel := relPath(s.root, path)
		if strings.Contains(strings.ToLower(name), needle) {
			hits = append(hits, rel)
			return nil
		}
		info, statErr := entry.Info()
		if statErr != nil || info.Size() > 256*1024 {
			return nil
		}
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		for i, line := range strings.Split(string(body), "\n") {
			if strings.Contains(strings.ToLower(line), needle) {
				hits = append(hits, redactLine(rel+": "+trimHit(line)))
				if i > 400 || len(hits) >= 20 {
					break
				}
			}
		}
		return nil
	})
	if len(hits) == 0 {
		hits = []string{"eşleşme yok"}
	}
	s.hits = hits
	s.notice = ""
}

func (s *session) addTrack(title string, state int) {
	title = strings.TrimSpace(redactLine(title))
	if title == "" {
		s.notice = "Madde boş."
		return
	}
	label := "açık"
	switch state {
	case 1:
		label = "sürüyor"
	case 2:
		label = "bitti"
	}
	for i := range s.tracks {
		if s.tracks[i].Title == title {
			s.tracks[i].State = label
			s.notice = ""
			s.syncProject()
			return
		}
	}
	s.tracks = append(s.tracks, trackItem{Title: title, State: label})
	s.notice = ""
	s.syncProject()
}

func (s *session) addNote(text string) {
	text = strings.TrimSpace(redactLine(text))
	if text == "" {
		s.notice = "Not boş."
		return
	}
	s.notes = append(s.notes, text)
	s.notice = ""
}

func (s *session) reviewNow() string {
	for _, line := range strings.Split(s.project.RunLine, "\n") {
		line = strings.TrimSpace(line)
		if !localURL(line) {
			continue
		}
		if err := pageLive(line); err != nil {
			s.review = "İnceleme kaldı. " + line
			return s.review
		}
		s.review = "İnceleme geçti. " + line
		return s.review
	}
	names := workspaceNames(s.root)
	if len(names) == 0 || names[0] == "dizinde dosya yok" {
		s.review = "İnceleme kaldı. Klasör boş."
		return s.review
	}
	s.review = "İnceleme geçti. " + names[0] + " duruyor."
	return s.review
}

// localText returns whether a file was read and whether its complete,
// unredacted contents were available to the model. A partial read is useful
// for inspection but must not authorize replacing the original file.
func (s *session) localText(rel string) (string, bool, bool) {
	path, err := s.safePath(rel)
	if err != nil {
		return "yol klasörün dışında", false, false
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "dosya yok", false, false
	}
	if info.Size() > 64*1024 {
		return "dosya büyük", false, false
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return "okunamadı", false, false
	}
	text := redactLine(string(body))
	complete := text == string(body)
	if len(text) > 4000 {
		cut := 4000
		for cut > 0 && !utf8.RuneStart(text[cut]) {
			cut--
		}
		text = text[:cut] + "\nkesildi."
		complete = false
	}
	return text, true, complete
}

func okuPaths(body string) []string {
	if strings.Contains(body, "DOSYA ") || strings.Contains(body, "DOSYA:") || strings.Contains(body, "```") {
		return nil
	}
	var out []string
	for _, line := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n") {
		trim := strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(trim, "OKU "); ok {
			rest = strings.TrimSpace(rest)
			if rest != "" {
				out = append(out, rest)
			}
		}
	}
	return out
}

func (s *session) safePath(rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" || refusedFolder(s.root) {
		return "", os.ErrInvalid
	}
	rel = filepath.Clean(rel)
	full := rel
	if !filepath.IsAbs(rel) {
		full = filepath.Join(s.root, rel)
	}
	if full != s.root && !insideRoot(s.root, full) {
		return "", os.ErrInvalid
	}
	return full, nil
}

func relPath(root, full string) string {
	rel, err := filepath.Rel(root, full)
	if err != nil {
		return filepath.Base(full)
	}
	return filepath.ToSlash(rel)
}

const (
	workspaceDepth = 5
	workspaceCap   = 80
)

func workspaceNameCount(names []string) (int, bool) {
	if len(names) == 1 {
		switch names[0] {
		case "dizinde dosya yok", "bu klasör taranmaz", "klasör açılmadı":
			return 0, false
		}
	}
	if len(names) > 0 && names[len(names)-1] == "ve daha" {
		return len(names) - 1, true
	}
	return len(names), false
}

func workspaceNames(root string) []string {
	if refusedFolder(root) {
		return []string{"bu klasör taranmaz"}
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return []string{"klasör açılmadı"}
	}
	drive := isRootPath(root)
	var names []string
	stopped := false
	walkErr := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || stopped {
			return nil
		}
		if path == root {
			return nil
		}
		name := entry.Name()
		if entry.IsDir() {
			if skipTreeDir(name) || drive || depth(root, path) > workspaceDepth {
				return filepath.SkipDir
			}
		} else if skipTreeDir(name) {
			return nil
		}
		rel := relPath(root, path)
		if entry.IsDir() {
			rel += "/"
		}
		names = append(names, rel)
		if len(names) >= workspaceCap {
			stopped = true
			return filepath.SkipAll
		}
		return nil
	})
	if walkErr != nil && len(names) == 0 {
		return []string{"klasör açılmadı"}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return []string{"dizinde dosya yok"}
	}
	if stopped {
		names = append(names, "ve daha")
	}
	return names
}

func trimHit(line string) string {
	line = strings.TrimSpace(line)
	runes := []rune(line)
	if len(runes) > 80 {
		return string(runes[:80])
	}
	return line
}

func localURL(raw string) bool {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return false
	}
	return strings.Contains(raw, "127.0.0.1") || strings.Contains(raw, "localhost")
}

func pageLive(url string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if resp.StatusCode != http.StatusOK || len(body) == 0 {
		return os.ErrInvalid
	}
	return nil
}
