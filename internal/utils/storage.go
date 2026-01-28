package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StorageBackend defines a storage interface
type StorageBackend interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Put(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) (map[string]interface{}, error)
	Watch(ctx context.Context, prefix string) (<-chan StorageEvent, error)
}

// StorageEvent represents a storage change
type StorageEvent struct {
	Type     string // PUT, DELETE
	Key      string
	Value    interface{}
	Revision int64
}

// ResourceSyncManager manages resource synchronization (like etcd sync)
type ResourceSyncManager struct {
	mu               sync.RWMutex
	backends         map[string]StorageBackend
	resourceVersions map[string]int64 // resource -> version
	lastSyncTime     time.Time
	resyncPeriod     time.Duration
	conflictResolver func(local, remote interface{}) interface{}
}

// NewResourceSyncManager creates a new sync manager
func NewResourceSyncManager(resyncPeriod time.Duration) *ResourceSyncManager {
	return &ResourceSyncManager{
		backends:         make(map[string]StorageBackend),
		resourceVersions: make(map[string]int64),
		resyncPeriod:     resyncPeriod,
		lastSyncTime:     time.Now(),
		conflictResolver: func(local, remote interface{}) interface{} {
			return remote // Default: server wins
		},
	}
}

// RegisterBackend registers a storage backend
func (rsm *ResourceSyncManager) RegisterBackend(name string, backend StorageBackend) {
	rsm.mu.Lock()
	defer rsm.mu.Unlock()

	rsm.backends[name] = backend
}

// SetConflictResolver sets a custom conflict resolver
func (rsm *ResourceSyncManager) SetConflictResolver(resolver func(local, remote interface{}) interface{}) {
	rsm.mu.Lock()
	defer rsm.mu.Unlock()

	rsm.conflictResolver = resolver
}

// Sync performs full synchronization with backends
func (rsm *ResourceSyncManager) Sync(ctx context.Context) error {
	rsm.mu.Lock()
	defer rsm.mu.Unlock()

	for name, backend := range rsm.backends {
		resources, err := backend.List(ctx, "")
		if err != nil {
			return fmt.Errorf("failed to sync %s: %w", name, err)
		}

		for key := range resources {
			rsm.resourceVersions[key]++
		}
	}

	rsm.lastSyncTime = time.Now()
	return nil
}

// GetResourceVersion returns the version of a resource
func (rsm *ResourceSyncManager) GetResourceVersion(resourceID string) int64 {
	rsm.mu.RLock()
	defer rsm.mu.RUnlock()

	return rsm.resourceVersions[resourceID]
}

// NeedsResync checks if resync is needed
func (rsm *ResourceSyncManager) NeedsResync() bool {
	rsm.mu.RLock()
	defer rsm.mu.RUnlock()

	return time.Since(rsm.lastSyncTime) > rsm.resyncPeriod
}

// TransactionManager manages atomic transactions (ACID-like guarantees)
type Transaction struct {
	ID         string
	Operations []TransactionOp
	Status     string // pending, committed, aborted
	StartTime  time.Time
	EndTime    time.Time
}

// TransactionOp represents a transaction operation
type TransactionOp struct {
	Type     string // PUT, DELETE
	Key      string
	Value    interface{}
	Previous interface{}
}

// TransactionManager manages transactions
type TransactionManager struct {
	mu           sync.RWMutex
	transactions map[string]*Transaction
	nextTxID     int
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager() *TransactionManager {
	return &TransactionManager{
		transactions: make(map[string]*Transaction),
	}
}

// BeginTransaction starts a new transaction
func (tm *TransactionManager) BeginTransaction() *Transaction {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.nextTxID++
	txID := fmt.Sprintf("tx-%d", tm.nextTxID)

	tx := &Transaction{
		ID:         txID,
		Operations: make([]TransactionOp, 0),
		Status:     "pending",
		StartTime:  time.Now(),
	}

	tm.transactions[txID] = tx
	return tx
}

// AddOperation adds an operation to a transaction
func (tm *TransactionManager) AddOperation(txID string, op TransactionOp) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx, ok := tm.transactions[txID]
	if !ok {
		return fmt.Errorf("transaction not found")
	}

	if tx.Status != "pending" {
		return fmt.Errorf("transaction already %s", tx.Status)
	}

	tx.Operations = append(tx.Operations, op)
	return nil
}

