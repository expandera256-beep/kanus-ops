//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// The helper runs using the same token/environment preparation as production,
// and the same Job Object resource limits through safeCmd.Start.
func TestF01WindowsChild(t *testing.T) {
	if os.Getenv("F01_LOW_CHILD") != "1" {
		t.Skip("yalnız Low IL çocuk yordamı")
	}
	fail := func(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(21) }
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		fail(err)
	}
	defer token.Close()
	buf := make([]byte, 1024)
	var size uint32
	if err := windows.GetTokenInformation(token, windows.TokenIntegrityLevel, &buf[0], uint32(len(buf)), &size); err != nil {
		fail(err)
	}
	label := (*windows.Tokenmandatorylabel)(unsafe.Pointer(&buf[0]))
	if label.Label.Sid.String() != "S-1-16-4096" {
		fail(fmt.Errorf("Low IL değil: %s", label.Label.Sid.String()))
	}
	groups, err := token.GetTokenGroups()
	if err != nil {
		fail(err)
	}
	admin, _ := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	for _, g := range groups.AllGroups() {
		if g.Sid.Equals(admin) && g.Attributes&windows.SE_GROUP_USE_FOR_DENY_ONLY == 0 {
			fail(fmt.Errorf("admin SID deny-only değil"))
		}
	}
	privbuf := make([]byte, 4096)
	if err := windows.GetTokenInformation(token, windows.TokenPrivileges, &privbuf[0], uint32(len(privbuf)), &size); err != nil {
		fail(err)
	}
	privs := (*windows.Tokenprivileges)(unsafe.Pointer(&privbuf[0]))
	// DISABLE_MAX_PRIVILEGE leaves only traversal privilege, if present.
	var traverse windows.LUID
	if err := windows.LookupPrivilegeValue(nil, windows.StringToUTF16Ptr("SeChangeNotifyPrivilege"), &traverse); err != nil {
		fail(err)
	}
	for _, p := range privs.AllPrivileges() {
		if p.Luid != traverse {
			fail(fmt.Errorf("ayrıcalık silinmedi: %v", p.Luid))
		}
	}
	if err := os.WriteFile("existing.txt", []byte("updated"), 0600); err != nil {
		fail(fmt.Errorf("varolan workspace dosyasına yazma: %w", err))
	}
	if err := os.WriteFile("inside.txt", []byte("ok"), 0600); err != nil {
		fail(fmt.Errorf("workspace yazma: %w", err))
	}
	outside := os.Getenv("F01_OUTSIDE")
	if err := os.WriteFile(filepath.Join(outside, "bad.txt"), []byte("bad"), 0600); !errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
		fail(fmt.Errorf("dışarı yazma ERROR_ACCESS_DENIED değil: %v", err))
	}
	// Explicitly demonstrate that Low IL does NOT prevent outside reads.
	if b, err := os.ReadFile(filepath.Join(outside, "synthetic.txt")); err != nil || string(b) != "F01-FAKE" {
		fail(fmt.Errorf("okuma-izolasyonu yok kanıtı: %v", err))
	}
	fmt.Println("LOW_IL_OK ADMIN_DENY_ONLY_OK PRIVILEGES_DROPPED_OK INSIDE_WRITE_OK OUTSIDE_WRITE_DENIED OUTSIDE_READ_STILL_ALLOWED")
	for _, tool := range []string{"go", "node", "python"} {
		p, err := exec.LookPath(tool)
		if err != nil {
			fmt.Printf("TOOL_SKIP %s: araç PATH içinde yok\n", tool)
			continue
		}
		arg := "--version"
		if tool == "go" {
			arg = "version"
		}
		out, err := exec.Command(p, arg).CombinedOutput()
		if err != nil {
			fail(fmt.Errorf("TOOL_FAIL %s (kalıtılan Low IL): %v %s", tool, err, out))
		}
		fmt.Printf("TOOL_LOW_IL_OK %s %s", tool, out)
	}
	os.Exit(0)
}

func TestF01WindowsLowIL(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "project")
	outside := filepath.Join(base, "outside")
	for _, d := range []string{root, outside} {
		if err := os.Mkdir(d, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(outside, "synthetic.txt"), []byte("F01-FAKE"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "existing.txt"), []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	// Ensure an explicit Medium IL sibling even if the caller's TEMP is Low.
	sd, err := windows.SecurityDescriptorFromString("S:(ML;OICI;NW;;;ME)")
	if err != nil {
		t.Fatal(err)
	}
	sacl, _, err := sd.SACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(outside, windows.SE_FILE_OBJECT, windows.LABEL_SECURITY_INFORMATION, nil, nil, nil, sacl); err != nil {
		t.Fatal(err)
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, self, "-test.run=^TestF01WindowsChild$", "-test.v")
	configureProcessTree(cmd)
	cmd.Dir = root
	cmd.Env = append(safeExecEnv(), "F01_LOW_CHILD=1", "F01_OUTSIDE="+outside)
	sc := &safeCmd{Cmd: cmd, root: root, Stdout: &cappedBuf{}, Stderr: &cappedBuf{}, ctx: ctx, cancel: cancel}
	cmd.Stdout = sc.Stdout
	cmd.Stderr = sc.Stderr
	state, _, err := sc.Run()
	out := sc.Stdout.String() + sc.Stderr.String()
	t.Log(out)
	if err != nil || state != "success" || !strings.Contains(out, "OUTSIDE_WRITE_DENIED") {
		t.Fatalf("state=%s err=%v", state, err)
	}
	if _, err := os.Stat(filepath.Join(outside, "bad.txt")); !os.IsNotExist(err) {
		t.Fatal("dış dosya oluştu")
	}
	if b, err := os.ReadFile(filepath.Join(root, "inside.txt")); err != nil || string(b) != "ok" {
		t.Fatal("iç dosya yok", err)
	}
}
