package licensesrv

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const payload = `{"id":555,"email":"a@example.com","financial_status":"paid","line_items":[{"id":1,"product_id":42,"quantity":1},{"id":2,"product_id":99,"quantity":1}]}`

func sign(secret, body string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(body))
	return base64.StdEncoding.EncodeToString(m.Sum(nil))
}

func setup(t *testing.T) (*Handler, string) {
	path := filepath.Join(t.TempDir(), "store.json")
	st, err := OpenStore(path)
	t.Cleanup(func() {
		if st != nil {
			st.Close()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC)
	return &Handler{Store: st, Cfg: Config{WebhookSecret: "whsec", LicenseSecret: "lic", Plans: map[int64]Plan{42: {Kind: "sub", Days: 30, Prompts: 500, Credits: 500}}, Now: func() time.Time { return now }}}, path
}

func post(h http.Handler, body, sig, id string) int {
	r := httptest.NewRequest("POST", "/webhooks/shopify/orders-paid", strings.NewReader(body))
	r.Header.Set("X-Shopify-Hmac-Sha256", sig)
	r.Header.Set("X-Shopify-Topic", "orders/paid")
	r.Header.Set("X-Shopify-Webhook-Id", id)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Code
}

func TestRejectsBadHMAC(t *testing.T) {
	h, _ := setup(t)
	if c := post(h, payload, sign("wrong", payload), "w1"); c != 401 {
		t.Fatal(c)
	}
	if c := post(h, payload, "", "w1"); c != 401 {
		t.Fatal(c)
	}
	if len(h.Store.Outbox()) != 0 {
		t.Fatal("mail queued")
	}
}

func TestPaidOrderMintsPersistsQueuesIdempotent(t *testing.T) {
	h, path := setup(t)
	if c := post(h, payload, sign("whsec", payload), "w1"); c != 200 {
		t.Fatal(c)
	}
	if c := post(h, payload, sign("whsec", payload), "w1"); c != 200 {
		t.Fatal(c)
	}
	if c := post(h, payload, sign("whsec", payload), "w2"); c != 200 {
		t.Fatal(c)
	}
	h.Store.Close()
	st, _ := OpenStore(path) // kalıcılık
	t.Cleanup(func() {
		if st != nil {
			st.Close()
		}
	})
	if n := len(st.st.Licenses); n != 1 {
		t.Fatalf("lisans=%d (yalnız eşlenen ürün, tek sefer)", n)
	}
	ob := st.Outbox()
	if len(ob) != 1 || ob[0].To != "a@example.com" || !strings.Contains(ob[0].Body, "us1.") {
		t.Fatalf("%+v", ob)
	}
}

type fakeSender struct {
	fail bool
	got  []Mail
}

func (f *fakeSender) Send(m Mail) error {
	if f.fail {
		return errors.New("x")
	}
	f.got = append(f.got, m)
	return nil
}

func TestOutboxRetryAndSend(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "w1")
	now := time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC)
	f := &fakeSender{fail: true}
	if s, fl := h.Store.FlushOutbox(f, now); s != 0 || fl != 1 {
		t.Fatal(s, fl)
	}
	if s, _ := h.Store.FlushOutbox(f, now); s != 0 {
		t.Fatal("geri çekilme yok")
	}
	f.fail = false
	if s, _ := h.Store.FlushOutbox(f, now.Add(time.Hour)); s != 1 || len(f.got) != 1 {
		t.Fatal(s)
	}
	if s, _ := h.Store.FlushOutbox(f, now.Add(2*time.Hour)); s != 0 {
		t.Fatal("tekrar gönderildi")
	}
}

func TestCreditReservation(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "w1")
	var key string
	for k := range h.Store.st.Licenses {
		key = k
	}
	now := time.Now()
	id, err := h.Store.Reserve(key, 400, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Store.Reserve(key, 200, time.Minute, now); err != ErrNoCredit {
		t.Fatal(err)
	}
	if err := h.Store.Settle(id, true); err != nil {
		t.Fatal(err)
	}
	id2, _ := h.Store.Reserve(key, 100, time.Minute, now)
	h.Store.Settle(id2, false)
	_, _ = h.Store.Reserve(key, 100, time.Second, now)
	if _, err := h.Store.Reserve(key, 100, time.Minute, now.Add(time.Minute)); err != nil {
		t.Fatal("süresi dolan rezervasyon iade edilmedi", err)
	}
	l, _ := h.Store.License(key)
	if l.Credits != 100 {
		t.Fatal(l.Credits)
	}
}
