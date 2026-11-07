package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

// Public service URL, set by the release build. Contains no credentials.
var embeddedGatewayURL string

func activeGateway() string {
	if base := strings.TrimSpace(os.Getenv("US_GATEWAY_URL")); base != "" {
		return strings.TrimRight(base, "/")
	}
	if base := strings.TrimSpace(embeddedGatewayURL); base != "" {
		return strings.TrimRight(base, "/")
	}
	if commercialRequired() {
		return ""
	}
	return door
}
func validateGateway(raw string) error {
	if commercialRequired() && (!strings.HasPrefix(raw, "https://") || isLoopbackGateway(raw)) {
		return fmt.Errorf("Ticari sürüm için uzak HTTPS hesap geçidi gerekir.")
	}
	u, err := url.Parse(raw)
	if err != nil || u == nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return fmt.Errorf("Geçerli bir kapı adresi girin (yalnızca sunucu ve isteğe bağlı port).")
	}
	return validateEndpoint(raw)
}
func validateEndpoint(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u == nil || u.Hostname() == "" || u.User != nil {
		return fmt.Errorf("Model adresi geçersiz.")
	}
	if u.Scheme == "https" {
		return nil
	}
	host := u.Hostname()
	ip := net.ParseIP(host)
	if u.Scheme == "http" && (strings.EqualFold(host, "localhost") || (ip != nil && ip.IsLoopback())) {
		return nil
	}
	return fmt.Errorf("Uzak model için HTTPS gerekir. HTTP yalnızca bu bilgisayardaki bir SSH tüneli için kullanılabilir.")
}

// A marker used only inside the process; no fake bearer token is transmitted.
const localTunnelKey = "__US_LOCAL_TUNNEL__"

func isLoopbackGateway(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u == nil || u.Hostname() == "" {
		return false
	}
	host := u.Hostname()
	ip := net.ParseIP(host)
	return strings.EqualFold(host, "localhost") || (ip != nil && ip.IsLoopback())
}
