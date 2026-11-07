package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Input is bounded in bytes, conservatively, rather than pretending to know
// the server's tokenizer. The server is always the final authority.
type contextLimits struct { promptBytes, initialBytes, outputTokens int }

func defaultContextLimits() contextLimits {
	return contextLimits{promptWireBudget, initialWireBudget, 2048}
}

func configuredContextLimit() int {
	n, err := strconv.Atoi(strings.TrimSpace(os.Getenv("US_CONTEXT_TOKENS")))
	if err != nil || n < 2048 || n > 8388608 { return 0 }
	return n
}

func limitsForContext(n int) contextLimits {
	if n < 2048 || n > 8388608 { return defaultContextLimits() }
	output := n / 4
	if output > 8192 { output = 8192 }
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv("US_OUTPUT_TOKENS"))); err == nil && v >= 128 && v <= n/2 { output = v }
	input := n - output - 512
	if input < 1024 { return defaultContextLimits() }
	initial := input - 600
	if initial < 256 { return defaultContextLimits() }
	return contextLimits{input, initial, output}
}

func selectedModel(worker workerID) string {
	if custom := strings.TrimSpace(os.Getenv("US_MODEL_NAME")); custom != "" { return custom }
	switch worker {
	case "q5": return "usta-q5"
	case "flash": return "flash"
	case "a100": return "a100"
	default: return "usta"
	}
}

// Only use metadata for the exact selected model, returned by the same
// endpoint as completions. Missing metadata never causes a guessed increase.
func resolveContextLimits(worker workerID, key string) contextLimits {
	if n := configuredContextLimit(); n > 0 { return limitsForContext(n) }
	// A fallback gateway may host a different model; never apply one gateway's
	// advertised budget to another gateway.
	if len(doorList()) != 1 { return defaultContextLimits() }
	client := &http.Client{Timeout: 2 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	for _, base := range doorList() {
		url := strings.TrimRight(base, "/") + "/v1/models"
		if validateEndpoint(url) != nil { continue }
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil { continue }
		if key != "" && key != localTunnelKey { req.Header.Set("Authorization", "Bearer " + key) }
		resp, err := client.Do(req)
		if err != nil { continue }
		data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if err != nil || resp.StatusCode != http.StatusOK { continue }
		var listing struct { Data []struct { ID string `json:"id"`; MaxModelLen int `json:"max_model_len"`; ContextLength int `json:"context_length"`; ContextWindow int `json:"context_window"`; MaxOutputTokens int `json:"max_output_tokens"`; Meta struct { NCtx int `json:"n_ctx"` } `json:"meta"` } `json:"data"` }
		if json.Unmarshal(data, &listing) != nil { continue }
		for _, model := range listing.Data {
			if model.ID != selectedModel(worker) { continue }
			n := model.MaxModelLen
			if n == 0 { n = model.ContextLength }; if n == 0 { n = model.ContextWindow }
			if n == 0 { n = model.Meta.NCtx }
			if n < 2048 || n > 8388608 { break }
			limits := limitsForContext(n)
			if model.MaxOutputTokens > 0 && limits.outputTokens > model.MaxOutputTokens { limits.outputTokens = model.MaxOutputTokens }
			return limits
		}
	}
	return defaultContextLimits()
}
