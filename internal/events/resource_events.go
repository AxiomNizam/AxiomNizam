package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ResourceEventType represents lifecycle events for resources
type ResourceEventType string

const (
	// Resource lifecycle events
	EventResourceCreated      ResourceEventType = "resource.created"
	EventResourceUpdated      ResourceEventType = "resource.updated"
	EventResourceDeleted      ResourceEventType = "resource.deleted"
	EventResourceReconciled   ResourceEventType = "resource.reconciled"
	EventResourceError        ResourceEventType = "resource.error"
	EventResourceWarning      ResourceEventType = "resource.warning"
	EventResourceHealthy      ResourceEventType = "resource.healthy"
	EventResourceUnhealthy    ResourceEventType = "resource.unhealthy"
	EventResourceStatusChange ResourceEventType = "resource.status_change"

	// Policy events
	EventPolicyEnforced  ResourceEventType = "policy.enforced"
	EventPolicyViolation ResourceEventType = "policy.violation"
	EventPolicyMutated   ResourceEventType = "policy.mutated"

	// RBAC events
	EventRBACDecision ResourceEventType = "rbac.decision"
	EventRBACDenied   ResourceEventType = "rbac.denied"

	// Admission events
	EventAdmissionAccepted ResourceEventType = "admission.accepted"
	EventAdmissionRejected ResourceEventType = "admission.rejected"

	// Backup events
	EventBackupStarted   ResourceEventType = "backup.started"
	EventBackupCompleted ResourceEventType = "backup.completed"
	EventBackupFailed    ResourceEventType = "backup.failed"

	// Cluster events
	EventClusterEvent ResourceEventType = "cluster.event"
)

// ResourceEvent represents an event in resource lifecycle
type ResourceEvent struct {
	ID               string                 `json:"id"`
	Type             ResourceEventType      `json:"type"`
	Kind             string                 `json:"kind"`
	Name             string                 `json:"name"`
	Namespace        string                 `json:"namespace"`
	ResourceVersion  int64                  `json:"resource_version"`
	Generation       int64                  `json:"generation"`
	Timestamp        time.Time              `json:"timestamp"`
	FirstTimestamp   time.Time              `json:"first_timestamp"`
	LastTimestamp    time.Time              `json:"last_timestamp"`
	Count            int                    `json:"count"`
	Reason           string                 `json:"reason"`
	Message          string                 `json:"message"`
	Source           string                 `json:"source"`
	InvolvedObject   *ObjectReference       `json:"involved_object"`
	Related          *ObjectReference       `json:"related,omitempty"`
	EventTime        *MicroTime             `json:"event_time,omitempty"`
	Series           *EventSeries           `json:"series,omitempty"`
	Action           string                 `json:"action,omitempty"`
	UserID           string                 `json:"user_id,omitempty"`
	ImpersonatedUser string                 `json:"impersonated_user,omitempty"`
	CorrelationID    string                 `json:"correlation_id,omitempty"`
	Severity         string                 `json:"severity"` // info, warning, error, critical
	Tags             map[string]string      `json:"tags,omitempty"`
	Data             map[string]interface{} `json:"data,omitempty"`
}

// ObjectReference is a reference to another object
type ObjectReference struct {
	Kind            string `json:"kind"`
	Namespace       string `json:"namespace,omitempty"`
	Name            string `json:"name"`
	UID             string `json:"uid,omitempty"`
	APIVersion      string `json:"api_version,omitempty"`
	ResourceVersion string `json:"resource_version,omitempty"`
}

// MicroTime represents timestamp with microsecond precision
type MicroTime struct {
	Time time.Time
}

// EventSeries represents aggregated events
type EventSeries struct {
	Count            int32
	LastObservedTime *MicroTime
	State            string // "ongoing", "completed"
}

// EventStore manages event storage and retrieval
type EventStore struct {
	mu              sync.RWMutex
	events          []*ResourceEvent
	maxEvents       int
	byKind          map[string][]*ResourceEvent
	byNamespace     map[string][]*ResourceEvent
	byType          map[ResourceEventType][]*ResourceEvent
	eventAggregator *EventAggregator
}

// EventAggregator aggregates similar events
type EventAggregator struct {
	mu             sync.RWMutex
	aggregations   map[string]*EventSeries
	aggregationTTL time.Duration
	lastAggregated map[string]time.Time
}

// EventWatcher watches for specific events
type EventWatcher struct {
	ID      string
	Filters *EventFilter
	Channel chan *ResourceEvent
	Done    chan struct{}
	Closed  bool
}

// EventFilter allows filtering events
type EventFilter struct {
	Kind        string
	Namespace   string
	Name        string
	Types       []ResourceEventType
	Severities  []string
	Sources     []string
	StartTime   *time.Time
	EndTime     *time.Time
	IncludeTags map[string]string
	ExcludeTags map[string]string
}

