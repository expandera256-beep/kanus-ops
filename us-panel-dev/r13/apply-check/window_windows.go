//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/jchv/go-webview2"
)

func init() {
	runAppWindow = runWebView
	spawnAppWindow = spawnWebView
}

func spawnWebView(url string) {
	if os.Getenv("US_WINDOW") == "1" {
		return
	}
	exe, err := os.Executable()
	if err != nil || exe == "" {
		return
	}
	cmd := exec.Command(exe, "--window", url)
	cmd.Env = append(os.Environ(), "US_WINDOW=1", "US_STANDALONE=1")
	_ = cmd.Start()
}

func runWebView(url string) bool {
	data := filepath.Join(os.TempDir(), "us-window")
	_ = os.MkdirAll(data, 0o755)
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		AutoFocus: true,
		DataPath:  data,
		WindowOptions: webview2.WindowOptions{
			Title:  "US",
			Width:  1280,
			Height: 840,
			Center: true,
		},
	})
	runtime.KeepAlive(data)
	if w == nil {
		return false
	}
	defer w.Destroy()
	w.Navigate(url)
	w.Run()
	return true
}
