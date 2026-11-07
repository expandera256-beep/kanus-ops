package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const changeHistoryDir = ".us-history"
const maxBackupBytes = 8 << 20

type changeRecord struct {
	Path    string `json:"path"`
	Existed bool   `json:"existed"`
	Old     []byte `json:"old,omitempty"`
	Mode    uint32 `json:"mode,omitempty"`
	NewHash string `json:"new_hash"`
}

func hashContent(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

// Refuse symlinks at every existing component, including the destination.
func noSymlinkPath(root, full string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	absFull, err := filepath.Abs(full)
	if err != nil || !insideRoot(absRoot, absFull) || absRoot == absFull {
		return errors.New("klasör dışı yol")
	}
	rootInfo, err := os.Lstat(absRoot)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return errors.New("geçersiz çalışma klasörü")
	}
	rel, err := filepath.Rel(absRoot, absFull)
	if err != nil {
		return err
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) > 0 && (strings.EqualFold(parts[0], changeHistoryDir) || strings.EqualFold(parts[0], proposalDir) || strings.EqualFold(parts[0], taskDir) || strings.EqualFold(parts[0], transactionDir)) {
		return errors.New("özel klasör korunuyor")
	}
	cur := absRoot
	for _, p := range parts {
		cur = filepath.Join(cur, p)
		info, err := os.Lstat(cur)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("bağlantı içeren yol reddedildi")
		}
	}
	return nil
}

// Snapshot creation is mandatory before model output changes a file.
func recordChange(root, rel string, next []byte) (string, error) {
	h, err := pinRoot(root)
	if err != nil {
		return "", err
	}
	defer h.Close()
	ft := &fileTools{Root: root, root: h}
	full, err := ft.resolve(rel)
	if err != nil {
		return "", err
	}
	rel, err = filepath.Rel(root, full)
	if err != nil {
		return "", err
	}
	rec := changeRecord{Path: filepath.ToSlash(rel), NewHash: hashContent(next)}
	f, err := openBeneathAt(h, root, rel)
	if err == nil {
		info, e := f.Stat()
		if e != nil || !info.Mode().IsRegular() || info.Size() > maxBackupBytes {
			f.Close()
			return "", errors.New("yedeklenemeyen dosya")
		}
		rec.Old, err = io.ReadAll(io.LimitReader(f, maxBackupBytes+1))
		f.Close()
		if err != nil {
			return "", err
		}
		if len(rec.Old) > maxBackupBytes {
			return "", errors.New("yedeklenemeyen dosya")
		}
		rec.Existed, rec.Mode = true, uint32(info.Mode().Perm())
	} else if !os.IsNotExist(err) {
		return "", err
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return "", err
	}
	name := filepath.Join(changeHistoryDir, fmt.Sprintf("%020d-%016x.json", time.Now().UnixNano(), rand.Uint64()))
	if err = writeBeneathAt(h, root, name, data); err != nil {
		return "", err
	}
	return filepath.Join(root, name), nil
}

func writeModelFile(root, full string, data []byte) error {
	return atomicWriteFile(root, full, data)
}

func undoLastChange(root string) (string, error) {
	h, err := pinRoot(root)
	if err != nil {
		return "", err
	}
	defer h.Close()
	entries, err := readDirBeneathAt(h, changeHistoryDir)
	if err != nil {
		return "", errors.New("geri alınacak kayıt yok")
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && e.Type()&os.ModeSymlink == 0 && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return "", errors.New("geri alınacak kayıt yok")
	}
	sort.Strings(names)
	record := filepath.Join(changeHistoryDir, names[len(names)-1])
	f, err := openBeneathAt(h, root, record)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(io.LimitReader(f, maxBackupBytes*2+1))
	f.Close()
	if err != nil || len(data) > maxBackupBytes*2 {
		return "", errors.New("geçersiz geçmiş kaydı")
	}
	var rec changeRecord
	if json.Unmarshal(data, &rec) != nil {
		return "", errors.New("geçersiz geçmiş kaydı")
	}
	full, rel, ok := replyTarget(root, rec.Path)
	if !ok || rel != rec.Path {
		return "", errors.New("geçersiz hedef")
	}
	relative, err := filepath.Rel(root, full)
	if err != nil {
		return "", err
	}
	currentFile, err := openBeneathAt(h, root, relative)
	if err != nil {
		return "", err
	}
	current, err := io.ReadAll(io.LimitReader(currentFile, maxBackupBytes+1))
	currentFile.Close()
	if err != nil || len(current) > maxBackupBytes || hashContent(current) != rec.NewHash {
		return "", errors.New("dosya sonradan değişmiş; geri alma yapılmadı")
	}
	if rec.Existed {
		err = writeBeneathAt(h, root, relative, rec.Old)
	} else {
		err = deleteBeneathAt(h, relative)
	}
	if err != nil {
		return "", err
	}
	if err = deleteBeneathAt(h, record); err != nil {
		return "", errors.New("dosya geri alındı ama geçmiş kaydı kaldırılamadı")
	}
	return rel, nil
}
