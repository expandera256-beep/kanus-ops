package imageapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"strconv"
	"strings"
	"unicode/utf8"
)

type parsedReq struct {
	Model  string
	Prompt string
	Size   string
	N      int
	Seed   *uint32
}

type fieldErr struct {
	status int
	code   string
	msg    string
	field  string
	reason string
}

func parseImageRequest(body []byte) (parsedReq, *fieldErr) {
	if hasDuplicateTopKeys(body) || hasTrailingData(body) {
		return parsedReq{}, &fieldErr{status: 400, code: "invalid_json", msg: "JSON geçersiz"}
	}
	var raw map[string]json.RawMessage
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		return parsedReq{}, &fieldErr{status: 400, code: "invalid_json", msg: "JSON geçersiz"}
	}
	allowed := map[string]struct{}{"model": {}, "prompt": {}, "size": {}, "n": {}, "seed": {}}
	for k := range raw {
		if _, ok := unsupportedFields[k]; ok {
			return parsedReq{}, &fieldErr{status: 400, code: "unsupported_parameter", msg: "Desteklenmeyen parametre", field: k}
		}
		if _, ok := allowed[k]; !ok {
			return parsedReq{}, &fieldErr{status: 400, code: "unknown_field", msg: "Bilinmeyen alan", field: k}
		}
	}

	out := parsedReq{Model: defaultModel, Size: defaultSize, N: defaultN}

	if v, ok := raw["model"]; ok {
		if string(v) == "null" {
			return parsedReq{}, &fieldErr{status: 400, code: "invalid_model", msg: "model geçersiz", field: "model"}
		}
		var m string
		if err := json.Unmarshal(v, &m); err != nil {
			return parsedReq{}, &fieldErr{status: 400, code: "invalid_model", msg: "model geçersiz", field: "model"}
		}
		if m != "klein" && m != "schnell" {
			return parsedReq{}, &fieldErr{status: 400, code: "invalid_model", msg: "model allowlist dışı", field: "model"}
		}
		out.Model = m
	}
	v, ok := raw["prompt"]
	if !ok {
		return parsedReq{}, &fieldErr{status: 400, code: "invalid_prompt", msg: "prompt zorunludur", field: "prompt"}
	}
	var prompt string
	if err := json.Unmarshal(v, &prompt); err != nil {
		return parsedReq{}, &fieldErr{status: 400, code: "invalid_prompt", msg: "prompt geçersiz", field: "prompt"}
	}
	if !utf8.ValidString(prompt) {
		return parsedReq{}, &fieldErr{status: 400, code: "invalid_prompt", msg: "prompt geçerli UTF-8 değil", field: "prompt"}
	}
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return parsedReq{}, &fieldErr{status: 400, code: "invalid_prompt", msg: "prompt boş olamaz", field: "prompt"}
	}
	for _, r := range prompt {
		if r < 0x20 && r != '\n' && r != '\t' {
			return parsedReq{}, &fieldErr{status: 400, code: "invalid_prompt", msg: "prompt kontrol karakteri içeriyor", field: "prompt"}
		}
	}
	out.Prompt = prompt

	if v, ok := raw["size"]; ok {
		var size string
		if err := json.Unmarshal(v, &size); err != nil {
			return parsedReq{}, &fieldErr{status: 400, code: "invalid_size", msg: "size geçersiz", field: "size"}
		}
		if _, ok := allowedSizes[size]; !ok {
			return parsedReq{}, &fieldErr{status: 400, code: "invalid_size", msg: "size allowlist dışı", field: "size"}
		}
		out.Size = size
	}
	if v, ok := raw["n"]; ok {
		n, err := parseJSONInt(v, 1, 4)
		if err != nil {
			return parsedReq{}, &fieldErr{status: 400, code: "invalid_n", msg: "n 1..4 tamsayı olmalı", field: "n"}
		}
		out.N = int(n)
	}
	if v, ok := raw["seed"]; ok {
		if string(v) == "null" {
			return parsedReq{}, &fieldErr{status: 400, code: "invalid_seed", msg: "seed geçersiz", field: "seed"}
		}
		n, err := parseJSONInt(v, 0, 4294967295)
		if err != nil {
			return parsedReq{}, &fieldErr{status: 400, code: "invalid_seed", msg: "seed 0..4294967295 olmalı", field: "seed"}
		}
		s := uint32(n)
		out.Seed = &s
	}
	return out, nil
}

func parseJSONInt(raw json.RawMessage, min, max uint64) (uint64, error) {
	s := strings.TrimSpace(string(raw))
	if s == "null" || s == "" {
		return 0, strconv.ErrSyntax
	}
	if strings.ContainsAny(s, ".eE\"") {
		return 0, strconv.ErrSyntax
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	if n < min || n > max {
		return 0, strconv.ErrRange
	}
	return n, nil
}

func validIdempotencyKey(s string) bool {
	if len(s) < idempotencyMin || len(s) > idempotencyMax {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '.' || c == '_' || c == ':' || c == '-'
		if !ok {
			return false
		}
	}
	return true
}

func isJSONContentType(ct string) bool {
	if ct == "" {
		return false
	}
	mt, _, err := mime.ParseMediaType(ct)
	return err == nil && mt == "application/json"
}

func hasDuplicateTopKeys(body []byte) bool {
	dec := json.NewDecoder(bytes.NewReader(body))
	t, err := dec.Token()
	if err != nil {
		return false
	}
	d, ok := t.(json.Delim)
	if !ok || d != '{' {
		return false
	}
	seen := map[string]struct{}{}
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return false
		}
		key, ok := kt.(string)
		if !ok {
			return false
		}
		if _, dup := seen[key]; dup {
			return true
		}
		seen[key] = struct{}{}
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return false
		}
	}
	return false
}

func hasTrailingData(body []byte) bool {
	dec := json.NewDecoder(bytes.NewReader(body))
	var v any
	if err := dec.Decode(&v); err != nil {
		return false
	}
	var extra any
	err := dec.Decode(&extra)
	return err == nil || (err != nil && !errors.Is(err, io.EOF))
}
