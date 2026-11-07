package licensesrv

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStoreWriterProcess(t *testing.T) {
	path := os.Getenv("US_LOCK_TEST_PATH")
	if path == "" {
		t.Skip("subprocess helper")
	}
	st, err := OpenStore(path)
	if os.Getenv("US_LOCK_TEST_EXPECT") == "locked" {
		if err == nil {
			st.Close()
			t.Fatal("concurrent writer accepted")
		}
		if !strings.Contains(err.Error(), "writer lock") {
			t.Fatal(err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	st.mu.Lock()
	st.st.Webhooks["child"] = time.Now()
	err = st.saveLocked()
	st.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
}
func TestStoreConcurrentProcessWriterLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.json")
	st, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	exe, _ := os.Executable()
	run := func(want string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, exe, "-test.run=^TestStoreWriterProcess$", "-test.v")
		cmd.Env = append(os.Environ(), "US_LOCK_TEST_PATH="+path, "US_LOCK_TEST_EXPECT="+want)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatal(err, string(out))
		}
	}
	run("locked")
	st.Close()
	run("open")
	again, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer again.Close()
	if _, ok := again.st.Webhooks["child"]; !ok {
		t.Fatal("child write not durable")
	}
}
