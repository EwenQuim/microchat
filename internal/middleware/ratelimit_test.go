package middleware

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestIPFromRequest(t *testing.T) {
	tests := []struct {
		name       string
		forwarded  string
		remoteAddr string
		want       string
	}{
		{"x-forwarded-for single", "203.0.113.7", "10.0.0.1:1234", "203.0.113.7"},
		{"x-forwarded-for chain", "203.0.113.7, 70.41.3.18", "10.0.0.1:1234", "203.0.113.7"},
		{"x-forwarded-for with spaces", "  203.0.113.7  ,70.41.3.18", "10.0.0.1:1234", "203.0.113.7"},
		{"fallback to remote addr", "", "10.0.0.1:1234", "10.0.0.1"},
		{"remote addr without port", "", "10.0.0.1", "10.0.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.forwarded != "" {
				r.Header.Set("X-Forwarded-For", tt.forwarded)
			}
			if got := IPFromRequest(r); got != tt.want {
				t.Errorf("IPFromRequest() = %q, want %q", got, tt.want)
			}
		})
	}
}

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
