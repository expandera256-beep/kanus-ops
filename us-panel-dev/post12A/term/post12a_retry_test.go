package main

// POST12-A EXE tarafı: en fazla 3 deneme, AYNI X-US-Request-ID ve aynı gövde,
// yalnız geçici hatalarda (bağlantı/timeout/reset ve 5xx; GPU warming hariç).
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"kiyas/licensesrv"
)

const p12aStream = "data: {\"choices\":[{\"delta\":{\"content\":\"tamam\"}}]}\n\ndata: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"

func TestPost12AModelRetrySameIDOnTransient(t *testing.T) {
	prev := releaseAuthority
	releaseAuthority = "go-hook" // written-request reset retried only against the hook
	t.Cleanup(func() { releaseAuthority = prev })
	var mu sync.Mutex
	var ids, bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		mu.Lock()
		ids = append(ids, r.Header.Get("X-US-Request-ID"))
		bodies = append(bodies, string(raw))
		n := len(ids)
		mu.Unlock()
		if n == 1 {
			w.Header().Set("X-US-Idempotency", "1")
			http.Error(w, "temporary", 502)
			return
		}
		if n == 2 {
			c, _, _ := w.(http.Hijacker).Hijack()
			c.Close() // connection reset before any response
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, p12aStream)
	}))
	defer srv.Close()
	out := postCompletion(srv.URL+"/v1/chat/completions", "a100", "test", "merhaba", "system", 32)
	if !out.ok || len(ids) != 3 {
		t.Fatalf("ok=%v attempts=%d body=%q", out.ok, len(ids), out.body)
	}
	if ids[0] == "" || ids[0] != ids[1] || ids[1] != ids[2] || bodies[0] != bodies[1] || bodies[1] != bodies[2] {
		t.Fatalf("retry changed id/body: %v", ids)
	}
}

func TestPost12AModelRetryBoundedAndNot4xxOrWarming(t *testing.T) {
	cases := []struct {
		name string
		h    http.HandlerFunc
		want int32
	}{
		{"5xx-max-3", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-US-Idempotency", "1")
			http.Error(w, "x", 500)
		}, 3},
		{"5xx-legacy-server-no-retry", func(w http.ResponseWriter, r *http.Request) { http.Error(w, "x", 500) }, 1},
		{"429-retried", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-US-Idempotency", "1")
			http.Error(w, "x", 429)
		}, 3},
		{"409", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-US-Idempotency", "1")
			w.WriteHeader(409)
			fmt.Fprint(w, `{"error":"request_terminal"}`)
		}, 1},
		{"402", func(w http.ResponseWriter, r *http.Request) { http.Error(w, "x", 402) }, 1},
		{"warming", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-US-GPU-State", "warming")
			w.Header().Set("Retry-After", "30")
			w.WriteHeader(503)
			fmt.Fprint(w, `{"status":"warming","message":"GPU ısınıyor."}`)
		}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1); tc.h(w, r) }))
			defer srv.Close()
			out := postCompletion(srv.URL+"/v1/chat/completions", "a100", "test", "merhaba", "system", 32)
			if out.ok || hits.Load() != tc.want {
				t.Fatalf("ok=%v attempts=%d want=%d", out.ok, hits.Load(), tc.want)
			}
		})
	}
}

// Reserve ACK kaybı: sunucu hold'u açar ama yanıt kaybolur; EXE aynı kimlikle
// tekrar eder, sunucu aynı rezervasyonu döner. Tek hold, tek ücret.
func TestPost12AGateReserveLostAckSameID(t *testing.T) {
	gate, st, key := gatewayCreditFixture(t)
	api := &licensesrv.CreditHandler{Store: st}
	var reserves atomic.Int32
	var mu sync.Mutex
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/credits/reserve" {
			mu.Lock()
			seen = append(seen, r.Header.Get("X-US-Request-ID"))
			mu.Unlock()
			if reserves.Add(1) == 1 {
				rec := httptest.NewRecorder()
				api.ServeHTTP(rec, r)
				c, _, _ := w.(http.Hijacker).Hijack()
				c.Close()
				return
			}
		}
		api.ServeHTTP(w, r)
	}))
	defer srv.Close()
	gate.base = srv.URL
	calls := 0
	out := gate.call(context.Background(), func() modelResult {
		calls++
		s, _ := st.CreditStatus(key, time.Now())
		if s.Reserved != 1 {
			t.Errorf("extra hold: reserved=%d", s.Reserved)
		}
		return modelResult{ok: true, body: "ok"}
	})
	s, _ := st.CreditStatus(key, time.Now())
	if !out.ok || calls != 1 || reserves.Load() != 2 || s.Reserved != 0 || s.Credits != 1 || s.Committed != 1 {
		t.Fatalf("ok=%v body=%q model=%d reserves=%d reserved=%d credits=%d commits=%d", out.ok, out.body, calls, reserves.Load(), s.Reserved, s.Credits, s.Committed)
	}
	if len(seen) != 2 || seen[0] == "" || seen[0] != seen[1] {
		t.Fatalf("reserve retry must reuse the request id: %v", seen)
	}
}

func TestPost12AGateReserveSendsBodyRequestID(t *testing.T) {
	var got struct {
		RequestID string `json:"idempotency_id"`
	}
	var header string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/credits/reserve" {
			header = r.Header.Get("X-US-Request-ID")
			_ = json.NewDecoder(r.Body).Decode(&got)
			fmt.Fprint(w, `{"id":"r1"}`)
			return
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer srv.Close()
	_ = (creditGate{base: srv.URL, token: "test", required: true}).call(context.Background(), func() modelResult { return modelResult{ok: true} })
	if len(header) != 32 || header != got.RequestID {
		t.Fatalf("header=%q body=%q", header, got.RequestID)
	}
}

func TestPost12ARetryWindowCapped(t *testing.T) {
	prev := retryWindow
	retryWindow = 0
	t.Cleanup(func() { retryWindow = prev })
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("X-US-Idempotency", "1")
		http.Error(w, "x", 503)
	}))
	defer srv.Close()
	time.Sleep(time.Millisecond)
	_ = postCompletion(srv.URL+"/v1/chat/completions", "a100", "test", "merhaba", "system", 32)
	err := retryTransient(context.Background(), func(context.Context) error {
		hits.Add(1)
		return &creditRequestError{transient: true, message: "x"}
	})
	if err == nil || hits.Load() != 2 {
		t.Fatalf("window exceeded but retried: %d", hits.Load())
	}
}
