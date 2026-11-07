package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func wantsSingleExe(text string) bool {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "pyinstaller") || strings.Contains(lower, "onefile") {
		return true
	}
	return strings.Contains(lower, "tek exe") || strings.Contains(lower, "exe haline")
}

func (s *session) keepBuild(note string) {
	note = strings.TrimSpace(note)
	if note == "" {
		note = "Exe yazılmadı."
	}
	s.review = note
	s.reaNote = note
	s.notice = ""
	s.draft = ""
	s.cursor = 0
	s.writing = false
	s.flow = append(s.flow, event{title: "exe", meta: "yerel", body: trimRunes(note, 700), kind: "yerel"})
	s.noteLog("işlem", "tek exe")
	s.noteLog("düşünce", "Model çağrılmadı. "+note)
	s.ran = true
}

func buildSingleExe(root string) string {
	if refusedFolder(root) {
		return "Bu klasör taranmaz."
	}
	spec := findSpec(root)
	if spec == "" {
		return "PyInstaller spec yok. Model çağrılmadı."
	}
	bin, args, err := pyinstallerCmd(spec)
	if err != nil {
		return "PyInstaller yok. Kurulum: python -m pip install pyinstaller. Model çağrılmadı."
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	out, _, _, runErr := safeRun(ctx, root, filepath.Dir(spec), append([]string{bin}, args...), 6*time.Minute, nil)
	if out == "" && runErr != nil {
		out = runErr.Error()
	}
	text := trimRunes(strings.TrimSpace(redactLine(out)), 4000)
	exe := newestExe(filepath.Join(filepath.Dir(spec), "dist"))
	if exe != "" {
		rel := relPath(root, exe)
		line := "Tek exe yazıldı. " + rel
		if text != "" {
			line += "\n" + text
		}
		return line
	}
	if text == "" {
		if runErr != nil {
			return "PyInstaller bitmedi. Model çağrılmadı."
		}
		return "PyInstaller çıktı vermedi."
	}
	if runErr != nil {
		return "PyInstaller durdu.\n" + text
	}
	return text
}

func findSpec(root string) string {
	root, err := filepath.Abs(root)
	if err != nil || refusedFolder(root) {
		return ""
	}
	var found []string
	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		if depth > 3 || len(found) >= 8 || refusedFolder(dir) {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			name := entry.Name()
			if skipTreeDir(name) {
				continue
			}
			full := filepath.Join(dir, name)
			if entry.IsDir() {
				walk(full, depth+1)
				continue
			}
			if strings.HasSuffix(strings.ToLower(name), ".spec") {
				found = append(found, full)
			}
		}
	}
	walk(root, 0)
	if len(found) == 0 {
		return ""
	}
	best := found[0]
	for _, path := range found[1:] {
		if len(path) < len(best) {
			best = path
		}
	}
	return best
}

func pyinstallerCmd(spec string) (string, []string, error) {
	if bin := strings.TrimSpace(os.Getenv("US_PYINSTALLER")); bin != "" {
		return bin, []string{"--noconfirm", "--clean", spec}, nil
	}
	if path, err := exec.LookPath("pyinstaller"); err == nil {
		return path, []string{"--noconfirm", "--clean", spec}, nil
	}
	for _, name := range []string{"python3", "python", "py"} {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		return path, []string{"-m", "PyInstaller", "--noconfirm", "--clean", spec}, nil
	}
	return "", nil, os.ErrNotExist
}

func newestExe(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var best string
	var bestTime time.Time
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".exe") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if best == "" || info.ModTime().After(bestTime) {
			best = filepath.Join(dir, entry.Name())
			bestTime = info.ModTime()
		}
	}
	return best
}
