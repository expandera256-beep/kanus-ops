package main

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

// Round 11 strict-DONE: the EXE must keep reading after finish_reason=stop
// until the gateway's [DONE] (closing early would make the hook release).
func TestStreamReaderReadsThroughDone(t *testing.T) {
	src := "data: {\"choices\":[{\"delta\":{\"content\":\"merhaba\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"YOK\"}}]}\n\n" +
		"data: [DONE]\n\nTRAILER"
	r := bufio.NewReader(strings.NewReader(src))
	content, _, errText := readModelStreamPolicy(r, nil, false)
	if errText != "" || content != "merhaba" {
		t.Fatalf("%q %q", content, errText)
	}
	rest, _ := io.ReadAll(r)
	if strings.TrimSpace(string(rest)) != "TRAILER" {
		t.Fatalf("[DONE] okunmadan durdu; kalan %q", rest)
	}
	// Non-strict gateway: stop then EOF without [DONE] is still a success.
	r2 := bufio.NewReader(strings.NewReader("data: {\"choices\":[{\"delta\":{\"content\":\"a\"},\"finish_reason\":\"stop\"}]}\n\n"))
	if c, _, e := readModelStreamPolicy(r2, nil, false); e != "" || c != "a" {
		t.Fatalf("%q %q", c, e)
	}
}
