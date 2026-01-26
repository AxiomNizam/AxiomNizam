package status

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ========== RESOURCE STATUS TRACKING ==========

// ResourceStatus represents current status of a resource
type ResourceStatus struct {
	Kind               string
	Namespace          string
	Name               string
	Phase              string // Pending, Validating, Acting, Succeeded, Failed
	Ready              bool
	Message            string
	Generation         int64
	ObservedGeneration int64
	Conditions         []Condition
	LastUpdateTime     time.Time
	ReconciliationID   string // Trace ID for debugging
}

// Condition represents a resource condition
type Condition struct {
	Type               string // Ready, Validating, Synced, etc.
	Status             string // True, False, Unknown
	Reason             string
	Message            string
	LastUpdateTime     time.Time
	LastTransitionTime time.Time
}

// StatusTracker tracks resource reconciliation status
type StatusTracker struct {
	mu       sync.RWMutex
	statuses map[string]*ResourceStatus // key = namespace/name
}

// NewStatusTracker creates tracker
func NewStatusTracker() *StatusTracker {
	return &StatusTracker{
		statuses: make(map[string]*ResourceStatus),
	}
}

// SetStatus sets or updates resource status
func (st *StatusTracker) SetStatus(status *ResourceStatus) {
	st.mu.Lock()
	defer st.mu.Unlock()

	key := fmt.Sprintf("%s/%s/%s", status.Kind, status.Namespace, status.Name)
	status.LastUpdateTime = time.Now()
	st.statuses[key] = status
}

// GetStatus gets resource status
func (st *StatusTracker) GetStatus(kind, namespace, name string) *ResourceStatus {
	st.mu.RLock()
	defer st.mu.RUnlock()

	key := fmt.Sprintf("%s/%s/%s", kind, namespace, name)
	return st.statuses[key]
}

// DeleteStatus deletes status
func (st *StatusTracker) DeleteStatus(kind, namespace, name string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	key := fmt.Sprintf("%s/%s/%s", kind, namespace, name)
	delete(st.statuses, key)
}

// ListStatuses lists all statuses
func (st *StatusTracker) ListStatuses() []*ResourceStatus {
	st.mu.RLock()
	defer st.mu.RUnlock()

	result := make([]*ResourceStatus, 0, len(st.statuses))
	for _, status := range st.statuses {
		result = append(result, status)
	}
	return result
}

// IsReady checks if resource is ready
func (st *StatusTracker) IsReady(kind, namespace, name string) bool {
	status := st.GetStatus(kind, namespace, name)
	return status != nil && status.Ready && status.Phase == "Succeeded"
}

// IsFailed checks if reconciliation failed
func (st *StatusTracker) IsFailed(kind, namespace, name string) bool {
	status := st.GetStatus(kind, namespace, name)
	return status != nil && status.Phase == "Failed"
}

// ========== GENERATION TRACKING FOR DRIFT DETECTION ==========

// GenerationTracker tracks resource generations for change detection
type GenerationTracker struct {
	mu          sync.RWMutex
	generations map[string]int64 // key = namespace/name -> generation
}

// NewGenerationTracker creates tracker
func NewGenerationTracker() *GenerationTracker {
	return &GenerationTracker{
		generations: make(map[string]int64),
	}
}

// SetGeneration sets resource generation
func (gt *GenerationTracker) SetGeneration(kind, namespace, name string, generation int64) {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	key := fmt.Sprintf("%s/%s/%s", kind, namespace, name)
	gt.generations[key] = generation
}

// GetGeneration gets resource generation
func (gt *GenerationTracker) GetGeneration(kind, namespace, name string) int64 {
	gt.mu.RLock()
	defer gt.mu.RUnlock()

	key := fmt.Sprintf("%s/%s/%s", kind, namespace, name)
	return gt.generations[key]
}

// HasChanged checks if generation has changed
func (gt *GenerationTracker) HasChanged(kind, namespace, name string, currentGen int64) bool {
	lastGen := gt.GetGeneration(kind, namespace, name)
	return currentGen > lastGen
}

// ========== RECONCILIATION EVENT LOG ==========

// ReconciliationEvent represents a reconciliation event
type ReconciliationEvent struct {
	Timestamp        time.Time
	ReconciliationID string
	Kind             string
	Namespace        string
	Name             string
	Event            string // Started, Observing, Validating, Acting, Succeeded, Failed
	Message          string
	Error            string
}

// ReconciliationEventLog logs reconciliation events for debugging
type ReconciliationEventLog struct {
	mu     sync.RWMutex
	events []ReconciliationEvent
	maxLen int
}

// NewReconciliationEventLog creates event log
func NewReconciliationEventLog(maxLen int) *ReconciliationEventLog {
	return &ReconciliationEventLog{
		events: make([]ReconciliationEvent, 0),
		maxLen: maxLen,
	}
}

// Log logs an event
func (rel *ReconciliationEventLog) Log(event ReconciliationEvent) {
	rel.mu.Lock()
	defer rel.mu.Unlock()

	event.Timestamp = time.Now()
	rel.events = append(rel.events, event)

	// Keep only last maxLen events
	if len(rel.events) > rel.maxLen {
		rel.events = rel.events[len(rel.events)-rel.maxLen:]
	}
}

