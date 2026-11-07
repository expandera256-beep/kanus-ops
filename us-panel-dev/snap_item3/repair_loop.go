package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Budgeted build -> error -> fix -> retest loop. The model is any
// OpenAI-compatible /v1/chat/completions endpoint. Edits are applied only
// through root-confined atomic fileTools and only after Approve says yes.

type repairBudget struct {
	MaxIterations int           // build+fix rounds
	MaxModelCalls int           // paid model calls
	MaxDuration   time.Duration // wall clock for the whole loop
	StepTimeout   time.Duration // one build/test command
	MaxEditBytes  int           // total bytes the model may write per round
}

var defaultRepairBudget = repairBudget{MaxIterations: 4, MaxModelCalls: 4, MaxDuration: 15 * time.Minute, StepTimeout: 3 * time.Minute, MaxEditBytes: 256 << 10}

type repairEdit struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type repairStep struct {
	Iteration int      `json:"iteration"`
	Argv      []string `json:"argv"`
	State     string   `json:"state"`
	ExitCode  int      `json:"exitCode"`
	Stdout    string   `json:"stdout"`
	Stderr    string   `json:"stderr"`
	Edits     []string `json:"edits,omitempty"`
	Note      string   `json:"note,omitempty"`
}

type repairResult struct {
	Status     string       `json:"status"` // fixed, already_ok, budget_exhausted, cancelled, rejected, model_error
	Iterations int          `json:"iterations"`
	ModelCalls int          `json:"modelCalls"`
	Steps      []repairStep `json:"steps"`
}

type repairLoop struct {
	Root     string
	Endpoint string // base URL, e.g. https://api-test.kanus.org
	Key      string
	Model    string
	Budget   repairBudget
	Client   *http.Client
	// Approve must return true for the edits to be written. nil = refuse.
	Approve func(iteration int, edits []repairEdit) bool
	// Checkpoint, if set, is called after every state change (item 4).
	Checkpoint func(r *repairResult) error
	// Resume state (item 4): steps already done are not repeated.
	resume *repairResult
}

// repairCommands picks the build/test argv for a project.
func repairCommands(root string) ([][]string, error) {
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
		return [][]string{{"go", "build", "./..."}, {"go", "vet", "./..."}, {"go", "test", "./..."}}, nil
	}
	return nil, errors.New("desteklenen derleme bulunamadı (go.mod yok)")
}

func (l *repairLoop) runChecks(ctx context.Context, iteration int) (repairStep, bool) {
	cmds, err := repairCommands(l.Root)
	if err != nil {
		return repairStep{Iteration: iteration, State: "start_failed", ExitCode: -1, Note: err.Error()}, false
	}
	var last repairStep
	for _, argv := range cmds {
		bin, err := lookPathSafe(argv[0])
		if err != nil {
			return repairStep{Iteration: iteration, Argv: argv, State: "start_failed", ExitCode: -1, Note: err.Error()}, false
		}
		sc, err := newSafeCmd(ctx, l.Root, l.Root, append([]string{bin}, argv[1:]...), l.Budget.StepTimeout)
		if err != nil {
			return repairStep{Iteration: iteration, Argv: argv, State: "start_failed", ExitCode: -1, Note: err.Error()}, false
		}
		state, code, _ := sc.Run()
		last = repairStep{Iteration: iteration, Argv: argv, State: state, ExitCode: code, Stdout: redactLine(sc.Stdout.String()), Stderr: redactLine(sc.Stderr.String())}
		if state != "success" {
			return last, false
		}
	}
	return last, true
}

