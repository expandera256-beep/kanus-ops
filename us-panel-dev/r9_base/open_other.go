//go:build !windows

package main

import "os/exec"

func init() { launchPanel = openPanel }

func openPanel(url string) {
	_ = exec.Command("xdg-open", url).Start()
}
