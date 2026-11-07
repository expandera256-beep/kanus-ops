package licensesrv

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func postTopic(h http.Handler, topic, body, id string) int {
	r := httptest.NewRequest("POST", "/webhooks/shopify/test", strings.NewReader(body))
	r.Header.Set("X-Shopify-Hmac-Sha256", sign("whsec", body))
	r.Header.Set("X-Shopify-Topic", topic)
	r.Header.Set("X-Shopify-Webhook-Id", id)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Code
}
func activeCount(st *Store) int {
	n := 0
	for _, l := range st.st.Licenses {
		if !l.Revoked {
			n++
		}
	}
	return n
}
func TestCancelRevokesPendingReservationsAndPersists(t *testing.T) {
	h, path := setup(t)
	post(h, payload, sign("whsec", payload), "paid")
	key := h.Store.st.Orders["555"][0]
	id, err := h.Store.Reserve(key, 1, time.Minute, h.now())
	if err != nil {
		t.Fatal(err)
	}
	for _, delivery := range []string{"cancel", "cancel", "cancel-new"} {
		if c := postTopic(h, "orders/cancelled", `{"id":555}`, delivery); c != 200 {
			t.Fatal(c)
		}
	}
	st, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	l, _ := st.License(key)
	if !l.Revoked || l.Credits != 0 || l.Reserved != 0 || len(st.st.Reservations) != 0 || !st.Outbox()[0].Dead {
		t.Fatalf("not revoked: %+v", l)
	}
	if err := st.settleFor(key, id, true, h.now()); err == nil {
		t.Fatal("revoked commit permitted")
	}
	if _, err := st.Reserve(key, 1, time.Minute, h.now()); err != ErrInvalid {
		t.Fatal(err)
	}
}
func TestCancelBeforePaidTombstone(t *testing.T) {
	h, path := setup(t)
	postTopic(h, "orders/cancelled", `{"id":555}`, "c")
	st, _ := OpenStore(path)
	h.Store = st
	if c := post(h, payload, sign("whsec", payload), "p"); c != 200 {
		t.Fatal(c)
	}
	if len(st.st.Licenses) != 0 || len(st.Outbox()) != 0 {
		t.Fatal("late paid reactivated order")
	}
}
func TestPartialRefundIdempotentAndOutOfOrder(t *testing.T) {
	for _, before := range []bool{true, false} {
		t.Run(fmt.Sprint(before), func(t *testing.T) {
			h, path := setup(t)
			paid := strings.Replace(payload, `"product_id":42,"quantity":1`, `"product_id":42,"quantity":3`, 1)
			refund := `{"id":11,"order_id":555,"refund_line_items":[{"line_item_id":1,"quantity":1}]}`
			if before {
				postTopic(h, "refunds/create", refund, "r")
			}
			post(h, paid, sign("whsec", paid), "p")
			for _, delivery := range []string{"r", "r-new", "r-new"} {
				if c := postTopic(h, "refunds/create", refund, delivery); c != 200 {
					t.Fatal(c)
				}
			}
			st, err := OpenStore(path)
			if err != nil {
				t.Fatal(err)
			}
			h.Store = st
			if activeCount(st) != 2 || len(st.st.Refunds) != 1 {
				t.Fatal(activeCount(st), len(st.st.Refunds))
			}
			for _, key := range st.st.Orders["555"] {
				l, _ := st.License(key)
				if l.Revoked && strings.Contains(st.Outbox()[0].Body, key) {
					t.Fatal("revoked key still queued")
				}
			}
			full := `{"id":12,"order_id":555,"refund_line_items":[{"line_item_id":1,"quantity":2}]}`
			if c := postTopic(h, "refunds/create", full, "r2"); c != 200 || activeCount(st) != 0 || !st.Outbox()[0].Dead {
				t.Fatal("full refund", c)
			}
		})
	}
}
func TestReversalHMACShopTopicAndInvalidPayload(t *testing.T) {
	for _, topic := range []string{"orders/cancelled", "refunds/create"} {
		t.Run(topic, func(t *testing.T) {
			h, _ := setup(t)
			h.Cfg.ShopDomain = "kanus-test.myshopify.com"
			r := httptest.NewRequest("POST", "/", strings.NewReader(`{"id":555}`))
			r.Header.Set("X-Shopify-Topic", topic)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 401 {
				t.Fatal(w.Code)
			}
			r.Header.Set("X-Shopify-Hmac-Sha256", sign("whsec", `{"id":555}`))
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 401 {
				t.Fatal("wrong shop", w.Code)
			}
		})
	}
	h, _ := setup(t)
	h.Topic = "orders/cancelled"
	if c := postTopic(h, "orders/paid", payload, "x"); c != 400 {
		t.Fatal(c)
	}
	for _, body := range []string{`{}`, `{"id":1,"order_id":555,"refund_line_items":[{"line_item_id":1,"quantity":-1}]}`} {
		h.Topic = "refunds/create"
		if c := postTopic(h, "refunds/create", body, "x"); c != 400 {
			t.Fatal(c)
		}
	}
}
func TestFinancialOnlyRefundDoesNotInventCreditConversion(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "p")
	if c := postTopic(h, "refunds/create", `{"id":21,"order_id":555,"transactions":[{"amount":"10.00"}]}`, "r"); c != 200 || activeCount(h.Store) != 1 {
		t.Fatal(c)
	}
}
func TestFailedWriteRollsBackAndRetrySucceeds(t *testing.T) {
	for _, topic := range []string{"orders/paid", "orders/cancelled", "refunds/create"} {
		t.Run(topic, func(t *testing.T) {
			h, path := setup(t)
			if topic != "orders/paid" {
				post(h, payload, sign("whsec", payload), "p")
			}
			// A file as the parent deterministically fails even when tests run as root.
			blocker := filepath.Join(t.TempDir(), "file")
			if err := os.WriteFile(blocker, []byte("x"), 0600); err != nil {
				t.Fatal(err)
			}
			h.Store.path = filepath.Join(blocker, "store.json")
			body := payload
			if topic == "orders/cancelled" {
				body = `{"id":555}`
			}
			if topic == "refunds/create" {
				body = `{"id":1,"order_id":555,"refund_line_items":[{"line_item_id":1,"quantity":1}]}`
			}
			if c := postTopic(h, topic, body, "retry"); c != 500 {
				t.Fatal(c)
			}
			if _, seen := h.Store.st.Webhooks["retry"]; seen {
				t.Fatal("failed write marked delivered")
			}
			h.Store.path = path
			if c := postTopic(h, topic, body, "retry"); c != 200 {
				t.Fatal(c)
			}
			if topic == "orders/paid" {
				if activeCount(h.Store) != 1 {
					t.Fatal("mint lost")
				}
			} else if activeCount(h.Store) != 0 {
				t.Fatal("revocation lost")
			}
		})
	}
}
