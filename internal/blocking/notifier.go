// Package blocking implements the long-polling "blocking query"
// pattern that Nomad and Consul use for their list APIs: the client
// submits its last-known index; the server blocks until either a
// newer index is available or the timeout elapses; the response
// carries the new index the client uses on its next call.
//
// This package provides the server-side watcher: Notifier.NotifyAll
// is called by mutators after every change; Notifier.Wait returns
// when the observed index advances past the caller's threshold or
// ctx fires.
package blocking

import (
	"context"
	"sync"
	"time"
)

// Notifier is the synchronisation primitive shared between writers
// (who call NotifyAll after updating the store) and readers (who
// call Wait to long-poll).
type Notifier struct {
	mu    sync.Mutex
	cond  *sync.Cond
	index uint64
}

// NewNotifier returns a notifier starting at index 1 — the same
// convention the stream broker uses so callers can share indexes.
func NewNotifier() *Notifier {
	n := &Notifier{index: 1}
	n.cond = sync.NewCond(&n.mu)
	return n
}

// Index returns the current index without blocking.
func (n *Notifier) Index() uint64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.index
}

// NotifyAll advances the index and wakes every blocked Wait.  The
// new index is returned so the caller can stamp its just-written
// object.
func (n *Notifier) NotifyAll() uint64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.index++
	n.cond.Broadcast()
	return n.index
}

// Wait blocks until the notifier's index is strictly greater than
// minIndex, ctx is cancelled, or maxWait elapses.  The returned
// index is the notifier's value at unblock time, guaranteed to be
// >= minIndex+1 on the happy path and == minIndex (i.e. unchanged)
// on timeout/cancel.
//
// minIndex of 0 is treated as "I have no prior index" and returns
// immediately with the current value.
func (n *Notifier) Wait(ctx context.Context, minIndex uint64, maxWait time.Duration) uint64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	if minIndex == 0 || n.index > minIndex {
		return n.index
	}

	// Use a timer so long-poll callers don't pin a goroutine forever.
	var timer *time.Timer
	var timerFired bool
	if maxWait > 0 {
		timer = time.AfterFunc(maxWait, func() {
			n.mu.Lock()
			timerFired = true
			n.cond.Broadcast()
			n.mu.Unlock()
		})
		defer timer.Stop()
	}

	// Park a helper goroutine so ctx cancellation unblocks the cond.
	stop := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			n.mu.Lock()
			n.cond.Broadcast()
			n.mu.Unlock()
		case <-stop:
		}
	}()
	defer close(stop)

	for n.index <= minIndex && ctx.Err() == nil && !timerFired {
		n.cond.Wait()
	}
	return n.index
}

// BlockingQuery is the upstream-Consul-style convenience wrapper.
// The caller supplies a query function that returns (result, index)
// for a given snapshot.  BlockingQuery re-runs the function whenever
// the notifier advances until the result's index exceeds minIndex.
// Useful for list endpoints that want to hide the Wait plumbing.
func BlockingQuery[T any](
	ctx context.Context,
	notifier *Notifier,
	minIndex uint64,
	maxWait time.Duration,
	query func() (T, uint64, error),
) (T, uint64, error) {
	deadline := time.Now().Add(maxWait)
	for {
		result, idx, err := query()
		if err != nil {
			return result, idx, err
		}
		if idx > minIndex {
			return result, idx, nil
		}
		remaining := time.Until(deadline)
		if maxWait > 0 && remaining <= 0 {
			return result, idx, nil
		}
		if maxWait <= 0 {
			remaining = 0
		}
		newIdx := notifier.Wait(ctx, idx, remaining)
		if newIdx == idx {
			// Timeout or cancel — return the stale result so the
			// client can refresh its cursor without an error.
			return result, idx, ctx.Err()
		}
	}
}
