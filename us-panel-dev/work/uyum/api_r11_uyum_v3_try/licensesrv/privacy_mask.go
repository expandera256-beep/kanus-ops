package licensesrv

import (
	"regexp"
	"strings"
)

// maskLicenseKey: "us1.…" + last 4 characters. Used for everything that is
// persisted (outbox, and therefore backups) or logged.
func maskLicenseKey(k string) string {
	r := []rune(k)
	if len(r) <= 8 {
		return "us1.…"
	}
	return "us1.…" + string(r[len(r)-4:])
}

func maskLicenseKeys(keys []string) []string {
	out := make([]string, len(keys))
	for i, k := range keys {
		out[i] = maskLicenseKey(k)
	}
	return out
}

var rawLicenseKeyRE = regexp.MustCompile(`us1\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`)

// scrubLicenseKeys masks every full license key in a mail body (legacy
// outbox entries written before round 11 stored full keys).
func scrubLicenseKeys(body string) string {
	return rawLicenseKeyRE.ReplaceAllStringFunc(body, maskLicenseKey)
}

// maskEmail: "d***@gmail.com".
func maskEmail(e string) string {
	e = strings.TrimSpace(e)
	at := strings.LastIndex(e, "@")
	if at < 1 {
		return "***"
	}
	r := []rune(e[:at])
	return string(r[0]) + "***" + e[at:]
}
