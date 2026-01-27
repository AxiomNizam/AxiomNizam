package distributedstate

import (
	"context"
	"sync"
	"time"
)

type CachedStateStore struct {
	underlying StateStore
	cache      map[string]cacheEntry
	mu         sync.RWMutex
	ttl        time.Duration
}

type cacheEntry struct {
	value     string
	revision  int64
	timestamp time.Time
}

func NewCachedStateStore(underlying StateStore, ttl time.Duration) *CachedStateStore {
	return &CachedStateStore{
		underlying: underlying,
		cache:      make(map[string]cacheEntry),
		ttl:        ttl,
	}
}

func (cs *CachedStateStore) Put(ctx context.Context, key string, value string) error {
	err := cs.underlying.Put(ctx, key, value)
	if err == nil {
		cs.mu.Lock()
		cs.cache[key] = cacheEntry{
			value:     value,
			timestamp: time.Now(),
		}
		cs.mu.Unlock()
	}
	return err
}

func (cs *CachedStateStore) Get(ctx context.Context, key string) (string, error) {
	cs.mu.RLock()
	if entry, exists := cs.cache[key]; exists {
		if time.Since(entry.timestamp) < cs.ttl {
			cs.mu.RUnlock()
			return entry.value, nil
		}
	}
	cs.mu.RUnlock()

	value, err := cs.underlying.Get(ctx, key)
	if err == nil {
		cs.mu.Lock()
		cs.cache[key] = cacheEntry{
			value:     value,
			timestamp: time.Now(),
		}
		cs.mu.Unlock()
	}
	return value, err
}

func (cs *CachedStateStore) Delete(ctx context.Context, key string) error {
	err := cs.underlying.Delete(ctx, key)
	if err == nil {
		cs.mu.Lock()
		delete(cs.cache, key)
		cs.mu.Unlock()
	}
	return err
}

func (cs *CachedStateStore) CompareAndSwap(ctx context.Context, key string, oldValue, newValue string) (bool, error) {
	success, err := cs.underlying.CompareAndSwap(ctx, key, oldValue, newValue)
	if success && err == nil {
		cs.mu.Lock()
		cs.cache[key] = cacheEntry{
			value:     newValue,
			timestamp: time.Now(),
		}
		cs.mu.Unlock()
	}
	return success, err
}

func (cs *CachedStateStore) GetWithRevision(ctx context.Context, key string) (string, int64, error) {
	value, revision, err := cs.underlying.GetWithRevision(ctx, key)
	if err == nil {
		cs.mu.Lock()
		cs.cache[key] = cacheEntry{
			value:     value,
			revision:  revision,
			timestamp: time.Now(),
		}
		cs.mu.Unlock()
	}
	return value, revision, err
}

func (cs *CachedStateStore) PutWithLease(ctx context.Context, key string, value string, ttl time.Duration) (int64, error) {
	leaseID, err := cs.underlying.PutWithLease(ctx, key, value, ttl)
	if err == nil {
		cs.mu.Lock()
		cs.cache[key] = cacheEntry{
			value:     value,
			timestamp: time.Now(),
		}
		cs.mu.Unlock()
	}
	return leaseID, err
}

func (cs *CachedStateStore) ReleaseLeaseID(ctx context.Context, leaseID int64) error {
	err := cs.underlying.ReleaseLeaseID(ctx, leaseID)
	if err == nil {
		cs.mu.Lock()
		for k, v := range cs.cache {
			if time.Since(v.timestamp) > cs.ttl {
				delete(cs.cache, k)
			}
		}
		cs.mu.Unlock()
	}
	return err
}

func (cs *CachedStateStore) Watch(ctx context.Context, key string, callback func(Event)) error {
	return cs.underlying.Watch(ctx, key, func(event Event) {
		cs.mu.Lock()
		if event.Type == EventTypeDelete {
			delete(cs.cache, key)
		} else {
			cs.cache[key] = cacheEntry{
				value:     event.Value,
				revision:  event.Revision,
				timestamp: time.Now(),
			}
		}
		cs.mu.Unlock()
		callback(event)
	})
}

func (cs *CachedStateStore) List(ctx context.Context, prefix string) (map[string]string, error) {
	return cs.underlying.List(ctx, prefix)
}

func (cs *CachedStateStore) Transaction(ctx context.Context, ops []Operation) (bool, error) {
	success, err := cs.underlying.Transaction(ctx, ops)
	if success && err == nil {
		cs.mu.Lock()
		for _, op := range ops {
			if op.Type == OperationDelete {
				delete(cs.cache, op.Key)
			} else if op.Type == OperationPut {
				cs.cache[op.Key] = cacheEntry{
					value:     op.Value,
					timestamp: time.Now(),
				}
			}
		}
		cs.mu.Unlock()
	}
	return success, err
}

func (cs *CachedStateStore) Invalidate(key string) {
	cs.mu.Lock()
	delete(cs.cache, key)
	cs.mu.Unlock()
}

func (cs *CachedStateStore) InvalidateAll() {
	cs.mu.Lock()
	cs.cache = make(map[string]cacheEntry)
	cs.mu.Unlock()
}

func (cs *CachedStateStore) Close() error {
	cs.mu.Lock()
	cs.cache = make(map[string]cacheEntry)
	cs.mu.Unlock()
	return cs.underlying.Close()
}
