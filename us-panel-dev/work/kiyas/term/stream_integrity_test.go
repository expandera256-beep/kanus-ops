package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type brokenStream struct { sent bool }
func (b *brokenStream) Read(p []byte) (int, error) {
	if !b.sent { b.sent = true; return copy(p, "data: {\"choices\":[{\"delta\":{\"content\":\"yarım\"}}]}\n\n"), nil }
	return 0, errors.New("timeout")
}

func TestRepairReadPhaseHasSmallBudget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct { MaxTokens int `json:"max_tokens"`; Messages []struct { Content string `json:"content"` } `json:"messages"` }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil { t.Fatal(err) }
		if body.MaxTokens != 512 || len(body.Messages) == 0 || !strings.Contains(body.Messages[0].Content, "OKU") { t.Errorf("unexpected read phase: %+v", body) }
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"OKU main.go\"}}]}\n\ndata: [DONE]\n\n"))
	}))
	defer srv.Close()
	old := doorList
	doorList = func() []string { return []string{srv.URL} }
	defer func() { doorList = old }()
	s := newView("a100")
	result := s.watchProgram("a100", "test-key", "incele", -1)
	if !result.ok || result.body != "OKU main.go" { t.Fatalf("%+v", result) }
}
func TestIncompleteStreamDoesNotSucceed(t *testing.T) {
	for _, in := range []string{
		"data: {\"choices\":[{\"delta\":{\"content\":\"yarım\"}}]}\n\n",
		"data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"length\"}]}\n\n",
	} {
		content, _, err := readModelStream(bufio.NewReader(strings.NewReader(in)), nil)
		if err == "" || content != "" { t.Fatalf("incomplete result: %q %q", content, err) }
	}
	if content, _, err := readModelStream(bufio.NewReader(&brokenStream{}), nil); err == "" || content != "" { t.Fatal("timeout accepted") }
}
func TestCompletedStreamIsAccepted(t *testing.T) {
	for _, end := range []string{"data: [DONE]\n\n", "data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n"} {
		in := "data: {\"choices\":[{\"delta\":{\"content\":\"tam\"}}]}\n\n" + end
		content, _, err := readModelStream(bufio.NewReader(strings.NewReader(in)), nil)
		if err != "" || content != "tam" { t.Fatalf("%q %q", content, err) }
	}
}
