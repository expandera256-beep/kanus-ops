package main

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

type projectCard struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	State    string      `json:"state"`
	Tracks   []trackItem `json:"tracks,omitempty"`
	Attached []string    `json:"attached,omitempty"`
}

type pluginCard struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Text string `json:"text"`
}

type deskFile struct {
	Projects []projectCard `json:"projects"`
	Plugins  []pluginCard  `json:"plugins"`
	Journal  []journalItem `json:"journal,omitempty"`
	Chat     []chatTurn    `json:"chat,omitempty"`
	Chats    []chatThread  `json:"chats,omitempty"`
	ChatID   string        `json:"chatId,omitempty"`
}

func (s *session) ensureDesk() {
	if s.deskReady {
		return
	}
	s.deskReady = true
	s.projects = nil
	s.plugins = nil
	path := deskPath()
	if path == "" {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var file deskFile
	if json.Unmarshal(raw, &file) != nil {
		return
	}
	s.projects = file.Projects
	s.plugins = file.Plugins
	if len(s.journal) == 0 {
		s.journal = file.Journal
	}
	if len(file.Chats) > 0 {
		s.chats = file.Chats
		s.chatID = file.ChatID
	} else if len(file.Chat) > 0 {
		s.chats = []chatThread{{ID: "s1", Title: "Sohbet", Turns: file.Chat}}
		s.chatID = "s1"
	}
	if s.chatID != "" {
		for i := range s.chats {
			if s.chats[i].ID == s.chatID {
				s.chat = s.chats[i].Turns
				break
			}
		}
	} else if s.chat == nil {
		s.chat = file.Chat
	}
}

func (s *session) saveDesk() {
	path := deskPath()
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	raw, err := json.MarshalIndent(deskFile{Projects: s.projects, Plugins: s.plugins, Journal: s.journal, Chat: s.chat, Chats: s.chats, ChatID: s.chatID}, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, raw, 0o644)
}

var deskSeq atomic.Int64

func isolateSharedDesk() {
	base := os.Getenv("US_DESK_BASE")
	if base == "" || os.Getenv("US_DESK") != base {
		return
	}
	os.Setenv("US_DESK", filepath.Join(filepath.Dir(base), "s"+itoa(int(deskSeq.Add(1)))+".json"))
}

func deskPath() string {
	if path := strings.TrimSpace(os.Getenv("US_DESK")); path != "" {
		return path
	}
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, "us", "desk.json")
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return aa == bb
}

func (s *session) stashPath(path string) {
	if path == "" {
		return
	}
	s.ensureDesk()
	for i := range s.projects {
		if !samePath(s.projects[i].Path, path) {
			continue
		}
		s.projects[i].Tracks = append([]trackItem(nil), s.tracks...)
		s.projects[i].Attached = append([]string(nil), s.attached...)
	}
}

func (s *session) syncProject() {
	s.stashPath(s.root)
	s.saveDesk()
}

func (s *session) noteOpen(path string) {
	if path == "" || refusedFolder(path) {
		return
	}
	s.ensureDesk()
	for i := range s.projects {
		s.projects[i].State = "duruyor"
		if !samePath(s.projects[i].Path, path) {
			continue
		}
		s.projects[i].State = "açık"
		s.tracks = append([]trackItem(nil), s.projects[i].Tracks...)
		s.attached = append([]string(nil), s.projects[i].Attached...)
		s.saveDesk()
		return
	}
	s.projects = append(s.projects, projectCard{Name: filepath.Base(path), Path: path, State: "açık", Tracks: append([]trackItem(nil), s.tracks...), Attached: append([]string(nil), s.attached...)})
	s.saveDesk()
}

