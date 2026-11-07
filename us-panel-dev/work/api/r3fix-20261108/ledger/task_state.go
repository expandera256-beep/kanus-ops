package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// Persistent, versioned, atomic JSON state for the agent build/fix loop so a
// task can resume after a crash or restart without re-applying edits,
// re-running completed phases, or silently replaying a paid model call.

const taskStateVersion = 1

type taskState struct {
	Version      int           `json:"version"`
	ID           string        `json:"id"`
	Root         string        `json:"root"`
	Model        string        `json:"model"`
	Phase        string        `json:"phase"` // checking, model_call, applying, done
	Result       repairResult  `json:"result"`
	PendingEdits []pendingEdit `json:"pendingEdits,omitempty"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

type pendingEdit struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Hash    string `json:"hash"`
	Applied bool   `json:"applied"`
}

var taskIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func taskStatePath(root, id string) (string, error) {
	if !taskIDPattern.MatchString(id) {
		return "", errors.New("geçersiz görev kimliği")
	}
	return filepath.Join(root, taskDir, "agent", id+".json"), nil
}

func saveTaskState(st *taskState) error {
	path, err := taskStatePath(st.Root, st.ID)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	for _, d := range []string{filepath.Join(st.Root, taskDir), dir} {
		if info, err := os.Lstat(d); err == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
			return errors.New("görev klasörü geçersiz")
		}
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	st.Version = taskStateVersion
	st.UpdatedAt = time.Now().UTC()
	raw, err := json.MarshalIndent(st, "", " ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".state-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0o600); err == nil {
		_, err = f.Write(raw)
	}
	if err == nil {
		err = f.Sync()
	}
	if c := f.Close(); err == nil {
		err = c
	}
	if err != nil {
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	syncDir(dir)
	return nil
}

// loadTaskState returns (nil, nil) when no state exists. A corrupt or
// foreign file is preserved as <id>.json.corrupt-<ts> and reported.
func loadTaskState(root, id string) (*taskState, error) {
	path, err := taskStatePath(root, id)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil || !info.Mode().IsRegular() || info.Size() > 16<<20 {
		return nil, errors.New("görev durumu okunamadı")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var st taskState
	if json.Unmarshal(raw, &st) != nil || st.Version != taskStateVersion || st.ID != id || st.Root != root {
		keep := fmt.Sprintf("%s.corrupt-%d", path, time.Now().UnixNano())
		_ = os.Rename(path, keep)
		return nil, fmt.Errorf("görev durumu bozuk; korundu: %s", filepath.Base(keep))
	}
	return &st, nil
}

// errUncertainModelCall: the process stopped while a model request was in
// flight. It is not replayed automatically because it may have been billed.
var errUncertainModelCall = errors.New("önceki model isteğinin sonucu belirsiz; otomatik tekrar yapılmadı")

// runPersistentRepair runs (or resumes) a repair loop whose every state
// change is checkpointed atomically. allowReplay must be set explicitly by
// the user to re-issue an in-flight model call after a crash.
func runPersistentRepair(ctx context.Context, l *repairLoop, id string, allowReplay bool) (*repairResult, error) {
	st, err := loadTaskState(l.Root, id)
	if err != nil {
		return nil, err
	}
	if st == nil {
		st = &taskState{ID: id, Root: l.Root, Model: l.Model, Phase: "checking"}
	} else {
		switch {
		case st.Phase == "done":
			r := st.Result
			return &r, nil
		case st.Phase == "model_call" && !allowReplay:
			return nil, errUncertainModelCall
		case st.Phase == "model_call":
			// User approved a replay: the uncounted call stays counted.
		}
		if len(st.PendingEdits) > 0 {
			ft, err := newFileTools(l.Root)
			if err != nil {
				return nil, err
			}
			for i := range st.PendingEdits {
				e := &st.PendingEdits[i]
				cur, err := ft.Read(e.Path)
				if err == nil && hashContent(cur) == e.Hash {
					e.Applied = true // already on disk: never write twice
					continue
				}
				if _, err := recordChange(ft.Root, filepath.ToSlash(e.Path), []byte(e.Content)); err != nil {
					return nil, err
				}
				if err := ft.Write(e.Path, []byte(e.Content)); err != nil {
					return nil, err
				}
				e.Applied = true
			}
			var names []string
			for _, e := range st.PendingEdits {
				names = append(names, e.Path)
			}
			st.Result.Steps = append(st.Result.Steps, repairStep{Iteration: st.Result.Iterations, State: "applied", ExitCode: -1, Edits: names, Note: "devam: bekleyen düzeltme tamamlandı"})
			st.PendingEdits = nil
		}
		st.Phase = "checking"
		if err := saveTaskState(st); err != nil {
			return nil, err
		}
		resume := st.Result
		l.resume = &resume
	}
	l.hooks = &repairHooks{
		beforeModel: func(r *repairResult) error {
			st.Result, st.Phase = *r, "model_call"
			return saveTaskState(st)
		},
		beforeApply: func(r *repairResult, edits []repairEdit) error {
			st.Result, st.Phase = *r, "applying"
			st.PendingEdits = nil
			for _, e := range edits {
				st.PendingEdits = append(st.PendingEdits, pendingEdit{Path: e.Path, Content: e.Content, Hash: hashContent([]byte(e.Content))})
			}
			return saveTaskState(st)
		},
	}
	l.Checkpoint = func(r *repairResult) error {
		st.Result = *r
		st.PendingEdits = nil
		st.Phase = "checking"
		switch r.Status {
		case "running":
		default:
			st.Phase = "done"
		}
		return saveTaskState(st)
	}
	return l.Run(ctx)
}
