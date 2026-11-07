package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepairWithoutFileProposalIsNotMarkedCompleted(t *testing.T) {
	t.Setenv("US_DESK", filepath.Join(t.TempDir(), "desk.json"))
	s := newView("q8")
	s.root = t.TempDir()
	s.key = "test-key"
	s.reviewBeforeApply = true
	if err := os.WriteFile(filepath.Join(s.root, "main.py"), []byte("print('broken')\n"), 0600); err != nil {
		t.Fatal(err)
	}
	calls := 0
	s.sendModel = func(_ workerID, _, wire string) modelResult {
		calls++
		if calls == 1 {
			return modelResult{ok: true, body: `{"version":1,"operations":[{"op":"read_file","path":"main.py"}]}`}
		}
		return modelResult{ok: true, body: "Dosya düzeltildi."}
	}
	s.askModel("projeyi onar", intent{})
	records := listTaskRecords(s.root)
	if len(records) != 1 || records[0].Status != "failed" || s.proposal != nil {
		t.Fatalf("history=%+v", records)
	}
	if calls != 2 {
		t.Fatalf("unexpected calls %d", calls)
	}
}

func TestTaskHistoryPersistsUniqueIDsAndStatus(t *testing.T) {
	root := t.TempDir()
	a, err := newTaskRecord(root, "same prompt")
	if err != nil {
		t.Fatal(err)
	}
	b, err := newTaskRecord(root, "same prompt")
	if err != nil {
		t.Fatal(err)
	}
	if a.ID == b.ID || a.PromptHash != b.PromptHash {
		t.Fatal("task identity not independent")
	}
	if err = updateTaskRecord(root, a, "failed", "model yanıt vermedi"); err != nil {
		t.Fatal(err)
	}
	rows := listTaskRecords(root)
	if len(rows) != 2 {
		t.Fatalf("records: %+v", rows)
	}
	loaded := findTaskRecord(root, a.ID)
	if loaded == nil || loaded.Status != "failed" {
		t.Fatalf("lost status: %+v", loaded)
	}
}

func TestTaskHistoryRejectsSymlinkDirectory(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, taskDir)); err != nil {
		t.Skip("symlink unavailable")
	}
	if _, err := newTaskRecord(root, "task"); err == nil {
		t.Fatal("symlink accepted")
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatalf("outside modified: %v %v", entries, err)
	}
}

func TestTaskHistoryTracksProposalUntilApplied(t *testing.T) {
	t.Setenv("US_DESK", filepath.Join(t.TempDir(), "desk.json"))
	s := newView("q8")
	s.root = t.TempDir()
	s.key = "test-key"
	s.reviewBeforeApply = true
	s.modelCh = make(chan modelResult, 2)
	s.sendModel = func(workerID, string, string) modelResult {
		return modelResult{ok: true, body: "DOSYA output.txt\nhello\nBİTTİ"}
	}
	s.draft = "dosya oluştur"
	s.submit()
	if s.proposal == nil || s.proposal.TaskID == "" {
		t.Fatal("proposal not associated with task")
	}
	got := findTaskRecord(s.root, s.proposal.TaskID)
	if got == nil || got.Status != "awaiting_review" {
		t.Fatalf("state: %+v", got)
	}
	id := s.proposal.TaskID
	s.applyProposal()
	got = findTaskRecord(s.root, id)
	if got == nil || got.Status != "completed" {
		t.Fatalf("after apply: %+v notice=%s", got, s.notice)
	}
}
