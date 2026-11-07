//go:build windows

package main

import "os"

// O_NOFOLLOW does not exist on Windows; reparse points are refused by
// noSymlinkPath (Lstat ModeSymlink) instead.
const openNoFollow = 0

// fileLinkCount: os.FileInfo on Windows does not expose the link count
// without opening a handle; hard links are NOT detected on Windows (documented).
func fileLinkCount(info os.FileInfo) uint64 { return 1 }
