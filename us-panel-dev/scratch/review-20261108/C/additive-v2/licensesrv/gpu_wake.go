package licensesrv

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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
	// ReadyGrace: after model_ready the VPS tunnel may lag; within this
	// window an unreachable upstream does not trigger another start.
	ReadyGrace time.Duration
	Client     *http.Client

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

func validateWakeURL(raw, want string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("invalid GPU wake/status URL")
	}
	p := strings.ToLower(u.EscapedPath())
	if strings.Contains(p, "health") || strings.Contains(p, "/v1/") || p != want {
		return errors.New("GPU URL must be the /a100 control endpoint; /health and model paths are refused (they wake the GPU)")
	}
	return nil
}

func NewGPUWaker(wakeURL, statusURL, tokenFile string, warmup time.Duration) (*GPUWaker, error) {
	if err := validateWakeURL(wakeURL, "/a100/start"); err != nil {
		return nil, err
	}
	if err := validateWakeURL(statusURL, "/a100/status"); err != nil {
		return nil, err
	}
	if tokenFile == "" {
		tokenFile = "/etc/us-model/wake.token"
	}
	if warmup < MinGPUWarmup {
		warmup = MinGPUWarmup
	}
	return &GPUWaker{WakeURL: wakeURL, StatusURL: statusURL, TokenFile: tokenFile, WarmupTimeout: warmup, PollMin: 5 * time.Second, PollMax: 30 * time.Second, FailCooldown: 30 * time.Second, ReadyGrace: 3 * time.Minute,
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

// MarkDown is called when the upstream is unreachable after being ready
// (e.g. idle stop): the next Trigger starts the GPU again.
func (g *GPUWaker) MarkDown() {
	g.mu.Lock()
	if !g.running && g.state == "ready" && time.Since(g.since) > g.ReadyGrace {
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
	resp, err := g.Client.Do(req)
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

// upstreamReachable is a TCP connect probe to the model upstream (the VPS
// loopback tunnel). It sends no HTTP and cannot wake anything.
func upstreamReachable(ctx context.Context, upstream string) bool {
	u, err := url.Parse(upstream)
	if err != nil {
		return false
	}
	host := u.Host
	if u.Port() == "" {
		if u.Scheme == "https" {
			host = net.JoinHostPort(u.Hostname(), "443")
		} else {
			host = net.JoinHostPort(u.Hostname(), "80")
		}
	}
	d := net.Dialer{Timeout: 2 * time.Second}
	c, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

func writeWarming(w http.ResponseWriter, st GPUStatus) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "30")
	w.Header().Set("X-US-GPU-State", st.State)
	w.WriteHeader(http.StatusServiceUnavailable)
	msg := "GPU ısınıyor; model hazırlanıyor. Kredi ayrılmadı veya düşülmedi. Lütfen yeniden deneyin."
	if st.State == "failed" {
		msg = "GPU başlatılamadı; kredi düşülmedi. Biraz sonra yeniden deneyin."
	}
	body, _ := json.Marshal(map[string]any{"status": "warming", "state": st.State, "reason": st.Reason, "message": msg, "retry_after": 30})
	_, _ = w.Write(body)
}
