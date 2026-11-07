package licensesrv

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIVersionHeaderOnV1Wrapper(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	w := httptest.NewRecorder()
	WithAPIVersion(inner).ServeHTTP(w, httptest.NewRequest("GET", "/v1/credits/status", nil))
	if w.Code != 204 || w.Header().Get("API-Version") != "1" {
		t.Fatal(w.Code, w.Header().Get("API-Version"))
	}
}

func TestV1AccountAliasPreservesLegacyAccountPath(t *testing.T) {
	h, _ := setup(t)
	a := accountTestHandler(t, h.Store)
	mux := http.NewServeMux()
	mux.Handle("/account/", a)
	mux.Handle("/v1/account/", StripV1AccountPrefix(a))
	// legacy
	w := accountPost(mux, "/account/register", "", map[string]string{"email": "ver@example.test", "password": accountPassword})
	if w.Code != 201 {
		t.Fatal("legacy", w.Code)
	}
	// versioned alias
	w = accountPost(mux, "/v1/account/login", "", map[string]string{"email": "ver@example.test", "password": accountPassword})
	if w.Code != 200 || w.Header().Get("API-Version") != "1" {
		t.Fatal("alias", w.Code, w.Header().Get("API-Version"))
	}
	_ = time.Now()
}
