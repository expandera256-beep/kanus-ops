package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

// PrefStore persists the user's language choice as a small JSON file
// ({"lang":"en"}) under a trusted config root (the panel's desk.json
// directory).
//
// Root is resolved ONCE per distinct value (filepath.EvalSymlinks, so a
// legitimately symlinked HOME/config such as /home -> /var/home works) and
// pinned as a directory handle. Everything below the root (Name and its
// intermediate directories) is opened relative to that handle without
// following symlinks or reparse points; the file is written as a temp file
// and atomically renamed relative to the same handle. Re-pointing the root
// symlink later does not move writes. If the root cannot be opened the
// choice is kept in memory only and Set returns ErrNotPersisted.
type PrefStore struct {
	mu    sync.Mutex
	Root  func() string
	Name  string // relative path under Root; default "lang.json"
	roots map[string]*prefRoot
	mem   map[string]string
}

type prefRoot struct {
	h       rootHandle
	err     error
	refs    int  // in-flight Get/Set using h; guarded by PrefStore.mu
	evicted bool // removed from the cache; closed when refs reaches 0
	closed  bool
}

// maxRoots bounds the pinned-root cache. The GUI uses one fixed root, so
// one entry is enough; a new Root() value evicts the old one and its
// handle is closed as soon as no Get/Set is using it (same on Unix and
// Windows), keeping the descriptor count constant.
const maxRoots = 1

// Root handle accounting (process-wide). Tests use these to measure leaks
// directly instead of OS-wide handle counts, which include Go runtime and
// AV/EDR handles on Windows.
var rootOpened, rootClosed atomic.Int64

// rootCloseHook is nil in production; tests set it to verify that a closed
// handle is really invalid.
var rootCloseHook func(rootHandle)

type prefDoc struct {
	Lang string `json:"lang"`
}

// maxPrefSize bounds the preference file; larger files are ignored.
const maxPrefSize = 4096

// errUnsafePath is returned when a component below the root is a symlink,
// reparse point or not a regular file/directory.
var errUnsafePath = errors.New("i18n: unsafe preference path")

// ErrNotPersisted means the config root is unavailable; the choice is
// kept in memory for this process only.
var ErrNotPersisted = errors.New("i18n: language preference could not be saved")

// ErrUnsupported is returned for languages without a locale file.
var ErrUnsupported = errors.New("i18n: unsupported language")

func (p *PrefStore) name() ([]string, error) {
	n := p.Name
	if n == "" {
		n = "lang.json"
	}
	n = filepath.ToSlash(n)
	if strings.HasPrefix(n, "/") || filepath.IsAbs(n) || filepath.VolumeName(n) != "" {
		return nil, errUnsafePath
	}
	parts := strings.Split(n, "/")
	for _, s := range parts {
		if s == "" || s == "." || s == ".." || strings.ContainsAny(s, "\\:") {
			return nil, errUnsafePath
		}
	}
	return parts, nil
}

// root returns the pinned root for the current Root() value, opening it on
// first use. Only this short lookup holds the mutex; file I/O does not.
func (p *PrefStore) root() (string, *prefRoot) {
	if p == nil || p.Root == nil {
		return "", nil
	}
	key := p.Root()
	if key == "" {
		return "", nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if r, ok := p.roots[key]; ok {
		r.refs++
		return key, r
	}
	if p.roots == nil {
		p.roots = map[string]*prefRoot{}
	}
	for len(p.roots) >= maxRoots {
		for k, old := range p.roots {
			delete(p.roots, k)
			old.evicted = true
			p.closeIfIdleLocked(old)
			break
		}
	}
	r := &prefRoot{}
	resolved, err := resolveRoot(key)
	if err == nil {
		r.h, err = openRoot(resolved)
		if err == nil {
			rootOpened.Add(1)
		}
	}
	r.err = err
	r.refs = 1
	p.roots[key] = r
	return key, r
}

// release ends one use of r returned by root(); it must not be called with
// p.mu held.
func (p *PrefStore) release(r *prefRoot) {
	if r == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	r.refs--
	p.closeIfIdleLocked(r)
}

func (p *PrefStore) closeIfIdleLocked(r *prefRoot) {
	if r.evicted && r.refs <= 0 && r.err == nil && !r.closed {
		r.closed = true
		closeRoot(r.h)
		rootClosed.Add(1)
		if rootCloseHook != nil {
			rootCloseHook(r.h)
		}
	}
}

// Close releases the pinned root handles. In-flight calls finish first;
// later calls re-open the root on demand.
func (p *PrefStore) Close() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for k, r := range p.roots {
		delete(p.roots, k)
		r.evicted = true
		p.closeIfIdleLocked(r)
	}
	return nil
}

// Path reports the (unresolved) preference file path for diagnostics.
func (p *PrefStore) Path() string {
	if p == nil || p.Root == nil || p.Root() == "" {
		return ""
	}
	n := p.Name
	if n == "" {
		n = "lang.json"
	}
	return filepath.Join(p.Root(), filepath.FromSlash(n))
}

func (p *PrefStore) memGet(key string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mem[key]
}

func resolveRoot(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(abs, 0o700); err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// Get returns the stored, supported language or "" (caller uses TR).
// Files are opened without following symlinks (O_NOFOLLOW|O_NONBLOCK /
// FILE_OPEN_REPARSE_POINT), must be regular, and at most maxPrefSize bytes
// are read. The mutex is never held during file I/O.
func (p *PrefStore) Get() string {
	key, r := p.root()
	if r == nil {
		return ""
	}
	defer p.release(r)
	if r.err != nil {
		return p.memGet(key)
	}
	parts, err := p.name()
	if err != nil {
		return ""
	}
	raw, err := readPrefFile(r.h, parts)
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

// Set validates and stores lang below the pinned root.
func (p *PrefStore) Set(lang string) error {
	b, err := Load()
	if err != nil {
		return err
	}
	lang = normalize(lang)
	if !b.has(lang) {
		return ErrUnsupported
	}
	key, r := p.root()
	if r == nil {
		return ErrNotPersisted
	}
	defer p.release(r) // runs after the Unlock deferred below
	if r.err != nil {
		p.mu.Lock()
		if p.mem == nil {
			p.mem = map[string]string{}
		}
		p.mem[key] = lang
		p.mu.Unlock()
		return fmt.Errorf("%w: %v", ErrNotPersisted, r.err)
	}
	parts, err := p.name()
	if err != nil {
		return err
	}
	raw, _ := json.Marshal(prefDoc{Lang: lang})
	p.mu.Lock() // serialize writers; operations below never block on FIFOs
	defer p.mu.Unlock()
	return writePrefFile(r.h, parts, raw)
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
				if errors.Is(err, ErrNotPersisted) {
					http.Error(w, "language preference could not be saved (kept for this session only)", http.StatusServiceUnavailable)
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
