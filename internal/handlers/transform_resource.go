package handlers

// Phase 6 P2 — Transformation resource-ification.

import (
	"context"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/reconciler"
	"example.com/axiomnizam/internal/resources"
)

const (
	TransformRuleKind       = "TransformRule"
	TransformRuleAPIVersion = "transform.axiomnizam.io/v1"
)

type TransformRuleSpec struct {
	RuleName    string                 `json:"ruleName"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

type TransformRuleResourceStatus struct {
	resources.ObjectStatus `json:",inline"`
	RuleActive             bool `json:"ruleActive"`
}

type TransformRuleResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`
	Spec                 TransformRuleSpec           `json:"spec"`
	Status               TransformRuleResourceStatus `json:"status"`
}

func (r *TransformRuleResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *TransformRuleResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *TransformRuleResource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *TransformRuleResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil { r.Status.ObjectStatus = *s }
}
func (r *TransformRuleResource) DeepCopy() resources.Resource { cp := *r; return &cp }
func (r *TransformRuleResource) GetKey() string {
	if r.Namespace == "" { return r.Name }
	return r.Namespace + "/" + r.Name
}
func (r *TransformRuleResource) GetGeneration() int64         { return r.Generation }
func (r *TransformRuleResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }

type TransformRuleReconciler struct{ store store.ResourceStore[*TransformRuleResource] }

func NewTransformRuleReconciler(rs store.ResourceStore[*TransformRuleResource]) *TransformRuleReconciler {
	return &TransformRuleReconciler{store: rs}
}

func (rec *TransformRuleReconciler) Reconcile(ctx context.Context, obj reconciler.Resource) reconciler.ReconcileResult {
	res, ok := obj.(*TransformRuleResource)
	if !ok { return reconciler.ReconcileResult{Error: transformErr("transform: wrong type")} }
	now := time.Now()
	phase := "Disabled"
	if res.Spec.Enabled { phase = "Active" }
	res.Status.Phase = phase
	res.Status.RuleActive = res.Spec.Enabled
	res.Status.ObservedGeneration = res.Generation
	res.Status.LastTransitionTime = now
	if rec.store != nil { _ = rec.store.Update(ctx, res) }
	return reconciler.ReconcileResult{}
}

type transformErr string
func (e transformErr) Error() string { return string(e) }
