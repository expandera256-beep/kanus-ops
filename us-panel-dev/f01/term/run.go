package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type launch struct {
	Dir     string
	Argv    []string
	Label   string
	Missing string
}

type programNote struct {
	gen    int
	status string
	line   string
}

func wantsRun(task string) bool {
	for _, field := range strings.FieldsFunc(foldText(task), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		switch field {
		case "calistir", "calistirin", "baslat", "baslatin", "run", "start":
			return true
		}
	}
	return false
}

// pythonEnv is the filtered allowlist env (no parent secrets).
func pythonEnv() []string { return safeExecEnv("PYTHONIOENCODING=utf-8") }

func (s *session) stopProgram() {
	if s.proc == nil {
		return
	}
	_ = killProcessTree(s.proc)
	s.proc = nil
}

func (s *session) startProgram(root, extra string) string {
	s.runGen++
	gen := s.runGen
	s.stopProgram()
	s.runStatus = ""
	s.runStarted = false
	if err := sandboxPreflight(); err != nil {
		s.runStatus = "durum  durdu"
		return s.runStatus + "\n" + err.Error()
	}
	found, ok := detectLaunch(root)
	if !ok {
		return "Çalıştırılacak program yok."
	}
	if found.Missing != "" {
		s.runStatus = statusReport("durdu")
		return formatRun(found.Label, found.Missing, "durdu")
	}
	argv := append([]string{}, found.Argv...)
	extra = strings.TrimSpace(extra)
	if extra != "" && commandSafe(append(argv, extra)) {
		argv = append(argv, extra)
	}
	sc, err := newSafeCmd(context.Background(), root, found.Dir, argv, runProgramTimeout, "PYTHONIOENCODING=utf-8")
	if err != nil {
		s.runStatus = "durum  durdu"
		return s.runStatus + "\n" + found.Label + " reddedildi: " + err.Error()
	}
	out := func() string { return runOutput(sc.Stdout, sc.Stderr) }
	if err := sc.Start(); err != nil {
		s.runStatus = "durum  durdu"
		return s.runStatus + "\n" + found.Label + " başlamadı: " + err.Error()
	}
	s.runStarted = true
	exited := make(chan error, 1)
	go func() {
		_, _, err := sc.Wait()
		exited <- err
	}()
	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(3 * time.Second)
	for {
		select {
		case err := <-exited:
			state := "bitti"
			if err != nil {
				state = "durdu"
			}
			s.runStatus = statusReport(state)
			return formatRun(found.Label, out(), state)
		case <-ticker.C:
			if programReady(out()) {
				s.proc = sc.Cmd.Process
				s.runStatus = statusReport("çalışıyor")
				s.watchExit(exited, gen, found.Label, out)
				return formatRun(found.Label, out(), "çalışıyor")
			}
		case <-deadline:
			s.proc = sc.Cmd.Process
			s.runStatus = statusReport("çalışıyor")
			s.watchExit(exited, gen, found.Label, out)
			return formatRun(found.Label, out(), "çalışıyor")
		}
	}
}

// runProgramTimeout bounds a Run-path program's lifetime (servers included).
const runProgramTimeout = 4 * time.Hour

// runOutput keeps stdout and stderr separate in the panel view.
func runOutput(stdout, stderr *cappedBuf) string {
	o, e := stdout.String(), stderr.String()
	if strings.TrimSpace(e) == "" {
		return o
	}
	if strings.TrimSpace(o) == "" {
		return "[stderr]\n" + e
	}
	return o + "\n[stderr]\n" + e
}

func (s *session) watchExit(exited <-chan error, gen int, label string, out func() string) {
	ch := s.procDone
	go func() {
		err := <-exited
		state := "bitti"
		if err != nil {
			state = "durdu"
		}
		note := programNote{gen: gen, status: statusReport(state), line: formatRun(label, out(), state)}
		if ch == nil {
			return
		}
		select {
		case ch <- note:
		default:
		}
	}()
}

func (s *session) applyProgram(note programNote) {
	if note.gen != s.runGen || note.status == "" {
		return
	}
	s.runStatus = note.status
	s.proc = nil
	for i := len(s.flow) - 1; i >= 0; i-- {
		if s.flow[i].title == "çalıştır" {
			s.flow[i].body = note.line
			s.project.RunLine = note.line
			return
		}
	}
}

func statusReport(state string) string {
	switch state {
	case "bitti", "durdu":
		return "durum  çalışıyor - " + state
	default:
		return "durum  " + state
	}
}

