package licensesrv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type finalTransportFunc func(*http.Request) (*http.Response, error)

func (f finalTransportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestReadinessInvalidatesEveryUpstreamFailure(t *testing.T) {
	cases := []struct {
		name, body, contentType                string
		status                                 int
		stream, truncated, timeout, connection bool
	}{
		{name: "500", status: 500, stream: true}, {name: "501", status: 501, stream: true}, {name: "502", status: 502, stream: true}, {name: "503", status: 503, stream: true}, {name: "504", status: 504, stream: true}, {name: "599", status: 599, stream: true},
		{name: "malformed-json", status: 200, body: "{", contentType: "application/json"},
		{name: "incomplete-json", status: 200, body: `{"choices":[{"message":{"content":"partial"}}]}`, contentType: "application/json"},
		{name: "json-read-cut", status: 200, body: "{", contentType: "application/json", truncated: true},
		{name: "wrong-sse-type", status: 200, body: goodEvents, contentType: "application/json", stream: true},
		{name: "sse-cut", status: 200, body: strings.ReplaceAll(goodEvents, "data: [DONE]\n\n", ""), contentType: "text/event-stream", stream: true},
		{name: "sse-malformed", status: 200, body: "data: {\n\n", contentType: "text/event-stream", stream: true},
		{name: "sse-truncated", status: 200, body: "data: {\"choices\":[{\"delta\":{\"content\":\"a\"},\"finish_reason\":\"length\"}]}\n\n", contentType: "text/event-stream", stream: true},
		{name: "timeout-json", timeout: true}, {name: "timeout-sse", timeout: true, stream: true},
		{name: "connection-json", connection: true}, {name: "connection-sse", connection: true, stream: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _ := setup(t)
			post(h, payload, sign("whsec", payload), "p")
			key := h.Store.st.Orders["555"][0]
			var chats atomic.Int32
			release := make(chan struct{})
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v1/models" {
					http.Error(w, "loading", 503)
					return
				}
				chats.Add(1)
				if tc.timeout {
					select {
					case <-r.Context().Done():
					case <-release:
					}
					return
				}
				w.Header().Set("Content-Type", tc.contentType)
				if tc.truncated {
					w.Header().Set("Content-Length", "10000")
				}
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer srv.Close()
			defer close(release)
			p, err := NewModelProxy(h.Store, srv.URL, "", "test-model", time.Second, 1, 32)
			if err != nil {
				t.Fatal(err)
			}
			p.readyUntil = time.Now().Add(20 * time.Second)
			if tc.timeout {
				p.Timeout = 30 * time.Millisecond
			}
			if tc.connection {
				p.Client.Transport = finalTransportFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("simulated connection error") })
			}
			before, _ := h.Store.License(key)
			input := strings.Replace(proxyInput, `"stream":true`, fmt.Sprintf(`"stream":%t`, tc.stream), 1)
			proxyRequest(p, key, input, context.Background())
			p.readyMu.Lock()
			cached := time.Now().Before(p.readyUntil)
			p.readyMu.Unlock()
			if cached {
				t.Error("upstream failure retained readiness cache")
			}
			after, _ := h.Store.License(key)
			if after.Reserved != 0 || after.Credits != before.Credits {
				t.Fatal("failure charged or left reservation")
			}
			n := chats.Load()
			w := proxyRequest(p, key, input, context.Background())
			if w.Code != 503 || chats.Load() != n {
				t.Errorf("next request skipped readiness probe: %d", w.Code)
			}
		})
	}
}

func TestMetadataFailureInvalidatesReadiness(t *testing.T) {
	for _, body := range []string{"upstream failed", "{"} {
		t.Run(body, func(t *testing.T) {
			h, _ := setup(t)
			post(h, payload, sign("whsec", payload), "p")
			key := h.Store.st.Orders["555"][0]
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if body == "upstream failed" {
					w.WriteHeader(500)
				}
				fmt.Fprint(w, body)
			}))
			defer s.Close()
			p, _ := NewModelProxy(h.Store, s.URL, "", "test-model", time.Second, 1, 2)
			p.readyUntil = time.Now().Add(20 * time.Second)
			req := httptest.NewRequest("GET", "/v1/models", nil)
			req.Header.Set("Authorization", "Bearer "+key)
			p.Models(httptest.NewRecorder(), req)
			if time.Now().Before(p.readyUntil) {
				t.Fatal("metadata failure retained cache")
			}
		})
	}
}