func (s *session) rememberProject(path string) {
	s.ensureDesk()
	found := -1
	for i := range s.projects {
		s.projects[i].State = "duruyor"
		if samePath(s.projects[i].Path, path) {
			found = i
		}
	}
	if found < 0 {
		s.tracks = nil
		s.attached = nil
		s.projects = append(s.projects, projectCard{Name: filepath.Base(path), Path: path, State: "açık"})
		if len(s.projects) > 30 {
			s.projects = s.projects[len(s.projects)-30:]
		}
	} else {
		s.projects[found].State = "açık"
		s.projects[found].Name = filepath.Base(path)
		s.tracks = append([]trackItem(nil), s.projects[found].Tracks...)
		s.attached = append([]string(nil), s.projects[found].Attached...)
	}
	s.saveDesk()
}

func (s *session) useProject(n int) {
	s.ensureDesk()
	if n < 0 || n >= len(s.projects) {
		s.notice = "Proje yok."
		return
	}
	s.stashPath(s.root)
	path := s.projects[n].Path
	s.enterPlace(path, false)
}

func (s *session) attach(rel string) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		s.notice = "Dosya seçilmedi."
		return
	}
	full, err := s.safePath(rel)
	if err != nil {
		s.notice = "Dosya klasörün dışında."
		return
	}
	info, err := os.Stat(full)
	if err != nil || info.IsDir() {
		s.notice = "Dosya yok."
		return
	}
	rel = relPath(s.root, full)
	for _, item := range s.attached {
		if item == rel {
			s.notice = ""
			return
		}
	}
	if len(s.attached) >= 12 {
		s.notice = "Ek çok."
		return
	}
	s.attached = append(s.attached, rel)
	s.notice = ""
	s.noteLog("işlem", "iliştir "+rel)
	s.syncProject()
}

func (s *session) detach(rel string) {
	rel = strings.TrimSpace(rel)
	next := s.attached[:0]
	for _, item := range s.attached {
		if item == rel {
			continue
		}
		next = append(next, item)
	}
	s.attached = next
	s.notice = ""
	s.syncProject()
}

func (s *session) addPlugin(text string) {
	fields := strings.Fields(text)
	if len(fields) < 2 {
		s.notice = "Eklenti eksik. Örnek: mcp ad komut"
		return
	}
	kind := strings.ToLower(fields[0])
	switch kind {
	case "mcp", "skill", "gorsel", "görsel", "otomasyon":
	default:
		s.notice = "Eklenti türü yok."
		return
	}
	if kind == "görsel" {
		kind = "gorsel"
	}
	name := fields[1]
	rest := trimRunes(redactLine(strings.TrimSpace(strings.Join(fields[2:], " "))), 180)
	if kind == "mcp" {
		rest = mcpRecord(rest)
	}
	s.ensureDesk()
	for i := range s.plugins {
		if s.plugins[i].Kind == kind && strings.EqualFold(s.plugins[i].Name, name) {
			s.plugins[i].Text = rest
			s.notice = ""
			s.saveDesk()
			return
		}
	}
	s.plugins = append(s.plugins, pluginCard{Name: name, Kind: kind, Text: rest})
	s.notice = ""
	s.saveDesk()
}

func mcpRecord(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "komut yok. Çalıştırılmadı."
	}
	if _, err := exec.LookPath(fields[0]); err != nil {
		return fields[0] + " yok. Çalıştırılmadı."
	}
	return fields[0] + " kayıtlı. Çalıştırılmadı."
}

func (s *session) runSkill(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "depo", "repo":
		s.repoLine = repoStatus(s.root)
		return s.repoLine
	case "github":
		s.hubLine = githubStatus(s.root)
		return s.hubLine
	case "görsel", "gorsel", "tuval":
		s.visual = "/view?engine=tuval"
		return "Tuval açıldı."
	}
	s.ensureDesk()
	for _, item := range s.plugins {
		if item.Kind == "skill" && strings.EqualFold(item.Name, name) {
			s.addNote(item.Text)
			return "Skill notu duruyor. " + item.Name
		}
	}
	return "Skill yok."
}

