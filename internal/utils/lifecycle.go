package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Finalizer represents a cleanup handler for resource deletion
type Finalizer struct {
	Name    string
	Handler func(context.Context, interface{}) error
}

// FinalizerManager manages resource finalizers (Kubernetes-style)
// Ensures cleanup operations run before resource deletion
type FinalizerManager struct {
	mu         sync.RWMutex
	finalizers map[string][]Finalizer
}

// NewFinalizerManager creates a new finalizer manager
func NewFinalizerManager() *FinalizerManager {
	return &FinalizerManager{
		finalizers: make(map[string][]Finalizer),
	}
}

// AddFinalizer adds a finalizer to a resource
func (fm *FinalizerManager) AddFinalizer(resourceID string, finalizer Finalizer) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Check if finalizer already exists
	for _, f := range fm.finalizers[resourceID] {
		if f.Name == finalizer.Name {
			return
		}
	}

	fm.finalizers[resourceID] = append(fm.finalizers[resourceID], finalizer)
}

// RemoveFinalizer removes a finalizer from a resource
func (fm *FinalizerManager) RemoveFinalizer(resourceID, name string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	finalizers, ok := fm.finalizers[resourceID]
	if !ok {
		return fmt.Errorf("resource not found")
	}

	filtered := make([]Finalizer, 0)
	found := false
	for _, f := range finalizers {
		if f.Name != name {
			filtered = append(filtered, f)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("finalizer not found: %s", name)
	}

	fm.finalizers[resourceID] = filtered
	if len(filtered) == 0 {
		delete(fm.finalizers, resourceID)
	}

	return nil
}

// RunFinalizers runs all finalizers for a resource before deletion
func (fm *FinalizerManager) RunFinalizers(ctx context.Context, resourceID string, resource interface{}) error {
	fm.mu.RLock()
	finalizers := fm.finalizers[resourceID]
	fm.mu.RUnlock()

	if len(finalizers) == 0 {
		return nil
	}

	// Run finalizers in reverse order (cleanup in LIFO order)
	for i := len(finalizers) - 1; i >= 0; i-- {
		finalizer := finalizers[i]
		if err := finalizer.Handler(ctx, resource); err != nil {
			return fmt.Errorf("finalizer %s failed: %w", finalizer.Name, err)
		}
		// Remove successfully completed finalizer
		_ = fm.RemoveFinalizer(resourceID, finalizer.Name)
	}

	return nil
}

// HasFinalizers checks if resource has pending finalizers
func (fm *FinalizerManager) HasFinalizers(resourceID string) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	finalizers, ok := fm.finalizers[resourceID]
	return ok && len(finalizers) > 0
}

// ListFinalizers returns all finalizers for a resource
func (fm *FinalizerManager) ListFinalizers(resourceID string) []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	finalizers := fm.finalizers[resourceID]
	names := make([]string, len(finalizers))
	for i, f := range finalizers {
		names[i] = f.Name
	}
	return names
}

// OwnerReference represents a resource ownership relationship
type OwnerReference struct {
	Kind       string
	Name       string
	UID        string
	Controller bool
}

// GarbageCollector manages orphaned resource cleanup (Kubernetes-style)
type GarbageCollector struct {
	mu       sync.RWMutex
	owners   map[string][]string        // owner -> children
	children map[string]*OwnerReference // child -> owner
}

// NewGarbageCollector creates a new garbage collector
func NewGarbageCollector() *GarbageCollector {
	return &GarbageCollector{
		owners:   make(map[string][]string),
		children: make(map[string]*OwnerReference),
	}
}

// AddOwnershipRelation records parent-child relationship
func (gc *GarbageCollector) AddOwnershipRelation(child string, owner *OwnerReference) error {
	if owner == nil {
		return fmt.Errorf("owner reference cannot be nil")
	}

	gc.mu.Lock()
	defer gc.mu.Unlock()

	ownerKey := fmt.Sprintf("%s/%s", owner.Kind, owner.Name)
	gc.owners[ownerKey] = append(gc.owners[ownerKey], child)
	gc.children[child] = owner

	return nil
}

// GetChildren returns all children of an owner
func (gc *GarbageCollector) GetChildren(ownerKind, ownerName string) []string {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	ownerKey := fmt.Sprintf("%s/%s", ownerKind, ownerName)
	children := make([]string, len(gc.owners[ownerKey]))
	copy(children, gc.owners[ownerKey])
	return children
}

