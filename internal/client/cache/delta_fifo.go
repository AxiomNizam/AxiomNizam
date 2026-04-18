// Package cache — DeltaFIFO is the queue the reflector writes to and
// the informer reads from.  Each queue entry is a list of Delta
// values for a single object key, in the order the reflector
// observed them.  The compression rules mirror upstream client-go:
//
//   - Added followed by Deleted → queue entry retained; both deltas
//     emitted so the listener sees the lifecycle.
//   - Updated followed by Updated → both retained (listeners may
//     care about intermediate state).
//   - Deleted followed by Deleted → deduplicated.
//
// The Sync delta is a special variant used during periodic
// resync-all operations; listeners typically treat it as Update.
package cache

import (
	"errors"
	"sync"
)

// DeltaType enumerates the change kinds.
type DeltaType string

const (
	DeltaAdded   DeltaType = "Added"
	DeltaUpdated DeltaType = "Updated"
	DeltaDeleted DeltaType = "Deleted"
	DeltaSync    DeltaType = "Sync"
)

// Delta is a single observation.
type Delta struct {
	Type   DeltaType
	Object interface{}
}

// Deltas is the ordered list of observations for one key.
type Deltas []Delta

// Newest returns the most-recent delta, or zero if empty.
func (d Deltas) Newest() Delta {
	if len(d) == 0 {
		return Delta{}
	}
	return d[len(d)-1]
}

// DeltaFIFO is an unbounded FIFO keyed by object key.  Entries are
// popped in order of first insertion — so a key added, updated,
// then popped yields both deltas in one Pop call, preserving
// ordering across keys.
type DeltaFIFO struct {
	mu      sync.Mutex
	cond    *sync.Cond
	keyFunc KeyFunc

	// items maps key → accumulated deltas awaiting Pop.
	items map[string]Deltas
	// queue is the FIFO order.
	queue []string
	// closed signals shutdown to blocked Pops.
	closed bool
}

// NewDeltaFIFO constructs an empty queue keyed by keyFunc.
func NewDeltaFIFO(keyFunc KeyFunc) *DeltaFIFO {
	if keyFunc == nil {
		panic("cache.NewDeltaFIFO: KeyFunc is required")
	}
	f := &DeltaFIFO{
		keyFunc: keyFunc,
		items:   map[string]Deltas{},
	}
	f.cond = sync.NewCond(&f.mu)
	return f
}

// ErrFIFOClosed is returned by Pop after Close has been called.
var ErrFIFOClosed = errors.New("cache: DeltaFIFO is closed")

// Add records an Added delta.
func (f *DeltaFIFO) Add(obj interface{}) error { return f.queueDelta(DeltaAdded, obj) }

// Update records an Updated delta.
func (f *DeltaFIFO) Update(obj interface{}) error { return f.queueDelta(DeltaUpdated, obj) }

// Delete records a Deleted delta.
func (f *DeltaFIFO) Delete(obj interface{}) error { return f.queueDelta(DeltaDeleted, obj) }

// Sync records a Sync delta — used by periodic relist.
func (f *DeltaFIFO) Sync(obj interface{}) error { return f.queueDelta(DeltaSync, obj) }

// queueDelta is the shared enqueue path.
func (f *DeltaFIFO) queueDelta(t DeltaType, obj interface{}) error {
	key, err := f.keyFunc(obj)
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return ErrFIFOClosed
	}
	existing, present := f.items[key]
	delta := Delta{Type: t, Object: obj}
	existing = dedupCompress(append(existing, delta))
	f.items[key] = existing
	if !present {
		f.queue = append(f.queue, key)
		f.cond.Signal()
	}
	return nil
}

// Pop blocks until at least one entry is available, then returns
// the full Deltas slice for the head key.  The caller MUST process
// every delta before the next Pop — losing deltas here means the
// cache drifts from reality.
func (f *DeltaFIFO) Pop() (Deltas, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for len(f.queue) == 0 && !f.closed {
		f.cond.Wait()
	}
	if f.closed {
		return nil, ErrFIFOClosed
	}
	key := f.queue[0]
	f.queue = f.queue[1:]
	deltas := f.items[key]
	delete(f.items, key)
	return deltas, nil
}

// Close wakes all blocked Pops and prevents further Add/Update/Delete.
func (f *DeltaFIFO) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	f.cond.Broadcast()
}

// Replace is the reflector's bulk-resync entry point.  Every key
// present in items is enqueued with a Sync delta; keys absent from
// items but present in the queue get a synthetic Deleted delta so
// listeners learn the object disappeared.
func (f *DeltaFIFO) Replace(items []interface{}) error {
	seen := map[string]struct{}{}
	for _, obj := range items {
		if err := f.Sync(obj); err != nil {
			return err
		}
		k, err := f.keyFunc(obj)
		if err == nil {
			seen[k] = struct{}{}
		}
	}
	// Tombstones for keys no longer present — only those still
	// queued get synthetic deletes; consumers that already popped a
	// key no longer in items have no way to learn about the deletion
	// from us, and must rely on the next Sync cycle.
	f.mu.Lock()
	stale := make([]string, 0)
	for k := range f.items {
		if _, ok := seen[k]; !ok {
			stale = append(stale, k)
		}
	}
	f.mu.Unlock()
	for _, k := range stale {
		// Synthesise a tombstone — carry only the key since the
		// object value is unknown at the reflector layer.  Listeners
		// must tolerate a string-valued Deleted delta.
		if err := f.Delete(tombstone{Key: k}); err != nil {
			return err
		}
	}
	return nil
}

// tombstone is the marker used for synthetic deletions.  The
// informer recognises this by type-switching on Object in Deleted.
type tombstone struct{ Key string }

// dedupCompress applies the upstream compression rules.  Called with
// the list already extended by the new delta.
func dedupCompress(d Deltas) Deltas {
	if len(d) < 2 {
		return d
	}
	last := d[len(d)-1]
	prev := d[len(d)-2]
	// Collapse consecutive Deletes of the same key.
	if last.Type == DeltaDeleted && prev.Type == DeltaDeleted {
		return d[:len(d)-1]
	}
	return d
}
