//go:build windows

package i18n

import (
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32                  = windows.NewLazySystemDLL("kernel32.dll")
	procGetProcessHandleCount = kernel32.NewProc("GetProcessHandleCount")
	procGetHandleInformation  = kernel32.NewProc("GetHandleInformation")
)

func openHandleCount(t *testing.T) int {
	t.Helper()
	var n uint32
	r, _, err := procGetProcessHandleCount.Call(uintptr(windows.CurrentProcess()), uintptr(unsafe.Pointer(&n)))
	if r == 0 {
		t.Fatal("GetProcessHandleCount:", err)
	}
	return int(n)
}

// handleValid: GetHandleInformation fails with ERROR_INVALID_HANDLE once
// the handle is closed.
func handleValid(h rootHandle) bool {
	var flags uint32
	r, _, _ := procGetHandleInformation.Call(uintptr(h), uintptr(unsafe.Pointer(&flags)))
	return r != 0
}
