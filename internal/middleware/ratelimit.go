package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// windowEntry holds the sliding-window state for a single rate-limit key.
type windowEntry struct {
	mu          sync.Mutex
	prevCount   int64
	currCount   int64
	windowStart time.Time
}

// RateLimiter is a sliding-window counter rate limiter backed by a sync.Map.
// It approximates a true sliding window using two consecutive fixed buckets:
//
//	rate ≈ prevCount*(1 - elapsed/window) + currCount
type RateLimiter struct {
	entries sync.Map
}

// NewRateLimiter creates a RateLimiter and starts a background goroutine that
// evicts stale entries every cleanupInterval.
func NewRateLimiter(cleanupInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{}
	go rl.cleanup(cleanupInterval)
	return rl
}

// Allow returns true if the request identified by key is within the rate limit,
// and increments the counter. Returns false (without incrementing) when the
// estimated rate equals or exceeds limit.
func (rl *RateLimiter) Allow(key string, limit int, window time.Duration) bool {
	now := time.Now()
	v, _ := rl.entries.LoadOrStore(key, &windowEntry{windowStart: now})
	entry := v.(*windowEntry)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	elapsed := now.Sub(entry.windowStart)

	switch {
	case elapsed >= 2*window:
		// Both buckets are stale — full reset.
		entry.prevCount = 0
		entry.currCount = 1
		entry.windowStart = now
		return true

	case elapsed >= window:
		// Rotate: current bucket becomes previous, open a new current bucket.
		entry.prevCount = entry.currCount
		entry.currCount = 0
		entry.windowStart = entry.windowStart.Add(window)
		elapsed = now.Sub(entry.windowStart)
	}

	ratio := float64(elapsed) / float64(window)
	rate := float64(entry.prevCount)*(1-ratio) + float64(entry.currCount)
	if rate >= float64(limit) {
		return false
	}

	entry.currCount++
	return true
}

func (rl *RateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		rl.entries.Range(func(k, v any) bool {
			entry := v.(*windowEntry)
			entry.mu.Lock()
			expired := now.Sub(entry.windowStart) >= 2*interval
			entry.mu.Unlock()
			if expired {
				rl.entries.Delete(k)
			}
			return true
		})
	}
}

// IPFromRequest extracts the client IP, respecting X-Forwarded-For for
// reverse-proxy deployments.
func IPFromRequest(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.SplitN(fwd, ",", 2)[0])
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		return host[:idx]
	}
	return host
}

func tooManyRequests(w http.ResponseWriter, window time.Duration) {
	w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "too many requests"})
}

// IPRateLimit returns middleware that limits requests by client IP.
func IPRateLimit(rl *RateLimiter, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow(IPFromRequest(r), limit, window) {
				tooManyRequests(w, window)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// MessageRateLimit returns middleware for POST /{room}/messages.
// It applies an IP-based limit and, when a pubkey is present in the JSON body,
// a separate per-pubkey limit. The request body is buffered so the handler
// can still read it.
func MessageRateLimit(rl *RateLimiter, ipLimit, pubkeyLimit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := IPFromRequest(r)
			if !rl.Allow("ip:"+ip, ipLimit, window) {
				tooManyRequests(w, window)
				return
			}

			if r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				_ = r.Body.Close()
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

				if err == nil && pubkeyLimit > 0 {
					var partial struct {
						Pubkey string `json:"pubkey"`
					}
					if json.Unmarshal(bodyBytes, &partial) == nil && partial.Pubkey != "" {
						if !rl.Allow("pubkey:"+partial.Pubkey, pubkeyLimit, window) {
							tooManyRequests(w, window)
							return
						}
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
