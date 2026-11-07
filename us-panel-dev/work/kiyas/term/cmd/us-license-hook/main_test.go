package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"kiyas/licensesrv"
)

func TestWebhookRoutesAreTopicBound(t *testing.T) {
	st, err := licensesrv.OpenStore(filepath.Join(t.TempDir(), "store"))
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
