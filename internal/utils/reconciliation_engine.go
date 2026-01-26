package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ReconciliationEngine manages controller reconciliation loops
type ReconciliationEngine struct {
	mu          sync.RWMutex
	reconcilers map[string]Reconciler
	queues      map[string]*ReconciliationQueue
	results     map[string]*ReconciliationResult
}

// Reconciler implements resource reconciliation
type Reconciler interface {
	Reconcile(ctx context.Context, resource *ManagedResource) (ReconciliationResult, error)
}

// ReconciliationRequest queues a resource for reconciliation
type ReconciliationRequest struct {
	Namespace string
	Kind      string
	Name      string
	Force     bool
	Retries   int
	CreatedAt time.Time
}

// ReconciliationResult indicates reconciliation result
type ReconciliationResult struct {
	Status           string        // Success, Pending, Failed, Retrying
	ObservedGeneration int64        // Generation this result applies to
	LastReconciled   time.Time
	Requeue          bool
	RequeueAfter     time.Duration
	Error            string
	Message          string
	Conditions       []ResourceCondition
	Details          map[string]interface{}
}

// ReconciliationQueue holds pending reconciliation requests
type ReconciliationQueue struct {
	mu       sync.Mutex
	requests []*ReconciliationRequest
	working  map[string]bool
}

// NewReconciliationEngine creates a new reconciliation engine
func NewReconciliationEngine() *ReconciliationEngine {
	return &ReconciliationEngine{
		reconcilers: make(map[string]Reconciler),
		queues:      make(map[string]*ReconciliationQueue),
		results:     make(map[string]*ReconciliationResult),
	}
}

// RegisterReconciler registers a reconciler for a resource kind
func (re *ReconciliationEngine) RegisterReconciler(kind string, reconciler Reconciler) {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.reconcilers[kind] = reconciler
	re.queues[kind] = &ReconciliationQueue{
		requests: make([]*ReconciliationRequest, 0),
		working:  make(map[string]bool),
	}
}

// Enqueue enqueues a reconciliation request
func (re *ReconciliationEngine) Enqueue(req *ReconciliationRequest) error {
	re.mu.RLock()
	queue, exists := re.queues[req.Kind]
	re.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no reconciler registered for kind %s", req.Kind)
	}

	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now()
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	// Check if already working on this resource
	key := req.Namespace + "/" + req.Name
	if queue.working[key] && !req.Force {
		return fmt.Errorf("already reconciling %s", key)
	}

	queue.requests = append(queue.requests, req)
	return nil
}

// Dequeue dequeues a reconciliation request
func (re *ReconciliationEngine) Dequeue(kind string) *ReconciliationRequest {
	re.mu.RLock()
	queue, exists := re.queues[kind]
	re.mu.RUnlock()

	if !exists {
		return nil
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	if len(queue.requests) == 0 {
		return nil
	}

	req := queue.requests[0]
	queue.requests = queue.requests[1:]

	key := req.Namespace + "/" + req.Name
	queue.working[key] = true

	return req
}

// Reconcile reconciles a resource
func (re *ReconciliationEngine) Reconcile(ctx context.Context, req *ReconciliationRequest, resource *ManagedResource) (*ReconciliationResult, error) {
	re.mu.RLock()
	reconciler, exists := re.reconcilers[req.Kind]
	re.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no reconciler registered for kind %s", req.Kind)
	}

	// Create reconciliation context with timeout
	recCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Run reconciler
	result, err := reconciler.Reconcile(recCtx, resource)
	if err != nil {
		result.Error = err.Error()
		result.Status = "Failed"

		// Check if retryable
		if req.Retries < 3 {
			result.Status = "Retrying"
			result.Requeue = true
			result.RequeueAfter = time.Duration((req.Retries+1)*5) * time.Second
			req.Retries++

			// Re-enqueue for retry
			re.Enqueue(req)
		}
	} else {
		result.Status = "Success"
	}

	result.ObservedGeneration = resource.Generation
	result.LastReconciled = time.Now()

	// Store result
	re.mu.Lock()
	defer re.mu.Unlock()

	resultKey := req.Namespace + "/" + req.Kind + "/" + req.Name
	re.results[resultKey] = &result

	// Mark as done
	queue := re.queues[req.Kind]
	queue.mu.Lock()
	key := req.Namespace + "/" + req.Name
	delete(queue.working, key)
	queue.mu.Unlock()

	return &result, nil
}

// GetResult gets the last reconciliation result
func (re *ReconciliationEngine) GetResult(namespace, kind, name string) *ReconciliationResult {
	re.mu.RLock()
	defer re.mu.RUnlock()

	resultKey := namespace + "/" + kind + "/" + name
	return re.results[resultKey]
}

