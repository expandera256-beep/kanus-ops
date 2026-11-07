package licensesrv

import (
	"net/http"
	"net/http/httptest"
)

// withModels makes a fake upstream answer the readiness probe
// GET /v1/models with the configured test model.
func withModels(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v1/models" && r.Header.Get("X-US-Ready-Probe") == "1" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"test-model"}]}`))
			return
		}
		h.ServeHTTP(w, r)
	})
}

func newModelsServer(h http.Handler) *httptest.Server { return httptest.NewServer(withModels(h)) }
