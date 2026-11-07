package licensesrv

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIErrorEnvelopeJSONAdditiveAndPlainFallback(t *testing.T) {
	r := httptest.NewRequest("POST", "/v1/credits/reserve", nil)
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	WriteAPIError(w, r, 402, "insufficient_credit", "reserve")
	if w.Code != 402 || w.Header().Get("X-Request-Id") == "" {
		t.Fatal(w.Code, w.Header())
	}
	var body ErrorBody
	if json.Unmarshal(w.Body.Bytes(), &body) != nil || body.Error != "reserve" || body.Code != "insufficient_credit" || body.Message != "reserve" || body.RequestID == "" {
		t.Fatal(w.Body.String())
	}
	// Plain fallback preserves http.Error shape (message + newline).
	r2 := httptest.NewRequest("POST", "/v1/credits/reserve", nil)
	w2 := httptest.NewRecorder()
	WriteAPIError(w2, r2, 401, "unauthorized", "license")
	if w2.Code != 401 || w2.Body.String() != "license\n" {
		t.Fatal(w2.Code, w2.Body.String())
	}
}

func TestAccountErrorKeepsLegacyErrorField(t *testing.T) {
	w := httptest.NewRecorder()
	WriteAccountError(w, 401, "unauthorized", "account request rejected")
	var body ErrorBody
	if w.Code != 401 || json.Unmarshal(w.Body.Bytes(), &body) != nil {
		t.Fatal(w.Code, w.Body.String())
	}
	if body.Error != "account request rejected" || body.Code != "unauthorized" || body.RequestID == "" {
		t.Fatal(body)
	}
	if !strings.Contains(w.Body.String(), `"error":"account request rejected"`) {
		t.Fatal("legacy field missing")
	}
}
