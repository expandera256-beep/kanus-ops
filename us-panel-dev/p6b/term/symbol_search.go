package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// symbolHit is a definition site. It deliberately carries no line text:
// only path, line, kind and the (already validated) name.
type symbolHit struct {
	Rel  string
	Line int
	Kind string
	Name string
}

func (h symbolHit) Loc() string { return h.Rel + ":" + strconv.Itoa(h.Line) }

type symbolStats struct {
	Files     int
	Bytes     int64
	Truncated bool
}

var (
	symbolTimeBudget  = 500 * time.Millisecond
	symbolByteBudget  = int64(8 << 20)
	symbolFileByteCap = int64(1 << 20)
	symbolFileCap     = 5000
)

var symbolNameRE = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

var symbolLang = map[string]string{
	".go": "go", ".py": "py",
	".js": "js", ".mjs": "js", ".cjs": "js", ".jsx": "js", ".ts": "js", ".tsx": "js", ".mts": "js", ".cts": "js",
	".php": "php",
}

type symbolRule struct {
	kind string
	re   *regexp.Regexp
}

func symbolRules(lang, name string) []symbolRule {
	q := regexp.QuoteMeta(name)
	b := `(?:^|[^A-Za-z0-9_$])`
	e := `(?:[^A-Za-z0-9_$]|$)`
	mk := func(kind, re string) symbolRule { return symbolRule{kind, regexp.MustCompile(re)} }
	switch lang {
	case "go":
		return []symbolRule{
			mk("func", `^\s*func\s+(?:\([^)]*\)\s*)?`+q+`\s*[\[(]`),
			mk("type", `^\s*type\s+`+q+e),
			mk("var", `^\s*(?:var|const)\s+`+q+e),
		}
	case "py":
		return []symbolRule{
			mk("func", `^\s*(?:async\s+)?def\s+`+q+`\s*\(`),
			mk("class", `^\s*class\s+`+q+`\s*[:(]`),
			mk("var", `^`+q+`\s*(?::[^=]*)?=[^=]`),
		}
	case "js":
		mods := `(?:(?:export|default|declare|abstract)\s+)*`
		return []symbolRule{
			mk("class", `^\s*`+mods+`class\s+`+q+e),
			mk("interface", `^\s*`+mods+`interface\s+`+q+e),
			mk("type", `^\s*`+mods+`type\s+`+q+`\s*(?:<[^=]*>)?\s*=`),
			mk("enum", `^\s*`+mods+`(?:const\s+)?enum\s+`+q+e),
			mk("func", `^\s*`+mods+`(?:async\s+)?function\s*\*?\s*`+q+`\s*[<(]`),
			mk("var", `^\s*`+mods+`(?:const|let|var)\s+`+q+`\s*[:=]`),
			mk("method", b+`(?:(?:static|async|get|set|public|private|protected|readonly)\s+)*\*?`+q+`\s*\([^()]*\)\s*(?::[^{]*)?\{`),
		}
	case "php":
		mods := `(?:(?:final|abstract|readonly|public|private|protected|static)\s+)*`
		return []symbolRule{
			mk("class", `^\s*`+mods+`class\s+`+q+e),
			mk("interface", `^\s*interface\s+`+q+e),
			mk("trait", `^\s*trait\s+`+q+e),
			mk("enum", `^\s*enum\s+`+q+e),
			mk("func", `^\s*`+mods+`function\s*&?\s*`+q+`\s*\(`),
		}
	}
	return nil
}

// validSymbolRoot: absolute, clean, an existing real directory (not a
// symlink) and not a refused system folder.
func validSymbolRoot(root string) bool {
	if root == "" || !filepath.IsAbs(root) || filepath.Clean(root) != root || filepath.Dir(root) == root || refusedFolder(root) {
		return false
	}
	st, err := os.Lstat(root)
	return err == nil && st.IsDir() && st.Mode()&os.ModeSymlink == 0
}

func findSymbols(root, name string, limit int) []symbolHit {
	h, _ := findSymbolsCtx(context.Background(), root, name, limit)
	return h
}

// findSymbolsCtx returns definition sites of name in Go, Python, JS/TS and
// PHP sources. Every file is opened relative to a pinned root handle with
// openBeneathAt (no symlink following, regular files only, non-blocking), so
// a swap between walk and open cannot pull data from outside root or block
// on a FIFO. Bounded by ctx, symbolTimeBudget, symbolByteBudget (total),
// symbolFileByteCap (per file) and symbolFileCap. Hidden files/dirs,
// ignored trees and secret-looking paths are skipped.
func findSymbolsCtx(parent context.Context, root, name string, limit int) ([]symbolHit, symbolStats) {
	var stats symbolStats
	if name == "" || len(name) > 128 || limit <= 0 || !symbolNameRE.MatchString(name) || !validSymbolRoot(root) {
		return nil, stats
	}
	ctx, cancel := context.WithTimeout(parent, symbolTimeBudget)
	defer cancel()
	if ctx.Err() != nil {
		stats.Truncated = true
		return nil, stats
	}
	h, err := pinRoot(root)
	if err != nil {
		return nil, stats
	}
	defer h.Close()
	rules := map[string][]symbolRule{}
	var out []symbolHit
	_ = filepath.WalkDir(root, func(path string, e os.DirEntry, err error) error {
		if ctx.Err() != nil || stats.Bytes >= symbolByteBudget {
			stats.Truncated = true
			return filepath.SkipAll
		}
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
		if strings.HasPrefix(e.Name(), ".") {
			return nil
		}
		lang := symbolLang[strings.ToLower(filepath.Ext(path))]
		if lang == "" {
			return nil
		}
		if stats.Files++; stats.Files > symbolFileCap {
			stats.Truncated = true
			return filepath.SkipAll
		}
		rel := relPath(root, path)
		low := strings.ToLower(rel)
		for _, s := range []string{"gizli", "secret", "credential", ".env", ".pem", ".key"} {
			if strings.Contains(low, s) {
				return nil
			}
		}
		f, err := openBeneathAt(h, root, rel)
		if err != nil {
			return nil
		}
		defer f.Close()
		st, err := f.Stat()
		if err != nil || !st.Mode().IsRegular() || st.Size() > symbolFileByteCap {
			return nil
		}
		remain := symbolByteBudget - stats.Bytes
		if st.Size() > remain {
			stats.Truncated = true
			return filepath.SkipAll
		}
		if rules[lang] == nil {
			rules[lang] = symbolRules(lang, name)
		}
		cr := &countingReader{r: io.LimitReader(f, symbolFileByteCap)}
		sc := bufio.NewScanner(cr)
		sc.Buffer(make([]byte, 64<<10), int(symbolFileByteCap))
		inDoc := false
		for n := 1; sc.Scan(); n++ {
			if n%256 == 0 && ctx.Err() != nil {
				break
			}
			line := sc.Text()
			if lang == "py" {
				q := strings.Count(line, `"""`) + strings.Count(line, `'''`)
				wasDoc := inDoc
				if q%2 == 1 {
					inDoc = !inDoc
				}
				if wasDoc || q > 0 {
					continue
				}
			}
			if !strings.Contains(line, name) {
				continue
			}
			for _, r := range rules[lang] {
				if r.re.MatchString(line) {
					out = append(out, symbolHit{Rel: rel, Line: n, Kind: r.kind, Name: name})
					break
				}
			}
			if len(out) >= limit {
				break
			}
		}
		stats.Bytes += cr.n
		return nil
	})
	if ctx.Err() != nil {
		stats.Truncated = true
	}
	return out, stats
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}
