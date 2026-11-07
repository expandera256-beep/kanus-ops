package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const proposalDir = ".us-pending"

type proposedFile struct {
	Path    string
	New     string
	Old     string
	Existed bool
	OldHash string
}
type fileProposal struct {
	Root   string
	Files  []proposedFile
	TaskID string `json:"task_id,omitempty"`
}

// Read requests are protocol instructions, never part of a proposed file.
func proposalHasReadDirective(body string) bool {
	for _, line := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "OKU ") || strings.HasPrefix(line, "OKU:") {
			return true
		}
	}
	return false
}

func prepareProposal(root, reply string, allowed map[string]bool, hashes map[string]string) (*fileProposal, error) {
	return prepareProposalWithEvidence(root, reply, allowed, hashes, nil)
}

func prepareProposalWithEvidence(root, reply string, allowed map[string]bool, hashes map[string]string, evidence map[string]patchEvidence) (*fileProposal, error) {
	if env, yes, err := parseModelActions(reply); yes {
		if err != nil {
			return nil, err
		}
		return prepareActionProposal(root, env, allowed, hashes, evidence)
	}
	p := &fileProposal{Root: root}
	seen := map[string]bool{}
	var name string
	var lines []string
	flush := func() error {
		if name == "" {
			return nil
		}
		full, rel, ok := replyTarget(root, name)
		if !ok || seen[rel] {
			return errors.New("geçersiz veya yinelenen dosya yolu")
		}
		seen[rel] = true
		body := strings.TrimSpace(strings.Join(lines, "\n"))
		if body == "" {
			return errors.New("boş dosya önerisi")
		}
		if proposalHasReadDirective(body) {
			return errors.New("dosya gövdesinde OKU komutu var; öneri reddedildi")
		}
		item := proposedFile{Path: rel, New: body + "\n"}
		info, err := os.Lstat(full)
		if err == nil {
			if !info.Mode().IsRegular() || info.Size() > maxBackupBytes {
				return errors.New("dosya incelenemiyor")
			}
			if allowed != nil && !allowed[rel] {
				return fmt.Errorf("tam okunmamış dosya: %s", rel)
			}
			b, e := os.ReadFile(full)
			if e != nil {
				return e
			}
			item.Existed = true
			item.Old = string(b)
			item.OldHash = hashContent(b)
			if h, ok := hashes[rel]; ok && h != item.OldHash {
				return fmt.Errorf("dosya okunduktan sonra değişti: %s", rel)
			}
			if _, ok := hashes[rel]; !ok {
				return fmt.Errorf("mevcut dosya okunmadı: %s", rel)
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		p.Files = append(p.Files, item)
		name = ""
		lines = nil
		return nil
	}
	for _, line := range strings.Split(strings.ReplaceAll(expandFences(reply), "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(line)
		n, ok := strings.CutPrefix(t, "DOSYA ")
		if !ok {
			n, ok = strings.CutPrefix(t, "DOSYA:")
		}
		if ok {
			if err := flush(); err != nil {
				return nil, err
			}
			name = strings.TrimSpace(n)
			continue
		}
		if isJobMark(t) {
			if err := flush(); err != nil {
				return nil, err
			}
			continue
		}
		if name != "" {
			lines = append(lines, line)
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	if len(p.Files) == 0 {
		return nil, nil
	}
	return p, nil
}

func proposalPreview(p *fileProposal) string {
	if p == nil {
		return ""
	}
	var b strings.Builder
	for _, f := range p.Files {
		fmt.Fprintf(&b, "DOSYA %s\n", f.Path)
		if f.Existed {
			b.WriteString("ÖNCE:\n")
			b.WriteString(f.Old)
		} else {
			b.WriteString("YENİ DOSYA\n")
		}
		b.WriteString("\nSONRA:\n")
		b.WriteString(f.New)
		b.WriteString("\n")
	}
	out := b.String()
	if len(out) > 256*1024 {
		clip, _ := clipUTF8(out, 256*1024)
		return clip + "\n[Önizleme kesildi; değişiklik uygulanamaz.]"
	}
	return out
}

func (s *session) applyProposal() {
	p := s.proposal
	if p == nil {
		s.notice = "Bekleyen değişiklik yok."
		return
	}
	if p.Root != s.root {
		s.notice = "Çalışma klasörü değişti; öneri uygulanmadı."
		return
	}
	for _, f := range p.Files {
		if proposalHasReadDirective(f.New) {
			s.notice = "Dosya gövdesinde OKU komutu var; öneri uygulanmadı."
			return
		}
	}
	if strings.Contains(proposalPreview(p), "[Önizleme kesildi; değişiklik uygulanamaz.]") {
		s.notice = "Öneri incelenemeyecek kadar büyük; daha küçük parçalara bölün."
		return
	}
	rec := findTaskRecord(s.root, p.TaskID)
	if rec != nil {
		if err := updateTaskRecord(s.root, rec, "applying", "Dosyalar doğrulanıp uygulanıyor."); err != nil {
			s.notice = "Görev uygulama durumu kaydedilemedi; dosyalar değiştirilmedi."
			return
		}
	}
	note, err := applyProposalBatch(s.root, p.Files, writeModelFile)
	if err != nil {
		if rec != nil {
			_ = updateTaskRecord(s.root, rec, "failed", err.Error())
		}
		s.notice = "Öneri uygulanamadı: " + err.Error()
		return
	}
	if rec != nil {
		if err := updateTaskRecord(s.root, rec, "completed", "Öneri uygulandı; proje testleri ayrıca çalıştırılmalıdır."); err != nil {
			s.notice = "Dosyalar uygulandı; görev durumu kaydedilemedi."
			s.noteLog("işlem", s.notice)
			return
		}
	}
	s.proposal = nil
	_ = removeProposal(s.root)
	s.notice = "Değişiklikler uygulandı: " + note
	s.noteLog("işlem", s.notice)
}

func proposalPath(root string) string { return filepath.Join(root, proposalDir, "pending.json") }
func saveProposal(p *fileProposal) error {
	dir := filepath.Join(p.Root, proposalDir)
	if info, err := os.Lstat(dir); err == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
		return errors.New("geçersiz öneri klasörü")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Mkdir(dir, 0700); err != nil && !os.IsExist(err) {
		return err
	}
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "pending-*")
	if err != nil {
		return err
	}
	temp := f.Name()
	defer os.Remove(temp)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(data)
	}
	if e := f.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	return os.Rename(temp, proposalPath(p.Root))
}
func removeProposal(root string) error {
	err := os.Remove(proposalPath(root))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
func loadProposal(root string) *fileProposal {
	dir := filepath.Join(root, proposalDir)
	d, e := os.Lstat(dir)
	if e != nil || !d.IsDir() || d.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	path := proposalPath(root)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 20<<20 {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var p fileProposal
	if json.Unmarshal(data, &p) != nil || p.Root != root || len(p.Files) == 0 {
		return nil
	}
	return &p
}
