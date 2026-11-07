package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProposalBatchRollsBackWhenSecondWriteFails(t *testing.T) {
	root := t.TempDir()
	for _, n := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(root, n), []byte("old\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := prepareProposal(root, "DOSYA a.txt\nfirst\nDOSYA b.txt\nsecond\n", map[string]bool{"a.txt": true, "b.txt": true}, map[string]string{"a.txt": hashContent([]byte("old\n")), "b.txt": hashContent([]byte("old\n"))})
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	_, err = applyProposalBatch(root, p.Files, func(root, full string, b []byte) error {
		calls++
		if calls == 2 {
			return errors.New("simulated failure")
		}
		return writeModelFile(root, full, b)
	})
	if err == nil || !strings.Contains(err.Error(), "geri alındı") {
		t.Fatalf("unexpected result: %v", err)
	}
	for _, n := range []string{"a.txt", "b.txt"} {
		b, e := os.ReadFile(filepath.Join(root, n))
		if e != nil || string(b) != "old\n" {
			t.Fatalf("%s: %q %v", n, b, e)
		}
	}
}

func TestProposalBatchDoesNotOverwriteExternalChangeDuringRollback(t *testing.T) {
	root := t.TempDir()
	for _, n := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(root, n), []byte("old\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := prepareProposal(root, "DOSYA a.txt\nfirst\nDOSYA b.txt\nsecond\n", map[string]bool{"a.txt": true, "b.txt": true}, map[string]string{"a.txt": hashContent([]byte("old\n")), "b.txt": hashContent([]byte("old\n"))})
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	_, err = applyProposalBatch(root, p.Files, func(root, full string, b []byte) error {
		calls++
		if calls == 2 {
			if e := os.WriteFile(filepath.Join(root, "a.txt"), []byte("user\n"), 0600); e != nil {
				t.Fatal(e)
			}
			return errors.New("failure")
		}
		return writeModelFile(root, full, b)
	})
	if err == nil || !strings.Contains(err.Error(), "geri alma tamamlanamadı") {
		t.Fatalf("unexpected result: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(root, "a.txt"))
	if string(b) != "user\n" {
		t.Fatalf("external change lost: %q", b)
	}
}

func TestProposalBatchSuccessAndPreflight(t *testing.T) {
	root := t.TempDir()
	original := []byte("old\n")
	name := filepath.Join(root, "a.txt")
	if err := os.WriteFile(name, original, 0600); err != nil {
		t.Fatal(err)
	}
	p, err := prepareProposal(root, "DOSYA a.txt\nnew\nDOSYA b.txt\ncreated\n", map[string]bool{"a.txt": true}, map[string]string{"a.txt": hashContent(original)})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "b.txt"), []byte("external"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = applyProposalBatch(root, p.Files, writeModelFile); err == nil {
		t.Fatal("conflict accepted")
	}
	if b, _ := os.ReadFile(name); string(b) != "old\n" {
		t.Fatal("preflight changed file")
	}
	if err = os.Remove(filepath.Join(root, "b.txt")); err != nil {
		t.Fatal(err)
	}
	if _, err = applyProposalBatch(root, p.Files, writeModelFile); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(name); string(b) != "new\n" {
		t.Fatalf("%q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(root, "b.txt")); string(b) != "created\n" {
		t.Fatalf("%q", b)
	}
}
