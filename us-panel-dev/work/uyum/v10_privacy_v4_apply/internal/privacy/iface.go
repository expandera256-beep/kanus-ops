package privacy

import "time"

// PreferenceStore is optional; wiring may pass nil.
type PreferenceStore interface {
	AsMap(accountID string) map[string]any
	Update(accountID string, marketing, product, cookie *bool, source string, now time.Time) error
	Delete(accountID string) error
}

// OpenPrefStoreFunc is set by prefs.go init when the preferences implementation is present.
var OpenPrefStoreFunc func(path string) (PreferenceStore, error)
