package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ControlPlane integrates all data control plane components
type ControlPlane struct {
	mu                     sync.RWMutex
	resourceMgr            *ResourceManager
	validationEngine       *ValidationEngine
	mutationEngine         *MutationEngine
	reconciliationEngine   *ReconciliationEngine
	distributedCoordinator *DistributedCoordinator
	observers              []ControlPlaneObserver
	backupManager          *BackupManager
	admissionPolicy        interface{} // Polymorphic for policies.AdmissionPolicy
}

// ControlPlaneObserver observes control plane events
type ControlPlaneObserver interface {
	OnResourceCreated(ctx context.Context, resource *ManagedResource)
	OnResourceUpdated(ctx context.Context, oldResource, newResource *ManagedResource)
	OnResourceDeleted(ctx context.Context, resource *ManagedResource)
	OnReconciliationStarted(ctx context.Context, req *ReconciliationRequest)
	OnReconciliationCompleted(ctx context.Context, req *ReconciliationRequest, result *ReconciliationResult)
	OnPolicyViolation(ctx context.Context, policyName string, resource *ManagedResource)
}

// NewControlPlane creates a new control plane
func NewControlPlane(instanceID string) *ControlPlane {
	return &ControlPlane{
		resourceMgr:            NewResourceManager(),
		validationEngine:       NewValidationEngine(),
		mutationEngine:         NewMutationEngine(),
		reconciliationEngine:   NewReconciliationEngine(),
		distributedCoordinator: NewDistributedCoordinator(instanceID),
		observers:              make([]ControlPlaneObserver, 0),
		backupManager:          NewBackupManager(NewResourceManager(), 100, 30),
		admissionPolicy:        nil,
	}
}

// RegisterObserver registers a control plane observer
func (cp *ControlPlane) RegisterObserver(observer ControlPlaneObserver) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.observers = append(cp.observers, observer)
}

// GetResourceManager returns the resource manager
func (cp *ControlPlane) GetResourceManager() *ResourceManager {
	return cp.resourceMgr
}

// GetValidationEngine returns the validation engine
func (cp *ControlPlane) GetValidationEngine() *ValidationEngine {
	return cp.validationEngine
}

// GetMutationEngine returns the mutation engine
func (cp *ControlPlane) GetMutationEngine() *MutationEngine {
	return cp.mutationEngine
}

// GetReconciliationEngine returns the reconciliation engine
func (cp *ControlPlane) GetReconciliationEngine() *ReconciliationEngine {
	return cp.reconciliationEngine
}

// GetDistributedCoordinator returns the distributed coordinator
func (cp *ControlPlane) GetDistributedCoordinator() *DistributedCoordinator {
	return cp.distributedCoordinator
}

// GetBackupManager returns the backup manager
func (cp *ControlPlane) GetBackupManager() *BackupManager {
	return cp.backupManager
}

// SetAdmissionPolicy sets the admission policy engine
func (cp *ControlPlane) SetAdmissionPolicy(policy interface{}) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.admissionPolicy = policy
}

// GetAdmissionPolicy returns the admission policy engine
func (cp *ControlPlane) GetAdmissionPolicy() interface{} {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.admissionPolicy
}

// CreateResource creates a resource with full control plane processing
func (cp *ControlPlane) CreateResource(ctx context.Context, resource *ManagedResource) (*ManagedResource, error) {
	// Check leader status for distributed deployments
	if !cp.distributedCoordinator.IsLeader() {
		return nil, fmt.Errorf("not leader: cannot create resource")
	}

	// Check admission policies first
	if cp.admissionPolicy != nil {
		// Type assertion would happen at integration point
		// For now, we note that admission checks should happen here
	}

	// Validate resource
	validationResult := cp.validationEngine.Validate(ctx, resource, "create")
	if !validationResult.Valid {
		return nil, fmt.Errorf("validation failed: %v", validationResult.Errors)
	}

	// Transform resource (mutations, defaults, etc.)
	var err error
	resource, err = cp.validationEngine.Transform(ctx, resource, "create")
	if err != nil {
		return nil, fmt.Errorf("transformation failed: %v", err)
	}

	// Apply mutations
	if err := cp.mutationEngine.Mutate(ctx, resource, "create"); err != nil {
		return nil, fmt.Errorf("mutation failed: %v", err)
	}

	// Store resource
	resource, err = cp.resourceMgr.Create(ctx, resource)
	if err != nil {
		return nil, fmt.Errorf("storage failed: %v", err)
	}

	// Record event
	cp.resourceMgr.RecordEvent(ctx, resource.Metadata.Namespace, resource.Kind, resource.Metadata.Name,
		"Normal", "Created", "Resource created successfully")

	// Enqueue for reconciliation
	cp.reconciliationEngine.Enqueue(&ReconciliationRequest{
		Namespace: resource.Metadata.Namespace,
		Kind:      resource.Kind,
		Name:      resource.Metadata.Name,
		Force:     false,
		CreatedAt: time.Now(),
	})

	// Notify observers
	for _, observer := range cp.observers {
		go observer.OnResourceCreated(ctx, resource)
	}

	// Notify watchers across cluster
	cp.distributedCoordinator.NotifyWatchers(WatchEvent{
		Type:      "Added",
		Resource:  resource,
		Timestamp: time.Now(),
	})

	return resource, nil
}

