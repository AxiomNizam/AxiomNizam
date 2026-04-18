// Package workqueue — rate-limiter variants.
//
// The k8s client-go workqueue provides a family of rate limiters that
// differ only in how they compute the delay for a requeued item:
//
//	ItemExponentialFailureRateLimiter
//	    baseDelay * 2^(numRequeues-1), clamped at maxDelay
//	ItemFastSlowRateLimiter
//	    fastDelay for the first N requeues, slowDelay thereafter
//	BucketRateLimiter
//	    token bucket: returns the positive duration a request must
//	    wait to acquire a token, or 0 when tokens are available
//	MaxOfRateLimiter
//	    returns the max of several limiters — used to enforce both
//	    a per-item backoff and a global QPS ceiling simultaneously
//
// This file supplies each of the four.  They all satisfy the existing
// RateLimiter interface defined in queue.go.
package workqueue

import (
	"math"
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// ItemExponentialFailureRateLimiter
// -----------------------------------------------------------------------------

// ItemExponentialFailureRateLimiter is equivalent to DefaultRateLimiter
// but uses its tracked requeue count rather than the item's field, so
// callers can drive it from a queue that does not populate Item.RetryCount.
type ItemExponentialFailureRateLimiter struct {
	mu        sync.Mutex
	baseDelay time.Duration
	maxDelay  time.Duration
	failures  map[string]int
}

// NewItemExponentialFailureRateLimiter constructs the limiter.  A
// typical baseline pair is (5ms, 1000s) matching client-go defaults.
func NewItemExponentialFailureRateLimiter(baseDelay, maxDelay time.Duration) *ItemExponentialFailureRateLimiter {
	return &ItemExponentialFailureRateLimiter{
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		failures:  map[string]int{},
	}
}

// When returns the next retry delay.  The very first call for a key
// returns baseDelay; each subsequent call doubles until maxDelay.
func (r *ItemExponentialFailureRateLimiter) When(item *Item) time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	// failures is incremented *before* computing the delay so that the
	// first failure yields baseDelay (exponent == 0).
	exp := r.failures[item.Key]
	r.failures[item.Key] = exp + 1

	// math.Pow is used instead of repeated shift to correctly saturate
	// on large exponents without risking int64 overflow from a literal
	// 2<<63.  Once the computed value exceeds maxDelay we clamp.
	backoff := float64(r.baseDelay.Nanoseconds()) * math.Pow(2, float64(exp))
	if backoff > math.MaxInt64 || time.Duration(backoff) > r.maxDelay {
		return r.maxDelay
	}
	return time.Duration(backoff)
}

// Forget clears the failure count for key.
func (r *ItemExponentialFailureRateLimiter) Forget(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.failures, key)
}

// NumRequeues returns the current failure count for key.
func (r *ItemExponentialFailureRateLimiter) NumRequeues(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.failures[key]
}

// -----------------------------------------------------------------------------
// ItemFastSlowRateLimiter
// -----------------------------------------------------------------------------

// ItemFastSlowRateLimiter emits fastDelay for the first maxFastAttempts
// attempts of a key, then slowDelay forever after.  This matches the
// client-go limiter of the same name, and suits workloads that expect
// transient failures to clear quickly but want to back off aggressively
// if they persist.
type ItemFastSlowRateLimiter struct {
	mu              sync.Mutex
	fastDelay       time.Duration
	slowDelay       time.Duration
	maxFastAttempts int
	failures        map[string]int
}

// NewItemFastSlowRateLimiter constructs the limiter.
func NewItemFastSlowRateLimiter(fastDelay, slowDelay time.Duration, maxFastAttempts int) *ItemFastSlowRateLimiter {
	return &ItemFastSlowRateLimiter{
		fastDelay:       fastDelay,
		slowDelay:       slowDelay,
		maxFastAttempts: maxFastAttempts,
		failures:        map[string]int{},
	}
}

// When returns fastDelay while within the fast window, else slowDelay.
func (r *ItemFastSlowRateLimiter) When(item *Item) time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failures[item.Key]++
	if r.failures[item.Key] <= r.maxFastAttempts {
		return r.fastDelay
	}
	return r.slowDelay
}

