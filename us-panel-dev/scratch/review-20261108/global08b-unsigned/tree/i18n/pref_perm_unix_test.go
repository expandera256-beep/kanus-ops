//go:build !windows

package i18n

import (
	"os"
	"testing"
)

// checkPrivateFile: on Unix the mode bits are the access control.
func checkPrivateFile(t *testing.T, path string) {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("perm %v", fi.Mode().Perm())
	}
}