// EventBusWithLifecycle extends event bus with resource lifecycle support
type EventBusWithLifecycle struct {
	mu               sync.RWMutex
	store            *EventStore
	subscribers      map[ResourceEventType][]EventHandler
	wildcardHandlers []EventHandler
	watchers         map[string]*EventWatcher
	auditLog         []*EventAuditLog
	metrics          *EventMetrics
	retentionPolicy  *RetentionPolicy
	streamBuffer     chan *ResourceEvent
}

// EventAuditLog tracks event publishing for compliance
type EventAuditLog struct {
	ID              string
	Timestamp       time.Time
	EventID         string
	EventType       ResourceEventType
	Kind            string
	Namespace       string
	Name            string
	Publisher       string
	SubscriberCount int
	Dropped         bool
	Error           string
}

// EventMetrics tracks event bus metrics
type EventMetrics struct {
	TotalEvents      int64
	EventsByType     map[ResourceEventType]int64
	EventsByKind     map[string]int64
	SubscriberCount  int
	WatcherCount     int
	AggregatedEvents int64
	DroppedEvents    int64
	PublishErrors    int64
}

// RetentionPolicy defines event retention
type RetentionPolicy struct {
	MaxAge       time.Duration
	MaxCount     int
	ArchiveAfter time.Duration
}

// NewEventBusWithLifecycle creates an event bus with lifecycle support
func NewEventBusWithLifecycle(maxEvents int, retentionPolicy *RetentionPolicy) *EventBusWithLifecycle {
	if retentionPolicy == nil {
		retentionPolicy = &RetentionPolicy{
			MaxAge:       7 * 24 * time.Hour,
			MaxCount:     100000,
			ArchiveAfter: 30 * 24 * time.Hour,
		}
	}

	return &EventBusWithLifecycle{
		store:            NewEventStore(maxEvents),
		subscribers:      make(map[ResourceEventType][]EventHandler),
		wildcardHandlers: make([]EventHandler, 0),
		watchers:         make(map[string]*EventWatcher),
		auditLog:         make([]*EventAuditLog, 0, 10000),
		metrics:          &EventMetrics{EventsByType: make(map[ResourceEventType]int64), EventsByKind: make(map[string]int64)},
		retentionPolicy:  retentionPolicy,
		streamBuffer:     make(chan *ResourceEvent, 1000),
	}
}

// PublishResourceEvent publishes a resource lifecycle event
func (eb *EventBusWithLifecycle) PublishResourceEvent(ctx context.Context, event *ResourceEvent) error {
	if event.ID == "" {
		event.ID = fmt.Sprintf("event-%d", time.Now().UnixNano())
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Store event
	eb.mu.Lock()
	eb.store.AddEvent(event)
	eb.metrics.TotalEvents++
	eb.metrics.EventsByType[event.Type]++
	eb.metrics.EventsByKind[event.Kind]++
	eb.mu.Unlock()

	// Dispatch to subscribers
	go eb.dispatchEvent(ctx, event)

	// Log audit entry
	eb.recordAuditLog(event)

	return nil
}

// dispatchEvent dispatches event to all subscribers
func (eb *EventBusWithLifecycle) dispatchEvent(ctx context.Context, event *ResourceEvent) {
	eb.mu.RLock()
	subscribers := make([]EventHandler, 0)

	// Get type-specific subscribers
	if subs, exists := eb.subscribers[event.Type]; exists {
		subscribers = append(subscribers, subs...)
	}

	// Add wildcard subscribers
	subscribers = append(subscribers, eb.wildcardHandlers...)

	// Notify watchers
	watchers := make([]*EventWatcher, 0)
	for _, w := range eb.watchers {
		if w.matchesFilter(event) {
			watchers = append(watchers, w)
		}
	}

	eb.mu.RUnlock()

	// Call subscribers
	for _, handler := range subscribers {
		go func(h EventHandler) {
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := h(ctx, &Event{
				ID:            event.ID,
				Type:          EventType(event.Type),
				Source:        event.Source,
				Timestamp:     event.Timestamp,
				UserID:        event.UserID,
				CorrelationID: event.CorrelationID,
				Data: map[string]interface{}{
					"kind":      event.Kind,
					"name":      event.Name,
					"namespace": event.Namespace,
					"reason":    event.Reason,
					"message":   event.Message,
				},
			}); err != nil {
				eb.mu.Lock()
				eb.metrics.PublishErrors++
				eb.mu.Unlock()
			}
		}(handler)
	}

	// Send to watchers
	for _, w := range watchers {
		select {
		case w.Channel <- event:
		default:
			eb.mu.Lock()
			eb.metrics.DroppedEvents++
			eb.mu.Unlock()
		}
	}
}

// SubscribeToResourceEvents subscribes to resource events of specific types
func (eb *EventBusWithLifecycle) SubscribeToResourceEvents(eventTypes []ResourceEventType, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for _, eventType := range eventTypes {
		eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
	}

	return nil
}

// WatchResourceEvents creates a watcher for filtered events
func (eb *EventBusWithLifecycle) WatchResourceEvents(ctx context.Context, filter *EventFilter) (*EventWatcher, error) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	watcher := &EventWatcher{
		ID:      fmt.Sprintf("watcher-%d", time.Now().UnixNano()),
		Filters: filter,
		Channel: make(chan *ResourceEvent, 100),
		Done:    make(chan struct{}),
	}

	eb.watchers[watcher.ID] = watcher
	eb.metrics.WatcherCount = len(eb.watchers)

	return watcher, nil
}

