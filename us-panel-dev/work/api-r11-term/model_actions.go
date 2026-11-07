package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type modelOperation struct {
	Op        string  `json:"op"`
	Path      string  `json:"path"`
	Content   *string `json:"content,omitempty"`
	SHA256    string  `json:"sha256,omitempty"`
	Before    *string `json:"before,omitempty"`
	After     *string `json:"after,omitempty"`
	Offset    int     `json:"offset,omitempty"`
	Limit     int     `json:"limit,omitempty"`
	StartLine int     `json:"start_line,omitempty"`
	EndLine   int     `json:"end_line,omitempty"`
}
type modelActionEnvelope struct {
	Version    int              `json:"version"`
	Message    string           `json:"message,omitempty"`
	Operations []modelOperation `json:"operations"`
}

// Do not fall back to free-text file parsing after a damaged JSON envelope.
func parseModelActions(reply string) (*modelActionEnvelope, bool, error) {
	raw := strings.TrimSpace(reply)
	if strings.HasPrefix(raw, "```json") {
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "```json"))
		if strings.HasSuffix(raw, "```") {
			raw = strings.TrimSpace(strings.TrimSuffix(raw, "```"))
		}
	}
	if !strings.HasPrefix(raw, "{") || (!strings.Contains(raw, "\"operations\"") && !strings.Contains(raw, "\"version\"")) {
		return nil, false, nil
	}
	if len(raw) > sourceSizeLimit || !utf8.ValidString(raw) {
		return nil, true, errors.New("işlem yanıtı çok büyük veya UTF-8 değil")
	}
	d := json.NewDecoder(strings.NewReader(raw))
	d.UseNumber()
	if err := uniqueJSONValue(d, 0); err != nil {
		return nil, true, fmt.Errorf("işlem JSON'u geçersiz: %w", err)
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, true, errors.New("işlem yanıtında ek veri var")
	}
	d = json.NewDecoder(strings.NewReader(raw))
	d.DisallowUnknownFields()
	var env modelActionEnvelope
	if err := d.Decode(&env); err != nil {
		return nil, true, fmt.Errorf("işlem şeması geçersiz: %w", err)
	}
	if env.Version != 1 || env.Operations == nil || len(env.Operations) > 16 || len(env.Message) > 16<<10 {
		return nil, true, errors.New("işlem sürümü veya sayısı geçersiz")
	}
	readMode, writeMode := false, false
	for _, op := range env.Operations {
		if op.Path == "" || len(op.Path) > 1024 {
			return nil, true, errors.New("işlem yolu boş veya çok uzun")
		}
		rangeSet := op.Offset != 0 || op.Limit != 0 || op.StartLine != 0 || op.EndLine != 0
		switch op.Op {
		case "list_files":
			readMode = true
			if !validListActionPath(op.Path) || op.Offset < 0 || op.Limit < 0 || op.Limit > 200 || op.StartLine != 0 || op.EndLine != 0 || op.Content != nil || op.Before != nil || op.After != nil || op.SHA256 != "" {
				return nil, true, errors.New("listeleme alanları geçersiz")
			}
		case "read_file":
			readMode = true
			if op.Content != nil || op.Before != nil || op.After != nil {
				return nil, true, errors.New("okuma gövde içeremez")
			}
		case "create_file", "write_file":
			writeMode = true
			if op.Content == nil || op.Before != nil || op.After != nil || rangeSet {
				return nil, true, errors.New("dosya yazma alanları geçersiz")
			}
		case "patch_file":
			writeMode = true
			if op.Content != nil || op.Before == nil || op.After == nil || rangeSet || len(op.SHA256) != 64 {
				return nil, true, errors.New("yama alanları geçersiz")
			}
		default:
			return nil, true, errors.New("desteklenmeyen işlem: " + op.Op)
		}
		if op.SHA256 != "" {
			if len(op.SHA256) != 64 {
				return nil, true, errors.New("geçersiz dosya hash'i")
			}
			if _, err := hex.DecodeString(op.SHA256); err != nil {
				return nil, true, errors.New("geçersiz dosya hash'i")
			}
		}
	}
	if readMode && writeMode {
		return nil, true, errors.New("okuma ile yazma aynı yanıtta karıştırılamaz")
	}
	return &env, true, nil
}

