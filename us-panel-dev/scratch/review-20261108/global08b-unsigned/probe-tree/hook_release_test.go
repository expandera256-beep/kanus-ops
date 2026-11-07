package main

import (
	"context"
	"fmt"
	"kiyas/licensesrv"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func useHookRelease(t *testing.T, base string) {
	t.Helper()
	oldAuthority, oldChannel, oldLicense, oldGateway := releaseAuthority, releaseChannel, embeddedLicenseURL, embeddedGatewayURL
	releaseAuthority, releaseChannel, embeddedLicenseURL, embeddedGatewayURL = "go-hook", "commercial", base, "https://old-python.example"
	t.Cleanup(func() {
		releaseAuthority, releaseChannel, embeddedLicenseURL, embeddedGatewayURL = oldAuthority, oldChannel, oldLicense, oldGateway
	})
	t.Setenv("US_LICENSE_URL", "")
	t.Setenv("US_GATEWAY_URL", "")
	t.Setenv("US_LICENSE_SECRET", "")
	t.Setenv("US_MODEL_NAME", "")
	t.Setenv("US_LICENSE_LEDGER", filepath.Join(t.TempDir(), "license.json"))
}
func TestHookReleaseDefaultsNeverSelectPythonOrCloudRun(t *testing.T) {
	useHookRelease(t, "https://api-test.kanus.org")
	if activeGateway() != activeLicenseURL() || activeLicenseURL() != "https://api-test.kanus.org" || len(doorBases()) != 1 || doorBases()[0] != activeLicenseURL() || selectedModel("a100") != tunnelModel {
		t.Fatal("incorrect hook defaults")
	}
	s := newView("a100")
	s.accountGateway = activeGateway()
	s.key = strings.Repeat("old-python-token", 4)
	s.admin = true
	if s.modelKey() != "" || s.licenseBlock(time.Now()) == "" {
		t.Fatal("stale account/admin bypass")
	}
	gate := s.creditGateway()
	if gate.required || gate.base != activeLicenseURL() {
		t.Fatal("hook bypassed")
	}
	if s.allowanceLine(time.Now()) == "" || !strings.Contains(s.connectionStatus(), "Go") {
		t.Fatal("wrong profile UI")
	}
}
func TestHookReleaseUsesRemoteLedgerWithoutDesktopSigningSecret(t *testing.T) {
	_, st, key := gatewayCreditFixture(t)
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		l, _ := st.License(key)
		if l.Reserved != 1 || r.Header.Get("Authorization") != "" {
			t.Error("reservation or credential leak", l)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"merhaba\"}}]}\n\ndata: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n")
	}))
	defer upstream.Close()
	proxy, err := licensesrv.NewModelProxy(st, upstream.URL, "", tunnelModel, 45*time.Minute, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(proxy)
	defer srv.Close()
	useHookRelease(t, srv.URL)
	s := newView("a100")
	s.root = t.TempDir()
	s.modelCh = make(chan modelResult, 4)
	if note := s.acceptLicense(key, time.Now()); !strings.Contains(note, "kaydedildi") {
		t.Fatal(note)
	}
	if s.licenseBlock(time.Now()) != "" || licenseSecret() != "" {
		t.Fatal("desktop secret required")
	}
	s.admin = true
	s.accountGateway = srv.URL
	s.key = strings.Repeat("old-session", 5)
	if s.modelKey() != key {
		t.Fatal("license missing from proxy request")
	}
	out := postDoorContext(context.Background(), "a100", s.modelKey(), "selam", "system", 64, "", 0, nil, defaultContextLimits())
	l, _ := st.License(key)
	if !out.ok || calls != 1 || l.Credits != 1 || l.Reserved != 0 {
		t.Fatal(out, calls, l)
	}
	out = postDoorContext(context.Background(), "a100", "us1.unknown.signature", "selam", "system", 64, "", 0, nil, defaultContextLimits())
	if out.ok || calls != 1 {
		t.Fatal("unknown license reached upstream")
	}
}
func TestHookReleaseDisablesPythonAccountOperationsAndForm(t *testing.T) {
	useHookRelease(t, "https://api-test.kanus.org")
	s := newView("a100")
	s.root = t.TempDir()
	if !strings.Contains(string(renderGUIPage(guiPage)), `<form id="accountform" hidden>`) {
		t.Fatal("account form not hidden")
	}
	mux := guiMux(&s)
	for _, op := range []string{"account-login", "account-register", "account-redeem", "account-balance", "account-requests", "account-workers"} {
		r := httptest.NewRequest(http.MethodPost, "/api/act", strings.NewReader(`{"op":"`+op+`"}`))
		r.Host = "127.0.0.1:8788"
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		// The hook profile must reject the action before any account request.
		if w.Code != http.StatusConflict {
			t.Fatal(op, w.Code)
		}
	}
}
func TestHookReleaseLicenseURLOverrideDoesNotChangeModelEndpoint(t *testing.T) {
	useHookRelease(t, "https://api-test.kanus.org")
	t.Setenv("US_GATEWAY_URL", "http://127.0.0.1:1923")
	t.Setenv("US_DEV_DIRECT_MODEL", "1")
	t.Setenv("US_LICENSE_URL", "https://another-license.example/")
	if activeLicenseURL() != "https://another-license.example" || activeGateway() != activeLicenseURL() {
		t.Fatal("license/model URLs mixed")
	}
}

func TestHookReleaseColdStartIsNotSuccessAndWaits45Minutes(t *testing.T) {
	useHookRelease(t, "https://api-test.kanus.org")
	if completionWait < 45*time.Minute {
		t.Fatal(completionWait)
	}
	s := newView("a100")
	s.markCallStart()
	_, load, _, _ := s.lifeLines(time.Now())
	if !s.calling || !strings.Contains(load, "GPU ısınıyor") || strings.Contains(load, "hazır") {
		t.Fatal(load)
	}
	s.noteLive(120)
	_, load, _, _ = s.lifeLines(time.Now())
	if load != "Model yanıtı akıyor" {
		t.Fatal(load)
	}
	t.Setenv("US_GATEWAY_URL", tunnelURL)
	t.Setenv("US_DEV_DIRECT_MODEL", "1")
	if activeGateway() != activeLicenseURL() {
		t.Fatal("commercial direct bypass")
	}
}
