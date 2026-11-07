package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type brokenStream struct{ sent bool }

func (b *brokenStream) Read(p []byte) (int, error) {
	if !b.sent {
		b.sent = true
		return copy(p, "data: {\"choices\":[{\"delta\":{\"content\":\"yarım\"}}]}\n\n"), nil
	}
	return 0, errors.New("timeout")
}

func TestRepairReadPhaseHasSmallBudget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			MaxTokens int `json:"max_tokens"`
			Messages  []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.MaxTokens != 512 || len(body.Messages) == 0 || !strings.Contains(body.Messages[0].Content, "read_file") {
			t.Errorf("unexpected read phase: %+v", body)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"OKU main.go\"}}]}\n\ndata: [DONE]\n\n"))
	}))
	defer srv.Close()
	old := doorList
	doorList = func() []string { return []string{srv.URL} }
	defer func() { doorList = old }()
	s := newView("a100")
	result := s.watchProgram("a100", "test-key", "incele", -1)
	if !result.ok || result.body != "OKU main.go" {
		t.Fatalf("%+v", result)
	}
}

func TestRepairRequestsJSONResponseFormatWithoutForcingChat(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if json.NewDecoder(r.Body).Decode(&body) != nil {
				t.Error("bad body")
			}
			format, has := body["response_format"].(map[string]any)
			if has != enabled || has && format["type"] != "json_object" {
				t.Errorf("format=%v enabled=%v", format, enabled)
			}
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
		}))
		ctx := context.WithValue(context.Background(), modelActionsJSONKey{}, enabled)
		result := postCompletionContext(ctx, srv.URL, "q8", "test-key", "task", programSystem, 128, "", 0, nil, defaultContextLimits())
		srv.Close()
		if !result.ok {
			t.Fatal(result.body)
		}
	}
}
func TestIncompleteStreamDoesNotSucceed(t *testing.T) {
	for _, in := range []string{
		"data: {\"choices\":[{\"delta\":{\"content\":\"yarım\"}}]}\n\n",
		"data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"length\"}]}\n\n",
	} {
		content, _, err := readModelStream(bufio.NewReader(strings.NewReader(in)), nil)
		if err == "" || content != "" {
			t.Fatalf("incomplete result: %q %q", content, err)
		}
	}
	if content, _, err := readModelStream(bufio.NewReader(&brokenStream{}), nil); err == "" || content != "" {
		t.Fatal("timeout accepted")
	}
}
func TestCompletedStreamIsAccepted(t *testing.T) {
	for _, end := range []string{"data: [DONE]\n\n", "data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n"} {
		in := "data: {\"choices\":[{\"delta\":{\"content\":\"tam\"}}]}\n\n" + end
		content, _, err := readModelStream(bufio.NewReader(strings.NewReader(in)), nil)
		if err != "" || content != "tam" {
			t.Fatalf("%q %q", content, err)
		}
	}
}

func TestMeteredStreamRequiresTerminalAndMatchesRequestID(t *testing.T) {
	for _, mode := range []string{"missing-terminal", "wrong-id", "complete"} {
		t.Run(mode, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id := r.Header.Get("X-US-Request-ID")
				if len(id) != 32 {
					t.Errorf("request ID missing: %q", id)
				}
				if mode == "wrong-id" {
					id = "another-request"
				}
				w.Header().Set("X-US-Request-ID", id)
				w.Header().Set("X-US-Metered", "1")
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"DOSYA a.txt\\nhello\\nBİTTİ\\n\"}}]}\n\n"))
				if mode == "complete" {
					_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"))
				}
			}))
			defer srv.Close()
			result := postCompletion(srv.URL, "q8", "test-key", "task", programSystem, 128)
			if result.ok != (mode == "complete") {
				t.Fatalf("ok=%v body=%s", result.ok, result.body)
			}
		})
	}
}

func TestSubmittedRequestWithoutResponseIsNotRetried(t *testing.T) {
	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		_ = conn.Close()
	}))
	defer first.Close()
	hits := 0
	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"duplicate"},"finish_reason":"stop"}]}`))
	}))
	defer second.Close()
	old := doorList
	doorList = func() []string { return []string{first.URL, second.URL} }
	defer func() { doorList = old }()
	result := postDoor("q8", "test-key", "task", programSystem, 128)
	if result.ok || hits != 0 || !strings.Contains(result.body, "belirsiz") {
		t.Fatalf("result=%+v fallback calls=%d", result, hits)
	}
}
