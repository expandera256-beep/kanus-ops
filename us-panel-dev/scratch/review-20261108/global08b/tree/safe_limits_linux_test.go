//go:build linux

package main

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestSafeExecAppliesLinuxRlimits(t *testing.T) {
	py := needPython(t)
	root := t.TempDir()
	sc, err := newSafeCmd(context.Background(), root, root, []string{py, "-c", "import resource,time\ntime.sleep(0.3)\nprint(resource.getrlimit(resource.RLIMIT_CPU)[0], resource.getrlimit(resource.RLIMIT_DATA)[0], resource.getrlimit(resource.RLIMIT_FSIZE)[0])"}, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	state, _, _ := sc.Run()
	got := strings.Fields(sc.Stdout.String())
	if state != "success" || len(got) != 3 || got[0] != "600" || got[1] != "4294967296" || got[2] != "536870912" {
		t.Fatalf("rlimit uygulanmadı: %s %q %s", state, got, sc.Stderr.String())
	}
}
