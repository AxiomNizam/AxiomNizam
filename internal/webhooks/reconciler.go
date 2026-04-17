package webhooks

// Reconciler for WebhookResource.
//
// Skeleton reconciler that treats Spec.Active as the sole runtime
// state: when Active=true the resource transitions to Ready, otherwise
// to Paused.  A production follow-up plugs in the actual webhook
// dispatch bookkeeping (delivery counters, retry schedule, circuit
// breaker state).

import (
	"context"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/reconciler"
)

// WebhookReconciler reconciles WebhookResource objects.
type WebhookReconciler struct {
	store store.ResourceStore[*WebhookResource]
}

// NewWebhookReconciler builds a reconciler.
func NewWebhookReconciler(rs store.ResourceStore[*WebhookResource]) *WebhookReconciler {
	return &WebhookReconciler{store: rs}
}

// Reconcile implements `reconciler.Reconciler`.
func (r *WebhookReconciler) Reconcile(ctx context.Context, obj reconciler.Resource) reconciler.ReconcileResult {
	res, ok := obj.(*WebhookResource)
	if !ok {
		return reconciler.ReconcileResult{Error: errWebhookType}
	}

	phase := "Paused"
	if res.Spec.Active {
		phase = "Ready"
	}
	now := time.Now()
	status := res.Status
	status.Phase = phase
	status.ObservedGeneration = res.Generation
	status.LastTransitionTime = now
	res.Status = status

	if r.store != nil {
		_ = r.store.Update(ctx, res)
	}
	return reconciler.ReconcileResult{}
}

var errWebhookType = webhookErr("webhooks: reconciler received non-WebhookResource")

type webhookErr string

func (e webhookErr) Error() string { return string(e) }