func formatRun(label, output, state string) string {
	lines := []string{statusReport(state), label + " " + state + "."}
	output = redactLine(strings.TrimSpace(output))
	if state == "durdu" {
		// Python tracebacks describe the actual exception at the end.
		parts := strings.Split(output, "\n")
		if len(parts) > 8 {
			parts = parts[len(parts)-8:]
		}
		output = strings.Join(parts, "\n")
	}
	n := 0
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		n++
		if (state == "durdu" && n == 8) || (state != "durdu" && n == 3) {
			break
		}
	}
	return strings.Join(lines, "\n")
}

func programReady(output string) bool {
	lower := strings.ToLower(output)
	for _, mark := range []string{"dinliyor", "uvicorn running", "running on", "listening", "started server"} {
		if strings.Contains(lower, mark) {
			return true
		}
	}
	return false
}

func detectLaunch(root string) (launch, bool) {
	if strings.TrimSpace(root) == "" {
		return launch{}, false
	}
	if found, ok := launchFromDocker(root); ok {
		return found, true
	}
	if found, ok := launchFromPython(root); ok {
		return found, true
	}
	if found, ok := launchFromGo(root); ok {
		return found, true
	}
	if found, ok := launchFromPackage(root); ok {
		return found, true
	}
	return launch{}, false
}

func launchFromDocker(root string) (launch, bool) {
	var hits []string
	walkProject(root, func(path string, entry os.DirEntry) {
		if !entry.IsDir() && entry.Name() == "Dockerfile" {
			hits = append(hits, path)
		}
	})
	sortShallow(hits)
	for _, path := range hits {
		body, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		argv := dockerArgv(string(body))
		if !commandSafe(argv) {
			continue
		}
		bin, ok := resolveBin(argv[0])
		if !ok {
			continue
		}
		dir := filepath.Dir(path)
		next := append([]string{bin}, argv[1:]...)
		if !commandExists(dir, next) || !insideRoot(root, dir) {
			continue
		}
		return launch{Dir: dir, Argv: next, Label: shortCommand(argv)}, true
	}
	return launch{}, false
}

func launchFromPython(root string) (launch, bool) {
	priority := map[string]int{"main.py": 0, "app.py": 1, "run.py": 2, "manage.py": 3, "asgi.py": 4, "wsgi.py": 5}
	bestScore := 1 << 30
	bestPath := ""
	walkProject(root, func(path string, entry os.DirEntry) {
		if entry.IsDir() {
			return
		}
		rank, ok := priority[strings.ToLower(entry.Name())]
		if !ok {
			return
		}
		score := depth(root, path)*10 + rank
		if score < bestScore {
			bestScore = score
			bestPath = path
		}
	})
	if bestPath == "" {
		return launch{}, false
	}
	dir := filepath.Dir(bestPath)
	name := filepath.Base(bestPath)
	if !insideRoot(root, dir) {
		return launch{}, false
	}
	bin, prefix, ok := findPython(root, dir)
	if !ok {
		return launch{Dir: dir, Label: "python3 " + name, Missing: "Python yok. python.org üzerinden Python 3 kur, sonra yeniden çalıştır."}, true
	}
	argv := append([]string{bin}, prefix...)
	argv = append(argv, name)
	return launch{Dir: dir, Argv: argv, Label: "python3 " + name}, true
}

func launchFromPackage(root string) (launch, bool) {
	var hits []string
	walkProject(root, func(path string, entry os.DirEntry) {
		if !entry.IsDir() && entry.Name() == "package.json" {
			hits = append(hits, path)
		}
	})
	sortShallow(hits)
	npm, ok := resolveBin("npm")
	if !ok {
		return launch{}, false
	}
	for _, path := range hits {
		body, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(body, &pkg) != nil {
			continue
		}
		start := strings.TrimSpace(pkg.Scripts["start"])
		if start == "" || strings.ContainsAny(start, ";|&`$") || strings.Contains(start, "..") {
			continue
		}
		dir := filepath.Dir(path)
		if !insideRoot(root, dir) {
			continue
		}
		return launch{Dir: dir, Argv: []string{npm, "start"}, Label: "npm start"}, true
	}
	return launch{}, false
}

func dockerArgv(text string) []string {
	var entry, cmd []string
	for _, line := range strings.Split(text, "\n") {
		if cut := strings.Index(line, "#"); cut >= 0 {
			line = line[:cut]
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		key := strings.ToUpper(fields[0])
		if key != "CMD" && key != "ENTRYPOINT" {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, fields[0]))
		argv := splitDocker(rest)
		if len(argv) == 0 {
			continue
		}
		if key == "ENTRYPOINT" {
			entry = argv
		} else {
			cmd = argv
		}
	}
	if len(entry) > 0 && len(cmd) > 0 {
		return append(entry, cmd...)
	}
	if len(cmd) > 0 {
		return cmd
	}
	return entry
}

