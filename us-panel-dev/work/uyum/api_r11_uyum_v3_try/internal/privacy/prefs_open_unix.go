//go:build unix

package privacy

import (
	"os"
	"syscall"
)

func openPrefsNoFollow(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
}
