package apiresource

import (
	"context"
	"sync"
	"time"
)

// WorkQueue manages items to be reconciled
type WorkQueue struct {
	mu         sync.Mutex
	items      map[string]*QueueItem // key -> item
	queue      []*QueueItem          // ordered list
	processing map[string]bool       // currently processing
	retries    map[string]int        // retry count
	maxRetries int
}

// QueueItem represents something to reconcile
type QueueItem struct {
	Key         string // namespace/name
	RetryCount  int
	AddedAt     time.Time
	ProcessedAt time.Time
}

// NewWorkQueue creates a new work queue
func NewWorkQueue(maxRetries int) *WorkQueue {
	if maxRetries == 0 {
		maxRetries = 5
	}
	return &WorkQueue{
		items:      make(map[string]*QueueItem),
		queue:      make([]*QueueItem, 0),
		processing: make(map[string]bool),
		retries:    make(map[string]int),
		maxRetries: maxRetries,
	}
}

// Add adds an item to the queue
func (wq *WorkQueue) Add(key string) {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	// If already exists and being processed, skip
	if wq.processing[key] {
		return
	}

	item := &QueueItem{
		Key:     key,
		AddedAt: time.Now(),
	}

	if _, exists := wq.items[key]; !exists {
		wq.items[key] = item
		wq.queue = append(wq.queue, item)
	}
}

// Get retrieves next item from queue
func (wq *WorkQueue) Get(ctx context.Context) (string, error) {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	// Wait for items if queue is empty
	for len(wq.queue) == 0 {
		wq.mu.Unlock()
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(100 * time.Millisecond):
			wq.mu.Lock()
		}
	}

	// Get first item
	item := wq.queue[0]
	wq.queue = wq.queue[1:]

	// Mark as processing
	wq.processing[item.Key] = true
	item.ProcessedAt = time.Now()

	return item.Key, nil
}

// Done marks item as processed successfully
func (wq *WorkQueue) Done(key string) {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	delete(wq.processing, key)
	delete(wq.items, key)
	wq.retries[key] = 0 // Clear retry count on success
}

// Forget removes an item from the queue and retries
func (wq *WorkQueue) Forget(key string) {
	wq.Done(key)
}

// AddAfter requeues item after delay
func (wq *WorkQueue) AddAfter(key string, delay time.Duration) {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	wq.processing[key] = false // Mark as no longer processing

	wq.retries[key]++
	if wq.retries[key] > wq.maxRetries {
		// Max retries exceeded, remove from queue
		delete(wq.items, key)
		delete(wq.retries, key)
		return
	}

	// Re-add after delay
	go func() {
		time.Sleep(delay)
		wq.Add(key)
	}()
}

// Len returns queue length
func (wq *WorkQueue) Len() int {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	return len(wq.queue)
}

// ProcessingLen returns number of items being processed
func (wq *WorkQueue) ProcessingLen() int {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	count := 0
	for _, v := range wq.processing {
		if v {
			count++
		}
	}
	return count
}

// Shutdown stops the work queue
func (wq *WorkQueue) Shutdown() {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	wq.items = make(map[string]*QueueItem)
	wq.queue = make([]*QueueItem, 0)
	wq.processing = make(map[string]bool)
	wq.retries = make(map[string]int)
}
