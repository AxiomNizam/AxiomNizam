package distributedstate

import (
	"context"
	"sync"
	"time"
)

type InMemoryStateStore struct {
	mu          sync.RWMutex
	data        map[string]string
	revisions   map[string]int64
	watchers    map[string][]func(Event)
	leases      map[int64]*leaseEntry
	nextLeaseID int64
	revision    int64
}

type leaseEntry struct {
	key       string
	expireAt  time.Time
	tickerCtx context.Context
}

func NewInMemoryStateStore() *InMemoryStateStore {
	store := &InMemoryStateStore{
		data:      make(map[string]string),
		revisions: make(map[string]int64),
		watchers:  make(map[string][]func(Event)),
		leases:    make(map[int64]*leaseEntry),
	}
	go store.leaseCleanup()
	return store
}

func (im *InMemoryStateStore) Put(ctx context.Context, key string, value string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	oldValue := im.data[key]
	im.data[key] = value
	im.revision++
	im.revisions[key] = im.revision

	im.notifyWatchers(Event{
		Type:      EventTypePut,
		Key:       key,
		Value:     value,
		OldValue:  oldValue,
		Revision:  im.revision,
		Timestamp: time.Now(),
	})

	return nil
}

func (im *InMemoryStateStore) Get(ctx context.Context, key string) (string, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.data[key], nil
}

func (im *InMemoryStateStore) Delete(ctx context.Context, key string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	oldValue := im.data[key]
	delete(im.data, key)
	delete(im.revisions, key)
	im.revision++

	im.notifyWatchers(Event{
		Type:      EventTypeDelete,
		Key:       key,
		OldValue:  oldValue,
		Revision:  im.revision,
		Timestamp: time.Now(),
	})

	return nil
}

func (im *InMemoryStateStore) CompareAndSwap(ctx context.Context, key string, oldValue, newValue string) (bool, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	currentValue := im.data[key]
	if currentValue != oldValue {
		return false, nil
	}

	im.data[key] = newValue
	im.revision++
	im.revisions[key] = im.revision

	im.notifyWatchers(Event{
		Type:      EventTypePut,
		Key:       key,
		Value:     newValue,
		OldValue:  oldValue,
		Revision:  im.revision,
		Timestamp: time.Now(),
	})

	return true, nil
}

func (im *InMemoryStateStore) GetWithRevision(ctx context.Context, key string) (string, int64, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.data[key], im.revisions[key], nil
}

func (im *InMemoryStateStore) PutWithLease(ctx context.Context, key string, value string, ttl time.Duration) (int64, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.data[key] = value
	im.revision++
	im.revisions[key] = im.revision

	im.nextLeaseID++
	leaseID := im.nextLeaseID
	im.leases[leaseID] = &leaseEntry{
		key:      key,
		expireAt: time.Now().Add(ttl),
	}

	return leaseID, nil
}

func (im *InMemoryStateStore) ReleaseLeaseID(ctx context.Context, leaseID int64) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	entry, exists := im.leases[leaseID]
	if exists {
		delete(im.data, entry.key)
		delete(im.leases, leaseID)
	}

	return nil
}

func (im *InMemoryStateStore) Watch(ctx context.Context, key string, callback func(Event)) error {
	im.mu.Lock()
	im.watchers[key] = append(im.watchers[key], callback)
	im.mu.Unlock()

	return nil
}

func (im *InMemoryStateStore) List(ctx context.Context, prefix string) (map[string]string, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range im.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			result[k] = v
		}
	}

	return result, nil
}

func (im *InMemoryStateStore) Transaction(ctx context.Context, ops []Operation) (bool, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	for _, op := range ops {
		if op.Type == OperationCompare {
			if im.data[op.Key] != op.OldValue {
				return false, nil
			}
		}
	}

	for _, op := range ops {
		switch op.Type {
		case OperationPut:
			im.data[op.Key] = op.Value
			im.revision++
			im.revisions[op.Key] = im.revision
		case OperationDelete:
			delete(im.data, op.Key)
			delete(im.revisions, op.Key)
			im.revision++
		}
	}

	return true, nil
}

func (im *InMemoryStateStore) notifyWatchers(event Event) {
	if callbacks, exists := im.watchers[event.Key]; exists {
		for _, callback := range callbacks {
			go callback(event)
		}
	}
}

func (im *InMemoryStateStore) leaseCleanup() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		im.mu.Lock()
		now := time.Now()
		for leaseID, entry := range im.leases {
			if now.After(entry.expireAt) {
				delete(im.data, entry.key)
				delete(im.leases, leaseID)
			}
		}
		im.mu.Unlock()
	}
}

func (im *InMemoryStateStore) Close() error {
	return nil
}
