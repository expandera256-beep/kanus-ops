package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Explicitly initiated by the user. Project tests can execute project code.
func runProjectTests(root string) string {
	return runProjectTestsWithContext(context.Background(), root, 60*time.Second)
}

func projectTestEnv() []string { return safeExecEnv("PYTHONIOENCODING=utf-8") }

func runProjectTestsWithContext(parent context.Context, root string, timeout time.Duration) string {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() || refusedFolder(root) {
		return "Test çalıştırılamadı: geçerli proje klasörü yok."
	}
	var bin string
	var args []string
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
		bin = "go"
		args = []string{"test", "-v", "./..."}
	} else if _, err := os.Stat(filepath.Join(root, "package.json")); err == nil {
		var reason string
		bin, args, reason = nodeTestCommand(root)
		if reason != "" {
			return reason
		}
	} else if kind, why := detectProjectKind(root); projectKindNotice(kind, why) != "" {
		return projectKindNotice(kind, why)
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
	started := time.Now()
	combined, state, _, err := safeRun(ctx, root, root, append([]string{bin}, args...), timeout, nil, "PYTHONIOENCODING=utf-8")
	if state == "start_failed" {
		return "Test çalıştırılamadı: " + bin + " güvenli çalıştırıcıdan geçmedi: " + redactLine(fmt.Sprint(err))
	}
	output := struct{ s string }{combined}
	elapsed := time.Since(started).Round(time.Millisecond)
	text := "Komut: " + strings.Join(append([]string{bin}, args...), " ") + "\n" + redactLine(strings.TrimSpace(output.s))
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
	if !projectTestEvidence(output.s) {
		if strings.Contains(text, "[no test files]") || strings.Contains(text, "[no tests to run]") {
			return "Test bulunamadı; başarı doğrulanmadı.\n" + text
		}
		return "Test doğrulanmadı (komut çıkış kodu 0; gerçek test kanıtı yok).\n" + text
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

// Exit 0 alone is not evidence that any test actually ran. Recognize explicit
// successful cases/counts from Go, unittest, Node TAP, Jest and Vitest. Unknown
// project runners remain unverified rather than fabricating a green result.
func projectTestEvidence(output string) bool {
	plain := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(output, "")
	if regexp.MustCompile(`(?m)^\s*--- PASS: `).MatchString(plain) {
		return true
	}
	if m := regexp.MustCompile(`Ran ([0-9]+) tests?`).FindStringSubmatch(plain); m != nil {
		count, _ := strconv.Atoi(m[1])
		skipped := 0
		if n := regexp.MustCompile(`skipped=([0-9]+)`).FindStringSubmatch(plain); n != nil {
			skipped, _ = strconv.Atoi(n[1])
		}
		return count > skipped
	}
	return regexp.MustCompile(`(?m)^(?:#|ℹ)\s*pass\s+[1-9][0-9]*\s*$`).MatchString(plain) || regexp.MustCompile(`(?m)Tests:?\s+[^\n]*[1-9][0-9]* passed`).MatchString(plain)
}
