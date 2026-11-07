package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type symbolHit struct {
	Rel  string
	Line int
	Text string
}

func (h symbolHit) Loc() string { return h.Rel + ":" + strconv.Itoa(h.Line) }

var symbolExts = map[string]bool{".go": true, ".py": true, ".js": true, ".mjs": true, ".cjs": true, ".ts": true, ".php": true}

// findSymbols returns definition sites (not call sites) of name in Go, Python,
// JS/TS and PHP sources. Bounded walk; skips symlinks, ignored trees, files
// whose path looks secret, and files over 1 MiB.
func findSymbols(root, name string, limit int) []symbolHit {
	if name == "" || limit <= 0 || !regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`).MatchString(name) {
		return nil
	}
	q := regexp.QuoteMeta(name)
	defRE := regexp.MustCompile(`^\s*(?:` +
		`func\s+(?:\([^)]*\)\s*)?` + q + `\s*[\[(]` + // Go func/method
		`|type\s+` + q + `\b` + // Go type
		`|(?:async\s+)?def\s+` + q + `\s*\(` + // Python def
		`|class\s+` + q + `\b` + // Python/JS/PHP class
		`|(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s*\*?\s*` + q + `\s*\(` + // JS function
		`|(?:export\s+)?(?:const|let|var)\s+` + q + `\s*=` + // JS binding
		`|(?:(?:public|private|protected|static)\s+)*function\s+` + q + `\s*\(` + // PHP
		`)`)
	var out []symbolHit
	scanned := 0
	_ = filepath.WalkDir(root, func(path string, e os.DirEntry, err error) error {
		if err != nil || path == root {
			return nil
		}
		if len(out) >= limit {
			return filepath.SkipAll
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
		if scanned++; scanned > 5000 {
			return filepath.SkipAll
		}
		if !symbolExts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		rel := relPath(root, path)
		low := strings.ToLower(rel)
		for _, s := range []string{"gizli", "secret", "credential", ".env", ".pem", ".key"} {
			if strings.Contains(low, s) {
				return nil
			}
		}
		if info, err := e.Info(); err != nil || !info.Mode().IsRegular() || info.Size() > 1<<20 {
			return nil
		}
		f, err := openBeneath(root, rel)
		if err != nil {
			return nil
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 64<<10), 1<<20)
		for n := 1; sc.Scan(); n++ {
			if line := sc.Text(); defRE.MatchString(line) {
				out = append(out, symbolHit{Rel: rel, Line: n, Text: redactLine(strings.TrimSpace(line))})
				if len(out) >= limit {
					break
				}
			}
		}
		return nil
	})
	return out
}
