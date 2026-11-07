package licensesrv

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

// ErrorBody is the standard API error envelope. The legacy "error" string field
// is preserved for Round-7 clients; code/message/request_id are additive.
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

func wantsErrorEnvelope(r *http.Request) bool {
	if r == nil {
		return false
	}
	if strings.EqualFold(r.Header.Get("X-Kanus-Error-Envelope"), "1") {
		return true
	}
	if r.URL != nil && r.URL.Query().Get("error_envelope") == "1" {
		return true
	}
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "application/json") || strings.Contains(accept, "application/problem+json")
}

// WriteAPIError writes a dual-mode error: JSON envelope when requested, else plain text
// matching historical http.Error bodies (message + newline).
func WriteAPIError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	if message == "" {
		message = http.StatusText(status)
	}
	if code == "" {
		code = "request_rejected"
	}
	rid := newRequestID()
	w.Header().Set("X-Request-Id", rid)
	if wantsErrorEnvelope(r) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(ErrorBody{Error: message, Code: code, Message: message, RequestID: rid})
		return
	}
	http.Error(w, message, status)
}

// WriteAccountError always returns JSON and keeps the legacy "error" string.
func WriteAccountError(w http.ResponseWriter, status int, code, message string) {
	if message == "" {
		message = "account request rejected"
	}
	if code == "" {
		code = "account_rejected"
	}
	rid := newRequestID()
	w.Header().Set("X-Request-Id", rid)
	accountJSON(w, status, ErrorBody{Error: message, Code: code, Message: message, RequestID: rid})
}
