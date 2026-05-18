package cache

import (
	"context"
	"sync"
	"time"
)

// SlidingWindowLimiter implements a sliding window rate limiter.
type SlidingWindowLimiter struct {
	mu       sync.Mutex
	windows  map[string]*window
	limit    int
	interval time.Duration
}

type window struct {
	current  int
	previous int
	start    time.Time
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter.
func NewSlidingWindowLimiter(limit int, interval time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windows:  make(map[string]*window),
		limit:    limit,
		interval: interval,
	}
}

// Allow checks if a request should be allowed under the rate limit.
func (l *SlidingWindowLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	w, exists := l.windows[key]
	if !exists {
		l.windows[key] = &window{
			current: 1,
			start:   now,
		}
		return true
	}

	// Check if we've moved to a new window
	elapsed := now.Sub(w.start)
	if elapsed >= l.interval {
		// Move to new window
		w.previous = w.current
		w.current = 0
		w.start = now
		elapsed = 0
	}

	// Calculate weighted count using sliding window
	weight := float64(l.interval-elapsed) / float64(l.interval)
	count := float64(w.previous)*weight + float64(w.current)

	if int(count) >= l.limit {
		return false
	}

	w.current++
	return true
}

// Reset clears the rate limit for a key.
func (l *SlidingWindowLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.windows, key)
}

// ResetAll clears all rate limits.
func (l *SlidingWindowLimiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.windows = make(map[string]*window)
}

// InMemoryRateLimiter is a simple in-memory rate limiter for testing.
type InMemoryRateLimiter struct {
	mu      sync.Mutex
	counts  map[string]int
	limits  map[string]int
	windows map[string]time.Time
	ttl     time.Duration
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter.
func NewInMemoryRateLimiter(ttl time.Duration) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		counts:  make(map[string]int),
		limits:  make(map[string]int),
		windows: make(map[string]time.Time),
		ttl:     ttl,
	}
}

// CheckAndIncrement checks if a request is allowed and increments the counter.
func (r *InMemoryRateLimiter) CheckAndIncrement(ctx context.Context, key string, limit int) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	start, exists := r.windows[key]

	if !exists || now.Sub(start) > r.ttl {
		r.windows[key] = now
		r.counts[key] = 1
		r.limits[key] = limit
		return true, nil
	}

	if r.counts[key] >= limit {
		return false, nil
	}

	r.counts[key]++
	return true, nil
}
