package licensesrv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestGPUWakeHTTPAllowlist(t *testing.T) {
	allow := ParseHTTPAllow("34.41.186.170:80, 10.0.0.5:8080")
	ok := [][2]string{
		{"http://34.41.186.170/a100/start", "http://34.41.186.170/a100/status"},
		{"http://34.41.186.170:80/a100/start", "http://34.41.186.170:80/a100/status"},
		{"http://10.0.0.5:8080/a100/start", "http://10.0.0.5:8080/a100/status"},
		{"https://gpu.example.org/a100/start", "https://gpu.example.org/a100/status"},
		{"http://127.0.0.1:9/a100/start", "http://127.0.0.1:9/a100/status"},
	}
	for _, c := range ok {
		if _, err := NewGPUWakerAllow(c[0], c[1], "/nonexistent", time.Hour, allow); err != nil {
			t.Fatalf("izinli reddedildi %v: %v", c, err)
		}
	}
	bad := []string{
		"http://34.41.186.171/a100/start",      // other host
		"http://34.41.186.170:8080/a100/start", // other port
		"http://10.0.0.5/a100/start",           // listed host, wrong port
		"ftp://34.41.186.170/a100/start",       // scheme
	}
	for _, u := range bad {
		if _, err := NewGPUWakerAllow(u, "https://gpu.example.org/a100/status", "/nonexistent", time.Hour, allow); err == nil {
			t.Fatalf("izinsiz kabul edildi: %s", u)
		}
	}
	if _, err := NewGPUWakerAllow("http://34.41.186.170/a100/start", "http://34.41.186.170/a100/status", "/x", time.Hour, nil); err == nil || !strings.Contains(err.Error(), "US_GPU_WAKE_HTTP_ALLOW") {
		t.Fatalf("allowlist yokken net hata beklenirdi: %v", err)
	}
	t.Setenv("US_GPU_WAKE_HTTP_ALLOW", "34.41.186.170:80")
	g, err := NewGPUWaker("http://34.41.186.170/a100/start", "http://34.41.186.170/a100/status", "/x", time.Hour)
	if err != nil || !g.InsecureHTTP() {
		t.Fatalf("env allowlist: %v", err)
	}
}

func TestGPUWakeRedirectToOtherHostRefused(t *testing.T) {
	var otherHits atomic.Int32
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		otherHits.Add(1)
		w.Write([]byte(`{"model_ready":true}`))
	}))
	defer other.Close()
	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+r.URL.Path, http.StatusFound)
	}))
	defer redir.Close()
	g, err := NewGPUWakerAllow(redir.URL+"/a100/start", redir.URL+"/a100/status", tokenFile(t), 45*time.Minute, nil)
	if err != nil {
		t.Fatal(err)
	}
	if g.ModelReady(context.Background()) {
		t.Fatal("yönlendirme sonrası hazır sayıldı")
	}
	g.PollMin, g.PollMax = time.Millisecond, time.Millisecond
	g.Trigger()
	waitState(t, g, "failed")
	if otherHits.Load() != 0 {
		t.Fatal("başka hosta yönlendirme izlendi (token sızabilirdi)")
	}
}

func TestGPUWarmingBodyHidesInternalReason(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "p")
	key := h.Store.st.Orders["555"][0]
	f := newFakeGPU(t, 202, 1)
	p, _ := NewModelProxy(h.Store, "http://127.0.0.1:1", "", "test-model", 45*time.Minute, 1, 32)
	g, _ := NewGPUWaker(f.srv.URL+"/a100/start", f.srv.URL+"/a100/status", filepath.Join(t.TempDir(), "missing.token"), 45*time.Minute)
	p.Waker = g
	proxyRequest(p, key, proxyInput, context.Background())
	waitState(t, g, "failed")
	w := proxyRequest(p, key, proxyInput, context.Background())
	body := w.Body.String()
	if w.Code != 503 || strings.Contains(body, "token") || strings.Contains(body, "reason") || !strings.Contains(body, "GPU ısınıyor / geçici hata") {
		t.Fatalf("istemciye iç durum sızdı: %s", body)
	}
	_ = os.Getenv
}
