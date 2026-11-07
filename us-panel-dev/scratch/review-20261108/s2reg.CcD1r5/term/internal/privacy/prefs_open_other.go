//go:build !linux

package privacy

import (
	"errors"
	"os"
)

// Explicit fail-closed limitation: Windows (and other non-Linux platforms) need
// a reviewed handle-relative, reparse-safe implementation before persistence is
// enabled. No plain os.Open fallback, no symlink race silently accepted.
var errPrefsPlatform = errors.New("privacy prefs: secure persistence requires Linux; unsupported platform")

func openPrefsNoFollow(path string) (*os.File, error) { return nil, errPrefsPlatform }
func writePrefsAtomic(path string, raw []byte, fault func(string) error) error {
	return errPrefsPlatform
}
