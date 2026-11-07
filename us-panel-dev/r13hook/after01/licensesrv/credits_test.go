package licensesrv

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestParsePlansBothExportSchemas(t *testing.T) {
	for _, file := range []string{"US_PLANS_JSON_TEST.json", "US_PLANS_JSON_CANLI.json"} {
		// Fixtures live in testdata (previously read from ../../../../reports,
		// which only exists when the tree sits at work*/kiyas/term).
		raw, err := os.ReadFile(filepath.Join("testdata", file))
		if err != nil {
			t.Fatal(err)
		}
		plans, err := ParsePlans(raw)
		if err != nil || len(plans) != 8 {
			t.Fatal(file, err, len(plans))
		}
		for id, p := range plans {
			if p.Kind == "sub" && p.Credits != p.Prompts {
				t.Fatal(id, p)
			}
		}
	}
	for _, raw := range []string{`null`, `{}`, `{"x":{"kind":"sub"}}`, `{"42":{"plan":"bad","days":30}}`, `{"42":{"plan":"credit","kind":"use","days":30,"credits":1}}`, `{"42":{"kind":"sub","days":-1,"credits":1}}`} {
		if _, err := ParsePlans([]byte(raw)); err == nil {
			t.Fatal(raw)
		}
	}
}
func TestCreditHTTPAuthorizationAndIdempotentSettlement(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "p")
	key := h.Store.st.Orders["555"][0]
	api := &CreditHandler{Store: h.Store, Now: h.Cfg.Now}
	request := func(path, token, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, r)
		return w
	}
	if w := request("/v1/credits/reserve", "bad", `{}`); w.Code != 401 {
		t.Fatal(w.Code)
	}
	w := request("/v1/credits/reserve", key, `{"amount":-99,"ttl":0}`)
	var out struct {
		ID string `json:"id"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &out) != nil || out.ID == "" {
		t.Fatal(w.Code, w.Body)
	}
	l, _ := h.Store.License(key)
	if l.Reserved != 1 {
		t.Fatal(l.Reserved)
	}
	other := *h.Store.st.Licenses[key]
	other.Key = "other"
	h.Store.st.Licenses[other.Key] = &other
	body := `{"id":"` + out.ID + `","commit":true}`
	if w := request("/v1/credits/settle", "other", body); w.Code != 409 {
		t.Fatal("cross-license", w.Code)
	}
	for i := 0; i < 2; i++ {
		if w := request("/v1/credits/settle", key, body); w.Code != 200 {
			t.Fatal(w.Code)
		}
	}
	l, _ = h.Store.License(key)
	if l.Credits != 499 || l.Reserved != 0 {
		t.Fatal(l)
	}
	if w := request("/v1/credits/settle", key, strings.Replace(body, "true", "false", 1)); w.Code != 409 {
		t.Fatal(w.Code)
	}
}
func TestCreditConcurrentLastCreditExpiryAndPersistenceFailure(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "p")
	key := h.Store.st.Orders["555"][0]
	h.Store.st.Licenses[key].Credits = 1
	var wg sync.WaitGroup
	var got atomic.Int32
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := h.Store.Reserve(key, 1, time.Minute, h.now()); err == nil {
				got.Add(1)
			}
		}()
	}
	wg.Wait()
	if got.Load() != 1 {
		t.Fatal(got.Load())
	}
	var id string
	for k := range h.Store.st.Reservations {
		id = k
	}
	if err := h.Store.settleFor(key, id, true, h.now().Add(time.Minute)); err == nil {
		t.Fatal("expired commit")
	}
	if err := h.Store.settleFor(key, id, false, h.now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(t.TempDir(), "file")
	_ = os.WriteFile(blocker, []byte("x"), 0600)
	old := h.Store.path
	h.Store.path = filepath.Join(blocker, "store")
	if _, err := h.Store.Reserve(key, 1, time.Minute, h.now()); err == nil {
		t.Fatal("save succeeded")
	}
	l, _ := h.Store.License(key)
	if l.Reserved != 0 {
		t.Fatal("reserve rollback", l)
	}
	h.Store.path = old
	id, _ = h.Store.Reserve(key, 1, time.Minute, h.now())
	h.Store.path = filepath.Join(blocker, "store")
	if err := h.Store.settleFor(key, id, true, h.now()); err == nil {
		t.Fatal("settle save succeeded")
	}
	l, _ = h.Store.License(key)
	if l.Credits != 1 || l.Reserved != 1 {
		t.Fatal("settle rollback", l)
	}
}