// GetEvents returns all logged events
func (rel *ReconciliationEventLog) GetEvents() []ReconciliationEvent {
	rel.mu.RLock()
	defer rel.mu.RUnlock()

	result := make([]ReconciliationEvent, len(rel.events))
	copy(result, rel.events)
	return result
}

// GetEventsForResource returns events for specific resource
func (rel *ReconciliationEventLog) GetEventsForResource(kind, namespace, name string) []ReconciliationEvent {
	rel.mu.RLock()
	defer rel.mu.RUnlock()

	result := make([]ReconciliationEvent, 0)
	for _, e := range rel.events {
		if e.Kind == kind && e.Namespace == namespace && e.Name == name {
			result = append(result, e)
		}
	}
	return result
}

// ========== CONDITION MANAGER ==========

// ConditionManager manages resource conditions
type ConditionManager struct {
	mu         sync.RWMutex
	conditions map[string][]Condition // key = namespace/name -> conditions
}

// NewConditionManager creates manager
func NewConditionManager() *ConditionManager {
	return &ConditionManager{
		conditions: make(map[string][]Condition),
	}
}

// SetCondition sets or updates condition
func (cm *ConditionManager) SetCondition(kind, namespace, name string, cond Condition) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	key := fmt.Sprintf("%s/%s/%s", kind, namespace, name)
	cond.LastUpdateTime = time.Now()

	if cm.conditions[key] == nil {
		cm.conditions[key] = make([]Condition, 0)
	}

	// Update existing or add new
	found := false
	for i, c := range cm.conditions[key] {
		if c.Type == cond.Type {
			if c.Status != cond.Status {
				cond.LastTransitionTime = time.Now()
			}
			cm.conditions[key][i] = cond
			found = true
			break
		}
	}

	if !found {
		cond.LastTransitionTime = time.Now()
		cm.conditions[key] = append(cm.conditions[key], cond)
	}
}

// GetCondition gets condition by type
func (cm *ConditionManager) GetCondition(kind, namespace, name, condType string) *Condition {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	key := fmt.Sprintf("%s/%s/%s", kind, namespace, name)
	for i := range cm.conditions[key] {
		if cm.conditions[key][i].Type == condType {
			return &cm.conditions[key][i]
		}
	}
	return nil
}

// GetConditions gets all conditions
func (cm *ConditionManager) GetConditions(kind, namespace, name string) []Condition {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	key := fmt.Sprintf("%s/%s/%s", kind, namespace, name)
	result := make([]Condition, len(cm.conditions[key]))
	copy(result, cm.conditions[key])
	return result
}

// RemoveCondition removes condition
func (cm *ConditionManager) RemoveCondition(kind, namespace, name, condType string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	key := fmt.Sprintf("%s/%s/%s", kind, namespace, name)
	filtered := make([]Condition, 0)
	for _, c := range cm.conditions[key] {
		if c.Type != condType {
			filtered = append(filtered, c)
		}
	}
	cm.conditions[key] = filtered
}

// ========== WAITERS FOR TESTS ==========

// StatusWaiter allows waiting for status changes
type StatusWaiter struct {
	tracker *StatusTracker
	done    chan struct{}
}

// NewStatusWaiter creates waiter
func NewStatusWaiter(tracker *StatusTracker) *StatusWaiter {
	return &StatusWaiter{
		tracker: tracker,
		done:    make(chan struct{}, 1),
	}
}

// Wait waits for resource to become ready
func (sw *StatusWaiter) Wait(ctx context.Context, kind, namespace, name string) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if sw.tracker.IsReady(kind, namespace, name) {
				return nil
			}
			if sw.tracker.IsFailed(kind, namespace, name) {
				status := sw.tracker.GetStatus(kind, namespace, name)
				return fmt.Errorf("reconciliation failed: %s", status.Message)
			}
		}
	}
}

// ========== STATUS QUERY BUILDER ==========

// StatusQuery allows building queries for statuses
type StatusQuery struct {
	tracker *StatusTracker
	filters map[string]string
}

// NewStatusQuery creates query
func NewStatusQuery(tracker *StatusTracker) *StatusQuery {
	return &StatusQuery{
		tracker: tracker,
		filters: make(map[string]string),
	}
}

// WithKind filters by kind
func (sq *StatusQuery) WithKind(kind string) *StatusQuery {
	sq.filters["kind"] = kind
	return sq
}

// WithNamespace filters by namespace
func (sq *StatusQuery) WithNamespace(namespace string) *StatusQuery {
	sq.filters["namespace"] = namespace
	return sq
}

// WithPhase filters by phase
func (sq *StatusQuery) WithPhase(phase string) *StatusQuery {
	sq.filters["phase"] = phase
	return sq
}

// Execute executes query
func (sq *StatusQuery) Execute() []*ResourceStatus {
	allStatuses := sq.tracker.ListStatuses()
	result := make([]*ResourceStatus, 0)

	for _, status := range allStatuses {
		if kind, ok := sq.filters["kind"]; ok && status.Kind != kind {
			continue
		}
		if namespace, ok := sq.filters["namespace"]; ok && status.Namespace != namespace {
			continue
		}
		if phase, ok := sq.filters["phase"]; ok && status.Phase != phase {
			continue
		}
		result = append(result, status)
	}

	return result
}
