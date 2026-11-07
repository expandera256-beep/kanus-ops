package licensesrv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func loginFromIP(h *AccountHandler, ip, email, pw string) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(map[string]string{"email": email, "password": pw})
	r := httptest.NewRequest("POST", "/account/login", strings.NewReader(string(raw)))
	r.RemoteAddr = "127.0.0.1:5555" // behind nginx: X-Real-IP is trusted from loopback
	r.Header.Set("X-Real-IP", ip)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestAccountLimiterRotatingIPs(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	h, _ := setup(t)
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "victim@example.test", "")
	bodies := map[string]string{}
	for _, email := range []string{"victim@example.test", "ghost@example.test"} {
		var last *httptest.ResponseRecorder
		for i := 1; i <= 11; i++ {
			last = loginFromIP(a, fmt.Sprintf("203.0.%d.%d", len(email), i), "  "+strings.ToUpper(email[:1])+email[1:]+" ", "wrong-password-xyz")
			if i <= 10 && last.Code != 401 {
				t.Fatalf("%s deneme %d: %d", email, i, last.Code)
			}
		}
		if last.Code != 429 || last.Header().Get("Retry-After") == "" || last.Header().Get(hdrQueueDepth) == "" || last.Header().Get(hdrQueueLimit) == "" {
			t.Fatalf("%s 11. deneme: %d %v", email, last.Code, last.Header())
		}
		bodies[email] = last.Body.String()
		// Correct password is also refused while locked.
		if email == "victim@example.test" {
			if w := loginFromIP(a, "198.18.0.1", email, accountPassword); w.Code != 429 {
				t.Fatalf("kilitliyken doğru şifre: %d", w.Code)
			}
		}
	}
	if bodies["victim@example.test"] != bodies["ghost@example.test"] {
		t.Fatalf("kayıtlı/kayıtsız gövdeler farklı: %q %q", bodies["victim@example.test"], bodies["ghost@example.test"])
	}
	accountRegister(t, a, "other@example.test", "")
	if w := loginFromIP(a, "198.18.0.2", "other@example.test", accountPassword); w.Code != 200 {
		t.Fatalf("farklı e-posta etkilendi: %d", w.Code)
	}
	if strings.Contains(buf.String(), "victim@example.test") || strings.Contains(buf.String(), "ghost@example.test") {
		t.Fatalf("ham e-posta loglandı: %s", buf.String())
	}
}

func TestAccountLimiterSuccessResetsAndWindow(t *testing.T) {
	t.Setenv("US_LOGIN_ACCOUNT_LIMIT", "3")
	h, _ := setup(t)
	a := accountTestHandler(t, h.Store)
	accountRegister(t, a, "u@example.test", "")
	for i := 0; i < 2; i++ {
		loginFromIP(a, fmt.Sprintf("203.0.113.%d", i+1), "u@example.test", "wrong-password-xyz")
	}
	if w := loginFromIP(a, "203.0.113.9", "u@example.test", accountPassword); w.Code != 200 {
		t.Fatal(w.Code)
	}
	for i := 0; i < 2; i++ {
		if w := loginFromIP(a, fmt.Sprintf("203.0.113.%d", i+20), "u@example.test", "wrong-password-xyz"); w.Code != 401 {
			t.Fatalf("başarı sayacı sıfırlamadı: %d", w.Code)
		}
	}
	l := a.acctLimit
	now := time.Now()
	l.fail("u@example.test", now)
	if b, _ := l.blocked("u@example.test", now); !b {
		t.Fatal("limit uygulanmadı")
	}
	if b, _ := l.blocked("u@example.test", now.Add(l.window+time.Second)); b {
		t.Fatal("pencere bitince kilit kalktı olmalı")
	}
}

func TestAccountLimiterMemoryBound(t *testing.T) {
	l := newAccountLimiter()
	l.maxKeys = 100
	now := time.Now()
	for i := 0; i < 1000; i++ {
		l.fail(fmt.Sprintf("u%d@example.test", i), now.Add(time.Duration(i)*time.Millisecond))
	}
	if n := l.size(); n > 100 {
		t.Fatalf("anahtar sınırı aşıldı: %d", n)
	}
}