func splitDocker(rest string) []string {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return nil
	}
	if strings.HasPrefix(rest, "[") {
		var argv []string
		if json.Unmarshal([]byte(rest), &argv) != nil {
			return nil
		}
		return argv
	}
	return strings.Fields(rest)
}

func commandSafe(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	for _, arg := range argv {
		if arg == "-c" || strings.ContainsAny(arg, ";|&`$") || strings.Contains(arg, "..") || strings.Contains(arg, "\n") {
			return false
		}
	}
	_, ok := resolveBin(argv[0])
	return ok
}

func commandExists(dir string, argv []string) bool {
	base := strings.ToLower(filepath.Base(argv[0]))
	if base == "npm" {
		return true
	}
	for _, arg := range argv[1:] {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if strings.Contains(arg, ":") {
			mod := strings.Split(arg, ":")[0]
			rel := filepath.Join(strings.Split(mod, ".")...) + ".py"
			if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
				return false
			}
			continue
		}
		lower := strings.ToLower(arg)
		if strings.HasSuffix(lower, ".py") || strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".php") {
			if _, err := os.Stat(filepath.Join(dir, arg)); err != nil {
				return false
			}
		}
	}
	return true
}

func resolveBin(name string) (string, bool) {
	path, _, ok := findBin(name)
	return path, ok
}

func findPython(root, dir string) (string, []string, bool) {
	if path := projectPython(root, dir); path != "" && pythonRuns(path, nil) {
		return path, nil, true
	}
	path, prefix, ok := findBin("python3")
	return path, prefix, ok
}

func findBin(name string) (string, []string, bool) {
	base := strings.ToLower(filepath.Base(name))
	var candidates []string
	switch base {
	case "python", "python3", "py":
		candidates = []string{"python3", "python", "py"}
	case "uvicorn", "flask", "gunicorn", "node", "npm", "php", "go":
		candidates = []string{base}
	default:
		return "", nil, false
	}
	python := base == "python" || base == "python3" || base == "py"
	for _, cand := range candidates {
		path, err := exec.LookPath(cand)
		if err != nil || storeAlias(path) {
			continue
		}
		prefix := pythonPrefix(path)
		if python && !pythonRuns(path, prefix) {
			continue
		}
		return path, prefix, true
	}
	if python {
		for _, path := range installedPythons() {
			if storeAlias(path) || !pythonRuns(path, nil) {
				continue
			}
			return path, nil, true
		}
	}
	return "", nil, false
}

func pythonPrefix(path string) []string {
	base := strings.ToLower(filepath.Base(path))
	if base == "py" || base == "py.exe" {
		return []string{"-3"}
	}
	return nil
}

