// Package builder — a minimal in-memory queue used when the caller
// does not supply a QueueFactory.  It implements the Queue interface
// with a lock-protected slice and a condition variable.
//
// Production code should wire a workqueue.RateLimitingInterface from
// internal/workqueue here — this simple implementation lacks rate
// limiting and exists so that controllers are testable in isolation
// without pulling the full workqueue stack.
package builder

import "sync"

// simpleQueue is a FIFO with deduplication and shutdown signalling.
// AddRateLimited collapses to Add — callers that need backoff must
// supply a real rate-limited queue via QueueFactory.
type simpleQueue struct {
	mu         sync.Mutex
	cond       *sync.Cond
	items      []interface{}
	dirty      map[interface{}]struct{}
	processing map[interface{}]struct{}
	shutdown   bool
}

// newSimpleQueue returns an empty queue.
func newSimpleQueue() *simpleQueue {
	q := &simpleQueue{
		dirty:      map[interface{}]struct{}{},
		processing: map[interface{}]struct{}{},
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Add enqueues item, deduplicating against anything already pending
// or currently being processed.  Returning to the queue while in
// processing is deferred until Done is called.
func (q *simpleQueue) Add(item interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.shutdown {
		return
	}
	if _, ok := q.dirty[item]; ok {
		return
	}
	q.dirty[item] = struct{}{}
	if _, ok := q.processing[item]; ok {
		// Will be added after Done clears processing.
		return
	}
	q.items = append(q.items, item)
	q.cond.Signal()
}

// AddRateLimited degrades to Add here — see package doc.
func (q *simpleQueue) AddRateLimited(item interface{}) { q.Add(item) }

// Get blocks until an item is available or the queue is shut down.
// Callers MUST call Done when finished.
func (q *simpleQueue) Get() (interface{}, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 && !q.shutdown {
		q.cond.Wait()
	}
	if q.shutdown {
		return nil, true
	}
	item := q.items[0]
	q.items = q.items[1:]
	q.processing[item] = struct{}{}
	delete(q.dirty, item)
	return item, false
}

// Done marks the item as no longer being processed.  If Add was
// called on the same item while it was in-flight, the item is
// immediately re-enqueued.
func (q *simpleQueue) Done(item interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.processing, item)
	if _, requeued := q.dirty[item]; requeued {
		q.items = append(q.items, item)
		q.cond.Signal()
	}
}

// Forget is a no-op for this trivial queue — it exists for API parity
// with the rate-limited variant, where Forget clears the backoff.
func (q *simpleQueue) Forget(_ interface{}) {}

// ShutDown marks the queue closed and wakes every waiting Get.
func (q *simpleQueue) ShutDown() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.shutdown = true
	q.cond.Broadcast()
}
