package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"kiyas/licensesrv"
)

func TestWebhookRoutesAreTopicBound(t *testing.T) {
	st, err := licensesrv.OpenStore(filepath.Join(t.TempDir(), "store"))
	t.Cleanup(func() {
		if st != nil {
			st.Close()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	mux := newMux(licensesrv.Config{WebhookSecret: "test"}, st)
	for path, topic := range map[string]string{"/webhooks/shopify/orders-paid": "orders/paid", "/webhooks/shopify/orders-cancelled": "orders/cancelled", "/webhooks/shopify/refunds-create": "refunds/create"} {
		body := `{"id":1,"order_id":2,"email":"a@example.com"}`
		mac := hmac.New(sha256.New, []byte("test"))
		mac.Write([]byte(body))
		for _, signed := range []bool{false, true} {
			r := httptest.NewRequest("POST", path, strings.NewReader(body))
			r.Header.Set("X-Shopify-Topic", topic)
			if signed {
				r.Header.Set("X-Shopify-Hmac-Sha256", base64.StdEncoding.EncodeToString(mac.Sum(nil)))
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			want := 401
			if signed {
				want = 200
			}
			if w.Code != want {
				t.Fatal(path, w.Code)
			}
		}
	}
	for _, path := range []string{"/v1/credits/reserve", "/v1/credits/settle"} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", path, nil))
		if w.Code != 401 {
			t.Fatal(path, w.Code)
		}
	}
}

type countingSender struct{ calls int }

func (s *countingSender) Send(licensesrv.Mail) error { s.calls++; return nil }
func TestSMTPDisabledDoesNotAttemptOrExhaustOutbox(t *testing.T) {
	st, err := licensesrv.OpenStore(filepath.Join(t.TempDir(), "store"))
	t.Cleanup(func() {
		if st != nil {
			st.Close()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg := licensesrv.Config{WebhookSecret: "test", LicenseSecret: "test-license", Plans: map[int64]licensesrv.Plan{42: {Kind: "sub", Days: 30, Credits: 1, Prompts: 1}}}
	body := `{"id":7,"email":"a@example.com","line_items":[{"id":1,"product_id":42,"quantity":1}]}`
	mac := hmac.New(sha256.New, []byte("test"))
	mac.Write([]byte(body))
	r := httptest.NewRequest("POST", "/webhooks/shopify/orders-paid", strings.NewReader(body))
	r.Header.Set("X-Shopify-Topic", "orders/paid")
	r.Header.Set("X-Shopify-Hmac-Sha256", base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	w := httptest.NewRecorder()
	newMux(cfg, st).ServeHTTP(w, r)
	if w.Code != 200 || len(st.Outbox()) != 1 {
		t.Fatal(w.Code)
	}
	sender := &countingSender{}
	for i := 0; i < 20; i++ {
		flushIfMailEnabled(st, sender, false, time.Now().Add(time.Duration(i)*time.Hour))
	}
	mail := st.Outbox()[0]
	if sender.calls != 0 || mail.Attempts != 0 || mail.Dead || !mail.SentAt.IsZero() {
		t.Fatal("disabled mail mutated", mail)
	}
}

func TestPlansFilePrecedenceAndFailClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plans.json")
	t.Setenv("US_PLANS_FILE", path)
	t.Setenv("US_PLANS_JSON", `{"99":{"kind":"sub","days":1,"credits":9}}`)
	os.WriteFile(path, []byte(`{"42":{"kind":"sub","days":1,"credits":2}}`), 0600)
	plans, err := loadPlans()
	if err != nil || plans[42].Credits != 2 || len(plans) != 1 {
		t.Fatal(plans, err)
	}
	os.WriteFile(path, []byte("{"), 0600)
	if _, err = loadPlans(); err == nil {
		t.Fatal("invalid file fell back to env")
	}
	t.Setenv("US_PLANS_FILE", filepath.Join(t.TempDir(), "missing"))
	if _, err = loadPlans(); err == nil {
		t.Fatal("missing file fell back")
	}
	t.Setenv("US_PLANS_FILE", "")
	plans, err = loadPlans()
	if err != nil || plans[99].Credits != 9 {
		t.Fatal(plans, err)
	}
}
func TestModelRoutesAreLicenseProtected(t *testing.T) {
	st, err := licensesrv.OpenStore(filepath.Join(t.TempDir(), "store"))
	t.Cleanup(func() {
		if st != nil {
			st.Close()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	mux := newMux(licensesrv.Config{}, st)
	for _, path := range []string{"/v1/chat", "/v1/chat/completions"} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(`{}`)))
		if w.Code != 401 {
			t.Fatal(path, w.Code)
		}
	}
	for _, path := range []string{"/v1/models", "/v1/credits/status"} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 401 {
			t.Fatal(path, w.Code)
		}
	}
}

func TestPrivacyRetentionDaysMounted(t *testing.T) {
	t.Setenv("US_PRIVACY_RETENTION_DAYS", "1")
	t.Setenv("US_PRIVACY_ADMIN_TOKEN", "adm-test-token")
	dir := t.TempDir()
	path := filepath.Join(dir, "s.json")
	st, err := licensesrv.OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	st.Close()
	// Seed an aged webhook via on-disk ledger (no internal field access from package main).
	raw := []byte(`{"accounts":{},"licenseOwners":{},"webhooks":{"old":"2020-01-01T00:00:00Z"},"orders":{},"licenses":{},"outbox":[],"reservations":{},"settlements":{},"cancelled":{},"refunds":{}}`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	st, err = licensesrv.OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	mux := newMux(licensesrv.Config{}, st)
	req := httptest.NewRequest("POST", "/account/privacy/retention-run", nil)
	req.Header.Set("Authorization", "Bearer adm-test-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("retention %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"webhooksRemoved":1`) && !strings.Contains(w.Body.String(), `"webhooksRemoved": 1`) {
		t.Fatalf("expected KeepDays=1 to remove aged webhook: %s", w.Body.String())
	}
}
