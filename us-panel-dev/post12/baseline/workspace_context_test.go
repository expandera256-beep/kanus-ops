package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectRequestIncludesWorkspacePaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("gizli içerik"), 0o644); err != nil {
		t.Fatal(err)
	}
	wire := chatWire("mevcut altyapıyı tamamla", root)
	if !strings.Contains(wire, "README.md") || !strings.Contains(wire, "OKU") || strings.Contains(wire, "gizli içerik") {
		t.Fatal(wire)
	}
	if strings.Contains(chatWire("merhaba", root), "README.md") {
		t.Fatal("ordinary chat disclosed workspace paths")
	}
}

func TestExamplePathsAreNotCreated(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"tam/yol/klasör/altyapi.txt", "path/to/file.txt", "<dosya>.txt"} {
		if note := applyReplyFiles(root, "DOSYA "+name+"\nörnek içerik"); note != "" {
			t.Fatalf("placeholder created: %s", note)
		}
	}
}
