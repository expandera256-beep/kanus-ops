package main

import (
	"os"
	"strings"
)

// Set by the release build. Environment variables cannot disable this gate.
var releaseChannel = "development"

func commercialRequired() bool { return releaseChannel == "commercial" }

func (s *session) commercialSessionReady() bool {
	return s.accountGateway != "" && s.accountGateway == strings.TrimRight(activeGateway(), "/") &&
		strings.HasPrefix(s.accountGateway, "https://") && !isLoopbackGateway(s.accountGateway) &&
		validateGateway(s.accountGateway) == nil && len(strings.TrimSpace(s.key)) >= 32 && s.key != localTunnelKey
}

// Release authority is fixed at build time. The hook release does not use
// Python accounts or ship a signing secret in the desktop executable.
var releaseAuthority = "accounts"
var embeddedLicenseURL string

func hookLicenseMode() bool { return releaseAuthority == "go-hook" }
func activeLicenseURL() string {
	if raw := strings.TrimSpace(os.Getenv("US_LICENSE_URL")); raw != "" {
		return strings.TrimRight(raw, "/")
	}
	return strings.TrimRight(strings.TrimSpace(embeddedLicenseURL), "/")
}
func hookTokenPresent(token string) bool {
	parts := strings.Split(strings.TrimSpace(token), ".")
	return len(token) <= 4096 && !strings.ContainsAny(token, "\r\n\t ") && len(parts) == 3 && parts[0] == "us1" && parts[1] != "" && parts[2] != ""
}
