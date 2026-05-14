// Package resourcelock — LeaseLock is the production lock adapter
// over a transactional key-value store (etcd, Consul, Postgres-with-
// SELECT-FOR-UPDATE).  This file defines the wiring; the Backend
// interface abstracts the store so the same LeaseLock works against
// any of them.
package resourcelock

import (
	"context"
	"encoding/json"
	"fmt"
)

// Backend is the minimal KV contract LeaseLock needs.  Revisions are
// monotonically-increasing per-key versions — etcd's mod_revision,
// Postgres's xmin, etc.  Callers fetch revision from Get and pass it
// back to Put as the compare-and-set precondition.
type Backend interface {
	// Get returns (value, revision, exists, err).
	Get(ctx context.Context, key string) ([]byte, int64, bool, error)
	// Put writes value at key only if the current revision matches
	// ifRevision.  ifRevision == 0 requires the key to not exist.
	// Returns ErrConflict-wrapping if the precondition fails.
	Put(ctx context.Context, key string, value []byte, ifRevision int64) error
}

// LeaseLock serialises a LeaderElectionRecord to JSON and stores it
// at Key in the provided Backend.
type LeaseLock struct {
	Key     string
	ID      string
	Backend Backend

	// lastRevision is remembered between Get and Update so Update
	// can supply the precondition.  Access is not guarded because
	// leader-election drives a single goroutine per LeaseLock.
	lastRevision int64
}

// Get fetches the record.
func (l *LeaseLock) Get(ctx context.Context) (*LeaderElectionRecord, error) {
	raw, rev, ok, err := l.Backend.Get(ctx, l.Key)
	if err != nil {
		return nil, fmt.Errorf("leaselock get: %w", err)
	}
	if !ok {
		l.lastRevision = 0
		return nil, ErrNotFound
	}
	var rec LeaderElectionRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, fmt.Errorf("leaselock decode: %w", err)
	}
	l.lastRevision = rev
	return &rec, nil
}

// Create writes the initial record with the precondition "key must
// not exist" (ifRevision=0).
func (l *LeaseLock) Create(ctx context.Context, r LeaderElectionRecord) error {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("leaselock encode: %w", err)
	}
	if err := l.Backend.Put(ctx, l.Key, data, 0); err != nil {
		return err
	}
	// A successful create means revision is now 1 (or whatever the
	// backend chose); subsequent Update calls need a real Get to
	// refresh lastRevision.  Leave it at 0 to force that Get.
	l.lastRevision = 0
	return nil
}

// Update writes the record with the precondition "revision ==
// lastRevision from prior Get".  Callers that did not Get first see
// ErrConflict because lastRevision is 0 and the key already exists.
func (l *LeaseLock) Update(ctx context.Context, r LeaderElectionRecord) error {
	if l.lastRevision == 0 {
		return fmt.Errorf("leaselock update: must Get before Update")
	}
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("leaselock encode: %w", err)
	}
	return l.Backend.Put(ctx, l.Key, data, l.lastRevision)
}

// Identity returns the caller's ID.
func (l *LeaseLock) Identity() string { return l.ID }

// Describe returns "lease:<key>".
func (l *LeaseLock) Describe() string { return "lease:" + l.Key }
