package main

import (
	"os"
	"path/filepath"
	"strings"
)

// detectProjectKind reports the project's build/test toolchain from manifest
// files first, then from source extensions (bounded, no symlinks). The
// "-nomod"/"-nopkg" kinds mean sources exist without the manifest the
// toolchain needs, so no other language's runner must be chosen instead.
func detectProjectKind(root string) (string, string) {
	has := func(name string) bool {
		st, err := os.Lstat(filepath.Join(root, name))
		return err == nil && st.Mode().IsRegular()
	}
	switch {
	case has("go.mod"):
		return "go", "go.mod"
	case has("package.json"):
		return "node", "package.json"
	case has("pyproject.toml"), has("setup.py"), has("requirements.txt"):
		return "python", "python manifest"
	}
	counts := map[string]int{}
	scanned := 0
	_ = filepath.WalkDir(root, func(path string, e os.DirEntry, err error) error {
		if err != nil || path == root {
			return nil
		}
		if e.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if e.IsDir() {
			if skipTreeDir(e.Name()) || depth(root, path) > workspaceDepth {
				return filepath.SkipDir
			}
			return nil
		}
		scanned++
		if scanned > 5000 {
			return filepath.SkipAll
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".go":
			counts["go"]++
		case ".py":
			counts["python"]++
		case ".js", ".mjs", ".cjs", ".ts":
			counts["node"]++
		}
		return nil
	})
	best, n := "", 0
	for _, k := range []string{"go", "python", "node"} {
		if counts[k] > n {
			best, n = k, counts[k]
		}
	}
	switch best {
	case "go":
		return "go-nomod", "*.go"
	case "node":
		return "node-nopkg", "*.js"
	case "python":
		return "python", "*.py"
	}
	return "unknown", ""
}
