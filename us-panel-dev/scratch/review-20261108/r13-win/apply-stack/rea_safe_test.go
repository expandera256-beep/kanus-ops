//go:build linux

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

func TestReaPinnedVersionNoLatest(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "npx"), []byte("#!/bin/sh\necho npx\n"), 0o755)
	t.Setenv("PATH", dir)
	t.Setenv("US_REA", "")
	_, prefix, err := reaBinary()
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(prefix, " ")
	if strings.Contains(joined, "@latest") || !strings.Contains(joined, "rea-agents@") {
		t.Fatalf("sürüm sabitlenmedi: %q", joined)
	}
}

// The REA run goes through safe exec: env allowlist (NODE_OPTIONS dropped),
// own process group, and the project root is not on the binary path.
func TestReaRunsThroughSafeExec(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "a.js"), []byte("1"), 0o644)
	bin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "env.txt")
	os.WriteFile(filepath.Join(bin, "rea"), []byte("#!/bin/sh\nenv > "+marker+"\nps -o pgid= -p $$ >> "+marker+"\necho REA-OK\n"), 0o755)
	t.Setenv("PATH", bin+":/usr/bin:/bin")
	t.Setenv("US_REA", "")
	t.Setenv("NODE_OPTIONS", "--require=/tmp/evil.js")
	out := runREA(root, "a.js", "js")
	if !strings.Contains(out, "REA-OK") {
		t.Fatalf("REA çıktı: %q", out)
	}
	env, _ := os.ReadFile(marker)
	if strings.Contains(string(env), "NODE_OPTIONS") {
		t.Fatal("NODE_OPTIONS REA'ya geçti (safe exec dışı)")
	}
	lines := strings.Fields(string(env))
	pg := lines[len(lines)-1]
	if pg == "" || pg == strconv.Itoa(syscall.Getpgrp()) {
		t.Fatalf("ayrı süreç grubu yok: %q", pg)
	}
}
