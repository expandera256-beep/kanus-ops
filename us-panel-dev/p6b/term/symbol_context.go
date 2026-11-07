package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Automatic symbol context (PANEL DEV option b). Off unless
// US_AGENT_SYMBOL_CONTEXT is 1/true/on/yes. Only "path:line kind name" is
// injected, never line text; ≤3 names, ≤5 hits each, ≤600 bytes and ≤25%
// of the given limit. No tool or protocol change.
const (
	symbolContextHeader   = "\nSembol tanımları (yol:satır tür ad; içerik için read_file kullan):\n"
	symbolContextMaxNames = 3
	symbolContextMaxHits  = 5
	symbolContextMaxBytes = 600
)

func symbolContextEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("US_AGENT_SYMBOL_CONTEXT"))) {
	case "1", "true", "on", "yes":
		return true
	}
	return false
}

var (
	symbolBacktickRE = regexp.MustCompile("`([^`\\s]{1,128})`")
	symbolWordRE     = regexp.MustCompile(`[A-Za-z_$][A-Za-z0-9_$]*`)
	symbolCamelRE    = regexp.MustCompile(`[a-z0-9][A-Z]`)
	symbolStop       = map[string]bool{"the": true, "and": true, "for": true, "with": true, "this": true, "that": true, "proje": true, "dosya": true}
)

// symbolTaskNames extracts at most 3 identifiers: backticked names first,
// then CamelCase or snake_case words; ≥3 chars, no stopwords, deduplicated.
func symbolTaskNames(task string) []string {
	var out []string
	seen := map[string]bool{}
	add := func(w string) {
		if len(out) >= symbolContextMaxNames || len(w) < 3 || seen[w] || symbolStop[strings.ToLower(w)] || !symbolNameRE.MatchString(w) {
			return
		}
		seen[w] = true
		out = append(out, w)
	}
	for _, m := range symbolBacktickRE.FindAllStringSubmatch(task, -1) {
		add(m[1])
	}
	for _, w := range symbolWordRE.FindAllString(task, -1) {
		if symbolCamelRE.MatchString(w) || (strings.Contains(strings.Trim(w, "_"), "_") && strings.ContainsAny(strings.ToLower(w), "abcdefghijklmnopqrstuvwxyz")) {
			add(w)
		}
	}
	return out
}

// symbolContextBlock renders the bounded block for task, or "" when there is
// nothing to add or it would not fit.
func symbolContextBlock(root, task string, limit int) string {
	budget := symbolContextMaxBytes
	if q := limit / 4; q < budget {
		budget = q
	}
	if budget <= len(symbolContextHeader) {
		return ""
	}
	names := symbolTaskNames(task)
	if len(names) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(symbolContextHeader)
	lines := 0
	for _, name := range names {
		for _, h := range findSymbols(root, name, symbolContextMaxHits) {
			ln := h.Rel + ":" + strconv.Itoa(h.Line) + " " + h.Kind + " " + h.Name + "\n"
			if strings.ContainsAny(h.Rel, " \n\r\t") || b.Len()+len(ln) > budget {
				continue
			}
			b.WriteString(ln)
			lines++
		}
	}
	if lines == 0 {
		return ""
	}
	return b.String()
}
