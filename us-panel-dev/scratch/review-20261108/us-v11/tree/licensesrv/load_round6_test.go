package licensesrv

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type loadSample struct {
	Index int     `json:"index"`
	Code  int     `json:"http"`
	MS    float64 `json:"ms"`
	Done  bool    `json:"sse_done"`
}

func loadSummary(t *testing.T, name string, s []loadSample) {
	t.Helper()
	ms := make([]float64, 0, len(s))
	counts := map[int]int{}
	for _, v := range s {
		ms = append(ms, v.MS)
		counts[v.Code]++
		raw, _ := json.Marshal(v)
		t.Logf("%s raw %s", name, raw)
	}
	sort.Float64s(ms)
	pct := func(p float64) float64 { return ms[int(math.Ceil(float64(len(ms))*p))-1] }
	t.Logf("METRIC %s n=%d p50_ms=%.3f p95_ms=%.3f http=%v", name, len(s), pct(.50), pct(.95), counts)
}
func loadCall(client *http.Client, url, key string, index int) loadSample {
	start := time.Now()
	req, _ := http.NewRequest("POST", url+"/v1/chat/completions", strings.NewReader(proxyInput))
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := client.Do(req)
	if err != nil {
		return loadSample{Index: index, Code: 0, MS: float64(time.Since(start).Microseconds()) / 1000}
	}
	body, e := io.ReadAll(resp.Body)
	resp.Body.Close()
	return loadSample{index, resp.StatusCode, float64(time.Since(start).Microseconds()) / 1000, e == nil && strings.Contains(string(body), "data: [DONE]")}
}
func TestRound6SSE50LicensesQueue32(t *testing.T) {
	h, _ := setup(t)
	keys := make([]string, 50)
	for i := range keys {
		body := strings.Replace(payload, `"id":555`, fmt.Sprintf(`"id":%d`, 1000+i), 1)
		if post(h, body, sign("whsec", body), fmt.Sprint(i)) != 200 {
			t.Fatal("mint")
		}
		keys[i] = h.Store.st.Orders[fmt.Sprint(1000+i)][0]
	}
	gate := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(gate) })
	up := newModelsServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-gate:
		case <-r.Context().Done():
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n")
		w.(http.Flusher).Flush()
		time.Sleep(5 * time.Millisecond)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n")
	}))
	defer up.Close()
	p, err := NewModelProxy(h.Store, up.URL, "", "test-model", time.Minute, 1, 32)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(p)
	defer srv.Close()
	client := &http.Client{Timeout: 15 * time.Second}
	results := make([]loadSample, 50)
	var wg sync.WaitGroup
	start := make(chan struct{})
	var rejected atomic.Int32
	for i := range keys {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i] = loadCall(client, srv.URL, keys[i], i)
			if results[i].Code == 429 {
				rejected.Add(1)
			}
		}(i)
	}
	close(start)
	deadline := time.Now().Add(5 * time.Second)
	for len(p.requests)+int(rejected.Load()) < 50 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	filled := len(p.requests)
	overflow := rejected.Load()
	once.Do(func() { close(gate) })
	wg.Wait()
	loadSummary(t, "50licenses_queue32", results)
	if filled != 33 || overflow != 17 {
		t.Fatalf("admission=%d rejection=%d want33/17", filled, overflow)
	}
	ok := 0
	for i, v := range results {
		l, _ := h.Store.License(keys[i])
		if v.Code == 200 && v.Done {
			ok++
			if l.Credits != 499 {
				t.Fatal("debit")
			}
		} else if v.Code != 429 || v.Done || l.Credits != 500 {
			t.Fatal("unexpected result", v)
		}
		if l.Reserved != 0 || l.Credits < 0 {
			t.Fatal("balance")
		}
	}
	if ok != 33 {
		t.Fatal("success", ok)
	}
	t.Log("PASS expected_overflow=17 unexpected_errors=0 peak_admitted=33 active_capacity=1")
}
func TestRound6SingleLicenseReserveRace(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "mint")
	key := h.Store.st.Orders["555"][0]
	h.Store.st.Licenses[key].Credits = 1
	srv := httptest.NewServer(&CreditHandler{Store: h.Store})
	defer srv.Close()
	client := &http.Client{Timeout: 10 * time.Second}
	samples := make([]loadSample, 50)
	ids := make(chan string, 50)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range samples {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			begin := time.Now()
			req, _ := http.NewRequest("POST", srv.URL+"/v1/credits/reserve", nil)
			req.Header.Set("Authorization", "Bearer "+key)
			resp, err := client.Do(req)
			if err != nil {
				t.Error("local request")
				return
			}
			defer resp.Body.Close()
			samples[i] = loadSample{Index: i, Code: resp.StatusCode, MS: float64(time.Since(begin).Microseconds()) / 1000}
			if resp.StatusCode == 200 {
				var v struct{ ID string }
				if json.NewDecoder(resp.Body).Decode(&v) != nil {
					t.Error("json")
				}
				ids <- v.ID
			} else {
				io.Copy(io.Discard, resp.Body)
			}
		}(i)
	}
	close(start)
	wg.Wait()
	close(ids)
	count := 0
	for id := range ids {
		count++
		if h.Store.Settle(id, true) != nil || h.Store.Settle(id, true) != nil {
			t.Fatal("settle")
		}
	}
	for _, s := range samples {
		if s.Code != 200 && s.Code != 402 {
			t.Fatal(s)
		}
	}
	l, _ := h.Store.License(key)
	if count != 1 || l.Credits != 0 || l.Reserved != 0 {
		t.Fatal(count, l)
	}
	loadSummary(t, "single_license_reserve_race", samples)
	t.Log("PASS successes=1 expected_402=49 negative_balance=0 double_spend=0")
}
func TestRound6ExpiredRevokedHourlySSE(t *testing.T) {
	p, st, key := proxyFixture(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n")
		w.(http.Flusher).Flush()
		time.Sleep(1100 * time.Millisecond)
		fmt.Fprint(w, goodEvents)
	})
	srv := httptest.NewServer(p)
	defer srv.Close()
	client := &http.Client{Timeout: 10 * time.Second}
	st.st.Licenses[key].ExpiresAt = time.Now().Add(-time.Second)
	a := loadCall(client, srv.URL, key, 0)
	if a.Code != 401 {
		t.Fatal(a)
	}
	st.st.Licenses[key].ExpiresAt = time.Now().Add(time.Hour)
	st.st.Licenses[key].Revoked = true
	b := loadCall(client, srv.URL, key, 1)
	if b.Code != 401 {
		t.Fatal(b)
	}
	l := st.st.Licenses[key]
	l.Revoked = false
	l.Plan = Plan{Kind: "use", Seconds: 3600, Days: 30}
	l.Credits = 3600
	l.MeterUnit = "seconds"
	start := time.Now()
	c := loadCall(client, srv.URL, key, 2)
	elapsed := time.Since(start)
	after, _ := st.License(key)
	if c.Code != 200 || !c.Done || after.Credits != 3598 || after.Reserved != 0 {
		t.Fatalf("hourly expected ceil1.1=2 actual=%d elapsed=%s code=%d", 3600-after.Credits, elapsed, c.Code)
	}
	loadSummary(t, "expired_revoked_hourly", []loadSample{a, b, c})
	t.Logf("PASS hourly_debit=2 expected_ceil_seconds=2 expired_http=%d revoked_http=%d unexpected_errors=0", a.Code, b.Code)
}
