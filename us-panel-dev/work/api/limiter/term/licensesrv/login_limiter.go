package licensesrv

import (
	"strconv"
	"sync"
	"time"
)

// accountLimiter stores opaque, fixed-length keys. It never evicts live entries:
// saturation refuses unknown keys rather than forgiving an attacker's target.
// Failure history expires only after window of quiet, not at a fixed boundary.
type accountLimiter struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	maxKeys   int
	m         map[string]*acctBucket
	lastSweep time.Time
}
type acctBucket struct {
	start time.Time // most recent counted failure
	fails int
	until time.Time
}

func newAccountLimiter() *accountLimiter {
	return &accountLimiter{limit: envInt("US_LOGIN_PAIR_THRESHOLD", 5, 3, 20), window: time.Duration(envInt("US_LOGIN_FAILURE_QUIET_SECONDS", 1800, 1800, 86400)) * time.Second, maxKeys: 10000, m: map[string]*acctBucket{}}
}
func (l *accountLimiter) blocked(key string, now time.Time) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b := l.m[key]
	if b != nil && now.Sub(b.start) >= l.window {
		delete(l.m, key)
		b = nil
	}
	if b != nil {
		if now.Before(b.until) {
			return true, secondsLeft(b.until, now)
		}
		return false, 0
	}
	if len(l.m) >= l.maxKeys {
		l.evictLocked(now)
		if len(l.m) >= l.maxKeys {
			return true, 5
		}
	}
	return false, 0
}
func secondsLeft(until, now time.Time) int {
	n := int((until.Sub(now) + time.Second - 1) / time.Second)
	if n < 1 {
		return 1
	}
	return n
}
func (l *accountLimiter) fail(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lastSweep.IsZero() || now.Sub(l.lastSweep) >= time.Minute {
		l.evictLocked(now)
		l.lastSweep = now
	}
	b := l.m[key]
	if b == nil || now.Sub(b.start) >= l.window {
		if b == nil && len(l.m) >= l.maxKeys {
			return
		}
		b = &acctBucket{}
		l.m[key] = b
	}
	// Saturating count also bounds the exponential shift.
	if b.fails < l.limit+10 {
		b.fails++
	}
	b.start = now
	if b.fails >= l.limit {
		delay := time.Second * time.Duration(1<<uint(b.fails-l.limit))
		if delay > 15*time.Minute {
			delay = 15 * time.Minute
		}
		b.until = now.Add(delay)
	}
}
func (l *accountLimiter) reset(key string) { l.mu.Lock(); delete(l.m, key); l.mu.Unlock() }
func (l *accountLimiter) evictLocked(now time.Time) {
	for k, b := range l.m {
		if now.Sub(b.start) >= l.window {
			delete(l.m, k)
		}
	}
}
func (l *accountLimiter) size() int  { l.mu.Lock(); defer l.mu.Unlock(); return len(l.m) }
func retryAfterSeconds(n int) string { return strconv.Itoa(n) }
