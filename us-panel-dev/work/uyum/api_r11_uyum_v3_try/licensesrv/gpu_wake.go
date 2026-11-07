package licensesrv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// GPUWaker starts the ds-a100 worker through usta-wake's /a100/start and
// polls /a100/status until model_ready. It never calls /health (on usta-wake
// port 80 that wakes the selected worker). The bearer token is read from
// TokenFile at call time and never stored in state, logs or errors.
type GPUWaker struct {
	WakeURL, StatusURL, TokenFile string
	WarmupTimeout                 time.Duration
	PollMin, PollMax              time.Duration
	FailCooldown                  time.Duration
	Client                        *http.Client
	// PinFile: SPKI/PEM pin file for https control calls (empty: system CA).
	PinFile string

	mu      sync.Mutex
	state   string // idle, starting, polling, ready, failed
	reason  string
	since   time.Time
	running bool
}

// GPUStatus is safe to show to clients.
type GPUStatus struct {
	State  string    `json:"state"`
	Reason string    `json:"reason,omitempty"`
	Since  time.Time `json:"since"`
}

const MinGPUWarmup = 45 * time.Minute

func validateWakeURL(raw, want string, httpAllow map[string]bool) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("invalid GPU wake/status URL")
	}
	p := strings.ToLower(u.EscapedPath())
	if strings.Contains(p, "health") || strings.Contains(p, "/v1/") || p != want {
		return errors.New("GPU URL must be the /a100 control endpoint; /health and model paths are refused (they wake the GPU)")
	}
	if u.Scheme == "http" && !isLoopbackHost(u.Hostname()) && !httpAllow[hostPort(u)] {
		return fmt.Errorf("plain http to %s is refused: use https, a loopback host, or list exactly %s in US_GPU_WAKE_HTTP_ALLOW", hostPort(u), hostPort(u))
	}
	return nil
}

func isLoopbackHost(h string) bool {
	ip := net.ParseIP(h)
	return h == "localhost" || ip != nil && ip.IsLoopback()
}

func hostPort(u *url.URL) string {
	port := u.Port()
	if port == "" {
		port = "80"
		if u.Scheme == "https" {
			port = "443"
		}
	}
	return net.JoinHostPort(strings.ToLower(u.Hostname()), port)
}

// ParseHTTPAllow parses US_GPU_WAKE_HTTP_ALLOW ("host:port[,host:port]").
func ParseHTTPAllow(raw string) map[string]bool {
	out := map[string]bool{}
	for _, item := range strings.Split(raw, ",") {
		item = strings.ToLower(strings.TrimSpace(item))
		if h, p, err := net.SplitHostPort(item); err == nil && h != "" && p != "" {
			out[net.JoinHostPort(h, p)] = true
		}
	}
	return out
}

// InsecureHTTP reports whether the configured wake URLs use plain http to a
// non-loopback host (allowed only via US_GPU_WAKE_HTTP_ALLOW).
func (g *GPUWaker) InsecureHTTP() bool {
	for _, raw := range []string{g.WakeURL, g.StatusURL} {
		if u, err := url.Parse(raw); err == nil && u.Scheme == "http" && !isLoopbackHost(u.Hostname()) {
			return true
		}
	}
	return false
}

// NewGPUWaker reads the plain-http allowlist from US_GPU_WAKE_HTTP_ALLOW.
func NewGPUWaker(wakeURL, statusURL, tokenFile string, warmup time.Duration) (*GPUWaker, error) {
	g, err := NewGPUWakerAllow(wakeURL, statusURL, tokenFile, warmup, ParseHTTPAllow(os.Getenv("US_GPU_WAKE_HTTP_ALLOW")))
	if err == nil {
		g.PinFile = pinFileFromEnv()
	}
	return g, err
}

func NewGPUWakerAllow(wakeURL, statusURL, tokenFile string, warmup time.Duration, httpAllow map[string]bool) (*GPUWaker, error) {
	if err := validateWakeURL(wakeURL, "/a100/start", httpAllow); err != nil {
		return nil, err
	}
	if err := validateWakeURL(statusURL, "/a100/status", httpAllow); err != nil {
		return nil, err
	}
	if tokenFile == "" {
		tokenFile = "/etc/us-model/wake.token"
	}
	if warmup < MinGPUWarmup {
		warmup = MinGPUWarmup
	}
	return &GPUWaker{WakeURL: wakeURL, StatusURL: statusURL, TokenFile: tokenFile, WarmupTimeout: warmup, PollMin: 5 * time.Second, PollMax: 30 * time.Second, FailCooldown: 30 * time.Second,
		Client: &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, state: "idle", since: time.Now()}, nil
}

