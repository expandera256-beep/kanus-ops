package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// This lock serializes proposal commits inside this process. External editors
// and process crashes are not covered by it.
var proposalBatchMu sync.Mutex

type proposalBatchItem struct {
	file   proposedFile
	full   string
	record string
}

func checkProposalVersion(root string, f proposedFile) (string, error) {
	full, rel, ok := replyTarget(root, f.Path)
	if !ok || rel != f.Path {
		return "", errors.New("geçersiz dosya yolu")
	}
	info, err := os.Lstat(full)
	if !f.Existed {
		if os.IsNotExist(err) {
			return full, nil
		}
		return "", fmt.Errorf("yeni dosya artık mevcut veya erişilemiyor: %s", f.Path)
	}
	if hashContent([]byte(f.Old)) != f.OldHash {
		return "", fmt.Errorf("yedek içeriği geçersiz: %s", f.Path)
	}
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxBackupBytes {
		return "", fmt.Errorf("dosya erişilemiyor: %s", f.Path)
	}
	b, err := os.ReadFile(full)
	if err != nil || hashContent(b) != f.OldHash {
		return "", fmt.Errorf("dosya değişmiş: %s", f.Path)
	}
	return full, nil
}

// Preflight and snapshot every file before the first write. If a write fails,
// attempt a conflict-aware rollback and retain history for anything unresolved.
// It is not an atomic multi-file filesystem transaction or crash recovery.
func applyProposalBatch(root string, files []proposedFile, write func(string, string, []byte) error) (string, error) {
	proposalBatchMu.Lock()
	defer proposalBatchMu.Unlock()
	if len(files) == 0 {
		return "", errors.New("boş öneri")
	}
	if pendingTransaction(root) {
		return "", errors.New("önce kesilen dosya işlemini kurtarın")
	}
	if err := checkProposedPython(files); err != nil {
		return "", err
	}
	items := make([]proposalBatchItem, 0, len(files))
	seen := map[string]bool{}
	for _, f := range files {
		if seen[strings.ToLower(f.Path)] || proposalHasReadDirective(f.New) {
			return "", errors.New("yinelenen veya geçersiz dosya önerisi")
		}
		seen[strings.ToLower(f.Path)] = true
		full, err := checkProposalVersion(root, f)
		if err != nil {
			return "", err
		}
		items = append(items, proposalBatchItem{file: f, full: full})
	}
	cleanup := func() {
		for _, item := range items {
			if item.record != "" {
				_ = os.Remove(item.record)
			}
		}
	}
	for i := range items {
		rec, err := recordChange(root, items[i].file.Path, []byte(items[i].file.New))
		if err != nil {
			cleanup()
			return "", fmt.Errorf("yedek oluşturulamadı: %w", err)
		}
		items[i].record = rec
		if _, err = checkProposalVersion(root, items[i].file); err != nil {
			cleanup()
			return "", err
		}
	}
	if err := beginTransaction(root, items); err != nil {
		cleanup()
		return "", err
	}
	rollback := func(last int) error {
		var failed []string
		for i := last; i >= 0; i-- {
			it := items[i]
			if err := noSymlinkPath(root, it.full); err != nil {
				failed = append(failed, it.file.Path)
				continue
			}
			current, err := os.ReadFile(it.full)
			if it.file.Existed && err == nil && hashContent(current) == it.file.OldHash {
				continue
			}
			if !it.file.Existed && os.IsNotExist(err) {
				continue
			}
			if err != nil || hashContent(current) != hashContent([]byte(it.file.New)) {
				failed = append(failed, it.file.Path)
				continue
			}
			if it.file.Existed {
				err = writeModelFile(root, it.full, []byte(it.file.Old))
			} else {
				err = os.Remove(it.full)
			}
			if err != nil {
				failed = append(failed, it.file.Path)
			}
		}
		if len(failed) > 0 {
			return fmt.Errorf("geri alma tamamlanamadı; geçmiş kayıtları korundu: %s", strings.Join(failed, ", "))
		}
		if err := os.Remove(transactionPath(root)); err != nil {
			return errors.New("dosyalar geri alındı ama işlem kaydı kaldırılamadı")
		}
		cleanup()
		return nil
	}
	for i, it := range items {
		if _, err := checkProposalVersion(root, it.file); err != nil {
			if r := rollback(i - 1); r != nil {
				return "", fmt.Errorf("%v; %v", err, r)
			}
			return "", err
		}
		err := write(root, it.full, []byte(it.file.New))
		if err == nil {
			var written []byte
			written, err = os.ReadFile(it.full)
			if err == nil && hashContent(written) != hashContent([]byte(it.file.New)) {
				err = errors.New("yazılan dosya doğrulanamadı")
			}
		}
		if err != nil {
			if r := rollback(i); r != nil {
				return "", fmt.Errorf("%v; %v", err, r)
			}
			return "", fmt.Errorf("%s: %w; önceki yazımlar geri alındı", it.file.Path, err)
		}
	}
	if err := os.Remove(transactionPath(root)); err != nil {
		return "", errors.New("dosyalar yazıldı ancak işlem kaydı kaldırılamadı; kurtarmayı inceleyin")
	}
	names := make([]string, 0, len(items))
	for _, it := range items {
		names = append(names, filepath.ToSlash(it.file.Path))
	}
	return "Model dosyaya yazdı: " + strings.Join(names, ", "), nil
}