// GetOwner returns the owner of a child resource
func (gc *GarbageCollector) GetOwner(child string) *OwnerReference {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	return gc.children[child]
}

// MarkForDeletion marks children for deletion when owner is deleted
func (gc *GarbageCollector) MarkForDeletion(ownerKind, ownerName string) []string {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	ownerKey := fmt.Sprintf("%s/%s", ownerKind, ownerName)
	children := gc.owners[ownerKey]

	// Remove owner and its children from tracking
	delete(gc.owners, ownerKey)
	for _, child := range children {
		delete(gc.children, child)
	}

	return children
}

// StatusCondition represents resource condition (Kubernetes-style)
type StatusCondition struct {
	Type               string
	Status             string // "True", "False", "Unknown"
	Reason             string
	Message            string
	LastTransitionTime time.Time
	ObservedGeneration int64
}

// ConditionManager manages resource conditions
type ConditionManager struct {
	mu         sync.RWMutex
	conditions map[string][]StatusCondition // resource -> conditions
}

// NewConditionManager creates a new condition manager
func NewConditionManager() *ConditionManager {
	return &ConditionManager{
		conditions: make(map[string][]StatusCondition),
	}
}

// SetCondition adds or updates a condition
func (cm *ConditionManager) SetCondition(resourceID string, condition StatusCondition) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	condition.LastTransitionTime = time.Now()

	conditions := cm.conditions[resourceID]
	idx := -1

	// Find existing condition
	for i, c := range conditions {
		if c.Type == condition.Type {
			idx = i
			break
		}
	}

	if idx >= 0 {
		conditions[idx] = condition
	} else {
		conditions = append(conditions, condition)
	}

	cm.conditions[resourceID] = conditions
}

// GetCondition returns a specific condition
func (cm *ConditionManager) GetCondition(resourceID, conditionType string) *StatusCondition {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conditions := cm.conditions[resourceID]
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}

	return nil
}

// GetConditions returns all conditions for a resource
func (cm *ConditionManager) GetConditions(resourceID string) []StatusCondition {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conditions := cm.conditions[resourceID]
	result := make([]StatusCondition, len(conditions))
	copy(result, conditions)
	return result
}

// IsReady checks if resource is ready (Ready condition = True)
func (cm *ConditionManager) IsReady(resourceID string) bool {
	condition := cm.GetCondition(resourceID, "Ready")
	return condition != nil && condition.Status == "True"
}

// Event represents a kubernetes-style event
type Event struct {
	Type               string
	Reason             string
	Message            string
	Timestamp          time.Time
	InvolvedObject     string
	Count              int
	FirstOccurrence    time.Time
	LastOccurrence     time.Time
	ReportingComponent string
	Action             string
}

// EventRecorder records resource events (Kubernetes audit trail)
type EventRecorder struct {
	mu     sync.RWMutex
	events map[string][]Event // resource -> events
}

// NewEventRecorder creates a new event recorder
func NewEventRecorder() *EventRecorder {
	return &EventRecorder{
		events: make(map[string][]Event),
	}
}

// Record records an event
func (er *EventRecorder) Record(resourceID string, event Event) {
	er.mu.Lock()
	defer er.mu.Unlock()

	event.Timestamp = time.Now()
	if event.FirstOccurrence.IsZero() {
		event.FirstOccurrence = event.Timestamp
	}
	event.LastOccurrence = event.Timestamp

	er.events[resourceID] = append(er.events[resourceID], event)
}

// GetEvents returns events for a resource
func (er *EventRecorder) GetEvents(resourceID string) []Event {
	er.mu.RLock()
	defer er.mu.RUnlock()

	events := er.events[resourceID]
	result := make([]Event, len(events))
	copy(result, events)
	return result
}

// ClearEvents removes old events (older than retention period)
func (er *EventRecorder) ClearEvents(resourceID string, retention time.Duration) {
	er.mu.Lock()
	defer er.mu.Unlock()

	cutoff := time.Now().Add(-retention)
	events := er.events[resourceID]
	filtered := make([]Event, 0)

	for _, e := range events {
		if e.Timestamp.After(cutoff) {
			filtered = append(filtered, e)
		}
	}

	if len(filtered) == 0 {
		delete(er.events, resourceID)
	} else {
		er.events[resourceID] = filtered
	}
}
