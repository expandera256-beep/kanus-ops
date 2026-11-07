package i18n

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
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

// maxPrefSize bounds the preference file; larger files are ignored.
const maxPrefSize = 4096

// errUnsafePath is returned when a path component is a symlink or not a
// regular file/directory.
var errUnsafePath = errors.New("i18n: unsafe preference path")

// Get returns the stored, supported language or "" (caller uses TR).
// It never holds the mutex while touching the file system, refuses
// symlinks/FIFOs/devices (O_NOFOLLOW|O_NONBLOCK + fstat) and reads at most
// maxPrefSize bytes.
func (p *PrefStore) Get() string {
	if p == nil || p.Path == nil {
		return ""
	}
	path := p.Path()
	if path == "" {
		return ""
	}
	raw, err := readPrefFile(path)
	if err != nil {
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

// Set validates and atomically stores lang. The parent directory is
// resolved without following symlinks and the file is written relative to
// that directory handle (temp file + rename).
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
	raw, _ := json.Marshal(prefDoc{Lang: lang})
	p.mu.Lock()
	defer p.mu.Unlock()
	return writePrefFile(path, raw)
}

func readLimited(r io.Reader) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(r, maxPrefSize+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > maxPrefSize {
		return nil, errUnsafePath
	}
	return raw, nil
}

// PrefHandler serves GET/POST /api/lang. POST requires a single JSON
// document ({"lang":"en"}) of at most maxBodySize bytes; allow (if set)
// must accept the request (same Origin check as /api/act).
func PrefHandler(p *PrefStore, allow func(*http.Request) bool) http.Handler {
	const maxBodySize = 4096
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
			if allow != nil && !allow(r) {
				http.Error(w, "forbidden origin", http.StatusForbidden)
				return
			}
			if mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mt != "application/json" {
				http.Error(w, "json required", http.StatusUnsupportedMediaType)
				return
			}
			dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize))
			var d prefDoc
			err := dec.Decode(&d)
			if err == nil {
				if extra := dec.Decode(&struct{}{}); extra != io.EOF {
					err = errors.New("trailing data")
					if extra != nil {
						err = extra
					}
				}
			}
			if err != nil {
				var tooBig *http.MaxBytesError
				if errors.As(err, &tooBig) {
					http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
					return
				}
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
