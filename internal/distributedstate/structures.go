package distributedstate

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type DistributedValue struct {
	store   StateStore
	key     string
	mu      sync.RWMutex
	lastRev int64
}

func NewDistributedValue(store StateStore, key string) *DistributedValue {
	return &DistributedValue{
		store: store,
		key:   key,
	}
}

func (dv *DistributedValue) Get(ctx context.Context) (string, error) {
	return dv.store.Get(ctx, dv.key)
}

func (dv *DistributedValue) Set(ctx context.Context, value string) error {
	return dv.store.Put(ctx, dv.key, value)
}

func (dv *DistributedValue) GetJSON(ctx context.Context, v interface{}) error {
	value, err := dv.store.Get(ctx, dv.key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), v)
}

func (dv *DistributedValue) SetJSON(ctx context.Context, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return dv.store.Put(ctx, dv.key, string(data))
}

type DistributedCounter struct {
	value *DistributedValue
	store StateStore
	key   string
	mu    sync.Mutex
}

func NewDistributedCounter(store StateStore, key string) *DistributedCounter {
	return &DistributedCounter{
		value: NewDistributedValue(store, key),
		store: store,
		key:   key,
	}
}

func (dc *DistributedCounter) Increment(ctx context.Context) (int64, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		current, err := dc.value.Get(ctx)
		if err != nil {
			return 0, err
		}

		var currentVal int64 = 0
		if current != "" {
			fmt.Sscanf(current, "%d", &currentVal)
		}

		newVal := currentVal + 1
		success, err := dc.store.CompareAndSwap(ctx, dc.key, current, fmt.Sprintf("%d", newVal))
		if err != nil {
			return 0, err
		}

		if success {
			return newVal, nil
		}

		time.Sleep(time.Millisecond * time.Duration(1+(i*10)))
	}

	return 0, fmt.Errorf("failed to increment counter after %d retries", maxRetries)
}

func (dc *DistributedCounter) Decrement(ctx context.Context) (int64, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		current, err := dc.value.Get(ctx)
		if err != nil {
			return 0, err
		}

		var currentVal int64 = 0
		if current != "" {
			fmt.Sscanf(current, "%d", &currentVal)
		}

		newVal := currentVal - 1
		success, err := dc.store.CompareAndSwap(ctx, dc.key, current, fmt.Sprintf("%d", newVal))
		if err != nil {
			return 0, err
		}

		if success {
			return newVal, nil
		}

		time.Sleep(time.Millisecond * time.Duration(1+(i*10)))
	}

	return 0, fmt.Errorf("failed to decrement counter after %d retries", maxRetries)
}

func (dc *DistributedCounter) Get(ctx context.Context) (int64, error) {
	value, err := dc.value.Get(ctx)
	if err != nil {
		return 0, err
	}

	var val int64 = 0
	if value != "" {
		fmt.Sscanf(value, "%d", &val)
	}
	return val, nil
}

type DistributedSet struct {
	store  StateStore
	prefix string
}

func NewDistributedSet(store StateStore, prefix string) *DistributedSet {
	return &DistributedSet{
		store:  store,
		prefix: prefix,
	}
}

func (ds *DistributedSet) Add(ctx context.Context, member string) error {
	key := fmt.Sprintf("%s/%s", ds.prefix, member)
	return ds.store.Put(ctx, key, "1")
}

func (ds *DistributedSet) Remove(ctx context.Context, member string) error {
	key := fmt.Sprintf("%s/%s", ds.prefix, member)
	return ds.store.Delete(ctx, key)
}

func (ds *DistributedSet) Contains(ctx context.Context, member string) (bool, error) {
	key := fmt.Sprintf("%s/%s", ds.prefix, member)
	value, err := ds.store.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return value == "1", nil
}

func (ds *DistributedSet) Members(ctx context.Context) ([]string, error) {
	items, err := ds.store.List(ctx, ds.prefix)
	if err != nil {
		return nil, err
	}

	members := make([]string, 0, len(items))
	for key := range items {
		members = append(members, key)
	}
	return members, nil
}

func (ds *DistributedSet) Size(ctx context.Context) (int, error) {
	members, err := ds.Members(ctx)
	if err != nil {
		return 0, err
	}
	return len(members), nil
}

type DistributedQueue struct {
	store  StateStore
	prefix string
	mu     sync.Mutex
	index  int64
}

func NewDistributedQueue(store StateStore, prefix string) *DistributedQueue {
	return &DistributedQueue{
		store:  store,
		prefix: prefix,
		index:  0,
	}
}

func (dq *DistributedQueue) Enqueue(ctx context.Context, value string) error {
	dq.mu.Lock()
	dq.index++
	index := dq.index
	dq.mu.Unlock()

	key := fmt.Sprintf("%s/%d", dq.prefix, index)
	return dq.store.Put(ctx, key, value)
}

func (dq *DistributedQueue) Dequeue(ctx context.Context) (string, error) {
	items, err := dq.store.List(ctx, dq.prefix)
	if err != nil {
		return "", err
	}

	if len(items) == 0 {
		return "", fmt.Errorf("queue is empty")
	}

	var minKey string
	for key := range items {
		if minKey == "" || key < minKey {
			minKey = key
		}
	}

	value := items[minKey]
	if err := dq.store.Delete(ctx, minKey); err != nil {
		return "", err
	}

	return value, nil
}

func (dq *DistributedQueue) Size(ctx context.Context) (int, error) {
	items, err := dq.store.List(ctx, dq.prefix)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}
