//go:build windows

package main

import (
	"errors"
	"os"
	"path/filepath"
)

// openBeneath on Windows: components are checked by noSymlinkPath (reparse
// points refused) before open; a race-free handle-relative open is NOT
// implemented here (NOT FIXED on Windows, documented).
func openBeneath(root, rel string) (*os.File, error) {
	full := filepath.Join(root, rel)
	if err := noSymlinkPath(root, full); err != nil {
		return nil, errFileToolEscape
	}
	f, err := os.Open(full)
	if err != nil {
		return nil, err
	}
	if info, err := f.Stat(); err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, errors.New("normal dosya değil")
	}
	return f, nil
}
