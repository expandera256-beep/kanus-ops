package main

import (
	"fmt"
	"time"
)

func (s *session) retarget(worker workerID) {
	if s.tracked == worker && s.worker == worker {
		s.worker = worker
		return
	}
	s.worker = worker
	s.tracked = worker
	s.loading = false
	s.seenAwake = false
	s.loadStart = time.Time{}
	s.loadStop = time.Time{}
	s.lastBeat = time.Time{}
}

func (s *session) markCallStart() {
	s.calling = true
	s.callStart = time.Now()
	s.callStop = time.Time{}
	s.lastBeat = s.callStart
	s.liveChars = 0
	s.modelStage = ""
	if hookLicenseMode() {
		s.modelStage = "GPU ısınıyor / sunucu kuyruğu · hazır durumu henüz doğrulanmadı"
		s.notice = s.modelStage
	}
	s.longStop = false
}

func (s *session) markCallStop() {
	if !s.calling && s.callStart.IsZero() {
		return
	}
	s.calling = false
	s.callStop = time.Now()
	s.lastBeat = s.callStop
	s.modelStage = ""
	s.longActive = false
}

func (s *session) noteLife(snap resourceSnap) {
	s.resources = snap
	if !snap.DoorOK {
		return
	}
	if s.tracked != s.worker {
		s.tracked = s.worker
		s.loading = false
		s.seenAwake = false
		s.loadStart = time.Time{}
		s.loadStop = time.Time{}
		s.lastBeat = time.Time{}
	}
	known, awake := listedAwake(snap.Workers, s.worker)
	if !known {
		return
	}
	if !awake {
		if !s.loading {
			s.loading = true
			s.loadStart = time.Now()
			s.loadStop = time.Time{}
		}
		s.seenAwake = false
		return
	}
	if s.loading {
		s.loading = false
		if s.loadStart.IsZero() {
			s.loadStart = time.Now()
		}
		s.loadStop = time.Now()
	}
	if !s.seenAwake {
		s.seenAwake = true
		if s.lastBeat.IsZero() {
			s.lastBeat = time.Now()
		}
	}
}

func listedAwake(workers []workerStatus, id workerID) (known, awake bool) {
	for _, item := range workers {
		if item.ID == string(id) {
			return true, item.Awake
		}
	}
	return false, false
}

func (s *session) lifeLines(now time.Time) (call, load, idle string, idleLeft int) {
	call = "yok"
	if s.calling && !s.callStart.IsZero() {
		call = "sürüyor · " + clockText(now.Sub(s.callStart))
		if s.liveChars > 0 {
			call += " · yazılıyor " + itoa(s.liveChars)
		}
	} else if !s.callStop.IsZero() && !s.callStart.IsZero() {
		call = "bitti · " + clockText(s.callStop.Sub(s.callStart))
	}

	known, awake := listedAwake(s.resources.Workers, s.worker)
	switch {
	case !s.resources.DoorOK && len(s.resources.Workers) == 0:
		load = "kapı yok"
	case s.resources.DoorOK && !known:
		load = "kapı bu işçiyi saymıyor"
	case s.loading && !s.loadStart.IsZero():
		load = "yükleniyor · " + clockText(now.Sub(s.loadStart))
	case !s.loadStop.IsZero() && !s.loadStart.IsZero():
		load = "yüklendi · " + clockText(s.loadStop.Sub(s.loadStart))
	case awake || s.seenAwake:
		load = "hazır"
	default:
		load = "bakılıyor"
	}

	if hookLicenseMode() && s.calling {
		if s.liveChars == 0 {
			load = "GPU ısınıyor / sunucu kuyruğu · bekleniyor"
		} else {
			load = "Model yanıtı akıyor"
		}
	}
	idle = "Otomatik kapanış durumu doğrulanmadı"
	idleLeft = 0
	for _, item := range s.resources.Managed {
		if item.Model != selectedModel(s.worker) {
			continue
		}
		if !item.Live || now.Unix()-item.ObservedAt > 60 || now.Unix() < item.ObservedAt {
			break
		}
		switch {
		case item.ProviderStatus == "TERMINATED":
			idle = "kapalı"
			load = "istem bekliyor"
		case item.Active > 0:
			idle = fmt.Sprintf("Sunucuda etkin istek: %d", item.Active)
			idleLeft = 100
		case item.ProviderStatus == "STOPPING":
			idle = "kapanış sürüyor"
		case item.IdleRemaining != nil:
			left := *item.IdleRemaining - int(now.Unix()-item.ObservedAt)
			if left < 0 {
				left = 0
			}
			if left == 0 {
				idle = "Kapanış denetimi bekleniyor"
			} else {
				idle = clockText(time.Duration(left)*time.Second) + " · sunucu kapanış eşiği"
			}
			idleLeft = remainPercent(time.Duration(item.IdleSeconds)*time.Second, time.Duration(item.IdleSeconds-left)*time.Second)
		default:
			idle = "Sunucu bekleme süresi henüz başlamadı"
		}
		break
	}
	return call, load, idle, idleLeft
}

func remainPercent(limit, elapsed time.Duration) int {
	if limit <= 0 {
		return 0
	}
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > limit {
		return 0
	}
	return int((limit - elapsed) * 100 / limit)
}

func clockText(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Second)
	sec := int(d / time.Second)
	minute, second := sec/60, sec%60
	if minute == 0 {
		return fmt.Sprintf("%d sn", second)
	}
	if second == 0 {
		return fmt.Sprintf("%d dk", minute)
	}
	return fmt.Sprintf("%d dk %d sn", minute, second)
}
