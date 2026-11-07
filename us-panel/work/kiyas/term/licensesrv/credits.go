package licensesrv

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// CreditHandler authenticates with the opaque license key, not a global secret.
// It never accepts a caller-supplied amount/TTL. Deploy only behind HTTPS.
type CreditHandler struct {
	Store *Store
	Now   func() time.Time
}

const ReservationTTL = 45 * time.Minute

func (h *CreditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	key := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	l, ok := h.Store.License(key)
	if key == "" || key == r.Header.Get("Authorization") || !ok {
		http.Error(w, "license", 401)
		return
	}
	now := time.Now()
	if h.Now != nil {
		now = h.Now()
	}
	var out any
	switch r.URL.Path {
	case "/v1/credits/reserve":
		if l.Plan.Kind != "sub" {
			http.Error(w, "hourly credit policy not configured", 403)
			return
		}
		id, err := h.Store.Reserve(key, 1, ReservationTTL, now)
		if err != nil {
			status := 503
			if err == ErrInvalid {
				status = 403
			}
			if err == ErrNoCredit {
				status = 402
			}
			http.Error(w, "reserve", status)
			return
		}
		out = map[string]string{"id": id}
	case "/v1/credits/settle":
		var in struct {
			ID     string `json:"id"`
			Commit bool   `json:"commit"`
		}
		dec := json.NewDecoder(io.LimitReader(r.Body, 4097))
		if dec.Decode(&in) != nil || in.ID == "" {
			http.Error(w, "json", 400)
			return
		}
		if err := h.Store.settleFor(key, in.ID, in.Commit, now); err != nil {
			http.Error(w, "settle", 409)
			return
		}
		out = map[string]bool{"ok": true}
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
