package main

import (
	"strings"
	"testing"
	"time"
)

func TestLifeBarShowsCallLoadAndIdle(t *testing.T) {
	s := newView("a100")
	now := time.Date(2026, 11, 3, 15, 0, 0, 0, time.UTC)
	s.calling = true
	s.callStart = now.Add(-42 * time.Second)
	call, _, idle, full := s.lifeLines(now)
	if !strings.Contains(call, "sürüyor") || !strings.Contains(call, "42 sn") || !strings.Contains(idle, "kapanmaz") || full != 100 {
		t.Fatalf("%s %s %d", call, idle, full)
	}
	s.liveChars = 380
	call, _, _, _ = s.lifeLines(now)
	if !strings.Contains(call, "yazılıyor 380") {
		t.Fatal(call)
	}

	s.calling = false
	s.resources.DoorOK = true
	s.resources.Workers = []workerStatus{{ID: "a100", Awake: false}}
	s.loading = true
	s.loadStart = now.Add(-time.Minute - 5*time.Second)
	_, load, idle, _ := s.lifeLines(now)
	if !strings.Contains(load, "yükleniyor") || !strings.Contains(load, "1 dk 5 sn") || !strings.Contains(idle, "70 dk") {
		t.Fatalf("%s %s", load, idle)
	}

	s.loading = false
	s.seenAwake = true
	s.loadStart = now.Add(-3*time.Minute - 20*time.Second)
	s.loadStop = now
	s.lastBeat = now.Add(-2 * time.Minute)
	s.resources.Workers[0].Awake = true
	_, load, idle, left := s.lifeLines(now)
	if !strings.Contains(load, "yüklendi") || !strings.Contains(load, "3 dk 20 sn") || !strings.Contains(idle, "sonra kapanır") || left <= 0 || left >= 100 {
		t.Fatalf("%s %s %d", load, idle, left)
	}

	s.worker = "q8"
	s.tracked = "q8"
	s.resources.Workers = []workerStatus{{ID: "q8", Awake: true}}
	s.lastBeat = now.Add(-time.Minute)
	_, _, idle, _ = s.lifeLines(now)
	if !strings.Contains(idle, "sonra uyur") {
		t.Fatal(idle)
	}

	s.lastBeat = now.Add(-16 * time.Minute)
	_, _, idle, left = s.lifeLines(now)
	if !strings.Contains(idle, "uyuma zamanı geldi") || left != 0 {
		t.Fatalf("%s %d", idle, left)
	}
}
