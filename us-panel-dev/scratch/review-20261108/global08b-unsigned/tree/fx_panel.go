package main

// Approximate EUR/USD next to TRY prices (display only). Source: the
// license server's GET /v1/fx (FX_CONTRACT_FOR_US_API.md, GLOBAL proposal).
// Demir 08.11.2026 01:13 TSİ: unsigned, HTTPS only, no new key. Trust rests
// on TLS certificate verification (system roots, never InsecureSkipVerify),
// a fixed host (fxBase), no redirects, 5 s timeout, 4 KB limit and the
// schema/freshness/range checks in i18n.ParseFX. Anything else => TRY only;
// there is no fixed-rate fallback. Payments and the ledger stay TRY.

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"kiyas/i18n"
)

const (
	fxPath   = "/v1/fx"
	fxMinTTL = 5 * time.Minute
	fxMaxTTL = 6 * time.Hour
)

// fxCache: last verified response next to lang.json (same safe-write rules).
var fxCache = &i18n.PrefStore{Name: "fx.json", Root: func() string {
	desk := deskPath()
	if desk == "" {
		return ""
	}
	return filepath.Dir(desk)
}}

var errFXUnavailable = errors.New("fx: server reports fx_unavailable")

// fxLogReason logs only a fixed category: never certificate content,
// SPKI hashes, URLs or response bodies.
func fxLogReason(err error) string {
	switch {
	case errors.Is(err, errFXUnavailable):
		return "fx-unavailable"
	case err != nil && strings.Contains(err.Error(), "certificate"):
		return "tls-certificate"
	default:
		return "fetch-or-validate"
	}
}

var fxLogf = log.Printf

// fxTransport: standard TLS only (Demir 08.11.2026 01:22 TSİ: no pin).
// System roots + hostname verification by crypto/tls, TLS >= 1.2; no
// InsecureSkipVerify, no custom verification callbacks.
func fxTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	return t
}

var fxHTTP = &http.Client{Timeout: 5 * time.Second, Transport: fxTransport(), CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

// fxNow is replaced in tests.
var fxNow = time.Now

type fxState struct {
	mu       sync.Mutex
	snap     i18n.FXSnapshot
	ok       bool
	next     time.Time // next allowed network attempt
	inflight bool
	loaded   bool // disk cache read once
}

var fx fxState

// FX host is fixed (HOST 08.11.2026): live https://api.kanus.org, test
// https://api-test.kanus.org. Only build-time constants select it
// (embeddedLicenseURL / embeddedGatewayURL via -ldflags, see US_DERLE.bat);
// US_GATEWAY_URL / US_LICENSE_URL and the GUI gateway field are ignored, so
// users cannot redirect FX. Any other build value falls back to live.
const fxLiveBase, fxTestBase = "https://api.kanus.org", "https://api-test.kanus.org"

// fxBaseOverride is for tests only (httptest servers); production never sets it.
var fxBaseOverride string

func fxBase() string {
	if fxBaseOverride != "" {
		return fxBaseOverride
	}
	for _, b := range []string{embeddedLicenseURL, embeddedGatewayURL} {
		switch strings.TrimRight(strings.TrimSpace(b), "/") {
		case fxTestBase:
			return fxTestBase
		case fxLiveBase:
			return fxLiveBase
		}
	}
	return fxLiveBase
}

// fxHTTPS: only an https:// base may be used for FX.
func fxHTTPS(base string) bool { return strings.HasPrefix(strings.ToLower(base), "https://") }

func fxTTL(cc string) time.Duration {
	ttl := time.Hour
	for _, part := range strings.Split(cc, ",") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(strings.ToLower(part)), "max-age="); ok {
			if n, err := strconv.Atoi(v); err == nil {
				ttl = time.Duration(n) * time.Second
			}
		}
	}
	return min(max(ttl, fxMinTTL), fxMaxTTL)
}

// fetchFX performs one verified GET base/v1/fx.
func fetchFX(base string) (i18n.FXSnapshot, []byte, time.Duration, error) {
	var z i18n.FXSnapshot
	base = strings.TrimRight(base, "/")
	if !fxHTTPS(base) {
		return z, nil, 0, errors.New("fx: https required")
	}
	req, err := http.NewRequest(http.MethodGet, base+fxPath, nil)
	if err != nil {
		return z, nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := fxHTTP.Do(req)
	if err != nil {
		return z, nil, 0, err
	}
	defer resp.Body.Close()
	if resp.TLS == nil {
		return z, nil, 0, errors.New("fx: no TLS")
	}
	if resp.Request == nil || resp.Request.URL.Host != req.URL.Host {
		return z, nil, 0, errors.New("fx: host changed")
	}
	if resp.StatusCode == http.StatusServiceUnavailable {
		var e struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}
		if b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096)); json.Unmarshal(b, &e) == nil && (e.Code == "fx_unavailable" || e.Error == "fx_unavailable") {
			return z, nil, 0, errFXUnavailable
		}
	}
	if resp.StatusCode != http.StatusOK {
		return z, nil, 0, errors.New("fx: status " + strconv.Itoa(resp.StatusCode))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil || len(body) > 4096 {
		return z, nil, 0, errors.New("fx: body")
	}
	s, err := i18n.ParseFX(body, fxNow())
	if err != nil {
		return z, nil, 0, err
	}
	return s, body, fxTTL(resp.Header.Get("Cache-Control")), nil
}

// fxFromCache re-validates the cached body (schema, ranges, freshness).
func fxFromCache() (i18n.FXSnapshot, bool) {
	raw, err := fxCache.ReadRaw()
	if err != nil {
		return i18n.FXSnapshot{}, false
	}
	s, err := i18n.ParseFX(raw, fxNow())
	return s, err == nil
}

// refreshFX fetches synchronously; on success updates memory and disk.
func refreshFX(base string) {
	s, body, ttl, err := fetchFX(base)
	fx.mu.Lock()
	fx.inflight = false
	if err != nil {
		fx.next = fxNow().Add(fxMinTTL)
		fx.mu.Unlock()
		// Last valid snapshot (memory/disk) stays in use until FXFresh says stale.
		fxLogf("us: fx refresh failed: %s", fxLogReason(err))
		return
	}
	fx.snap, fx.ok, fx.next = s, true, fxNow().Add(ttl)
	fx.mu.Unlock()
	if err := fxCache.WriteRaw(body); err != nil {
		fxLogf("us: fx cache not saved")
	}
}

// currentFX never blocks on the network: it returns the last verified,
// still-fresh snapshot (memory, then disk) and starts a background refresh
// when due. ok=false => show TRY only.
func currentFX() (i18n.FXSnapshot, bool) {
	base := fxBase()
	fx.mu.Lock()
	defer fx.mu.Unlock()
	if !fxHTTPS(base) {
		return i18n.FXSnapshot{}, false // http/empty gateway: TRY only, no request
	}
	if !fx.loaded {
		fx.loaded = true
		fx.mu.Unlock()
		s, ok := fxFromCache()
		fx.mu.Lock()
		if ok && !fx.ok {
			fx.snap, fx.ok = s, true
		}
	}
	now := fxNow()
	if fx.ok && fxStale(fx.snap, now) {
		fx.ok = false
	}
	if !fx.inflight && !now.Before(fx.next) {
		fx.inflight = true
		go refreshFX(base)
	}
	return fx.snap, fx.ok
}

func fxStale(s i18n.FXSnapshot, now time.Time) bool { return !i18n.FXFresh(s.Date, now) }
