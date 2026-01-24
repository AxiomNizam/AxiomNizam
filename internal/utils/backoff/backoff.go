package backoff

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// Strategy defines the backoff strategy interface
type Strategy interface {
	// NextDuration returns the next backoff duration for the given attempt
	NextDuration(attempt int) time.Duration
	// Reset resets the backoff strategy
	Reset()
}

// ExponentialBackoff implements exponential backoff with jitter
type ExponentialBackoff struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	maxAttempts     int
	currentAttempt  int
}

// NewExponentialBackoff creates a new exponential backoff strategy
func NewExponentialBackoff(initial, max time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialInterval: initial,
		MaxInterval:     max,
		Multiplier:      2.0,
		currentAttempt:  0,
	}
}

// NextDuration returns the next backoff duration
func (eb *ExponentialBackoff) NextDuration(attempt int) time.Duration {
	interval := time.Duration(float64(eb.InitialInterval) * math.Pow(eb.Multiplier, float64(attempt-1)))
	if interval > eb.MaxInterval {
		interval = eb.MaxInterval
	}
	return interval
}

// Reset resets the backoff state
func (eb *ExponentialBackoff) Reset() {
	eb.currentAttempt = 0
}

// WithJitter applies jitter to the backoff duration
func (eb *ExponentialBackoff) WithJitter() time.Duration {
	duration := eb.NextDuration(eb.currentAttempt)
	jitter := time.Duration(rand.Int63n(int64(duration)))
	return duration - jitter
}

// LinearBackoff implements linear backoff
type LinearBackoff struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Increment       time.Duration
}

// NewLinearBackoff creates a new linear backoff strategy
func NewLinearBackoff(initial, max, increment time.Duration) *LinearBackoff {
	return &LinearBackoff{
		InitialInterval: initial,
		MaxInterval:     max,
		Increment:       increment,
	}
}

// NextDuration returns the next backoff duration
func (lb *LinearBackoff) NextDuration(attempt int) time.Duration {
	interval := lb.InitialInterval + lb.Increment*time.Duration(attempt-1)
	if interval > lb.MaxInterval {
		interval = lb.MaxInterval
	}
	return interval
}

// Reset resets the backoff state
func (lb *LinearBackoff) Reset() {}

// FibonacciBackoff implements Fibonacci backoff sequence
type FibonacciBackoff struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	prev            time.Duration
	current         time.Duration
}

// NewFibonacciBackoff creates a new Fibonacci backoff strategy
func NewFibonacciBackoff(initial, max time.Duration) *FibonacciBackoff {
	return &FibonacciBackoff{
		InitialInterval: initial,
		MaxInterval:     max,
		prev:            0,
		current:         initial,
	}
}

// NextDuration returns the next backoff duration
func (fb *FibonacciBackoff) NextDuration(attempt int) time.Duration {
	interval := fb.current
	if interval > fb.MaxInterval {
		interval = fb.MaxInterval
	}

	// Calculate next Fibonacci value
	next := fb.current + fb.prev
	fb.prev = fb.current
	fb.current = next

	return interval
}

// Reset resets the backoff state
func (fb *FibonacciBackoff) Reset() {
	fb.prev = 0
	fb.current = fb.InitialInterval
}

// FullJitter applies full jitter to backoff duration
func FullJitter(duration time.Duration) time.Duration {
	return time.Duration(rand.Int63n(int64(duration)))
}

// EqualJitter applies equal jitter to backoff duration
func EqualJitter(duration time.Duration) time.Duration {
	half := duration / 2
	jitter := time.Duration(rand.Int63n(int64(half)))
	return half + jitter
}

// DecorrelatedJitter applies decorrelated jitter to backoff duration
func DecorrelatedJitter(prevDuration, maxDuration time.Duration) time.Duration {
	temp := 3 * prevDuration
	if temp > maxDuration {
		temp = maxDuration
	}
	return time.Duration(rand.Int63n(int64(temp)))
}

// Wait implements a wait with backoff strategy
func Wait(ctx context.Context, strategy Strategy, attempt int) error {
	duration := strategy.NextDuration(attempt)

	select {
	case <-time.After(duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Controller manages backoff with context awareness
type Controller struct {
	strategy Strategy
	maxWait  time.Duration
	deadline time.Time
}

// NewController creates a new backoff controller
func NewController(strategy Strategy, maxWait time.Duration) *Controller {
	return &Controller{
		strategy: strategy,
		maxWait:  maxWait,
		deadline: time.Now().Add(maxWait),
	}
}

// Next returns the next backoff duration respecting context deadline
func (c *Controller) Next(ctx context.Context, attempt int) (time.Duration, error) {
	duration := c.strategy.NextDuration(attempt)

	// Respect context deadline
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < duration {
			duration = remaining
		}
		if remaining <= 0 {
			return 0, ctx.Err()
		}
	}

	// Respect max wait
	if duration > c.maxWait {
		duration = c.maxWait
	}

	return duration, nil
}

// Wait waits for the next backoff duration
func (c *Controller) Wait(ctx context.Context, attempt int) error {
	duration, err := c.Next(ctx, attempt)
	if err != nil {
		return err
	}

	if duration == 0 {
		return nil
	}

	select {
	case <-time.After(duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Reset resets the controller
func (c *Controller) Reset() {
	c.strategy.Reset()
	c.deadline = time.Now().Add(c.maxWait)
}

// AdaptiveBackoff adjusts backoff based on success/failure
type AdaptiveBackoff struct {
	strategy         Strategy
	successCount     int
	failureCount     int
	aggressionFactor float64
}

// NewAdaptiveBackoff creates a new adaptive backoff strategy
func NewAdaptiveBackoff(strategy Strategy) *AdaptiveBackoff {
	return &AdaptiveBackoff{
		strategy:         strategy,
		aggressionFactor: 0.9, // Reduce backoff after success
	}
}

// NextDuration returns the next backoff duration
func (ab *AdaptiveBackoff) NextDuration(attempt int) time.Duration {
	baseDuration := ab.strategy.NextDuration(attempt)

	// Reduce backoff after successful attempts
	if ab.successCount > 0 {
		reduction := math.Pow(ab.aggressionFactor, float64(ab.successCount))
		baseDuration = time.Duration(float64(baseDuration) * reduction)
	}

	return baseDuration
}

// Reset resets the adaptive backoff state
func (ab *AdaptiveBackoff) Reset() {
	ab.strategy.Reset()
	ab.successCount = 0
	ab.failureCount = 0
}

// RecordSuccess records a successful attempt
func (ab *AdaptiveBackoff) RecordSuccess() {
	ab.successCount++
}

// RecordFailure records a failed attempt
func (ab *AdaptiveBackoff) RecordFailure() {
	ab.failureCount++
	ab.successCount = 0 // Reset success count on failure
}

// DeadlineExceeded checks if deadline has been exceeded
func DeadlineExceeded(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// TimeUntilDeadline returns time remaining until deadline
func TimeUntilDeadline(ctx context.Context) time.Duration {
	if deadline, ok := ctx.Deadline(); ok {
		return time.Until(deadline)
	}
	return 0
}

// ShouldRetry determines if retry should continue based on context
func ShouldRetry(ctx context.Context, attempt int, maxAttempts int) bool {
	if attempt >= maxAttempts {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}
