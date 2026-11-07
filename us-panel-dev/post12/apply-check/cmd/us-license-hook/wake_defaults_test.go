package main

import (
	"os"
	"testing"
	"time"
)

// All existing route tests explicitly disable wake. Default configuration tests
// below only inspect objects; they never invoke Trigger/ModelReady or dial DNS.
func TestMain(m *testing.M) {
	os.Setenv("US_GPU_WAKE_DISABLE", "1")
	os.Exit(m.Run())
}

func TestWakeBuiltInDefaultsAndDisable(t *testing.T) {
	t.Setenv("US_GPU_WAKE_DISABLE", "0")
	t.Setenv("US_GPU_WAKE_URL", "")
	t.Setenv("US_GPU_STATUS_URL", "")
	t.Setenv("US_GPU_WAKE_TOKEN_FILE", "")
	t.Setenv("US_GPU_WAKE_PIN_FILE", "")
	t.Setenv("US_GPU_WAKE_USE_SYSTEM_CA", "0")
	w, err := configuredWaker()
	if err != nil || w == nil {
		t.Fatalf("defaults unavailable: %v", err)
	}
	if w.WakeURL != "https://34.41.186.170:8443/a100/start" || w.StatusURL != "https://34.41.186.170:8443/a100/status" || w.PinFile != "/etc/us-model/wake.pin" || w.TokenFile != "/etc/us-model/wake.token" || w.WarmupTimeout != 15*time.Minute {
		t.Fatal("built-in defaults differ")
	}
	t.Setenv("US_GPU_WAKE_DISABLE", "1")
	if w, err = configuredWaker(); err != nil || w != nil {
		t.Fatal("explicit disable ignored")
	}
}

func TestWakeInjectedURLsStayLocal(t *testing.T) {
	t.Setenv("US_GPU_WAKE_DISABLE", "0")
	t.Setenv("US_GPU_WAKE_URL", "http://127.0.0.1:9/a100/start")
	t.Setenv("US_GPU_STATUS_URL", "http://127.0.0.1:9/a100/status")
	w, err := configuredWaker()
	if err != nil || w == nil || w.WakeURL != "http://127.0.0.1:9/a100/start" {
		t.Fatalf("local config: %v", err)
	}
}
