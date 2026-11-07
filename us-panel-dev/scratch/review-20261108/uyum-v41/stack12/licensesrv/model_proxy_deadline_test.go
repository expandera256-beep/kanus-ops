package licensesrv

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

type deadlineResponse struct {
	deadline time.Time
	flushErr error
}

func (d *deadlineResponse) Header() http.Header                { return http.Header{} }
func (d *deadlineResponse) WriteHeader(int)                    {}
func (d *deadlineResponse) Write(b []byte) (int, error)        { return len(b), nil }
func (d *deadlineResponse) SetWriteDeadline(v time.Time) error { d.deadline = v; return nil }
func (d *deadlineResponse) FlushError() error                  { return d.flushErr }
func TestProxyWriteDeadlineAndFlushFailure(t *testing.T) {
	w := &deadlineResponse{}
	if !proxyWrite(w, "data: test\n\n") || time.Until(w.deadline) <= 0 || time.Until(w.deadline) > 21*time.Second {
		t.Fatal(w.deadline)
	}
	w.flushErr = errors.New("client disconnected")
	if proxyWrite(w, "data: test\n\n") {
		t.Fatal("flush failure ignored")
	}
}
