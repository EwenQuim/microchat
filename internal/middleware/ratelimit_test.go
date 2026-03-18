package middleware

import (
	"testing"
	"time"
)

func TestAllow_UnderLimit(t *testing.T) {
	rl := NewRateLimiter(time.Minute)
	for i := range 5 {
		if !rl.Allow("key", 10, time.Minute) {
			t.Fatalf("request %d should be allowed (under limit)", i+1)
		}
	}
}

func TestAllow_AtLimit(t *testing.T) {
	rl := NewRateLimiter(time.Minute)
	for range 10 {
		rl.Allow("key", 10, time.Minute)
	}
	if rl.Allow("key", 10, time.Minute) {
		t.Fatal("request over limit should be rejected")
	}
}

func TestAllow_SlidingWindow(t *testing.T) {
	rl := NewRateLimiter(time.Minute)
	window := 200 * time.Millisecond
	limit := 10

	// Fill window 1 completely.
	for range limit {
		rl.Allow("key", limit, window)
	}
	// Must be rejected at limit.
	if rl.Allow("key", limit, window) {
		t.Fatal("should be rejected at limit")
	}

	// Move into window 2 (just past the boundary).
	time.Sleep(window + 20*time.Millisecond)

	// Sliding window: previous count carries over, so fewer than limit requests
	// should be allowed at the start of the new window.
	allowed := 0
	for range limit {
		if rl.Allow("key", limit, window) {
			allowed++
		}
	}
	if allowed == limit {
		t.Errorf("sliding window should carry over previous count; got %d/%d allowed", allowed, limit)
	}
	if allowed == 0 {
		t.Error("some requests should be allowed at the start of a new window")
	}
}

func TestAllow_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(time.Minute)

	// Exhaust key "a".
	for range 10 {
		rl.Allow("a", 10, time.Minute)
	}
	if rl.Allow("a", 10, time.Minute) {
		t.Fatal("key 'a' should be rate limited")
	}

	// Key "b" must be independent.
	if !rl.Allow("b", 10, time.Minute) {
		t.Fatal("key 'b' should not be rate limited")
	}
}

func TestCleanup(t *testing.T) {
	interval := 50 * time.Millisecond
	rl := NewRateLimiter(interval)

	rl.Allow("key", 10, interval)

	if _, ok := rl.entries.Load("key"); !ok {
		t.Fatal("entry should exist before cleanup")
	}

	// Wait long enough for the entry to expire (>2*interval) and cleanup to run.
	time.Sleep(4 * interval)

	if _, ok := rl.entries.Load("key"); ok {
		t.Fatal("entry should have been cleaned up after expiry")
	}
}
