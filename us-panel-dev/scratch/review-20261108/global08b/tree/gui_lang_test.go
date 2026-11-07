package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestGUILangPreferencePersistsNextToDesk(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("US_DESK", filepath.Join(dir, "desk.json"))
	s := newView("lang")
	s.root = t.TempDir()
	h := guiMux(&s)
	get := func() string {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1/api/i18n", nil))
		return w.Body.String()
	}
	if !strings.Contains(get(), `"lang":"tr"`) {
		t.Fatal("default must stay tr")
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1/api/lang", strings.NewReader(`{"lang":"en"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	if langPref.Path() != filepath.Join(filepath.Dir(deskPath()), "lang.json") || langPref.Get() != "en" {
		t.Fatal("pref not stored next to desk")
	}
	if !strings.Contains(get(), `"lang":"en"`) {
		t.Fatal("pref not applied")
	}
}
