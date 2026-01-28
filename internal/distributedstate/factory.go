package distributedstate

import (
	"context"
	"fmt"
)

type StateStoreFactory struct{}

func (ssf *StateStoreFactory) CreateEtcdStore(endpoints []string) (StateStore, error) {
	return NewEtcdStateStore(endpoints)
}

func (ssf *StateStoreFactory) CreateInMemoryStore() StateStore {
	return NewInMemoryStateStore()
}

type DistributedConfig struct {
	Mode      string
	Endpoints []string
	Prefix    string
	CacheTTL  int
}

func NewStateStoreFromConfig(cfg DistributedConfig) (StateStore, error) {
	factory := &StateStoreFactory{}

	switch cfg.Mode {
	case "etcd":
		if len(cfg.Endpoints) == 0 {
			return nil, fmt.Errorf("etcd endpoints required")
		}
		return factory.CreateEtcdStore(cfg.Endpoints)
	case "memory":
		return factory.CreateInMemoryStore(), nil
	default:
		return nil, fmt.Errorf("unknown state store mode: %s", cfg.Mode)
	}
}

func NewCachedManager(cfg DistributedConfig) (*DistributedManager, error) {
	store, err := NewStateStoreFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	var finalStore StateStore = store
	if cfg.CacheTTL > 0 {
		finalStore = NewCachedStateStore(store, 0)
	}

	return NewDistributedManager(finalStore, cfg.Prefix), nil
}

type BatchOperation struct {
	Key   string
	Value string
	Type  OperationType
}

func (dm *DistributedManager) BatchUpdate(ctx context.Context, operations []BatchOperation) (bool, error) {
	ops := make([]Operation, len(operations))
	for i, bop := range operations {
		ops[i] = Operation{
			Type:  bop.Type,
			Key:   bop.Key,
			Value: bop.Value,
		}
	}
	return dm.ExecuteTransaction(ctx, ops)
}

func (dm *DistributedManager) GetOrDefault(ctx context.Context, name string, defaultValue string) string {
	value, err := dm.GetState(ctx, name)
	if err != nil || value == "" {
		return defaultValue
	}
	return value
}

func (dm *DistributedManager) Exists(ctx context.Context, name string) (bool, error) {
	value, err := dm.GetState(ctx, name)
	if err != nil {
		return false, err
	}
	return value != "", nil
}

func (dm *DistributedManager) IncrementCounter(ctx context.Context, name string) (int64, error) {
	key := dm.getKey(name)
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		current, err := dm.GetState(ctx, name)
		if err != nil {
			return 0, err
		}

		var currentVal int64 = 0
		if current != "" {
			fmt.Sscanf(current, "%d", &currentVal)
		}

		newVal := currentVal + 1
		success, err := dm.store.CompareAndSwap(ctx, key, current, fmt.Sprintf("%d", newVal))
		if err != nil {
			return 0, err
		}

		if success {
			return newVal, nil
		}
	}

	return 0, fmt.Errorf("failed to increment counter after %d retries", maxRetries)
}

func (dm *DistributedManager) Replicate(ctx context.Context, sourcePrefix string, targetPrefix string) error {
	items, err := dm.store.List(ctx, sourcePrefix)
	if err != nil {
		return err
	}

	for key, value := range items {
		newKey := fmt.Sprintf("%s%s", targetPrefix, key[len(sourcePrefix):])
		if err := dm.store.Put(ctx, newKey, value); err != nil {
			return err
		}
	}

	return nil
}
