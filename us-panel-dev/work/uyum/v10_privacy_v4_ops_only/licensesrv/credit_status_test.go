package licensesrv

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCreditStatusReadOnlyScopedAndPrivate(t *testing.T) {
	h, path := setup(t)
	post(h, payload, sign("whsec", payload), "p")
	key := h.Store.st.Orders["555"][0]
	id, err := h.Store.Reserve(key, 1, time.Minute, h.now())
	if err != nil {
		t.Fatal(err)
	}
	if err = h.Store.settleAmount(key, id, 1, h.now()); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	r := httptest.NewRequest("GET", "/v1/credits/status", nil)
	r.Header.Set("Authorization", "Bearer "+key)
	w := httptest.NewRecorder()
	(&CreditHandler{Store: h.Store, Now: h.now}).ServeHTTP(w, r)
	after, _ := os.ReadFile(path)
	if w.Code != 200 || string(before) != string(after) || !strings.Contains(w.Body.String(), `"committed":1`) || !strings.Contains(w.Body.String(), `"amount":1`) || !strings.Contains(w.Body.String(), `"kind":"sub"`) || !strings.Contains(w.Body.String(), `"remaining":499`) || !strings.Contains(w.Body.String(), `"expiresAt":`) || strings.Contains(w.Body.String(), key) || strings.Contains(w.Body.String(), "a@example.com") {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestCreditStatusIsGETAndRequiresLicense(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "p")
	key := h.Store.st.Orders["555"][0]
	handler := &CreditHandler{Store: h.Store, Now: h.now}
	for _, tc := range []struct {
		method, token string
		want          int
	}{{"POST", key, 405}, {"GET", "", 401}, {"GET", "us1.invalid.key", 401}, {"GET", key, 200}} {
		r := httptest.NewRequest(tc.method, "/v1/credits/status", nil)
		if tc.token != "" {
			r.Header.Set("Authorization", "Bearer "+tc.token)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatal(tc.method, w.Code)
		}
	}
}
