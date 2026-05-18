package testutil

import "time"

// FakeClock is a controllable clock for testing.
type FakeClock struct {
	now time.Time
}

// NewFakeClock creates a FakeClock starting at the given time.
func NewFakeClock(start time.Time) *FakeClock {
	return &FakeClock{now: start}
}

// Now returns the current fake time.
func (c *FakeClock) Now() time.Time {
	return c.now
}

// Advance moves the clock forward by the given duration.
func (c *FakeClock) Advance(d time.Duration) {
	c.now = c.now.Add(d)
}

// Set sets the clock to a specific time.
func (c *FakeClock) Set(t time.Time) {
	c.now = t
}
