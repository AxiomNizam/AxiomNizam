package rate

import (
	"fmt"
	"sync"
	"time"
)

// RatePolicy defines rate limiting policies
type RatePolicy struct {
	ID            string
	Name          string
	Type          string
	Version       string
	Enabled       bool
	Limits        []RateLimit
	Algorithms    []RateLimitAlgorithm
	GlobalLimit   int64
	BurstSize     int64
	TimeWindow    time.Duration
	Description   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// RateLimit defines a rate limit rule
type RateLimit struct {
	ID            string
	Name          string
	Resource      string
	RequestsPerSecond int64
	BurstSize     int64
	Window        time.Duration
	Priority      int
	Condition     string
}

// RateLimitAlgorithm defines the rate limiting algorithm
type RateLimitAlgorithm struct {
	Type       string // "token-bucket", "sliding-window", "fixed-window", "leaky-bucket"
	Parameters map[string]interface{}
}

// GetID returns policy ID
func (rp *RatePolicy) GetID() string {
	return rp.ID
}

// GetName returns policy name
func (rp *RatePolicy) GetName() string {
	return rp.Name
}

// GetType returns policy type
func (rp *RatePolicy) GetType() string {
	return rp.Type
}

// GetVersion returns version
func (rp *RatePolicy) GetVersion() string {
	return rp.Version
}

// GetEnabled returns if enabled
func (rp *RatePolicy) GetEnabled() bool {
	return rp.Enabled
}

// Validate validates the policy
func (rp *RatePolicy) Validate() error {
	if rp.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}
	if rp.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}
	if len(rp.Limits) == 0 {
		return fmt.Errorf("at least one rate limit must be defined")
	}
	return nil
}

// RateLimiter manages rate limiting
type RateLimiter struct {
	mu              sync.RWMutex
	ratePolicies    map[string]*RatePolicy
	buckets         map[string]TokenBucket
	slidingWindows  map[string]SlidingWindow
	clientLimiters  map[string]*ClientLimiter
}

// TokenBucket implements token bucket algorithm
type TokenBucket struct {
	Capacity      int64
	Tokens        float64
	RefillRate    float64 // tokens per second
	LastRefillTime time.Time
}

// SlidingWindow implements sliding window algorithm
type SlidingWindow struct {
	Window        time.Duration
	RequestCount  int64
	WindowStart   time.Time
	Requests      []time.Time
}

// ClientLimiter tracks rate limits per client
type ClientLimiter struct {
	ClientID    string
	RateLimit   int64 // requests per second
	BurstSize   int64
	TokenBucket *TokenBucket
	Exceeded    bool
	ExceededAt  time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		ratePolicies:   make(map[string]*RatePolicy),
		buckets:        make(map[string]TokenBucket),
		slidingWindows: make(map[string]SlidingWindow),
		clientLimiters: make(map[string]*ClientLimiter),
	}
}

// RegisterRatePolicy registers a rate policy
func (rl *RateLimiter) RegisterRatePolicy(policy *RatePolicy) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if err := policy.Validate(); err != nil {
		return err
	}

	rl.ratePolicies[policy.ID] = policy
	return nil
}

// AllowRequest checks if a request should be allowed based on rate limits
func (rl *RateLimiter) AllowRequest(clientID string, rateLimit int64, burstSize int64) (bool, int64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.clientLimiters[clientID]
	if !exists {
		// Create new limiter for client
		limiter = &ClientLimiter{
			ClientID:  clientID,
			RateLimit: rateLimit,
			BurstSize: burstSize,
			TokenBucket: &TokenBucket{
				Capacity:       burstSize,
				Tokens:         float64(burstSize),
				RefillRate:     float64(rateLimit),
				LastRefillTime: time.Now(),
			},
		}
		rl.clientLimiters[clientID] = limiter
	}

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(limiter.TokenBucket.LastRefillTime).Seconds()
	tokensToAdd := elapsed * limiter.TokenBucket.RefillRate
	limiter.TokenBucket.Tokens = min(
		limiter.TokenBucket.Tokens+tokensToAdd,
		float64(limiter.TokenBucket.Capacity),
	)
	limiter.TokenBucket.LastRefillTime = now

	// Check if request allowed
	if limiter.TokenBucket.Tokens >= 1.0 {
		limiter.TokenBucket.Tokens -= 1.0
		limiter.Exceeded = false
		return true, int64(limiter.TokenBucket.Tokens)
	}

	// Rate limit exceeded
	if !limiter.Exceeded {
		limiter.Exceeded = true
		limiter.ExceededAt = now
	}

	return false, 0
}

// ResetClient resets rate limit for a client
func (rl *RateLimiter) ResetClient(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.clientLimiters, clientID)
}

// GetClientStatus returns the rate limit status for a client
func (rl *RateLimiter) GetClientStatus(clientID string) (int64, int64, bool, error) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limiter, exists := rl.clientLimiters[clientID]
	if !exists {
		return 0, 0, false, fmt.Errorf("client not found")
	}

	return limiter.RateLimit, int64(limiter.TokenBucket.Tokens), limiter.Exceeded, nil
}

