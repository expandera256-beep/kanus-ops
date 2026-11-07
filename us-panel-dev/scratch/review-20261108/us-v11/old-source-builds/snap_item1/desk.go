package main

import (
	"bytes"
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
	text, ok, complete := s.localTextWithLimit(rel, 8000)
	if !ok {
		s.notice = "Dosya açılamadı: " + text
		return
	}
	s.openPath = relPath(s.root, path)
	s.openText = text
	s.openRoot = s.root
	s.openEditable = complete
	s.openHash = ""
	s.openRevision++
	if complete {
		s.openHash = hashContent([]byte(text))
		s.notice = ""
	} else {
		s.notice = "Dosya kısmi veya gizlenmiş gösteriliyor; üzerine kaydedilemez."
	}
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
	info, err := os.Lstat(path)
	if err == nil {
		if !info.Mode().IsRegular() || s.openRoot != s.root || s.openPath != relPath(s.root, path) || !s.openEditable {
			s.notice = "Mevcut dosyayı kaydetmek için tamamını güvenle açın."
			return
		}
		current, e := os.ReadFile(path)
		if e != nil || hashContent(current) != s.openHash {
			s.notice = "Dosya açıldıktan sonra değişti; kaydedilmedi."
			return
		}
	} else if !os.IsNotExist(err) {
		s.notice = "Dosya doğrulanamadı."
		return
	}
	record, err := recordChange(s.root, relPath(s.root, path), []byte(text))
	if err != nil {
		s.notice = "Dosya yedeği oluşturulamadı."
		return
	}
	if err := writeModelFile(s.root, path, []byte(text)); err != nil {
		_ = os.Remove(record)
		s.notice = "Dosya yazılamadı."
		return
	}
	if written, e := os.ReadFile(path); e != nil || hashContent(written) != hashContent([]byte(text)) {
		s.notice = "Kaydedilen dosya doğrulanamadı; yedek korundu."
		return
	}
	s.openPath = relPath(s.root, path)
	s.openText = redactLine(text)
	s.openRoot = s.root
	s.openEditable = s.openText == text
	s.openHash = hashContent([]byte(text))
	s.openRevision++
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
	return s.localTextWithLimit(rel, 4000)
}

func (s *session) localTextWithLimit(rel string, byteLimit int) (string, bool, bool) {
	path, err := s.safePath(rel)
	if err != nil {
		return "yol klasörün dışında", false, false
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "dosya yok", false, false
	}
	if info.Size() > 8*1024*1024 {
		return "dosya büyük", false, false
	}
	f, err := os.Open(path)
	if err != nil {
		return "okunamadı", false, false
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return "dosya değişti veya okunamadı", false, false
	}
	body, err := io.ReadAll(io.LimitReader(f, 8*1024*1024+1))
	if err != nil {
		return "okunamadı", false, false
	}
	if len(body) > 8*1024*1024 {
		return "dosya büyük", false, false
	}
	if !utf8.Valid(body) || bytes.IndexByte(body, 0) >= 0 {
		return "ikili dosya veya desteklenmeyen metin kodlaması", false, false
	}
	text := redactLine(string(body))
	complete := text == string(body)
	if byteLimit < 0 {
		byteLimit = 0
	}
	if len(text) > byteLimit {
		cut := byteLimit
		for cut > 0 && !utf8.RuneStart(text[cut]) {
			cut--
		}
		text = text[:cut] + "\nkesildi."
		complete = false
	}
	return text, true, complete
}

func okuPaths(body string) []string {
	if env, yes, err := parseModelActions(body); yes {
		if err != nil {
			return nil
		}
		var paths []string
		for _, op := range env.Operations {
			if op.Op == "read_file" {
				paths = append(paths, op.Path)
			}
		}
		return paths
	}
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
	if full != s.root && noSymlinkPath(s.root, full) != nil {
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

// workspacePage traverses the tree in a stable lexical order and reads at most
// one page plus one item. Hidden/dependency directories and symlinks are omitted.
func workspacePage(root string, offset, limit int) ([]string, bool, error) {
	if offset < 0 || offset > 1000000 || limit < 1 || limit > 200 || refusedFolder(root) || isRootPath(root) {
		return nil, false, os.ErrInvalid
	}
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, false, os.ErrInvalid
	}
	items := make([]string, 0, limit)
	seen := 0
	more := false
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if skipTreeDir(entry.Name()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if seen < offset {
			seen++
			return nil
		}
		if len(items) == limit {
			more = true
			return filepath.SkipAll
		}
		rel := relPath(root, path)
		if entry.IsDir() {
			rel += "/"
		}
		items = append(items, rel)
		seen++
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return items, more, nil
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
