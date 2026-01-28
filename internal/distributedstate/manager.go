package distributedstate

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type DistributedManager struct {
	store   StateStore
	mu      sync.RWMutex
	prefix  string
	timeout time.Duration
}

func NewDistributedManager(store StateStore, prefix string) *DistributedManager {
	return &DistributedManager{
		store:   store,
		prefix:  prefix,
		timeout: 5 * time.Second,
	}
}

func (dm *DistributedManager) getKey(name string) string {
	return fmt.Sprintf("%s/%s", dm.prefix, name)
}

func (dm *DistributedManager) WithTimeout(timeout time.Duration) *DistributedManager {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.timeout = timeout
	return dm
}

func (dm *DistributedManager) PutState(ctx context.Context, name string, value string) error {
	key := dm.getKey(name)
	ctx, cancel := context.WithTimeout(ctx, dm.timeout)
	defer cancel()
	return dm.store.Put(ctx, key, value)
}

func (dm *DistributedManager) GetState(ctx context.Context, name string) (string, error) {
	key := dm.getKey(name)
	ctx, cancel := context.WithTimeout(ctx, dm.timeout)
	defer cancel()
	return dm.store.Get(ctx, key)
}

func (dm *DistributedManager) DeleteState(ctx context.Context, name string) error {
	key := dm.getKey(name)
	ctx, cancel := context.WithTimeout(ctx, dm.timeout)
	defer cancel()
	return dm.store.Delete(ctx, key)
}

func (dm *DistributedManager) UpdateStateIfUnchanged(ctx context.Context, name string, oldValue, newValue string) (bool, error) {
	key := dm.getKey(name)
	ctx, cancel := context.WithTimeout(ctx, dm.timeout)
	defer cancel()
	return dm.store.CompareAndSwap(ctx, key, oldValue, newValue)
}

func (dm *DistributedManager) GetStateWithVersion(ctx context.Context, name string) (string, int64, error) {
	key := dm.getKey(name)
	ctx, cancel := context.WithTimeout(ctx, dm.timeout)
	defer cancel()
	return dm.store.GetWithRevision(ctx, key)
}

func (dm *DistributedManager) RegisterStateWithTTL(ctx context.Context, name string, value string, ttl time.Duration) (int64, error) {
	key := dm.getKey(name)
	ctx, cancel := context.WithTimeout(ctx, dm.timeout)
	defer cancel()
	return dm.store.PutWithLease(ctx, key, value, ttl)
}

func (dm *DistributedManager) DeregisterState(ctx context.Context, leaseID int64) error {
	ctx, cancel := context.WithTimeout(ctx, dm.timeout)
	defer cancel()
	return dm.store.ReleaseLeaseID(ctx, leaseID)
}

func (dm *DistributedManager) WatchState(ctx context.Context, name string, handler func(Event)) error {
	key := dm.getKey(name)
	return dm.store.Watch(ctx, key, handler)
}

func (dm *DistributedManager) ListStates(ctx context.Context) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, dm.timeout)
	defer cancel()
	return dm.store.List(ctx, dm.prefix)
}

func (dm *DistributedManager) ExecuteTransaction(ctx context.Context, operations []Operation) (bool, error) {
	for i := range operations {
		operations[i].Key = dm.getKey(operations[i].Key)
	}
	ctx, cancel := context.WithTimeout(ctx, dm.timeout)
	defer cancel()
	return dm.store.Transaction(ctx, operations)
}

func (dm *DistributedManager) Close() error {
	return dm.store.Close()
}
