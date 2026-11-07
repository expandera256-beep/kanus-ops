package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
)

// Per-desktop admission control, NOT a cross-client GPU scheduler. A trusted
// server-side model proxy is still required to enforce fleet-wide capacity.
type admissionState struct {
	Active   int    `json:"active"`
	Queued   int    `json:"queued"`
	Capacity int    `json:"capacity"`
	Error    string `json:"error,omitempty"`
}
type admissionGate struct {
	slots, requests chan struct{}
	err             error
}

func newAdmission(capacity, queued int) *admissionGate {
	return &admissionGate{slots: make(chan struct{}, capacity), requests: make(chan struct{}, capacity+queued)}
}
func configuredAdmission() *admissionGate {
	values := []int{1, 32}
	for i, key := range []string{"US_MODEL_CONCURRENCY", "US_MODEL_QUEUE_LIMIT"} {
		if raw := os.Getenv(key); raw != "" {
			v, err := strconv.Atoi(raw)
			max := 100
			if i == 0 {
				max = 10
			}
			if err != nil || v < 1 || v > max {
				return &admissionGate{err: fmt.Errorf("%s geçersiz; model çağrılmadı.", key)}
			}
			values[i] = v
		}
	}
	return newAdmission(values[0], values[1])
}

var modelAdmission = configuredAdmission()

func (g *admissionGate) acquire(ctx context.Context) (func(), error) {
	if g.err != nil {
		return nil, g.err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case g.requests <- struct{}{}:
	default:
		return nil, fmt.Errorf("Yerel model kuyruğu dolu; daha sonra yeniden deneyin.")
	}
	select {
	case g.slots <- struct{}{}:
	case <-ctx.Done():
		<-g.requests
		return nil, ctx.Err()
	}
	var once sync.Once
	release := func() { once.Do(func() { <-g.slots; <-g.requests }) }
	if err := ctx.Err(); err != nil {
		release()
		return nil, err
	}
	return release, nil
}
func (g *admissionGate) snapshot() admissionState {
	if g.err != nil {
		return admissionState{Error: g.err.Error()}
	}
	active := len(g.slots)
	queued := len(g.requests) - active
	if queued < 0 {
		queued = 0
	}
	return admissionState{Active: active, Queued: queued, Capacity: cap(g.slots)}
}