// String/GoString keep fmt from printing internals.
func (g *GPUWaker) String() string   { return "GPUWaker(" + g.Status().State + ")" }
func (g *GPUWaker) GoString() string { return g.String() }

func (g *GPUWaker) Status() GPUStatus {
	g.mu.Lock()
	defer g.mu.Unlock()
	return GPUStatus{State: g.state, Reason: g.reason, Since: g.since}
}

func (g *GPUWaker) set(state, reason string) {
	if state == "failed" {
		log.Printf("gpu wake: failed (%s)", reason) // fixed reason codes only; never the token
	}
	g.mu.Lock()
	g.state, g.reason, g.since = state, reason, time.Now()
	g.mu.Unlock()
}

// Trigger starts one warmup if none is running (singleflight). After a
// failure a new attempt is allowed only after FailCooldown.
func (g *GPUWaker) Trigger() {
	g.mu.Lock()
	if g.running || g.state == "ready" || (g.state == "failed" && time.Since(g.since) < g.FailCooldown) {
		g.mu.Unlock()
		return
	}
	g.running = true
	g.state, g.reason, g.since = "starting", "", time.Now()
	g.mu.Unlock()
	go g.run()
}

// MarkDown is called when the model is not verifiably ready (readiness
// probe failed, upstream connection error or 503): the next Trigger runs
// start (200 if already up) and polls model_ready again. No time grace.
func (g *GPUWaker) MarkDown() {
	g.mu.Lock()
	if !g.running && g.state == "ready" {
		g.state, g.reason, g.since = "idle", "upstream_down", time.Now()
	}
	g.mu.Unlock()
}

func (g *GPUWaker) token() (string, error) {
	raw, err := os.ReadFile(g.TokenFile)
	if err != nil {
		return "", errors.New("token_unreadable")
	}
	tok := strings.TrimSpace(string(raw))
	if tok == "" || strings.ContainsAny(tok, "\r\n") {
		return "", errors.New("token_invalid")
	}
	return tok, nil
}

func (g *GPUWaker) call(ctx context.Context, method, u string) (int, []byte, error) {
	tok, err := g.token()
	if err != nil {
		return 0, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return 0, nil, errors.New("request")
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Connection", "close")
	cl, err := g.client(u)
	if err != nil {
		return 0, nil, err
	}
	resp, err := cl.Do(req)
	if err != nil {
		return 0, nil, errors.New("unreachable") // never wrap: URL errors are fine but keep uniform
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return resp.StatusCode, body, nil
}

func (g *GPUWaker) run() {
	defer func() {
		g.mu.Lock()
		g.running = false
		g.mu.Unlock()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), g.WarmupTimeout)
	defer cancel()
	code, _, err := g.call(ctx, http.MethodPost, g.WakeURL)
	switch {
	case err != nil:
		g.set("failed", "wake_"+err.Error())
		return
	case code == 401 || code == 403:
		g.set("failed", "wake_auth")
		return
	case code != 200 && code != 202:
		g.set("failed", "wake_http")
		return
	}
	g.set("polling", "")
	wait := g.PollMin
	for {
		code, body, err := g.call(ctx, http.MethodGet, g.StatusURL)
		if err == nil && (code == 401 || code == 403) {
			g.set("failed", "status_auth")
			return
		}
		if err == nil && code == 200 {
			var st struct {
				ModelReady bool `json:"model_ready"`
			}
			if json.Unmarshal(body, &st) == nil && st.ModelReady {
				g.set("ready", "")
				return
			}
		}
		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			g.set("failed", "warmup_timeout")
			return
		case <-t.C:
		}
		if wait *= 2; wait > g.PollMax {
			wait = g.PollMax
		}
	}
}

// ModelReady asks /a100/status once and reports model_ready. It is a
// read-only status call (never /health, never start).
func (g *GPUWaker) ModelReady(ctx context.Context) bool {
	code, body, err := g.call(ctx, http.MethodGet, g.StatusURL)
	if err != nil || code != 200 {
		return false
	}
	var st struct {
		ModelReady bool `json:"model_ready"`
	}
	return json.Unmarshal(body, &st) == nil && st.ModelReady
}

func writeWarming(w http.ResponseWriter, st GPUStatus) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "30")
	w.Header().Set("X-US-GPU-State", st.State)
	w.WriteHeader(http.StatusServiceUnavailable)
	// Client sees only a coarse state and a generic message; internal
	// reasons (token_unreadable, wake_auth, ...) stay in the server log.
	msg := "GPU ısınıyor / geçici hata. Kredi ayrılmadı veya düşülmedi. Lütfen yeniden deneyin."
	body, _ := json.Marshal(map[string]any{"status": "warming", "state": st.State, "message": msg, "retry_after": 30})
	_, _ = w.Write(body)
}
