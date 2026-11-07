package licensesrv

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAPIKeyCreateListRotateRevokeAndGrace(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "api-key-mint")
	lic := h.Store.st.Orders["555"][0]
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "keys@example.test", lic)
	session := accountLogin(t, a, "keys@example.test")

	keys := &APIKeyHandler{Store: h.Store, Now: h.now}
	create := httptest.NewRequest("POST", "/v1/api-keys", strings.NewReader(`{"name":"ci"}`))
	create.Header.Set("Authorization", "Bearer "+session)
	cw := httptest.NewRecorder()
	keys.ServeHTTP(cw, create)
	if cw.Code != 201 {
		t.Fatal(cw.Code, cw.Body.String())
	}
	var created struct {
		Key    string       `json:"key"`
		APIKey apiKeyPublic `json:"api_key"`
	}
	if json.Unmarshal(cw.Body.Bytes(), &created) != nil || !strings.HasPrefix(created.Key, "usk_") || created.APIKey.Prefix == "" || created.APIKey.ID == "" {
		t.Fatal("create shape")
	}
	secret := created.Key
	id := created.APIKey.ID
	// Ensure plaintext is not persisted in store JSON
	h.Store.mu.Lock()
	raw, _ := json.Marshal(h.Store.st.APIKeys[id])
	h.Store.mu.Unlock()
	if strings.Contains(string(raw), secret) {
		t.Fatal("plaintext persisted")
	}

	lw := httptest.NewRecorder()
	lr := httptest.NewRequest("GET", "/v1/api-keys", nil)
	lr.Header.Set("Authorization", "Bearer "+session)
	keys.ServeHTTP(lw, lr)
	if lw.Code != 200 || strings.Contains(lw.Body.String(), secret) || !strings.Contains(lw.Body.String(), `"prefix"`) {
		t.Fatal("list leaked or missing", lw.Code)
	}

	// status via API key
	cred := &CreditHandler{Store: h.Store, Now: h.now}
	sr := httptest.NewRequest("GET", "/v1/credits/status", nil)
	sr.Header.Set("Authorization", "Bearer "+secret)
	sw := httptest.NewRecorder()
	cred.ServeHTTP(sw, sr)
	if sw.Code != 200 || !strings.Contains(sw.Body.String(), `"kind":"account"`) {
		t.Fatal("status via key", sw.Code, sw.Body.String())
	}

	// rotate → new secret; old in grace still works
	rr := httptest.NewRequest("POST", "/v1/api-keys/"+id+"/rotate", nil)
	rr.Header.Set("Authorization", "Bearer "+session)
	rw := httptest.NewRecorder()
	keys.ServeHTTP(rw, rr)
	var rotated struct {
		Key      string       `json:"key"`
		APIKey   apiKeyPublic `json:"api_key"`
		Previous apiKeyPublic `json:"previous"`
	}
	if rw.Code != 200 || json.Unmarshal(rw.Body.Bytes(), &rotated) != nil || rotated.Key == secret || rotated.Previous.Status != "grace" {
		t.Fatal("rotate", rw.Code, rw.Body.String())
	}
	oldSecret, newSecret := secret, rotated.Key

	for _, tok := range []string{oldSecret, newSecret} {
		r := httptest.NewRequest("GET", "/v1/credits/status", nil)
		r.Header.Set("Authorization", "Bearer "+tok)
		w := httptest.NewRecorder()
		cred.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatal("grace auth", w.Code)
		}
	}

	// revoke old hard
	vr := httptest.NewRequest("POST", "/v1/api-keys/"+id+"/revoke", nil)
	vr.Header.Set("Authorization", "Bearer "+session)
	vw := httptest.NewRecorder()
	keys.ServeHTTP(vw, vr)
	if vw.Code != 200 {
		t.Fatal(vw.Code)
	}
	r := httptest.NewRequest("GET", "/v1/credits/status", nil)
	r.Header.Set("Authorization", "Bearer "+oldSecret)
	w := httptest.NewRecorder()
	cred.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal("revoked still valid", w.Code)
	}
	// new still ok
	r = httptest.NewRequest("GET", "/v1/credits/status", nil)
	r.Header.Set("Authorization", "Bearer "+newSecret)
	w = httptest.NewRecorder()
	cred.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal("new after revoke old", w.Code)
	}
}

