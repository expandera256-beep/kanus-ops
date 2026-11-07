package licensesrv

import (
	"net"
	"net/http"
	"strings"
	"time"
)

const loginCookieName = "__Host-kanus_login"

// Keep the existing trusted-loopback proxy boundary. Never trust forwarded
// headers from arbitrary remote clients. IPv4-mapped IPv6 is canonical IPv4.
func accountClientSource(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		if v := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Real-IP"))); v != nil {
			ip = v
		}
	}
	if ip == nil {
		return "invalid"
	}
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	return ip.Mask(net.CIDRMask(64, 128)).String() + "/64"
}
func loginPairKey(email, source string, trusted bool) string {
	prefix := "unknown:"
	if trusted {
		prefix = "trusted:"
	}
	return prefix + digest(email+"\x00"+source)
}
func (h *AccountHandler) trustedLogin(r *http.Request, email string, now time.Time) bool {
	tokens := []string{}
	if v := r.Header.Get("Authorization"); strings.HasPrefix(v, "Bearer ") {
		tokens = append(tokens, strings.TrimPrefix(v, "Bearer "))
	}
	if c, err := r.Cookie(loginCookieName); err == nil {
		tokens = append(tokens, c.Value)
	}
	h.Store.mu.Lock()
	defer h.Store.mu.Unlock()
	for _, token := range tokens {
		if a := h.Store.accountForTokenLocked(token, now); a != nil && a.Email == email {
			return true
		}
	}
	return false
}
func (h *AccountHandler) loginBlocked(email, source string, trusted bool, now time.Time) (bool, int) {
	b, left := h.acctLimit.blocked(loginPairKey(email, source, trusted), now)
	if !trusted {
		if gb, gl := h.globalLogin.blocked(digest(email), now); gb {
			b = true
			if gl > left {
				left = gl
			}
		}
	}
	return b, left
}
func (h *AccountHandler) loginFailed(email, source string, trusted bool, now time.Time) {
	h.acctLimit.fail(loginPairKey(email, source, trusted), now)
	h.globalLogin.fail(digest(email), now)
}

// Same additive envelope as api-0003, without depending on that patch's helper.
// No account, source or credential detail is returned.
func loginRateError(w http.ResponseWriter, left int) {
	rid := randID()
	w.Header().Set("Retry-After", retryAfterSeconds(left))
	w.Header().Set("X-Request-Id", rid)
	accountJSON(w, 429, map[string]string{"error": "account request rejected", "code": "rate_limited", "message": "account request rejected", "request_id": rid})
}
func (h *AccountHandler) rememberLogin(w http.ResponseWriter, token string, now time.Time) {
	http.SetCookie(w, &http.Cookie{Name: loginCookieName, Value: token, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: int(h.TTL / time.Second), Expires: now.Add(h.TTL)})
}
