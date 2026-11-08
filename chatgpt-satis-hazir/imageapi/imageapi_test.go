package imageapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testKey = "test-license-token-01"
const testIdem = "7f3c2a1e-img-20261108-0001"

// fixturePNG is a 1x1 PNG used only as a generator stand-in. It is not FLUX output.
var fixturePNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
	0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
	0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}

func fixtureGen(model, prompt, size string, seed uint32) (string, []byte, error) {
	return "image/png", fixturePNG, nil
}

func subDir() MapDirectory {
	return MapDirectory{testKey: {ID: "lic-1", Kind: "sub", Active: true}}
}

func doPOST(t *testing.T, h http.Handler, body string, hdr map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/image", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func authIdem() map[string]string {
	return map[string]string{"Authorization": "Bearer " + testKey, "Idempotency-Key": testIdem}
}

func TestMissingIdempotency(t *testing.T) {
	s := New(Options{Directory: subDir()})
	rec := doPOST(t, s, `{"prompt":"merhaba"}`, map[string]string{"Authorization": "Bearer " + testKey})
	if rec.Code != 400 {
		t.Fatalf("status %d body %s", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), "missing_idempotency_key") {
		t.Fatalf("body %s", rec.Body)
	}
}

func TestInvalidModelAndSizeAndN(t *testing.T) {
	s := New(Options{Directory: subDir()})
	cases := []struct {
		body, code string
	}{
		{`{"prompt":"x","model":"FLUX.2-klein-4B"}`, "invalid_model"},
		{`{"prompt":"x","size":"1024×1024"}`, "invalid_size"},
		{`{"prompt":"x","size":"1024X1024"}`, "invalid_size"},
		{`{"prompt":"x","n":0}`, "invalid_n"},
		{`{"prompt":"x","n":5}`, "invalid_n"},
		{`{"prompt":"x","n":1.5}`, "invalid_n"},
		{`{"prompt":"x","extra":1}`, "unknown_field"},
		{`{"prompt":"x","fallback":true}`, "unsupported_parameter"},
		{`{"prompt":"x","allow_fallback":true}`, "unsupported_parameter"},
		{`{"prompt":"x","amount":1}`, "unsupported_parameter"},
		{`{"prompt":"x","ttl":1}`, "unsupported_parameter"},
		{`{"prompt":"x","idempotency_key":"abc"}`, "unsupported_parameter"},
		{`{"prompt":"x","steps":4}`, "unsupported_parameter"},
	}
	for _, tc := range cases {
		rec := doPOST(t, s, tc.body, authIdem())
		if rec.Code != 400 || !strings.Contains(rec.Body.String(), tc.code) {
			t.Fatalf("%s => %d %s want %s", tc.body, rec.Code, rec.Body, tc.code)
		}
	}
}

