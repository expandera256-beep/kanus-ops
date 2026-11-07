package licensesrv

import (
	"net/http"
	"strings"
)

const APIVersion = "1"

// WithAPIVersion sets API-Version: 1 on responses for versioned surfaces.
func WithAPIVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("API-Version", APIVersion)
		next.ServeHTTP(w, r)
	})
}

// StripV1AccountPrefix rewrites /v1/account/... → /account/... then delegates.
func StripV1AccountPrefix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("API-Version", APIVersion)
		path := r.URL.Path
		if strings.HasPrefix(path, "/v1/account/") {
			r2 := r.Clone(r.Context())
			u := *r.URL
			u.Path = strings.TrimPrefix(path, "/v1")
			r2.URL = &u
			next.ServeHTTP(w, r2)
			return
		}
		http.NotFound(w, r)
	})
}
