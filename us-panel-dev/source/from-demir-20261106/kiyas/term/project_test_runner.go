package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Explicitly initiated by the user. Project tests can execute project code.
func runProjectTests(root string) string {
	return runProjectTestsWithContext(context.Background(), root, 60*time.Second)
}

func projectTestEnv() []string {
	var env []string
	allowed := map[string]bool{"PATH": true, "HOME": true, "USERPROFILE": true, "SYSTEMROOT": true, "WINDIR": true, "TEMP": true, "TMP": true, "TMPDIR": true, "LANG": true, "LC_ALL": true, "PATHEXT": true, "COMSPEC": true}
	for _, value := range os.Environ() {
		key, _, _ := strings.Cut(value, "=")
		if allowed[strings.ToUpper(key)] {
			env = append(env, value)
		}
	}
	return append(env, "PYTHONIOENCODING=utf-8")
}

func runProjectTestsWithContext(parent context.Context, root string, timeout time.Duration) string {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() || refusedFolder(root) {
		return "Test çalıştırılamadı: geçerli proje klasörü yok."
	}
	var bin string
	var args []string
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
		bin = "go"
		args = []string{"test", "./..."}
	} else if _, err := os.Stat(filepath.Join(root, "package.json")); err == nil {
		var reason string
		bin, args, reason = nodeTestCommand(root)
		if reason != "" {
			return reason
		}
	} else {
		var prefix []string
		var found bool
		bin, prefix, found = findPython(root, root)
		if !found {
			return "Test çalıştırılamadı: kullanılabilir Python yorumlayıcısı bulunamadı."
		}
		args = append(prefix, "-m", "unittest", "discover")
	}
	if _, err = exec.LookPath(bin); err != nil {
		return "Test çalıştırılamadı: " + bin + " bulunamadı."
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	configureProcessTree(cmd)
	cmd.Cancel = func() error { return killProcessTree(cmd.Process) }
	cmd.WaitDelay = 2 * time.Second
	cmd.Dir = root
	cmd.Env = projectTestEnv()
	var output cappedBuf
	cmd.Stdout = &output
	cmd.Stderr = &output
	started := time.Now()
	err = cmd.Run()
	elapsed := time.Since(started).Round(time.Millisecond)
	text := "Komut: " + strings.Join(append([]string{bin}, args...), " ") + "\n" + redactLine(strings.TrimSpace(output.String()))
	if ctx.Err() != nil {
		if parent.Err() != nil {
			return fmt.Sprintf("Test iptal edildi (%s).\n%s", elapsed, text)
		}
		return fmt.Sprintf("Test zaman aşımı (%s).\n%s", elapsed, text)
	}
	if strings.Contains(text, "Ran 0 tests") {
		return fmt.Sprintf("Test bulunamadı; başarı doğrulanmadı (süre %s).\n%s", elapsed, text)
	}
	if err != nil {
		return fmt.Sprintf("Test başarısız (%v; süre %s).\n%s", err, elapsed, text)
	}
	return fmt.Sprintf("Test komutu başarılı (çıkış kodu 0; süre %s).\n%s", elapsed, text)
}

// Only the project's explicitly configured test script is selected. No model
// output is evaluated as a command, and package installation is never implied.
func nodeTestCommand(root string) (string, []string, string) {
	data, _, err := sourceSnapshot(root, "package.json")
	if err != nil || len(data) > 1<<20 {
		return "", nil, "Test çalıştırılamadı: package.json güvenle okunamadı."
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return "", nil, "Test çalıştırılamadı: package.json geçersiz."
	}
	if strings.TrimSpace(pkg.Scripts["test"]) == "" {
		return "", nil, "Test bulunamadı; package.json içinde test betiği tanımlı değil."
	}
	npm, err := exec.LookPath("npm")
	if err != nil {
		return "", nil, "Test çalıştırılamadı: npm bulunamadı; bağımlılıklar otomatik yüklenmez."
	}
	if runtime.GOOS != "windows" {
		return npm, []string{"test", "--ignore-scripts"}, ""
	}
	// Execute npm's JS entry with node directly, avoiding cmd.exe's command
	// parsing and Windows batch-file quoting entirely.
	node, err := exec.LookPath("node")
	if err != nil {
		return "", nil, "Test çalıştırılamadı: Node.js bulunamadı."
	}
	for _, base := range []string{filepath.Dir(npm), filepath.Dir(node)} {
		cli := filepath.Join(base, "node_modules", "npm", "bin", "npm-cli.js")
		if st, e := os.Stat(cli); e == nil && st.Mode().IsRegular() {
			return node, []string{cli, "test", "--ignore-scripts"}, ""
		}
	}
	return "", nil, "Test çalıştırılamadı: npm-cli.js bulunamadı; Node.js kurulumunu denetleyin."
}
