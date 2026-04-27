package streaming

// =====================================================
// P2 resource-ification — Streaming.
//
// StreamResource wraps the imperative StreamSubscription so a
// controller can reconcile stream subscriptions as first-class
// platform resources.
// =====================================================

import (
	"time"

	"example.com/axiomnizam/internal/resources"
)

const (
	StreamKind       = "Stream"
	StreamAPIVersion = "streaming.axiomnizam.io/v1"
)

// StreamSpec is the desired state of a stream subscription.
type StreamSpec struct {
	TenantID     string                 `json:"tenantId"`
	Topic        string                 `json:"topic"`
	EventTypes   []string               `json:"eventTypes,omitempty"`
	Filter       map[string]interface{} `json:"filter,omitempty"`
	DeliveryMode DeliveryMode           `json:"deliveryMode,omitempty"`
	ChunkSize    int                    `json:"chunkSize,omitempty"`
	Format       OutputFormat           `json:"format,omitempty"`
	Timeout      int                    `json:"timeout,omitempty"`
	Active       bool                   `json:"active"`
}

// StreamResourceStatus extends the canonical object status.
type StreamResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	StreamActive bool       `json:"streamActive"`
	MessageCount int64      `json:"messageCount"`
	LastActivity *time.Time `json:"lastActivity,omitempty"`
}

// StreamResource is the declarative resource for a Stream.
type StreamResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   StreamSpec           `json:"spec"`
	Status StreamResourceStatus `json:"status"`
}

func (r *StreamResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *StreamResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *StreamResource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *StreamResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *StreamResource) DeepCopy() resources.Resource {
	cp := *r
	if len(r.Spec.EventTypes) > 0 {
		cp.Spec.EventTypes = append([]string(nil), r.Spec.EventTypes...)
	}
	return &cp
}
func (r *StreamResource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *StreamResource) GetGeneration() int64         { return r.Generation }
func (r *StreamResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }
