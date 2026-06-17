package auth

import (
	"sync"
	"time"
)

type LoginLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	max      int
	window   time.Duration
}

func NewLoginLimiter(max int, window time.Duration) *LoginLimiter {
	return &LoginLimiter{
		attempts: make(map[string][]time.Time),
		max:      max,
		window:   window,
	}
}

func (l *LoginLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)
	entries := l.attempts[key]
	filtered := entries[:0]
	for _, t := range entries {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	l.attempts[key] = filtered
	return len(filtered) < l.max
}

func (l *LoginLimiter) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.attempts[key] = append(l.attempts[key], time.Now())
}

func (l *LoginLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}
