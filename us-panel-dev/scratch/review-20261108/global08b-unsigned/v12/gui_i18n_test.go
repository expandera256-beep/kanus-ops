package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"kiyas/i18n"
)

// TR markup must equal tr.json so the default (TR) panel is unchanged.
func TestGUIDataI18nMatchesTurkishDefault(t *testing.T) {
	b, err := i18n.Load()
	if err != nil {
		t.Fatal(err)
	}
	tr := b.For("tr")
	page := string(guiPage)
	re := regexp.MustCompile(`data-i18n="([^"]+)">([^<]*)<`)
	ms := re.FindAllStringSubmatch(page, -1)
	if len(ms) < 30 {
		t.Fatalf("expected >=30 translatable elements, got %d", len(ms))
	}
	for _, m := range ms {
		if got := tr.T(m[1]); got != m[2] {
			t.Errorf("%s: markup %q != tr.json %q", m[1], m[2], got)
		}
		if en := b.For("en").T(m[1]); en == m[1] || en == "" {
			t.Errorf("%s: no English text", m[1])
		}
	}
	for _, m := range regexp.MustCompile(`placeholder="([^"]*)"[^>]*data-i18n-ph="([^"]+)"`).FindAllStringSubmatch(page, -1) {
		if got := tr.T(m[2]); got != m[1] {
			t.Errorf("%s: placeholder %q != %q", m[2], m[1], got)
		}
	}
	for _, want := range []string{`<html lang="tr">`, `<title>US</title>`, `id="langsel"`, `fetch("/api/i18n"`, `document.documentElement.dir`} {
		if !strings.Contains(page, want) {
			t.Errorf("page missing %q", want)
		}
	}
}

func TestGUII18nEndpoint(t *testing.T) {
	s := newView("i18n")
	s.root = t.TempDir()
	h := guiMux(&s)
	for _, c := range []struct{ q, lang, sample string }{
		{"", "tr", "Dosya"}, {"?lang=en", "en", "Files"}, {"?lang=xx", "tr", "Dosya"}, {"?lang=EN-us", "en", "Files"},
	} {
		req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/api/i18n"+c.q, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Fatalf("%s: status %d", c.q, rec.Code)
		}
		var d i18n.Document
		if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
			t.Fatal(err)
		}
		if d.Lang != c.lang || d.Messages["gui.nav.files"] != c.sample || d.Dir != "ltr" || len(d.Languages) < 2 {
			t.Errorf("%s: got lang=%s dir=%s sample=%q", c.q, d.Lang, d.Dir, d.Messages["gui.nav.files"])
		}
	}
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/api/i18n", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST status %d", rec.Code)
	}
	req = httptest.NewRequest(http.MethodGet, "http://evil.example/api/i18n", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("foreign host status %d", rec.Code)
	}
}
