package netintel

// Phase 6 P2 — NetIntel resource-ification.

import (
	"context"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/reconciler"
	"example.com/axiomnizam/internal/resources"
)

const (
	ConfigKind       = "NetIntelConfig"
	ConfigAPIVersion = "netintel.axiomnizam.io/v1"
)

type ConfigSpec struct {
	TenantID string `json:"tenantId,omitempty"`
	Enabled  bool   `json:"enabled"`
}

type ConfigResourceStatus struct {
	resources.ObjectStatus `json:",inline"`
	ConfigActive           bool `json:"configActive"`
}

type ConfigResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`
	Spec                 ConfigSpec           `json:"spec"`
	Status               ConfigResourceStatus `json:"status"`
}

func (r *ConfigResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *ConfigResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *ConfigResource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *ConfigResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *ConfigResource) DeepCopy() resources.Resource { cp := *r; return &cp }
func (r *ConfigResource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *ConfigResource) GetGeneration() int64         { return r.Generation }
func (r *ConfigResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }

type ConfigReconciler struct {
	store store.ResourceStore[*ConfigResource]
}

func NewConfigReconciler(rs store.ResourceStore[*ConfigResource]) *ConfigReconciler {
	return &ConfigReconciler{store: rs}
}

func (rec *ConfigReconciler) Reconcile(ctx context.Context, obj reconciler.Resource) reconciler.ReconcileResult {
	res, ok := obj.(*ConfigResource)
	if !ok {
		return reconciler.ReconcileResult{Error: configErr("netintel: wrong type")}
	}
	now := time.Now()
	phase := "Disabled"
	if res.Spec.Enabled {
		phase = "Active"
	}
	res.Status.Phase = phase
	res.Status.ConfigActive = res.Spec.Enabled
	res.Status.ObservedGeneration = res.Generation
	res.Status.LastTransitionTime = now
	if rec.store != nil {
		_ = rec.store.Update(ctx, res)
	}
	return reconciler.ReconcileResult{}
}

type configErr string

func (e configErr) Error() string { return string(e) }
