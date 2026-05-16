package totp

import "time"

// Clock provides time for testability.
type Clock interface {
	Now() time.Time
}

// RealClock provides actual time.
type RealClock struct{}

// NewRealClock creates a new real clock.
func NewRealClock() Clock {
	return &RealClock{}
}

// Now returns the current time.
func (c *RealClock) Now() time.Time {
	return time.Now().UTC()
}

// MockClock provides fixed time for testing.
type MockClock struct {
	Time time.Time
}

// Now returns the mock time.
func (c *MockClock) Now() time.Time {
	return c.Time
}
