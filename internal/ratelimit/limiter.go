package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type Limiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     rate.Limit
	burst    int
	ttl      time.Duration
}

func New(r rate.Limit, burst int, ttl time.Duration) *Limiter {
	l := &Limiter{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    burst,
		ttl:      ttl,
	}

	go l.cleanupLook()
	return l
}

func (l *Limiter) Allow(Key string) bool {
	l.mu.Lock()

	v, exits := l.visitors[Key]
	if !exits {
		v = &visitor{limiter: rate.NewLimiter(l.rate, l.burst)}
		l.visitors[Key] = v
	}

	v.lastSeen = time.Now()
	l.mu.Unlock()

	return v.limiter.Allow()
}

func (l *Limiter) cleanupLook() {
	ticker := time.NewTicker(time.Minute)

	for range ticker.C {
		l.mu.Lock()

		for key, v := range l.visitors {
			if time.Since(v.lastSeen) > l.ttl {
				delete(l.visitors, key)
			}
		}
		l.mu.Unlock()
	}
}
