//go:build !linux && !windows

package main

import "os"

// No portable rlimit for an already started child on this platform.
func applySafeLimits(proc *os.Process, l safeLimits) (func(), error) { return func() {}, nil }
