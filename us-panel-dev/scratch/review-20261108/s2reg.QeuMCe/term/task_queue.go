package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type queuedTask struct {
	ID         string `json:"id"`
	Prompt     string `json:"prompt"`
	PromptHash string `json:"prompt_hash"`
	State      string `json:"state"`
	Created    string `json:"created"`
	Updated    string `json:"updated"`
	Owner      string `json:"owner,omitempty"`
	TaskID     string `json:"task_id,omitempty"`
}
type queuedTaskView struct {
	ID      string `json:"id"`
	State   string `json:"state"`
	Created string `json:"created"`
	TaskID  string `json:"task_id,omitempty"`
}

var requestKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{8,128}$`)

func newQueueOwner() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(b[:])
}

func queueDirectory(root string) (string, error) {
	if _, err := taskHistoryDir(root); err != nil {
		return "", err
	}
	dir := filepath.Join(root, taskDir, "queue")
	st, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		if err = os.Mkdir(dir, 0700); err != nil && !os.IsExist(err) {
			return "", err
		}
		st, err = os.Lstat(dir)
	}
	if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("geçersiz görev kuyruğu")
	}
	return dir, nil
}

func queueFile(root, id string) (string, error) {
	if len(id) != 64 {
		return "", errors.New("geçersiz kuyruk kimliği")
	}
	if _, err := hex.DecodeString(id); err != nil {
		return "", err
	}
	dir, err := queueDirectory(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, id+".json"), nil
}

func loadQueuedTask(root, id string) (*queuedTask, error) {
	path, err := queueFile(root, id)
	if err != nil {
		return nil, err
	}
	st, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() || st.Size() > 16<<20 {
		return nil, errors.New("geçersiz kuyruk kaydı")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var job queuedTask
	if json.Unmarshal(b, &job) != nil || job.ID != id || job.PromptHash != hashContent([]byte(job.Prompt)) || len(job.Prompt) > longInputLimit {
		return nil, errors.New("bozuk kuyruk kaydı")
	}
	return &job, nil
}

func enqueueTask(root, key, prompt string) (*queuedTask, bool, error) {
	if !requestKeyPattern.MatchString(key) || prompt == "" || len(prompt) > longInputLimit || !utf8.ValidString(prompt) || strings.IndexByte(prompt, 0) >= 0 {
		return nil, false, errors.New("geçersiz görev isteği")
	}
	if redactLine(prompt) != prompt {
		return nil, false, errors.New("kalıcı görev gizli anahtar içeriyor; anahtarı istemden çıkarın")
	}
	id := hashContent([]byte(key))
	path, err := queueFile(root, id)
	if err != nil {
		return nil, false, err
	}
	if existing, err := loadQueuedTask(root, id); err == nil {
		if existing.PromptHash != hashContent([]byte(prompt)) {
			return nil, false, errors.New("aynı istek kimliği farklı istemle kullanılamaz")
		}
		return existing, false, nil
	} else if !os.IsNotExist(err) {
		return nil, false, err
	}
	jobs, err := listQueuedTasks(root)
	if err != nil {
		return nil, false, err
	}
	pending := 0
	for _, j := range jobs {
		if j.State == "queued" || j.State == "running" {
			pending++
		}
	}
	if pending >= 100 {
		return nil, false, errors.New("görev kuyruğu dolu")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	job := &queuedTask{ID: id, Prompt: prompt, PromptHash: hashContent([]byte(prompt)), State: "queued", Created: now, Updated: now}
	b, err := json.Marshal(job)
	if err != nil {
		return nil, false, err
	}
	if len(b) > (16<<20)-1024 {
		return nil, false, errors.New("kalıcı görev kaydı çok büyük; daha küçük metin kullanın")
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, false, err
	}
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	if e := f.Close(); err == nil {
		err = e
	}
	if err != nil {
		_ = os.Remove(path)
		return nil, false, err
	}
	return job, true, nil
}

func saveQueuedTask(root string, job *queuedTask) error {
	path, err := queueFile(root, job.ID)
	if err != nil {
		return err
	}
	if _, err = loadQueuedTask(root, job.ID); err != nil {
		return err
	}
	job.Updated = time.Now().UTC().Format(time.RFC3339Nano)
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}
	if len(b) > 16<<20 {
		return errors.New("kalıcı görev kaydı çok büyük")
	}
	f, err := os.CreateTemp(filepath.Dir(path), "queue-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(b)
	}
	if err == nil {
		err = f.Sync()
	}
	if e := f.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func listQueuedTasks(root string) ([]queuedTask, error) {
	dir := filepath.Join(root, taskDir, "queue")
	for _, path := range []string{root, filepath.Join(root, taskDir), dir} {
		st, err := os.Lstat(path)
		if os.IsNotExist(err) {
			return nil, nil
		}
		if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("geçersiz görev kuyruğu")
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var jobs []queuedTask
	for _, e := range entries {
		if len(e.Name()) != 69 || e.IsDir() || e.Type()&os.ModeSymlink != 0 {
			continue
		}
		job, err := loadQueuedTask(root, e.Name()[:64])
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, *job)
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].Created < jobs[j].Created })
	return jobs, nil
}

func (g *guiServer) queueView() []queuedTaskView {
	jobs, err := listQueuedTasks(g.s.root)
	if err != nil {
		return nil
	}
	var out []queuedTaskView
	for _, j := range jobs {
		state := j.State
		if state == "running" && j.Owner != g.owner {
			state = "interrupted"
		}
		if j.TaskID != "" {
			if rec := findTaskRecord(g.s.root, j.TaskID); rec != nil && (rec.Status == "completed" || rec.Status == "cancelled") {
				state = rec.Status
			}
		}
		out = append(out, queuedTaskView{ID: j.ID, State: state, Created: j.Created, TaskID: j.TaskID})
	}
	if len(out) > 100 {
		out = out[len(out)-100:]
	}
	return out
}

// One worker runs per local GUI. Interrupted work requires explicit resumption;
// it is never silently resubmitted after a restart (which could charge again).
func (g *guiServer) runQueue(root string) {
	g.mu.Lock()
	if g.queueRunning {
		g.mu.Unlock()
		return
	}
	g.queueRunning = true
	g.mu.Unlock()
	for {
		g.mu.Lock()
		unlock := g.s.hold()
		stop := func() { g.queueRunning = false; unlock(); g.mu.Unlock() }
		if g.s.root != root || g.s.calling || g.s.testing || g.s.proposal != nil {
			stop()
			return
		}
		jobs, err := listQueuedTasks(root)
		if err != nil {
			g.s.notice = "Görev kuyruğu okunamadı: " + err.Error()
			stop()
			return
		}
		var job *queuedTask
		for i := range jobs {
			if jobs[i].State == "queued" {
				copy := jobs[i]
				job = &copy
				break
			}
		}
		if job == nil {
			stop()
			return
		}
		job.State = "running"
		job.Owner = g.owner
		if err = saveQueuedTask(root, job); err != nil {
			g.s.notice = "Görev başlatılmadan kaydedilemedi."
			stop()
			return
		}
		previousID := g.s.lastTaskID
		g.s.draft = job.Prompt
		g.s.cursor = len([]rune(job.Prompt))
		g.s.writing = true
		g.mu.Unlock()
		g.s.submit()
		g.s.finishFlow()
		unlock()
		g.mu.Lock()
		unlock = g.s.hold()
		job.State = "finished"
		if g.s.lastTaskID != previousID {
			job.TaskID = g.s.lastTaskID
			if rec := findTaskRecord(g.s.root, job.TaskID); rec != nil {
				job.State = rec.Status
			}
		}
		if err = saveQueuedTask(root, job); err != nil {
			g.s.notice = "Görev sonucu kaydedilemedi; yeniden göndermeden kayıtları inceleyin."
			g.queueRunning = false
			unlock()
			g.mu.Unlock()
			return
		}
		if g.s.proposal != nil {
			g.queueRunning = false
			unlock()
			g.mu.Unlock()
			return
		}
		unlock()
		g.mu.Unlock()
	}
}
