package handlers

// Phase 6 P2 — Analytics resource-ification.

import (
	"context"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/reconciler"
	"example.com/axiomnizam/internal/resources"
)

const (
	AnalyticsDashboardKind       = "AnalyticsDashboard"
	AnalyticsDashboardAPIVersion = "analytics.axiomnizam.io/v1"
)

type AnalyticsDashboardSpec struct {
	DisplayName string            `json:"displayName"`
	Description string            `json:"description,omitempty"`
	Category    string            `json:"category,omitempty"`
	Widgets     []AnalyticsWidget `json:"widgets,omitempty"`
	Filters     []DashboardFilter `json:"filters,omitempty"`
}

type AnalyticsDashboardResourceStatus struct {
	resources.ObjectStatus `json:",inline"`
	WidgetCount            int `json:"widgetCount"`
}

type AnalyticsDashboardResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`
	Spec                 AnalyticsDashboardSpec           `json:"spec"`
	Status               AnalyticsDashboardResourceStatus `json:"status"`
}

func (r *AnalyticsDashboardResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *AnalyticsDashboardResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *AnalyticsDashboardResource) GetStatus() *resources.ObjectStatus {
	return &r.Status.ObjectStatus
}
func (r *AnalyticsDashboardResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *AnalyticsDashboardResource) DeepCopy() resources.Resource { cp := *r; return &cp }
func (r *AnalyticsDashboardResource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *AnalyticsDashboardResource) GetGeneration() int64         { return r.Generation }
func (r *AnalyticsDashboardResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }

// AnalyticsDashboardReconciler reconciles AnalyticsDashboardResource.
type AnalyticsDashboardReconciler struct {
	store store.ResourceStore[*AnalyticsDashboardResource]
}

func NewAnalyticsDashboardReconciler(rs store.ResourceStore[*AnalyticsDashboardResource]) *AnalyticsDashboardReconciler {
	return &AnalyticsDashboardReconciler{store: rs}
}

func (r *AnalyticsDashboardReconciler) Reconcile(ctx context.Context, obj reconciler.Resource) reconciler.ReconcileResult {
	res, ok := obj.(*AnalyticsDashboardResource)
	if !ok {
		return reconciler.ReconcileResult{Error: analyticsErr("analytics: wrong type")}
	}
	now := time.Now()
	res.Status.Phase = "Active"
	res.Status.WidgetCount = len(res.Spec.Widgets)
	res.Status.ObservedGeneration = res.Generation
	res.Status.LastTransitionTime = now
	if r.store != nil {
		_ = r.store.Update(ctx, res)
	}
	return reconciler.ReconcileResult{}
}

type analyticsErr string

func (e analyticsErr) Error() string { return string(e) }
