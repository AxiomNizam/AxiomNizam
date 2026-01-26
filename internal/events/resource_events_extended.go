package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventSeverity represents event severity level
type EventSeverity string

const (
	EventSeverityNormal  EventSeverity = "Normal"
	EventSeverityWarning EventSeverity = "Warning"
)

// ResourceEvent represents a Kubernetes-style event
type ResourceEvent struct {
	// Core fields
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Type      EventSeverity `json:"type"`
	Reason    string        `json:"reason"` // e.g., "PolicyApplied", "QuotaExceeded"
	Message   string        `json:"message"`
	Source    string        `json:"source"` // e.g., "axiom-controller"
	FirstTime time.Time     `json:"firstTime"`
	LastTime  time.Time     `json:"lastTime"`
	Count     int           `json:"count"`

	// Involved object
	InvolvedObject ObjectReference `json:"involvedObject"`

	// Additional context
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ObjectReference references an object
type ObjectReference struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`
	UID        string `json:"uid,omitempty"`
}

// EventRecorder records events for resources
type EventRecorder interface {
	// Event records a normal event
	Event(ctx context.Context, obj ObjectReference, reason, message string)

	// Eventf records a normal event with formatting
	Eventf(ctx context.Context, obj ObjectReference, reason, messageFmt string, args ...interface{})

	// Warning records a warning event
	Warning(ctx context.Context, obj ObjectReference, reason, message string)

	// Warningf records a warning event with formatting
	Warningf(ctx context.Context, obj ObjectReference, reason, messageFmt string, args ...interface{})

	// GetEvents returns events for an object
	GetEvents(ctx context.Context, obj ObjectReference) ([]*ResourceEvent, error)
}

// SimpleEventRecorder implements EventRecorder
type SimpleEventRecorder struct {
	mu     sync.RWMutex
	events []*ResourceEvent
	store  map[string]*ResourceEvent // Key: namespace/name/reason
}

// NewEventRecorder creates a new event recorder
func NewEventRecorder() EventRecorder {
	return &SimpleEventRecorder{
		events: make([]*ResourceEvent, 0),
		store:  make(map[string]*ResourceEvent),
	}
}

// Event records a normal event
func (r *SimpleEventRecorder) Event(ctx context.Context, obj ObjectReference, reason, message string) {
	r.recordEvent(ctx, obj, EventSeverityNormal, reason, message)
}

// Eventf records a formatted normal event
func (r *SimpleEventRecorder) Eventf(ctx context.Context, obj ObjectReference, reason, messageFmt string, args ...interface{}) {
	message := fmt.Sprintf(messageFmt, args...)
	r.Event(ctx, obj, reason, message)
}

// Warning records a warning event
func (r *SimpleEventRecorder) Warning(ctx context.Context, obj ObjectReference, reason, message string) {
	r.recordEvent(ctx, obj, EventSeverityWarning, reason, message)
}

// Warningf records a formatted warning event
func (r *SimpleEventRecorder) Warningf(ctx context.Context, obj ObjectReference, reason, messageFmt string, args ...interface{}) {
	message := fmt.Sprintf(messageFmt, args...)
	r.Warning(ctx, obj, reason, message)
}

// GetEvents returns events for an object
func (r *SimpleEventRecorder) GetEvents(ctx context.Context, obj ObjectReference) ([]*ResourceEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*ResourceEvent, 0)
	for _, event := range r.events {
		if event.InvolvedObject.Name == obj.Name && event.InvolvedObject.Namespace == obj.Namespace {
			result = append(result, event)
		}
	}

	return result, nil
}

// recordEvent records an event, merging with existing if same reason
func (r *SimpleEventRecorder) recordEvent(ctx context.Context, obj ObjectReference, severity EventSeverity, reason, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	key := fmt.Sprintf("%s/%s/%s", obj.Namespace, obj.Name, reason)

	if existing, ok := r.store[key]; ok {
		// Update existing event
		existing.LastTime = now
		existing.Count++
		existing.Message = message
		return
	}

	// Create new event
	event := &ResourceEvent{
		Name:           fmt.Sprintf("%s.%x", obj.Name, now.Unix()),
		Namespace:      obj.Namespace,
		Type:           severity,
		Reason:         reason,
		Message:        message,
		Source:         "axiom-controller",
		FirstTime:      now,
		LastTime:       now,
		Count:          1,
		InvolvedObject: obj,
	}

	r.events = append(r.events, event)
	r.store[key] = event
}

// CommonEventReasons defines common event reasons
var CommonEventReasons = struct {
	PolicyApplied     string
	PolicyDenied      string
	QuotaExceeded     string
	ValidationFailed  string
	ReconcileFailed   string
	ResourceCreated   string
	ResourceUpdated   string
	ResourceDeleted   string
	WorkflowStarted   string
	WorkflowCompleted string
	WorkflowFailed    string
}{
	PolicyApplied:     "PolicyApplied",
	PolicyDenied:      "PolicyDenied",
	QuotaExceeded:     "QuotaExceeded",
	ValidationFailed:  "ValidationFailed",
	ReconcileFailed:   "ReconcileFailed",
	ResourceCreated:   "ResourceCreated",
	ResourceUpdated:   "ResourceUpdated",
	ResourceDeleted:   "ResourceDeleted",
	WorkflowStarted:   "WorkflowStarted",
	WorkflowCompleted: "WorkflowCompleted",
	WorkflowFailed:    "WorkflowFailed",
}

// GlobalEventRecorder is the package-level event recorder
var GlobalEventRecorder = NewEventRecorder()
