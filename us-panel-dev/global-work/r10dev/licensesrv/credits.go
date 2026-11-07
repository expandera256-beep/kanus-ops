package licensesrv

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
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
	if (r.URL.Path == "/v1/credits/status" && r.Method != http.MethodGet) || (r.URL.Path != "/v1/credits/status" && r.Method != http.MethodPost) {
		WriteAPIError(w, r, 405, "method_not_allowed", "method")
		return
	}
	key := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if r.URL.Path == "/v1/credits/status" && strings.HasPrefix(key, "usa1.") {
		if key == r.Header.Get("Authorization") {
			WriteAPIError(w, r, 401, "unauthorized", "session")
			return
		}
		out, e := h.Store.accountCreditStatus(key, time.Now())
		if e != nil {
			WriteAPIError(w, r, 401, "unauthorized", "session")
			return
		}
		accountJSON(w, 200, out)
		return
	}
	l, ok := h.Store.License(key)
	if key == "" || key == r.Header.Get("Authorization") || !ok {
		WriteAPIError(w, r, 401, "unauthorized", "license")
		return
	}
	now := time.Now()
	if h.Now != nil {
		now = h.Now()
	}
	var out any
	switch r.URL.Path {
	case "/v1/credits/status":
		out, _ = h.Store.CreditStatus(key, now)
	case "/v1/credits/reserve":
		if l.Plan.Kind != "sub" {
			WriteAPIError(w, r, 403, "hourly_requires_proxy", "hourly seconds metering requires the server model proxy")
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
			WriteAPIError(w, r, status, "reserve_rejected", "reserve")
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
			WriteAPIError(w, r, 400, "invalid_json", "json")
			return
		}
		if err := h.Store.settleFor(key, in.ID, in.Commit, now); err != nil {
			WriteAPIError(w, r, 409, "settle_conflict", "settle")
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

// CreditStatus exposes only the authenticated license's balance and aggregate
// settlement counts. A read neither sweeps reservations nor wakes a worker.
type CreditUsage struct {
	At     time.Time `json:"at"`
	Amount int       `json:"amount"`
	Commit bool      `json:"commit"`
}
type CreditStatus struct {
	Kind       string        `json:"kind"`
	Remaining  int           `json:"remaining"`
	Unit       string        `json:"unit"`
	History    []CreditUsage `json:"history"`
	Credits    int           `json:"credits"`
	Reserved   int           `json:"reserved"`
	Available  int           `json:"available"`
	Committed  int           `json:"committed"`
	Released   int           `json:"released"`
	Revoked    bool          `json:"revoked"`
	Valid      bool          `json:"valid"`
	ExpiresAt  time.Time     `json:"expiresAt"`
	ObservedAt time.Time     `json:"observedAt"`
}

func (s *Store) CreditStatus(key string, now time.Time) (CreditStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.st.Licenses[key]
	if !ok {
		return CreditStatus{}, false
	}
	unit := l.MeterUnit
	if l.Plan.Kind == "use" {
		unit = "seconds"
	}
	if unit == "" {
		unit = "credits"
	}
	out := CreditStatus{Kind: l.Plan.Kind, Remaining: l.Credits, Unit: unit, History: []CreditUsage{}, Credits: l.Credits, Reserved: l.Reserved, Available: l.Credits - l.Reserved, Revoked: l.Revoked, Valid: !l.Revoked && now.Before(l.ExpiresAt), ExpiresAt: l.ExpiresAt, ObservedAt: now}
	for _, v := range s.st.Settlements {
		if v.Key == key {
			out.History = append(out.History, CreditUsage{At: v.SettledAt, Amount: v.Amount, Commit: v.Commit})
			if v.Commit {
				out.Committed++
			} else {
				out.Released++
			}
		}
	}
	sort.Slice(out.History, func(i, j int) bool { return out.History[i].At.After(out.History[j].At) })
	if len(out.History) > 20 {
		out.History = out.History[:20]
	}
	return out, true
}
