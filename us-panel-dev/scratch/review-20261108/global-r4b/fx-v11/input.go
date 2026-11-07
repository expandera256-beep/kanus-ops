package main

import (
	"unicode/utf8"
)

type keyEvent struct {
	kind string
	r    rune
}

func parseKeys(buf []byte) ([]keyEvent, []byte) {
	var events []keyEvent
	for len(buf) > 0 {
		if hasPrefix(buf, "\x1b[200~") {
			end := indexBytes(buf, "\x1b[201~")
			if end < 0 {
				return events, buf
			}
			payload := buf[len("\x1b[200~"):end]
			for len(payload) > 0 {
				r, size := utf8.DecodeRune(payload)
				if r == utf8.RuneError && !utf8.FullRune(payload) {
					break
				}
				if r == '\r' || r == '\n' {
					events = append(events, keyEvent{kind: "rune", r: '\n'})
				} else if r >= 32 {
					events = append(events, keyEvent{kind: "rune", r: r})
				}
				payload = payload[size:]
			}
			buf = buf[end+len("\x1b[201~"):]
			continue
		}
		switch buf[0] {
		case 3:
			events = append(events, keyEvent{kind: "ctrlc"})
			buf = buf[1:]
			continue
		case 16:
			events = append(events, keyEvent{kind: "ctrlp"})
			buf = buf[1:]
			continue
		case 21:
			events = append(events, keyEvent{kind: "clear"})
			buf = buf[1:]
			continue
		case 8, 127:
			events = append(events, keyEvent{kind: "backspace"})
			buf = buf[1:]
			continue
		case '\r', '\n':
			events = append(events, keyEvent{kind: "enter"})
			buf = buf[1:]
			continue
		case '\t':
			events = append(events, keyEvent{kind: "tab"})
			buf = buf[1:]
			continue
		}
		if buf[0] == 27 {
			if len(buf) == 1 {
				return events, buf
			}
			if buf[1] != '[' && buf[1] != 'O' {
				events = append(events, keyEvent{kind: "escape"})
				buf = buf[1:]
				continue
			}
			if len(buf) >= 3 && buf[1] == '[' && (buf[2] == 'A' || buf[2] == 'B' || buf[2] == 'C' || buf[2] == 'D') {
				kind := map[byte]string{'A': "up", 'B': "down", 'C': "right", 'D': "left"}[buf[2]]
				events = append(events, keyEvent{kind: kind})
				buf = buf[3:]
				continue
			}
			if hasPrefix(buf, "\x1b[3~") {
				events = append(events, keyEvent{kind: "backspace"})
				buf = buf[4:]
				continue
			}
			i := 2
			for i < len(buf) && !isFinal(buf[i]) {
				i++
			}
			if i >= len(buf) {
				return events, buf
			}
			buf = buf[i+1:]
			continue
		}
		r, size := utf8.DecodeRune(buf)
		if r == utf8.RuneError && !utf8.FullRune(buf) {
			return events, buf
		}
		if r >= 32 {
			events = append(events, keyEvent{kind: "rune", r: r})
		}
		buf = buf[size:]
	}
	return events, nil
}

func hasPrefix(buf []byte, prefix string) bool {
	if len(buf) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if buf[i] != prefix[i] {
			return false
		}
	}
	return true
}

func indexBytes(buf []byte, needle string) int {
	if len(needle) == 0 || len(buf) < len(needle) {
		return -1
	}
	for i := 0; i+len(needle) <= len(buf); i++ {
		if hasPrefix(buf[i:], needle) {
			return i
		}
	}
	return -1
}

func isFinal(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '~'
}

func (s *session) insert(r rune) {
	runes := []rune(s.draft)
	if len(runes) >= 4000 {
		return
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
	if s.cursor > len(runes) {
		s.cursor = len(runes)
	}
	next := make([]rune, 0, len(runes)+1)
	next = append(next, runes[:s.cursor]...)
	next = append(next, r)
	next = append(next, runes[s.cursor:]...)
	s.draft = string(next)
	s.cursor++
}

func (s *session) backspace() {
	runes := []rune(s.draft)
	if s.cursor <= 0 || len(runes) == 0 {
		return
	}
	if s.cursor > len(runes) {
		s.cursor = len(runes)
	}
	next := append([]rune{}, runes[:s.cursor-1]...)
	next = append(next, runes[s.cursor:]...)
	s.draft = string(next)
	s.cursor--
}

func (s *session) keyInsert(r rune) {
	if r < 32 || r == 127 {
		return
	}
	runes := []rune(s.key)
	if len(runes) >= 200 {
		return
	}
	if s.keyCursor < 0 {
		s.keyCursor = 0
	}
	if s.keyCursor > len(runes) {
		s.keyCursor = len(runes)
	}
	next := make([]rune, 0, len(runes)+1)
	next = append(next, runes[:s.keyCursor]...)
	next = append(next, r)
	next = append(next, runes[s.keyCursor:]...)
	s.key = string(next)
	s.keyCursor++
}

func (s *session) keyBackspace() {
	runes := []rune(s.key)
	if s.keyCursor <= 0 || len(runes) == 0 {
		return
	}
	if s.keyCursor > len(runes) {
		s.keyCursor = len(runes)
	}
	next := append([]rune{}, runes[:s.keyCursor-1]...)
	next = append(next, runes[s.keyCursor:]...)
	s.key = string(next)
	s.keyCursor--
}
