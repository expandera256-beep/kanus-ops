package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestModelAvailabilityRequiresMatchingModel(t *testing.T) {
	found := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, `{"data":[{"id":"deepseek-v4.1-flash"}]}`)
	}))
	defer found.Close()
	if !modelAvailable(found.URL, found.Client()) {
		t.Fatal("expected model")
	}
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"data":[{"id":"other"}]}`) }))
	defer other.Close()
	if modelAvailable(other.URL, other.Client()) {
		t.Fatal("unrelated service accepted")
	}
}
func TestTunnelClaimDoesNotStartTwice(t *testing.T) {
	m := &tunnelManager{}
	if !m.claim() || m.claim() {
		t.Fatal("concurrent tunnel allowed")
	}
	m.done()
	if !m.claim() {
		t.Fatal("retry blocked")
	}
}

func TestTunnelUsesRequestedLocalPort(t *testing.T) {
	if tunnelURL != "http://127.0.0.1:1923" || sshTunnelForward != "127.0.0.1:1923:127.0.0.1:8080" {
		t.Fatal("unexpected SSH forwarding")
	}
}

func TestTunnelRequiresReadyModelAndDoesNotFollowRedirects(t *testing.T) {
	loading := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[{"id":"deepseek-v4.1-flash","status":{"value":"loading"}}]}`)
	}))
	defer loading.Close()
	if modelAvailable(loading.URL, loading.Client()) {
		t.Fatal("loading model accepted")
	}
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, loading.URL, http.StatusFound) }))
	defer redirect.Close()
	if modelAvailable(redirect.URL, redirect.Client()) {
		t.Fatal("redirect accepted")
	}
}
func TestCloudCommandQuotesPathAndPort(t *testing.T) {
	command, err := cloudSSHCommand(`C:\Program Files\Google Cloud SDK\gcloud.cmd`)
	if err != nil || !strings.Contains(command, `"C:\Program Files\Google Cloud SDK\gcloud.cmd"`) || !strings.Contains(command, sshTunnelForward) {
		t.Fatal(command, err)
	}
	if _, err := cloudSSHCommand(`C:\x & bad.cmd`); err == nil {
		t.Fatal("unsafe path accepted")
	}
}

func TestModelReadinessWaitTimeoutCancelAndLoading(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		if cancelled {
			cancel()
		}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":[{"id":"deepseek-v4.1-flash","status":{"value":"loading"}}]}`)
		}))
		started := time.Now()
		err := waitForModel(ctx, srv.URL, &http.Client{}, time.Millisecond)
		cancel()
		srv.Close()
		if err == nil || time.Since(started) > time.Second {
			t.Fatal("unbounded/loading ready", err)
		}
	}
}
func TestModelReadinessBoundsHungHTTPWithoutClientTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := waitForModel(ctx, srv.URL, &http.Client{}, time.Millisecond); err != context.DeadlineExceeded {
		t.Fatal(err)
	}
}
func TestModelReadinessLoadedAndSSHFailureOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[{"id":"deepseek-v4.1-flash","status":{"value":"loaded"}}]}`)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := waitForModel(ctx, srv.URL, srv.Client(), time.Millisecond); err != nil {
		t.Fatal(err)
	}
	cmd, err := cloudSSHCommand(`C:\gcloud.cmd`)
	if err != nil {
		t.Fatal(err)
	}
	for _, option := range []string{"ExitOnForwardFailure=yes", "ConnectTimeout=20", "ServerAliveInterval=15", "ServerAliveCountMax=3"} {
		if !strings.Contains(cmd, option) {
			t.Fatal(cmd)
		}
	}
}
