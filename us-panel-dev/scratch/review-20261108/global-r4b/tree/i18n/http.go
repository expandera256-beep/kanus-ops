package i18n

import (
	"encoding/json"
	"net/http"
)

// LangInfo describes one selectable language.
type LangInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Dir  string `json:"dir"`
}

// Document is the JSON payload served to the panel UI.
type Document struct {
	Lang      string            `json:"lang"`
	Dir       string            `json:"dir"`
	BCP47     string            `json:"bcp47"`
	Languages []LangInfo        `json:"languages"`
	Messages  map[string]string `json:"messages"`
}

// Doc builds the UI document for lang with fallback-resolved messages.
func (b *Bundle) Doc(lang string) Document {
	l := b.For(lang)
	b.mu.RLock()
	keys := make([]string, 0, len(b.cats[Default].msgs))
	for k := range b.cats[Default].msgs {
		keys = append(keys, k)
	}
	langs := make([]LangInfo, 0, len(b.cats))
	b.mu.RUnlock()
	for _, code := range b.Languages() {
		m := b.For(code).Meta()
		langs = append(langs, LangInfo{Code: code, Name: m.Name, Dir: m.Dir})
	}
	msgs := make(map[string]string, len(keys))
	for _, k := range keys {
		msgs[k] = l.T(k)
	}
	m := l.Meta()
	return Document{Lang: l.Lang(), Dir: m.Dir, BCP47: m.BCP47, Languages: langs, Messages: msgs}
}

// Handler serves GET /api/i18n?lang=xx. Without lang it negotiates from
// the resolver (user preference) and Accept-Language; default is TR.
func Handler(pref func() string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		b, err := Load()
		if err != nil {
			http.Error(w, "i18n unavailable", http.StatusInternalServerError)
			return
		}
		p := r.URL.Query().Get("lang")
		if p == "" && pref != nil {
			p = pref()
		}
		lang := b.Negotiate(p, "") // unknown or empty -> Default (TR)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(b.Doc(lang))
	})
}
