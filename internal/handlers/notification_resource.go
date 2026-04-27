package handlers

// Phase 6 P2 — Notification resource-ification.

import (
	"context"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/reconciler"
	"example.com/axiomnizam/internal/resources"
)

const (
	NotificationChannelKind       = "NotificationChannel"
	NotificationChannelAPIVersion = "notification.axiomnizam.io/v1"
)

type NotificationChannelSpec struct {
	ChannelType string `json:"channelType"` // discord, slack, email, webhook
	WebhookURL  string `json:"webhookUrl,omitempty"`
	Enabled     bool   `json:"enabled"`
}

type NotificationChannelResourceStatus struct {
	resources.ObjectStatus `json:",inline"`
	ChannelActive          bool `json:"channelActive"`
}

type NotificationChannelResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`
	Spec                 NotificationChannelSpec           `json:"spec"`
	Status               NotificationChannelResourceStatus `json:"status"`
}

func (r *NotificationChannelResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *NotificationChannelResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *NotificationChannelResource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *NotificationChannelResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil { r.Status.ObjectStatus = *s }
}
func (r *NotificationChannelResource) DeepCopy() resources.Resource { cp := *r; return &cp }
func (r *NotificationChannelResource) GetKey() string {
	if r.Namespace == "" { return r.Name }
	return r.Namespace + "/" + r.Name
}
func (r *NotificationChannelResource) GetGeneration() int64         { return r.Generation }
func (r *NotificationChannelResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }

type NotificationChannelReconciler struct{ store store.ResourceStore[*NotificationChannelResource] }

func NewNotificationChannelReconciler(rs store.ResourceStore[*NotificationChannelResource]) *NotificationChannelReconciler {
	return &NotificationChannelReconciler{store: rs}
}

func (rec *NotificationChannelReconciler) Reconcile(ctx context.Context, obj reconciler.Resource) reconciler.ReconcileResult {
	res, ok := obj.(*NotificationChannelResource)
	if !ok { return reconciler.ReconcileResult{Error: notifErr("notification: wrong type")} }
	now := time.Now()
	phase := "Disabled"
	if res.Spec.Enabled { phase = "Active" }
	res.Status.Phase = phase
	res.Status.ChannelActive = res.Spec.Enabled
	res.Status.ObservedGeneration = res.Generation
	res.Status.LastTransitionTime = now
	if rec.store != nil { _ = rec.store.Update(ctx, res) }
	return reconciler.ReconcileResult{}
}

type notifErr string
func (e notifErr) Error() string { return string(e) }