// GetEvents retrieves events matching criteria
func (eb *EventBusWithLifecycle) GetEvents(ctx context.Context, filter *EventFilter, limit int) []*ResourceEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return eb.store.QueryEvents(filter, limit)
}

// GetEventHistory retrieves event history for a resource
func (eb *EventBusWithLifecycle) GetEventHistory(ctx context.Context, kind string, namespace string, name string, limit int) []*ResourceEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return eb.store.GetResourceEventHistory(kind, namespace, name, limit)
}

// recordAuditLog records event publishing
func (eb *EventBusWithLifecycle) recordAuditLog(event *ResourceEvent) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	audit := &EventAuditLog{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		EventID:   event.ID,
		EventType: event.Type,
		Kind:      event.Kind,
		Namespace: event.Namespace,
		Name:      event.Name,
		Publisher: event.Source,
	}

	eb.auditLog = append(eb.auditLog, audit)
	if len(eb.auditLog) > 10000 {
		eb.auditLog = eb.auditLog[len(eb.auditLog)-10000:]
	}
}

// GetMetrics returns event bus metrics
func (eb *EventBusWithLifecycle) GetMetrics() *EventMetrics {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.metrics
}

// NewEventStore creates a new event store
func NewEventStore(maxEvents int) *EventStore {
	return &EventStore{
		events:      make([]*ResourceEvent, 0, maxEvents),
		maxEvents:   maxEvents,
		byKind:      make(map[string][]*ResourceEvent),
		byNamespace: make(map[string][]*ResourceEvent),
		byType:      make(map[ResourceEventType][]*ResourceEvent),
		eventAggregator: &EventAggregator{
			aggregations:   make(map[string]*EventSeries),
			aggregationTTL: 1 * time.Minute,
			lastAggregated: make(map[string]time.Time),
		},
	}
}

// AddEvent adds an event to the store
func (es *EventStore) AddEvent(event *ResourceEvent) {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.events = append(es.events, event)
	es.byKind[event.Kind] = append(es.byKind[event.Kind], event)
	es.byNamespace[event.Namespace] = append(es.byNamespace[event.Namespace], event)
	es.byType[event.Type] = append(es.byType[event.Type], event)

	if len(es.events) > es.maxEvents {
		removed := es.events[0]
		es.events = es.events[1:]

		// Update indexes
		es.byKind[removed.Kind] = es.events
		es.byNamespace[removed.Namespace] = es.events
		es.byType[removed.Type] = es.events
	}
}

// QueryEvents queries events with filters
func (es *EventStore) QueryEvents(filter *EventFilter, limit int) []*ResourceEvent {
	es.mu.RLock()
	defer es.mu.RUnlock()

	result := make([]*ResourceEvent, 0)
	count := 0

	for i := len(es.events) - 1; i >= 0 && count < limit; i-- {
		event := es.events[i]

		// Apply filters
		if filter.Kind != "" && event.Kind != filter.Kind {
			continue
		}
		if filter.Namespace != "" && event.Namespace != filter.Namespace {
			continue
		}
		if filter.Name != "" && event.Name != filter.Name {
			continue
		}
		if len(filter.Types) > 0 {
			found := false
			for _, t := range filter.Types {
				if event.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		result = append(result, event)
		count++
	}

	return result
}

// GetResourceEventHistory gets event history for a resource
func (es *EventStore) GetResourceEventHistory(kind string, namespace string, name string, limit int) []*ResourceEvent {
	es.mu.RLock()
	defer es.mu.RUnlock()

	result := make([]*ResourceEvent, 0)
	count := 0

	for i := len(es.events) - 1; i >= 0 && count < limit; i-- {
		event := es.events[i]
		if event.Kind == kind && event.Namespace == namespace && event.Name == name {
			result = append(result, event)
			count++
		}
	}

	return result
}

// matchesFilter checks if event matches watcher filter
func (w *EventWatcher) matchesFilter(event *ResourceEvent) bool {
	if w.Filters == nil {
		return true
	}

	if w.Filters.Kind != "" && event.Kind != w.Filters.Kind {
		return false
	}
	if w.Filters.Namespace != "" && event.Namespace != w.Filters.Namespace {
		return false
	}
	if w.Filters.Name != "" && event.Name != w.Filters.Name {
		return false
	}
	if len(w.Filters.Types) > 0 {
		found := false
		for _, t := range w.Filters.Types {
			if event.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// Close closes the watcher
func (w *EventWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.Closed {
		w.Closed = true
		close(w.Channel)
		close(w.Done)
	}

	return nil
}

// AddMutex for concurrent access
var wm sync.RWMutex

func (w *EventWatcher) mu() sync.RWMutex {
	return wm
}
