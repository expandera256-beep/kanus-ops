package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCancellationClosesActualHTTPModelRequest(t *testing.T) {
	started := make(chan struct{})
	serverCancelled := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"partial\"}}]}\n\n"))
		w.(http.Flusher).Flush()
		close(started)
		<-r.Context().Done()
		close(serverCancelled)
	}))
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan modelResult, 1)
	go func() {
		done <- postCompletionContext(ctx, srv.URL, "q8", "test-key", "task", "system", 128, "", 0, nil, defaultContextLimits())
	}()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case result := <-done:
		if result.ok {
			t.Fatal("cancelled result accepted")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("model request did not cancel")
	}
	select {
	case <-serverCancelled:
	case <-time.After(3 * time.Second):
		t.Fatal("server did not observe disconnect")
	}
}

func TestPanelCancelWorksForShortTaskAndBlocksRootMutation(t *testing.T) {
	s := newView("q8")
	s.root = t.TempDir()
	s.calling = true
	s.longActive = false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.modelCancel = cancel
	h := guiMux(&s)
	for _, op := range []string{"goto", "worker", "save"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "http://127.0.0.1:4477/api/act", bytes.NewBufferString(`{"op":"`+op+`","text":"x","path":"a"}`)))
		if w.Code != 409 {
			t.Fatalf("%s allowed: %d", op, w.Code)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "http://127.0.0.1:4477/api/act", bytes.NewBufferString(`{"op":"model-stop"}`)))
	if w.Code != 200 || ctx.Err() == nil || !s.longStop {
		t.Fatalf("cancel failed: %d %v", w.Code, ctx.Err())
	}
}
