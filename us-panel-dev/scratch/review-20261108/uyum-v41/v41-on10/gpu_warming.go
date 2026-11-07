package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// gpuWarmingMessage turns the hook's 503 "GPU warming" reply into a user
// message. No credit was reserved by the server for this request.
func gpuWarmingMessage(resp *http.Response) (string, bool) {
	if resp == nil || resp.StatusCode != http.StatusServiceUnavailable || resp.Header.Get("X-US-GPU-State") == "" {
		return "", false
	}
	var body struct {
		State      string `json:"state"`
		Message    string `json:"message"`
		RetryAfter int    `json:"retry_after"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&body)
	msg := strings.TrimSpace(body.Message)
	if msg == "" || len(msg) > 300 {
		msg = "GPU ısınıyor."
	}
	retry := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if retry == "" && body.RetryAfter > 0 {
		retry = fmt.Sprint(body.RetryAfter)
	}
	if body.State == "failed" {
		return msg + " GPU başlatılamadı; biraz sonra tekrar deneyin. Kredi düşülmedi.", true
	}
	if retry != "" {
		msg += " Yaklaşık " + retry + " sn sonra tekrar deneyin."
	}
	return msg + " Kredi düşülmedi.", true
}
