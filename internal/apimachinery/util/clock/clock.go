// Package clock provides an abstraction over time.Now and time.After
// so that code which schedules future work can be driven
// deterministically by tests.  The production implementation is a
// thin passthrough to the time package; the FakeClock used in tests
// advances only when the test explicitly calls Step.
//
// Pattern lifted from k8s.io/utils/clock.  The interface is kept
// minimal — only the calls AxiomNizam schedulers actually need.
package clock

import "time"

// Clock is the minimal abstraction.  Callers that want to observe
// the current wall-clock time must accept a Clock rather than call
// time.Now directly.
type Clock interface {
	// Now returns the current wall-clock moment.
	Now() time.Time
	// Since returns Now()-t.
	Since(t time.Time) time.Duration
	// After fires a single value on the returned channel after d.
	After(d time.Duration) <-chan time.Time
	// NewTimer returns a timer equivalent to time.NewTimer.
	NewTimer(d time.Duration) Timer
	// Sleep blocks for d, honouring cancellation only if the caller
	// wraps this call in a select (matching time.Sleep).
	Sleep(d time.Duration)
}

// Timer abstracts *time.Timer for the fake-clock variant.
type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(d time.Duration) bool
}

// RealClock is the production implementation.
type RealClock struct{}

// Now returns time.Now.
func (RealClock) Now() time.Time { return time.Now() }

// Since returns time.Since.
func (RealClock) Since(t time.Time) time.Duration { return time.Since(t) }

// After returns time.After.
func (RealClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// NewTimer wraps time.NewTimer.
func (RealClock) NewTimer(d time.Duration) Timer { return &realTimer{t: time.NewTimer(d)} }

// Sleep wraps time.Sleep.
func (RealClock) Sleep(d time.Duration) { time.Sleep(d) }

// realTimer adapts *time.Timer to the Timer interface.
type realTimer struct{ t *time.Timer }

// C exposes the timer's channel.
func (r *realTimer) C() <-chan time.Time { return r.t.C }

// Stop proxies to time.Timer.Stop.
func (r *realTimer) Stop() bool { return r.t.Stop() }

// Reset proxies to time.Timer.Reset.
func (r *realTimer) Reset(d time.Duration) bool { return r.t.Reset(d) }
