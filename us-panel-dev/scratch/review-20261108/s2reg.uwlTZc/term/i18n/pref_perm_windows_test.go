//go:build windows

package i18n

import (
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Rights that let a principal modify, replace or re-permission the file.
const winWriteMask = windows.FILE_WRITE_DATA | windows.FILE_APPEND_DATA | windows.FILE_WRITE_EA |
	windows.FILE_WRITE_ATTRIBUTES | windows.DELETE | windows.WRITE_DAC | windows.WRITE_OWNER |
	windows.GENERIC_WRITE | windows.GENERIC_ALL

// checkPrivateFile: on Windows Go's FileMode only mirrors the read-only
// attribute (always -rw-rw-rw- for writable files), so the DACL is checked:
// owner is the current user, DACL is protected (not inherited) and
// Everyone / Users / Authenticated Users have no write-class right.
func checkPrivateFile(t *testing.T, path string) {
	t.Helper()
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	tu, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := sd.Owner()
	if err != nil || !owner.Equals(tu.User.Sid) {
		t.Errorf("owner %v, want current user %v", owner, tu.User.Sid)
	}
	ctl, _, err := sd.Control()
	if err != nil || ctl&windows.SE_DACL_PROTECTED == 0 {
		t.Errorf("DACL not protected (inherits from parent): %v", sd)
	}
	dacl, _, err := sd.DACL()
	if err != nil || dacl == nil {
		t.Fatalf("no DACL (= everyone full access): %v", err)
	}
	var bad []*windows.SID
	for _, w := range []windows.WELL_KNOWN_SID_TYPE{windows.WinWorldSid, windows.WinBuiltinUsersSid, windows.WinAuthenticatedUserSid} {
		s, err := windows.CreateWellKnownSid(w)
		if err != nil {
			t.Fatal(err)
		}
		bad = append(bad, s)
	}
	for i := uint32(0); i < uint32(dacl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, i, &ace); err != nil {
			t.Fatal(err)
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			continue
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		for _, b := range bad {
			if sid.Equals(b) && uint32(ace.Mask)&winWriteMask != 0 {
				t.Errorf("write access for %v (mask %#x): %v", b, ace.Mask, sd)
			}
		}
	}
}