// Forget clears accumulated failures for key.
func (r *ItemFastSlowRateLimiter) Forget(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.failures, key)
}

// NumRequeues returns the current failure count.
func (r *ItemFastSlowRateLimiter) NumRequeues(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.failures[key]
}

// -----------------------------------------------------------------------------
// BucketRateLimiter (token bucket)
// -----------------------------------------------------------------------------

// BucketRateLimiter enforces a global QPS ceiling.  Unlike the item-
// based limiters above, this one is stateless with respect to item
// identity — it simply returns how long a caller must wait before the
// next token is available.
//
// The implementation uses a lazy token-regeneration model: on every
// call we compute how many tokens should have accumulated since the
// last call, cap at burst, decrement by one, and return the wait time
// implied by any remaining deficit.
type BucketRateLimiter struct {
	mu        sync.Mutex
	qps       float64   // tokens per second
	burst     float64   // bucket capacity
	tokens    float64   // current token count
	lastCheck time.Time // when tokens were last updated
}

// NewBucketRateLimiter constructs a bucket producing qps tokens per
// second, capped at burst.  The bucket starts full.
func NewBucketRateLimiter(qps float64, burst int) *BucketRateLimiter {
	return &BucketRateLimiter{
		qps:       qps,
		burst:     float64(burst),
		tokens:    float64(burst),
		lastCheck: time.Now(),
	}
}

// When returns the wait time before a token is available.
func (r *BucketRateLimiter) When(item *Item) time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(r.lastCheck).Seconds()
	r.tokens += elapsed * r.qps
	if r.tokens > r.burst {
		r.tokens = r.burst
	}
	r.lastCheck = now

	if r.tokens >= 1 {
		r.tokens--
		return 0
	}
	// We need (1 - r.tokens) more tokens; wait time = deficit / qps.
	deficit := 1 - r.tokens
	r.tokens--
	return time.Duration(deficit / r.qps * float64(time.Second))
}

// Forget is a no-op for the bucket — it has no per-item state.
func (r *BucketRateLimiter) Forget(_ string) {}

// NumRequeues is always 0 — the bucket does not track per-item retries.
func (r *BucketRateLimiter) NumRequeues(_ string) int { return 0 }

// -----------------------------------------------------------------------------
// MaxOfRateLimiter — composition
// -----------------------------------------------------------------------------

// MaxOfRateLimiter returns the largest delay among its delegates and
// the max of their NumRequeues.  Use this to stack a global QPS
// ceiling on top of a per-item backoff:
//
//	rl := NewMaxOfRateLimiter(
//	    NewItemExponentialFailureRateLimiter(5*time.Millisecond, 60*time.Second),
//	    NewBucketRateLimiter(10, 100),
//	)
type MaxOfRateLimiter struct {
	limiters []RateLimiter
}

// NewMaxOfRateLimiter constructs a composite over the provided limiters.
func NewMaxOfRateLimiter(limiters ...RateLimiter) *MaxOfRateLimiter {
	return &MaxOfRateLimiter{limiters: limiters}
}

// When returns the max delay across all delegates.
func (r *MaxOfRateLimiter) When(item *Item) time.Duration {
	var max time.Duration
	for _, l := range r.limiters {
		if d := l.When(item); d > max {
			max = d
		}
	}
	return max
}

// Forget calls Forget on every delegate.
func (r *MaxOfRateLimiter) Forget(key string) {
	for _, l := range r.limiters {
		l.Forget(key)
	}
}

// NumRequeues returns the max across delegates.  Callers that need a
// specific delegate's count should query it directly instead.
func (r *MaxOfRateLimiter) NumRequeues(key string) int {
	var max int
	for _, l := range r.limiters {
		if n := l.NumRequeues(key); n > max {
			max = n
		}
	}
	return max
}

// DefaultControllerRateLimiter returns the rate limiter that k8s
// controllers use by default: exponential per-item backoff composed
// with a 10 QPS / 100-burst token bucket.  Provided here as a
// convenience constructor so callers don't need to hand-assemble it.
func DefaultControllerRateLimiter() RateLimiter {
	return NewMaxOfRateLimiter(
		NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		NewBucketRateLimiter(10, 100),
	)
}
