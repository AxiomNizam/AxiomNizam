// Package informer — shared informer fan-out.
//
// A SharedInformer wraps a ListerWatcher and a single Indexer, and
// multiplexes events to any number of ResourceEventHandlers.  Sharing
// the cache across handlers is the whole point of the pattern: it
// amortises the cost of the initial List across N controllers that
// would otherwise each pay it independently.
package informer

import (
	"context"
	"sync"
)

// SharedInformer is the composed type.  Callers should treat it as
// write-once (call Run exactly once) and read-many.
type SharedInformer struct {
	lw      ListerWatcher
	indexer *Indexer

	mu         sync.RWMutex
	handlers   []ResourceEventHandler
	started    bool
	synced     bool
	syncedCh   chan struct{}
	stopReason error
}

// NewSharedInformer constructs an informer backed by lw and indexed by
// indexer.  Indexer is exposed via Indexer() for read-only queries.
func NewSharedInformer(lw ListerWatcher, indexer *Indexer) *SharedInformer {
	return &SharedInformer{
		lw:       lw,
		indexer:  indexer,
		syncedCh: make(chan struct{}),
	}
}

// Indexer returns the underlying read-optimised cache.
func (s *SharedInformer) Indexer() *Indexer { return s.indexer }

// AddEventHandler registers h.  Handlers added before Run receive the
// initial list as a synthetic stream of OnAdd calls; handlers added
// after Run has synced receive a replay of the current cache so that
// the handler's state machine starts from a consistent snapshot.
func (s *SharedInformer) AddEventHandler(h ResourceEventHandler) {
	s.mu.Lock()
	s.handlers = append(s.handlers, h)
	synced := s.synced
	items := s.indexer.List()
	s.mu.Unlock()

	if synced {
		for _, obj := range items {
			h.OnAdd(obj)
		}
	}
}

// HasSynced reports whether the initial List has completed.  Callers
// should gate reconciler startup on this: acting on a partial cache
// can cause false deletes when an object that actually exists simply
// has not been listed yet.
func (s *SharedInformer) HasSynced() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.synced
}

// WaitForCacheSync blocks until HasSynced returns true or ctx expires.
// Returns the context's error on cancellation; nil on successful sync.
func (s *SharedInformer) WaitForCacheSync(ctx context.Context) error {
	select {
	case <-s.syncedCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Run starts the event pump.  It blocks until ctx is done; callers
// typically invoke it in a goroutine at startup.  Run is idempotent:
// calling it twice returns immediately on the second invocation.
func (s *SharedInformer) Run(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return nil
	}
	s.started = true
	s.mu.Unlock()

	// 1) Initial LIST — populates the cache before any handlers run.
	initial, err := s.lw.List(ctx)
	if err != nil {
		return err
	}
	for _, obj := range initial {
		if err := s.indexer.Add(obj); err != nil {
			// Indexing errors are non-fatal at this stage — they
			// indicate a malformed object that will never be
			// reconciled.  Log-style reporting is left to callers
			// which may wrap this type.
			continue
		}
	}
	s.mu.Lock()
	handlers := append([]ResourceEventHandler(nil), s.handlers...)
	s.mu.Unlock()
	for _, obj := range initial {
		for _, h := range handlers {
			h.OnAdd(obj)
		}
	}

	// 2) Mark cache synced and release any WaitForCacheSync waiters.
	s.mu.Lock()
	s.synced = true
	close(s.syncedCh)
	s.mu.Unlock()

	// 3) Watch loop.
	watchCh, err := s.lw.Watch(ctx)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, open := <-watchCh:
			if !open {
				// Upstream closed; surface as nil so that caller can
				// decide whether to restart Run.
				return nil
			}
			s.dispatch(ev)
		}
	}
}

// dispatch updates the cache and invokes handlers for a single event.
func (s *SharedInformer) dispatch(ev Event) {
	var old map[string]interface{}
	switch ev.Type {
	case EventAdd:
		_ = s.indexer.Add(ev.Obj)
	case EventUpdate:
		if key, err := s.indexer.keyFunc(ev.Obj); err == nil {
			old, _ = s.indexer.Get(key)
		}
		_ = s.indexer.Add(ev.Obj)
	case EventDelete:
		_ = s.indexer.Delete(ev.Obj)
	}

	s.mu.RLock()
	handlers := append([]ResourceEventHandler(nil), s.handlers...)
	s.mu.RUnlock()

	for _, h := range handlers {
		switch ev.Type {
		case EventAdd:
			h.OnAdd(ev.Obj)
		case EventUpdate:
			if old == nil {
				// No prior object in cache → treat as Add for
				// handlers that care about transition direction.
				h.OnAdd(ev.Obj)
			} else {
				h.OnUpdate(old, ev.Obj)
			}
		case EventDelete:
			h.OnDelete(ev.Obj)
		}
	}
}
