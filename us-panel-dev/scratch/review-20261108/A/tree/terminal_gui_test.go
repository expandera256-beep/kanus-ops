package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTerminalGUIActionReturnsRealExitCode(t *testing.T) {
	s := newView("a100")
	s.root = t.TempDir()
	os.WriteFile(filepath.Join(s.root, "main.py"), []byte("print('gui-terminal-proof')\n"), 0600)
	mux := guiMux(&s)
	call := func(op string) guiSnap {
		t.Helper()
		var r *httptest.ResponseRecorder
		r = httptest.NewRecorder()
		if op != "" {
			raw, _ := json.Marshal(map[string]string{"op": op, "text": `["python3","main.py"]`})
			req := httptest.NewRequest("POST", "http://127.0.0.1:4477/api/act", bytes.NewReader(raw))
			mux.ServeHTTP(r, req)
		} else {
			mux.ServeHTTP(r, httptest.NewRequest("GET", "http://127.0.0.1:4477/api/state", nil))
		}
		if r.Code != 200 {
			t.Fatal(r.Code, r.Body.String())
		}
		var v guiSnap
		if err := json.Unmarshal(r.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	v := call("terminal-start")
	deadline := time.Now().Add(4 * time.Second)
	for v.Terminal.State == "running" && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
		v = call("")
	}
	if v.Terminal.State != "success" || v.Terminal.ExitCode != 0 || !strings.Contains(v.Terminal.Stdout, "gui-terminal-proof") {
		t.Fatal(v.Terminal)
	}
}
