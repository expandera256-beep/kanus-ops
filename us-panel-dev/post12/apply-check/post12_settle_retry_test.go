package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"kiyas/licensesrv"
)

func TestPost12SettleLostAcknowledgementRetriesSameID(t *testing.T) {
	gate, st, key := gatewayCreditFixture(t)
	var hits atomic.Int32
	api := &licensesrv.CreditHandler{Store: st}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/credits/settle" && hits.Add(1) == 1 {
			rec := httptest.NewRecorder()
			api.ServeHTTP(rec, r)
			if rec.Code != 200 {
				t.Errorf("first settle rejected: %d", rec.Code)
			}
			c, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			_ = c.Close()
			return
		}
		api.ServeHTTP(w, r)
	}))
	defer srv.Close()
	gate.base = srv.URL
	calls := 0
	out := gate.call(context.Background(), func() modelResult { calls++; return modelResult{ok: true, body: "answer"} })
	status, _ := st.CreditStatus(key, time.Now())
	if !out.ok || calls != 1 || hits.Load() != 2 || status.Credits != 1 || status.Reserved != 0 || status.Committed != 1 {
		t.Fatalf("ok=%v model=%d settle=%d credit=%d reserved=%d commits=%d", out.ok, calls, hits.Load(), status.Credits, status.Reserved, status.Committed)
	}
}

func TestPost12ReleaseTransientFailureRetries(t *testing.T) {
	gate, st, key := gatewayCreditFixture(t)
	var hits atomic.Int32
	api := &licensesrv.CreditHandler{Store: st}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/credits/settle" && hits.Add(1) == 1 {
			http.Error(w, "temporary", 503)
			return
		}
		api.ServeHTTP(w, r)
	}))
	defer srv.Close()
	gate.base = srv.URL
	out := gate.call(context.Background(), func() modelResult { return modelResult{body: "model failed"} })
	status, _ := st.CreditStatus(key, time.Now())
	if out.ok || hits.Load() != 2 || status.Reserved != 0 || status.Credits != 2 || status.Released != 1 {
		t.Fatalf("settle=%d reserved=%d credits=%d released=%d", hits.Load(), status.Reserved, status.Credits, status.Released)
	}
}

func TestPost12ReserveNeverRetriedAndSettleNeverRetries4xx(t *testing.T) {
	for _, path := range []string{"/v1/credits/reserve", "/v1/credits/settle"} {
		t.Run(path, func(t *testing.T) {
			var hits atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == path {
					hits.Add(1)
					http.Error(w, "reject", 409)
					return
				}
				fmt.Fprint(w, `{"id":"r"}`)
			}))
			defer srv.Close()
			_ = (creditGate{base: srv.URL, token: "test", required: true}).call(context.Background(), func() modelResult { return modelResult{ok: true} })
			if hits.Load() != 1 {
				t.Fatalf("retried non-idempotent reserve or 4xx: %d", hits.Load())
			}
		})
	}
}
