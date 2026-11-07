//go:build !windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func configureProcessTree(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} }

// killProcessTree freezes the child's process group, collects every live
// descendant (including ones that called setsid and left the group — PANEL
// DEV MEDIUM-1) and SIGKILLs them and their own groups. Limitation: a
// descendant whose parent chain already exited (double-fork reparented to
// init) before cancellation cannot be attributed without cgroups.
func killProcessTree(proc *os.Process) error {
	if proc == nil {
		return nil
	}
	own := syscall.Getpgrp()
	_ = syscall.Kill(-proc.Pid, syscall.SIGSTOP)
	seen := map[int]bool{}
	for round := 0; round < 3; round++ {
		for _, pid := range descendantPids(proc.Pid) {
			if !seen[pid] {
				seen[pid] = true
				_ = syscall.Kill(pid, syscall.SIGSTOP)
			}
		}
	}
	for pid := range seen {
		if pg, err := syscall.Getpgid(pid); err == nil && pg == pid && pg != own {
			_ = syscall.Kill(-pg, syscall.SIGKILL)
		}
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	if err := syscall.Kill(-proc.Pid, syscall.SIGKILL); err != nil {
		return proc.Kill()
	}
	return nil
}
