package main

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"kiyas/i18n"
)

// langPref stores the panel language next to the existing desk.json
// (honours US_DESK, defaults to <UserConfigDir>/us/lang.json). The desk
// directory is the trusted root: resolved once (symlinked HOME/config is
// fine) and pinned as a handle; nothing below it may be a symlink.
var langPref = &i18n.PrefStore{Name: "lang.json", Root: func() string {
	desk := deskPath()
	if desk == "" {
		return ""
	}
	return filepath.Dir(desk)
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

// sameOriginRequest is the shared Origin check for mutating panel routes
// (/api/act, /api/lang): an Origin header, when present, must be
// http://<same host>; requests without Origin (non-browser) are allowed.
func sameOriginRequest(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	return err == nil && u.Scheme == "http" && strings.EqualFold(u.Host, r.Host)
}