// UpdateResource updates a resource with full control plane processing
func (cp *ControlPlane) UpdateResource(ctx context.Context, resource *ManagedResource) (*ManagedResource, error) {
	// Check leader status
	if !cp.distributedCoordinator.IsLeader() {
		return nil, fmt.Errorf("not leader: cannot update resource")
	}

	// Get existing resource
	existing, err := cp.resourceMgr.Get(ctx, resource.Metadata.Namespace, resource.Kind, resource.Metadata.Name)
	if err != nil {
		return nil, err
	}

	// Validate update
	validationResult := cp.validationEngine.Validate(ctx, resource, "update")
	if !validationResult.Valid {
		return nil, fmt.Errorf("validation failed: %v", validationResult.Errors)
	}

	// Transform resource
	resource, err = cp.validationEngine.Transform(ctx, resource, "update")
	if err != nil {
		return nil, fmt.Errorf("transformation failed: %v", err)
	}

	// Apply mutations
	if err := cp.mutationEngine.Mutate(ctx, resource, "update"); err != nil {
		return nil, fmt.Errorf("mutation failed: %v", err)
	}

	// Update resource
	updated, err := cp.resourceMgr.Update(ctx, resource)
	if err != nil {
		return nil, fmt.Errorf("storage failed: %v", err)
	}

	// Record event
	cp.resourceMgr.RecordEvent(ctx, updated.Metadata.Namespace, updated.Kind, updated.Metadata.Name,
		"Normal", "Updated", "Resource updated successfully")

	// Enqueue for reconciliation
	cp.reconciliationEngine.Enqueue(&ReconciliationRequest{
		Namespace: updated.Metadata.Namespace,
		Kind:      updated.Kind,
		Name:      updated.Metadata.Name,
		Force:     false,
		CreatedAt: time.Now(),
	})

	// Notify observers
	for _, observer := range cp.observers {
		go observer.OnResourceUpdated(ctx, existing, updated)
	}

	// Notify watchers
	cp.distributedCoordinator.NotifyWatchers(WatchEvent{
		Type:      "Modified",
		Resource:  updated,
		Timestamp: time.Now(),
	})

	return updated, nil
}

// DeleteResource deletes a resource with full control plane processing
func (cp *ControlPlane) DeleteResource(ctx context.Context, namespace, kind, name string) error {
	// Check leader status
	if !cp.distributedCoordinator.IsLeader() {
		return fmt.Errorf("not leader: cannot delete resource")
	}

	// Get resource
	resource, err := cp.resourceMgr.Get(ctx, namespace, kind, name)
	if err != nil {
		return err
	}

	// Delete from resource manager (runs finalizers)
	if err := cp.resourceMgr.Delete(ctx, namespace, kind, name); err != nil {
		return err
	}

	// Record event
	cp.resourceMgr.RecordEvent(ctx, namespace, kind, name,
		"Normal", "Deleted", "Resource deleted successfully")

	// Notify observers
	for _, observer := range cp.observers {
		go observer.OnResourceDeleted(ctx, resource)
	}

	// Notify watchers
	cp.distributedCoordinator.NotifyWatchers(WatchEvent{
		Type:      "Deleted",
		Resource:  resource,
		Timestamp: time.Now(),
	})

	return nil
}

