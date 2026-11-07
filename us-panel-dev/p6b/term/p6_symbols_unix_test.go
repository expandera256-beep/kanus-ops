//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// PANEL DEV S1: dosya-symlink swap yarışı kök dışını okutmamalı.
func TestP6SymbolFileSwapTOCTOU(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	p6write(t, out, "s.go", "func Leak(){} // OUTSIDE\n")
	p6write(t, root, "a.go", "package a\n")
	stop, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		tmp := filepath.Join(root, "t")
		for {
			select {
			case <-stop:
				return
			default:
			}
			os.Remove(tmp)
			os.Symlink(filepath.Join(out, "s.go"), tmp)
			os.Rename(tmp, filepath.Join(root, "a.go"))
			os.WriteFile(tmp, []byte("package a\n"), 0o600)
			os.Rename(tmp, filepath.Join(root, "a.go"))
		}
	}()
	leaked := 0
	for end := time.Now().Add(2 * time.Second); time.Now().Before(end); {
		if h := findSymbols(root, "Leak", 5); len(h) > 0 {
			leaked++
		}
	}
	close(stop)
	<-done
	if leaked > 0 {
		t.Fatalf("TOCTOU: outside file read %d times", leaked)
	}
}

// PANEL DEV S2: kontrol ile açma arasında FIFO konursa askıda kalmamalı.
func TestP6SymbolFIFOSwapNoHang(t *testing.T) {
	root := t.TempDir()
	p6write(t, root, "a.go", "package a\n")
	stop, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		tmp := filepath.Join(root, "f")
		for {
			select {
			case <-stop:
				return
			default:
			}
			os.Remove(tmp)
			syscall.Mkfifo(tmp, 0o600)
			os.Rename(tmp, filepath.Join(root, "a.go"))
			os.WriteFile(tmp, []byte("package a\n"), 0o600)
			os.Rename(tmp, filepath.Join(root, "a.go"))
		}
	}()
	res := make(chan bool, 1)
	go func() {
		for end := time.Now().Add(2 * time.Second); time.Now().Before(end); {
			findSymbols(root, "X", 5)
		}
		res <- true
	}()
	select {
	case <-res:
		close(stop)
		<-done
	case <-time.After(6 * time.Second):
		close(stop)
		t.Fatal("findSymbols hung on FIFO")
	}
}

func TestP6SymbolStaticFIFOSkipped(t *testing.T) {
	root := t.TempDir()
	syscall.Mkfifo(filepath.Join(root, "p.go"), 0o600)
	c := make(chan int, 1)
	go func() { c <- len(findSymbols(root, "X", 5)) }()
	select {
	case <-c:
	case <-time.After(3 * time.Second):
		t.Fatal("static FIFO hang")
	}
}
