package licensesrv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWorkerMetadataIdlePolicyTenMinutes(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "p")
	key := h.Store.st.Orders["555"][0]
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"data":[{"id":"test-model"}]}`) }))
	defer s.Close()
	p, _ := NewModelProxy(h.Store, s.URL, "", "test-model", time.Second, 1, 2)
	req := httptest.NewRequest("GET", "/account/workers", nil)
	req.Header.Set("Authorization", "Bearer "+key)
	w := httptest.NewRecorder()
	p.AccountWorkers(w, req)
	var response struct {
		Workers []struct {
			IdleSeconds   int  `json:"idle_seconds"`
			IdleRemaining *int `json:"idle_remaining"`
			Live          bool `json:"live"`
		} `json:"workers"`
	}
	if json.Unmarshal(w.Body.Bytes(), &response) != nil || len(response.Workers) != 1 {
		t.Fatal("invalid metadata")
	}
	worker := response.Workers[0]
	if worker.IdleSeconds != 600 || worker.IdleRemaining != nil || worker.Live {
		t.Fatal("idle target must be 10 minutes, without inventing a verified countdown")
	}
	if DefaultGPUWarmup != 15*time.Minute {
		t.Fatal("idle policy changed warmup timeout")
	}
}