func storeAlias(path string) bool {
	clean := strings.ToLower(strings.ReplaceAll(path, `\`, `/`))
	slash := strings.LastIndex(clean, "/")
	if slash < 0 {
		return false
	}
	base := clean[slash+1:]
	dir := clean[:slash]
	if cut := strings.LastIndex(dir, "/"); cut >= 0 {
		dir = dir[cut+1:]
	}
	if dir != "windowsapps" {
		return false
	}
	return base == "python.exe" || base == "python3.exe" || base == "python" || base == "python3"
}

func pythonRuns(path string, prefix []string) bool {
	argv := append(append([]string{}, prefix...), "--version")
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()
	out, state, _, _ := safeRun(ctx, neutralDir(), "", append([]string{path}, argv...), 1200*time.Millisecond, nil)
	if state != "success" {
		return false
	}
	return pythonVersionOK(out)
}

func pythonVersionOK(out string) bool {
	lower := strings.ToLower(out)
	if strings.Contains(lower, "bulunamad") || strings.Contains(lower, "microsoft store") || strings.Contains(lower, "was not found") {
		return false
	}
	return strings.Contains(lower, "python 3") || strings.Contains(lower, "python 2")
}

func installedPythons() []string {
	var roots []string
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		roots = append(roots, filepath.Join(local, "Programs", "Python"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, filepath.Join(home, "AppData", "Local", "Programs", "Python"))
	}
	for _, env := range []string{"PROGRAMFILES", "PROGRAMFILES(X86)"} {
		if dir := os.Getenv(env); dir != "" {
			roots = append(roots, dir)
		}
	}
	seen := map[string]bool{}
	type hit struct {
		path string
		rank int
	}
	var hits []hit
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || pythonRank(entry.Name()) == 0 {
				continue
			}
			for _, file := range []string{"python.exe", "python3.exe", "python"} {
				path := filepath.Join(root, entry.Name(), file)
				if seen[path] {
					continue
				}
				info, err := os.Stat(path)
				if err != nil || info.IsDir() {
					continue
				}
				seen[path] = true
				hits = append(hits, hit{path: path, rank: pythonRank(entry.Name())})
			}
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].rank > hits[j].rank })
	out := make([]string, 0, len(hits))
	for _, item := range hits {
		out = append(out, item.path)
	}
	return out
}

func pythonRank(name string) int {
	digits := ""
	for _, r := range name {
		if r >= '0' && r <= '9' {
			digits += string(r)
		}
	}
	if digits == "" {
		return 0
	}
	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0
	}
	return n
}

// projectPython is disabled: a project-local interpreter (.venv/bin/python)
// is project-controlled code. Running it (even for --version) leaked the
// panel environment (PANEL DEV HIGH-1/HIGH-3). trustedBin also refuses any
// executable inside the workspace. Use a system Python on PATH instead.
func projectPython(root, dir string) string {
	if true {
		return ""
	}
	candidates := []string{
		filepath.Join(dir, ".venv", "bin", "python"),
		filepath.Join(dir, "venv", "bin", "python"),
		filepath.Join(dir, ".venv", "Scripts", "python.exe"),
		filepath.Join(dir, "venv", "Scripts", "python.exe"),
		filepath.Join(root, ".venv", "bin", "python"),
		filepath.Join(root, "venv", "bin", "python"),
		filepath.Join(root, ".venv", "Scripts", "python.exe"),
		filepath.Join(root, "venv", "Scripts", "python.exe"),
	}
	for _, cand := range candidates {
		info, err := os.Stat(cand)
		if err != nil || info.IsDir() || !insideRoot(root, cand) {
			continue
		}
		return cand
	}
	return ""
}

func shortCommand(argv []string) string {
	parts := []string{filepath.Base(argv[0])}
	parts = append(parts, argv[1:]...)
	return strings.Join(parts, " ")
}

func walkProject(root string, visit func(string, os.DirEntry)) {
	if refusedFolder(root) {
		return
	}
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := entry.Name()
		if entry.IsDir() {
			if path != root && skipRunDir(name) {
				return filepath.SkipDir
			}
			if depth(root, path) > 4 {
				return filepath.SkipDir
			}
			return nil
		}
		visit(path, entry)
		return nil
	})
}

func skipRunDir(name string) bool {
	return skipTreeDir(name)
}

func skipTreeDir(name string) bool {
	switch strings.ToLower(name) {
	case "node_modules", "vendor", "__pycache__", ".venv", "venv", "dist", "build", ".git":
		return true
	}
	return strings.HasPrefix(name, ".") || systemDirName(name)
}

func depth(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator))
}

func sortShallow(paths []string) {
	for i := 1; i < len(paths); i++ {
		for j := i; j > 0; j-- {
			if depthCompare(paths[j], paths[j-1]) {
				paths[j], paths[j-1] = paths[j-1], paths[j]
			}
		}
	}
}

func depthCompare(a, b string) bool {
	da := strings.Count(a, string(filepath.Separator))
	db := strings.Count(b, string(filepath.Separator))
	if da != db {
		return da < db
	}
	return a < b
}

type cappedBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (c *cappedBuf) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := len(p)
	const maxOutput = 16 << 10
	if len(p) >= maxOutput {
		c.b.Reset()
		_, _ = c.b.Write(p[len(p)-maxOutput:])
		return n, nil
	}
	if overflow := c.b.Len() + len(p) - maxOutput; overflow > 0 {
		c.b.Next(overflow)
	}
	_, _ = c.b.Write(p)
	return n, nil
}

func (c *cappedBuf) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.b.String()
}

func launchFromGo(root string) (launch, bool) {
	if _, _, err := sourceSnapshot(root, "go.mod"); err != nil {
		return launch{}, false
	}
	if _, _, err := sourceSnapshot(root, "main.go"); err != nil {
		return launch{}, false
	}
	bin, err := exec.LookPath("go")
	if err != nil {
		return launch{Dir: root, Label: "go run .", Missing: "Go bulunamadı; Go kurulumunu kontrol edin."}, true
	}
	return launch{Dir: root, Argv: []string{bin, "run", "."}, Label: "go run ."}, true
}
