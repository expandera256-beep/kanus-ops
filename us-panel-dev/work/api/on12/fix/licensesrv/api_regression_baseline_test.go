package licensesrv

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// API-0001: §7.4–7.6 öncesi kredi defteri / accounts regresyon kilidi.
// Üretim kodu değiştirilmeden mevcut davranışı sabitler.

func TestAPIRegressionLedgerBalanceReserveCommitRelease(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "api-reg-mint")
	key := h.Store.st.Orders["555"][0]
	now := h.now()

	st, ok := h.Store.CreditStatus(key, now)
	if !ok || st.Remaining != 500 || st.Available != 500 || st.Reserved != 0 || st.Kind != "sub" {
		t.Fatalf("initial status %+v", st)
	}

	id, err := h.Store.Reserve(key, 1, time.Minute, now)
	if err != nil || id == "" {
		t.Fatal(err)
	}
	st, _ = h.Store.CreditStatus(key, now)
	if st.Remaining != 500 || st.Reserved != 1 || st.Available != 499 {
		t.Fatalf("after reserve %+v", st)
	}

	if err := h.Store.settleFor(key, id, false, now); err != nil {
		t.Fatal(err)
	}
	st, _ = h.Store.CreditStatus(key, now)
	if st.Remaining != 500 || st.Reserved != 0 || st.Available != 500 || st.Released != 1 {
		t.Fatalf("after release %+v", st)
	}

	id, err = h.Store.Reserve(key, 1, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Store.settleFor(key, id, true, now); err != nil {
		t.Fatal(err)
	}
	st, _ = h.Store.CreditStatus(key, now)
	if st.Remaining != 499 || st.Reserved != 0 || st.Available != 499 || st.Committed != 1 {
		t.Fatalf("after commit %+v", st)
	}
}

func TestAPIRegressionInsufficientCreditAndIdempotentSettle(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "api-reg-nocredit")
	key := h.Store.st.Orders["555"][0]
	now := h.now()
	h.Store.st.Licenses[key].Credits = 1

	if _, err := h.Store.Reserve(key, 2, time.Minute, now); err != ErrNoCredit {
		t.Fatal("want ErrNoCredit for amount>available", err)
	}
	id, err := h.Store.Reserve(key, 1, time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Store.Reserve(key, 1, time.Minute, now); err != ErrNoCredit {
		t.Fatal("want ErrNoCredit when fully reserved", err)
	}
	if err := h.Store.settleFor(key, id, true, now); err != nil {
		t.Fatal(err)
	}
	// Identical settle is idempotent (no double spend).
	if err := h.Store.settleFor(key, id, true, now); err != nil {
		t.Fatal("idempotent commit", err)
	}
	st, _ := h.Store.CreditStatus(key, now)
	if st.Remaining != 0 || st.Committed != 1 || st.Reserved != 0 {
		t.Fatalf("double spend %+v", st)
	}
	// Conflicting settle (release after commit) rejected.
	if err := h.Store.settleFor(key, id, false, now); err == nil {
		t.Fatal("conflicting settle accepted")
	}
}

func TestAPIRegressionConcurrentLastCreditNoDoubleSpend(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "api-reg-race")
	key := h.Store.st.Orders["555"][0]
	now := h.now()
	h.Store.st.Licenses[key].Credits = 1

	var wg sync.WaitGroup
	var ok atomic.Int32
	ids := make(chan string, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := h.Store.Reserve(key, 1, time.Minute, now)
			if err == nil {
				ok.Add(1)
				ids <- id
			} else if err != ErrNoCredit {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	close(ids)
	if ok.Load() != 1 {
		t.Fatalf("expected exactly 1 reserve, got %d", ok.Load())
	}
	var id string
	for v := range ids {
		id = v
	}
	if err := h.Store.settleFor(key, id, true, now); err != nil {
		t.Fatal(err)
	}
	l, _ := h.Store.License(key)
	if l.Credits != 0 || l.Reserved != 0 {
		t.Fatalf("ledger %+v", l)
	}
}

func TestAPIRegressionAccountBalanceAggregatesOwnedLicenses(t *testing.T) {
	h, _ := setup(t)
	body := strings.Replace(payload, `"product_id":42,"quantity":1`, `"product_id":42,"quantity":2`, 1)
	post(h, body, sign("whsec", body), "api-reg-acct")
	keys := h.Store.st.Orders["555"]
	if len(keys) != 2 {
		t.Fatalf("want 2 keys, got %d", len(keys))
	}
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "api-reg@example.test", keys[0])
	token := accountLogin(t, a, "api-reg@example.test")
	if accountPost(a, "/account/redeem", token, map[string]string{"license": keys[1]}).Code != 200 {
		t.Fatal("redeem second")
	}
	b, err := h.Store.AccountBalance(token, time.Now())
	if err != nil || b.Remaining != 1000 || b.Reserved != 0 || b.Kind != "account" {
		t.Fatalf("aggregate balance %+v err=%v", b, err)
	}
	w := accountPost(a, "/account/balance", token, map[string]string{})
	var got AccountBalance
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &got) != nil || got.Remaining != 1000 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestAPIRegressionCreditHTTPStatusContract(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "api-reg-http")
	key := h.Store.st.Orders["555"][0]
	api := &CreditHandler{Store: h.Store, Now: h.now}
	call := func(method, path, token, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		api.ServeHTTP(w, r)
		return w
	}
	if w := call("POST", "/v1/credits/reserve", "", `{}`); w.Code != 401 {
		t.Fatal("unauth reserve", w.Code)
	}
	h.Store.st.Licenses[key].Credits = 0
	if w := call("POST", "/v1/credits/reserve", key, `{}`); w.Code != 402 {
		t.Fatal("no credit", w.Code)
	}
	h.Store.st.Licenses[key].Credits = 5
	w := call("POST", "/v1/credits/reserve", key, `{}`)
	var out struct {
		ID string `json:"id"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &out) != nil || out.ID == "" {
		t.Fatal(w.Code, w.Body.String())
	}
	body := `{"id":"` + out.ID + `","commit":true}`
	if w := call("POST", "/v1/credits/settle", key, body); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := call("POST", "/v1/credits/settle", key, body); w.Code != 200 {
		t.Fatal("idempotent settle http", w.Code)
	}
	if w := call("GET", "/v1/credits/status", key, ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"committed":1`) {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestAPIRegressionAccountSessionCreditStatus(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "api-reg-sess-status")
	key := h.Store.st.Orders["555"][0]
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "status@example.test", key)
	token := accountLogin(t, a, "status@example.test")
	api := &CreditHandler{Store: h.Store, Now: h.now}
	r := httptest.NewRequest("GET", "/v1/credits/status", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"kind":"account"`) || !strings.Contains(w.Body.String(), `"remaining":500`) {
		t.Fatal(w.Code, w.Body.String())
	}
}
