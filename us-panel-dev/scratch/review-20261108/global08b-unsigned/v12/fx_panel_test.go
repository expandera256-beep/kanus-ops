package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"kiyas/i18n"
)

// Test binaries never reach a real host: only loopback is allowed.
type fxLoopbackOnly struct{ base http.RoundTripper }

func (t fxLoopbackOnly) RoundTrip(r *http.Request) (*http.Response, error) {
	if h := r.URL.Hostname(); h != "127.0.0.1" && h != "::1" && h != "localhost" {
		return nil, errors.New("test: non-loopback fx request blocked: " + h)
	}
	return t.base.RoundTrip(r)
}

func init() { fxHTTP.Transport = fxLoopbackOnly{fxHTTP.Transport} } // production TLS config, loopback only

var fxSnapZero i18n.FXSnapshot

const fxBody = `{"schema":"kanus.fx/1","source":"TCMB today.xml ForexSelling (döviz satış)","tcmb_tarih":"06.11.2026","date":"2026-11-06","base":"TRY","rates":{"USD":50.0724,"EUR":55.9609}}`

type fxSrv struct {
	*httptest.Server
	hits atomic.Int64
}

// newFXServer: TLS test server trusted via its own test CA (stands in for
// the system roots); production verification otherwise unchanged.
func newFXServer(t *testing.T, h func(w http.ResponseWriter, r *http.Request)) *fxSrv {
	s := &fxSrv{}
	s.Server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { s.hits.Add(1); h(w, r) }))
	old, oldBase := fxHTTP, fxBaseOverride
	fxHTTP = &http.Client{Timeout: 2 * time.Second, Transport: fxLoopbackOnly{s.Server.Client().Transport}, CheckRedirect: old.CheckRedirect}
	fxBaseOverride = s.URL
	t.Cleanup(func() { fxWaitIdle(t); s.Close(); fxHTTP = old; fxBaseOverride = oldBase })
	return s
}

func serve(body string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != fxPath {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write([]byte(body))
	}
}

var fxLogMu sync.Mutex

func fxCaptureLog(t *testing.T) *[]string {
	var lines []string
	old := fxLogf
	fxLogf = func(f string, a ...any) {
		fxLogMu.Lock()
		lines = append(lines, fmt.Sprintf(f, a...))
		fxLogMu.Unlock()
	}
	t.Cleanup(func() { fxLogf = old })
	return &lines
}

func fxReset(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("US_DESK", filepath.Join(dir, "desk.json"))
	fxSetState(false, time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)) // no background refresh unless a test asks
	oldNow := fxNow
	fxNow = func() time.Time { return time.Date(2026, 11, 8, 0, 37, 0, 0, time.UTC) }
	t.Cleanup(func() { fxWaitIdle(t); fxNow = oldNow; fxSetState(false, time.Time{}); _ = fxCache.Close() })
}

func fxSetState(loaded bool, next time.Time) {
	fx.mu.Lock()
	fx.snap, fx.ok, fx.inflight, fx.loaded, fx.next = fxSnapZero, false, false, loaded, next
	fx.mu.Unlock()
}

