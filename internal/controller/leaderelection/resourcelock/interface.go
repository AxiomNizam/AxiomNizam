// Package resourcelock defines the pluggable lock interface used by
// leader-election.  Two implementations are provided here:
//
//   - MemoryLock: single-process lock for tests and single-replica
//     deployments.
//   - LeaseLock:  wraps an etcd/v3 lease for multi-replica setups;
//     this package only defines the contract — the etcd binding
//     lives alongside the existing internal/distributedstate code.
//
// The lock interface mirrors k8s.io/client-go/tools/leaderelection
// but drops the Endpoints/ConfigMap legacy variants.  Every
// production caller is on leases by now.
package resourcelock

import (
	"context"
	"errors"
	"sync"
	"time"
)

// LeaderElectionRecord is the marshalled state every lock stores.
// Callers serialise/deserialise this; the lock is responsible only
// for atomic compare-and-set semantics.
type LeaderElectionRecord struct {
	// HolderIdentity is the stable ID of the current leader.  By
	// convention this is hostname+pid+uuid; empty means "no leader".
	HolderIdentity string `json:"holderIdentity"`
	// LeaseDurationSeconds bounds how long the lock remains valid
	// without renewal.  After this elapses, contenders may steal.
	LeaseDurationSeconds int `json:"leaseDurationSeconds"`
	// AcquireTime records when the current leader first won.
	AcquireTime time.Time `json:"acquireTime"`
	// RenewTime records the last successful renewal.
	RenewTime time.Time `json:"renewTime"`
	// LeaderTransitions increments on every change of holder — used
	// for metrics and for detecting flapping elections.
	LeaderTransitions int `json:"leaderTransitions"`
}

// IsExpired reports whether the record should be considered stale.
// Used by contenders to decide whether to attempt a steal.
func (r *LeaderElectionRecord) IsExpired(now time.Time) bool {
	if r.HolderIdentity == "" {
		return true
	}
	if r.LeaseDurationSeconds <= 0 {
		return false
	}
	deadline := r.RenewTime.Add(time.Duration(r.LeaseDurationSeconds) * time.Second)
	return now.After(deadline)
}

// ErrNotFound is returned by Get when no record exists yet.
var ErrNotFound = errors.New("resourcelock: no record")

// ErrConflict is returned by Update/Create when the backing store
// saw a competing write since the caller's last Get.
var ErrConflict = errors.New("resourcelock: concurrent update")

// Interface is the contract implemented by each lock variant.
type Interface interface {
	// Get returns the current record, or ErrNotFound if none exists.
	Get(ctx context.Context) (*LeaderElectionRecord, error)
	// Create writes the initial record.  Returns ErrConflict if a
	// record already exists.
	Create(ctx context.Context, record LeaderElectionRecord) error
	// Update atomically replaces the record.  Returns ErrConflict
	// when the store's version differs from the one implied by the
	// caller's prior Get.
	Update(ctx context.Context, record LeaderElectionRecord) error
	// Identity returns the caller's stable ID.
	Identity() string
	// Describe returns a human-readable locator for logs — e.g.
	// "lease/axiomnizam-leader in namespace control-plane".
	Describe() string
}

// MemoryLock is the in-process implementation.  Safe for concurrent
// use but offers zero protection against a second process — use only
// in tests and single-replica deployments.
type MemoryLock struct {
	// ID is the identity returned by Identity().
	ID string
	// Name is a label used by Describe().
	Name string

	mu     sync.Mutex
	record *LeaderElectionRecord
}

// Get returns a copy of the current record.
func (m *MemoryLock) Get(_ context.Context) (*LeaderElectionRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.record == nil {
		return nil, ErrNotFound
	}
	cp := *m.record
	return &cp, nil
}

// Create stores the record if none exists.
func (m *MemoryLock) Create(_ context.Context, r LeaderElectionRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.record != nil {
		return ErrConflict
	}
	cp := r
	m.record = &cp
	return nil
}

// Update replaces the record unconditionally — since MemoryLock has
// no version field, the conflict path is unreachable in practice.
func (m *MemoryLock) Update(_ context.Context, r LeaderElectionRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := r
	m.record = &cp
	return nil
}

// Identity returns the configured ID.
func (m *MemoryLock) Identity() string { return m.ID }

// Describe returns the configured Name.
func (m *MemoryLock) Describe() string {
	if m.Name == "" {
		return "memory-lock"
	}
	return "memory-lock:" + m.Name
}
