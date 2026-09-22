package auth

import (
	"testing"
	"time"
)

func TestLoginLimiterBlocksAfterMaxAttempts(t *testing.T) {
	limiter := NewLoginLimiter(2, time.Minute)
	key := "127.0.0.1"

	if !limiter.Allow(key) {
		t.Fatal("first attempt should be allowed")
	}
	limiter.RecordFailure(key)
	if !limiter.Allow(key) {
		t.Fatal("second attempt should be allowed")
	}
	limiter.RecordFailure(key)
	if limiter.Allow(key) {
		t.Fatal("third attempt should be blocked")
	}
}

func TestLoginLimiterResetClearsAttempts(t *testing.T) {
	limiter := NewLoginLimiter(1, time.Minute)
	key := "127.0.0.1"

	limiter.RecordFailure(key)
	if limiter.Allow(key) {
		t.Fatal("attempt should be blocked after failure")
	}
	limiter.Reset(key)
	if !limiter.Allow(key) {
		t.Fatal("attempt should be allowed after reset")
	}
}
