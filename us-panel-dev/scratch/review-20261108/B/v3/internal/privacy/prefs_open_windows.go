//go:build windows

package privacy

import "os"

// On Windows, Lstat already refused reparse/symlink targets; plain open follows.
func openPrefsNoFollow(path string) (*os.File, error) {
	return os.Open(path)
}
