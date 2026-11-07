package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContextLimitsDiscoverSelectedModel(t *testing.T) {
	t.Setenv("US_CONTEXT_TOKENS", "")
	t.Setenv("US_OUTPUT_TOKENS", "")
	t.Setenv("US_MODEL_NAME", "specific-model")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected request")
		}
		fmt.Fprint(w, `{"data":[{"id":"wrong","max_model_len":100000},{"id":"specific-model","max_model_len":32768,"max_output_tokens":4096}]}`)
	}))
	defer srv.Close()
	old := doorList
	doorList = func() []string { return []string{srv.URL} }
	defer func() { doorList = old }()
	got := resolveContextLimits("a100", "test-key")
	if got.promptBytes <= promptWireBudget || got.outputTokens != 4096 || got.promptBytes+got.outputTokens+512 > 32768 {
		t.Fatalf("%+v", got)
	}
}

func TestContextLimitsUnknownModelUsesFallback(t *testing.T) {
	t.Setenv("US_CONTEXT_TOKENS", "")
	t.Setenv("US_MODEL_NAME", "selected")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[{"id":"other","max_model_len":100000}]}`)
	}))
	defer srv.Close()
	old := doorList
	doorList = func() []string { return []string{srv.URL} }
	defer func() { doorList = old }()
	if got := resolveContextLimits("a100", localTunnelKey); got != defaultContextLimits() {
		t.Fatalf("%+v", got)
	}
}

func TestContextLimitsReadsNestedServerMetadata(t *testing.T) {
	t.Setenv("US_CONTEXT_TOKENS", "")
	t.Setenv("US_OUTPUT_TOKENS", "")
	t.Setenv("US_MODEL_NAME", "deepseek-v4.1-flash")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[{"id":"deepseek-v4.1-flash","meta":{"n_ctx":8192}}]}`)
	}))
	defer srv.Close()
	old := doorList
	doorList = func() []string { return []string{srv.URL} }
	defer func() { doorList = old }()
	got := resolveContextLimits("a100", localTunnelKey)
	if got.promptBytes+got.outputTokens+512 != 8192 {
		t.Fatalf("%+v", got)
	}
}

func TestConfiguredContextKeepsOutputReservation(t *testing.T) {
	t.Setenv("US_CONTEXT_TOKENS", "32768")
	t.Setenv("US_OUTPUT_TOKENS", "4096")
	got := resolveContextLimits("a100", "test")
	if got.outputTokens != 4096 || got.promptBytes+got.outputTokens+512 != 32768 || got.initialBytes >= got.promptBytes {
		t.Fatalf("%+v", got)
	}
	if len(strings.Repeat("x", 10000)) >= got.initialBytes {
		t.Fatal("input should fit without summarization")
	}
}

func TestConfiguredLargePromptReachesModelWithoutSummary(t *testing.T) {
	t.Setenv("US_CONTEXT_TOKENS", "32768")
	t.Setenv("US_OUTPUT_TOKENS", "4096")
	s := newView("q8")
	s.root, s.key = t.TempDir(), "local-test"
	s.modelCh = make(chan modelResult, 1)
	text := strings.Repeat("özgün içerik ", 700) + "SON-İŞARET"
	calls := 0
	s.sendModel = func(_ workerID, _, wire string) modelResult {
		calls++
		if !strings.Contains(wire, text) || strings.Contains(wire, "--- PARÇA ---") {
			t.Error("prompt was lost or summarized")
		}
		return modelResult{ok: true, body: "tamam"}
	}
	s.askModel(text, intent{})
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
}
