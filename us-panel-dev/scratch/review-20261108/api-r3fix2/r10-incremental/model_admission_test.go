package main

import (
	"context"
	"testing"
	"time"
)

func TestAdmissionBoundedQueueCancelAndRelease(t *testing.T) {
	g := newAdmission(1, 1)
	release, err := g.acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		r, e := g.acquire(ctx)
		if e == nil {
			r()
		}
		done <- e
	}()
	deadline := time.Now().Add(time.Second)
	for g.snapshot().Queued != 1 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if _, err := g.acquire(context.Background()); err == nil {
		t.Fatal("overflow admitted")
	}
	cancel()
	if <-done == nil {
		t.Fatal("cancel admitted")
	}
	release()
	release()
	if g.snapshot().Active != 0 || g.snapshot().Queued != 0 {
		t.Fatal(g.snapshot())
	}
}
func TestAdmissionInvalidConfigFailsClosed(t *testing.T) {
	t.Setenv("US_MODEL_CONCURRENCY", "0")
	if _, err := configuredAdmission().acquire(context.Background()); err == nil {
		t.Fatal("invalid capacity")
	}
}
