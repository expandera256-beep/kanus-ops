package privacy

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestV41OriginParserAndAllowlist(t *testing.T) {
	valid := []string{"https://api.example", "https://api.example:1", "https://api.example:65535", "https://192.0.2.1:443", "https://[2001:db8::1]:443", "http://localhost:1", "http://localhost:65535", "http://127.0.0.1:3000", "http://[::1]:3000", "https://xn--bcher-kva.example"}
	invalid := []string{"http://localhost:65536", "http://[::1]:65536", "https://api.example:65536", "http://localhost:", "https://api.example:", "https://api.example:0", "https://api.example:-1", "https://api.example:+1", "https://api.example:abc", "https://api.example:00080", "https://api.example:999999999999999999999999", "https://", "https://-bad.example", "https://bad-.example", "https://bad..example", "https://bad_name.example", "https://999.999.999.999", "https://[127.0.0.1]", "https://[fe80::1%25eth0]", "http://public.example:80", "http://localhost.evil:80", "ftp://api.example:443", "https://name@api.example", "https://api.example/path", "https://api.example/", "https://api.example?", "https://api.example#", "https://api.example?x=1", "https://api.example#fragment", "https://" + strings.Repeat("a", 64) + ".example"}
	for _, raw := range valid {
		t.Run("valid_"+raw, func(t *testing.T) {
			if _, err := parsePrivacyOrigin(raw); err != nil {
				t.Fatal("valid origin rejected")
			}
			allowed := loadAllowedOrigins(raw)
			if len(allowed) != 1 {
				t.Fatal("valid entry not loaded")
			}
			r := httptest.NewRequest("POST", "/missing", nil)
			r.Host = "unrelated.example"
			r.Header.Set("Origin", raw)
			if !originAllowed(r, allowed) {
				t.Fatal("valid exact match rejected")
			}
		})
	}
	for _, raw := range invalid {
		t.Run("invalid_"+raw, func(t *testing.T) {
			if _, err := parsePrivacyOrigin(raw); err == nil {
				t.Fatal("invalid origin parsed")
			}
			allowed := loadAllowedOrigins(raw)
			if len(allowed) != 0 {
				t.Fatal("invalid config entry loaded")
			}
			// Even a matching Host and manually injected allowlist cannot bypass parser.
			r := httptest.NewRequest("POST", "/missing", nil)
			r.Host = strings.TrimPrefix(raw, "https://")
			r.Header.Set("Origin", raw)
			if originAllowed(r, map[string]struct{}{strings.ToLower(raw): {}}) {
				t.Fatal("incoming invalid origin accepted")
			}
		})
	}
}

func TestV41AllowlistLoadLogsNoRawValueAndIsSnapshot(t *testing.T) {
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)
	const bad = "https://synthetic-private-host.example:65536"
	t.Setenv("US_PRIVACY_ALLOWED_ORIGINS", bad+",http://localhost:3000")
	h := NewHandlerFromEnv(nil, nil)
	if len(h.allowedOrigins) != 1 {
		t.Fatal("invalid entry retained")
	}
	if !strings.Contains(buf.String(), "invalid entry ignored") || strings.Contains(buf.String(), bad) || strings.Contains(buf.String(), "synthetic-private-host") {
		t.Fatal("config warning missing or raw config leaked")
	}
	before := buf.Len()
	// HTTP route consumes validated startup config. An env edit needs reload/restart.
	t.Setenv("US_PRIVACY_ALLOWED_ORIGINS", "http://localhost:3001")
	for _, tt := range []struct {
		o    string
		want int
	}{{"http://localhost:3000", 404}, {"http://localhost:3001", 403}, {bad, 403}} {
		r := httptest.NewRequest("POST", "/missing", nil)
		r.Host = "api.example"
		r.Header.Set("Origin", tt.o)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tt.want {
			t.Fatalf("got %d want %d", w.Code, tt.want)
		}
	}
	if buf.Len() != before {
		t.Fatal("request unexpectedly reloaded/logged config")
	}
}

func TestV41UnavailableRoutesAndDisabledMount(t *testing.T) {
	// A configured store failure takes priority over erase flag-off: capability
	// unavailable is distinct from an intentional disabled route.
	m := http.NewServeMux()
	Mount(m, nil, MountOptions{PrivacyUnavailable: true, EraseEnabled: false})
	for _, path := range []string{"/account/privacy/export", "/account/privacy/preferences", "/account/privacy/erase/request", "/account/privacy/erase/confirm"} {
		w := httptest.NewRecorder()
		m.ServeHTTP(w, httptest.NewRequest("POST", path, nil))
		if w.Code != 503 || w.Header().Get("Retry-After") != "60" || w.Body.String() != "{\"error\":\"privacy unavailable\"}\n" {
			t.Fatalf("%s unexpected unavailable response", path)
		}
	}
}
