//go:build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

func driveRoots() []string {
	kernel := syscall.NewLazyDLL("kernel32.dll")
	mask, _, _ := kernel.NewProc("GetLogicalDrives").Call()
	var out []string
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		out = append(out, string(rune('A'+i))+`:\`)
	}
	return collectDrives(out)
}

func becomeStandalone() bool {
	if os.Getenv("US_STANDALONE") == "1" {
		return false
	}
	exe, err := os.Executable()
	if err != nil || exe == "" {
		exe = os.Args[0]
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Env = append(os.Environ(), "US_STANDALONE=1")
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x01000000 | 0x00000200 | 0x00000010,
	}
	if err := cmd.Start(); err != nil {
		return false
	}
	return true
}

func attachConsole() error {
	kernel := syscall.NewLazyDLL("kernel32.dll")
	hwnd, _, _ := kernel.NewProc("GetConsoleWindow").Call()
	if hwnd == 0 {
		if r, _, err := kernel.NewProc("AllocConsole").Call(); r == 0 {
			return err
		}
	}
	if hwnd != 0 && os.Getenv("US_STANDALONE") != "1" {
		return nil
	}
	in, err := syscall.Open("CONIN$", syscall.O_RDWR, 0)
	if err != nil {
		return err
	}
	out, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		return err
	}
	os.Stdin = os.NewFile(uintptr(in), "stdin")
	os.Stdout = os.NewFile(uintptr(out), "stdout")
	os.Stderr = os.NewFile(uintptr(out), "stderr")
	return nil
}

func enableWindowsVT() {
	kernel := syscall.NewLazyDLL("kernel32.dll")
	getMode := kernel.NewProc("GetConsoleMode")
	setMode := kernel.NewProc("SetConsoleMode")
	setCP := kernel.NewProc("SetConsoleOutputCP")
	setCP.Call(65001)
	for _, file := range []*os.File{os.Stdout, os.Stdin} {
		var mode uint32
		handle := syscall.Handle(file.Fd())
		if r, _, _ := getMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode))); r == 0 {
			continue
		}
		const enableVirtualTerminal = 0x0004
		setMode.Call(uintptr(handle), uintptr(mode|enableVirtualTerminal))
	}
}
