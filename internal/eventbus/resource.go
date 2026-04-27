package eventbus

// =====================================================
// P2 resource-ification — EventBus Topic & Subscription.
//
// TopicResource and SubscriptionResource wrap the imperative
// EventTopic and EventSubscription structs so a controller can
// reconcile event-bus primitives as first-class platform resources.
// =====================================================

import (
	"time"

	"example.com/axiomnizam/internal/resources"
)

const (
	TopicKind              = "EventBusTopic"
	TopicAPIVersion        = "eventbus.axiomnizam.io/v1"
	SubscriptionKind       = "EventBusSubscription"
	SubscriptionAPIVersion = "eventbus.axiomnizam.io/v1"
)

// --- TopicResource ---

// TopicSpec is the desired state of an event-bus topic.
type TopicSpec struct {
	Description       string          `json:"description,omitempty"`
	Schema            EventSchema     `json:"schema,omitempty"`
	Partitions        int             `json:"partitions,omitempty"`
	ReplicationFactor int             `json:"replicationFactor,omitempty"`
	Retention         RetentionConfig `json:"retention,omitempty"`
	Config            TopicConfig     `json:"config,omitempty"`
	Active            bool            `json:"active"`
}

// TopicResourceStatus extends the canonical object status.
type TopicResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	MessageCount int64 `json:"messageCount"`
	TopicActive  bool  `json:"topicActive"`
}

// TopicResource is the declarative resource for an EventBusTopic.
type TopicResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   TopicSpec           `json:"spec"`
	Status TopicResourceStatus `json:"status"`
}

func (r *TopicResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *TopicResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *TopicResource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *TopicResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *TopicResource) DeepCopy() resources.Resource { cp := *r; return &cp }
func (r *TopicResource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *TopicResource) GetGeneration() int64         { return r.Generation }
func (r *TopicResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }

// --- SubscriptionResource ---

// SubscriptionSpec is the desired state of an event-bus subscription.
type SubscriptionSpec struct {
	TenantID      string             `json:"tenantId"`
	Topics        []string           `json:"topics"`
	ConsumerGroup string             `json:"consumerGroup,omitempty"`
	Handler       string             `json:"handler,omitempty"`
	Filter        EventFilter        `json:"filter,omitempty"`
	Config        SubscriptionConfig `json:"config,omitempty"`
	Paused        bool               `json:"paused,omitempty"`
}

// SubscriptionResourceStatus extends the canonical object status.
type SubscriptionResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	SubscriptionStatus string    `json:"subscriptionStatus"`
	Offset             int64     `json:"offset"`
	Lag                int64     `json:"lag"`
	ProcessedCount     int64     `json:"processedCount"`
	FailedCount        int64     `json:"failedCount"`
	LastProcessed      time.Time `json:"lastProcessed,omitempty"`
}

// SubscriptionResource is the declarative resource for an EventBusSubscription.
type SubscriptionResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   SubscriptionSpec           `json:"spec"`
	Status SubscriptionResourceStatus `json:"status"`
}

func (r *SubscriptionResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *SubscriptionResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *SubscriptionResource) GetStatus() *resources.ObjectStatus {
	return &r.Status.ObjectStatus
}
func (r *SubscriptionResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *SubscriptionResource) DeepCopy() resources.Resource {
	cp := *r
	if len(r.Spec.Topics) > 0 {
		cp.Spec.Topics = append([]string(nil), r.Spec.Topics...)
	}
	return &cp
}
func (r *SubscriptionResource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *SubscriptionResource) GetGeneration() int64         { return r.Generation }
func (r *SubscriptionResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }
