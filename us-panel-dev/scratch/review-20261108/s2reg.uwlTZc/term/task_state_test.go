package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const utilFixed = "package main\n\nfunc helper() int { return 1 }\n"

func twoFileFix() string {
	return `{"files":[{"path":"main.go","content":` + jsonString(fixedMain) + `},{"path":"util.go","content":` + jsonString(utilFixed) + `}]}`
}

func jsonString(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\t", `\t`)
	return `"` + r.Replace(s) + `"`
}

func crashTestLoop(root, url string) *repairLoop {
	return &repairLoop{Root: root, Endpoint: url, Key: "test-key", Model: "mock",
		Budget:  repairBudget{MaxIterations: 3, MaxModelCalls: 3, MaxDuration: 2 * time.Minute, StepTimeout: time.Minute, MaxEditBytes: 1 << 16},
		Approve: func(int, []repairEdit) bool { return true }}
}

// Child process body: runs the persistent loop and hard-exits at the crash
// point, like a killed panel process (no deferred cleanup, no final save).
func TestTaskStateCrashChild(t *testing.T) {
	point := os.Getenv("US_TASK_CRASH_POINT")
	if point == "" {
		t.Skip("yalnız çökme simülasyonu alt süreci")
	}
	repairCrashHook = func(p string) {
		if p == point {
			os.Exit(3)
		}
	}
	_, _ = runPersistentRepair(context.Background(), crashTestLoop(os.Getenv("US_TASK_CRASH_ROOT"), os.Getenv("US_TASK_CRASH_URL")), "gorev1", false)
	os.Exit(0)
}

func runCrashChild(t *testing.T, root, url, point string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestTaskStateCrashChild$", "-test.count=1")
	cmd.Env = append(os.Environ(), "US_TASK_CRASH_POINT="+point, "US_TASK_CRASH_ROOT="+root, "US_TASK_CRASH_URL="+url)
	out, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 {
		t.Fatalf("alt süreç çökme noktasında durmadı (%s): %v\n%s", point, err, out)
	}
}

func historyCount(t *testing.T, root, rel string) int {
	entries, _ := os.ReadDir(filepath.Join(root, changeHistoryDir))
	n := 0
	for _, e := range entries {
		raw, _ := os.ReadFile(filepath.Join(root, changeHistoryDir, e.Name()))
		if strings.Contains(string(raw), `"path":"`+rel+`"`) || strings.Contains(string(raw), `"Path":"`+rel+`"`) {
			n++
		}
	}
	return n
}

func TestTaskStateResumeAfterCrashMidApply(t *testing.T) {
	t.Setenv("US_TASK_STATE_KEY_FILE", filepath.Join(t.TempDir(), "k"))
	root := brokenGoProject(t)
	if err := os.WriteFile(filepath.Join(root, "util.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	srv, calls := mockModel(t, func(string, int) string { return twoFileFix() })
	runCrashChild(t, root, srv.URL, "mid_apply")
	if *calls != 1 {
		t.Fatalf("çökme öncesi model çağrısı %d", *calls)
	}
	st, err := loadTaskState(root, "gorev1")
	if err != nil || st == nil || st.Phase != "applying" || len(st.PendingEdits) != 2 {
		t.Fatalf("kalıcı durum yanlış: %+v %v", st, err)
	}
	if got, _ := os.ReadFile(filepath.Join(root, "main.go")); string(got) != fixedMain {
		t.Fatal("ilk dosya çökme öncesi yazılmamış")
	}
	if got, _ := os.ReadFile(filepath.Join(root, "util.go")); string(got) == utilFixed {
		t.Fatal("ikinci dosya beklenmedik şekilde yazılmış")
	}
	res, err := runPersistentRepair(context.Background(), crashTestLoop(root, srv.URL), "gorev1", false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "fixed" || *calls != 1 || res.ModelCalls != 1 {
		t.Fatalf("devam yanlış: %s calls=%d modelCalls=%d", res.Status, *calls, res.ModelCalls)
	}
	if historyCount(t, root, "main.go") != 1 || historyCount(t, root, "util.go") != 1 {
		t.Fatalf("düzeltme çift uygulandı: main=%d util=%d", historyCount(t, root, "main.go"), historyCount(t, root, "util.go"))
	}
	if got, _ := os.ReadFile(filepath.Join(root, "util.go")); string(got) != utilFixed {
		t.Fatal("bekleyen düzeltme tamamlanmadı")
	}
	// Completed task is not re-run.
	again, err := runPersistentRepair(context.Background(), crashTestLoop(root, srv.URL), "gorev1", false)
	if err != nil || again.Status != "fixed" || *calls != 1 {
		t.Fatalf("tamamlanan görev tekrarlandı: %+v %v", again, err)
	}
}

func TestTaskStateUncertainModelCallNotReplayed(t *testing.T) {
	t.Setenv("US_TASK_STATE_KEY_FILE", filepath.Join(t.TempDir(), "k"))
	root := brokenGoProject(t)
	srv, calls := mockModel(t, func(string, int) string { return fileJSON("main.go", fixedMain) })
	runCrashChild(t, root, srv.URL, "model_in_flight")
	if _, err := runPersistentRepair(context.Background(), crashTestLoop(root, srv.URL), "gorev1", false); !errors.Is(err, errUncertainModelCall) {
		t.Fatalf("belirsiz istek otomatik tekrarlandı: %v", err)
	}
	if *calls != 0 {
		t.Fatalf("model çağrıldı: %d", *calls)
	}
	res, err := runPersistentRepair(context.Background(), crashTestLoop(root, srv.URL), "gorev1", true)
	if err != nil || res.Status != "fixed" || *calls != 1 {
		t.Fatalf("onaylı tekrar başarısız: %+v %v calls=%d", res, err, *calls)
	}
}

func TestTaskStateCorruptFileIsPreserved(t *testing.T) {
	t.Setenv("US_TASK_STATE_KEY_FILE", filepath.Join(t.TempDir(), "k"))
	root := t.TempDir()
	path, _ := taskStatePath(root, "bozuk")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{yarım"), 0o600); err != nil {
		t.Fatal(err)
	}
	if st, err := loadTaskState(root, "bozuk"); err == nil || st != nil {
		t.Fatal("bozuk kayıt kabul edildi")
	}
	matches, _ := filepath.Glob(path + ".corrupt-*")
	if len(matches) != 1 {
		t.Fatal("bozuk kayıt korunmadı")
	}
	if _, err := taskStatePath(root, "../x"); err == nil {
		t.Fatal("kimlik ile yol kaçışı")
	}
}