func validListActionPath(path string) bool {
	if path == "." {
		return true
	}
	if path == "" || filepath.IsAbs(path) || strings.ContainsAny(path, "\\:") {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." || part == ".." || skipTreeDir(part) {
			return false
		}
	}
	return true
}

func requestsModelRead(reply string) bool {
	env, yes, err := parseModelActions(reply)
	return yes && err == nil && len(env.Operations) > 0 && (env.Operations[0].Op == "read_file" || env.Operations[0].Op == "list_files")
}

func uniqueJSONValue(d *json.Decoder, depth int) error {
	if depth > 24 {
		return errors.New("JSON çok derin")
	}
	t, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			t, err := d.Token()
			if err != nil {
				return err
			}
			key, ok := t.(string)
			if !ok || seen[key] {
				return errors.New("yinelenen JSON alanı")
			}
			seen[key] = true
			if err = uniqueJSONValue(d, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := uniqueJSONValue(d, depth+1); err != nil {
				return err
			}
		}
	default:
		return errors.New("beklenmeyen JSON kapanışı")
	}
	_, err = d.Token()
	return err
}

func actionReceipt(reply string) string {
	env, yes, err := parseModelActions(reply)
	if !yes {
		return reply
	}
	if err != nil {
		return "Yapılandırılmış model yanıtı reddedildi: " + err.Error()
	}
	var b strings.Builder
	b.WriteString(env.Message)
	for _, op := range env.Operations {
		fmt.Fprintf(&b, "\n%s %s", op.Op, op.Path)
	}
	return strings.TrimSpace(b.String())
}

func readActionChunks(root string, env *modelActionEnvelope, budget int) ([]sourceChunk, string, error) {
	return readActionChunksBounded(root, env, budget, 0)
}

func readActionChunksBounded(root string, env *modelActionEnvelope, budget, capBytes int) ([]sourceChunk, string, error) {
	if len(env.Operations) == 0 || len(env.Operations) > 5 {
		return nil, "", errors.New("tek turda 1–5 dosya parçası isteyin")
	}
	// Use the actual serialized response bound instead of permanently halving
	// every source read for hypothetical escaping. Retry smaller if necessary.
	maxBytes := (budget - len(env.Operations)*420) / len(env.Operations)
	if capBytes > 0 && maxBytes > capBytes {
		maxBytes = capBytes
	}
	if maxBytes > 32<<10 {
		maxBytes = 32 << 10
	}
	if maxBytes < 4 {
		return nil, "", errors.New("okuma isteği bağlama sığmıyor")
	}
	var chunks []sourceChunk
	type fileListing struct {
		Path       string   `json:"path"`
		Offset     int      `json:"offset"`
		NextOffset int      `json:"next_offset"`
		More       bool     `json:"more"`
		Items      []string `json:"items"`
	}
	var listings []fileListing
	marshal := func() ([]byte, error) {
		return json.Marshal(struct {
			Files    []sourceChunk `json:"files"`
			Listings []fileListing `json:"listings,omitempty"`
		}{chunks, listings})
	}
	for _, op := range env.Operations {
		if op.Op == "list_files" {
			listRoot := root
			if op.Path != "." {
				full, _, ok := replyTarget(root, op.Path)
				if !ok {
					return nil, "", errors.New("listeleme yolu çalışma alanı dışında")
				}
				info, err := os.Lstat(full)
				if err != nil || !info.IsDir() {
					return nil, "", errors.New("list_files için gerçek dizin gerekir; dosya içeriği için read_file kullanın")
				}
				listRoot = full
			}
			limit := op.Limit
			if limit == 0 {
				limit = 80
			}
			for {
				items, more, err := workspacePage(listRoot, op.Offset, limit)
				if err != nil {
					return nil, "", err
				}
				if op.Path != "." {
					for i := range items {
						items[i] = op.Path + "/" + items[i]
					}
				}
				listings = append(listings, fileListing{op.Path, op.Offset, op.Offset + len(items), more, items})
				raw, err := marshal()
				if err == nil && len(raw) <= budget {
					break
				}
				listings = listings[:len(listings)-1]
				if limit == 1 {
					return nil, "", errors.New("dosya listesi bağlama sığmıyor")
				}
				limit = (limit + 1) / 2
			}
			continue
		}
		if op.Op != "read_file" {
			return nil, "", errors.New("okuma ile yazma aynı turda olamaz")
		}
		chunk, err := readSourceRange(root, sourceRangeRequest{Path: op.Path, SHA256: op.SHA256, Offset: op.Offset, Limit: op.Limit, StartLine: op.StartLine, EndLine: op.EndLine}, maxBytes)
		if err != nil {
			return nil, "", fmt.Errorf("%s: %w", op.Path, err)
		}
		chunks = append(chunks, chunk)
	}
	raw, err := marshal()
	if err != nil || len(raw) > budget {
		if err == nil && maxBytes > 4 {
			return readActionChunksBounded(root, env, budget, maxBytes/2)
		}
		return nil, "", errors.New("dosya parçaları bağlama sığmıyor; daha küçük aralık isteyin")
	}
	return chunks, string(raw), nil
}

