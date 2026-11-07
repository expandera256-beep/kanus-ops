package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGPUWarmingMessageShown(t *testing.T) {
	r := &http.Response{StatusCode: 503, Header: http.Header{"X-Us-Gpu-State": {"warming"}, "Retry-After": {"30"}},
		Body: io.NopCloser(strings.NewReader(`{"status":"warming","state":"warming","message":"GPU ısınıyor…","retry_after":30}`))}
	msg, ok := gpuWarmingMessage(r)
	if !ok || !strings.Contains(msg, "GPU ısınıyor") || !strings.Contains(msg, "30") {
		t.Fatalf("mesaj: %q %v", msg, ok)
	}
	plain := &http.Response{StatusCode: 503, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("x"))}
	if _, ok := gpuWarmingMessage(plain); ok {
		t.Fatal("GPU başlığı olmayan 503 ısınma sayıldı")
	}
}
