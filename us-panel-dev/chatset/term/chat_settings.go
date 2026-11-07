package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Chat settings (sohbet ayarları). Stored only in the user profile
// (UserConfigDir/us-panel/chat-settings.json, 0600, dir 0700). The file is
// never logged, never added to telemetry, crash reports or support bundles,
// and is removed by resetChatSettings (reset / uninstall). Defaults equal
// today's behaviour: nothing is added to the request.
type chatSettings struct {
	Temperature *float64 `json:"temperature,omitempty"`
	Length      string   `json:"length,omitempty"`
	Instruction string   `json:"instruction,omitempty"`
	Lang        string   `json:"lang,omitempty"`
}

const chatInstructionMax = 2000

var chatLengths = map[string]bool{"": true, "short": true, "medium": true, "long": true}
var chatLangs = map[string]bool{"": true, "tr": true, "en": true, "de": true, "ru": true}

var (
	chatSettingsMu  sync.Mutex
	chatSettingsCur chatSettings
	chatSettingsOK  bool
)

func chatSettingsPath() string {
	if p := strings.TrimSpace(os.Getenv("US_CHAT_SETTINGS_FILE")); p != "" {
		return p
	}
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, "us-panel", "chat-settings.json")
}

func (c chatSettings) isDefault() bool {
	return c.Temperature == nil && c.Length == "" && strings.TrimSpace(c.Instruction) == "" && c.Lang == ""
}

func (c chatSettings) validate() error {
	if c.Temperature != nil && (*c.Temperature < 0 || *c.Temperature > 1.5) {
		return errors.New("sıcaklık 0.0–1.5 aralığında olmalı")
	}
	if !chatLengths[c.Length] {
		return errors.New("geçersiz yanıt uzunluğu")
	}
	if !chatLangs[c.Lang] {
		return errors.New("geçersiz dil")
	}
	if len([]rune(c.Instruction)) > chatInstructionMax {
		return errors.New("sistem talimatı 2000 karakteri aşıyor")
	}
	return nil
}

func loadChatSettings() chatSettings {
	chatSettingsMu.Lock()
	defer chatSettingsMu.Unlock()
	if chatSettingsOK {
		return chatSettingsCur
	}
	chatSettingsOK = true
	chatSettingsCur = chatSettings{}
	if p := chatSettingsPath(); p != "" {
		if raw, err := os.ReadFile(p); err == nil {
			var c chatSettings
			if json.Unmarshal(raw, &c) == nil && c.validate() == nil {
				chatSettingsCur = c
			}
		}
	}
	return chatSettingsCur
}

func saveChatSettings(c chatSettings) error {
	if err := c.validate(); err != nil {
		return err
	}
	p := chatSettingsPath()
	if p == "" {
		return errors.New("kullanıcı profili bulunamadı")
	}
	if c.isDefault() {
		return resetChatSettings()
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return errors.New("ayar klasörü oluşturulamadı")
	}
	raw, _ := json.Marshal(c)
	tmp := p + ".tmp"
	_ = os.Remove(tmp)
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return errors.New("ayar dosyası yazılamadı")
	}
	_, werr := f.Write(raw)
	_ = f.Chmod(0o600)
	cerr := f.Close()
	if werr != nil || cerr != nil {
		_ = os.Remove(tmp)
		return errors.New("ayar dosyası yazılamadı")
	}
	if err := os.Rename(tmp, p); err != nil {
		_ = os.Remove(tmp)
		return errors.New("ayar dosyası yazılamadı")
	}
	chatSettingsMu.Lock()
	chatSettingsCur, chatSettingsOK = c, true
	chatSettingsMu.Unlock()
	return nil
}

// resetChatSettings deletes the stored file (reset / uninstall: "US.exe -reset-chat-settings").
func resetChatSettings() error {
	chatSettingsMu.Lock()
	chatSettingsCur, chatSettingsOK = chatSettings{}, true
	chatSettingsMu.Unlock()
	if p := chatSettingsPath(); p != "" {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return errors.New("ayar dosyası silinemedi")
		}
	}
	return nil
}

// applyChatSettingsPayload adds "us_chat" only for plain chat (modelSystem)
// and only when the user changed something; the hook enforces the bounds.
func applyChatSettingsPayload(payload map[string]any, system string) {
	if system != modelSystem {
		return
	}
	c := loadChatSettings()
	if c.isDefault() {
		return
	}
	payload["us_chat"] = c
}

func chatSettingsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if origin := r.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme != "http" || !strings.EqualFold(u.Host, r.Host) {
			http.Error(w, "Kaynak doğrulanamadı.", http.StatusForbidden)
			return
		}
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, loadChatSettings())
	case http.MethodPost:
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
		dec.DisallowUnknownFields()
		var c chatSettings
		if dec.Decode(&c) != nil {
			http.Error(w, "Ayarlar okunamadı.", http.StatusBadRequest)
			return
		}
		if err := saveChatSettings(c); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, loadChatSettings())
	case http.MethodDelete:
		if err := resetChatSettings(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, chatSettings{})
	default:
		http.Error(w, "yöntem yok", http.StatusMethodNotAllowed)
	}
}
