package workqueue

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Item represents an item in the work queue
type Item struct {
	// Key is the unique identifier for the item
	Key string

	// RetryCount how many times this item has been retried
	RetryCount int

	// MaxRetries maximum attempts allowed
	MaxRetries int

	// LastError from last processing attempt
	LastError error

	// AddedAt when item was added
	AddedAt time.Time
}

// WorkQueue manages a queue of items to be processed
type WorkQueue interface {
	// Add adds an item to the queue
	Add(key string) error

	// AddAfter adds an item after a delay
	AddAfter(key string, duration time.Duration) error

	// AddRateLimited adds with rate limiting backoff
	AddRateLimited(key string) error

	// Get blocks until an item is available
	Get() (*Item, error)

	// Done marks an item as processed
	Done(key string) error

	// Forget forgets tracking of an item
	Forget(key string) error

	// NumRequeues returns retry count
	NumRequeues(key string) int

	// Shutdown stops the queue
	Shutdown() error

	// Len returns queue length
	Len() int
}

// RateLimiter defines rate limiting behavior
type RateLimiter interface {
	// When returns when the item should be retried
	When(item *Item) time.Duration

	// Forget tells limiter to forget the item
	Forget(key string)

	// NumRequeues returns retry count
	NumRequeues(key string) int
}

// DefaultRateLimiter implements exponential backoff
type DefaultRateLimiter struct {
	mu        sync.RWMutex
	baseDelay time.Duration
	maxDelay  time.Duration
	requeues  map[string]int
}

// NewDefaultRateLimiter creates a rate limiter
func NewDefaultRateLimiter(baseDelay, maxDelay time.Duration) *DefaultRateLimiter {
	return &DefaultRateLimiter{
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		requeues:  make(map[string]int),
	}
}

// When returns when to retry
func (drl *DefaultRateLimiter) When(item *Item) time.Duration {
	drl.mu.Lock()
	defer drl.mu.Unlock()

	drl.requeues[item.Key]++

	// Exponential backoff: baseDelay * 2^retries
	delay := drl.baseDelay
	for i := 0; i < item.RetryCount; i++ {
		delay = delay * 2
		if delay > drl.maxDelay {
			delay = drl.maxDelay
			break
		}
	}
	return delay
}

// Forget removes tracking
func (drl *DefaultRateLimiter) Forget(key string) {
	drl.mu.Lock()
	defer drl.mu.Unlock()
	delete(drl.requeues, key)
}

// NumRequeues returns retry count
func (drl *DefaultRateLimiter) NumRequeues(key string) int {
	drl.mu.RLock()
	defer drl.mu.RUnlock()
	return drl.requeues[key]
}

// SimpleQueue implements a basic work queue
type SimpleQueue struct {
	mu          sync.Mutex
	queue       []string
	inProgress  map[string]bool
	dirty       map[string]bool
	rateLimiter RateLimiter
	cond        *sync.Cond
	shutdown    bool
}

// NewSimpleQueue creates a new work queue
func NewSimpleQueue(rateLimiter RateLimiter) *SimpleQueue {
	if rateLimiter == nil {
		rateLimiter = NewDefaultRateLimiter(1*time.Millisecond, 16*time.Second)
	}

	sq := &SimpleQueue{
		queue:       make([]string, 0),
		inProgress:  make(map[string]bool),
		dirty:       make(map[string]bool),
		rateLimiter: rateLimiter,
	}

	sq.cond = sync.NewCond(&sq.mu)
	return sq
}

// Add adds an item to the queue
func (sq *SimpleQueue) Add(key string) error {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	if sq.shutdown {
		return fmt.Errorf("queue is shut down")
	}

	if sq.dirty[key] {
		return nil // Already queued
	}

	sq.queue = append(sq.queue, key)
	sq.dirty[key] = true
	sq.cond.Signal()

	return nil
}

// AddAfter adds an item after a delay
func (sq *SimpleQueue) AddAfter(key string, duration time.Duration) error {
	go func() {
		<-time.After(duration)
		sq.Add(key)
	}()
	return nil
}

// AddRateLimited adds with rate limiting
func (sq *SimpleQueue) AddRateLimited(key string) error {
	numRequeues := sq.rateLimiter.NumRequeues(key)
	item := &Item{
		Key:        key,
		RetryCount: numRequeues,
	}
	delay := sq.rateLimiter.When(item)
	return sq.AddAfter(key, delay)
}

// Get blocks until an item is available
func (sq *SimpleQueue) Get() (*Item, error) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	for {
		if sq.shutdown {
			return nil, fmt.Errorf("queue is shut down")
		}

		if len(sq.queue) > 0 {
			key := sq.queue[0]
			sq.queue = sq.queue[1:]
			sq.inProgress[key] = true
			delete(sq.dirty, key)

			return &Item{
				Key:        key,
				RetryCount: sq.rateLimiter.NumRequeues(key),
				AddedAt:    time.Now(),
			}, nil
		}

		sq.cond.Wait()
	}
}

