package retry

import (
	"context"
	"fmt"
	"time"
)

const errMaxAttemptsExceeded = "max attempts (%d) exceeded"

// Config defines retry configuration
type Config struct {
	MaxAttempts       int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
	Jitter            bool
}

// DefaultConfig returns default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxAttempts:       3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
	}
}

// Result represents the result of a retry attempt
type Result struct {
	Attempt   int
	Duration  time.Duration
	Error     error
	Success   bool
	LastError error
}

// Func defines the retry function signature
type Func func(ctx context.Context, attempt int) error

// Do executes a function with retry logic
func Do(ctx context.Context, config Config, fn Func) Result {
	result := Result{
		LastError: fmt.Errorf(errMaxAttemptsExceeded, config.MaxAttempts),
	}

	backoff := config.InitialBackoff

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		result.Attempt = attempt

		// Execute the function
		err := fn(ctx, attempt)
		if err == nil {
			result.Success = true
			result.Error = nil
			return result
		}

		result.Error = err
		result.LastError = err

		// If this was the last attempt, don't backoff
		if attempt == config.MaxAttempts {
			break
		}

		// Apply backoff with optional jitter
		waitTime := backoff
		if config.Jitter {
			waitTime = applyJitter(backoff)
		}

		result.Duration = waitTime

		// Wait before retrying
		select {
		case <-time.After(waitTime):
			// Continue to next attempt
		case <-ctx.Done():
			result.Error = ctx.Err()
			return result
		}

		// Increase backoff for next attempt
		backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}
	}

	return result
}

// DoAsync executes a function with retry logic asynchronously
func DoAsync(ctx context.Context, config Config, fn Func) <-chan Result {
	resultChan := make(chan Result, 1)

	go func() {
		defer close(resultChan)
		result := Do(ctx, config, fn)
		select {
		case resultChan <- result:
		case <-ctx.Done():
		}
	}()

	return resultChan
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Add custom retryable error checking here
	// For now, all errors are potentially retryable
	return true
}

// Once executes a function once (no retries)
func Once(ctx context.Context, fn Func) error {
	return fn(ctx, 1)
}

// WithLimit returns a retry function with max attempts limit
func WithLimit(fn Func, maxAttempts int) Func {
	return func(ctx context.Context, attempt int) error {
		if attempt > maxAttempts {
			return fmt.Errorf(errMaxAttemptsExceeded, maxAttempts)
		}
		return fn(ctx, attempt)
	}
}

// WithTimeout returns a retry function with timeout
func WithTimeout(fn Func, timeout time.Duration) Func {
	return func(ctx context.Context, attempt int) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return fn(ctx, attempt)
	}
}

// WithFilter returns a retry function that filters retryable errors
func WithFilter(fn Func, isRetryable func(error) bool) Func {
	return func(ctx context.Context, attempt int) error {
		err := fn(ctx, attempt)
		if err != nil && !isRetryable(err) {
			return err // Don't retry non-retryable errors
		}
		return err
	}
}

func applyJitter(backoff time.Duration) time.Duration {
	// Add jitter to prevent thundering herd
	jitterAmount := time.Duration(float64(backoff) * 0.1) // 10% jitter
	return backoff + jitterAmount
}

// Backoff strategy interface
type BackoffStrategy interface {
	NextBackoff(attempt int) time.Duration
}

// ExponentialBackoff implements exponential backoff strategy
type ExponentialBackoff struct {
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
}

// NextBackoff returns next backoff duration
func (eb *ExponentialBackoff) NextBackoff(attempt int) time.Duration {
	backoff := eb.InitialBackoff
	for i := 1; i < attempt; i++ {
		backoff = time.Duration(float64(backoff) * eb.BackoffMultiplier)
		if backoff > eb.MaxBackoff {
			return eb.MaxBackoff
		}
	}
	return backoff
}

// LinearBackoff implements linear backoff strategy
type LinearBackoff struct {
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Step           time.Duration
}

// NextBackoff returns next backoff duration
func (lb *LinearBackoff) NextBackoff(attempt int) time.Duration {
	backoff := lb.InitialBackoff + lb.Step*time.Duration(attempt-1)
	if backoff > lb.MaxBackoff {
		return lb.MaxBackoff
	}
	return backoff
}

// ConstantBackoff implements constant backoff strategy
type ConstantBackoff struct {
	Backoff time.Duration
}

// NextBackoff returns next backoff duration
func (cb *ConstantBackoff) NextBackoff(attempt int) time.Duration {
	return cb.Backoff
}

// DoWithStrategy executes function with custom backoff strategy
func DoWithStrategy(ctx context.Context, maxAttempts int, strategy BackoffStrategy, fn Func) Result {
	result := Result{
		LastError: fmt.Errorf(errMaxAttemptsExceeded, maxAttempts),
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result.Attempt = attempt

		// Execute the function
		err := fn(ctx, attempt)
		if err == nil {
			result.Success = true
			result.Error = nil
			return result
		}

		result.Error = err
		result.LastError = err

		// If this was the last attempt, don't backoff
		if attempt == maxAttempts {
			break
		}

		// Get next backoff
		waitTime := strategy.NextBackoff(attempt)
		result.Duration = waitTime

		// Wait before retrying
		select {
		case <-time.After(waitTime):
			// Continue to next attempt
		case <-ctx.Done():
			result.Error = ctx.Err()
			return result
		}
	}

	return result
}
