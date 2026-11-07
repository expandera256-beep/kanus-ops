package i18n

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrefStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "lang.json")
	p := &PrefStore{Path: func() string { return path }}
	if p.Get() != "" {
		t.Fatal("empty store should return empty")
	}
	if err := p.Set("EN-gb"); err != nil {
		t.Fatal(err)
	}
	if p.Get() != "en" {
		t.Fatal("got", p.Get())
	}
	if fi, _ := os.Stat(path); fi.Mode().Perm() != 0o600 {
		t.Errorf("perm %v", fi.Mode().Perm())
	}
	if err := p.Set("xx"); err != ErrUnsupported {
		t.Fatal("unsupported accepted", err)
	}
	if p.Get() != "en" {
		t.Fatal("bad set changed value")
	}
	os.WriteFile(path, []byte(`{"lang":"zz"}`), 0o600)
	if p.Get() != "" {
		t.Fatal("unknown stored lang must be ignored")
	}
	os.WriteFile(path, []byte(`garbage`), 0o600)
	if p.Get() != "" {
		t.Fatal("corrupt file must be ignored")
	}
	var nilStore *PrefStore
	if nilStore.Get() != "" {
		t.Fatal("nil store")
	}
}

func TestPrefHandler(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lang.json")
	p := &PrefStore{Path: func() string { return path }}
	h := PrefHandler(p, nil)
	do := func(method, ct, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/api/lang", strings.NewReader(body))
		if ct != "" {
			r.Header.Set("Content-Type", ct)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := do("GET", "", ""); !strings.Contains(w.Body.String(), `"tr"`) {
		t.Fatal("default must be tr:", w.Body.String())
	}
	if w := do("POST", "application/x-www-form-urlencoded", "lang=en"); w.Code != http.StatusUnsupportedMediaType {
		t.Fatal("form post accepted", w.Code)
	}
	if w := do("POST", "application/json", `{"lang":"de"}`); w.Code != http.StatusBadRequest {
		t.Fatal("unsupported lang accepted", w.Code)
	}
	if w := do("POST", "application/json", `{`); w.Code != http.StatusBadRequest {
		t.Fatal("bad json", w.Code)
	}
	if w := do("POST", "application/json; charset=utf-8", `{"lang":"en"}`); w.Code != 200 {
		t.Fatal("valid post", w.Code)
	}
	if w := do("GET", "", ""); !strings.Contains(w.Body.String(), `"en"`) {
		t.Fatal("persisted value", w.Body.String())
	}
	if w := do("DELETE", "", ""); w.Code != http.StatusMethodNotAllowed {
		t.Fatal(w.Code)
	}
	// Handler honours stored preference when no lang query is given.
	r := httptest.NewRequest("GET", "/api/i18n", nil)
	w := httptest.NewRecorder()
	Handler(p.Get).ServeHTTP(w, r)
	if !strings.Contains(w.Body.String(), `"lang":"en"`) {
		t.Fatal("handler ignored preference")
	}
}