func (l *repairLoop) Run(parent context.Context) (*repairResult, error) {
	b := l.Budget
	if b.MaxIterations <= 0 {
		b = defaultRepairBudget
	}
	l.Budget = b
	ft, err := newFileTools(l.Root)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(parent, b.MaxDuration)
	defer cancel()
	res := &repairResult{Status: "running"}
	if l.resume != nil {
		*res = *l.resume
		res.Status = "running"
	}
	save := func() {
		if l.Checkpoint != nil {
			_ = l.Checkpoint(res)
		}
	}
	finish := func(status string) (*repairResult, error) {
		res.Status = status
		save()
		return res, nil
	}
	for {
		if ctx.Err() != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) && parent.Err() == nil {
				return finish("budget_exhausted")
			}
			return finish("cancelled")
		}
		step, ok := l.runChecks(ctx, res.Iterations)
		res.Steps = append(res.Steps, step)
		save()
		if ok {
			if res.Iterations == 0 {
				return finish("already_ok")
			}
			return finish("fixed")
		}
		if step.State == "cancelled" || ctx.Err() != nil {
			if parent.Err() != nil {
				return finish("cancelled")
			}
			return finish("budget_exhausted")
		}
		if step.State == "start_failed" {
			return finish("start_failed")
		}
		if res.Iterations >= b.MaxIterations || res.ModelCalls >= b.MaxModelCalls {
			return finish("budget_exhausted")
		}
		res.Iterations++
		res.ModelCalls++
		save()
		edits, err := l.askFix(ctx, ft, step)
		if err != nil {
			res.Steps = append(res.Steps, repairStep{Iteration: res.Iterations, State: "model_error", ExitCode: -1, Note: err.Error()})
			if ctx.Err() != nil && parent.Err() != nil {
				return finish("cancelled")
			}
			return finish("model_error")
		}
		total := 0
		for _, e := range edits {
			total += len(e.Content)
			if _, err := ft.resolve(e.Path); err != nil {
				res.Steps = append(res.Steps, repairStep{Iteration: res.Iterations, State: "rejected", ExitCode: -1, Note: "model kök dışı yol önerdi: " + e.Path})
				return finish("rejected")
			}
		}
		if total > b.MaxEditBytes {
			res.Steps = append(res.Steps, repairStep{Iteration: res.Iterations, State: "rejected", ExitCode: -1, Note: "düzeltme boyut bütçesini aşıyor"})
			return finish("rejected")
		}
		if l.Approve == nil || !l.Approve(res.Iterations, edits) {
			res.Steps = append(res.Steps, repairStep{Iteration: res.Iterations, State: "rejected", ExitCode: -1, Note: "kullanıcı onayı yok"})
			return finish("rejected")
		}
		var names []string
		for _, e := range edits {
			if _, err := recordChange(ft.Root, filepath.ToSlash(e.Path), []byte(e.Content)); err != nil {
				return nil, fmt.Errorf("yedek alınamadı: %w", err)
			}
			if err := ft.Write(e.Path, []byte(e.Content)); err != nil {
				return nil, fmt.Errorf("düzeltme yazılamadı: %w", err)
			}
			names = append(names, e.Path)
		}
		res.Steps = append(res.Steps, repairStep{Iteration: res.Iterations, State: "applied", ExitCode: -1, Edits: names})
		save()
	}
}

const repairSystemPrompt = `You fix build/test errors in a project. Reply with ONLY a JSON object: {"files":[{"path":"relative/path","content":"full new file content"}]}. Paths are relative to the project root.`

func (l *repairLoop) askFix(ctx context.Context, ft *fileTools, step repairStep) ([]repairEdit, error) {
	var prompt strings.Builder
	fmt.Fprintf(&prompt, "Komut: %s\nÇıkış kodu: %d\nstdout:\n%s\nstderr:\n%s\n\nDosyalar:\n", strings.Join(step.Argv, " "), step.ExitCode, step.Stdout, step.Stderr)
	budget := 64 << 10
	_ = filepath.WalkDir(ft.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil || budget <= 0 {
			return nil
		}
		if d.IsDir() {
			if path != ft.Root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if !(strings.HasSuffix(name, ".go") || name == "go.mod") {
			return nil
		}
		rel, _ := filepath.Rel(ft.Root, path)
		body, err := ft.Read(rel)
		if err != nil || len(body) > budget {
			return nil
		}
		budget -= len(body)
		fmt.Fprintf(&prompt, "--- %s\n%s\n", filepath.ToSlash(rel), redactLine(string(body)))
		return nil
	})
	req := map[string]any{
		"model":    l.Model,
		"stream":   false,
		"messages": []map[string]string{{"role": "system", "content": repairSystemPrompt}, {"role": "user", "content": prompt.String()}},
	}
	raw, _ := json.Marshal(req)
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(l.Endpoint, "/")+"/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	hreq.Header.Set("Content-Type", "application/json")
	if l.Key != "" {
		hreq.Header.Set("Authorization", "Bearer "+l.Key)
	}
	client := l.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Minute}
	}
	resp, err := client.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("model HTTP %d", resp.StatusCode)
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &out); err != nil || len(out.Choices) == 0 {
		return nil, errors.New("model yanıtı çözümlenemedi")
	}
	return parseRepairEdits(out.Choices[0].Message.Content)
}

func parseRepairEdits(content string) ([]repairEdit, error) {
	content = strings.TrimSpace(content)
	if i := strings.Index(content, "{"); i >= 0 {
		if j := strings.LastIndex(content, "}"); j > i {
			content = content[i : j+1]
		}
	}
	var payload struct {
		Files []repairEdit `json:"files"`
	}
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return nil, errors.New("model düzeltme JSON'u geçersiz")
	}
	if len(payload.Files) == 0 || len(payload.Files) > 20 {
		return nil, errors.New("model düzeltme önermedi")
	}
	return payload.Files, nil
}

func lookPathSafe(name string) (string, error) {
	bin, ok := resolveBin(name)
	if !ok {
		return "", fmt.Errorf("%s bulunamadı", name)
	}
	return bin, nil
}
