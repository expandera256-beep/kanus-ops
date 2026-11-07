package main

// S2 saldırı regresyonu — ortak yardımcılar.
// UYUM NOKTASI (adapter): newView, session.root/browse, guiMux. Panelin
// dışa açık girişi HTTP /api/act ve /api/state'tir; PoC'ler bunları kullanır.
// .9'da bu üç sembol değişirse yalnız bu dosyadaki s2regServer güncellenir.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

type s2regSnap struct {
	Notice     string `json:"notice"`
	TestReport string `json:"testReport"`
	Terminal   struct {
		State    string   `json:"state"`
		Argv     []string `json:"argv"`
		Stdout   string   `json:"stdout"`
		Stderr   string   `json:"stderr"`
		ExitCode int      `json:"exitCode"`
	} `json:"terminal"`
}

func s2regServer(t *testing.T, root string) *httptest.Server {
	t.Helper()
	s := newView("s2reg")
	s.root = root
	s.browse = root
	srv := httptest.NewServer(guiMux(&s))
	t.Cleanup(func() { s.stopProgram(); srv.Close() })
	return srv
}

func s2regAct(t *testing.T, srv *httptest.Server, body any) s2regSnap {
	t.Helper()
	raw, _ := json.Marshal(body)
	resp, err := http.Post(srv.URL+"/api/act", "application/json", strings.NewReader(string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var snap s2regSnap
	_ = json.NewDecoder(resp.Body).Decode(&snap)
	return snap
}

func s2regState(t *testing.T, srv *httptest.Server) s2regSnap {
	resp, err := http.Get(srv.URL + "/api/state")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var snap s2regSnap
	_ = json.NewDecoder(resp.Body).Decode(&snap)
	return snap
}

// s2regTerminal starts a terminal command via the public API and waits for a final state.
func s2regTerminal(t *testing.T, srv *httptest.Server, argv []string, seconds int, wait time.Duration) (s2regSnap, bool) {
	raw, _ := json.Marshal(argv)
	start := s2regAct(t, srv, map[string]any{"op": "terminal-start", "text": string(raw), "n": seconds})
	if start.Terminal.State != "running" && !strings.Contains(start.Notice, "başladı") {
		return start, false // rejected
	}
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		st := s2regState(t, srv)
		if st.Terminal.State != "running" && st.Terminal.State != "" {
			return st, true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return s2regState(t, srv), true
}

func s2regResult(t *testing.T, id string, pass bool, detail string) {
	res := "FAIL"
	if pass {
		res = "PASS"
	}
	detail = strings.ReplaceAll(detail, "\n", " ")
	if len(detail) > 300 {
		detail = detail[:300]
	}
	t.Logf("S2REG|%s|%s|%s", id, res, detail)
}

func s2regBlocked(t *testing.T, id, why string) { t.Logf("S2REG|%s|BLOCKED|%s", id, why) }

func s2regAlive(pid int) bool { return pid > 0 && syscall.Kill(pid, 0) == nil }

func s2regWrite(t *testing.T, path, body string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
}

func s2regGoProject(t *testing.T, root string) {
	s2regWrite(t, root+"/go.mod", "module p\n\ngo 1.21\n", 0o644)
	s2regWrite(t, root+"/main.go", "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"inside\") }\n", 0o644)
	s2regWrite(t, root+"/main_test.go", "package main\n\nimport \"testing\"\n\nfunc TestX(t *testing.T) {}\n", 0o644)
}