// SlidingWindowLimiter implements sliding window rate limiting
type SlidingWindowLimiter struct {
	mu              sync.RWMutex
	windows         map[string][]time.Time
	requestLimit    int64
	window          time.Duration
}

// NewSlidingWindowLimiter creates a new sliding window limiter
func NewSlidingWindowLimiter(requestLimit int64, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windows:      make(map[string][]time.Time),
		requestLimit: requestLimit,
		window:       window,
	}
}

// Allow checks if a request is allowed
func (swl *SlidingWindowLimiter) Allow(identifier string) bool {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	now := time.Now()
	requests, exists := swl.windows[identifier]

	if !exists {
		requests = make([]time.Time, 0)
	}

	// Remove old requests outside the window
	validRequests := make([]time.Time, 0)
	for _, reqTime := range requests {
		if now.Sub(reqTime) < swl.window {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if new request allowed
	if int64(len(validRequests)) < swl.requestLimit {
		validRequests = append(validRequests, now)
		swl.windows[identifier] = validRequests
		return true
	}

	return false
}

// FixedWindowLimiter implements fixed window rate limiting
type FixedWindowLimiter struct {
	mu           sync.RWMutex
	counters     map[string]FixedWindowCounter
	requestLimit int64
	window       time.Duration
}

// FixedWindowCounter holds counter data
type FixedWindowCounter struct {
	Count      int64
	WindowStart time.Time
}

// NewFixedWindowLimiter creates a new fixed window limiter
func NewFixedWindowLimiter(requestLimit int64, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		counters:     make(map[string]FixedWindowCounter),
		requestLimit: requestLimit,
		window:       window,
	}
}

// Allow checks if a request is allowed
func (fwl *FixedWindowLimiter) Allow(identifier string) bool {
	fwl.mu.Lock()
	defer fwl.mu.Unlock()

	now := time.Now()
	counter, exists := fwl.counters[identifier]

	if !exists || now.Sub(counter.WindowStart) > fwl.window {
		// New window
		fwl.counters[identifier] = FixedWindowCounter{
			Count:       1,
			WindowStart: now,
		}
		return true
	}

	if counter.Count < fwl.requestLimit {
		counter.Count++
		fwl.counters[identifier] = counter
		return true
	}

	return false
}

// LeakyBucketLimiter implements leaky bucket algorithm
type LeakyBucketLimiter struct {
	mu         sync.RWMutex
	buckets    map[string]*LeakyBucket
	capacity   int64
	leakRate   float64 // units per second
}

// LeakyBucket represents a leaky bucket
type LeakyBucket struct {
	Water          float64
	LastLeakTime   time.Time
	Capacity       int64
	LeakRate       float64
}

// NewLeakyBucketLimiter creates a new leaky bucket limiter
func NewLeakyBucketLimiter(capacity int64, leakRate float64) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		buckets:  make(map[string]*LeakyBucket),
		capacity: capacity,
		leakRate: leakRate,
	}
}

// Allow checks if a request is allowed
func (lbl *LeakyBucketLimiter) Allow(identifier string, amount int64) bool {
	lbl.mu.Lock()
	defer lbl.mu.Unlock()

	now := time.Now()
	bucket, exists := lbl.buckets[identifier]

	if !exists {
		bucket = &LeakyBucket{
			Water:        0,
			LastLeakTime: now,
			Capacity:     lbl.capacity,
			LeakRate:     lbl.leakRate,
		}
	}

	// Calculate leaked water
	elapsed := now.Sub(bucket.LastLeakTime).Seconds()
	bucket.Water = max(0, bucket.Water-elapsed*bucket.LeakRate)
	bucket.LastLeakTime = now

	// Check if request fits
	if bucket.Water+float64(amount) <= float64(bucket.Capacity) {
		bucket.Water += float64(amount)
		lbl.buckets[identifier] = bucket
		return true
	}

	return false
}

// AdaptiveRateLimiter adapts rate limits based on system load
type AdaptiveRateLimiter struct {
	mu              sync.RWMutex
	baseLimiters    map[string]*ClientLimiter
	cpuThreshold    float64
	memoryThreshold float64
	currentLoad     float64
}

// NewAdaptiveRateLimiter creates a new adaptive rate limiter
func NewAdaptiveRateLimiter() *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		baseLimiters:    make(map[string]*ClientLimiter),
		cpuThreshold:    80.0,
		memoryThreshold: 85.0,
		currentLoad:     0,
	}
}

// SetSystemLoad sets current system load
func (arl *AdaptiveRateLimiter) SetSystemLoad(load float64) {
	arl.mu.Lock()
	defer arl.mu.Unlock()
	arl.currentLoad = load
}

// GetAdaptiveLimit returns adaptive limit based on load
func (arl *AdaptiveRateLimiter) GetAdaptiveLimit(baseLimit int64) int64 {
	arl.mu.RLock()
	defer arl.mu.RUnlock()

	if arl.currentLoad > arl.cpuThreshold {
		// Reduce limit under high load
		return int64(float64(baseLimit) * (1.0 - (arl.currentLoad / 100.0)))
	}

	return baseLimit
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
