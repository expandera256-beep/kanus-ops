//go:build linux

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

// applySafeLimits sets RLIMIT_CPU/DATA/FSIZE (DATA, not AS, so Go/V8 address-space reservations still work) on the started child via
// prlimit(2). Children forked later inherit these limits.
func applySafeLimits(proc *os.Process, l safeLimits) (func(), error) {
	if proc == nil {
		return func() {}, nil
	}
	set := func(res int, v uint64) error {
		if v == 0 {
			return nil
		}
		var cur unix.Rlimit
		if err := unix.Prlimit(proc.Pid, res, nil, &cur); err != nil {
			return err
		}
		if cur.Max != unix.RLIM_INFINITY && cur.Max < v {
			v = cur.Max
		}
		return unix.Prlimit(proc.Pid, res, &unix.Rlimit{Cur: v, Max: v}, nil)
	}
	for _, item := range []struct {
		res int
		v   uint64
	}{{unix.RLIMIT_CPU, l.CPUSeconds}, {unix.RLIMIT_DATA, l.MemoryBytes}, {unix.RLIMIT_FSIZE, l.FileBytes}} {
		if err := set(item.res, item.v); err != nil {
			return nil, err
		}
	}
	// RLIMIT_NPROC is deliberately NOT set: on Linux it counts every
	// thread of the real UID (the panel itself included), so a low value
	// breaks threaded servers. Process count is bounded on Windows by the
	// Job Object; on Linux use a cgroup/container for multi-tenant hosting.
	return func() {}, nil
}
