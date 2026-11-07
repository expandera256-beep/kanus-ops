package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func archiveFixture(t *testing.T, bad string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	names := []string{"google-cloud-sdk/bin/gcloud.cmd", "google-cloud-sdk/platform/bundledpython/python.exe"}
	if bad != "" {
		names = append(names, bad)
	}
	for _, name := range names {
		f, e := w.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = f.Write([]byte("test")); e != nil {
			t.Fatal(e)
		}
	}
	if e := w.Close(); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}
func TestInstallVerifiesAndCachesArchive(t *testing.T) {
	data := archiveFixture(t, "")
	digest := sha256.Sum256(data)
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits++; w.Write(data) }))
	defer srv.Close()
	dir := t.TempDir()
	sha := hex.EncodeToString(digest[:])
	cli, py, err := installCloudArchive(dir, srv.Client(), srv.URL, sha, nil)
	if err != nil || !regularFile(cli) || !regularFile(py) {
		t.Fatal(err)
	}
	_, _, err = installCloudArchive(dir, srv.Client(), srv.URL, sha, nil)
	if err != nil || hits != 1 {
		t.Fatal(err, hits)
	}
}
func TestInstallRejectsTamperingAndTraversal(t *testing.T) {
	for _, tc := range []struct {
		bad       string
		wrongHash bool
	}{{"", true}, {"../escape", false}, {"google-cloud-sdk/../../escape", false}} {
		data := archiveFixture(t, tc.bad)
		digest := sha256.Sum256(data)
		sha := hex.EncodeToString(digest[:])
		if tc.wrongHash {
			sha = string(bytes.Repeat([]byte{'0'}, 64))
		}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write(data) }))
		dir := t.TempDir()
		_, _, err := installCloudArchive(dir, srv.Client(), srv.URL, sha, nil)
		srv.Close()
		if err == nil {
			t.Fatalf("bad archive accepted: %+v", tc)
		}
		if _, err := os.Stat(filepath.Join(dir, "escape")); !os.IsNotExist(err) {
			t.Fatal("unexpected output")
		}
	}
}