// Done marks an item as processed
func (sq *SimpleQueue) Done(key string) error {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	delete(sq.inProgress, key)
	sq.rateLimiter.Forget(key)

	return nil
}

// Forget forgets an item
func (sq *SimpleQueue) Forget(key string) error {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	sq.rateLimiter.Forget(key)
	delete(sq.dirty, key)
	delete(sq.inProgress, key)

	return nil
}

// NumRequeues returns retry count
func (sq *SimpleQueue) NumRequeues(key string) int {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	return sq.rateLimiter.NumRequeues(key)
}

// Shutdown stops the queue
func (sq *SimpleQueue) Shutdown() error {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	sq.shutdown = true
	sq.cond.Broadcast()

	return nil
}

// Len returns queue length
func (sq *SimpleQueue) Len() int {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	return len(sq.queue)
}

// PriorityQueue wraps a work queue with priority support
type PriorityQueue struct {
	queues     []WorkQueue
	priorities map[string]int
	mu         sync.Mutex
}

// NewPriorityQueue creates a priority queue
func NewPriorityQueue(numQueues int) *PriorityQueue {
	queues := make([]WorkQueue, numQueues)
	for i := 0; i < numQueues; i++ {
		queues[i] = NewSimpleQueue(nil)
	}

	return &PriorityQueue{
		queues:     queues,
		priorities: make(map[string]int),
	}
}

// Add adds an item with priority
func (pq *PriorityQueue) Add(key string) error {
	return pq.AddWithPriority(key, 0)
}

// AddWithPriority adds an item with specific priority
func (pq *PriorityQueue) AddWithPriority(key string, priority int) error {
	pq.mu.Lock()
	if priority >= len(pq.queues) {
		priority = len(pq.queues) - 1
	}
	if priority < 0 {
		priority = 0
	}
	pq.priorities[key] = priority
	pq.mu.Unlock()

	return pq.queues[priority].Add(key)
}

// Get blocks until an item from highest priority queue is available
func (pq *PriorityQueue) Get() (*Item, error) {
	// Try higher priority queues first
	for i := len(pq.queues) - 1; i >= 0; i-- {
		if pq.queues[i].Len() > 0 {
			return pq.queues[i].Get()
		}
	}

	// Block on highest priority queue
	return pq.queues[len(pq.queues)-1].Get()
}

// Done marks item as done in all queues
func (pq *PriorityQueue) Done(key string) error {
	for _, q := range pq.queues {
		q.Done(key)
	}
	return nil
}

// AddAfter adds item after delay
func (pq *PriorityQueue) AddAfter(key string, duration time.Duration) error {
	pq.mu.Lock()
	priority := pq.priorities[key]
	pq.mu.Unlock()

	go func() {
		<-time.After(duration)
		pq.AddWithPriority(key, priority)
	}()
	return nil
}

// AddRateLimited adds with rate limiting
func (pq *PriorityQueue) AddRateLimited(key string) error {
	pq.mu.Lock()
	priority := pq.priorities[key]
	pq.mu.Unlock()

	return pq.queues[priority].AddRateLimited(key)
}

// Forget forgets an item
func (pq *PriorityQueue) Forget(key string) error {
	pq.mu.Lock()
	delete(pq.priorities, key)
	pq.mu.Unlock()

	for _, q := range pq.queues {
		q.Forget(key)
	}
	return nil
}

// NumRequeues returns retry count
func (pq *PriorityQueue) NumRequeues(key string) int {
	pq.mu.Lock()
	priority := pq.priorities[key]
	pq.mu.Unlock()

	return pq.queues[priority].NumRequeues(key)
}

// Shutdown shuts down all queues
func (pq *PriorityQueue) Shutdown() error {
	for _, q := range pq.queues {
		q.Shutdown()
	}
	return nil
}

// Len returns total queue length
func (pq *PriorityQueue) Len() int {
	total := 0
	for _, q := range pq.queues {
		total += q.Len()
	}
	return total
}

// ProcessFunc is called to process items from queue
type ProcessFunc func(ctx context.Context, item *Item) error

// Worker processes items from a queue
type Worker struct {
	queue       WorkQueue
	processFunc ProcessFunc
	maxRetries  int
}

// NewWorker creates a new worker
func NewWorker(queue WorkQueue, processFunc ProcessFunc, maxRetries int) *Worker {
	return &Worker{
		queue:       queue,
		processFunc: processFunc,
		maxRetries:  maxRetries,
	}
}

// Run starts the worker processing loop
func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return w.queue.Shutdown()
		default:
		}

		item, err := w.queue.Get()
		if err != nil {
			return err
		}

		// Process the item
		err = w.processFunc(ctx, item)

		if err != nil {
			if item.RetryCount < w.maxRetries {
				item.RetryCount++
				item.LastError = err
				w.queue.AddRateLimited(item.Key)
			} else {
				// Max retries exceeded, give up
				w.queue.Forget(item.Key)
			}
		} else {
			// Success
			w.queue.Done(item.Key)
		}
	}
}
