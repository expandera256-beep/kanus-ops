package main

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"
    "strings"
)
func TestModelAvailabilityRequiresMatchingModel(t *testing.T) {
    found := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
        if r.URL.Path != "/v1/models" { http.NotFound(w,r); return }
        fmt.Fprint(w, `{"data":[{"id":"deepseek-v4.1-flash"}]}`)
    }))
    defer found.Close()
    if !modelAvailable(found.URL, found.Client()) { t.Fatal("expected model") }
    other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) { fmt.Fprint(w, `{"data":[{"id":"other"}]}`) }))
    defer other.Close()
    if modelAvailable(other.URL, other.Client()) { t.Fatal("unrelated service accepted") }
}
func TestTunnelClaimDoesNotStartTwice(t *testing.T) {
    m := &tunnelManager{}
    if !m.claim() || m.claim() { t.Fatal("concurrent tunnel allowed") }
    m.done()
    if !m.claim() { t.Fatal("retry blocked") }
}

func TestTunnelUsesRequestedLocalPort(t *testing.T) {
    if tunnelURL != "http://127.0.0.1:1923" || sshTunnelForward != "127.0.0.1:1923:127.0.0.1:8080" { t.Fatal("unexpected SSH forwarding") }
}

func TestTunnelRequiresReadyModelAndDoesNotFollowRedirects(t *testing.T) {
    loading := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ fmt.Fprint(w, `{"data":[{"id":"deepseek-v4.1-flash","status":{"value":"loading"}}]}`) }))
    defer loading.Close()
    if modelAvailable(loading.URL, loading.Client()) { t.Fatal("loading model accepted") }
    redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ http.Redirect(w,r,loading.URL,http.StatusFound) }))
    defer redirect.Close()
    if modelAvailable(redirect.URL, redirect.Client()) { t.Fatal("redirect accepted") }
}
func TestCloudCommandQuotesPathAndPort(t *testing.T) {
    command, err := cloudSSHCommand(`C:\Program Files\Google Cloud SDK\gcloud.cmd`)
    if err != nil || !strings.Contains(command, `"C:\Program Files\Google Cloud SDK\gcloud.cmd"`) || !strings.Contains(command, sshTunnelForward) { t.Fatal(command, err) }
    if _, err := cloudSSHCommand(`C:\x & bad.cmd`); err == nil { t.Fatal("unsafe path accepted") }
}
