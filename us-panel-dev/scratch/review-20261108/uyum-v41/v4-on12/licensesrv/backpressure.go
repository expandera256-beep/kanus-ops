package licensesrv

import (
	"net/http"
	"strconv"
)

// Header names follow PERF proposal 01 (Retry-After, X-US-Queue-Wait-Ms)
// plus X-US-Queue-Depth / X-US-Queue-Limit.
const (
	hdrQueueDepth = "X-US-Queue-Depth"
	hdrQueueLimit = "X-US-Queue-Limit"
)

// backpressureWriter guarantees that every 429 and 503 written through it
// carries Retry-After (unless the handler set a more specific one) and the
// current queue depth/limit.
type backpressureWriter struct {
	http.ResponseWriter
	retry        string
	depth, limit func() int
	wrote        bool
}

func withBackpressure(w http.ResponseWriter, retry string, depth, limit func() int) *backpressureWriter {
	return &backpressureWriter{ResponseWriter: w, retry: retry, depth: depth, limit: limit}
}

func (b *backpressureWriter) WriteHeader(code int) {
	if !b.wrote && (code == http.StatusTooManyRequests || code == http.StatusServiceUnavailable || code == http.StatusGatewayTimeout) {
		h := b.Header()
		if h.Get("Retry-After") == "" {
			h.Set("Retry-After", b.retry)
		}
		h.Set(hdrQueueDepth, strconv.Itoa(b.depth()))
		h.Set(hdrQueueLimit, strconv.Itoa(b.limit()))
	}
	b.wrote = true
	b.ResponseWriter.WriteHeader(code)
}

func (b *backpressureWriter) Write(p []byte) (int, error) {
	if !b.wrote {
		b.WriteHeader(http.StatusOK)
	}
	return b.ResponseWriter.Write(p)
}

func (b *backpressureWriter) Flush() {
	if f, ok := b.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (b *backpressureWriter) Unwrap() http.ResponseWriter { return b.ResponseWriter }
