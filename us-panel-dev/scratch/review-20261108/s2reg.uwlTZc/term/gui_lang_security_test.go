package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func postLang(t *testing.T, h http.Handler, origin string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:4477/api/lang", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func langHandler(t *testing.T) http.Handler {
	s := newView("langsec")
	s.root = t.TempDir()
	_ = os.Remove(langPref.Path())
	t.Cleanup(func() { _ = os.Remove(langPref.Path()) })
	return guiMux(&s)
}

// G-03: foreign Origin must get 403 like /api/act; same-origin still works.
func TestLangPostRejectsForeignOrigin(t *testing.T) {
	h := langHandler(t)
	if w := postLang(t, h, "https://evil.example", []byte(`{"lang":"en"}`)); w.Code != http.StatusForbidden {
		t.Fatalf("foreign origin status %d", w.Code)
	}
	if langPref.Get() == "en" {
		t.Fatal("foreign origin changed preference")
	}
	if w := postLang(t, h, "http://127.0.0.1:4477", []byte(`{"lang":"en"}`)); w.Code != http.StatusOK {
		t.Fatalf("same origin status %d", w.Code)
	}
	if w := postLang(t, h, "", []byte(`{"lang":"tr"}`)); w.Code != http.StatusOK {
		t.Fatalf("no origin status %d", w.Code)
	}
}

// G-04: trailing data / second document / oversized body are rejected.
func TestLangPostRejectsTrailingAndOversized(t *testing.T) {
	h := langHandler(t)
	for name, body := range map[string][]byte{
		"trailing garbage": []byte(`{"lang":"en"} xxx`),
		"second object":    []byte(`{"lang":"en"}{"lang":"tr"}`),
		"2MB body":         append([]byte(`{"lang":"en"}`), bytes.Repeat([]byte(" "), 2<<20)...),
	} {
		w := postLang(t, h, "", body)
		if w.Code != http.StatusBadRequest && w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("%s: status %d", name, w.Code)
		}
	}
	if langPref.Get() == "en" {
		t.Fatal("rejected body changed preference")
	}
	if w := postLang(t, h, "", []byte(`{"lang":"en"}`+"\n")); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "en") {
		t.Fatalf("valid body with newline: %d", w.Code)
	}
}
