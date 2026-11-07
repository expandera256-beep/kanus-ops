package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileToolsRefuseEscapes(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("gizli"), 0o644); err != nil {
		t.Fatal(err)
	}
	ft, err := newFileTools(root)
	if err != nil {
		t.Fatal(err)
	}
	symOK := os.Symlink(outside, filepath.Join(root, "linkdir")) == nil
	if symOK {
		_ = os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(root, "link.txt"))
	}
	bad := []string{"../x", "a/../../x", `..\x`, filepath.Join(outside, "secret.txt"), "/etc/passwd", "", "."}
	if symOK {
		bad = append(bad, "linkdir/secret.txt", "linkdir/new.txt", "link.txt")
	}
	for _, rel := range bad {
		if _, err := ft.Read(rel); err == nil {
			t.Fatalf("okuma kaçışı kabul edildi: %q", rel)
		}
		if err := ft.Write(rel, []byte("x")); err == nil {
			t.Fatalf("yazma kaçışı kabul edildi: %q", rel)
		}
		if err := ft.Delete(rel); err == nil {
			t.Fatalf("silme kaçışı kabul edildi: %q", rel)
		}
	}
	if raw, _ := os.ReadFile(filepath.Join(outside, "secret.txt")); string(raw) != "gizli" {
		t.Fatal("kök dışı dosya değişti")
	}
	if entries, _ := os.ReadDir(outside); len(entries) != 1 {
		t.Fatal("kök dışına dosya yazıldı")
	}
	if _, err := newFileTools(filepath.Join(root, "linkdir")); symOK && err == nil {
		t.Fatal("symlink kök kabul edildi")
	}
}

func TestFileToolsAtomicWriteAndLimits(t *testing.T) {
	root := t.TempDir()
	ft, _ := newFileTools(root)
	ft.MaxWrite = 1024
	ft.MaxRead = 2048
	if err := ft.Write("pkg/a.go", []byte("package a\n")); err != nil {
		t.Fatal(err)
	}
	if got, err := ft.Read("pkg/a.go"); err != nil || string(got) != "package a\n" {
		t.Fatalf("%q %v", got, err)
	}
	if err := ft.Write("pkg/big.txt", make([]byte, 1025)); err == nil {
		t.Fatal("yazma sınırı uygulanmadı")
	}
	if err := os.WriteFile(filepath.Join(root, "huge.txt"), make([]byte, 4096), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ft.Read("huge.txt"); err == nil {
		t.Fatal("okuma sınırı uygulanmadı")
	}
	// Simulated crash between temp write and rename: original stays intact,
	// no partial or temp file remains.
	atomicWriteTestHook = func(string) error { return errors.New("çökme simülasyonu") }
	err := ft.Write("pkg/a.go", []byte("package a // YARIM"))
	atomicWriteTestHook = nil
	if err == nil {
		t.Fatal("çökme simülasyonu hata vermedi")
	}
	if got, _ := os.ReadFile(filepath.Join(root, "pkg", "a.go")); string(got) != "package a\n" {
		t.Fatalf("kısmi yazım görünür: %q", got)
	}
	entries, _ := os.ReadDir(filepath.Join(root, "pkg"))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".us-write-") {
			t.Fatal("geçici dosya kaldı")
		}
	}
	if err := ft.Delete("pkg/a.go"); err != nil {
		t.Fatal(err)
	}
}

func TestFileToolsWorkspacesAreSeparate(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	fa, _ := newFileTools(a)
	fb, _ := newFileTools(b)
	if err := fa.Write("x.txt", []byte("A")); err != nil {
		t.Fatal(err)
	}
	if _, err := fb.Read(filepath.Join(a, "x.txt")); err == nil {
		t.Fatal("B çalışma alanı A dosyasını okudu")
	}
	if _, err := fb.Read("x.txt"); err == nil {
		t.Fatal("B'de olmayan dosya okundu")
	}
}

func TestAppendFileIsAtomicAndBackedUp(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "n.txt"), []byte("bir\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newView("q-append")
	s.root = root
	s.appendFile("n.txt", "iki\n")
	if got, _ := os.ReadFile(filepath.Join(root, "n.txt")); string(got) != "bir\niki\n" {
		t.Fatalf("%q notice=%q", got, s.notice)
	}
	hist, _ := os.ReadDir(filepath.Join(root, changeHistoryDir))
	if len(hist) == 0 {
		t.Fatal("ekleme öncesi yedek yok")
	}
	if os.Symlink(t.TempDir(), filepath.Join(root, "dis")) == nil {
		s.appendFile("dis/y.txt", "x")
		if _, err := os.Stat(filepath.Join(root, "dis", "y.txt")); err == nil {
			t.Fatal("symlink üzerinden ekleme yapıldı")
		}
	}
}
