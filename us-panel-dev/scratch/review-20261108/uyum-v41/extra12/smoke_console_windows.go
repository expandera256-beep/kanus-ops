//go:build windows

package main

import (
	"os"
	"syscall"
)

func attachSmokeConsole() {
	syscall.NewLazyDLL("kernel32.dll").NewProc("AttachConsole").Call(uintptr(^uint32(0)))
	if _, err := os.Stdout.Stat(); err != nil {
		if f, e := os.OpenFile("CONOUT$", os.O_WRONLY, 0); e == nil {
			os.Stdout = f
		}
	}
}
