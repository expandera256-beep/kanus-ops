package main

import (
	"context"
	"strings"
	"testing"
)

func TestLegacyDirectDispatchDefaultOff(t *testing.T) {
	t.Setenv("US_DEV_DIRECT_MODEL", "")
	old := releaseAuthority
	releaseAuthority = "accounts"
	defer func() { releaseAuthority = old }()
	out := postDoorContext(context.Background(), "a100", "x", "hello", "system", 16, "", 0, nil, defaultContextLimits())
	if out.ok || !strings.Contains(out.body, "kapalı") {
		t.Fatal(out)
	}
}
