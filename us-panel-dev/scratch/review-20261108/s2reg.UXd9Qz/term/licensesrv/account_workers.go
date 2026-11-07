package licensesrv

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type metadataBuffer struct {
	header http.Header
	code   int
	body   bytes.Buffer
}

func (b *metadataBuffer) Header() http.Header  { return b.header }
func (b *metadataBuffer) WriteHeader(code int) { b.code = code }
func (b *metadataBuffer) Write(p []byte) (int, error) {
	if b.code == 0 {
		b.code = 200
	}
	return b.body.Write(p)
}

// Reuses authenticated read-only metadata. Does not claim GCP provider status.
func (p *ModelProxy) AccountWorkers(w http.ResponseWriter, r *http.Request) {
	q := r.Clone(r.Context())
	q.Method = "GET"
	b := &metadataBuffer{header: make(http.Header)}
	p.Models(b, q)
	if b.code != 200 {
		accountError(w, b.code)
		return
	}
	var listing struct {
		Data []struct {
			ID     string `json:"id"`
			Status struct {
				Value string `json:"value"`
			} `json:"status"`
		} `json:"data"`
	}
	if json.Unmarshal(b.body.Bytes(), &listing) != nil {
		accountError(w, 503)
		return
	}
	workers := []map[string]any{}
	for _, m := range listing.Data {
		if m.ID != p.Model {
			continue
		}
		phase := "unknown"
		ready := m.Status.Value == "loaded"
		if ready {
			phase = "ready"
		}
		workers = append(workers, map[string]any{"worker": "model-upstream", "model": m.ID, "phase": phase, "provider_status": "unknown", "active": len(p.slots), "idle_seconds": 900, "idle_remaining": nil, "observed_at": time.Now().Unix(), "live": false, "model_ready": ready})
	}
	accountJSON(w, 200, map[string]any{"workers": workers})
}
