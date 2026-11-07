//go:build !windows

package main

import (
	"os"
	"syscall"
)

const openNoFollow = syscall.O_NOFOLLOW

// fileLinkCount returns the hard link count (1 when unknown).
func fileLinkCount(info os.FileInfo) uint64 {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return uint64(st.Nlink)
	}
	return 1
}
