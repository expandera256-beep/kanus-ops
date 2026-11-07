package main

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

type journalItem struct {
	N    int    `json:"n"`
	Kind string `json:"kind"`
	Text string `json:"text"`
}

func (s *session) noteLog(kind, text string) {
	text = journalText(text)
	if text == "" {
		return
	}
	s.ensureDesk()
	s.journal = append(s.journal, journalItem{N: len(s.journal) + 1, Kind: kind, Text: text})
	if os.Getenv("US_TRACE") == "1" {
		fmt.Fprintf(os.Stderr, "%s: %s\n", kind, text)
	}
	if len(s.journal) > 80 {
		s.journal = s.journal[len(s.journal)-80:]
		for i := range s.journal {
			s.journal[i].N = i + 1
		}
	}
	s.saveDesk()
}

func journalText(text string) string { return journalTextLimit(text, 1400) }

func journalTextLimit(text string, limit int) string {
	var kept []string
	skip := false
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "DOSYA ") || strings.HasPrefix(trim, "DOSYA:") {
			kept = append(kept, trim)
			skip = true
			continue
		}
		if skip {
			continue
		}
		kept = append(kept, line)
	}
	if skip {
		kept = append(kept, "dosya gövdesi loglanmadı.")
	}
	return trimLog(strings.TrimSpace(redactLine(strings.Join(kept, "\n"))), limit)
}

func trimLog(text string, max int) string {
	if utf8.RuneCountInString(text) <= max {
		return text
	}
	cut := len(text)
	count := 0
	for i := range text {
		if count == max {
			cut = i
			break
		}
		count++
	}
	head := text[:cut]
	for i := len(head) - 1; i > len(head)/2; i-- {
		switch head[i] {
		case ',', ' ', '\n':
			head = strings.TrimRight(head[:i], " ,")
			return head + "\nkesildi."
		}
	}
	return strings.TrimRight(head, " ,") + "\nkesildi."
}

type chatTurn struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type chatThread struct {
	ID    string     `json:"id"`
	Title string     `json:"title"`
	Turns []chatTurn `json:"turns,omitempty"`
}

func (s *session) noteChat(role, text string) {
	text = chatText(text)
	if text == "" {
		return
	}
	s.ensureChats()
	idx := s.chatIndex()
	s.chats[idx].Turns = append(s.chats[idx].Turns, chatTurn{Role: role, Text: text})
	if len(s.chats[idx].Turns) > 40 {
		s.chats[idx].Turns = s.chats[idx].Turns[len(s.chats[idx].Turns)-40:]
	}
	if role == "panel" && defaultChatTitle(s.chats[idx].Title) {
		if title := titleFromJob(text); title != "" {
			s.chats[idx].Title = title
		}
	}
	s.chat = s.chats[idx].Turns
	s.saveDesk()
}

func (s *session) ensureChats() {
	s.ensureDesk()
	if len(s.chats) == 0 {
		s.chats = []chatThread{{ID: "s1", Title: "Sohbet 1"}}
		s.chatID = "s1"
	}
	if s.chatIndex() < 0 {
		s.chatID = s.chats[0].ID
	}
	s.chat = s.chats[s.chatIndex()].Turns
}

func (s *session) chatIndex() int {
	for i := range s.chats {
		if s.chats[i].ID == s.chatID {
			return i
		}
	}
	return -1
}

func (s *session) freshChatID() string {
	n := 0
	for _, item := range s.chats {
		if len(item.ID) < 2 || item.ID[0] != 's' {
			continue
		}
		v := 0
		ok := true
		for _, r := range item.ID[1:] {
			if r < '0' || r > '9' {
				ok = false
				break
			}
			v = v*10 + int(r-'0')
		}
		if ok && v > n {
			n = v
		}
	}
	return "s" + itoa(n+1)
}