// GetQueueSize gets the queue size for a kind
func (re *ReconciliationEngine) GetQueueSize(kind string) int {
	re.mu.RLock()
	queue, exists := re.queues[kind]
	re.mu.RUnlock()

	if !exists {
		return 0
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	return len(queue.requests)
}

// ControllerReconciler provides a default reconciliation pattern
type ControllerReconciler struct {
	resourceMgr *ResourceManager
	handlers    map[string]ReconciliationHandler
}

// ReconciliationHandler handles reconciliation for a phase
type ReconciliationHandler func(ctx context.Context, resource *ManagedResource) error

// NewControllerReconciler creates a new controller reconciler
func NewControllerReconciler(resourceMgr *ResourceManager) *ControllerReconciler {
	return &ControllerReconciler{
		resourceMgr: resourceMgr,
		handlers:    make(map[string]ReconciliationHandler),
	}
}

// RegisterPhaseHandler registers a handler for a phase
func (cr *ControllerReconciler) RegisterPhaseHandler(phase string, handler ReconciliationHandler) {
	cr.handlers[phase] = handler
}

// Reconcile reconciles a resource
func (cr *ControllerReconciler) Reconcile(ctx context.Context, resource *ManagedResource) (ReconciliationResult, error) {
	result := ReconciliationResult{
		Details: make(map[string]interface{}),
	}

	// Get current phase
	phase := resource.Status.Phase
	if phase == "" {
		phase = "Pending"
	}

	// Handle deletion
	if resource.DeletedAt != nil {
		if handler, exists := cr.handlers["Terminating"]; exists {
			if err := handler(ctx, resource); err != nil {
				result.Error = err.Error()
				return result, err
			}
		}
		phase = "Deleted"
	} else if handler, exists := cr.handlers[phase]; exists {
		// Handle normal reconciliation
		if err := handler(ctx, resource); err != nil {
			result.Error = err.Error()

			// Update status to failed
			resource.Status.Phase = "Failed"
			resource.Status.Message = err.Error()

			return result, err
		}

		// Transition to next phase or complete
		if phase == "Pending" {
			phase = "Active"
		}
	}

	// Update resource status
	resource.Status.Phase = phase
	resource.Status.LastUpdate = time.Now()

	// Store updated generation
	resource.ObservedGen = resource.Generation

	result.Message = fmt.Sprintf("Reconciliation completed in phase %s", phase)
	result.Details["phase"] = phase
	result.Details["generation"] = resource.Generation

	return result, nil
}

// WorkqueueReconciler manages a work queue for reconciliation
type WorkqueueReconciler struct {
	mu           sync.RWMutex
	queue        []*ReconciliationRequest
	rateLimiter  RateLimiter
	maxRetries   int
	defaultRequeueInterval time.Duration
}

// RateLimiter limits reconciliation rate
type RateLimiter interface {
	When(key string) time.Duration
	Forget(key string)
	NumRequeues(key string) int
}

// NewWorkqueueReconciler creates a new workqueue reconciler
func NewWorkqueueReconciler(maxRetries int, defaultRequeueInterval time.Duration) *WorkqueueReconciler {
	return &WorkqueueReconciler{
		queue:                  make([]*ReconciliationRequest, 0),
		maxRetries:             maxRetries,
		defaultRequeueInterval: defaultRequeueInterval,
	}
}

// AddToWorkqueue adds a resource to the work queue
func (wr *WorkqueueReconciler) AddToWorkqueue(namespace, kind, name string) {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	req := &ReconciliationRequest{
		Namespace: namespace,
		Kind:      kind,
		Name:      name,
		CreatedAt: time.Now(),
	}

	wr.queue = append(wr.queue, req)
}

// Get gets the next item from the work queue
func (wr *WorkqueueReconciler) Get() *ReconciliationRequest {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	if len(wr.queue) == 0 {
		return nil
	}

	item := wr.queue[0]
	wr.queue = wr.queue[1:]

	return item
}

// Requeue requeues an item
func (wr *WorkqueueReconciler) Requeue(req *ReconciliationRequest) {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	if req.Retries < wr.maxRetries {
		req.Retries++
		wr.queue = append(wr.queue, req)
	}
}

// GetQueueLength returns the queue length
func (wr *WorkqueueReconciler) GetQueueLength() int {
	wr.mu.RLock()
	defer wr.mu.RUnlock()

	return len(wr.queue)
}

// StatusUpdateReconciler manages status subresource updates
type StatusUpdateReconciler struct {
	mu          sync.RWMutex
	pendingUpdates map[string]*ResourceStatus
}

// NewStatusUpdateReconciler creates a new status update reconciler
func NewStatusUpdateReconciler() *StatusUpdateReconciler {
	return &StatusUpdateReconciler{
		pendingUpdates: make(map[string]*ResourceStatus),
	}
}

// QueueStatusUpdate queues a status update
func (sur *StatusUpdateReconciler) QueueStatusUpdate(namespace, kind, name string, status ResourceStatus) {
	sur.mu.Lock()
	defer sur.mu.Unlock()

	key := namespace + "/" + kind + "/" + name
	sur.pendingUpdates[key] = &status
}

// GetPendingStatusUpdate gets a pending status update
func (sur *StatusUpdateReconciler) GetPendingStatusUpdate(namespace, kind, name string) *ResourceStatus {
	sur.mu.Lock()
	defer sur.mu.Unlock()

	key := namespace + "/" + kind + "/" + name
	status, exists := sur.pendingUpdates[key]
	if exists {
		delete(sur.pendingUpdates, key)
	}
	return status
}

// GetPendingCount returns count of pending updates
func (sur *StatusUpdateReconciler) GetPendingCount() int {
	sur.mu.RLock()
	defer sur.mu.RUnlock()

	return len(sur.pendingUpdates)
}
