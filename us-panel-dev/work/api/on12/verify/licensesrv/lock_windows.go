//go:build windows

package licensesrv

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var lockFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("LockFileEx")

func lockStore(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	var overlap syscall.Overlapped
	ok, _, e := lockFileEx.Call(f.Fd(), 3, 0, 1, 0, uintptr(unsafe.Pointer(&overlap)))
	if ok == 0 {
		f.Close()
		return nil, fmt.Errorf("ledger writer lock unavailable: %w", e)
	}
	return f, nil
}
func unlockStore(f *os.File) error {
	if f == nil {
		return nil
	}
	return f.Close()
}

// Windows directory fsync is not exposed portably; file fsync precedes rename.
func syncStoreDir(string) error { return nil }