func TestAuthAndUseKind(t *testing.T) {
	s := New(Options{Directory: MapDirectory{
		testKey: {ID: "lic-1", Kind: "use", Active: true},
	}})
	rec := doPOST(t, s, `{"prompt":"yağmur"}`, authIdem())
	if rec.Code != 403 || !strings.Contains(rec.Body.String(), "plan_kind_not_supported") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	rec = doPOST(t, New(Options{Directory: subDir()}), `{"prompt":"yağmur"}`, map[string]string{"Idempotency-Key": testIdem})
	if rec.Code != 401 || !strings.Contains(rec.Body.String(), `"license"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestPendingRateTableDoesNotCharge(t *testing.T) {
	s := New(Options{Directory: subDir(), Snapshot: EngineSnapshot{
		State: "IMAGE_READY", SlotReady: true, ModelStaged: map[string]bool{"klein": true}, MaxUptimeOK: true,
	}, Generator: fixtureGen})
	rec := doPOST(t, s, `{"prompt":"yağmur"}`, authIdem())
	if rec.Code != 503 || !strings.Contains(rec.Body.String(), "reservation_failed") || !strings.Contains(rec.Body.String(), "rate_table_pending") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestModelNotStagedWhenRatesPresent(t *testing.T) {
	s := New(Options{
		Directory: subDir(),
		Rates:     map[RateKey]int{{"klein", "1024x1024"}: 3}, // test fixture only; not a product tariff
		Snapshot:  EngineSnapshot{State: "COLD", ModelStaged: map[string]bool{"klein": false, "schnell": false}},
	})
	rec := doPOST(t, s, `{"prompt":"yağmur"}`, authIdem())
	if rec.Code != 503 || !strings.Contains(rec.Body.String(), "model_not_staged") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestWarmSync200AndDownload(t *testing.T) {
	s := New(Options{
		Directory: subDir(),
		Rates:     map[RateKey]int{{"klein", "1344x768"}: 3}, // test fixture only
		Snapshot: EngineSnapshot{
			State: "IMAGE_READY", SlotReady: true, MaxUptimeOK: true,
			ModelStaged: map[string]bool{"klein": true},
		},
		Generator: fixtureGen,
	})
	body := `{"prompt":"İstanbul'da yağmur","size":"1344x768","n":2,"seed":10}`
	rec := doPOST(t, s, body, authIdem())
	if rec.Code != 200 {
		t.Fatalf("status %d %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-US-Metered") != "1" {
		t.Fatalf("headers %v", rec.Header())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "succeeded" || resp["object"] != "image.generation" {
		t.Fatalf("%v", resp)
	}
	if strings.Contains(rec.Body.String(), "rsv_") || strings.Contains(rec.Body.String(), "Reservation") {
		t.Fatalf("reservation id leaked: %s", rec.Body)
	}
	id := resp["id"].(string)
	if hidden := s.reservationIDExposed(id); hidden == "" || strings.Contains(rec.Body.String(), hidden) {
		t.Fatalf("reservation handling %q body %s", hidden, rec.Body)
	}
	credits := resp["credits"].(map[string]any)
	if credits["reserved"] != float64(6) || credits["committed"] != float64(6) || credits["released"] != float64(0) {
		t.Fatalf("credits %v", credits)
	}
	ret := resp["retention"].(map[string]any)
	if ret["policy"] != "authorized_24h" {
		t.Fatalf("retention %v", ret)
	}
	fb := resp["fallback"].(map[string]any)
	if fb["used"] != false {
		t.Fatalf("fallback %v", fb)
	}
	outs := resp["outputs"].([]any)
	if len(outs) != 2 {
		t.Fatalf("outputs %v", outs)
	}
	o0 := outs[0].(map[string]any)
	if o0["seed"] != float64(10) {
		t.Fatalf("seed0 %v", o0["seed"])
	}
	o1 := outs[1].(map[string]any)
	if o1["seed"] != float64(11) {
		t.Fatalf("seed1 %v", o1["seed"])
	}
	sum := sha256.Sum256(fixturePNG)
	if o0["sha256"] != hex.EncodeToString(sum[:]) {
		t.Fatalf("sha %v", o0["sha256"])
	}
	ref := o0["ref"].(string)
	req := httptest.NewRequest(http.MethodGet, "/v1/image/outputs/"+ref, nil)
	req.Header.Set("Authorization", "Bearer "+testKey)
	dl := httptest.NewRecorder()
	s.ServeHTTP(dl, req)
	if dl.Code != 200 || !bytes.Equal(dl.Body.Bytes(), fixturePNG) {
		t.Fatalf("download %d %v", dl.Code, dl.Body.Bytes())
	}
	// other license cannot see the output
	s2dir := MapDirectory{"other": {ID: "lic-other", Kind: "sub", Active: true}}
	// same server, other token unknown
	req = httptest.NewRequest(http.MethodGet, "/v1/image/outputs/"+ref, nil)
	req.Header.Set("Authorization", "Bearer other")
	deny := httptest.NewRecorder()
	s.ServeHTTP(deny, req)
	if deny.Code != 401 {
		t.Fatalf("other lic %d %s", deny.Code, deny.Body)
	}
	_ = s2dir
}

func TestIdempotencyReplayAndConflict(t *testing.T) {
	s := New(Options{
		Directory: subDir(),
		Rates:     map[RateKey]int{{"klein", "1024x1024"}: 2},
		Snapshot: EngineSnapshot{
			State: "IMAGE_READY", SlotReady: true, MaxUptimeOK: true,
			ModelStaged: map[string]bool{"klein": true},
		},
		Generator: fixtureGen,
	})
	rec1 := doPOST(t, s, `{"prompt":"a"}`, authIdem())
	rec2 := doPOST(t, s, `{"prompt":"a"}`, authIdem())
	if rec1.Code != 200 || rec2.Code != 200 {
		t.Fatalf("%d %d", rec1.Code, rec2.Code)
	}
	if rec2.Header().Get("Idempotent-Replay") != "true" {
		t.Fatalf("missing replay header")
	}
	if rec1.Body.String() == "" || !strings.Contains(rec2.Body.String(), `"id"`) {
		t.Fatalf("replay body %s", rec2.Body)
	}
	rec3 := doPOST(t, s, `{"prompt":"b"}`, authIdem())
	if rec3.Code != 409 || !strings.Contains(rec3.Body.String(), "idempotency_key_reused") {
		t.Fatalf("%d %s", rec3.Code, rec3.Body)
	}
}

func TestCold202RecordsBootIntentWithoutGPU(t *testing.T) {
	var intents []string
	s := New(Options{
		Directory:      subDir(),
		Rates:          map[RateKey]int{{"klein", "1024x1024"}: 2},
		AcceptColdSwap: true,
		Snapshot:       EngineSnapshot{State: "COLD", ModelStaged: map[string]bool{"klein": true}},
		OnBootIntent:   func(id string) { intents = append(intents, id) },
	})
	rec := doPOST(t, s, `{"prompt":"cold"}`, authIdem())
	if rec.Code != 202 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Location") == "" {
		t.Fatalf("missing Location")
	}
	var acc map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &acc)
	id := acc["id"].(string)
	if !s.JobBootIntent(id) || len(intents) != 1 {
		t.Fatalf("intent %v jobs %v", intents, s.JobBootIntent(id))
	}
	// poll must not add another boot intent
	req := httptest.NewRequest(http.MethodGet, "/v1/image/jobs/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+testKey)
	pol := httptest.NewRecorder()
	s.ServeHTTP(pol, req)
	if pol.Code != 200 || !strings.Contains(pol.Body.String(), "starting") {
		t.Fatalf("poll %d %s", pol.Code, pol.Body)
	}
	if len(intents) != 1 {
		t.Fatalf("poll triggered boot: %v", intents)
	}
	// replay POST returns same 202 job
	rec2 := doPOST(t, s, `{"prompt":"cold"}`, authIdem())
	if rec2.Code != 202 || !strings.Contains(rec2.Body.String(), id) {
		t.Fatalf("replay %d %s", rec2.Code, rec2.Body)
	}
}

func TestMethodNotAllowedAndTooLarge(t *testing.T) {
	s := New(Options{Directory: subDir()})
	req := httptest.NewRequest(http.MethodGet, "/v1/image", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != 405 || rec.Header().Get("Allow") != "POST" {
		t.Fatalf("%d %v", rec.Code, rec.Header())
	}
	big := `{"prompt":"` + strings.Repeat("a", 33*1024) + `"}`
	rec = doPOST(t, s, big, authIdem())
	if rec.Code != 413 || !strings.Contains(rec.Body.String(), "body_too_large") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestRetentionDoesNotDelete(t *testing.T) {
	now := time.Date(2026, 11, 8, 4, 0, 0, 0, time.UTC)
	s := New(Options{
		Directory: subDir(),
		Rates:     map[RateKey]int{{"klein", "1024x1024"}: 1},
		Snapshot: EngineSnapshot{
			State: "IMAGE_READY", SlotReady: true, MaxUptimeOK: true,
			ModelStaged: map[string]bool{"klein": true},
		},
		Generator: fixtureGen,
		Now:       func() time.Time { return now },
	})
	rec := doPOST(t, s, `{"prompt":"x"}`, authIdem())
	if rec.Code != 200 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	ref := resp["outputs"].([]any)[0].(map[string]any)["ref"].(string)
	now = now.Add(25 * time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/v1/image/outputs/"+ref, nil)
	req.Header.Set("Authorization", "Bearer "+testKey)
	deny := httptest.NewRecorder()
	s.ServeHTTP(deny, req)
	if deny.Code != 401 {
		t.Fatalf("after 24h %d %s", deny.Code, deny.Body)
	}
	s.mu.Lock()
	_, still := s.outputs[ref]
	s.mu.Unlock()
	if !still {
		t.Fatal("auto-deleted output; contract forbids automatic deletion")
	}
}

func TestNoPublicOutputWithoutAuth(t *testing.T) {
	s := New(Options{Directory: subDir()})
	req := httptest.NewRequest(http.MethodGet, "/v1/image/outputs/imgout_nope", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestDuplicateKeyInvalidJSON(t *testing.T) {
	s := New(Options{Directory: subDir()})
	rec := doPOST(t, s, `{"prompt":"a","prompt":"b"}`, authIdem())
	if rec.Code != 400 || !strings.Contains(rec.Body.String(), "invalid_json") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
