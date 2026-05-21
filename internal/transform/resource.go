package transform

// Phase 6 P2 — Transformation resource-ification.

import (
	"context"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/reconciler"
	"example.com/axiomnizam/internal/resources"
)

const (
	RuleKind       = "TransformRule"
	RuleAPIVersion = "transform.axiomnizam.io/v1"
)

type RuleSpec struct {
	RuleName    string                 `json:"ruleName"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

type RuleResourceStatus struct {
	resources.ObjectStatus `json:",inline"`
	RuleActive             bool `json:"ruleActive"`
}

type RuleResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`
	Spec                 RuleSpec           `json:"spec"`
	Status               RuleResourceStatus `json:"status"`
}

func (r *RuleResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *RuleResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *RuleResource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *RuleResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *RuleResource) DeepCopy() resources.Resource { cp := *r; return &cp }
func (r *RuleResource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *RuleResource) GetGeneration() int64         { return r.Generation }
func (r *RuleResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }

type RuleReconciler struct {
	store store.ResourceStore[*RuleResource]
}

func NewRuleReconciler(rs store.ResourceStore[*RuleResource]) *RuleReconciler {
	return &RuleReconciler{store: rs}
}

func (rec *RuleReconciler) Reconcile(ctx context.Context, obj reconciler.Resource) reconciler.ReconcileResult {
	res, ok := obj.(*RuleResource)
	if !ok {
		return reconciler.ReconcileResult{Error: ruleErr("transform: wrong type")}
	}
	now := time.Now()
	phase := "Disabled"
	if res.Spec.Enabled {
		phase = "Active"
	}
	res.Status.Phase = phase
	res.Status.RuleActive = res.Spec.Enabled
	res.Status.ObservedGeneration = res.Generation
	res.Status.LastTransitionTime = now
	if rec.store != nil {
		_ = rec.store.Update(ctx, res)
	}
	return reconciler.ReconcileResult{}
}

type ruleErr string

func (e ruleErr) Error() string { return string(e) }
