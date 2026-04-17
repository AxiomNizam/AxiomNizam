package apiscanner

// Reconciler for APIScanResource.
//
// Translates the resource Spec into a ScanRequest, drives the existing
// `Engine.Scan`, and records the result on Status.

import (
	"context"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/platform/timing"
	"example.com/axiomnizam/internal/reconciler"
)

// APIScanReconciler reconciles APIScanResource objects.
type APIScanReconciler struct {
	store  store.ResourceStore[*APIScanResource]
	engine *Engine
}

// NewAPIScanReconciler builds a reconciler.  `engine` must be non-nil
// for actual scans to run; if nil the reconciler only marks Phase.
func NewAPIScanReconciler(rs store.ResourceStore[*APIScanResource], eng *Engine) *APIScanReconciler {
	return &APIScanReconciler{store: rs, engine: eng}
}

// Reconcile implements `reconciler.Reconciler`.
func (r *APIScanReconciler) Reconcile(ctx context.Context, obj reconciler.Resource) reconciler.ReconcileResult {
	res, ok := obj.(*APIScanResource)
	if !ok {
		return reconciler.ReconcileResult{Error: scanErr("apiscanner: reconciler received non-APIScanResource")}
	}

	now := time.Now()
	status := res.Status
	status.ObservedGeneration = res.Generation
	status.LastTransitionTime = now

	if r.engine == nil {
		status.Phase = "Pending"
		res.Status = status
		if r.store != nil {
			_ = r.store.Update(ctx, res)
		}
		return reconciler.ReconcileResult{}
	}

	req := ScanRequest{
		Endpoint:           res.Spec.Endpoint,
		Timeout:            res.Spec.Timeout,
		RetryCount:         res.Spec.RetryCount,
		RetryBackoff:       res.Spec.RetryBackoff,
		InsecureSkipVerify: res.Spec.InsecureSkipVerify,
		AuthHeader:         res.Spec.AuthHeader,
		AuthValue:          res.Spec.AuthValue,
		Format:             res.Spec.Format,
	}
	result, err := r.engine.Scan(ctx, req)
	status.LastScanAt = &now
	status.ScanCount++
	if err != nil {
		status.Phase = "Failed"
		status.LastError = err.Error()
		res.Status = status
		if r.store != nil {
			_ = r.store.Update(ctx, res)
		}
		return reconciler.ReconcileResult{Error: err, Requeue: true, RequeueAfter: timing.DefaultRequeueAfter}
	}
	status.LastResult = &result
	status.LastError = ""
	if res.Spec.RunOnce {
		status.Phase = "Completed"
		res.Status = status
		if r.store != nil {
			_ = r.store.Update(ctx, res)
		}
		return reconciler.ReconcileResult{}
	}
	status.Phase = "Ready"
	res.Status = status
	if r.store != nil {
		_ = r.store.Update(ctx, res)
	}
	return reconciler.ReconcileResult{Requeue: true, RequeueAfter: timing.DefaultRequeueAfter}
}

type scanErr string

func (e scanErr) Error() string { return string(e) }
