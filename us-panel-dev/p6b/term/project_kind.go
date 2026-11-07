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
	nested := map[string]string{} // shallowest sub-folder manifest per toolchain
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
		if e.Type().IsRegular() {
			for name, k := range map[string]string{"go.mod": "go", "package.json": "node"} {
				if e.Name() == name {
					rel := relPath(root, path)
					if old, ok := nested[k]; !ok || strings.Count(rel, "/") < strings.Count(old, "/") {
						nested[k] = rel
					}
				}
			}
		}
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
	if best == "" {
		if nested["go"] != "" {
			best = "go"
		} else if nested["node"] != "" {
			best = "node"
		}
	}
	switch best {
	case "go":
		if nested["go"] != "" {
			return "go-nomod", nested["go"]
		}
		return "go-nomod", "*.go"
	case "node":
		if nested["node"] != "" {
			return "node-nopkg", nested["node"]
		}
		return "node-nopkg", "*.js"
	case "python":
		return "python", "*.py"
	}
	return "unknown", ""
}

// projectKindNotice explains why no test runner was started for kinds that
// have no runnable root manifest; "" means a runner may be chosen.
func projectKindNotice(kind, evidence string) string {
	sub := strings.Contains(evidence, "/")
	switch kind {
	case "go-nomod":
		if sub {
			return "Test çalıştırılamadı: go.mod kök klasörde değil, alt klasörde (" + evidence + "); o alt klasörü proje olarak seçin. Başka dilin test aracı seçilmedi."
		}
		return "Test çalıştırılamadı: Go kaynakları var ama go.mod yok; başka dilin test aracı seçilmedi."
	case "node-nopkg":
		if sub {
			return "Test çalıştırılamadı: package.json kök klasörde değil, alt klasörde (" + evidence + "); o alt klasörü proje olarak seçin. Başka dilin test aracı seçilmedi."
		}
		return "Test çalıştırılamadı: JavaScript kaynakları var ama package.json yok; başka dilin test aracı seçilmedi."
	case "unknown":
		return "Test bulunamadı; bu proje türü için desteklenen test aracı yok (Go, Node, Python desteklenir); başarı doğrulanmadı."
	}
	return ""
}
