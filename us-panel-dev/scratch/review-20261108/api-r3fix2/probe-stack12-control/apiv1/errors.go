package apiv1

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

type ErrorBody struct {
	Error     string `json:"error"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "0000000000000000"
	}
	return hex.EncodeToString(b)
}

func wantsEnvelope(r *http.Request) bool {
	if r == nil {
		return false
	}
	if strings.EqualFold(r.Header.Get("X-Kanus-Error-Envelope"), "1") {
		return true
	}
	if r.URL != nil && r.URL.Query().Get("error_envelope") == "1" {
		return true
	}
	a := r.Header.Get("Accept")
	return strings.Contains(a, "application/json") || strings.Contains(a, "application/problem+json")
}

func WriteAPIError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	if message == "" {
		message = http.StatusText(status)
	}
	if code == "" {
		code = "request_rejected"
	}
	rid := newRequestID()
	w.Header().Set("X-Request-Id", rid)
	if wantsEnvelope(r) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(ErrorBody{Error: message, Code: code, Message: message, RequestID: rid})
		return
	}
	http.Error(w, message, status)
}
