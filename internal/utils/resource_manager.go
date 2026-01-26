package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ResourceManager provides comprehensive resource lifecycle management with reconciliation support
// Implements Kubernetes-style resource management patterns
type ResourceManager struct {
	mu          sync.RWMutex
	resources   map[string]map[string]*ManagedResource // namespace -> kind -> resources
	finalizers  map[string]*FinalizerManager
	conditions  map[string]*ConditionManager
	events      map[string]*EventTracker
	generation  map[string]int64
	observers   []ResourceChangeObserver
}

// ManagedResource represents a resource with full lifecycle management
type ManagedResource struct {
	APIVersion  string                 `json:"apiVersion"`
	Kind        string                 `json:"kind"`
	Metadata    ResourceMetadata        `json:"metadata"`
	Spec        map[string]interface{} `json:"spec"`
	Status      ResourceStatus         `json:"status"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	DeletedAt   *time.Time             `json:"deletedAt,omitempty"`
	Finalizers  []string               `json:"finalizers"`
	OwnerRefs   []OwnerReference       `json:"ownerReferences,omitempty"`
	Conditions  []ResourceCondition    `json:"conditions,omitempty"`
	Generation  int64                  `json:"generation"`
	ObservedGen int64                  `json:"observedGeneration"`
}

// ResourceMetadata contains metadata about a resource
type ResourceMetadata struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	UID        string            `json:"uid"`
	Labels     map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Finalizers []string          `json:"finalizers"`
}

// ResourceStatus represents resource status with conditions
type ResourceStatus struct {
	Phase      string                  `json:"phase"` // Pending, Active, Terminating, Failed
	Conditions []ResourceCondition     `json:"conditions"`
	LastUpdate time.Time               `json:"lastUpdate"`
	Message    string                  `json:"message,omitempty"`
	Details    map[string]interface{}  `json:"details,omitempty"`
}

// ResourceCondition represents a condition on a resource
type ResourceCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"` // True, False, Unknown
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
	ObservedGeneration int64     `json:"observedGeneration"`
	LastTransition     time.Time `json:"lastTransitionTime"`
}

// OwnerReference represents an ownership relationship
type OwnerReference struct {
	APIVersion         string `json:"apiVersion"`
	Kind               string `json:"kind"`
	Name               string `json:"name"`
	UID                string `json:"uid"`
	BlockOwnerDeletion bool   `json:"blockOwnerDeletion,omitempty"`
}

// ResourceChangeObserver observes resource changes
type ResourceChangeObserver interface {
	OnCreated(ctx context.Context, resource *ManagedResource)
	OnUpdated(ctx context.Context, oldResource, newResource *ManagedResource)
	OnDeleted(ctx context.Context, resource *ManagedResource)
	OnStatusChanged(ctx context.Context, resource *ManagedResource, oldStatus, newStatus ResourceStatus)
}

// EventTracker tracks events for a resource
type EventTracker struct {
	mu     sync.RWMutex
	events []*ResourceEvent
}

// ResourceEvent represents an event
type ResourceEvent struct {
	Type    string    // Normal, Warning
	Reason  string
	Message string
	Time    time.Time
	Count   int
}

// NewResourceManager creates a new resource manager
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources:  make(map[string]map[string]*ManagedResource),
		finalizers: make(map[string]*FinalizerManager),
		conditions: make(map[string]*ConditionManager),
		events:     make(map[string]*EventTracker),
		generation: make(map[string]int64),
		observers:  make([]ResourceChangeObserver, 0),
	}
}

// Create creates a new resource
func (rm *ResourceManager) Create(ctx context.Context, resource *ManagedResource) (*ManagedResource, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Validate resource
	if resource.Metadata.Name == "" || resource.Kind == "" {
		return nil, fmt.Errorf("resource name and kind are required")
	}

	namespace := resource.Metadata.Namespace
	if namespace == "" {
		namespace = "default"
		resource.Metadata.Namespace = namespace
	}

	// Initialize namespace map if needed
	if _, exists := rm.resources[namespace]; !exists {
		rm.resources[namespace] = make(map[string]*ManagedResource)
	}

	key := resource.Metadata.Name
	if _, exists := rm.resources[namespace][key]; exists {
		return nil, fmt.Errorf("resource %s already exists", key)
	}

	// Set defaults
	resource.CreatedAt = time.Now()
	resource.UpdatedAt = time.Now()
	resource.Generation = 1
	resource.ObservedGen = 0

	// Set status to pending
	if resource.Status.Phase == "" {
		resource.Status.Phase = "Pending"
	}
	resource.Status.LastUpdate = time.Now()

	// Store resource
	rm.resources[namespace][key] = resource

	// Initialize managers
	resourceID := namespace + "/" + resource.Kind + "/" + key
	rm.finalizers[resourceID] = NewFinalizerManager()
	rm.conditions[resourceID] = NewConditionManager()
	rm.events[resourceID] = &EventTracker{events: make([]*ResourceEvent, 0)}
	rm.generation[resourceID] = 1

	// Notify observers
	for _, observer := range rm.observers {
		go observer.OnCreated(ctx, resource)
	}

	return resource, nil
}

