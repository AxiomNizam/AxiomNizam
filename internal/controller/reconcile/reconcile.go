// Package reconcile defines the controller-runtime reconciler
// contract used across AxiomNizam controllers.  A Reconciler receives
// a Request (namespace/name of the object that changed), performs
// whatever convergence logic the controller owns, and returns a
// Result telling the workqueue whether to requeue.
//
// The surface is deliberately kept to three types: Reconciler,
// Request, Result.  Controllers that need richer context (for
// example, an Object already fetched) wrap this Reconciler with a
// decorator — we do not push that into the interface because the
// reconciler-per-GVK model does not always want it.
package reconcile

import (
	"context"
	"time"
)

// Request identifies the object whose observed state may have drifted
// from its desired state.  It carries only the namespace/name tuple
// because the controller already knows which GVK it is reconciling.
type Request struct {
	// Namespace is empty for cluster-scoped resources.
	Namespace string
	// Name is the metadata.name of the object.
	Name string
}

// String renders the canonical "namespace/name" form used in logs.
// Cluster-scoped resources render as just the name.
func (r Request) String() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}

// Result tells the workqueue what to do after the reconcile returns.
// The zero value means "don't requeue" — the most common outcome
// when the observed state already matches desired state.
type Result struct {
	// Requeue=true re-adds the request to the queue with standard
	// rate-limiting.  Used when the controller made progress but has
	// more work to do (for example, waiting for a dependent object
	// to become Ready).
	Requeue bool
	// RequeueAfter scheduled the request for re-processing after
	// this duration.  Useful for TTL-driven logic (expiring certs,
	// cleaning up orphaned children).  Non-zero value implies Requeue.
	RequeueAfter time.Duration
}

// IsZero reports whether the caller requested no follow-up.  Used by
// the controller loop to decide whether to call the queue at all.
func (r Result) IsZero() bool { return !r.Requeue && r.RequeueAfter == 0 }

// Reconciler is the core controller contract.  Implementations MUST
// be idempotent: Reconcile may be invoked many times with the same
// Request, in any order, with arbitrary delays between calls.
type Reconciler interface {
	// Reconcile converges the cluster toward the object's desired
	// state.  If Reconcile returns an error, the workqueue re-adds
	// the Request with exponential backoff; if it returns a non-zero
	// Result with a zero error, the queue honours the Result's
	// requeue preference without backoff.
	Reconcile(ctx context.Context, req Request) (Result, error)
}

// Func adapts a plain function to the Reconciler interface — useful
// for tests and simple controllers that do not need to carry state
// on a struct receiver.
type Func func(ctx context.Context, req Request) (Result, error)

// Reconcile satisfies Reconciler.
func (f Func) Reconcile(ctx context.Context, req Request) (Result, error) {
	return f(ctx, req)
}
