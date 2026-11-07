package i18n

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// PrefStore persists the user's language choice as a small JSON file
// ({"lang":"en"}) next to the panel's existing desk config.
type PrefStore struct {
	mu   sync.Mutex
	Path func() string
}

type prefDoc struct {
	Lang string `json:"lang"`
}

// Get returns the stored, supported language or "" if none.
func (p *PrefStore) Get() string {
	if p == nil || p.Path == nil {
		return ""
	}
	path := p.Path()
	if path == "" {
		return ""
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	raw, err := os.ReadFile(path)
	if err != nil || len(raw) > 4096 {
		return ""
	}
	var d prefDoc
	if json.Unmarshal(raw, &d) != nil {
		return ""
	}
	b, err := Load()
	if err != nil || !b.has(normalize(d.Lang)) {
		return ""
	}
	return normalize(d.Lang)
}

// ErrUnsupported is returned for languages without a locale file.
var ErrUnsupported = errors.New("i18n: unsupported language")

// Set validates and atomically stores lang.
func (p *PrefStore) Set(lang string) error {
	b, err := Load()
	if err != nil {
		return err
	}
	lang = normalize(lang)
	if !b.has(lang) {
		return ErrUnsupported
	}
	if p == nil || p.Path == nil || p.Path() == "" {
		return errors.New("i18n: no preference path")
	}
	path := p.Path()
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, _ := json.Marshal(prefDoc{Lang: lang})
	tmp, err := os.CreateTemp(filepath.Dir(path), ".lang-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err = tmp.Write(raw); err == nil {
		err = tmp.Sync()
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(name, 0o600)
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	if err != nil {
		_ = os.Remove(name)
	}
	return err
}

// PrefHandler serves GET/POST /api/lang. POST requires a JSON body
// ({"lang":"en"}) so cross-site simple form posts cannot change it.
func PrefHandler(p *PrefStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		switch r.Method {
		case http.MethodGet:
			lang := p.Get()
			if lang == "" {
				lang = Default
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(prefDoc{Lang: lang})
		case http.MethodPost:
			if mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mt != "application/json" {
				http.Error(w, "json required", http.StatusUnsupportedMediaType)
				return
			}
			var d prefDoc
			if err := json.NewDecoder(io.LimitReader(r.Body, 1024)).Decode(&d); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if err := p.Set(d.Lang); err != nil {
				if errors.Is(err, ErrUnsupported) {
					http.Error(w, "unsupported language", http.StatusBadRequest)
					return
				}
				http.Error(w, "could not save", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(prefDoc{Lang: normalize(d.Lang)})
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
