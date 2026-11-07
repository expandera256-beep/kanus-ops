package main

import (
	"net/http"
	"path/filepath"

	"kiyas/i18n"
)

// langPref stores the panel language next to the existing desk.json
// (honours US_DESK, defaults to <UserConfigDir>/us/lang.json).
var langPref = &i18n.PrefStore{Path: func() string {
	desk := deskPath()
	if desk == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(desk), "lang.json")
}}

// guiError writes a localized plain-text error. The language is the stored
// panel preference; without one it is TR, so default output is unchanged.
func guiError(w http.ResponseWriter, key string, code int) {
	msg := key
	if b, err := i18n.Load(); err == nil {
		msg = b.For(langPref.Get()).T(key)
	}
	http.Error(w, msg, code)
}
