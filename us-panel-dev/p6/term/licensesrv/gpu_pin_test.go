package licensesrv

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func tlsGPU(t *testing.T) (*httptest.Server, *atomic.Int32) {
	var hits atomic.Int32
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+wakeSecret {
			http.Error(w, "forbidden", 403)
			return
		}
		hits.Add(1)
		if r.URL.Path == "/a100/start" {
			w.WriteHeader(202)
			return
		}
		w.Write([]byte(`{"instance":"RUNNING","model_ready":true}`))
	}))
	t.Cleanup(s.Close)
	return s, &hits
}

func writePin(t *testing.T, body string) string {
	p := filepath.Join(t.TempDir(), "wake.pin")
	os.WriteFile(p, []byte(body), 0o600)
	return p
}

func pinnedWaker(t *testing.T, s *httptest.Server, pinFile string) *GPUWaker {
	t.Setenv("US_GPU_WAKE_PIN_FILE", pinFile)
	g, err := NewGPUWaker(s.URL+"/a100/start", s.URL+"/a100/status", tokenFile(t), 45*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	g.PollMin, g.PollMax = time.Millisecond, 5*time.Millisecond
	return g
}

func TestGPUPinSPKIAndPEMAndRotation(t *testing.T) {
	s, _ := tlsGPU(t)
	leaf := s.Certificate()
	sum := sha256.Sum256(leaf.RawSubjectPublicKeyInfo)
	other := sha256.Sum256([]byte("old key"))
	cases := map[string]string{
		"hex":           hex.EncodeToString(sum[:]),
		"base64+prefix": "sha256/" + base64.StdEncoding.EncodeToString(sum[:]),
		"rotation":      "# eski\nsha256/" + base64.StdEncoding.EncodeToString(other[:]) + "\n" + hex.EncodeToString(sum[:]) + "\n",
		"pem":           string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leaf.Raw})),
	}
	for name, body := range cases {
		g := pinnedWaker(t, s, writePin(t, body))
		if !g.ModelReady(context.Background()) {
			t.Fatalf("%s: eşleşen pin reddedildi", name)
		}
	}
}

func TestGPUPinMismatchMissingAndSystemCA(t *testing.T) {
	s, hits := tlsGPU(t)
	wrong := sha256.Sum256([]byte("başka sertifika"))
	g := pinnedWaker(t, s, writePin(t, hex.EncodeToString(wrong[:])))
	if g.ModelReady(context.Background()) {
		t.Fatal("eşleşmeyen sertifika kabul edildi")
	}
	g.Trigger()
	waitState(t, g, "failed")
	// Missing pin file (configured): fail closed, nothing reaches the server.
	g2 := pinnedWaker(t, s, filepath.Join(t.TempDir(), "yok.pin"))
	g2.Trigger()
	if st := waitState(t, g2, "failed"); !strings.Contains(st.Reason, "pin") {
		t.Fatalf("sebep: %+v", st)
	}
	if hits.Load() != 0 {
		t.Fatalf("pin yok/yanlışken sunucuya %d istek ulaştı", hits.Load())
	}
	// No pin file: system CA → self-signed rejected.
	t.Setenv("US_GPU_WAKE_PIN_FILE", "")
	g3, _ := NewGPUWaker(s.URL+"/a100/start", s.URL+"/a100/status", tokenFile(t), 45*time.Minute)
	if g3.PinFile != "" || g3.ModelReady(context.Background()) {
		t.Fatal("sistem CA modu kendinden imzalı sertifikayı kabul etti")
	}
}

func TestGPUPinRedirectAndTokenNeverLogged(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	var otherHits atomic.Int32
	other := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { otherHits.Add(1) }))
	defer other.Close()
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+r.URL.Path, http.StatusFound)
	}))
	defer s.Close()
	sum := sha256.Sum256(s.Certificate().RawSubjectPublicKeyInfo)
	g := pinnedWaker(t, s, writePin(t, hex.EncodeToString(sum[:])))
	g.Trigger()
	waitState(t, g, "failed")
	if otherHits.Load() != 0 {
		t.Fatal("başka hosta yönlendirme izlendi")
	}
	if strings.Contains(buf.String(), wakeSecret) {
		t.Fatal("token loglandı")
	}
}

func TestGPUPinRightKeyWrongHostRejected(t *testing.T) {
	s, _ := tlsGPU(t)
	sum := sha256.Sum256(s.Certificate().RawSubjectPublicKeyInfo)
	u := strings.Replace(s.URL, "127.0.0.1", "localhost", 1) // not in the cert's SANs
	t.Setenv("US_GPU_WAKE_PIN_FILE", writePin(t, hex.EncodeToString(sum[:])))
	g, err := NewGPUWaker(u+"/a100/start", u+"/a100/status", tokenFile(t), 45*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if g.ModelReady(context.Background()) {
		t.Fatal("pin doğru ama host/IP SAN uyuşmazken kabul edildi")
	}
}

func TestGPUWakeInfoLogLine(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	s, _ := tlsGPU(t)
	sum := sha256.Sum256(s.Certificate().RawSubjectPublicKeyInfo)
	g := pinnedWaker(t, s, writePin(t, hex.EncodeToString(sum[:])))
	g.ModelReady(context.Background())
	out := buf.String()
	host := strings.TrimPrefix(s.URL, "https://")
	for _, want := range []string{"gpu_wake", "scheme=https", "host=" + host, "path=/a100/status", "status=200", "pin_ok=yes", "duration_ms="} {
		if !strings.Contains(out, want) {
			t.Fatalf("log satırında %q yok: %s", want, out)
		}
	}
	if strings.Contains(out, wakeSecret) {
		t.Fatal("token loglandı")
	}
}