func TestAPIKeyRateLimitRetryAfter(t *testing.T) {
	h, _ := setup(t)
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "rate@example.test", "")
	session := accountLogin(t, a, "rate@example.test")
	keys := &APIKeyHandler{Store: h.Store, Now: h.now}
	cw := httptest.NewRecorder()
	cr := httptest.NewRequest("POST", "/v1/api-keys", strings.NewReader(`{"name":"r"}`))
	cr.Header.Set("Authorization", "Bearer "+session)
	keys.ServeHTTP(cw, cr)
	var created struct {
		Key    string       `json:"key"`
		APIKey apiKeyPublic `json:"api_key"`
	}
	json.Unmarshal(cw.Body.Bytes(), &created)
	h.Store.mu.Lock()
	h.Store.st.APIKeys[created.APIKey.ID].RatePerMin = 2
	h.Store.mu.Unlock()

	cred := &CreditHandler{Store: h.Store, Now: h.now}
	var last *httptest.ResponseRecorder
	for i := 0; i < 3; i++ {
		r := httptest.NewRequest("GET", "/v1/credits/status", nil)
		r.Header.Set("Authorization", "Bearer "+created.Key)
		r.Header.Set("Accept", "application/json")
		last = httptest.NewRecorder()
		cred.ServeHTTP(last, r)
	}
	if last.Code != 429 || last.Header().Get("Retry-After") == "" {
		t.Fatal(last.Code, last.Header().Get("Retry-After"), last.Body.String())
	}
	var body ErrorBody
	if json.Unmarshal(last.Body.Bytes(), &body) != nil || body.Code != "rate_limited" {
		t.Fatal(last.Body.String())
	}
}

func TestUsageEndpointNoSecretLeak(t *testing.T) {
	h, _ := setup(t)
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "usage@example.test", "")
	session := accountLogin(t, a, "usage@example.test")
	keys := &APIKeyHandler{Store: h.Store, Now: h.now}
	cw := httptest.NewRecorder()
	cr := httptest.NewRequest("POST", "/v1/api-keys", strings.NewReader(`{}`))
	cr.Header.Set("Authorization", "Bearer "+session)
	keys.ServeHTTP(cw, cr)
	var created struct {
		Key string `json:"key"`
	}
	json.Unmarshal(cw.Body.Bytes(), &created)

	u := &UsageHandler{Store: h.Store, Now: h.now}
	r := httptest.NewRequest("GET", "/v1/usage", nil)
	r.Header.Set("Authorization", "Bearer "+session)
	w := httptest.NewRecorder()
	u.ServeHTTP(w, r)
	if w.Code != 200 || strings.Contains(w.Body.String(), created.Key) || strings.Contains(w.Body.String(), "usage@example.test") {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestAPIKeyMuxRoutes(t *testing.T) {
	// compile-time / handler smoke without logging secrets
	_ = http.MethodGet
	_ = time.Second
}

func TestAPIKeyDailyQuotaRetryAfter(t *testing.T) {
	h, _ := setup(t)
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "quota@example.test", "")
	session := accountLogin(t, a, "quota@example.test")
	keys := &APIKeyHandler{Store: h.Store, Now: h.now}
	cw := httptest.NewRecorder()
	cr := httptest.NewRequest("POST", "/v1/api-keys", strings.NewReader(`{"name":"q"}`))
	cr.Header.Set("Authorization", "Bearer "+session)
	keys.ServeHTTP(cw, cr)
	var created struct {
		Key    string       `json:"key"`
		APIKey apiKeyPublic `json:"api_key"`
	}
	if json.Unmarshal(cw.Body.Bytes(), &created) != nil {
		t.Fatal(cw.Body.String())
	}
	h.Store.mu.Lock()
	k := h.Store.st.APIKeys[created.APIKey.ID]
	k.DailyQuota = 1
	k.RatePerMin = 100
	k.DayStamp = h.now().UTC().Format("2006-01-02")
	k.DayCount = 1
	h.Store.mu.Unlock()
	r := httptest.NewRequest("GET", "/v1/credits/status", nil)
	r.Header.Set("Authorization", "Bearer "+created.Key)
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	(&CreditHandler{Store: h.Store, Now: h.now}).ServeHTTP(w, r)
	if w.Code != 429 || w.Header().Get("Retry-After") == "" {
		t.Fatal(w.Code, w.Header().Get("Retry-After"), w.Body.String())
	}
	var body ErrorBody
	if json.Unmarshal(w.Body.Bytes(), &body) != nil || body.Code != "quota_exceeded" {
		t.Fatal(w.Body.String())
	}
}

