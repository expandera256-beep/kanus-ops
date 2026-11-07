//go:build !linux && !windows

package i18n

import "testing"

func openHandleCount(t *testing.T) int {
	t.Helper()
	t.Skip("handle count not implemented on this OS")
	return 0
}
