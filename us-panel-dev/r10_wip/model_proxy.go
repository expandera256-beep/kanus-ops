package licensesrv

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ModelProxy struct {
	Store                  *Store
	Upstream, Token, Model string
	Timeout                time.Duration
	Client                 *http.Client
	// Waker, if set, starts the GPU when the upstream is unreachable.
	// Warmup never reserves or charges credit.
	Waker           *GPUWaker
	slots, requests chan struct{}
}

func NewModelProxy(st *Store, upstream, token, model string, timeout time.Duration, capacity, queued int) (*ModelProxy, error) {
	if timeout < time.Second || timeout > 50*time.Minute || capacity < 1 || capacity > 10 || queued < 1 || queued > 100 {
		return nil, errors.New("invalid model limits")
	}
	if upstream != "" {
		u, e := url.Parse(upstream)
		if e != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" && u.Path != "/" {
			return nil, errors.New("invalid model upstream")
		}
		ip := net.ParseIP(u.Hostname())
		local := u.Hostname() == "localhost" || ip != nil && ip.IsLoopback()
		if u.Scheme != "https" && !(u.Scheme == "http" && local) {
			return nil, errors.New("model upstream requires HTTPS or loopback tunnel")
		}
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return &ModelProxy{Store: st, Upstream: strings.TrimRight(upstream, "/"), Token: token, Model: model, Timeout: timeout, Client: client, slots: make(chan struct{}, capacity), requests: make(chan struct{}, capacity+queued)}, nil
}
func (p *ModelProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-US-Metered", "1")
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	key := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if key == r.Header.Get("Authorization") || !p.Store.authorizedCredential(key, time.Now()) {
		http.Error(w, "license", 401)
		return
	}
	if p.Upstream == "" {
		http.Error(w, "model upstream not configured", 503)
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, (1<<20)+1))
	if err != nil || len(raw) > 1<<20 {
		http.Error(w, "body", 413)
		return
	}
	var in map[string]json.RawMessage
	if json.Unmarshal(raw, &in) != nil {
		http.Error(w, "json", 400)
		return
	}
	// A single credit buys one bounded candidate, not arbitrary n/best_of.
	for _, field := range []string{"n", "best_of", "max_completion_tokens"} {
		if _, exists := in[field]; exists {
			http.Error(w, "unsupported generation override", 400)
			return
		}
	}
	var model string
	_ = json.Unmarshal(in["model"], &model)
	if model != p.Model {
		http.Error(w, "model", 400)
		return
	}
	var messages []json.RawMessage
	var maxTokens int
	var stream bool
	if json.Unmarshal(in["messages"], &messages) != nil || len(messages) < 1 || len(messages) > 256 || json.Unmarshal(in["max_tokens"], &maxTokens) != nil || maxTokens < 1 || maxTokens > 8192 {
		http.Error(w, "model request bounds", 400)
		return
	}
	_ = json.Unmarshal(in["stream"], &stream)
	if p.Waker != nil && !upstreamReachable(r.Context(), p.Upstream) {
		p.Waker.MarkDown()
		p.Waker.Trigger()
		writeWarming(w, p.Waker.Status())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), p.Timeout)
	defer cancel()
	select {
	case p.requests <- struct{}{}:
	default:
		http.Error(w, "model queue full", 429)
		return
	}
	defer func() { <-p.requests }()
	select {
	case p.slots <- struct{}{}:
	case <-ctx.Done():
		http.Error(w, "model queue timeout", 504)
		return
	}
	defer func() { <-p.slots }()
	id, held, hourly, err := p.Store.reserveModel(key, p.Timeout, time.Now())
	if err != nil {
		status := 503
		if err == ErrNoCredit {
			status = 402
		}
		if err == ErrInvalid {
			status = 403
		}
		http.Error(w, "credit reservation rejected", status)
		return
	}
	key = p.Store.reservationKey(id)
	if key == "" {
		http.Error(w, "reservation no longer active", 403)
		return
	}
	settled := false
	defer func() {
		if !settled {
			_ = p.Store.settleAmount(key, id, 0, time.Now())
		}
	}()
	if hourly {
		var stop context.CancelFunc
		ctx, stop = context.WithTimeout(ctx, time.Duration(held)*time.Second)
		defer stop()
	}
	started := time.Now()
	settle := func() error {
		amount := 1
		if hourly {
			amount = int(math.Ceil(time.Since(started).Seconds()))
			if amount < 1 {
				amount = 1
			}
			if amount > held {
				amount = held
			}
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := p.Store.settleAmount(key, id, amount, time.Now()); err != nil {
			return err
		}
		settled = true
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Upstream+"/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		http.Error(w, "upstream request", 502)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if p.Token != "" {
		req.Header.Set("Authorization", "Bearer "+p.Token)
	}
	type result struct {
		resp *http.Response
		err  error
	}
	ready := make(chan result)
	go func() {
		resp, e := p.Client.Do(req)
		select {
		case ready <- result{resp, e}:
		case <-ctx.Done():
			if resp != nil {
				resp.Body.Close()
			}
		}
	}()
	var resp *http.Response
	if stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("X-Accel-Buffering", "no")
		if !proxyWrite(w, ": GPU isiniyor; model hazirligi dogrulanmadi\n\n") {
			return
		}
		tick := time.NewTicker(15 * time.Second)
		defer tick.Stop()
		for resp == nil {
			select {
			case v := <-ready:
				if v.err != nil {
					proxyFault(w, "upstream unavailable")
					return
				}
				resp = v.resp
			case <-tick.C:
				if !proxyWrite(w, ": GPU isiniyor / kuyruk bekleniyor\n\n") {
					return
				}
			case <-ctx.Done():
				proxyFault(w, "model timeout or cancellation")
				return
			}
		}
	} else {
		select {
		case v := <-ready:
			if v.err != nil {
				http.Error(w, "upstream unavailable", 502)
				return
			}
			resp = v.resp
		case <-ctx.Done():
			http.Error(w, "model timeout or cancellation", 504)
			return
		}
	}
	defer resp.Body.Close()
	fail := func(message string) {
		if stream {
			proxyFault(w, message)
		} else {
			http.Error(w, message, 502)
		}
	}
	if resp.StatusCode != 200 {
		fail("upstream rejected model request")
		return
	}
	if !stream {
		body, e := io.ReadAll(io.LimitReader(resp.Body, (8<<20)+1))
		if e != nil || len(body) > 8<<20 || !proxyCompleted(body) {
			fail("incomplete model response")
			return
		}
		if err := settle(); err != nil {
			fail("settlement not confirmed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		proxyWrite(w, string(body))
		return
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "event-stream") {
		fail("expected upstream event stream")
		return
	}
	events := make(chan proxyEvent, 1)
	go scanProxyEvents(ctx, resp.Body, events)
	tick := time.NewTicker(15 * time.Second)
	defer tick.Stop()
	hasOutput := false
	for {
		select {
		case event := <-events:
			if event.err != nil {
				fail("incomplete model stream")
				return
			}
			if event.data == "[DONE]" {
				if !hasOutput {
					fail("empty model response")
					return
				}
				if err := settle(); err != nil {
					fail("settlement not confirmed")
					return
				}
				proxyWrite(w, "data: [DONE]\n\n")
				return
			}
			var doc struct {
				Error   json.RawMessage `json:"error"`
				Choices []struct {
					Delta struct {
						Content   string `json:"content"`
						Reasoning string `json:"reasoning_content"`
					} `json:"delta"`
					Finish string `json:"finish_reason"`
				} `json:"choices"`
			}
			if json.Unmarshal([]byte(event.data), &doc) != nil || len(doc.Error) > 0 && string(doc.Error) != "null" {
				fail("model stream error")
				return
			}
			terminal := false
			for _, v := range doc.Choices {
				if v.Delta.Content != "" || v.Delta.Reasoning != "" {
					hasOutput = true
				}
				if v.Finish != "" {
					if v.Finish != "stop" {
						fail("model response truncated")
						return
					}
					terminal = true
				}
			}
			if terminal {
				if !hasOutput {
					fail("empty model response")
					return
				}
				if err := settle(); err != nil {
					fail("settlement not confirmed")
					return
				}
			}
			if !proxyWrite(w, "data: "+event.data+"\n\n") {
				return
			}
			if terminal {
				proxyWrite(w, "data: [DONE]\n\n")
				return
			}
		case <-tick.C:
			if !proxyWrite(w, ": bekleniyor\n\n") {
				return
			}
		case <-ctx.Done():
			fail("model timeout or cancellation")
			return
		}
	}
}
func proxyCompleted(body []byte) bool {
	var doc struct {
		Error   json.RawMessage `json:"error"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Finish string `json:"finish_reason"`
		} `json:"choices"`
	}
	if json.Unmarshal(body, &doc) != nil || len(doc.Error) > 0 && string(doc.Error) != "null" || len(doc.Choices) == 0 {
		return false
	}
	return doc.Choices[0].Message.Content != "" && doc.Choices[0].Finish == "stop"
}
func proxyWrite(w http.ResponseWriter, text string) bool {
	controller := http.NewResponseController(w)
	// Bound slow-reader writes independently of the upstream request timeout.
	// Tests/custom writers may not expose deadline controls; the native Go HTTP
	// server does. Each chunk/heartbeat renews this short write deadline.
	_ = controller.SetWriteDeadline(time.Now().Add(20 * time.Second))
	if _, err := io.WriteString(w, text); err != nil {
		return false
	}
	if err := controller.Flush(); err != nil && !errors.Is(err, http.ErrNotSupported) {
		return false
	}
	return true
}

func proxyFault(w http.ResponseWriter, message string) {
	body, _ := json.Marshal(map[string]any{"error": map[string]string{"message": message}})
	proxyWrite(w, "data: "+string(body)+"\n\n")
}

type proxyEvent struct {
	data string
	err  error
}

func scanProxyEvents(ctx context.Context, body io.Reader, out chan<- proxyEvent) {
	send := func(v proxyEvent) bool {
		select {
		case out <- v:
			return true
		case <-ctx.Done():
			return false
		}
	}
	scanner := bufio.NewScanner(io.LimitReader(body, (8<<20)+1))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	parts := []string{}
	total := 0
	for scanner.Scan() {
		line := scanner.Text()
		total += len(line) + 1
		if total > 8<<20 {
			send(proxyEvent{err: fmt.Errorf("stream bound")})
			return
		}
		if line == "" {
			if len(parts) > 0 {
				if !send(proxyEvent{data: strings.Join(parts, "\n")}) {
					return
				}
				parts = nil
			}
		} else if strings.HasPrefix(line, "data:") {
			parts = append(parts, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	send(proxyEvent{err: io.ErrUnexpectedEOF})
}

// Metadata is read-only and cannot reserve credit or start cloud resources.
func (p *ModelProxy) Models(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != "GET" {
		http.Error(w, "method", 405)
		return
	}
	key := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if key == r.Header.Get("Authorization") || !p.Store.authorizedCredential(key, time.Now()) {
		http.Error(w, "license", 401)
		return
	}
	if p.Upstream == "" {
		http.Error(w, "model upstream not configured", 503)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", p.Upstream+"/v1/models", nil)
	if p.Token != "" {
		req.Header.Set("Authorization", "Bearer "+p.Token)
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		http.Error(w, "model metadata unavailable", 502)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if err != nil || len(body) > 1<<20 || resp.StatusCode != 200 || !json.Valid(body) {
		http.Error(w, "model metadata invalid", 502)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	proxyWrite(w, string(body))
}