// CommitTransaction commits a transaction
func (tm *TransactionManager) CommitTransaction(txID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx, ok := tm.transactions[txID]
	if !ok {
		return fmt.Errorf("transaction not found")
	}

	if tx.Status != "pending" {
		return fmt.Errorf("cannot commit %s transaction", tx.Status)
	}

	tx.Status = "committed"
	tx.EndTime = time.Now()
	return nil
}

// RollbackTransaction rolls back a transaction
func (tm *TransactionManager) RollbackTransaction(txID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx, ok := tm.transactions[txID]
	if !ok {
		return fmt.Errorf("transaction not found")
	}

	if tx.Status != "pending" {
		return fmt.Errorf("cannot rollback %s transaction", tx.Status)
	}

	tx.Status = "aborted"
	tx.EndTime = time.Now()
	return nil
}

// GetTransaction returns transaction info
func (tm *TransactionManager) GetTransaction(txID string) *Transaction {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.transactions[txID]
}

// ConsistencyLevel defines consistency guarantees
type ConsistencyLevel int

const (
	Eventual ConsistencyLevel = iota
	StrongConsistency
	Sequential
	LinearizableRead
)

// Snapshot represents a point-in-time resource snapshot
type Snapshot struct {
	Name          string
	CreatedAt     time.Time
	ExpiresAt     time.Time
	Data          map[string]interface{}
	ResourceCount int
	Metadata      map[string]string
}

// SnapshotManager manages resource snapshots
type SnapshotManager struct {
	mu        sync.RWMutex
	snapshots map[string]*Snapshot
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager() *SnapshotManager {
	return &SnapshotManager{
		snapshots: make(map[string]*Snapshot),
	}
}

// CreateSnapshot creates a snapshot
func (sm *SnapshotManager) CreateSnapshot(name string, data map[string]interface{}, ttl time.Duration) *Snapshot {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshot := &Snapshot{
		Name:          name,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(ttl),
		Data:          data,
		ResourceCount: len(data),
		Metadata:      make(map[string]string),
	}

	sm.snapshots[name] = snapshot
	return snapshot
}

// GetSnapshot retrieves a snapshot
func (sm *SnapshotManager) GetSnapshot(name string) *Snapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	snapshot := sm.snapshots[name]
	if snapshot != nil && snapshot.ExpiresAt.Before(time.Now()) {
		return nil // Snapshot expired
	}
	return snapshot
}

// ListSnapshots lists all snapshots
func (sm *SnapshotManager) ListSnapshots() []*Snapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	snapshots := make([]*Snapshot, 0)
	now := time.Now()
	for _, snapshot := range sm.snapshots {
		if snapshot.ExpiresAt.After(now) {
			snapshots = append(snapshots, snapshot)
		}
	}
	return snapshots
}

// RestoreSnapshot restores from a snapshot
func (sm *SnapshotManager) RestoreSnapshot(name string) (map[string]interface{}, error) {
	snapshot := sm.GetSnapshot(name)
	if snapshot == nil {
		return nil, fmt.Errorf("snapshot not found or expired")
	}

	data := make(map[string]interface{})
	for k, v := range snapshot.Data {
		data[k] = v
	}
	return data, nil
}

// DeleteSnapshot deletes a snapshot
func (sm *SnapshotManager) DeleteSnapshot(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.snapshots[name]; !ok {
		return fmt.Errorf("snapshot not found")
	}

	delete(sm.snapshots, name)
	return nil
}

// CleanupExpiredSnapshots removes expired snapshots
func (sm *SnapshotManager) CleanupExpiredSnapshots() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	count := 0

	for name, snapshot := range sm.snapshots {
		if snapshot.ExpiresAt.Before(now) {
			delete(sm.snapshots, name)
			count++
		}
	}

	return count
}
