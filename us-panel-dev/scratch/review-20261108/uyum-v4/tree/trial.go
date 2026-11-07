package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type trialResult struct {
	Ran  bool
	OK   bool
	Line string
}

func materialize(dir string) (string, error) {
	scene := makeScenario(0)
	path := filepath.Join(dir, "pricing.php")
	return path, os.WriteFile(path, []byte(scene.Expected), 0o644)
}

func runTrial() trialResult {
	dir, err := os.MkdirTemp("", "kiyas-gorev-")
	if err != nil {
		return trialResult{Line: "geçici klasör açılmadı"}
	}
	defer os.RemoveAll(dir)
	file, err := materialize(dir)
	if err != nil {
		return trialResult{Line: "pricing.php yazılamadı"}
	}
	if _, err := exec.LookPath("php"); err != nil {
		body, readErr := os.ReadFile(file)
		if readErr != nil {
			return trialResult{Line: "dosya okunamadı"}
		}
		line, ok := checkPhpText(string(body))
		return trialResult{Ran: true, OK: ok, Line: line}
	}
	out, err := exec.Command("php", "-l", file).CombinedOutput()
	shown := strings.TrimSpace(strings.ReplaceAll(string(out), file, "pricing.php"))
	shown = strings.ReplaceAll(shown, dir, "ayrik://")
	if shown == "" {
		shown = "php -l cevap vermedi"
	}
	return trialResult{Ran: true, OK: err == nil && strings.Contains(shown, "No syntax errors"), Line: shown}
}
