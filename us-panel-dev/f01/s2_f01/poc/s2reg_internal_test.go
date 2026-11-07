package main

// S2 saldırı regresyonu — İÇ API kullanan PoC'ler.
// UYUM NOKTALARI (.9'da değişebilir): newSafeCmd, (*safeCmd).Run/Stdout,
// newFileTools, (*fileTools).Read/resolve, taskState, pendingEdit,
// saveTaskState, runPersistentRepair, repairLoop, hashContent.
// Bu dosya derlenmezse koşucu yalnız bu dosyayı çıkarır ve senaryolarını BLOCKED yazar.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestS2Reg_H3_SAFECMD_AbsoluteBasename(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(t.TempDir(), "go")
	s2regWrite(t, bin, "#!/bin/sh\necho NOT_REAL_GO\n", 0o755)
	sc, err := newSafeCmd(context.Background(), root, root, []string{bin}, 5*time.Second)
	if err != nil {
		s2regResult(t, "H3-SAFECMD", true, "reddedildi: "+err.Error())
		return
	}
	st, _, _ := sc.Run()
	ran := strings.Contains(sc.Stdout.String(), "NOT_REAL_GO")
	s2regResult(t, "H3-SAFECMD", !ran, fmt.Sprintf("taban adı 'go' olan keyfi betik çalıştı=%v durum=%s", ran, st))
}

func TestS2Reg_M2_TASKSTATE_Tamper(t *testing.T) {
	root := t.TempDir()
	s2regWrite(t, filepath.Join(root, "go.mod"), "module p\n\ngo 1.21\n", 0o644)
	s2regWrite(t, filepath.Join(root, "main.go"), "package main\nfunc main(){}\n", 0o644)
	evil := "package main\n\nfunc main() {}\n\n// INJECTED_BY_TAMPERED_STATE\n"
	st := &taskState{ID: "t1", Root: root, Phase: "applying", PendingEdits: []pendingEdit{{Path: "main.go", Content: evil, Hash: hashContent([]byte(evil))}}}
	if err := saveTaskState(st); err != nil {
		s2regBlocked(t, "M2-TASKSTATE", "durum yazılamadı: "+err.Error())
		return
	}
	approved := false
	l := &repairLoop{Root: root, Endpoint: "http://127.0.0.1:1", Approve: func(int, []repairEdit) bool { approved = true; return false }}
	_, _ = runPersistentRepair(context.Background(), l, "t1", false)
	b, _ := os.ReadFile(filepath.Join(root, "main.go"))
	inj := strings.Contains(string(b), "INJECTED")
	s2regResult(t, "M2-TASKSTATE", !inj, fmt.Sprintf("kurcalanmış kayıt onaysız uygulandı=%v approve_çağrıldı=%v", inj, approved))
}

func TestS2Reg_M3_HARDLINK(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(t.TempDir(), "outside.txt")
	s2regWrite(t, out, "OUTSIDE-SECRET", 0o600)
	if err := os.Link(out, filepath.Join(root, "hl.txt")); err != nil {
		s2regBlocked(t, "M3-HARDLINK", "hard link oluşturulamadı: "+err.Error())
		return
	}
	ft, err := newFileTools(root)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ft.Read("hl.txt")
	leak := err == nil && strings.Contains(string(b), "OUTSIDE-SECRET")
	s2regResult(t, "M3-HARDLINK", !leak, fmt.Sprintf("kök dışı içerik hard link ile okundu=%v err=%v", leak, err))
}

func TestS2Reg_M4_SYMLINK_RACE(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	s2regWrite(t, filepath.Join(outside, "f.txt"), "OUTSIDE-SECRET", 0o600)
	real := filepath.Join(root, "realdir")
	os.MkdirAll(real, 0o755)
	s2regWrite(t, filepath.Join(real, "f.txt"), "inside", 0o644)
	ft, err := newFileTools(root)
	if err != nil {
		t.Fatal(err)
	}
	d := filepath.Join(root, "d")
	var stop atomic.Bool
	done := make(chan struct{})
	go func() { // d: gerçek dizin <-> kök dışı symlink arasında sürekli değişir
		defer close(done)
		for !stop.Load() {
			os.Remove(d)       // d: symlink ise kaldır
			os.Rename(real, d) // d: gerçek dizin (resolve bunu kabul eder)
			os.Rename(d, real)
			os.Symlink(outside, d)
		}
	}()
	leaks, okReads := 0, 0
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		b, err := ft.Read("d/f.txt")
		if err == nil {
			okReads++
			if strings.Contains(string(b), "OUTSIDE-SECRET") {
				leaks++
			}
		}
	}
	stop.Store(true)
	<-done
	s2regResult(t, "M4-SYMLINK-RACE", leaks == 0, fmt.Sprintf("3 sn yarışta başarılı okuma=%d kök dışı okuma=%d (0 sızıntı yarışın kaybedildiğini de gösterebilir; kesin kanıt değil)", okReads, leaks))
}

func TestS2Reg_L1_WindowsPaths(t *testing.T) {
	root := t.TempDir()
	ft, err := newFileTools(root)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ id, p string }{
		{"L1-ADS", "a.txt:gizli"},
		{"L1-ADS-DATA", "a.txt::$DATA"},
		{"L1-DRIVE-REL", "C:evil.txt"},
		{"L1-RESERVED-CON", "CON"},
		{"L1-RESERVED-NUL", "sub/nul.txt"},
		{"L1-RESERVED-AUX", "AUX.log"},
		{"L1-LONGPATH", `\\?\C:\Windows\win.ini`},
		{"L1-UNC", `\\server\share\x.txt`},
		{"L1-DEVICE", `\\.\PhysicalDrive0`},
	}
	for _, c := range cases {
		_, err := ft.resolve(c.p)
		s2regResult(t, c.id, err != nil, fmt.Sprintf("resolve(%q) reddedildi=%v (Linux'ta denetim; Windows davranışı manuel plan W-6/W-7)", c.p, err != nil))
	}
}
