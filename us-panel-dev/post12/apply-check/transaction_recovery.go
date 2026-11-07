package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const transactionDir = ".us-transaction"

type transactionItem struct {
	Path    string
	Existed bool
	Old     string
	OldHash string
	NewHash string
}
type transactionRecord struct {
	Root  string
	Files []transactionItem
}

func transactionPath(root string) string { return filepath.Join(root, transactionDir, "active.json") }

// Write the recovery description before any target is touched. An existing
// journal blocks new transactions until an operator recovers it.
func beginTransaction(root string, items []proposalBatchItem) error {
	dir := filepath.Join(root, transactionDir)
	st, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		if err = os.Mkdir(dir, 0700); err != nil && !os.IsExist(err) {
			return err
		}
		st, err = os.Lstat(dir)
	}
	if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return errors.New("geçersiz işlem kayıt klasörü")
	}
	rec := transactionRecord{Root: root}
	for _, it := range items {
		rec.Files = append(rec.Files, transactionItem{Path: it.file.Path, Existed: it.file.Existed, Old: it.file.Old, OldHash: it.file.OldHash, NewHash: hashContent([]byte(it.file.New))})
	}
	data, err := json.Marshal(rec)
	if err != nil || len(data) > 20<<20 {
		return errors.New("işlem kaydı hazırlanamadı")
	}
	f, err := os.OpenFile(transactionPath(root), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if os.IsExist(err) {
		return errors.New("önce kesilen dosya işlemini kurtarın")
	}
	if err != nil {
		return err
	}
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	if e := f.Close(); err == nil {
		err = e
	}
	if err != nil {
		_ = os.Remove(transactionPath(root))
		return err
	}
	return nil
}

func pendingTransaction(root string) bool {
	_, err := os.Lstat(transactionPath(root))
	return err == nil || !os.IsNotExist(err)
}

// This is recovery, not cross-file atomic visibility. External writers and
// power-loss durability need stronger OS/filesystem guarantees.
func recoverProposalBatch(root string) error {
	proposalBatchMu.Lock()
	defer proposalBatchMu.Unlock()
	dir := filepath.Join(root, transactionDir)
	st, err := os.Lstat(dir)
	if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return errors.New("kurtarılacak işlem kaydı yok")
	}
	path := transactionPath(root)
	st, err = os.Lstat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() > 20<<20 {
		return errors.New("işlem kaydı okunamıyor")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(io.LimitReader(f, 20<<20+1))
	_ = f.Close()
	if err != nil || len(data) > 20<<20 {
		return errors.New("işlem kaydı okunamıyor")
	}
	var rec transactionRecord
	if json.Unmarshal(data, &rec) != nil || rec.Root != root || len(rec.Files) == 0 || len(rec.Files) > 100 {
		return errors.New("işlem kaydı geçersiz")
	}
	seen := map[string]bool{}
	for _, item := range rec.Files {
		key := strings.ToLower(item.Path)
		if seen[key] {
			return errors.New("yinelenen işlem yolu")
		}
		seen[key] = true
		full, rel, ok := replyTarget(root, item.Path)
		if !ok || rel != item.Path || len(item.NewHash) != 64 || item.Existed && hashContent([]byte(item.Old)) != item.OldHash {
			return errors.New("işlem kaydı geçersiz")
		}
		current, e := os.ReadFile(full)
		if item.Existed && e == nil && hashContent(current) == item.OldHash {
			continue
		}
		if !item.Existed && os.IsNotExist(e) {
			continue
		}
		if e != nil || hashContent(current) != item.NewHash {
			return fmt.Errorf("dosya dışarıdan değişmiş: %s", item.Path)
		}
	}
	for i := len(rec.Files) - 1; i >= 0; i-- {
		item := rec.Files[i]
		full, _, ok := replyTarget(root, item.Path)
		if !ok {
			return errors.New("hedef değişti")
		}
		current, e := os.ReadFile(full)
		if item.Existed && e == nil && hashContent(current) == item.OldHash {
			continue
		}
		if !item.Existed && os.IsNotExist(e) {
			continue
		}
		if e != nil || hashContent(current) != item.NewHash {
			return fmt.Errorf("dosya dışarıdan değişmiş: %s", item.Path)
		}
		if item.Existed {
			e = writeModelFile(root, full, []byte(item.Old))
		} else {
			e = os.Remove(full)
		}
		if e != nil {
			return fmt.Errorf("%s kurtarılamadı: %w", item.Path, e)
		}
	}
	return os.Remove(path)
}
