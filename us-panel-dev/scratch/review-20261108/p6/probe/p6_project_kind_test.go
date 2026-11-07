package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func p6write(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestP6DetectProjectKind(t *testing.T) {
	cases := []struct {
		files map[string]string
		want  string
	}{
		{map[string]string{"go.mod": "module x\n", "main.go": "package main\n"}, "go"},
		{map[string]string{"package.json": "{}"}, "node"},
		{map[string]string{"pyproject.toml": "", "a.py": ""}, "python"},
		{map[string]string{"app.py": "print(1)\n"}, "python"},
		{map[string]string{"main.go": "package main\n"}, "go-nomod"},
		{map[string]string{"index.js": "1\n"}, "node-nopkg"},
		{map[string]string{"README.md": "x"}, "unknown"},
	}
	for i, c := range cases {
		root := t.TempDir()
		for k, v := range c.files {
			p6write(t, root, k, v)
		}
		if got, _ := detectProjectKind(root); got != c.want {
			t.Fatalf("case %d: got %q want %q", i, got, c.want)
		}
	}
}

// Kök neden: go.mod olmayan Go projesinde çalıştırıcı Python unittest'e düşüyordu
// (yanlış test aracı). Artık açık bir "go.mod yok" mesajı verilmeli.
func TestP6RunnerDoesNotPickPythonForGoWithoutMod(t *testing.T) {
	root := t.TempDir()
	p6write(t, root, "main.go", "package main\nfunc main(){}\n")
	out := runProjectTests(root)
	if strings.Contains(out, "unittest") || strings.Contains(out, "Python") {
		t.Fatalf("wrong tool chosen: %s", out)
	}
	if !strings.Contains(out, "go.mod") {
		t.Fatalf("expected go.mod notice: %s", out)
	}
}

func TestP6RunnerDoesNotPickPythonForJSWithoutPackage(t *testing.T) {
	root := t.TempDir()
	p6write(t, root, "index.js", "console.log(1)\n")
	out := runProjectTests(root)
	if strings.Contains(out, "unittest") || strings.Contains(out, "Python") {
		t.Fatalf("wrong tool chosen: %s", out)
	}
	if !strings.Contains(out, "package.json") {
		t.Fatalf("expected package.json notice: %s", out)
	}
}
