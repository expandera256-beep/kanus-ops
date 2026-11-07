package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Session root identity (r13b, PANEL DEV R13-01/R13-02).
//
// The workspace root is pinned exactly once, when the session selects it
// (setRoot), and every desk/history operation goes through that handle. The
// root path is never re-resolved per call, so renaming the root or any of its
// ancestors after selection (and planting a symlink/junction in its place)
// cannot redirect reads, writes, history records, undo or cleanup.
//
// A session built as a struct literal (tests, helpers) is pinned lazily on
// its first desk operation; that first operation is the session start.

// setRoot changes the session workspace and pins it immediately.
func (s *session) setRoot(dir string) {
	s.root = dir
	s.repinRoot()
}

func (s *session) repinRoot() {
	if s.rootPin != nil {
		_ = s.rootPin.Close()
	}
	s.rootPin, s.rootPinAbs = nil, ""
	s.rootPinPath = s.root
	if strings.TrimSpace(s.root) == "" || refusedFolder(s.root) {
		return
	}
	abs, err := filepath.Abs(s.root)
	if err != nil {
		return
	}
	h, err := pinRoot(abs)
	if err != nil {
		return
	}
	s.rootPin, s.rootPinAbs = h, abs
}

// pinnedRoot returns the handle pinned for the current s.root. It pins only
// when s.root was changed without setRoot (once per root value).
func (s *session) pinnedRoot() (*rootHandle, string, error) {
	if s.rootPinPath != s.root || (s.rootPin == nil && s.rootPinPath == "") {
		s.repinRoot()
	}
	if s.rootPin == nil {
		return nil, "", errFileToolEscape
	}
	return s.rootPin, s.rootPinAbs, nil
}

// pinnedRel validates rel lexically against the pinned root and returns the
// OS-relative path to use with the *BeneathAt API. No ambient lstat/stat.
func (s *session) pinnedRel(rel string) (*rootHandle, string, string, error) {
	h, abs, err := s.pinnedRoot()
	if err != nil {
		return nil, "", "", err
	}
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return nil, "", "", os.ErrInvalid
	}
	ft := &fileTools{Root: abs, root: h}
	full, err := ft.resolve(rel)
	if err != nil {
		return nil, "", "", err
	}
	r, err := filepath.Rel(abs, full)
	if err != nil {
		return nil, "", "", errFileToolEscape
	}
	return h, abs, r, nil
}

func (s *session) undoLast() (string, error) {
	h, abs, err := s.pinnedRoot()
	if err != nil {
		return "", err
	}
	return undoLastChangeAt(h, abs)
}
