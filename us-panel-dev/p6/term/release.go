package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

const buildVersion = "2026.11.06.11"

// An occupied port must not silently open an older panel after an EXE update.
func currentPanelVersion(base string) bool {
	client := http.Client{Timeout: 800 * time.Millisecond, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Get(strings.TrimRight(base, "/") + "/api/version")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var value struct {
		Version string `json:"version"`
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1024)).Decode(&value) == nil && value.Version == buildVersion
}
