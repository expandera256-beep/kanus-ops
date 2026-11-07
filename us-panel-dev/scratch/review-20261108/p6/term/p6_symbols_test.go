package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestP6FindSymbolDefinitions(t *testing.T) {
	root := t.TempDir()
	p6write(t, root, "a/calc.go", "package a\n\nfunc Total(x int) int { return x }\ntype Cart struct{}\nfunc (c *Cart) Total() int { return 0 }\n")
	p6write(t, root, "b/util.py", "class Cart:\n    def total(self):\n        return 0\n\ndef Total():\n    pass\n")
	p6write(t, root, "c/app.js", "export function Total(a) { return a }\nconst Cart = () => 1\nclass Cart2 {}\n")
	p6write(t, root, "secret/.env", "func Total() {}\n")
	p6write(t, root, "node_modules/x/i.js", "function Total(){}\n")
	got := findSymbols(root, "Total", 50)
	want := map[string]bool{"a/calc.go:3": true, "a/calc.go:5": true, "b/util.py:5": true, "c/app.js:1": true}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for _, h := range got {
		if !want[h.Loc()] {
			t.Fatalf("unexpected hit %v", h)
		}
	}
	if h := findSymbols(root, "Cart", 50); len(h) != 3 { // go type, py class, js const
		t.Fatalf("Cart hits %v", h)
	}
	if h := findSymbols(root, "Total", 1); len(h) != 1 {
		t.Fatalf("limit not applied %v", h)
	}
	if h := findSymbols(root, "", 10); len(h) != 0 {
		t.Fatalf("empty name must return nothing")
	}
}

func TestP6FindSymbolsSkipsSymlinks(t *testing.T) {
	root := t.TempDir()
	out := t.TempDir()
	p6write(t, out, "x.go", "package x\nfunc Leak(){}\n")
	if err := os.Symlink(filepath.Join(out, "x.go"), filepath.Join(root, "x.go")); err != nil {
		t.Skip("symlink unsupported")
	}
	if h := findSymbols(root, "Leak", 10); len(h) != 0 {
		t.Fatalf("followed symlink: %v", h)
	}
}
