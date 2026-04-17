package reconciler

import (
	"context"
	"time"
)

// ============================================================================
// Reconciler Unification (P0.1)
// ============================================================================
//
// The canonical Reconciler contract in AxiomNizam is:
//
//     Reconcile(ctx context.Context, obj Resource) ReconcileResult
//
// Historically several parallel Reconciler interfaces existed across the
// codebase (controllers.Reconciler, client.UnifiedReconciler,
// controlplane.Reconciler, storage/controller.BucketController,
// apiresource.APIResourceReconciler). The helpers in this file convert those
// legacy signatures into the canonical Reconciler so that the work-queue,
// runtime, and controller-manager can drive any implementation uniformly.
//
// NEW CODE MUST IMPLEMENT THE CANONICAL INTERFACE DIRECTLY.
// The adapters below exist only so legacy controllers keep working during
// the P0 → P1 migration.

// ReconcilerFunc lets a plain function satisfy Reconciler.
type ReconcilerFunc func(ctx context.Context, obj Resource) ReconcileResult

// Reconcile implements Reconciler for the function type.
func (f ReconcilerFunc) Reconcile(ctx context.Context, obj Resource) ReconcileResult {
	return f(ctx, obj)
}

// KeyedReconciler is a legacy reconciler addressed by an opaque string key
// (e.g. "namespace/name"). This matches controllers.ControllerReconciler and
// client.UnifiedReconciler-style signatures.
type KeyedReconciler interface {
	Reconcile(ctx context.Context, key string) (time.Duration, error)
}

// FromKeyed adapts a KeyedReconciler into the canonical Reconciler interface.
// The adapter forwards obj.GetKey() to the keyed implementation and maps
// (time.Duration, error) to ReconcileResult.
func FromKeyed(k KeyedReconciler) Reconciler {
	return ReconcilerFunc(func(ctx context.Context, obj Resource) ReconcileResult {
		requeueAfter, err := k.Reconcile(ctx, obj.GetKey())
		return ReconcileResult{
			Requeue:      requeueAfter > 0 || err != nil,
			RequeueAfter: requeueAfter,
			Error:        err,
		}
	})
}

// TypedReconciler is a legacy reconciler that operates on a concrete resource
// type T (e.g. storage/controller.BucketController.Reconcile(ctx, *BucketResource)).
type TypedReconciler[T Resource] interface {
	Reconcile(ctx context.Context, obj T) error
}

// FromTyped adapts a typed, error-returning reconciler into the canonical
// Reconciler interface. Non-nil errors trigger a requeue with exponential
// backoff handled by the workqueue layer.
func FromTyped[T Resource](r TypedReconciler[T]) Reconciler {
	return ReconcilerFunc(func(ctx context.Context, obj Resource) ReconcileResult {
		typed, ok := obj.(T)
		if !ok {
			var zero T
			return ReconcileResult{
				Error: errTypeMismatch{want: anyTypeName(zero), got: anyTypeName(obj)},
			}
		}
		if err := r.Reconcile(ctx, typed); err != nil {
			return ReconcileResult{Requeue: true, Error: err}
		}
		return ReconcileResult{}
	})
}

// errTypeMismatch is returned when a typed adapter receives an unexpected
// concrete type.
type errTypeMismatch struct {
	want string
	got  string
}

func (e errTypeMismatch) Error() string {
	return "reconciler: type mismatch (want " + e.want + ", got " + e.got + ")"
}

// anyTypeName returns a best-effort name for diagnostic messages.
func anyTypeName(v any) string {
	if v == nil {
		return "<nil>"
	}
	type named interface{ GetTypeMetaName() string }
	if n, ok := v.(named); ok {
		return n.GetTypeMetaName()
	}
	return "unknown"
}
