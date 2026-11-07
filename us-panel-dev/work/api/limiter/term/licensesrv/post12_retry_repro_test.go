//go:build post12repro

package licensesrv

// These acceptance reproducers intentionally FAIL until the idempotency
// protocol proposal is approved. Not included in the default regression run.
import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPost12ReproChatRetryMustSettleOnce(t *testing.T) {
	p, st, key := proxyFixture(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, goodEvents)
	})
	for i := 0; i < 2; i++ {
		r := httptest.NewRequest("POST", "/v1/chat", strings.NewReader(proxyInput)).WithContext(context.Background())
		r.Header.Set("Authorization", "Bearer "+key)
		r.Header.Set("X-US-Request-ID", "same-logical-request")
		w := httptest.NewRecorder()
		p.ServeHTTP(w, r)
		if !strings.Contains(w.Body.String(), "[DONE]") {
			t.Fatal("request failed")
		}
	}
	l, _ := st.License(key)
	if l.Credits != 499 || l.Reserved != 0 {
		t.Fatalf("same request charged twice: credits=%d reserved=%d want=499/0", l.Credits, l.Reserved)
	}
}
func TestPost12ReproReserveLostResponseMustNotLeak(t *testing.T) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "post12")
	key := h.Store.st.Orders["555"][0]
	api := &CreditHandler{Store: h.Store, Now: h.Cfg.Now}
	for i := 0; i < 2; i++ {
		r := httptest.NewRequest("POST", "/v1/credits/reserve", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer "+key)
		r.Header.Set("X-US-Request-ID", "same-reserve-attempt")
		w := httptest.NewRecorder()
		api.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatal(w.Code)
		}
	}
	l, _ := h.Store.License(key)
	if l.Reserved != 1 {
		t.Fatalf("lost reserve response leaks hold: reserved=%d want=1", l.Reserved)
	}
}