func prepareActionProposal(root string, env *modelActionEnvelope, allowed map[string]bool, hashes map[string]string, evidence map[string]patchEvidence) (*fileProposal, error) {
	p := &fileProposal{Root: root}
	seen := map[string]bool{}
	for _, op := range env.Operations {
		if op.Op == "read_file" || op.Op == "list_files" {
			return nil, errors.New("model ek okuma istiyor; dosyalar uygulanmadı")
		}
		full, rel, ok := replyTarget(root, op.Path)
		if !ok || seen[strings.ToLower(rel)] {
			return nil, errors.New("geçersiz veya yinelenen dosya yolu")
		}
		seen[strings.ToLower(rel)] = true
		var item proposedFile
		if op.Op == "patch_file" {
			var err error
			item, err = patchFromEvidence(root, op.Path, op.SHA256, *op.Before, *op.After, evidence)
			if err != nil {
				return nil, err
			}
		} else {
			if op.Content == nil || len(*op.Content) > sourceSizeLimit || !utf8.ValidString(*op.Content) || strings.IndexByte(*op.Content, 0) >= 0 {
				return nil, errors.New("geçersiz dosya gövdesi")
			}
			item = proposedFile{Path: rel, New: *op.Content}
			_, err := os.Lstat(full)
			if err == nil {
				if op.Op == "create_file" {
					return nil, errors.New("yeni dosya zaten mevcut")
				}
				b, _, err := sourceSnapshot(root, rel)
				if err != nil {
					return nil, err
				}
				h := hashContent(b)
				if !allowed[rel] || hashes[rel] != h {
					return nil, fmt.Errorf("tam ve güncel okuma yok: %s", rel)
				}
				if op.SHA256 != "" && op.SHA256 != h {
					return nil, errors.New("dosya hash'i uyuşmuyor")
				}
				item.Existed = true
				item.Old = string(b)
				item.OldHash = h
			} else if !os.IsNotExist(err) {
				return nil, err
			} else if op.Op == "write_file" {
				return nil, errors.New("write_file hedefi mevcut değil; create_file kullanın")
			}
		}
		p.Files = append(p.Files, item)
	}
	if len(p.Files) == 0 {
		return nil, nil
	}
	return p, nil
}
