package licensesrv

import (
	"log"
	"strconv"
	"sync"
	"time"
)

// accountLimiter counts FAILED logins per normalized email, independent of
// the client IP (rotating IPs cannot bypass it). Unknown emails are counted
// exactly like registered ones, so the limiter does not reveal whether an
// account exists. A successful login resets the counter (only the correct
// password can do that, so it gives an attacker nothing; a real user who
// mistyped a few times is not punished afterwards). Memory: at most maxKeys
// buckets; expired buckets are swept, then the oldest is evicted.
type accountLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	maxKeys int
	m       map[string]*acctBucket
}

type acctBucket struct {
	start time.Time
	fails int
}

func newAccountLimiter() *accountLimiter {
	return &accountLimiter{
		limit:   envInt("US_LOGIN_ACCOUNT_LIMIT", 10, 1, 1000),
		window:  time.Duration(envInt("US_LOGIN_ACCOUNT_WINDOW_SECONDS", 900, 10, 86400)) * time.Second,
		maxKeys: 10000,
		m:       map[string]*acctBucket{},
	}
}

// blocked reports whether email is locked and for how many seconds.
func (l *accountLimiter) blocked(email string, now time.Time) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b := l.m[email]
	if b == nil {
		return false, 0
	}
	if now.Sub(b.start) >= l.window {
		delete(l.m, email)
		return false, 0
	}
	if b.fails >= l.limit {
		left := int((l.window - now.Sub(b.start)).Seconds())
		if left < 1 {
			left = 1
		}
		return true, left
	}
	return false, 0
}

func (l *accountLimiter) fail(email string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b := l.m[email]
	if b == nil || now.Sub(b.start) >= l.window {
		if len(l.m) >= l.maxKeys {
			l.evictLocked(now)
		}
		b = &acctBucket{start: now}
		l.m[email] = b
	}
	b.fails++
	if b.fails == l.limit {
		log.Printf("account login limiter: %s locked for %s after %d failures", maskEmail(email), l.window, b.fails)
	}
}

func (l *accountLimiter) reset(email string) {
	l.mu.Lock()
	delete(l.m, email)
	l.mu.Unlock()
}

func (l *accountLimiter) evictLocked(now time.Time) {
	var oldestK string
	var oldest time.Time
	for k, b := range l.m {
		if now.Sub(b.start) >= l.window {
			delete(l.m, k)
			continue
		}
		if oldestK == "" || b.start.Before(oldest) {
			oldestK, oldest = k, b.start
		}
	}
	if len(l.m) >= l.maxKeys && oldestK != "" {
		delete(l.m, oldestK)
	}
}

func (l *accountLimiter) size() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.m)
}

func retryAfterSeconds(n int) string { return strconv.Itoa(n) }