func fxWaitIdle(t *testing.T) {
	for i := 0; i < 500; i++ {
		fx.mu.Lock()
		busy := fx.inflight
		fx.mu.Unlock()
		if !busy {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("fx refresh did not finish")
}

func cacheFile() string { return filepath.Join(filepath.Dir(os.Getenv("US_DESK")), "fx.json") }

// FX08B-01: a valid 4096-byte response must persist and reload.
func TestFXCache4096Threshold(t *testing.T) {
	fxReset(t)
	body := fxBody + strings.Repeat(" ", 4096-len(fxBody))
	s := newFXServer(t, serve(body))
	if _, _, _, err := fetchFX(s.URL); err != nil {
		t.Fatal(err)
	}
	refreshFX(s.URL)
	raw, err := os.ReadFile(cacheFile())
	if err != nil || len(raw) != 4096 || string(raw) != body {
		t.Fatal("4096-byte body not cached", len(raw), err)
	}
	fxSetState(false, time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC))
	if got, ok := fxFromCache(); !ok || got.USD != 50.0724 {
		t.Fatal("cache not reusable", got, ok)
	}
	// 4097 bytes is rejected at fetch, so nothing larger can reach the cache.
	s2 := newFXServer(t, serve(body+" "))
	if _, _, _, err := fetchFX(s2.URL); err == nil {
		t.Fatal("4097-byte body accepted")
	}
}

func TestFXCacheWriteFailureLogged(t *testing.T) {
	fxReset(t)
	logs := fxCaptureLog(t)
	s := newFXServer(t, serve(fxBody))
	if err := os.Mkdir(cacheFile(), 0o700); err != nil { // a directory where the file should be
		t.Fatal(err)
	}
	refreshFX(s.URL)
	fxLogMu.Lock()
	defer fxLogMu.Unlock()
	if len(*logs) != 1 || (*logs)[0] != "us: fx cache not saved" {
		t.Fatal(*logs)
	}
}

func TestFXRejects(t *testing.T) {
	fxReset(t)
	cases := map[string]func(w http.ResponseWriter, r *http.Request){
		"stale":     serve(strings.Replace(fxBody, "2026-11-06", "2026-11-02", 1)),
		"future":    serve(strings.Replace(fxBody, "2026-11-06", "2026-11-09", 1)),
		"schema":    serve(strings.Replace(fxBody, "kanus.fx/1", "x", 1)),
		"range":     serve(strings.Replace(fxBody, "55.9609", "5000", 1)),
		"string":    serve(strings.Replace(fxBody, "55.9609", `"55.9609"`, 1)),
		"rate_type": serve(strings.Replace(fxBody, `"base"`, `"rate_type":"forex_buying","base"`, 1)),
		"500":       func(w http.ResponseWriter, r *http.Request) { http.Error(w, "x", 500) },
		"huge":      func(w http.ResponseWriter, r *http.Request) { serve(strings.Repeat(" ", 5000)+fxBody)(w, r) },
		"slow": func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(3 * time.Second)
			serve(fxBody)(w, r)
		},
	}
	for name, h := range cases {
		s := newFXServer(t, h)
		if _, _, _, err := fetchFX(s.URL); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestFXOptionalRateTypeAndUnknownFields(t *testing.T) {
	fxReset(t)
	body := strings.Replace(fxBody, `"base"`, `"rate_type":"forex_selling","future_field":{"x":1},"base"`, 1)
	s := newFXServer(t, serve(body))
	if snap, _, _, err := fetchFX(s.URL); err != nil || snap.EUR != 55.9609 {
		t.Fatal(err, snap)
	}
}

func TestFX503EnvelopeKeepsLastValid(t *testing.T) {
	fxReset(t)
	logs := fxCaptureLog(t)
	var down atomic.Bool
	s := newFXServer(t, func(w http.ResponseWriter, r *http.Request) {
		if down.Load() {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"fx_unavailable","code":"fx_unavailable","message":"FX rate unavailable","request_id":"req-1"}`))
			return
		}
		serve(fxBody)(w, r)
	})
	refreshFX(s.URL)
	down.Store(true)
	if _, _, _, err := fetchFX(s.URL); !errors.Is(err, errFXUnavailable) {
		t.Fatal("envelope not recognised", err)
	}
	refreshFX(s.URL)
	if snap, ok := currentFX(); !ok || snap.EUR != 55.9609 {
		t.Fatal("last valid snapshot dropped on 503", ok)
	}
	fxLogMu.Lock()
	joined := strings.Join(*logs, "\n")
	fxLogMu.Unlock()
	if !strings.Contains(joined, "fx-unavailable") || strings.Contains(joined, "127.0.0.1") || strings.Contains(joined, "req-1") {
		t.Fatal("log must hold only a category", joined)
	}
	// After the 5-day window the snapshot is no longer shown.
	fx.mu.Lock()
	fx.next = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	fx.mu.Unlock()
	fxNow = func() time.Time { return time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC) }
	if _, ok := currentFX(); ok {
		t.Fatal("stale snapshot shown")
	}
}

func TestFXStandardTLSOnly(t *testing.T) {
	fxReset(t)
	c := fxTransport().TLSClientConfig
	if c == nil || c.InsecureSkipVerify || c.MinVersion < tls.VersionTLS12 || c.VerifyPeerCertificate != nil || c.VerifyConnection != nil || c.RootCAs != nil {
		t.Fatal("FX must use standard verification with system roots", c)
	}
	src, err := os.ReadFile("fx_panel.go")
	if err != nil {
		t.Fatal(err)
	}
	if regexp.MustCompile(`InsecureSkipVerify\s*:|VerifyPeerCertificate\s*[:=]|VerifyConnection\s*[:=]|RootCAs\s*[:=]`).Match(src) {
		t.Fatal("fx_panel.go must not override certificate verification")
	}
	// Self-signed certificate with the production client: rejected.
	srv := httptest.NewTLSServer(http.HandlerFunc(serve(fxBody)))
	defer srv.Close()
	if _, _, _, err := fetchFX(srv.URL); err == nil || !strings.Contains(err.Error(), "certificate") {
		t.Fatal("untrusted certificate accepted", err)
	}
	// Plain HTTP: refused before any request.
	var hits atomic.Int64
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1); serve(fxBody)(w, r) }))
	defer plain.Close()
	if _, _, _, err := fetchFX(plain.URL); err == nil || hits.Load() != 0 {
		t.Fatal("http accepted", err, hits.Load())
	}
	// Redirect is not followed (target never contacted).
	target := newFXServer(t, serve(fxBody))
	redir := newFXServer(t, func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL+fxPath, http.StatusFound) })
	if _, _, _, err := fetchFX(redir.URL); err == nil || target.hits.Load() != 0 {
		t.Fatal("redirect followed", err, target.hits.Load())
	}
}

func TestFXHostFixed(t *testing.T) {
	oldL, oldG, oldO := embeddedLicenseURL, embeddedGatewayURL, fxBaseOverride
	defer func() { embeddedLicenseURL, embeddedGatewayURL, fxBaseOverride = oldL, oldG, oldO }()
	fxBaseOverride = ""
	t.Setenv("US_GATEWAY_URL", "https://evil.example")
	t.Setenv("US_LICENSE_URL", "https://evil.example")
	for _, c := range []struct{ lic, gw, want string }{
		{"", "", fxLiveBase},
		{"https://api-test.kanus.org", "", fxTestBase},
		{"", "https://api-test.kanus.org/", fxTestBase},
		{"https://api.kanus.org", "", fxLiveBase},
		{"https://evil.example", "http://127.0.0.1:1", fxLiveBase},
	} {
		embeddedLicenseURL, embeddedGatewayURL = c.lic, c.gw
		if got := fxBase(); got != c.want {
			t.Errorf("lic=%q gw=%q: %s want %s", c.lic, c.gw, got, c.want)
		}
	}
}

func TestFXHTTPBaseShowsTRYOnly(t *testing.T) {
	fxReset(t)
	var hits atomic.Int64
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1); serve(fxBody)(w, r) }))
	defer plain.Close()
	old := fxBaseOverride
	fxBaseOverride = plain.URL
	defer func() { fxBaseOverride = old }()
	fxSetState(true, time.Time{})
	if _, ok := currentFX(); ok {
		t.Fatal("fx from http base")
	}
	fxWaitIdle(t)
	if txt := pricingSummaryFor("tr"); strings.Contains(txt, "≈") || hits.Load() != 0 {
		t.Fatal("http base contacted or FX shown", txt, hits.Load())
	}
}

func TestFXCacheRevalidated(t *testing.T) {
	fxReset(t)
	for name, body := range map[string]string{
		"range": strings.Replace(fxBody, "55.9609", "5000", 1),
		"stale": strings.Replace(fxBody, "2026-11-06", "2026-11-01", 1),
		"junk":  "not json",
	} {
		if err := os.WriteFile(cacheFile(), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, ok := fxFromCache(); ok {
			t.Errorf("%s cache accepted", name)
		}
	}
}

func TestPricingSummaryFXAndFallback(t *testing.T) {
	fxReset(t)
	for _, lang := range []string{"tr", "en", "de", "ru"} {
		txt := pricingSummaryFor(lang)
		if strings.Contains(txt, "≈") || strings.Contains(txt, "USD") || strings.Contains(txt, "50,05") || strings.Contains(txt, "50.05") || strings.Contains(txt, "price.") {
			t.Fatalf("%s: %s", lang, txt)
		}
	}
	if txt := pricingSummaryFor("tr"); !strings.Contains(txt, "18.900 TL") || !strings.Contains(txt, "100 istem 9.900 TL") {
		t.Fatal(txt)
	}
	s := newFXServer(t, serve(fxBody))
	refreshFX(s.URL)
	want := map[string][]string{
		"tr": {"Aylık abonelik 18.900 TL (≈ 337,74 € · ≈ 377,45 $)", "ödeme TL ile alınır", "06.11.2026"},
		"en": {"Monthly subscription 18,900 TRY (≈ 337.74 € · ≈ $377.45)", "charged in Turkish lira"},
		"de": {"Monatsabo 18.900 TL (≈ 337,74 € · ≈ 377,45 $)", "türkischen Lira"},
		"ru": {"Месячная подписка 18\u00a0900 TL", "турецких лирах"},
	}
	for lang, w := range want {
		txt := pricingSummaryFor(lang)
		for _, s := range w {
			if !strings.Contains(txt, s) {
				t.Errorf("%s: missing %q in %q", lang, s, txt)
			}
		}
	}
}

func TestPricingHandler(t *testing.T) {
	fxReset(t)
	w := httptest.NewRecorder()
	pricingHandler(w, httptest.NewRequest("GET", "/api/pricing?lang=de", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Monatsabo") {
		t.Fatal(w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	pricingHandler(w, httptest.NewRequest("POST", "/api/pricing", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatal(w.Code)
	}
}

func TestFXBackgroundRefreshAndDiskReload(t *testing.T) {
	fxReset(t)
	s := newFXServer(t, serve(fxBody))
	fxSetState(true, time.Time{})
	if _, ok := currentFX(); ok {
		t.Fatal("first call must not block on network")
	}
	fxWaitIdle(t)
	if snap, ok := currentFX(); !ok || snap.EUR != 55.9609 || s.hits.Load() != 1 {
		t.Fatal(ok, snap, s.hits.Load())
	}
	fxSetState(false, time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC))
	if snap, ok := currentFX(); !ok || snap.USD != 50.0724 {
		t.Fatal("disk cache not used", ok)
	}
}
