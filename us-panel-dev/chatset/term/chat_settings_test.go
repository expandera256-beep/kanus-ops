package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func csTempFile(t *testing.T) string {
	p := filepath.Join(t.TempDir(), "prof", "us-panel", "chat-settings.json")
	t.Setenv("US_CHAT_SETTINGS_FILE", p)
	chatSettingsMu.Lock()
	chatSettingsOK = false
	chatSettingsMu.Unlock()
	t.Cleanup(func() {
		chatSettingsMu.Lock()
		chatSettingsOK = false
		chatSettingsCur = chatSettings{}
		chatSettingsMu.Unlock()
	})
	return p
}

func TestChatSettingsClientGoldenDefaults(t *testing.T) {
	csTempFile(t)
	a := map[string]any{"model": "m", "max_tokens": 4096, "messages": []map[string]string{{"role": "system", "content": modelSystem}}}
	b := map[string]any{"model": "m", "max_tokens": 4096, "messages": []map[string]string{{"role": "system", "content": modelSystem}}}
	applyChatSettingsPayload(b, modelSystem)
	ra, _ := json.Marshal(a)
	rb, _ := json.Marshal(b)
	if string(ra) != string(rb) {
		t.Fatalf("defaults must be byte-identical: %s vs %s", ra, rb)
	}
}

func TestChatSettingsClientStoragePermsAndReset(t *testing.T) {
	p := csTempFile(t)
	temp := 0.7
	if err := saveChatSettings(chatSettings{Temperature: &temp, Length: "short", Lang: "de", Instruction: "kısa yaz"}); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if fi.Mode().Perm() != 0o600 {
			t.Fatal("file perm", fi.Mode().Perm())
		}
		if di, _ := os.Stat(filepath.Dir(p)); di.Mode().Perm() != 0o700 {
			t.Fatal("dir perm", di.Mode().Perm())
		}
	}
	payload := map[string]any{}
	applyChatSettingsPayload(payload, modelSystem)
	if payload["us_chat"] == nil {
		t.Fatal("chat payload missing settings")
	}
	other := map[string]any{}
	applyChatSettingsPayload(other, programSystem)
	if len(other) != 0 {
		t.Fatal("settings must not touch file-operation/agent calls")
	}
	if err := resetChatSettings(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("reset must delete the file")
	}
	if !loadChatSettings().isDefault() {
		t.Fatal("reset must restore defaults")
	}
}

func TestChatSettingsClientHandlerValidation(t *testing.T) {
	csTempFile(t)
	secret := "GIZLI-" + strings.Repeat("x", 2100)
	for _, body := range []string{`{"lang":"fr"}`, `{"length":"huge"}`, `{"temperature":3}`, `{"model":"x"}`, `{"instruction":"` + secret + `"}`} {
		w := httptest.NewRecorder()
		chatSettingsHandler(w, httptest.NewRequest("POST", "/api/chat-settings", strings.NewReader(body)))
		if w.Code != http.StatusBadRequest || strings.Contains(w.Body.String(), "GIZLI") {
			t.Fatal(body[:10], w.Code, w.Body.String()[:40])
		}
	}
	r := httptest.NewRequest("POST", "/api/chat-settings", strings.NewReader(`{}`))
	r.Header.Set("Origin", "http://evil.example")
	w := httptest.NewRecorder()
	chatSettingsHandler(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatal("cross-origin must be refused", w.Code)
	}
}

// UYUM-4: only chat_settings.go and the single model.go call site know the file;
// no telemetry/crash/support/journal path references it.
func TestChatSettingsClientNotInOtherPaths(t *testing.T) {
	files, _ := filepath.Glob("*.go")
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || f == "chat_settings.go" {
			continue
		}
		raw, _ := os.ReadFile(f)
		s := string(raw)
		for _, needle := range []string{"chat-settings.json", "chatSettingsCur", "loadChatSettings", ".Instruction"} {
			if strings.Contains(s, needle) {
				t.Fatalf("%s references %s", f, needle)
			}
		}
	}
}

func TestChatSettingsPrivacyNotice(t *testing.T) {
	raw, _ := os.ReadFile("gui.html")
	if strings.Count(string(raw), "sistem talimatı model sağlayıcıya iletilir") < 4 {
		t.Fatal("privacy line missing in TR/EN/DE/RU")
	}
	if strings.Contains(string(raw), "localStorage") {
		t.Fatal("instruction must not be stored in browser storage")
	}
}
