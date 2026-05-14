// Package workqueue — delaying queue backed by a single-timer min-heap.
//
// The pre-existing SimpleQueue.AddAfter implementation spawns a
// goroutine per delayed item:
//
//	go func() {
//	    <-time.After(duration)
//	    sq.Add(key)
//	}()
//
// That works for low cardinality but collapses under load: retrying a
// thousand resources with exponential backoff creates a thousand long-
// lived goroutines, each pinning a timer in the runtime's timer heap.
// DelayingQueue replaces that model with a single background goroutine
// that multiplexes a heap of pending items onto one timer — the same
// design used by client-go's DelayingInterface.
//
// The queue embeds a caller-supplied WorkQueue (typically SimpleQueue)
// and delegates every method except AddAfter to it, so existing code
// can upgrade by swapping the constructor without touching consumer
// logic.
package workqueue

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

// delayedItem is a pending entry in the heap.  Sorted ascending by
// readyAt so that the next item to release is always heap[0].
type delayedItem struct {
	key     string
	readyAt time.Time
	index   int // maintained by heap.Interface
}

// delayedHeap implements heap.Interface.  The heap stores pointers so
// that index back-references remain valid after Push/Pop reshuffling.
type delayedHeap []*delayedItem

// Len implements heap.Interface.
func (h delayedHeap) Len() int { return len(h) }

// Less implements heap.Interface with ascending readyAt ordering.
func (h delayedHeap) Less(i, j int) bool { return h[i].readyAt.Before(h[j].readyAt) }

// Swap implements heap.Interface and maintains the index back-reference.
func (h delayedHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

// Push implements heap.Interface.
func (h *delayedHeap) Push(x interface{}) {
	item := x.(*delayedItem)
	item.index = len(*h)
	*h = append(*h, item)
}

// Pop implements heap.Interface.
func (h *delayedHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

// DelayingQueue wraps an underlying WorkQueue with an efficient
// AddAfter implementation.  All non-delay methods are simple pass-
// throughs; callers can use the resulting value anywhere a WorkQueue
// is expected.
type DelayingQueue struct {
	WorkQueue // underlying queue (SimpleQueue, metrics-wrapped queue, …)

	mu       sync.Mutex
	heap     delayedHeap
	wake     chan struct{} // buffered size 1; signalled when head may have moved earlier
	shutdown bool
	wg       sync.WaitGroup
}

// NewDelayingQueue wraps inner with a delayed-insertion timer.  The
// caller is responsible for calling Shutdown to stop the background
// goroutine; the queue's own Shutdown method chains through.
func NewDelayingQueue(inner WorkQueue) *DelayingQueue {
	dq := &DelayingQueue{
		WorkQueue: inner,
		wake:      make(chan struct{}, 1),
	}
	dq.wg.Add(1)
	go dq.run(context.Background())
	return dq
}

// signal performs a non-blocking send on the wake channel.  The
// channel is buffered to size 1 so a pending signal coalesces multiple
// rapid broadcasts into a single wake-up.
func (dq *DelayingQueue) signal() {
	select {
	case dq.wake <- struct{}{}:
	default:
	}
}

// AddAfter schedules key to be added to the underlying queue after
// duration.  Duplicate keys with different readyAt times are collapsed
// to the earliest time — matching the semantics of client-go.
func (dq *DelayingQueue) AddAfter(key string, duration time.Duration) error {
	if duration <= 0 {
		return dq.WorkQueue.Add(key)
	}
	readyAt := time.Now().Add(duration)

	dq.mu.Lock()
	if dq.shutdown {
		dq.mu.Unlock()
		return nil
	}

	// Check for an existing scheduled entry with the same key; collapse
	// to the earliest deadline so callers don't accidentally postpone
	// a more-urgent retry by re-adding with a longer backoff.
	for _, it := range dq.heap {
		if it.key == key {
			if readyAt.Before(it.readyAt) {
				it.readyAt = readyAt
				heap.Fix(&dq.heap, it.index)
				dq.mu.Unlock()
				dq.signal()
				return nil
			}
			dq.mu.Unlock()
			return nil
		}
	}

	heap.Push(&dq.heap, &delayedItem{key: key, readyAt: readyAt})
	dq.mu.Unlock()
	dq.signal()
	return nil
}

// Shutdown stops the background goroutine after flushing any items
// whose deadline has already passed into the underlying queue.  It
// then chains the call through to the underlying queue.
func (dq *DelayingQueue) Shutdown() error {
	dq.mu.Lock()
	dq.shutdown = true
	dq.mu.Unlock()
	dq.signal()
	dq.wg.Wait()
	return dq.WorkQueue.Shutdown()
}

// run is the single goroutine that advances the heap.  It blocks on
// the wake channel when the heap is empty and otherwise sleeps in a
// select that wakes early when a newly-added item has an earlier
// deadline than the current head.
func (dq *DelayingQueue) run(_ context.Context) {
	defer dq.wg.Done()
	const idleSleep = time.Hour // upper bound when heap is empty

	for {
		dq.mu.Lock()
		if dq.shutdown && len(dq.heap) == 0 {
			dq.mu.Unlock()
			return
		}

		var wait time.Duration
		if len(dq.heap) == 0 {
			wait = idleSleep
		} else {
			head := dq.heap[0]
			wait = time.Until(head.readyAt)
			if wait <= 0 {
				// Head is due — pop and forward outside the lock.
				item := heap.Pop(&dq.heap).(*delayedItem)
				dq.mu.Unlock()
				_ = dq.WorkQueue.Add(item.key)
				continue
			}
		}
		dq.mu.Unlock()

		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
		case <-dq.wake:
			timer.Stop()
		}
	}
}