func (s *session) pluginView() []pluginCard {
	base := []pluginCard{
		{Name: "tuval", Kind: "gorsel", Text: "yerel pano"},
		{Name: "rea", Kind: "skill", Text: "seçili dosyayı statik incele"},
		{Name: "tek exe", Kind: "skill", Text: "spec ile pyinstaller"},
		{Name: "depo", Kind: "skill", Text: "dal ve değişen adlar"},
		{Name: "github", Kind: "skill", Text: "gh repo view"},
		{Name: "bakış", Kind: "otomasyon", Text: "depo, github, tuval"},
	}
	return append(base, s.plugins...)
}

func repoStatus(root string) string {
	if refusedFolder(root) {
		return "Bu klasör taranmaz."
	}
	if _, err := exec.LookPath("git"); err != nil {
		return "git yok."
	}
	if strings.TrimSpace(gitOut(root, "rev-parse", "--is-inside-work-tree")) != "true" {
		return "Bu klasör depo değil."
	}
	branch := strings.TrimSpace(gitOut(root, "rev-parse", "--abbrev-ref", "HEAD"))
	if branch == "" {
		branch = "bilinmiyor"
	}
	var names []string
	for _, line := range strings.Split(gitOut(root, "status", "--porcelain"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(line) > 3 {
			line = strings.TrimSpace(line[2:])
		}
		names = append(names, line)
		if len(names) == 20 {
			break
		}
	}
	text := "dal " + branch
	if len(names) == 0 {
		return text + "\ntemiz"
	}
	return text + "\n" + strings.Join(names, "\n")
}

func githubStatus(root string) string {
	if refusedFolder(root) {
		return "Bu klasör taranmaz."
	}
	bin, args := githubCmd()
	if bin == "" {
		return "GitHub CLI yok. Kurulum: gh"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = root
	cmd.Env = reaEnv()
	out, err := cmd.CombinedOutput()
	text := trimRunes(strings.TrimSpace(redactLine(string(out))), 300)
	if err != nil {
		if text == "" {
			return "GitHub deposu okunamadı."
		}
		return text
	}
	if text == "" {
		return "GitHub deposu yok."
	}
	return "GitHub " + text
}

func githubCmd() (string, []string) {
	if bin := strings.TrimSpace(os.Getenv("US_GH")); bin != "" {
		return bin, []string{"repo", "view"}
	}
	path, err := exec.LookPath("gh")
	if err != nil {
		return "", nil
	}
	return path, []string{"repo", "view", "--json", "nameWithOwner,url", "-q", `.nameWithOwner + " " + .url`}
}

func gitOut(root string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	cmd.Env = reaEnv()
	out, _ := cmd.CombinedOutput()
	return strings.TrimSpace(redactLine(string(out)))
}

func boardSVG(name string, tracks []trackItem) string {
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="640" height="360" viewBox="0 0 640 360">`)
	b.WriteString(`<rect width="640" height="360" fill="#14161c"/>`)
	b.WriteString(`<text x="24" y="42" fill="#d6f25c" font-size="22" font-family="Segoe UI, sans-serif">`)
	b.WriteString(xmlText(name))
	b.WriteString(`</text>`)
	if len(tracks) == 0 {
		b.WriteString(`<text x="24" y="84" fill="#a39e93" font-size="16" font-family="Segoe UI, sans-serif">takip yok</text>`)
	}
	y := 84
	for i, item := range tracks {
		if i == 8 {
			break
		}
		b.WriteString(`<text x="24" y="`)
		b.WriteString(itoa(y))
		b.WriteString(`" fill="#f4f1ea" font-size="16" font-family="Segoe UI, sans-serif">`)
		b.WriteString(xmlText(item.State + " · " + item.Title))
		b.WriteString(`</text>`)
		y += 28
	}
	b.WriteString(`</svg>`)
	return b.String()
}

func xmlText(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return text
}

func visualPath(rel string) (string, bool) {
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".svg", ".html", ".htm", ".png":
		return "/view?path=" + url.QueryEscape(rel), true
	default:
		return "", false
	}
}