func defaultChatTitle(title string) bool {
	title = strings.TrimSpace(title)
	if title == "Sohbet" {
		return true
	}
	rest, ok := strings.CutPrefix(title, "Sohbet ")
	if !ok || rest == "" {
		return false
	}
	for _, r := range rest {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func titleFromJob(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trim := strings.TrimSpace(line)
		title := ""
		switch {
		case strings.HasPrefix(trim, "iş: "):
			title = strings.TrimSpace(strings.TrimPrefix(trim, "iş: "))
		case strings.HasPrefix(trim, "kullanıcı: "):
			title = strings.TrimSpace(strings.TrimPrefix(trim, "kullanıcı: "))
		default:
			continue
		}
		runes := []rune(title)
		if len(runes) > 42 {
			title = string(runes[:42])
		}
		return title
	}
	return ""
}

func (s *session) newChat() {
	if s.calling {
		s.notice = "Model çağrılıyor."
		return
	}
	s.ensureChats()
	id := s.freshChatID()
	s.chats = append(s.chats, chatThread{ID: id, Title: "Sohbet " + itoa(len(s.chats)+1)})
	s.chatID = id
	s.chat = nil
	s.flow = nil
	s.notice = ""
	s.saveDesk()
}

func (s *session) useChat(id string) {
	if s.calling {
		s.notice = "Model çağrılıyor."
		return
	}
	s.ensureChats()
	for i := range s.chats {
		if s.chats[i].ID != id {
			continue
		}
		s.chatID = id
		s.chat = append([]chatTurn(nil), s.chats[i].Turns...)
		s.flow = nil
		s.notice = ""
		s.saveDesk()
		return
	}
}

func (s *session) deleteChat() {
	if s.calling {
		s.notice = "Model çağrılıyor."
		return
	}
	s.ensureChats()
	idx := s.chatIndex()
	if idx >= 0 {
		s.chats = append(s.chats[:idx], s.chats[idx+1:]...)
	}
	if len(s.chats) == 0 {
		s.chats = []chatThread{{ID: s.freshChatID(), Title: "Sohbet 1"}}
	}
	if idx >= len(s.chats) {
		idx = len(s.chats) - 1
	}
	if idx < 0 {
		idx = 0
	}
	s.chatID = s.chats[idx].ID
	s.chat = append([]chatTurn(nil), s.chats[idx].Turns...)
	s.flow = nil
	s.notice = ""
	s.saveDesk()
}

func (s *session) chatMemory() string {
	s.ensureChats()
	turns := s.chat
	if len(turns) > 8 {
		turns = turns[len(turns)-8:]
	}
	const maxRunes = 6000
	var lines []string
	remaining := maxRunes
	for i := len(turns) - 1; i >= 0 && remaining > 0; i-- {
		turn := turns[i]
		if turn.Role == "model" && thoughtOnly(turn.Text) { continue }
		text := strings.TrimSpace(chatText(turn.Text))
		if text == "" { continue }
		line := turn.Role + ": " + text
		runes := []rune(line)
		if len(runes) > remaining { runes = runes[:remaining] }
		lines = append(lines, string(runes))
		remaining -= len(runes) + 1
	}
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 { lines[i], lines[j] = lines[j], lines[i] }
	return strings.Join(lines, "\n")
}

func (s *session) withChatMemory(wire string) string {
	mem := s.chatMemory()
	if mem == "" {
		return wire
	}
	parts := strings.SplitN(wire, "\n", 2)
	head := parts[0] + "\nbellek:\n" + mem
	if len(parts) == 1 {
		return head
	}
	return head + "\n" + parts[1]
}

func chatText(text string) string {
	text = journalTextLimit(text, 12000)
	var kept []string
	skip := false
	for _, line := range strings.Split(text, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "onarım parçası:") {
			kept = append(kept, "onarım parçası gitti. Gövde kayda geçmedi.")
			break
		}
		if strings.HasPrefix(trim, "Yerel okuma:") {
			kept = append(kept, "Yerel okuma. Gövde kayda geçmedi.")
			skip = true
			continue
		}
		if strings.HasPrefix(trim, "OKU ") {
			if !strings.Contains(trim, "Gövde kayda geçmedi") {
				trim = trim + " Gövde kayda geçmedi."
			}
			kept = append(kept, trim)
			skip = true
			continue
		}
		if skip {
			if strings.HasPrefix(trim, "Şimdi ") || strings.HasPrefix(trim, "bellek:") || strings.HasPrefix(trim, "sıradaki") || strings.HasPrefix(trim, "test:") {
				skip = false
			} else {
				continue
			}
		}
		kept = append(kept, line)
	}
	return trimLog(strings.TrimSpace(strings.Join(kept, "\n")), 12000)
}

func machineBrief(task string, count int, limited bool, written []string) string {
	made := createdList(written)
	countText := itoa(count)
	if limited {
		countText = "en az " + countText
	}
	line := "İş modele gitti: " + redactLine(task) + ". Klasörde " + countText + " ad."
	if limited {
		line += " Liste ilk " + itoa(count) + " adla sınırlı."
	}
	if made == "yok" {
		return line
	}
	return line + " Yazılan: " + made + "."
}

func createdList(notes []string) string {
	var names []string
	for _, note := range notes {
		name := strings.TrimSpace(strings.TrimPrefix(note, "Model dosyaya yazdı:"))
		if name != "" && name != note {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "yok"
	}
	return strings.Join(names, ", ")
}

func nameList(items []string) string {
	if len(items) == 0 {
		return "yok"
	}
	return strings.Join(items, ", ")
}

func (s *session) noteJobSummary(task string, reads, written []string, ending string) string {
	text := strings.Join([]string{
		"Özet.",
		"İstem: " + redactLine(task) + ".",
		"Okunan: " + nameList(reads) + ".",
		"Yazılan: " + createdList(written) + ".",
		"Sonuç: " + ending + ".",
	}, " ")
	s.noteLog("özet", text)
	return text
}
