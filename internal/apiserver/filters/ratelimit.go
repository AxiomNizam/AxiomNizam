// Package filters — rate limiting.
//
// A token-bucket limiter applied per client identity (User.Name or
// source IP for anonymous).  Used to shed load during traffic spikes
// and to bound the damage a single runaway client can do.
//
// The implementation is deliberately simple — one bucket per key,
// LRU-evicted when the bucket count exceeds the configured maximum.
// Callers who need richer semantics (priority-fair queuing, shaped
// per-verb limits) should compose this with a dedicated library.
package filters

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiterOptions configures the middleware.
type RateLimiterOptions struct {
	// RequestsPerSecond is the steady-state fill rate per bucket.
	RequestsPerSecond float64
	// Burst is the bucket capacity.
	Burst int
	// MaxBuckets caps the number of distinct keys tracked.  Zero
	// means unbounded — useful in tests, dangerous in production.
	MaxBuckets int
	// KeyFunc returns the rate-limit identity for a request.  Nil
	// defaults to UserFrom(ctx).Name falling back to remote address.
	KeyFunc func(r *http.Request) string
}

// RateLimit returns the middleware.
func RateLimit(opts RateLimiterOptions) func(http.Handler) http.Handler {
	if opts.KeyFunc == nil {
		opts.KeyFunc = defaultLimitKey
	}
	if opts.Burst <= 0 {
		opts.Burst = 10
	}
	if opts.RequestsPerSecond <= 0 {
		opts.RequestsPerSecond = 5
	}
	l := &limiter{opts: opts, buckets: map[string]*bucket{}}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := opts.KeyFunc(r)
			if !l.allow(key) {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

// defaultLimitKey prefers the authenticated user name, falling back
// to RemoteAddr.  Anonymous clients share the same bucket per IP.
func defaultLimitKey(r *http.Request) string {
	if u := UserFrom(r.Context()); u != nil && u.Name != "" {
		return "user:" + u.Name
	}
	return "ip:" + clientIP(r)
}

// bucket is one caller's token bucket.
type bucket struct {
	tokens   float64
	lastFill time.Time
}

// limiter is the shared state.
type limiter struct {
	opts    RateLimiterOptions
	mu      sync.Mutex
	buckets map[string]*bucket
	order   []string // LRU order, newest last
}

// allow charges one token if available.
func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: float64(l.opts.Burst), lastFill: now}
		l.buckets[key] = b
		l.order = append(l.order, key)
		l.evictIfNeeded()
	} else {
		elapsed := now.Sub(b.lastFill).Seconds()
		b.tokens = minF(float64(l.opts.Burst), b.tokens+elapsed*l.opts.RequestsPerSecond)
		b.lastFill = now
	}
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// evictIfNeeded drops the oldest bucket when the cap is exceeded.
// Called under l.mu.
func (l *limiter) evictIfNeeded() {
	if l.opts.MaxBuckets <= 0 || len(l.buckets) <= l.opts.MaxBuckets {
		return
	}
	// O(n) LRU: shift the slice.  Acceptable because eviction is
	// rare compared to allow() calls.
	oldest := l.order[0]
	l.order = l.order[1:]
	delete(l.buckets, oldest)
}

// minF is a local helper; math.Min allocates.
func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
