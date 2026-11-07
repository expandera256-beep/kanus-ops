//go:build linux

package i18n

import (
	"os"
	"testing"
)

func openHandleCount(t *testing.T) int {
	t.Helper()
	ents, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Skip("/proc/self/fd unavailable:", err)
	}
	return len(ents)
}
