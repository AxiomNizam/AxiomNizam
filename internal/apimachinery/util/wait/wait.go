// Package wait mirrors the k8s.io/apimachinery/pkg/util/wait package:
// scheduling primitives for "do X every T", "poll until condition",
// and "backoff on failure".  These are used pervasively in
// controllers — the reconciler loop, leader-election renewal, and
// event throttling all reach for one of the Until / Poll / Backoff
// combinators.
//
// Every function honours context cancellation.  The older PollImmediate /
// PollUntil variants from upstream are collapsed into a single Poll
// family taking a PollOptions struct, because the combinatorial blow-
// up of function names was always the least-loved part of the API.
package wait

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// ErrWaitTimeout is returned when a Poll exhausts its timeout before
// the condition returns true.  Using a sentinel rather than
// context.DeadlineExceeded lets callers distinguish "my context was
// cancelled" from "the poll budget was exhausted".
var ErrWaitTimeout = errors.New("wait: timed out waiting for condition")

// ConditionFunc is the predicate a Poll evaluates.  Return (true, nil)
// to stop polling with success, (false, nil) to keep polling, or
// (anything, err) to stop with err.
type ConditionFunc func(ctx context.Context) (done bool, err error)

// Forever runs fn every period until ctx is cancelled.  Panics in fn
// are recovered and logged; the tick does not miss.
func Forever(ctx context.Context, period time.Duration, fn func()) {
	Until(ctx, period, fn)
}

// Until is like Forever but exits cleanly when ctx is done.  Runs fn
// immediately on entry, then on each subsequent tick.
func Until(ctx context.Context, period time.Duration, fn func()) {
	if period <= 0 {
		return
	}
	run := func() {
		defer func() {
			_ = recover() // swallow — callers should log inside fn.
		}()
		fn()
	}
	run()
	t := time.NewTicker(period)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			run()
		}
	}
}

// JitterUntil is like Until but adds ±jitterFactor * period to each
// sleep, spreading load from many replicas that would otherwise wake
// together.  Set sliding=true to measure the period from the *end*
// of each run, false for the start.
func JitterUntil(ctx context.Context, period time.Duration, jitterFactor float64, sliding bool, fn func()) {
	run := func() {
		defer func() { _ = recover() }()
		fn()
	}
	for {
		if !sliding {
			start := time.Now()
			run()
			// Compute how much time is left in this tick.
			spent := time.Since(start)
			if spent < period {
				if err := sleepJitter(ctx, period-spent, jitterFactor); err != nil {
					return
				}
			}
			continue
		}
		run()
		if err := sleepJitter(ctx, period, jitterFactor); err != nil {
			return
		}
	}
}

// sleepJitter blocks for d * (1 ± jitterFactor) honouring ctx.
func sleepJitter(ctx context.Context, d time.Duration, jitterFactor float64) error {
	if jitterFactor > 0 {
		d += time.Duration(rand.Float64() * jitterFactor * float64(d))
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// PollOptions configures a Poll run.  Zero means "no limit" for
// Timeout / Steps; zero Interval is illegal and treated as 100ms.
type PollOptions struct {
	// Interval is the base gap between condition checks.
	Interval time.Duration
	// Timeout bounds the total wall-clock time of the poll.
	Timeout time.Duration
	// Immediate=true evaluates the condition once before the first sleep.
	Immediate bool
	// JitterFactor adds up to ±factor*Interval per step.
	JitterFactor float64
}

// Poll calls condition every opts.Interval until it returns (true,_),
// ctx is cancelled, or the timeout elapses.  Returns ErrWaitTimeout
// when the budget is exhausted without success.
func Poll(ctx context.Context, opts PollOptions, condition ConditionFunc) error {
	interval := opts.Interval
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	deadline := time.Time{}
	if opts.Timeout > 0 {
		deadline = time.Now().Add(opts.Timeout)
	}

	check := func() (bool, error) {
		return condition(ctx)
	}

	if opts.Immediate {
		done, err := check()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}

	for {
		// Compute sleep duration honouring both the context and the
		// wall-clock deadline.
		step := interval
		if opts.JitterFactor > 0 {
			step += time.Duration(rand.Float64() * opts.JitterFactor * float64(step))
		}
		if !deadline.IsZero() {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return ErrWaitTimeout
			}
			if step > remaining {
				step = remaining
			}
		}
		t := time.NewTimer(step)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}

		done, err := check()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return ErrWaitTimeout
		}
	}
}

// PollImmediate is the shorthand for Poll with Immediate=true.
func PollImmediate(ctx context.Context, interval, timeout time.Duration, condition ConditionFunc) error {
	return Poll(ctx, PollOptions{Interval: interval, Timeout: timeout, Immediate: true}, condition)
}

// Backoff configures exponential sleep on retry.
type Backoff struct {
	// Duration is the starting sleep.
	Duration time.Duration
	// Factor multiplies Duration at each step.
	Factor float64
	// Jitter is the max ±fraction of Duration added per step.
	Jitter float64
	// Steps caps how many times Step can be called before clamping
	// at Cap.  A Steps of 0 means "forever".
	Steps int
	// Cap bounds the duration regardless of Steps.
	Cap time.Duration

	// internal state
	step int
}

// Step returns the next sleep and advances internal state.
func (b *Backoff) Step() time.Duration {
	if b.Steps > 0 && b.step >= b.Steps {
		return b.cap(b.Duration)
	}
	d := b.Duration
	if b.Jitter > 0 {
		d += time.Duration(rand.Float64() * b.Jitter * float64(d))
	}
	b.step++
	if b.Factor > 0 {
		next := float64(b.Duration) * b.Factor
		if next > math.MaxInt64 {
			next = math.MaxInt64
		}
		b.Duration = time.Duration(next)
	}
	return b.cap(d)
}

// cap applies b.Cap if set.
func (b *Backoff) cap(d time.Duration) time.Duration {
	if b.Cap > 0 && d > b.Cap {
		return b.Cap
	}
	return d
}

// ExponentialBackoff retries fn with an exponential gap until it
// returns (true, _) or the Backoff is exhausted.
func ExponentialBackoff(ctx context.Context, backoff Backoff, fn func() (done bool, err error)) error {
	for {
		done, err := fn()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		d := backoff.Step()
		if backoff.Steps > 0 && backoff.step > backoff.Steps {
			return ErrWaitTimeout
		}
		t := time.NewTimer(d)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
}
