package main

import (
    "bufio"
    "strings"
    "testing"
)
func TestChatMemoryPrefersRecentMessages(t *testing.T) {
    s:=newView("q8")
    s.ensureChats()
    s.noteChat("panel", "kullanıcı: "+strings.Repeat("x", 12000))
    s.noteChat("model", "EN_YENİ_YANIT")
    memory:=s.chatMemory()
    if !strings.Contains(memory, "EN_YENİ_YANIT") { t.Fatal("latest turn lost") }
    if len([]rune(memory)) > 6000 { t.Fatal("memory exceeded budget") }
}
func TestStreamErrorAfterPartialTextIsNotSuccess(t *testing.T) {
    data := "data: {\"choices\":[{\"delta\":{\"content\":\"yarım\"}}]}\n\ndata: {\"error\":{\"message\":\"backend failed\"}}\n\n"
    _, _, errText := readModelStream(bufio.NewReader(strings.NewReader(data)), nil)
    if !strings.Contains(errText, "backend failed") { t.Fatal(errText) }
}
