//go:build linux

package i18n

import (
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func openHandleCount(t *testing.T) int {
	t.Helper()
	ents, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Skip("/proc/self/fd unavailable:", err)
	}
	return len(ents)
}

func handleValid(h rootHandle) bool {
	_, err := unix.FcntlInt(uintptr(h), unix.F_GETFD, 0)
	return err == nil
}
