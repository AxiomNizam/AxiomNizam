// Package stream implements a Nomad-style event broker: many
// publishers, many subscribers, topic-scoped filtering, and
// resumable subscriptions via a monotonic index.
//
// The shape matches nomad/nomad/stream: Broker.Publish appends events
// to an internal ring; Subscribe returns a Subscription whose Next()
// yields events strictly newer than the subscriber's last-seen
// index.  The ring trims to MaxSize so a slow subscriber cannot
// pin memory indefinitely — slow readers lose the oldest events and
// see an explicit ErrSubscriptionClosed on the next Next.
package stream

import (
	"context"
	"errors"
	"sync"
)

// Event is a single broker payload.  Topic and Key are used for
// subscriber filtering; Index is the monotonic sequence number the
// broker assigns on Publish.
type Event struct {
	// Topic is the coarse channel — e.g. "Job", "Deployment".
	Topic string
	// Key is the fine-grained identifier, typically "namespace/name".
	Key string
	// Index is the broker-assigned sequence — subscribers resume here.
	Index uint64
	// Payload is the opaque body; downstream consumers type-assert.
	Payload interface{}
}

// Broker is the publisher/dispatcher.
type Broker struct {
	// MaxSize bounds the in-memory ring.  Zero disables trimming.
	MaxSize int

	mu   sync.Mutex
	cond *sync.Cond
	ring []Event
	// nextIndex is the sequence number to assign on the next Publish.
	// Starts at 1 so subscribers can use 0 as "from the beginning".
	nextIndex uint64
	// firstIndex is the index of ring[0] — needed to answer "is the
	// subscriber's resume point still in memory?".
	firstIndex uint64
	closed     bool
}

// NewBroker returns an empty broker bounded to maxSize events.
func NewBroker(maxSize int) *Broker {
	b := &Broker{MaxSize: maxSize, nextIndex: 1, firstIndex: 1}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Publish appends evt to the ring and wakes every subscriber.  The
// caller-supplied Index field is overwritten with the assigned value.
func (b *Broker) Publish(evt Event) uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return 0
	}
	evt.Index = b.nextIndex
	b.nextIndex++
	b.ring = append(b.ring, evt)
	if b.MaxSize > 0 && len(b.ring) > b.MaxSize {
		drop := len(b.ring) - b.MaxSize
		b.ring = b.ring[drop:]
		b.firstIndex += uint64(drop)
	}
	b.cond.Broadcast()
	return evt.Index
}

// Close stops all subscribers.  Idempotent.
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	b.cond.Broadcast()
}

// LastIndex returns the index of the most recent published event,
// or 0 if nothing has been published.  Used by callers that want to
// start a new subscription "from now".
func (b *Broker) LastIndex() uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.nextIndex - 1
}

// ErrSubscriptionClosed is returned by Next after Close or after the
// ring has moved past the subscriber's resume point.
var ErrSubscriptionClosed = errors.New("stream: subscription closed or lagging")

// SubscribeReq carries filter parameters.  An empty Topics list
// means "every topic"; an empty Keys list means "every key within
// the matched topics".
type SubscribeReq struct {
	Topics []string
	Keys   []string
	// StartIndex is the first index the subscriber wants.  Zero means
	// "start from the broker's current head".
	StartIndex uint64
}

// matches tests whether evt passes the request's filters.
func (r *SubscribeReq) matches(evt Event) bool {
	if len(r.Topics) > 0 && !contains(r.Topics, evt.Topic) {
		return false
	}
	if len(r.Keys) > 0 && !contains(r.Keys, evt.Key) {
		return false
	}
	return true
}

// contains is a tiny case-sensitive membership test.
func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

// Subscribe returns a Subscription honouring req.
func (b *Broker) Subscribe(req SubscribeReq) *Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()
	startIdx := req.StartIndex
	if startIdx == 0 {
		startIdx = b.nextIndex
	}
	return &Subscription{broker: b, req: req, nextIdx: startIdx}
}

// Subscription is a long-lived pull cursor.
type Subscription struct {
	broker  *Broker
	req     SubscribeReq
	nextIdx uint64
	done    bool
}

// Next blocks until an event matching the subscription is available,
// ctx is cancelled, or the subscription has been closed.  A single
// Next call may skip many non-matching events to find the next match.
func (s *Subscription) Next(ctx context.Context) (Event, error) {
	if s.done {
		return Event{}, ErrSubscriptionClosed
	}
	// Park a goroutine to wake the cond on ctx cancellation so Next
	// can return when the caller gives up.
	cancelWake := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			s.broker.mu.Lock()
			s.broker.cond.Broadcast()
			s.broker.mu.Unlock()
		case <-cancelWake:
		}
	}()
	defer close(cancelWake)

	s.broker.mu.Lock()
	defer s.broker.mu.Unlock()
	for {
		if s.broker.closed {
			s.done = true
			return Event{}, ErrSubscriptionClosed
		}
		if ctx.Err() != nil {
			return Event{}, ctx.Err()
		}
		// If the subscriber's resume point has fallen off the ring,
		// fail fast — forcing clients to reset rather than silently
		// skipping events.
		if s.nextIdx < s.broker.firstIndex {
			s.done = true
			return Event{}, ErrSubscriptionClosed
		}
		// Find the first matching event at or after nextIdx.
		for i := range s.broker.ring {
			evt := s.broker.ring[i]
			if evt.Index < s.nextIdx {
				continue
			}
			s.nextIdx = evt.Index + 1
			if s.req.matches(evt) {
				return evt, nil
			}
		}
		// Nothing matched; wait for a new publish or close/cancel.
		s.broker.cond.Wait()
	}
}

// Close releases the subscription.  Safe to call from any goroutine.
func (s *Subscription) Close() {
	s.broker.mu.Lock()
	defer s.broker.mu.Unlock()
	s.done = true
}
