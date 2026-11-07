package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// Login/register only: the server hashes with bcrypt and can be slow under
// load, so these two calls get a longer timeout (nginx proxy_read_timeout is
// 90s) and exactly one retry on a network error or timeout. An HTTP status
// (401, 429, 5xx, ...) is never retried. Other account calls keep 20s.
var (
	accountAuthTimeout    = accountAuthTimeoutDefault()
	accountAuthRetryDelay = 2 * time.Second
)

func accountAuthDo(url string, payload []byte) (*http.Response, error) {
	client := &http.Client{Timeout: accountAuthTimeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(accountAuthRetryDelay)
		}
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err == nil {
			return resp, nil // any HTTP status: no retry
		}
		lastErr = err
	}
	return nil, lastErr
}

func accountRequest(base, action, email, password, license string) (string, error) {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if err := validateGateway(base); err != nil {
		return "", err
	}
	if action != "register" && action != "login" {
		return "", errors.New("hesap işlemi geçersiz")
	}
	if len(email) > 254 || len(password) > 256 || len(license) > 4096 {
		return "", errors.New("hesap bilgisi çok uzun")
	}
	payload, _ := json.Marshal(map[string]string{"email": email, "password": password, "license": license})
	resp, err := accountAuthDo(base+"/account/"+action, payload)
	if err != nil {
		return "", errors.New("hesap sunucusuna ulaşılamadı")
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("hesap işlemi reddedildi")
	}
	if action == "register" {
		return "", nil
	}
	var v struct {
		Session string `json:"session"`
	}
	if json.Unmarshal(raw, &v) != nil || len(v.Session) < 32 || len(v.Session) > 256 || strings.ContainsAny(v.Session, "\r\n") {
		return "", errors.New("geçersiz hesap yanıtı")
	}
	return v.Session, nil
}

// Redeem a purchased pack or retrieve the balance from the authenticated
// gateway. The bearer token is never put in a URL or persisted by this call.
func accountCredits(base, action, session, license string) (int, error) {
	n, _, err := accountBalances(base, action, session, license)
	return n, err
}

func accountBalances(base, action, session, license string) (int, int, error) {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if err := validateGateway(base); err != nil {
		return 0, 0, err
	}
	if action != "balance" && action != "redeem" || len(session) < 32 || len(session) > 256 || len(license) > 4096 || strings.ContainsAny(session, "\r\n") {
		return 0, 0, errors.New("hesap bilgisi geçersiz")
	}
	payload, _ := json.Marshal(map[string]string{"license": license})
	req, err := http.NewRequest(http.MethodPost, base+"/account/"+action, bytes.NewReader(payload))
	if err != nil {
		return 0, 0, errors.New("hesap bağlantısı kurulamadı")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+session)
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, errors.New("hesap sunucusuna ulaşılamadı")
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, 0, errors.New("kredi işlemi reddedildi")
	}
	var v struct {
		Remaining *int `json:"remaining"`
		Seconds   int  `json:"seconds_remaining"`
	}
	if json.Unmarshal(raw, &v) != nil || v.Remaining == nil || v.Seconds < 0 || *v.Remaining < 0 || *v.Remaining > 100000000 {
		return 0, 0, errors.New("geçersiz bakiye yanıtı")
	}
	return *v.Remaining, v.Seconds, nil
}

type accountModelRequest struct {
	ID      string `json:"request_id"`
	State   string `json:"state"`
	Created int64  `json:"created"`
	Updated int64  `json:"updated"`
}

func accountRequests(base, session string) ([]accountModelRequest, error) {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if err := validateGateway(base); err != nil {
		return nil, err
	}
	if len(session) < 32 || len(session) > 256 || strings.ContainsAny(session, "\r\n") {
		return nil, errors.New("hesap bilgisi geçersiz")
	}
	req, err := http.NewRequest(http.MethodPost, base+"/account/requests", strings.NewReader("{}"))
	if err != nil {
		return nil, errors.New("hesap bağlantısı kurulamadı")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+session)
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("hesap sunucusuna ulaşılamadı")
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, (16<<10)+1))
	if err != nil || len(raw) > 16<<10 || resp.StatusCode != 200 {
		return nil, errors.New("istek geçmişi alınamadı")
	}
	var v struct {
		Requests []accountModelRequest `json:"requests"`
	}
	if json.Unmarshal(raw, &v) != nil || len(v.Requests) > 20 {
		return nil, errors.New("geçersiz istek geçmişi")
	}
	for _, item := range v.Requests {
		if len(item.ID) < 8 || len(item.ID) > 128 || item.Created < 0 || item.Updated < item.Created {
			return nil, errors.New("geçersiz istek kaydı")
		}
		for _, c := range item.ID {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("._-", c)) {
				return nil, errors.New("geçersiz istek kimliği")
			}
		}
		if item.State != "reserved" && item.State != "completed" && item.State != "refunded" {
			return nil, errors.New("geçersiz istek durumu")
		}
	}
	return v.Requests, nil
}

func accountAuthTimeoutDefault() time.Duration { return 90 * time.Second }
