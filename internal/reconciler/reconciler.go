package reconciler

import (
	"context"
	"time"
)

// Resource defines the interface for any reconcilable resource
type Resource interface {
	// GetKey returns a unique identifier for the resource (namespace/name)
	GetKey() string

	// GetGeneration returns the generation of the spec
	GetGeneration() int64

	// GetObservedGeneration returns the generation observed by the controller
	GetObservedGeneration() int64
}

// ReconcileResult holds the result of a reconciliation attempt
type ReconcileResult struct {
	// Requeue indicates if the item should be requeued
	Requeue bool

	// RequeueAfter specifies duration to requeue after (Requeue must be true)
	RequeueAfter time.Duration

	// Error is any error encountered during reconciliation
	Error error
}

// Reconciler defines the standard reconciliation interface
// All controllers must implement this pattern:
// 1. Observe: Read current state from storage and runtime
// 2. Diff: Compare desired vs actual state
// 3. Act: Make changes to achieve desired state
// 4. Update Status: Record the result in storage
type Reconciler interface {
	// Reconcile implements the standard Observe → Diff → Act → Update Status pattern
	//
	// Input: ctx (context with deadline for cancellation)
	//        obj (the resource to reconcile)
	//
	// Process:
	//   1. Observe: Fetch desired state from obj and actual state from runtime
	//   2. Diff: Identify what changed between desired and actual
	//   3. Act: Make changes to achieve desired state (idempotent)
	//   4. Update Status: Record phase, ready, conditions in storage
	//
	// Output: ReconcileResult with requeue decision and error
	//
	// Contract:
	//   - Must be idempotent (safe to call multiple times)
	//   - Must update status even on error
	//   - Must use context deadline for timeout
	//   - Must return proper requeue decisions
	Reconcile(ctx context.Context, obj Resource) ReconcileResult
}

// Observer observes the current state of a resource
type Observer interface {
	// ObserveDesiredState reads the desired state from storage
	ObserveDesiredState(ctx context.Context, key string) (Resource, error)

	// ObserveActualState reads the actual state from the runtime
	ObserveActualState(ctx context.Context, key string) (map[string]interface{}, error)
}

// Differ compares desired vs actual state
type Differ interface {
	// Diff returns the differences between desired and actual
	// Returns a list of changes needed to be made
	Diff(ctx context.Context, desired Resource, actual map[string]interface{}) []string
}

// Actor performs changes to achieve desired state
type Actor interface {
	// Act executes changes to move toward desired state
	// Must be idempotent
	Act(ctx context.Context, obj Resource, changes []string) error
}

// StatusUpdater updates resource status in storage
type StatusUpdater interface {
	// UpdateStatus updates the resource status and conditions
	UpdateStatus(ctx context.Context, obj Resource) error
}

// StandardReconciler provides a base implementation of the Reconciler interface
// Subclasses should override Diff and Act methods
type StandardReconciler struct {
	Observer      Observer
	Differ        Differ
	Actor         Actor
	StatusUpdater StatusUpdater
}

// Reconcile implements the standard pattern
func (r *StandardReconciler) Reconcile(ctx context.Context, obj Resource) ReconcileResult {
	// Phase 1: OBSERVE
	// Read desired state from storage
	desired, err := r.Observer.ObserveDesiredState(ctx, obj.GetKey())
	if err != nil {
		return ReconcileResult{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
			Error:        err,
		}
	}

	// Read actual state from runtime
	actual, err := r.Observer.ObserveActualState(ctx, obj.GetKey())
	if err != nil {
		return ReconcileResult{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
			Error:        err,
		}
	}

	// Phase 2: DIFF
	// Identify what changed
	changes := r.Differ.Diff(ctx, desired, actual)

	// Phase 3: ACT
	// Make changes if needed (idempotent)
	if len(changes) > 0 {
		err := r.Actor.Act(ctx, desired, changes)
		if err != nil {
			// Update status to reflect error
			r.StatusUpdater.UpdateStatus(ctx, desired)
			return ReconcileResult{
				Requeue:      true,
				RequeueAfter: 10 * time.Second,
				Error:        err,
			}
		}
	}

	// Phase 4: UPDATE STATUS
	// Record the result
	err = r.StatusUpdater.UpdateStatus(ctx, desired)
	if err != nil {
		return ReconcileResult{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
			Error:        err,
		}
	}

	// Success - no requeue needed
	return ReconcileResult{
		Requeue: false,
		Error:   nil,
	}
}

// Helper function to create a simple reconciler from components
func New(observer Observer, differ Differ, actor Actor, updater StatusUpdater) Reconciler {
	return &StandardReconciler{
		Observer:      observer,
		Differ:        differ,
		Actor:         actor,
		StatusUpdater: updater,
	}
}
