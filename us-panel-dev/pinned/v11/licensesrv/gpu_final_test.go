package licensesrv

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNonLoopbackPinRequiredUnlessExplicitSystemCA(t *testing.T) {
	t.Setenv("US_GPU_WAKE_PIN_FILE", "")
	t.Setenv("US_GPU_WAKE_USE_SYSTEM_CA", "0")
	// Configuration-only: reserved .invalid name is never resolved or dialed.
	g, err := NewGPUWaker("https://gpu.invalid:8443/a100/start", "https://gpu.invalid:8443/a100/status", "/unused", 0)
	if err != nil {
		t.Fatal(err)
	}
	if g.PinFile != DefaultGPUWakePinFile {
		t.Fatal("missing mandatory default pin path")
	}
	a, err := NewGPUWakerAllow("https://gpu.invalid:8443/a100/start", "https://gpu.invalid:8443/a100/status", "/unused", 0, nil)
	if err != nil || a.PinFile != DefaultGPUWakePinFile {
		t.Fatal("allowlist constructor bypassed mandatory pin")
	}
	t.Setenv("US_GPU_WAKE_USE_SYSTEM_CA", "1")
	g, err = NewGPUWaker("https://gpu.invalid:8443/a100/start", "https://gpu.invalid:8443/a100/status", "/unused", 0)
	if err != nil || g.PinFile != "" {
		t.Fatal("explicit system CA opt-in ignored")
	}
	// An explicit pin must win even with CA opt-in, including if unreadable.
	t.Setenv("US_GPU_WAKE_PIN_FILE", filepath.Join(t.TempDir(), "missing"))
	g, _ = NewGPUWaker("https://gpu.invalid:8443/a100/start", "https://gpu.invalid:8443/a100/status", "/unused", 0)
	if _, err := g.client(g.WakeURL); err == nil {
		t.Fatal("missing explicit pin failed open")
	}
}

func unrelatedPEM(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cert := &x509.Certificate{SerialNumber: big.NewInt(42), Subject: pkix.Name{CommonName: "unrelated"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, IsCA: true, BasicConstraintsValid: true}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func TestPEMPinsUseNormalTLSAndWrongPinsFail(t *testing.T) {
	s, hits := tlsGPU(t)
	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: s.Certificate().Raw}))
	g := pinnedWaker(t, s, writePin(t, certPEM))
	cl, err := g.client(g.StatusURL)
	if err != nil {
		t.Fatal(err)
	}
	tr, ok := cl.Transport.(*http.Transport)
	if !ok || tr.TLSClientConfig == nil || tr.TLSClientConfig.InsecureSkipVerify || tr.TLSClientConfig.RootCAs == nil {
		t.Fatal("PEM mode must use normal root/SAN verification, not InsecureSkipVerify")
	}
	if !g.ModelReady(context.Background()) {
		t.Fatal("matching PEM rejected")
	}
	// Rotation: two PEM certificates, one current.
	g = pinnedWaker(t, s, writePin(t, unrelatedPEM(t)+certPEM))
	if !g.ModelReady(context.Background()) {
		t.Fatal("PEM rotation rejected")
	}
	before := hits.Load()
	wrongSPKI := sha256.Sum256([]byte("unrelated key"))
	for _, bad := range []string{unrelatedPEM(t), hex.EncodeToString(wrongSPKI[:])} {
		g = pinnedWaker(t, s, writePin(t, bad))
		if g.ModelReady(context.Background()) {
			t.Fatal("wrong PEM/SPKI accepted")
		}
	}
	if hits.Load() != before {
		t.Fatal("wrong pins reached application")
	}
}

func TestMissingPinNoWakeAndGenericResponse(t *testing.T) {
	s, hits := tlsGPU(t)
	g := pinnedWaker(t, s, filepath.Join(t.TempDir(), "missing.pin"))
	g.Trigger()
	st := waitState(t, g, "failed")
	w := httptest.NewRecorder()
	writeWarming(w, st)
	if hits.Load() != 0 || w.Code != 503 {
		t.Fatal("missing pin made wake request")
	}
	for _, v := range []string{"pin_unreadable", "token_unreadable", wakeSecret} {
		if contains := strings.Contains(w.Body.String(), v); contains {
			t.Fatal("internal state or token leaked")
		}
	}
}
