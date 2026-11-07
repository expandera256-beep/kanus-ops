package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const changeHistoryDir = ".us-history"
const maxBackupBytes = 8 << 20

type changeRecord struct {
	Path string `json:"path"`
	Existed bool `json:"existed"`
	Old []byte `json:"old,omitempty"`
	Mode uint32 `json:"mode,omitempty"`
	NewHash string `json:"new_hash"`
}

func hashContent(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

// Refuse symlinks at every existing component, including the destination.
func noSymlinkPath(root, full string) error {
	absRoot, err := filepath.Abs(root); if err != nil { return err }
	absFull, err := filepath.Abs(full); if err != nil || !insideRoot(absRoot, absFull) || absRoot == absFull { return errors.New("klasör dışı yol") }
	rootInfo, err := os.Lstat(absRoot); if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 { return errors.New("geçersiz çalışma klasörü") }
	rel, err := filepath.Rel(absRoot, absFull); if err != nil { return err }
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) > 0 && (strings.EqualFold(parts[0], changeHistoryDir) || strings.EqualFold(parts[0], proposalDir) || strings.EqualFold(parts[0], taskDir) || strings.EqualFold(parts[0], transactionDir)) { return errors.New("özel klasör korunuyor") }
	cur := absRoot
	for _, p := range parts {
		cur = filepath.Join(cur, p)
		info, err := os.Lstat(cur)
		if os.IsNotExist(err) { continue }
		if err != nil { return err }
		if info.Mode()&os.ModeSymlink != 0 { return errors.New("bağlantı içeren yol reddedildi") }
	}
	return nil
}

// Snapshot creation is mandatory before model output changes a file.
func recordChange(root, rel string, next []byte) (string, error) {
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := noSymlinkPath(root, full); err != nil { return "", err }
	rec := changeRecord{Path:rel, NewHash:hashContent(next)}
	info, err := os.Lstat(full)
	if err == nil {
		if !info.Mode().IsRegular() || info.Size() > maxBackupBytes { return "", errors.New("yedeklenemeyen dosya") }
		rec.Old, err = os.ReadFile(full); if err != nil { return "", err }
		rec.Existed, rec.Mode = true, uint32(info.Mode().Perm())
	} else if !os.IsNotExist(err) { return "", err }
	dir := filepath.Join(root, changeHistoryDir)
	if info, err := os.Lstat(dir); err == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) { return "", errors.New("geçersiz geçmiş klasörü") } else if err != nil && !os.IsNotExist(err) { return "", err }
	if err := os.Mkdir(dir, 0700); err != nil && !os.IsExist(err) { return "", err }
	data, err := json.Marshal(rec); if err != nil { return "", err }
	f, err := os.CreateTemp(dir, fmt.Sprintf("%020d-*.json", time.Now().UnixNano())); if err != nil { return "", err }
	name := f.Name()
	if err = f.Chmod(0600); err == nil { _, err = f.Write(data) }
	if closeErr := f.Close(); err == nil { err = closeErr }
	if err != nil { _ = os.Remove(name); return "", err }
	return name, nil
}

func writeModelFile(root, full string, data []byte) error {
	return atomicWriteFile(root, full, data)
}

func undoLastChange(root string) (string, error) {
	dir := filepath.Join(root, changeHistoryDir)
	info, err := os.Lstat(dir); if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 { return "", errors.New("geri alınacak kayıt yok") }
	entries, err := os.ReadDir(dir); if err != nil { return "", err }
	var names []string
	for _, e := range entries { if !e.IsDir() && e.Type()&os.ModeSymlink == 0 && strings.HasSuffix(e.Name(), ".json") { names = append(names,e.Name()) } }
	if len(names)==0 { return "", errors.New("geri alınacak kayıt yok") }; sort.Strings(names)
	file := filepath.Join(dir,names[len(names)-1]); info, err = os.Lstat(file)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxBackupBytes*2 { return "", errors.New("geçersiz geçmiş kaydı") }
	data, err := os.ReadFile(file); if err != nil { return "", err }
	var rec changeRecord; if json.Unmarshal(data,&rec) != nil { return "", errors.New("geçersiz geçmiş kaydı") }
	full, rel, ok := replyTarget(root,rec.Path); if !ok || rel != rec.Path { return "", errors.New("geçersiz hedef") }
	current, err := os.ReadFile(full); if err != nil || len(current)>maxBackupBytes || hashContent(current)!=rec.NewHash { return "", errors.New("dosya sonradan değişmiş; geri alma yapılmadı") }
	if rec.Existed {
		f, err := os.CreateTemp(filepath.Dir(full), ".us-restore-*"); if err != nil { return "", err }
		tmp := f.Name(); defer os.Remove(tmp)
		if err = f.Chmod(os.FileMode(rec.Mode)); err == nil { _, err = f.Write(rec.Old) }; if closeErr := f.Close(); err == nil { err = closeErr }; if err != nil { return "", err }
		if err = os.Rename(tmp,full); err != nil { return "", err }
	} else if err = os.Remove(full); err != nil { return "", err }
	if err = os.Remove(file); err != nil { return "", errors.New("dosya geri alındı ama geçmiş kaydı kaldırılamadı") }
	return rel,nil
}
