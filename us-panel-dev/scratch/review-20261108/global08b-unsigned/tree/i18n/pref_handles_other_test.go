//go:build !linux && !windows

package i18n

import (
	"testing"

	"golang.org/x/sys/unix"
)

func openHandleCount(t *testing.T) int {
	t.Helper()
	t.Skip("handle count not implemented on this OS")
	return 0
}

func handleValid(h rootHandle) bool {
	_, err := unix.FcntlInt(uintptr(h), unix.F_GETFD, 0)
	return err == nil
}
