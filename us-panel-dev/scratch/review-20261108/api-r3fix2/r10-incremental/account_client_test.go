package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAccountClientLoginAndGatewayBinding(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/account/login" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]string
		if json.NewDecoder(r.Body).Decode(&body) != nil || body["email"] != "a@example.com" || body["password"] != "long-password-123" {
			t.Error("wrong credentials")
		}
		_, _ = w.Write([]byte(`{"session":"abcdefghijklmnopqrstuvwxyz1234567890"}`))
	}))
	defer srv.Close()
	token, err := accountRequest(srv.URL, "login", "a@example.com", "long-password-123", "")
	if err != nil || len(token) < 32 {
		t.Fatalf("%q %v", token, err)
	}
	t.Setenv("US_GATEWAY_URL", srv.URL)
	s := newView("q8")
	s.key = token
	s.accountGateway = srv.URL
	if s.modelKey() != token || s.licenseBlock(time.Now()) != "" {
		t.Fatal("account not recognized")
	}
	t.Setenv("US_GATEWAY_URL", "https://different.example")
	if s.accountGateway == activeGateway() {
		t.Fatal("changed endpoint")
	}
}

func TestAccountClientRejectsRemoteCleartext(t *testing.T) {
	if _, err := accountRequest("http://example.com", "login", "a@example.com", "long-password-123", ""); err == nil {
		t.Fatal("cleartext accepted")
	}
}

func TestAccountCreditsUsesBearerAndValidatesBalance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/account/redeem" || r.Header.Get("Authorization") != "Bearer abcdefghijklmnopqrstuvwxyz1234567890" {
			t.Error("wrong request")
		}
		var body map[string]string
		if json.NewDecoder(r.Body).Decode(&body) != nil || body["license"] != "package" {
			t.Error("wrong body")
		}
		_, _ = w.Write([]byte(`{"remaining":5}`))
	}))
	defer srv.Close()
	remaining, err := accountCredits(srv.URL, "redeem", "abcdefghijklmnopqrstuvwxyz1234567890", "package")
	if err != nil || remaining != 5 {
		t.Fatalf("remaining=%d error=%v", remaining, err)
	}
}

func TestAccountRequestsUsesBearerAndRejectsInvalidMetadata(t *testing.T) {
	for _, state := range []string{"reserved", "completed", "refunded", "unknown"} {
		t.Run(state, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/account/requests" || r.Header.Get("Authorization") != "Bearer abcdefghijklmnopqrstuvwxyz1234567890" {
					t.Error("wrong request")
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"requests": []accountModelRequest{{ID: "request-0001", State: state, Created: 1, Updated: 2}}})
			}))
			defer srv.Close()
			items, err := accountRequests(srv.URL, "abcdefghijklmnopqrstuvwxyz1234567890")
			if state == "unknown" {
				if err == nil {
					t.Fatal("invalid state accepted")
				}
				return
			}
			if err != nil || len(items) != 1 || items[0].State != state {
				t.Fatalf("items=%+v error=%v", items, err)
			}
		})
	}
}
