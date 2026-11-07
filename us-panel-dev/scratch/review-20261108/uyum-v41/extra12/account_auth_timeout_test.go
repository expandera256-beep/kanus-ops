package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func withAuthTimeout(t *testing.T, timeout time.Duration) {
	old, oldDelay := accountAuthTimeout, accountAuthRetryDelay
	accountAuthTimeout, accountAuthRetryDelay = timeout, 10*time.Millisecond
	t.Cleanup(func() { accountAuthTimeout, accountAuthRetryDelay = old, oldDelay })
}

const okSession = `{"session":"abcdefghijklmnopqrstuvwxyz1234567890"}`

func TestAccountLoginSlowServerWithinBudgetSucceeds(t *testing.T) {
	withAuthTimeout(t, 900*time.Millisecond) // scaled 90s budget
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(500 * time.Millisecond) // > old 20s-equivalent ratio, < budget
		_, _ = w.Write([]byte(okSession))
	}))
	defer srv.Close()
	if tok, err := accountRequest(srv.URL, "login", "a@example.com", "long-password-123", ""); err != nil || tok == "" {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("calls=%d", atomic.LoadInt32(&calls))
	}
	if accountAuthTimeoutDefault() != 90*time.Second {
		t.Fatal("varsayılan login/register süresi 90 sn değil")
	}
}

func TestAccountLoginTimeoutRetriesExactlyOnce(t *testing.T) {
	withAuthTimeout(t, 150*time.Millisecond)
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			time.Sleep(400 * time.Millisecond)
			return
		}
		_, _ = w.Write([]byte(okSession))
	}))
	defer srv.Close()
	if _, err := accountRequest(srv.URL, "login", "a@example.com", "long-password-123", ""); err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("calls=%d", atomic.LoadInt32(&calls))
	}
	// Both attempts time out: still exactly 2 attempts, then error.
	atomic.StoreInt32(&calls, -100)
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(400 * time.Millisecond)
	}))
	defer slow.Close()
	if _, err := accountRequest(slow.URL, "register", "a@example.com", "long-password-123", ""); err == nil {
		t.Fatal("hata bekleniyordu")
	}
	if atomic.LoadInt32(&calls) != -98 {
		t.Fatalf("deneme sayısı %d", atomic.LoadInt32(&calls)+100)
	}
}

func TestAccountLoginNetworkErrorRetriesOnce(t *testing.T) {
	withAuthTimeout(t, time.Second)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var calls int32
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			atomic.AddInt32(&calls, 1)
			_ = c.Close() // reset: network error
		}
	}()
	defer ln.Close()
	if _, err := accountRequest("http://"+ln.Addr().String(), "login", "a@example.com", "long-password-123", ""); err == nil {
		t.Fatal("hata bekleniyordu")
	}
	time.Sleep(50 * time.Millisecond)
	if n := atomic.LoadInt32(&calls); n != 2 {
		t.Fatalf("bağlantı sayısı %d", n)
	}
}

func TestAccountLoginNoRetryOnHTTPStatus(t *testing.T) {
	withAuthTimeout(t, time.Second)
	for _, code := range []int{401, 429, 500, 503} {
		var calls int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(code)
		}))
		if _, err := accountRequest(srv.URL, "login", "a@example.com", "long-password-123", ""); err == nil {
			t.Fatalf("%d: hata bekleniyordu", code)
		}
		srv.Close()
		if atomic.LoadInt32(&calls) != 1 {
			t.Fatalf("%d: %d deneme", code, atomic.LoadInt32(&calls))
		}
	}
}
