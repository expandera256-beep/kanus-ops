package licensesrv

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const DefaultGPUWakePinFile = "/etc/us-model/wake.pin"

// loadPins reads a pin file: SPKI sha256 pins (hex or base64, optional
// "sha256/" prefix), one per line, or PEM certificate(s) whose SPKI is
// pinned. Several entries allow rotation. Re-read on every call.
func loadPins(path string) ([][32]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil || len(raw) > 64<<10 {
		return nil, errors.New("pin_unreadable")
	}
	var pins [][32]byte
	if bytes.Contains(raw, []byte("-----BEGIN")) {
		rest := raw
		for {
			var b *pem.Block
			b, rest = pem.Decode(rest)
			if b == nil {
				break
			}
			if b.Type != "CERTIFICATE" {
				continue
			}
			c, err := x509.ParseCertificate(b.Bytes)
			if err != nil {
				return nil, errors.New("pin_invalid")
			}
			pins = append(pins, sha256.Sum256(c.RawSubjectPublicKeyInfo))
		}
	} else {
		for _, line := range strings.Split(string(raw), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			line = strings.TrimPrefix(strings.TrimPrefix(line, "sha256/"), "sha256:")
			var b []byte
			if len(line) == 64 {
				b, err = hex.DecodeString(line)
			} else {
				b, err = base64.StdEncoding.DecodeString(line)
				if err != nil {
					b, err = base64.RawStdEncoding.DecodeString(line)
				}
			}
			if err != nil || len(b) != 32 {
				return nil, errors.New("pin_invalid")
			}
			var p [32]byte
			copy(p[:], b)
			pins = append(pins, p)
		}
	}
	if len(pins) == 0 {
		return nil, errors.New("pin_invalid")
	}
	return pins, nil
}

// pinnedTLS: Go's CA check is skipped ONLY because VerifyConnection below
// enforces both (1) leaf SPKI sha256 ∈ pins and (2) the leaf matches the
// URL host (DNS name or IP SAN). Any failure aborts the handshake.
func pinnedTLS(host string, pins [][32]byte) *tls.Config {
	return &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ServerName:         host,
		InsecureSkipVerify: true, // replaced by VerifyConnection (pin + SAN)
		VerifyConnection: func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) == 0 {
				return errors.New("no peer certificate")
			}
			leaf := cs.PeerCertificates[0]
			sum := sha256.Sum256(leaf.RawSubjectPublicKeyInfo)
			ok := false
			for _, p := range pins {
				if p == sum {
					ok = true
				}
			}
			if !ok {
				return errors.New("certificate pin mismatch")
			}
			if err := leaf.VerifyHostname(host); err != nil {
				return errors.New("certificate host/IP SAN mismatch")
			}
			now := time.Now()
			if now.Before(leaf.NotBefore) || now.After(leaf.NotAfter) {
				return errors.New("certificate expired or not yet valid")
			}
			return nil
		},
	}
}

// client returns the HTTP client for one control call. Redirects are never
// followed. With a pin file: pinned TLS; without: system CA verification.
func (g *GPUWaker) client(rawURL string) (*http.Client, error) {
	noRedirect := func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	if g.PinFile == "" {
		c := g.Client
		if c == nil {
			c = &http.Client{Timeout: 20 * time.Second}
		}
		cc := *c
		cc.CheckRedirect = noRedirect
		return &cc, nil
	}
	pins, err := loadPins(g.PinFile)
	if err != nil {
		return nil, err // fail closed: no wake, no status
	}
	host := rawURL
	if i := strings.Index(host, "://"); i >= 0 {
		host = host[i+3:]
	}
	if i := strings.IndexAny(host, "/?#"); i >= 0 {
		host = host[:i]
	}
	if h, _, e := net.SplitHostPort(host); e == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	tr := &http.Transport{TLSClientConfig: pinnedTLS(host, pins), DisableKeepAlives: true, TLSHandshakeTimeout: 10 * time.Second, Proxy: nil}
	return &http.Client{Timeout: 20 * time.Second, Transport: tr, CheckRedirect: noRedirect}, nil
}

// pinFileFromEnv: explicit US_GPU_WAKE_PIN_FILE (must exist at call time,
// else fail closed); unset → DefaultGPUWakePinFile if present, else "" (CA).
func pinFileFromEnv() string {
	if v, set := os.LookupEnv("US_GPU_WAKE_PIN_FILE"); set {
		return strings.TrimSpace(v)
	}
	if _, err := os.Stat(DefaultGPUWakePinFile); err == nil {
		return DefaultGPUWakePinFile
	}
	return ""
}
