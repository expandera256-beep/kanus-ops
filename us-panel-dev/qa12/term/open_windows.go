//go:build windows

package main

import (
	"os/exec"
	"runtime"
	"syscall"
	"unsafe"
)

func init() {
	launchPanel = openPanel
	notifyUser = messageBox
}

func messageBox(text string) {
	msg, err1 := syscall.UTF16PtrFromString(text)
	title, err2 := syscall.UTF16PtrFromString("US")
	if err1 != nil || err2 != nil {
		return
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	user32.NewProc("MessageBoxW").Call(0, uintptr(unsafe.Pointer(msg)), uintptr(unsafe.Pointer(title)), 0)
	runtime.KeepAlive(msg)
	runtime.KeepAlive(title)
}

func openPanel(url string) {
	if shellOpen(url) {
		return
	}
	if exec.Command("explorer.exe", url).Start() == nil {
		return
	}
	if exec.Command("cmd", "/c", "start", "", url).Start() == nil {
		return
	}
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

func shellOpen(url string) bool {
	verb, err1 := syscall.UTF16PtrFromString("open")
	target, err2 := syscall.UTF16PtrFromString(url)
	if err1 != nil || err2 != nil {
		return false
	}
	shell32 := syscall.NewLazyDLL("shell32.dll")
	r, _, _ := shell32.NewProc("ShellExecuteW").Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(target)), 0, 0, 1)
	runtime.KeepAlive(verb)
	runtime.KeepAlive(target)
	return r > 32
}
