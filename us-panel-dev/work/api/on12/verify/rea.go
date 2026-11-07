package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

func (s *session) keepREA(note string) {
	note = strings.TrimSpace(note)
	if note == "" {
		note = "REA çıktı vermedi."
	}
	s.reaNote = note
	s.review = note
	s.notice = ""
	s.flow = append(s.flow, event{title: "rea", meta: "yerel", body: trimRunes(note, 700), kind: "yerel"})
	s.noteLog("işlem", "rea")
	s.noteLog("düşünce", note)
	s.ran = true
}

func reaPrompt(text string) bool {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "rea.tools") || strings.Contains(lower, "rea-agents") {
		return true
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return false
	}
	head := strings.ToLower(strings.Trim(fields[0], "/"))
	return head == "rea"
}

func runREA(root, rel, command string) string {
	if refusedFolder(root) {
		return "Bu klasör taranmaz."
	}
	lower := strings.ToLower(command)
	if strings.Contains(lower, "hopper") || strings.Contains(lower, "decompile") || strings.Contains(lower, "debug") || strings.Contains(lower, "gdb") || strings.Contains(lower, "rizin") {
		return "REA komutu yok. Statik inceleme: inspect-artifact, js, pe, doctor."
	}
	command, rel = foldREA(command, rel)
	path, err := reaPath(root, rel)
	if err != nil {
		return "Dosya klasörün dışında."
	}
	if _, statErr := os.Stat(path); statErr != nil {
		return "Dosya yok."
	}
	args, ok := reaArgv(command, path)
	if !ok {
		return "REA komutu yok. Statik inceleme: inspect-artifact, js, pe, doctor."
	}
	bin, prefix, err := reaBinary()
	if err != nil {
		return "REA yok. Kurulum: npx rea-agents@" + reaAgentsVersion()
	}
	// Safe exec: trusted binary resolution, env allowlist, own process
	// group (tree killed on timeout), rlimits / Job Object, 90s timeout.
	stdout, _, _, runErr := safeRun(context.Background(), root, root, append([]string{bin}, append(prefix, args...)...), reaTimeout, nil)
	out := []byte(stdout)
	text := trimRunes(strings.TrimSpace(redactLine(string(out))), 6000)
	if text == "" {
		if runErr != nil {
			return "REA incelemesi bitmedi."
		}
		return "REA çıktı vermedi."
	}
	return text
}

func foldREA(command, rel string) (string, string) {
	lower := strings.ToLower(command)
	if strings.Contains(lower, "rea.tools") || strings.Contains(lower, "rea-agents") {
		rel = firstLocalREA(command, rel)
		if strings.Contains(lower, ".js") || strings.Contains(lower, "javascript") {
			return "js", rel
		}
		return "inspect-artifact", rel
	}
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "inspect-artifact", rel
	}
	head := strings.ToLower(strings.Trim(fields[0], "/"))
	rest := fields[1:]
	if head == "rea" {
		if len(rest) == 0 {
			return "inspect-artifact", rel
		}
		head = strings.ToLower(strings.Trim(rest[0], "/"))
		rest = rest[1:]
	}
	if !knownREA(head) {
		if strings.TrimSpace(rel) == "" {
			return "inspect-artifact", head
		}
		return head, rel
	}
	if strings.TrimSpace(rel) == "" && len(rest) > 0 {
		rel = rest[0]
	}
	return head, rel
}

func firstLocalREA(command, rel string) string {
	if strings.TrimSpace(rel) != "" {
		return rel
	}
	for _, field := range strings.Fields(command) {
		low := strings.ToLower(field)
		if strings.Contains(low, "rea.tools") || strings.Contains(low, "rea-agents") || strings.Contains(field, "://") {
			continue
		}
		return field
	}
	return ""
}

func knownREA(command string) bool {
	switch command {
	case "inspect-artifact", "incele", "inspect", "doctor", "js", "javascript", "analyze-javascript-application", "pe", "inspect-pe-imports", "duzen", "düzen", "layout", "inspect-binary-layout":
		return true
	default:
		return false
	}
}

func reaArgv(command, path string) ([]string, bool) {
	limit := []string{"--format", "md", "--token-limit", "1200"}
	switch command {
	case "", "incele", "inspect", "inspect-artifact":
		return append([]string{"inspect-artifact", path}, limit...), true
	case "doctor":
		return append([]string{"doctor", "--target", path}, limit...), true
	case "js", "javascript", "analyze-javascript-application":
		return append([]string{"analyze-javascript-application", path}, limit...), true
	case "pe", "inspect-pe-imports":
		return append([]string{"inspect-pe-imports", path}, limit...), true
	case "duzen", "düzen", "layout", "inspect-binary-layout":
		return append([]string{"inspect-binary-layout", path}, limit...), true
	default:
		return nil, false
	}
}

func reaPath(root, rel string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	rel = strings.TrimSpace(rel)
	if rel == "" || rel == "." {
		return root, nil
	}
	full := rel
	if !filepath.IsAbs(rel) {
		full = filepath.Join(root, rel)
	}
	full, err = filepath.Abs(full)
	if err != nil {
		return "", err
	}
	if full != root && !insideRoot(root, full) {
		return "", os.ErrInvalid
	}
	return full, nil
}

func reaBinary() (string, []string, error) {
	if bin := strings.TrimSpace(os.Getenv("US_REA")); bin != "" {
		return bin, nil, nil
	}
	if path, err := exec.LookPath("rea"); err == nil {
		return path, nil, nil
	}
	npx, err := exec.LookPath("npx")
	if err != nil {
		return "", nil, err
	}
	return npx, []string{"--yes", "rea-agents@" + reaAgentsVersion()}, nil
}

// reaAgentsVersionDefault: pinned instead of @latest. The version in use
// could not be determined offline (no lockfile / npx cache entry on the
// build box) — FLAG: confirm the vetted version and set it here or via
// US_REA_AGENTS_VERSION (strict semver only).
const reaAgentsVersionDefault = "1.0.0"

const reaTimeout = 90 * time.Second

var reaSemver = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

func reaAgentsVersion() string {
	if v := strings.TrimSpace(os.Getenv("US_REA_AGENTS_VERSION")); reaSemver.MatchString(v) {
		return v
	}
	return reaAgentsVersionDefault
}

// reaEnv is the shared env allowlist (previously only three keys were removed).
func reaEnv() []string { return safeExecEnv() }

func trimRunes(text string, max int) string {
	if utf8.RuneCountInString(text) <= max {
		return text
	}
	count := 0
	for i := range text {
		if count == max {
			return text[:i] + "\nkesildi."
		}
		count++
	}
	return text
}
