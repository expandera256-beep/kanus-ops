package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// FINDING-02: the repair loop must never follow redirects (cross-port 307 would
// resend Authorization and the project body to another origin).
func TestRepairLoopRefusesRedirects(t *testing.T) {
	const key = "SYNTHETIC_REPAIR_KEY_FOR_TEST"
	for _, tc := range []struct {
		name   string
		client func() *http.Client
	}{
		{"default_client", func() *http.Client { return nil }},
		{"caller_client", func() *http.Client { return &http.Client{Timeout: 5 * time.Second} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var hit, auth atomic.Bool
			sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hit.Store(true)
				if r.Header.Get("Authorization") != "" {
					auth.Store(true)
				}
				w.Write([]byte(`{"choices":[{"message":{"content":"{\"files\":[]}"}}]}`))
			}))
			defer sink.Close()
			origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, sink.URL+"/sink", http.StatusTemporaryRedirect)
			}))
			defer origin.Close()
			root := t.TempDir()
			os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main(){}\n"), 0o600)
			ft, err := newFileTools(root)
			if err != nil {
				t.Fatal(err)
			}
			caller := tc.client()
			loop := &repairLoop{Root: root, Endpoint: origin.URL, Key: key, Model: "m", Client: caller}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_, err = loop.askFix(ctx, ft, repairStep{State: "failed", Stderr: "x"})
			if hit.Load() || auth.Load() {
				t.Fatalf("redirect followed: sink_hit=%t authorization_leaked=%t", hit.Load(), auth.Load())
			}
			if err == nil {
				t.Fatal("expected error on 307 response")
			}
			if caller != nil && caller.CheckRedirect != nil {
				t.Fatal("caller client was mutated")
			}
		})
	}
}