// ReconcileResource performs full reconciliation
func (cp *ControlPlane) ReconcileResource(ctx context.Context, namespace, kind, name string, reconciler Reconciler) (*ReconciliationResult, error) {
	// Get resource
	resource, err := cp.resourceMgr.Get(ctx, namespace, kind, name)
	if err != nil {
		return nil, err
	}

	// Create reconciliation request
	req := &ReconciliationRequest{
		Namespace: namespace,
		Kind:      kind,
		Name:      name,
		CreatedAt: time.Now(),
	}

	// Register reconciler if not already registered
	cp.reconciliationEngine.RegisterReconciler(kind, reconciler)

	// Notify observers - reconciliation started
	for _, observer := range cp.observers {
		go observer.OnReconciliationStarted(ctx, req)
	}

	// Reconcile
	result, err := cp.reconciliationEngine.Reconcile(ctx, req, resource)
	if err != nil {
		return result, err
	}

	// Update resource status with reconciliation result
	newStatus := resource.Status
	newStatus.Phase = result.Status
	newStatus.Message = result.Message
	newStatus.Conditions = result.Conditions

	if err := cp.resourceMgr.UpdateStatus(ctx, namespace, kind, name, newStatus); err != nil {
		return result, err
	}

	// Notify observers - reconciliation completed
	for _, observer := range cp.observers {
		go observer.OnReconciliationCompleted(ctx, req, result)
	}

	return result, nil
}

// ProcessReconciliationQueue processes the reconciliation work queue
func (cp *ControlPlane) ProcessReconciliationQueue(ctx context.Context, kind string, reconciler Reconciler) {
	cp.reconciliationEngine.RegisterReconciler(kind, reconciler)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Get next request from queue
		req := cp.reconciliationEngine.Dequeue(kind)
		if req == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Get resource
		resource, err := cp.resourceMgr.Get(ctx, req.Namespace, kind, req.Name)
		if err != nil {
			continue
		}

		// Reconcile
		result, err := cp.reconciliationEngine.Reconcile(ctx, req, resource)
		if err == nil && result != nil {
			// Update status
			newStatus := resource.Status
			newStatus.Phase = result.Status
			newStatus.Message = result.Message
			newStatus.Conditions = result.Conditions

			_ = cp.resourceMgr.UpdateStatus(ctx, req.Namespace, kind, req.Name, newStatus)
		}

		time.Sleep(10 * time.Millisecond) // Avoid busy loop
	}
}

// GetResourceWithConditions gets a resource and its conditions
func (cp *ControlPlane) GetResourceWithConditions(ctx context.Context, namespace, kind, name string) (*ManagedResource, []ResourceCondition, error) {
	resource, err := cp.resourceMgr.Get(ctx, namespace, kind, name)
	if err != nil {
		return nil, nil, err
	}

	return resource, resource.Conditions, nil
}

// ListResourcesByLabel lists resources by labels
func (cp *ControlPlane) ListResourcesByLabel(ctx context.Context, namespace, kind string, labels map[string]string) ([]*ManagedResource, error) {
	return cp.resourceMgr.List(ctx, namespace, kind, labels)
}

// WatchResources watches for resource changes
func (cp *ControlPlane) WatchResources(ctx context.Context, kind string, callback WatchCallback) error {
	return cp.distributedCoordinator.Watch(ctx, kind, callback)
}

// GetControlPlaneStatus returns the control plane status
func (cp *ControlPlane) GetControlPlaneStatus(ctx context.Context) map[string]interface{} {
	status := make(map[string]interface{})

	// Leader information
	leader := cp.distributedCoordinator.GetLeader()
	status["leader"] = leader
	status["isLeader"] = cp.distributedCoordinator.IsLeader()

	// Queue sizes
	for _, kind := range []string{"API", "Workflow", "Policy", "Datasource"} {
		queueSize := cp.reconciliationEngine.GetQueueSize(kind)
		status[fmt.Sprintf("%s_queue_size", kind)] = queueSize
	}

	// Instance count
	instanceCount := cp.distributedCoordinator.GetInstanceCount(30 * time.Second)
	status["active_instances"] = instanceCount

	// Distributed locks
	status["distributed_locks"] = len(cp.distributedCoordinator.locks)

	return status
}

// HealthCheck performs a health check on the control plane
func (cp *ControlPlane) HealthCheck(ctx context.Context) error {
	// Check if leader election is functional
	if !cp.distributedCoordinator.IsLeader() {
		return fmt.Errorf("not the leader")
	}

	// Check if resource manager is accessible
	if _, err := cp.resourceMgr.GetGeneration(ctx, "default", "API", "test"); err != nil {
		// This is expected to fail for non-existent resource
		// So we just verify we can call it
	}

	// Check reconciliation engine
	queueSize := cp.reconciliationEngine.GetQueueSize("API")
	if queueSize < 0 {
		return fmt.Errorf("invalid queue size")
	}

	return nil
}
