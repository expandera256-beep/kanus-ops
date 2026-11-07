package apiv1

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithAPIVersionHeader(t *testing.T) {
	h := WithAPIVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/v1/x", nil))
	if w.Header().Get("API-Version") != "1" {
		t.Fatal(w.Header().Get("API-Version"))
	}
}

func TestMountAccountAlias(t *testing.T) {
	mux := http.NewServeMux()
	var saw string
	accounts := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		saw = r.URL.Path
		w.WriteHeader(200)
	})
	mux.Handle("/account/", accounts)
	MountAccountAlias(mux, accounts)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/v1/account/login", nil))
	if w.Code != 200 || saw != "/account/login" || w.Header().Get("API-Version") != "1" {
		t.Fatal(w.Code, saw, w.Header().Get("API-Version"))
	}
}
