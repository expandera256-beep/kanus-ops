package licensesrv

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fakeModelUpstream: port always open; /v1/models behaviour is switchable.
type fakeModelUpstream struct {
	mode       atomic.Int32 // 0 models 500, 1 lists other model, 2 lists test-model
	chat503    atomic.Bool
	chatCalls  atomic.Int32
	probeCalls atomic.Int32
	srv        *httptest.Server
}

func newFakeModelUpstream(t *testing.T, onChat func()) *fakeModelUpstream {
	f := &fakeModelUpstream{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(strings.ToLower(r.URL.Path), "health"):
			t.Errorf("/health çağrıldı")
		case r.URL.Path == "/v1/models":
			f.probeCalls.Add(1)
			switch f.mode.Load() {
			case 0:
				http.Error(w, "loading", 500)
			case 1:
				fmt.Fprint(w, `{"data":[{"id":"other-model"}]}`)
			default:
				fmt.Fprint(w, `{"data":[{"id":"test-model"}]}`)
			}
		case r.URL.Path == "/v1/chat/completions":
			f.chatCalls.Add(1)
			if onChat != nil {
				onChat()
			}
			if f.chat503.Load() {
				http.Error(w, "busy", 503)
				return
			}
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, goodEvents)
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func readyFixture(t *testing.T, withWaker bool, readyAfter int32) (*ModelProxy, *fakeModelUpstream, *fakeGPU, *Store, string) {
	h, _ := setup(t)
	post(h, payload, sign("whsec", payload), "p")
	key := h.Store.st.Orders["555"][0]
	f := newFakeModelUpstream(t, func() {
		if l, _ := h.Store.License(key); l.Reserved != 1 {
			t.Error("upstream çağrısında rezervasyon yok")
		}
	})
	p, err := NewModelProxy(h.Store, f.srv.URL, "", "test-model", 45*time.Minute, 1, 32)
	if err != nil {
		t.Fatal(err)
	}
	var g *fakeGPU
	if withWaker {
		g = newFakeGPU(t, 200, readyAfter)
		p.Waker = testWaker(t, g)
	}
	return p, f, g, h.Store, key
}

func TestReadyPortOpenModelNotLoadedNoChatNoReservation(t *testing.T) {
	for _, mode := range []int32{0, 1} {
		p, f, _, st, key := readyFixture(t, false, 0)
		f.mode.Store(mode)
		before, _ := st.License(key)
		for i := 0; i < 3; i++ {
			w := proxyRequest(p, key, proxyInput, context.Background())
			if w.Code != 503 || !strings.Contains(w.Body.String(), "GPU ısınıyor") || w.Header().Get("Retry-After") == "" {
				t.Fatalf("mode %d: %d %s", mode, w.Code, w.Body.String())
			}
		}
		after, _ := st.License(key)
		if f.chatCalls.Load() != 0 || after.Reserved != 0 || after.Credits != before.Credits {
			t.Fatalf("mode %d: chat=%d %+v", mode, f.chatCalls.Load(), after)
		}
	}
}

func TestReadyModelReadyFalseThenTrue(t *testing.T) {
	p, f, g, st, key := readyFixture(t, true, 4)
	f.mode.Store(2) // port open and listing, but /a100/status says not ready
	if w := proxyRequest(p, key, proxyInput, context.Background()); w.Code != 503 {
		t.Fatalf("model_ready=false iken %d", w.Code)
	}
	if f.chatCalls.Load() != 0 {
		t.Fatal("model_ready=false iken chat çağrıldı")
	}
	if l, _ := st.License(key); l.Reserved != 0 {
		t.Fatal("rezervasyon")
	}
	waitState(t, p.Waker, "ready")
	w := proxyRequest(p, key, proxyInput, context.Background())
	if w.Code != 200 || !strings.Contains(w.Body.String(), "[DONE]") || !strings.Contains(w.Header().Get("Content-Type"), "event-stream") {
		t.Fatalf("hazırken SSE: %d %s", w.Code, w.Body.String())
	}
	if atomic.LoadInt32(&g.wakeCalls) != 1 || f.chatCalls.Load() != 1 {
		t.Fatalf("start=%d chat=%d", g.wakeCalls, f.chatCalls.Load())
	}
}

func TestReadyCacheExpiry(t *testing.T) {
	p, f, _, _, key := readyFixture(t, false, 0)
	f.mode.Store(2)
	p.ReadyTTL = 80 * time.Millisecond
	for i := 0; i < 3; i++ {
		if w := proxyRequest(p, key, proxyInput, context.Background()); w.Code != 200 {
			t.Fatal(w.Code)
		}
	}
	if n := f.probeCalls.Load(); n != 1 {
		t.Fatalf("önbellek içinde %d prob", n)
	}
	time.Sleep(120 * time.Millisecond)
	f.mode.Store(0) // model unloaded after cache expired
	if w := proxyRequest(p, key, proxyInput, context.Background()); w.Code != 503 {
		t.Fatalf("süre dolunca yeniden prob yok: %d", w.Code)
	}
	if f.probeCalls.Load() != 2 || f.chatCalls.Load() != 3 {
		t.Fatalf("prob=%d chat=%d", f.probeCalls.Load(), f.chatCalls.Load())
	}
}

func TestReadyInvalidatedAfterUpstreamFailure(t *testing.T) {
	p, f, _, st, key := readyFixture(t, false, 0)
	f.mode.Store(2)
	p.ReadyTTL = time.Hour
	if w := proxyRequest(p, key, proxyInput, context.Background()); w.Code != 200 {
		t.Fatal(w.Code)
	}
	f.chat503.Store(true)
	f.mode.Store(0)
	before, _ := st.License(key)
	if w := proxyRequest(p, key, proxyInput, context.Background()); strings.Contains(w.Body.String(), "[DONE]") {
		t.Fatal("503 upstream başarı sayıldı")
	}
	// Cache must be invalidated: next request probes and gets warming, no chat.
	chats := f.chatCalls.Load()
	w := proxyRequest(p, key, proxyInput, context.Background())
	after, _ := st.License(key)
	if w.Code != 503 || !strings.Contains(w.Body.String(), `"status":"warming"`) || f.chatCalls.Load() != chats || after.Credits != before.Credits || after.Reserved != 0 {
		t.Fatalf("geçersizleştirme yok: %d chat %d→%d %+v", w.Code, chats, f.chatCalls.Load(), after)
	}
	// Connection error also invalidates.
	p2, f2, _, _, key2 := readyFixture(t, false, 0)
	f2.mode.Store(2)
	p2.ReadyTTL = time.Hour
	proxyRequest(p2, key2, proxyInput, context.Background())
	f2.srv.CloseClientConnections()
	f2.srv.Close()
	proxyRequest(p2, key2, proxyInput, context.Background())
	p2.readyMu.Lock()
	cached := time.Now().Before(p2.readyUntil)
	p2.readyMu.Unlock()
	if cached {
		t.Fatal("bağlantı hatası sonrası önbellek geçersizleşmedi")
	}
}

func TestReadyWakerReprobedAfterUpstream503(t *testing.T) {
	p, f, g, _, key := readyFixture(t, true, 1)
	f.mode.Store(2)
	p.ReadyTTL = time.Hour
	if w := proxyRequest(p, key, proxyInput, context.Background()); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := proxyRequest(p, key, proxyInput, context.Background()); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if n := atomic.LoadInt32(&g.statusCalls); n != 1 {
		t.Fatalf("önbellekte status %d kez", n)
	}
	f.chat503.Store(true)
	proxyRequest(p, key, proxyInput, context.Background())
	proxyRequest(p, key, proxyInput, context.Background())
	if n := atomic.LoadInt32(&g.statusCalls); n < 2 {
		t.Fatalf("503 sonrası model_ready yeniden sorulmadı: %d", n)
	}
}