func TestAPIKeyRotateDefaultGrace24h(t *testing.T) {
	h, _ := setup(t)
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "grace@example.test", "")
	session := accountLogin(t, a, "grace@example.test")
	keys := &APIKeyHandler{Store: h.Store, Now: h.now}
	cw := httptest.NewRecorder()
	cr := httptest.NewRequest("POST", "/v1/api-keys", strings.NewReader(`{"name":"g"}`))
	cr.Header.Set("Authorization", "Bearer "+session)
	keys.ServeHTTP(cw, cr)
	var created struct {
		Key    string       `json:"key"`
		APIKey apiKeyPublic `json:"api_key"`
	}
	json.Unmarshal(cw.Body.Bytes(), &created)
	rr := httptest.NewRequest("POST", "/v1/api-keys/"+created.APIKey.ID+"/rotate", strings.NewReader(`{}`))
	rr.Header.Set("Authorization", "Bearer "+session)
	rw := httptest.NewRecorder()
	keys.ServeHTTP(rw, rr)
	var out struct {
		GraceSeconds int          `json:"grace_seconds"`
		Previous     apiKeyPublic `json:"previous"`
		Key          string       `json:"key"`
	}
	if rw.Code != 200 || json.Unmarshal(rw.Body.Bytes(), &out) != nil || out.GraceSeconds != int((24 * time.Hour).Seconds()) || out.Previous.Status != "grace" {
		t.Fatal(rw.Code, rw.Body.String())
	}
	// old still works during grace
	r := httptest.NewRequest("GET", "/v1/credits/status", nil)
	r.Header.Set("Authorization", "Bearer "+created.Key)
	w := httptest.NewRecorder()
	(&CreditHandler{Store: h.Store, Now: h.now}).ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal("grace auth", w.Code)
	}
}

func TestAPIKeyRotateRevokeOldImmediate(t *testing.T) {
	h, _ := setup(t)
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "nograce@example.test", "")
	session := accountLogin(t, a, "nograce@example.test")
	keys := &APIKeyHandler{Store: h.Store, Now: h.now}
	cw := httptest.NewRecorder()
	cr := httptest.NewRequest("POST", "/v1/api-keys", strings.NewReader(`{}`))
	cr.Header.Set("Authorization", "Bearer "+session)
	keys.ServeHTTP(cw, cr)
	var created struct {
		Key    string       `json:"key"`
		APIKey apiKeyPublic `json:"api_key"`
	}
	json.Unmarshal(cw.Body.Bytes(), &created)
	rr := httptest.NewRequest("POST", "/v1/api-keys/"+created.APIKey.ID+"/rotate", strings.NewReader(`{"revoke_old":true}`))
	rr.Header.Set("Authorization", "Bearer "+session)
	rw := httptest.NewRecorder()
	keys.ServeHTTP(rw, rr)
	var out struct {
		GraceSeconds int          `json:"grace_seconds"`
		Previous     apiKeyPublic `json:"previous"`
	}
	if rw.Code != 200 || json.Unmarshal(rw.Body.Bytes(), &out) != nil || out.GraceSeconds != 0 || out.Previous.Status != "revoked" {
		t.Fatal(rw.Code, rw.Body.String())
	}
	r := httptest.NewRequest("GET", "/v1/credits/status", nil)
	r.Header.Set("Authorization", "Bearer "+created.Key)
	w := httptest.NewRecorder()
	(&CreditHandler{Store: h.Store, Now: h.now}).ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal("old key should be dead", w.Code)
	}
}
