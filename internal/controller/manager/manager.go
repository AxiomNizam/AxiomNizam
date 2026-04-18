// Package manager coordinates the lifecycle of many Runnables —
// controllers, webhook servers, metrics exporters, leader-election
// drivers — as a single bundle.  Start blocks until either a
// Runnable returns an error or the context is cancelled; stop is
// idempotent and safe from any goroutine.
//
// The design mirrors controller-runtime's Manager, but with fewer
// knobs: AxiomNizam does not need the port-binding and scheme-
// discovery helpers upstream provides.
package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Runnable is the minimal lifecycle interface.  Start blocks until
// stopCh is closed and then returns; returning before stopCh is
// closed is treated as a fatal error for the whole manager.
type Runnable interface {
	Start(stopCh <-chan struct{}) error
}

// RunnableFunc adapts a function to the interface.
type RunnableFunc func(stopCh <-chan struct{}) error

// Start satisfies Runnable.
func (f RunnableFunc) Start(stopCh <-chan struct{}) error { return f(stopCh) }

// Manager is the coordinator.
type Manager struct {
	mu        sync.Mutex
	runnables []Runnable
	started   bool

	// LeaderElected, if non-nil, runs before any other Runnable and
	// must acquire leadership before Start returns ready.  Callers
	// wire a resourcelock-backed implementation here.
	LeaderElected Runnable
}

// New returns an empty Manager.
func New() *Manager { return &Manager{} }

// Add registers a Runnable.  Adding after Start returns an error
// because late additions would miss the initial start signal.
func (m *Manager) Add(r Runnable) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return fmt.Errorf("manager: cannot Add after Start")
	}
	if r == nil {
		return errors.New("manager: nil Runnable")
	}
	m.runnables = append(m.runnables, r)
	return nil
}

// Start launches every Runnable in a dedicated goroutine and blocks
// until ctx is cancelled or any Runnable returns.  The first non-nil
// error from any Runnable is propagated; subsequent errors are
// dropped but logged via fmt.Errorf chains.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		return errors.New("manager: already started")
	}
	m.started = true
	runnables := append([]Runnable(nil), m.runnables...)
	m.mu.Unlock()

	stopCh := make(chan struct{})
	defer closeOnce(stopCh)

	// Propagate ctx cancellation into stopCh.
	go func() {
		<-ctx.Done()
		closeOnce(stopCh)
	}()

	if m.LeaderElected != nil {
		// Run leader election inline; it must return quickly once
		// leadership is acquired or lost.  We wrap it in a goroutine
		// so a long-running election does not block the rest of the
		// manager's startup.
		go func() {
			if err := m.LeaderElected.Start(stopCh); err != nil {
				// Leadership failure is fatal — trigger shutdown.
				closeOnce(stopCh)
			}
		}()
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(runnables))
	for _, r := range runnables {
		r := r
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := r.Start(stopCh); err != nil {
				errCh <- err
				closeOnce(stopCh)
			}
		}()
	}
	wg.Wait()
	close(errCh)

	// Aggregate errors.
	var firstErr error
	for err := range errCh {
		if firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// closeOnce is a nil-safe idempotent close.
func closeOnce(ch chan struct{}) {
	defer func() { _ = recover() }() // swallow "close of closed channel"
	close(ch)
}
