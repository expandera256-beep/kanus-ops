package main

import "testing"

func TestR13PortableAliasAndDevicePolicy(t *testing.T) {
	for _, name := range []string{"LONGFI~1.TXT", "a:stream", "CON.txt", "COM¹", "LPT².txt", "LPT³", "a.", "a "} {
		if windowsSafeRel(name) {
			t.Errorf("unsafe Windows component accepted: %q", name)
		}
	}
	for _, name := range []string{"long filename.txt", "src/main.go", "company.txt"} {
		if !windowsSafeRel(name) {
			t.Errorf("safe component refused: %q", name)
		}
	}
}
