// Package clock — FakeClock for deterministic tests.
//
// Usage in a test:
//
//	c := clock.NewFake(time.Unix(0, 0))
//	go worker(c)          // worker calls c.NewTimer(1 * time.Second)
//	c.Step(1 * time.Second)
//	// worker's timer has now fired.
//
// All timers registered before a Step call that crosses their
// deadline fire in registration order.  Thread-safe for concurrent
// test goroutines.
package clock

import (
	"sort"
	"sync"
	"time"
)

// FakeClock implements Clock with manual time advancement.
type FakeClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*fakeTimer
}

// NewFake returns a FakeClock starting at start.
func NewFake(start time.Time) *FakeClock { return &FakeClock{now: start} }

// Now returns the current fake time.
func (f *FakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

// Since returns f.Now() - t.
func (f *FakeClock) Since(t time.Time) time.Duration {
	return f.Now().Sub(t)
}

// After returns a channel that fires the next time Step advances past
// the current time + d.
func (f *FakeClock) After(d time.Duration) <-chan time.Time {
	t := f.NewTimer(d)
	return t.C()
}

// NewTimer registers a fake timer.
func (f *FakeClock) NewTimer(d time.Duration) Timer {
	f.mu.Lock()
	defer f.mu.Unlock()
	ft := &fakeTimer{
		clock:    f,
		deadline: f.now.Add(d),
		ch:       make(chan time.Time, 1),
		active:   true,
	}
	f.timers = append(f.timers, ft)
	return ft
}

// Sleep on a FakeClock blocks until Step advances past d.  Matches
// time.Sleep semantics — goroutines calling Sleep on a FakeClock will
// remain parked until the test drives the clock forward.
func (f *FakeClock) Sleep(d time.Duration) {
	<-f.After(d)
}

// Step advances the fake clock by d and fires every timer whose
// deadline now lies at or before the new `now`.
func (f *FakeClock) Step(d time.Duration) {
	f.mu.Lock()
	f.now = f.now.Add(d)
	nowCopy := f.now
	// Capture timers that are due; release the lock before firing to
	// avoid deadlocks with callers that take another lock in their
	// handler.
	var due []*fakeTimer
	var remaining []*fakeTimer
	// Fire in deadline order so consumers can reason about ordering.
	sort.Slice(f.timers, func(i, j int) bool {
		return f.timers[i].deadline.Before(f.timers[j].deadline)
	})
	for _, t := range f.timers {
		if t.active && !t.deadline.After(nowCopy) {
			due = append(due, t)
		} else {
			remaining = append(remaining, t)
		}
	}
	f.timers = remaining
	f.mu.Unlock()

	for _, t := range due {
		// Non-blocking send — tests that never read the channel
		// should not hang the whole Step call.
		select {
		case t.ch <- nowCopy:
		default:
		}
		t.markInactive()
	}
}

// SetTime absolutely sets the fake clock; fires every timer whose
// deadline lies at or before the new time.  Equivalent to
// Step(target.Sub(Now())) when target >= Now(); rewinds without
// firing anything when target < Now().
func (f *FakeClock) SetTime(target time.Time) {
	f.mu.Lock()
	if target.Before(f.now) {
		f.now = target
		f.mu.Unlock()
		return
	}
	delta := target.Sub(f.now)
	f.mu.Unlock()
	f.Step(delta)
}

// fakeTimer is the per-timer state.
type fakeTimer struct {
	mu       sync.Mutex
	clock    *FakeClock
	deadline time.Time
	ch       chan time.Time
	active   bool
}

// C returns the fire channel.
func (t *fakeTimer) C() <-chan time.Time { return t.ch }

// Stop deactivates the timer.  Returns true when the timer was active.
func (t *fakeTimer) Stop() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	wasActive := t.active
	t.active = false
	return wasActive
}

// Reset reschedules the timer to fire after d from the current fake now.
func (t *fakeTimer) Reset(d time.Duration) bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	t.mu.Lock()
	defer t.mu.Unlock()
	wasActive := t.active
	t.deadline = t.clock.now.Add(d)
	t.active = true
	// Ensure we are re-registered if Stop had removed us; search by pointer.
	found := false
	for _, existing := range t.clock.timers {
		if existing == t {
			found = true
			break
		}
	}
	if !found {
		t.clock.timers = append(t.clock.timers, t)
	}
	return wasActive
}

// markInactive flips the state under the timer's own lock.
func (t *fakeTimer) markInactive() {
	t.mu.Lock()
	t.active = false
	t.mu.Unlock()
}
