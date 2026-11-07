package privacy

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Preferences holds optional consent flags. Defaults are false until set.
type Preferences struct {
	AccountID        string    `json:"accountId"`
	MarketingConsent bool      `json:"marketingConsent"`
	ProductUpdates   bool      `json:"productUpdates"`
	CookieAnalytics  bool      `json:"cookieAnalytics"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Source           string    `json:"source,omitempty"`
}

// PrefStore persists preferences beside the license store (sidecar JSON).
type PrefStore struct {
	mu   sync.Mutex
	path string
	byID map[string]Preferences
}

func OpenPrefStore(path string) (*PrefStore, error) {
	ps := &PrefStore{path: path, byID: map[string]Preferences{}}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ps, nil
		}
		return nil, err
	}
	if len(raw) == 0 {
		return ps, nil
	}
	if err := json.Unmarshal(raw, &ps.byID); err != nil {
		return nil, err
	}
	if ps.byID == nil {
		ps.byID = map[string]Preferences{}
	}
	return ps, nil
}

func (p *PrefStore) Get(accountID string) Preferences {
	p.mu.Lock()
	defer p.mu.Unlock()
	if v, ok := p.byID[accountID]; ok {
		return v
	}
	return Preferences{AccountID: accountID}
}

func (p *PrefStore) Put(pref Preferences, now time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	pref.UpdatedAt = now.UTC()
	p.byID[pref.AccountID] = pref
	return p.saveLocked()
}

func (p *PrefStore) Delete(accountID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.byID, accountID)
	return p.saveLocked()
}

func (p *PrefStore) AsMap(accountID string) map[string]any {
	v := p.Get(accountID)
	return map[string]any{
		"accountId":        v.AccountID,
		"marketingConsent": v.MarketingConsent,
		"productUpdates":   v.ProductUpdates,
		"cookieAnalytics":  v.CookieAnalytics,
		"updatedAt":        v.UpdatedAt,
		"source":           v.Source,
	}
}

func (p *PrefStore) saveLocked() error {
	if p.path == "" {
		return errors.New("privacy prefs path empty")
	}
	if err := os.MkdirAll(filepath.Dir(p.path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(p.byID, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p.path), ".privacy-prefs-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	tmp.Close()
	_ = os.Chmod(name, 0o600)
	return os.Rename(name, p.path)
}

func (p *PrefStore) Update(accountID string, marketing, product, cookie *bool, source string, now time.Time) error {
	cur := p.Get(accountID)
	cur.AccountID = accountID
	if marketing != nil {
		cur.MarketingConsent = *marketing
	}
	if product != nil {
		cur.ProductUpdates = *product
	}
	if cookie != nil {
		cur.CookieAnalytics = *cookie
	}
	if source != "" {
		cur.Source = source
	}
	return p.Put(cur, now)
}

func init() {
	OpenPrefStoreFunc = func(path string) (PreferenceStore, error) {
		return OpenPrefStore(path)
	}
}
