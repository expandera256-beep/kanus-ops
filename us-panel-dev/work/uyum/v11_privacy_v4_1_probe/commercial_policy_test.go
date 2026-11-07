package main

import (
	"strings"
	"testing"
	"time"
)

func TestCommercialReleaseCannotBypassAccountWithLocalOrLegacyKeys(t *testing.T) {
	old := releaseChannel
	releaseChannel = "commercial"
	t.Cleanup(func() { releaseChannel = old })
	t.Setenv("US_REQUIRE_COMMERCIAL", "0")
	t.Setenv("US_KEY", "legacy-key")
	t.Setenv("US_GATEWAY_URL", tunnelURL)
	t.Setenv("US_MODEL_NAME", tunnelModel)
	s := newView("a100")
	s.key = localTunnelKey
	if s.modelKey() != "" || s.licenseBlock(time.Now()) == "" {
		t.Fatal("local bypass in commercial release")
	}
	t.Setenv("US_GATEWAY_URL", "https://gateway.example.com")
	s.key = strings.Repeat("s", 40)
	if s.modelKey() != "" {
		t.Fatal("raw token accepted before account login")
	}
	s.accountGateway = "https://gateway.example.com"
	if s.modelKey() != s.key || s.licenseBlock(time.Now()) != "" {
		t.Fatal("authenticated account rejected")
	}
	t.Setenv("US_GATEWAY_URL", "https://other.example.com")
	if s.modelKey() != "" || s.licenseBlock(time.Now()) == "" {
		t.Fatal("credential crossed gateway")
	}
}

func TestNewProjectCreationUsesFileActionProtocol(t *testing.T) {
	if projectRepairRequest("Yeni proje oluştur.\nEksik veya yanlış oturum için 401 döndür.") {
		t.Fatal("requirements mistaken for repair intent")
	}
	for _, task := range []string{"app.py dosyasını oluştur", "Yeni proje oluştur", "Implement project files", "app.py hatasını düzelt"} {
		if !fileMutationRequest(task) {
			t.Fatal(task)
		}
	}
	for _, task := range []string{"selam", "bir şiir yaz", "projenin amacı nedir"} {
		if fileMutationRequest(task) {
			t.Fatal(task)
		}
	}
}

func TestCommercialGatewayUsesReleaseURLWithoutLegacyFallback(t *testing.T) {
	oldChannel, oldURL := releaseChannel, embeddedGatewayURL
	releaseChannel, embeddedGatewayURL = "commercial", "https://verified-gateway.example/"
	t.Cleanup(func() { releaseChannel, embeddedGatewayURL = oldChannel, oldURL })
	t.Setenv("US_GATEWAY_URL", "")
	if activeGateway() != "https://verified-gateway.example" || len(doorBases()) != 1 || doorBases()[0] != activeGateway() {
		t.Fatal("release gateway not selected exclusively")
	}
	embeddedGatewayURL = ""
	if activeGateway() != "" || validateGateway(activeGateway()) == nil {
		t.Fatal("missing release URL fell back to legacy gateway")
	}
	t.Setenv("US_GATEWAY_URL", "https://configured.example")
	if activeGateway() != "https://configured.example" || len(doorBases()) != 1 {
		t.Fatal("explicit gateway ignored")
	}
}
