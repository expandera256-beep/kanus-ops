package apiv1

import (
	"net/http"
	"strings"
)

const APIVersion = "1"

func WithAPIVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("API-Version", APIVersion)
		next.ServeHTTP(w, r)
	})
}

// MountAccountAlias rewrites /v1/account/* → /account/* without replacing legacy routes.
func MountAccountAlias(mux *http.ServeMux, accounts http.Handler) {
	mux.Handle("/v1/account/", WithAPIVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := r.Clone(r.Context())
		u := *r.URL
		u.Path = strings.TrimPrefix(r.URL.Path, "/v1")
		r2.URL = &u
		accounts.ServeHTTP(w, r2)
	})))
}