// Get retrieves a resource
func (rm *ResourceManager) Get(ctx context.Context, namespace, kind, name string) (*ManagedResource, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if namespace == "" {
		namespace = "default"
	}

	nsMap, exists := rm.resources[namespace]
	if !exists {
		return nil, fmt.Errorf("namespace %s not found", namespace)
	}

	resource, exists := nsMap[name]
	if !exists {
		return nil, fmt.Errorf("resource %s not found", name)
	}

	if resource.Kind != kind {
		return nil, fmt.Errorf("resource is of kind %s, not %s", resource.Kind, kind)
	}

	return resource, nil
}

// Update updates a resource spec
func (rm *ResourceManager) Update(ctx context.Context, resource *ManagedResource) (*ManagedResource, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	namespace := resource.Metadata.Namespace
	if namespace == "" {
		namespace = "default"
	}

	existing, err := rm.getUnsafe(namespace, resource.Kind, resource.Metadata.Name)
	if err != nil {
		return nil, err
	}

	oldResource := *existing
	resource.CreatedAt = existing.CreatedAt
	resource.UpdatedAt = time.Now()
	resource.Generation = existing.Generation + 1
	resource.ObservedGen = existing.ObservedGen

	rm.resources[namespace][resource.Metadata.Name] = resource

	resourceID := namespace + "/" + resource.Kind + "/" + resource.Metadata.Name
	rm.generation[resourceID] = resource.Generation

	// Notify observers
	for _, observer := range rm.observers {
		go observer.OnUpdated(ctx, &oldResource, resource)
	}

	return resource, nil
}

// UpdateStatus updates resource status
func (rm *ResourceManager) UpdateStatus(ctx context.Context, namespace, kind, name string, status ResourceStatus) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if namespace == "" {
		namespace = "default"
	}

	resource, err := rm.getUnsafe(namespace, kind, name)
	if err != nil {
		return err
	}

	oldStatus := resource.Status
	resource.Status = status
	resource.Status.LastUpdate = time.Now()
	resource.UpdatedAt = time.Now()

	// Notify observers
	for _, observer := range rm.observers {
		go observer.OnStatusChanged(ctx, resource, oldStatus, status)
	}

	return nil
}

// Delete deletes a resource
func (rm *ResourceManager) Delete(ctx context.Context, namespace, kind, name string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if namespace == "" {
		namespace = "default"
	}

	resource, err := rm.getUnsafe(namespace, kind, name)
	if err != nil {
		return err
	}

	now := time.Now()
	resource.DeletedAt = &now
	resource.Status.Phase = "Terminating"

	// Notify observers
	for _, observer := range rm.observers {
		go observer.OnDeleted(ctx, resource)
	}

	// Run finalizers
	resourceID := namespace + "/" + kind + "/" + name
	if fm, exists := rm.finalizers[resourceID]; exists {
		_ = fm.RunFinalizers(ctx, resource)
	}

	// Clean up
	delete(rm.resources[namespace], name)
	delete(rm.finalizers, resourceID)
	delete(rm.conditions, resourceID)
	delete(rm.events, resourceID)
	delete(rm.generation, resourceID)

	return nil
}

// List lists resources by kind and labels
func (rm *ResourceManager) List(ctx context.Context, namespace, kind string, labels map[string]string) ([]*ManagedResource, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if namespace == "" {
		namespace = "default"
	}

	nsMap, exists := rm.resources[namespace]
	if !exists {
		return nil, nil
	}

	var results []*ManagedResource
	for _, resource := range nsMap {
		if resource.Kind != kind {
			continue
		}

		// Check label match
		if len(labels) > 0 {
			matched := true
			for k, v := range labels {
				if resource.Metadata.Labels[k] != v {
					matched = false
					break
				}
			}
			if !matched {
				continue
			}
		}

		results = append(results, resource)
	}

	return results, nil
}

