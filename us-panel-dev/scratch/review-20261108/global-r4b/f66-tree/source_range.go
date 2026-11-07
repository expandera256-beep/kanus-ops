package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode/utf8"
)

const sourceSizeLimit = 8 << 20

type sourceRangeRequest struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256,omitempty"`
	Offset    int    `json:"offset,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
}

type sourceChunk struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	Size       int    `json:"size"`
	Offset     int    `json:"offset"`
	NextOffset int    `json:"next_offset"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
	TotalLines int    `json:"total_lines"`
	Complete   bool   `json:"complete"`
	Redacted   bool   `json:"redacted"`
	Truncated  bool   `json:"truncated"`
	EOF        bool   `json:"eof"`
	Content    string `json:"content"`
}

// Load a bounded, stable UTF-8 snapshot. This does not authorize a write.
func sourceSnapshot(root, rel string) ([]byte, string, error) {
	s := session{root: root}
	path, err := s.safePath(rel)
	if err != nil {
		return nil, "", errors.New("geçersiz dosya yolu")
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > sourceSizeLimit {
		return nil, "", errors.New("dosya okunamıyor veya 8 MiB sınırını aşıyor")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil || !os.SameFile(info, before) {
		return nil, "", errors.New("dosya açılırken değişti")
	}
	b, err := io.ReadAll(io.LimitReader(f, sourceSizeLimit+1))
	if err != nil {
		return nil, "", err
	}
	after, err := f.Stat()
	if err != nil || after.Size() != before.Size() || !after.ModTime().Equal(before.ModTime()) || len(b) != int(after.Size()) || len(b) > sourceSizeLimit {
		return nil, "", errors.New("dosya okunurken değişti")
	}
	if !utf8.Valid(b) || bytes.IndexByte(b, 0) >= 0 {
		return nil, "", errors.New("ikili dosya veya geçersiz UTF-8")
	}
	return b, relPath(root, path), nil
}

func readSourceRange(root string, req sourceRangeRequest, maxBytes int) (sourceChunk, error) {
	var out sourceChunk
	if maxBytes < 4 || maxBytes > 64<<10 {
		return out, errors.New("geçersiz parça bütçesi")
	}
	b, rel, err := sourceSnapshot(root, req.Path)
	if err != nil {
		return out, err
	}
	hash := hashContent(b)
	if req.SHA256 != "" && req.SHA256 != hash {
		return out, errors.New("dosya sürümü değişti; parçaları yeniden okuyun")
	}
	lines := bytes.Count(b, []byte("\n"))
	if len(b) > 0 && b[len(b)-1] != '\n' {
		lines++
	}
	start, end := req.Offset, len(b)
	truncated := false
	if req.StartLine != 0 || req.EndLine != 0 {
		if req.Offset != 0 || req.Limit != 0 || req.StartLine < 1 || req.EndLine < req.StartLine || req.EndLine-req.StartLine >= 200 || req.StartLine > lines {
			return out, errors.New("geçersiz satır aralığı")
		}
		start, end = 0, len(b)
		line := 1
		for i, c := range b {
			if line == req.StartLine && i == 0 {
				start = 0
			}
			if c == '\n' {
				line++
				if line == req.StartLine {
					start = i + 1
				}
				if line == req.EndLine+1 {
					end = i + 1
					break
				}
			}
		}
		if end-start > maxBytes {
			end = start + maxBytes
			truncated = true
			for end > start && !utf8.RuneStart(b[end]) {
				end--
			}
		}
	} else {
		if start < 0 || start > len(b) || (start < len(b) && !utf8.RuneStart(b[start])) {
			return out, errors.New("geçersiz bayt başlangıcı")
		}
		limit := req.Limit
		if limit == 0 {
			limit = maxBytes
		}
		if limit < 1 || limit > 64<<10 {
			return out, errors.New("geçersiz parça uzunluğu")
		}
		if limit > maxBytes {
			limit = maxBytes
			truncated = end-start > maxBytes
		}
		if end-start > limit {
			end = start + limit
			for end > start && !utf8.RuneStart(b[end]) {
				end--
			}
		}
		if end == start && start < len(b) {
			return out, errors.New("parça bir UTF-8 karakterini almaya yetmiyor")
		}
	}
	text := string(b[start:end])
	clean := text
	// Detect secrets before slicing: otherwise a token split across adjacent
	// chunks could evade redaction. Never provide write evidence in that case.
	redacted := redactLine(string(b)) != string(b)
	if redacted {
		clean = "[Dosya gizli bilgi içeriyor; parça içeriği gizlendi.]"
	}
	out = sourceChunk{Path: rel, SHA256: hash, Size: len(b), Offset: start, NextOffset: end, StartLine: bytes.Count(b[:start], []byte("\n")) + 1, TotalLines: lines, Complete: start == 0 && end == len(b), Redacted: redacted, Truncated: truncated, EOF: end == len(b), Content: clean}
	out.EndLine = out.StartLine + strings.Count(text, "\n")
	if strings.HasSuffix(text, "\n") {
		out.EndLine--
	}
	if text == "" {
		out.EndLine = out.StartLine - 1
	}
	return out, nil
}

type patchEvidence struct {
	SHA256   string
	Snippets []string
	Ranges   []observedSourceRange
}

type observedSourceRange struct {
	Start, End int
	Content    string
}

func rememberSourceChunk(evidence map[string]patchEvidence, chunk sourceChunk) {
	if chunk.Redacted {
		return
	}
	prev := evidence[chunk.Path]
	if prev.SHA256 != chunk.SHA256 {
		prev = patchEvidence{SHA256: chunk.SHA256}
	}
	// Evidence is bounded independently from the model window.
	if len(prev.Snippets) == 32 {
		prev.Snippets = append([]string(nil), prev.Snippets[1:]...)
	}
	prev.Snippets = append(prev.Snippets, chunk.Content)
	if chunk.Offset >= 0 && chunk.NextOffset > chunk.Offset && chunk.NextOffset-chunk.Offset == len(chunk.Content) {
		prev.Ranges = append(prev.Ranges, observedSourceRange{chunk.Offset, chunk.NextOffset, chunk.Content})
		sort.Slice(prev.Ranges, func(i, j int) bool { return prev.Ranges[i].Start < prev.Ranges[j].Start })
		var merged []observedSourceRange
		for _, r := range prev.Ranges {
			if len(merged) > 0 {
				last := &merged[len(merged)-1]
				if r.Start <= last.End && r.End >= last.Start {
					overlap := last.End - r.Start
					if overlap > len(r.Content) {
						overlap = len(r.Content)
					}
					at := r.Start - last.Start
					if at >= 0 && at+overlap <= len(last.Content) && last.Content[at:at+overlap] == r.Content[:overlap] {
						if r.End > last.End {
							last.Content += r.Content[overlap:]
							last.End = r.End
						}
						continue
					}
				}
			}
			merged = append(merged, r)
		}
		// Keep the same bounded recent evidence as independent snippets.
		if len(merged) > 32 {
			merged = merged[len(merged)-32:]
		}
		for i := range merged {
			if len(merged[i].Content) > 256<<10 {
				cut := len(merged[i].Content) - (256 << 10)
				for cut < len(merged[i].Content) && !utf8.RuneStart(merged[i].Content[cut]) {
					cut++
				}
				merged[i].Start += cut
				merged[i].Content = merged[i].Content[cut:]
			}
		}
		prev.Ranges = merged
	}
	evidence[chunk.Path] = prev
}

func patchFromEvidence(root, path, hash, before, after string, evidence map[string]patchEvidence) (proposedFile, error) {
	var item proposedFile
	if before == "" || len(before)+len(after) > sourceSizeLimit {
		return item, errors.New("boş veya çok büyük yama")
	}
	b, rel, err := sourceSnapshot(root, path)
	if err != nil {
		return item, err
	}
	proof, ok := evidence[rel]
	if !ok || hash == "" || hash != proof.SHA256 || hash != hashContent(b) {
		return item, errors.New("yama için okuma ve sürüm kanıtı yok")
	}
	observed := false
	for _, snippet := range proof.Snippets {
		if strings.Contains(snippet, before) {
			observed = true
			break
		}
	}
	if !observed {
		for _, r := range proof.Ranges {
			if strings.Contains(r.Content, before) {
				observed = true
				break
			}
		}
	}
	if !observed || strings.Count(string(b), before) != 1 {
		return item, errors.New("yama hedefi okunmadı veya benzersiz değil")
	}
	next := strings.Replace(string(b), before, after, 1)
	if len(next) > sourceSizeLimit || !utf8.ValidString(next) || strings.IndexByte(next, 0) >= 0 {
		return item, fmt.Errorf("yama sonucu geçersiz")
	}
	return proposedFile{Path: rel, Existed: true, Old: string(b), OldHash: hash, New: next}, nil
}
