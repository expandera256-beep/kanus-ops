//go:build !windows

package main

func enableWindowsVT() {}

func attachConsole() error { return nil }

func becomeStandalone() bool { return false }

func driveRoots() []string {
	return collectDrives(nil)
}
