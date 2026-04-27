package handlers

// Phase 6 P2 — NetIntel resource-ification.

import (
	"context"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/reconciler"
	"example.com/axiomnizam/internal/resources"
)

const (
	NetIntelConfigKind       = "NetIntelConfig"
	NetIntelConfigAPIVersion = "netintel.axiomnizam.io/v1"
)

type NetIntelConfigSpec struct {
	TenantID string `json:"tenantId,omitempty"`
	Enabled  bool   `json:"enabled"`
}

type NetIntelConfigResourceStatus struct {
	resources.ObjectStatus `json:",inline"`
	ConfigActive           bool `json:"configActive"`
}

type NetIntelConfigResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`
	Spec                 NetIntelConfigSpec           `json:"spec"`
	Status               NetIntelConfigResourceStatus `json:"status"`
}

func (r *NetIntelConfigResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *NetIntelConfigResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *NetIntelConfigResource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *NetIntelConfigResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil { r.Status.ObjectStatus = *s }
}
func (r *NetIntelConfigResource) DeepCopy() resources.Resource { cp := *r; return &cp }
func (r *NetIntelConfigResource) GetKey() string {
	if r.Namespace == "" { return r.Name }
	return r.Namespace + "/" + r.Name
}
func (r *NetIntelConfigResource) GetGeneration() int64         { return r.Generation }
func (r *NetIntelConfigResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }

type NetIntelConfigReconciler struct{ store store.ResourceStore[*NetIntelConfigResource] }

func NewNetIntelConfigReconciler(rs store.ResourceStore[*NetIntelConfigResource]) *NetIntelConfigReconciler {
	return &NetIntelConfigReconciler{store: rs}
}

func (rec *NetIntelConfigReconciler) Reconcile(ctx context.Context, obj reconciler.Resource) reconciler.ReconcileResult {
	res, ok := obj.(*NetIntelConfigResource)
	if !ok { return reconciler.ReconcileResult{Error: netintelErr("netintel: wrong type")} }
	now := time.Now()
	phase := "Disabled"
	if res.Spec.Enabled { phase = "Active" }
	res.Status.Phase = phase
	res.Status.ConfigActive = res.Spec.Enabled
	res.Status.ObservedGeneration = res.Generation
	res.Status.LastTransitionTime = now
	if rec.store != nil { _ = rec.store.Update(ctx, res) }
	return reconciler.ReconcileResult{}
}

type netintelErr string
func (e netintelErr) Error() string { return string(e) }
