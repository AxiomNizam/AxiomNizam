package waitx

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Checker describes a wait condition.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// RetryStrategy controls delay calculation between attempts.
type RetryStrategy interface {
	NextDelay(attempt int, baseInterval, maxInterval time.Duration, lastErr error) time.Duration
}

// AttemptEvent includes metadata for each retry callback.
type AttemptEvent struct {
	Checker string
	Attempt int
	Delay   time.Duration
	Elapsed time.Duration
	Err     error
}

// WaitOptions configures wait behavior.
type WaitOptions struct {
	Timeout       time.Duration
	Interval      time.Duration
	MaxInterval   time.Duration
	InvertCheck   bool
	RetryStrategy RetryStrategy
	OnRetry       func(event AttemptEvent)
}

// DefaultWaitOptions returns a safe baseline config.
func DefaultWaitOptions() WaitOptions {
	return WaitOptions{
		Timeout:       30 * time.Second,
		Interval:      time.Second,
		MaxInterval:   30 * time.Second,
		InvertCheck:   false,
		RetryStrategy: LinearRetry{},
	}
}

// WaitContext waits until checker returns ready, or context/timeout cancels.
func WaitContext(ctx context.Context, checker Checker, opts WaitOptions) error {
	if checker == nil {
		return errors.New("checker is required")
	}
	applyDefaultWaitOptions(&opts)
	ctx, cancel := withOptionalTimeout(ctx, opts.Timeout)
	defer cancel()

	checkerName := checker.Name()
	started := time.Now()

	for attempt := 1; ; attempt++ {
		checkErr := checker.Check(ctx)
		ready, effectiveErr := resolveCheckResult(checkErr, opts.InvertCheck)
		if ready {
			return nil
		}

		if abortErr := contextAbortError(ctx, checkerName, started, effectiveErr); abortErr != nil {
			return abortErr
		}

		delay := clampNonNegativeDuration(opts.RetryStrategy.NextDelay(attempt, opts.Interval, opts.MaxInterval, effectiveErr))
		notifyRetry(opts.OnRetry, checkerName, attempt, delay, started, effectiveErr)

		if waitErr := waitDelayOrAbort(ctx, delay, checkerName, started, effectiveErr); waitErr != nil {
			return waitErr
		}
	}
}

func applyDefaultWaitOptions(opts *WaitOptions) {
	if opts.Interval <= 0 {
		opts.Interval = time.Second
	}
	if opts.MaxInterval <= 0 {
		opts.MaxInterval = 30 * time.Second
	}
	if opts.RetryStrategy == nil {
		opts.RetryStrategy = LinearRetry{}
	}
}

func withOptionalTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return context.WithCancel(ctx)
}

func resolveCheckResult(checkErr error, invertCheck bool) (bool, error) {
	if !invertCheck {
		return checkErr == nil, checkErr
	}
	if checkErr != nil {
		return true, checkErr
	}
	return false, errors.New("condition is still ready")
}

func contextAbortError(ctx context.Context, checkerName string, started time.Time, lastErr error) error {
	if ctx.Err() == nil {
		return nil
	}
	return formatWaitExitError(ctx.Err(), checkerName, time.Since(started), lastErr)
}

func clampNonNegativeDuration(delay time.Duration) time.Duration {
	if delay < 0 {
		return 0
	}
	return delay
}

func notifyRetry(onRetry func(event AttemptEvent), checkerName string, attempt int, delay time.Duration, started time.Time, err error) {
	if onRetry == nil {
		return
	}
	onRetry(AttemptEvent{
		Checker: checkerName,
		Attempt: attempt,
		Delay:   delay,
		Elapsed: time.Since(started),
		Err:     err,
	})
}

func waitDelayOrAbort(ctx context.Context, delay time.Duration, checkerName string, started time.Time, lastErr error) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return formatWaitExitError(ctx.Err(), checkerName, time.Since(started), lastErr)
	case <-timer.C:
		return nil
	}
}

func formatWaitExitError(ctxErr error, checkerName string, elapsed time.Duration, lastErr error) error {
	rounded := elapsed.Round(time.Millisecond)
	if errors.Is(ctxErr, context.DeadlineExceeded) {
		return fmt.Errorf("wait for %s timed out after %s: %w", checkerName, rounded, lastErr)
	}
	return fmt.Errorf("wait for %s aborted after %s: %w", checkerName, rounded, ctxErr)
}
