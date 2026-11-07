package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const tunnelAddress = "127.0.0.1:1923"
const tunnelURL = "http://" + tunnelAddress
const sshTunnelForward = tunnelAddress + ":127.0.0.1:8080"
const tunnelModel = "deepseek-v4.1-flash"

type tunnelManager struct {
	mu     sync.Mutex
	state  string
	active bool
}

var autoTunnel = &tunnelManager{state: "Otomatik bağlantı bekliyor."}

func (m *tunnelManager) status() string  { m.mu.Lock(); defer m.mu.Unlock(); return m.state }
func (m *tunnelManager) update(s string) { m.mu.Lock(); m.state = s; m.mu.Unlock() }
func (m *tunnelManager) claim() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.active {
		return false
	}
	m.active = true
	return true
}
func (m *tunnelManager) done() { m.mu.Lock(); m.active = false; m.mu.Unlock() }

func modelAvailable(base string, client *http.Client) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return modelAvailableContext(ctx, base, client)
}
func modelAvailableContext(ctx context.Context, base string, client *http.Client) bool {
	if ctx.Err() != nil {
		return false
	}
	safeClient := *client
	safeClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+"/v1/models", nil)
	if err != nil {
		return false
	}
	resp, err := safeClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var listing struct {
		Data []struct {
			ID     string `json:"id"`
			Status struct {
				Value string `json:"value"`
			} `json:"status"`
		} `json:"data"`
	}
	if json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&listing) != nil {
		return false
	}
	for _, item := range listing.Data {
		if item.ID == tunnelModel && (item.Status.Value == "" || item.Status.Value == "loaded") {
			return true
		}
	}
	return false
}
func portOccupied() bool {
	conn, err := net.DialTimeout("tcp", tunnelAddress, 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func startAutoTunnel(s *session) {
	if !autoTunnel.claim() {
		return
	}
	go func() {
		defer autoTunnel.done()
		client := &http.Client{Timeout: 2 * time.Second}
		if modelAvailable(tunnelURL, client) {
			activateTunnel(s)
			return
		}
		if portOccupied() {
			autoTunnel.update("1923 portu başka bir servis tarafından kullanılıyor; bağlantı başlatılmadı.")
			return
		}
		autoTunnel.update("Google Cloud bağlantısı açılıyor. Gerekirse açılan pencerede oturum açın.")
		if err := startCloudSSH(); err != nil {
			autoTunnel.update(fmt.Sprintf("Bağlantı başlatılamadı: %v", err))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if waitForModel(ctx, tunnelURL, client, 2*time.Second) == nil {
			activateTunnel(s)
			return
		}
		if portOccupied() {
			autoTunnel.update("SSH portu açık, fakat DeepSeek modeli hazır değil; ds-boot/ds-serve durumunu kontrol edin. Hazır sayılmadı.")
		} else {
			autoTunnel.update("Bağlantı zaman aşımına uğradı. VM/SSH veya oturum açma durumunu kontrol edip Yeniden bağlan'a basın.")
		}
	}()
}
func activateTunnel(s *session) {
	unlock := s.hold()
	defer unlock()
	current := strings.TrimSpace(os.Getenv("US_GATEWAY_URL"))
	if current != "" && strings.TrimRight(current, "/") != tunnelURL {
		autoTunnel.update("Tünel hazır. Özel model adresiniz korunuyor; otomatik seçilmedi.")
		return
	}
	if err := os.Setenv("US_GATEWAY_URL", tunnelURL); err != nil {
		autoTunnel.update("Model adresi ayarlanamadı.")
		return
	}
	if err := os.Setenv("US_MODEL_NAME", tunnelModel); err != nil {
		autoTunnel.update("Model adı ayarlanamadı.")
		return
	}
	s.retarget("a100")
	autoTunnel.update("DeepSeek bağlantısı hazır. SSH penceresini açık tutun.")
	go watchTunnel()
}

func watchTunnel() {
	client := &http.Client{Timeout: 2 * time.Second}
	failures := 0
	for {
		time.Sleep(8 * time.Second)
		if strings.TrimRight(strings.TrimSpace(os.Getenv("US_GATEWAY_URL")), "/") != tunnelURL {
			return
		}
		if modelAvailable(tunnelURL, client) {
			failures = 0
			continue
		}
		failures++
		if failures >= 3 {
			autoTunnel.update("DeepSeek bağlantısı kesildi. Google Cloud penceresini kontrol edip Yeniden bağlan'a basın.")
			return
		}
	}
}

// Build only a fixed command, using a validated local CLI path.
func cloudSSHCommand(gcloud string) (string, error) {
	if gcloud == "" || strings.ContainsAny(gcloud, "\"\r\n&|<>^%!") {
		return "", fmt.Errorf("Google Cloud CLI yolu geçersiz")
	}
	return `call "` + gcloud + `" compute ssh ds-a100 --project=project-27082a40-6083-4952-9db --zone=europe-west4-b -- -N -o ExitOnForwardFailure=yes -o ConnectTimeout=20 -o ServerAliveInterval=15 -o ServerAliveCountMax=3 -L ` + sshTunnelForward, nil
}

// waitForModel bounds the entire readiness loop, including each HTTP request.
// This is model readiness, not evidence that a VM has been started/stopped.
func waitForModel(ctx context.Context, base string, client *http.Client, interval time.Duration) error {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	for {
		probe, cancel := context.WithTimeout(ctx, 2*time.Second)
		ready := modelAvailableContext(probe, base, client)
		cancel()
		if ready {
			return nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