// RegisterObserver registers a change observer
func (rm *ResourceManager) RegisterObserver(observer ResourceChangeObserver) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.observers = append(rm.observers, observer)
}

// GetGeneration returns current generation of a resource
func (rm *ResourceManager) GetGeneration(ctx context.Context, namespace, kind, name string) (int64, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if namespace == "" {
		namespace = "default"
	}

	resourceID := namespace + "/" + kind + "/" + name
	gen, exists := rm.generation[resourceID]
	if !exists {
		return 0, fmt.Errorf("resource not found")
	}
	return gen, nil
}

// RecordEvent records an event for a resource
func (rm *ResourceManager) RecordEvent(ctx context.Context, namespace, kind, name, eventType, reason, message string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if namespace == "" {
		namespace = "default"
	}

	resourceID := namespace + "/" + kind + "/" + name
	tracker, exists := rm.events[resourceID]
	if !exists {
		tracker = &EventTracker{events: make([]*ResourceEvent, 0)}
		rm.events[resourceID] = tracker
	}

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	// Find existing event with same reason
	for _, e := range tracker.events {
		if e.Reason == reason {
			e.Count++
			e.Time = time.Now()
			return
		}
	}

	// Add new event
	tracker.events = append(tracker.events, &ResourceEvent{
		Type:    eventType,
		Reason:  reason,
		Message: message,
		Time:    time.Now(),
		Count:   1,
	})

	// Keep only last 100 events
	if len(tracker.events) > 100 {
		tracker.events = tracker.events[len(tracker.events)-100:]
	}
}

// GetEvents returns events for a resource
func (rm *ResourceManager) GetEvents(ctx context.Context, namespace, kind, name string) ([]*ResourceEvent, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if namespace == "" {
		namespace = "default"
	}

	resourceID := namespace + "/" + kind + "/" + name
	tracker, exists := rm.events[resourceID]
	if !exists {
		return nil, nil
	}

	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	events := make([]*ResourceEvent, len(tracker.events))
	copy(events, tracker.events)
	return events, nil
}

// AddFinalizer adds a finalizer to a resource
func (rm *ResourceManager) AddFinalizer(ctx context.Context, namespace, kind, name, finalizerName string, handler func(context.Context, *ManagedResource) error) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if namespace == "" {
		namespace = "default"
	}

	resourceID := namespace + "/" + kind + "/" + name

	fm, exists := rm.finalizers[resourceID]
	if !exists {
		fm = NewFinalizerManager()
		rm.finalizers[resourceID] = fm
	}

	fm.AddFinalizer(name, Finalizer{
		Name:    finalizerName,
		Handler: handler,
	})

	return nil
}

// getUnsafe gets a resource without lock (must be called with lock held)
func (rm *ResourceManager) getUnsafe(namespace, kind, name string) (*ManagedResource, error) {
	nsMap, exists := rm.resources[namespace]
	if !exists {
		return nil, fmt.Errorf("namespace %s not found", namespace)
	}

	resource, exists := nsMap[name]
	if !exists {
		return nil, fmt.Errorf("resource %s not found", name)
	}

	if resource.Kind != kind {
		return nil, fmt.Errorf("resource is of kind %s, not %s", resource.Kind, kind)
	}

	return resource, nil
}

// ConditionManager manages conditions on a resource
type ConditionManager struct {
	mu         sync.RWMutex
	conditions map[string]*ResourceCondition
}

// NewConditionManager creates a new condition manager
func NewConditionManager() *ConditionManager {
	return &ConditionManager{
		conditions: make(map[string]*ResourceCondition),
	}
}

// SetCondition sets a condition
func (cm *ConditionManager) SetCondition(condType, status, reason, message string, generation int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.conditions[condType] = &ResourceCondition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: generation,
		LastTransition:     time.Now(),
	}
}

// GetCondition gets a condition
func (cm *ConditionManager) GetCondition(condType string) *ResourceCondition {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.conditions[condType]
}

// GetConditions returns all conditions
func (cm *ConditionManager) GetConditions() []ResourceCondition {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conditions := make([]ResourceCondition, 0, len(cm.conditions))
	for _, cond := range cm.conditions {
		conditions = append(conditions, *cond)
	}
	return conditions
}

// IsReady returns true if all required conditions are met
func (cm *ConditionManager) IsReady() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if len(cm.conditions) == 0 {
		return false
	}

	for _, cond := range cm.conditions {
		if cond.Status != "True" {
			return false
		}
	}
	return true
}
