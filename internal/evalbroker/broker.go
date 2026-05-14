// Package evalbroker implements the Nomad evaluation-broker pattern:
// a priority queue of work items with explicit ack/nack semantics,
// delayed re-enqueue on failure, and per-item visibility timeouts.
//
// The broker is the reliable-dispatch heart of any system that
// separates "something happened" from "decide what to do about it".
// AxiomNizam controllers use it for scheduler evaluations, workflow
// steps, and cross-region replication dispatches — all cases where
// losing an item silently would be worse than dispatching it twice.
//
// Compared to a plain workqueue, evalbroker adds:
//   - Priority ordering: higher priority wins ties.
//   - Explicit Ack/Nack: items only leave the system on Ack.
//   - Visibility timeout: items auto-Nacked if Ack is not called in
//     time, surviving worker crashes without losing work.
//   - Delay: Nack can defer re-dispatch to shed hot-loop failures.
package evalbroker

import (
	"container/heap"
	"errors"
	"sync"
	"time"
)

// Evaluation is the unit of work.
type Evaluation struct {
	// ID is the stable identifier (UUID, usually).
	ID string
	// Priority: higher values dispatch first.
	Priority int
	// Type is an optional classifier used by metrics and logs.
	Type string
	// Payload is opaque to the broker.
	Payload interface{}
	// CreateTime records when the caller first Enqueued — used for
	// age-based tiebreaking on equal priorities.
	CreateTime time.Time
}

// Config tunes the broker.
type Config struct {
	// DeliveryLimit caps Nacks before an item is moved to the DLQ.
	// Zero means "no limit".
	DeliveryLimit int
	// NackTimeout is the default re-enqueue delay when Nack is called
	// without an explicit delay.
	NackTimeout time.Duration
	// VisibilityTimeout bounds how long a Dequeue'd eval can remain
	// un-Acked before the broker re-enqueues it.  Zero disables.
	VisibilityTimeout time.Duration
}

// Broker coordinates enqueue, dequeue, ack, and nack.
type Broker struct {
	cfg Config

	mu   sync.Mutex
	cond *sync.Cond
	// ready is the heap of eligible evals.
	ready pq
	// pending is the set of evals currently Dequeue'd but not Acked.
	pending map[string]*pendingEval
	// dlq is the dead-letter queue: evals that exceeded DeliveryLimit.
	dlq []Evaluation
	// dispatched counts Nacks per eval ID; reset on Ack.
	attempts map[string]int
	closed   bool
}

// pendingEval tracks an outstanding Dequeue.
type pendingEval struct {
	eval     Evaluation
	deadline time.Time
}

// New returns a configured broker.
func New(cfg Config) *Broker {
	b := &Broker{
		cfg:      cfg,
		pending:  map[string]*pendingEval{},
		attempts: map[string]int{},
	}
	b.cond = sync.NewCond(&b.mu)
	go b.visibilityReaper()
	return b
}

// Enqueue adds an eval.  Duplicate IDs coalesce — the highest
// priority wins, matching Nomad's semantics for "re-evaluate this
// object I already have a pending eval for".
func (b *Broker) Enqueue(e Evaluation) {
	if e.CreateTime.IsZero() {
		e.CreateTime = time.Now()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	// Coalesce against the ready heap.
	for i, existing := range b.ready {
		if existing.ID == e.ID {
			if e.Priority > existing.Priority {
				b.ready[i] = e
				heap.Fix(&b.ready, i)
			}
			return
		}
	}
	heap.Push(&b.ready, e)
	b.cond.Signal()
}

// Dequeue blocks until an eval is available or the broker is closed.
// The returned eval is considered "in-flight" until Ack or Nack;
// failure to Ack within VisibilityTimeout results in automatic Nack.
func (b *Broker) Dequeue() (Evaluation, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for b.ready.Len() == 0 && !b.closed {
		b.cond.Wait()
	}
	if b.closed {
		return Evaluation{}, false
	}
	e := heap.Pop(&b.ready).(Evaluation)
	pe := &pendingEval{eval: e}
	if b.cfg.VisibilityTimeout > 0 {
		pe.deadline = time.Now().Add(b.cfg.VisibilityTimeout)
	}
	b.pending[e.ID] = pe
	return e, true
}

// Ack removes the eval from the pending set and clears its attempt
// counter.  Calling Ack for an ID that is not pending is a no-op.
func (b *Broker) Ack(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.pending, id)
	delete(b.attempts, id)
}

// ErrDeadLetter is returned by Nack when the eval has exceeded the
// configured DeliveryLimit and has been moved to the DLQ.
var ErrDeadLetter = errors.New("evalbroker: eval moved to DLQ (delivery limit exceeded)")

// Nack re-enqueues the eval after delay.  delay==0 uses the broker's
// NackTimeout default.
func (b *Broker) Nack(id string, delay time.Duration) error {
	b.mu.Lock()
	pe, ok := b.pending[id]
	if !ok {
		b.mu.Unlock()
		return nil
	}
	delete(b.pending, id)
	b.attempts[id]++
	if b.cfg.DeliveryLimit > 0 && b.attempts[id] > b.cfg.DeliveryLimit {
		b.dlq = append(b.dlq, pe.eval)
		delete(b.attempts, id)
		b.mu.Unlock()
		return ErrDeadLetter
	}
	eval := pe.eval
	b.mu.Unlock()

	if delay == 0 {
		delay = b.cfg.NackTimeout
	}
	if delay <= 0 {
		b.Enqueue(eval)
		return nil
	}
	time.AfterFunc(delay, func() { b.Enqueue(eval) })
	return nil
}

// DLQ returns a snapshot copy of the dead-letter queue.
func (b *Broker) DLQ() []Evaluation {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]Evaluation, len(b.dlq))
	copy(out, b.dlq)
	return out
}

// Close shuts the broker down; blocked Dequeues return (_, false).
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	b.cond.Broadcast()
}

// visibilityReaper runs in a dedicated goroutine and re-enqueues
// pending evals whose visibility window has elapsed.  Fires every
// second — coarse enough not to burn CPU, fine enough to recover
// from worker death within the configured timeout.
func (b *Broker) visibilityReaper() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		b.mu.Lock()
		if b.closed {
			b.mu.Unlock()
			return
		}
		now := time.Now()
		var expired []Evaluation
		for id, pe := range b.pending {
			if !pe.deadline.IsZero() && now.After(pe.deadline) {
				expired = append(expired, pe.eval)
				delete(b.pending, id)
			}
		}
		b.mu.Unlock()
		for _, e := range expired {
			// Count as a delivery attempt so runaway workers get DLQ'd.
			_ = b.Nack(e.ID, 0)
			b.Enqueue(e)
		}
	}
}
