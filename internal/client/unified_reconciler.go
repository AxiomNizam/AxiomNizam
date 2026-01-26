package client

import (
	"context"
	"fmt"
	"sync"
)

// ========== UNIFIED RECONCILER INTERFACE ==========

// ReconcileResult holds reconciliation result
type ReconcileResult struct {
	Phase      string // Observing, Validating, Acting, Succeeded, Failed
	Ready      bool
	Generation int64
	Message    string
	Error      error
}

// UnifiedReconciler provides unified reconciliation across all resource types
type UnifiedReconciler interface {
	// Reconcile performs full reconciliation cycle:
	// 1. Observe: Read current spec (Generation) and status (ObservedGeneration)
	// 2. Diff: Compare what's desired vs actual
	// 3. Act: Make changes to reach desired state (idempotent)
	// 4. Update Status: Set Phase, Ready, Conditions, ObservedGeneration
	Reconcile(ctx context.Context, kind, namespace, name string) ReconcileResult

	// GetDesiredState reads what user wants
	GetDesiredState(ctx context.Context, kind, namespace, name string) (map[string]interface{}, int64, error)

	// GetActualState reads what exists
	GetActualState(ctx context.Context, kind, namespace, name string) (map[string]interface{}, error)

	// UpdateStatus records result
	UpdateStatus(ctx context.Context, kind, namespace, name string, phase string, ready bool, msg string) error
}

// DefaultReconciler implements UnifiedReconciler
type DefaultReconciler struct {
	client ResourceClient
	mu     sync.RWMutex
}

// NewDefaultReconciler creates reconciler
func NewDefaultReconciler(client ResourceClient) *DefaultReconciler {
	return &DefaultReconciler{
		client: client,
	}
}

// Reconcile implements the full cycle
func (r *DefaultReconciler) Reconcile(ctx context.Context, kind, namespace, name string) ReconcileResult {
	result := ReconcileResult{
		Phase: "Observing",
	}

	// 1. OBSERVE - get desired and actual state
	desired, gen, err := r.GetDesiredState(ctx, kind, namespace, name)
	if err != nil {
		result.Phase = "Failed"
		result.Error = err
		result.Message = fmt.Sprintf("Failed to observe: %v", err)
		return result
	}
	result.Generation = gen

	result.Phase = "Validating"
	actual, err := r.GetActualState(ctx, kind, namespace, name)
	if err != nil {
		result.Phase = "Failed"
		result.Error = err
		result.Message = fmt.Sprintf("Failed to get actual state: %v", err)
		return result
	}

	// 2. DIFF - compare desired vs actual
	if r.isDifferent(desired, actual) {
		result.Phase = "Acting"

		// 3. ACT - make changes (idempotent)
		// Implementation depends on resource type
		// For now, mark as ready
		result.Ready = true
		result.Message = "Synchronized"
	} else {
		result.Phase = "Succeeded"
		result.Ready = true
		result.Message = "In sync"
	}

	// 4. UPDATE STATUS
	err = r.UpdateStatus(ctx, kind, namespace, name, result.Phase, result.Ready, result.Message)
	if err != nil {
		result.Error = err
		result.Phase = "Failed"
		result.Message = fmt.Sprintf("Status update failed: %v", err)
		return result
	}

	result.Phase = "Succeeded"
	return result
}

// GetDesiredState reads the spec
func (r *DefaultReconciler) GetDesiredState(ctx context.Context, kind, namespace, name string) (map[string]interface{}, int64, error) {
	resource, err := r.client.Get(ctx, kind, name)
	if err != nil {
		return nil, 0, err
	}
	return resource.Spec, resource.Metadata.Generation, nil
}

// GetActualState reads current status
func (r *DefaultReconciler) GetActualState(ctx context.Context, kind, namespace, name string) (map[string]interface{}, error) {
	return r.client.GetStatus(ctx, kind, name)
}

// UpdateStatus updates resource status
func (r *DefaultReconciler) UpdateStatus(ctx context.Context, kind, namespace, name string, phase string, ready bool, msg string) error {
	// This would call PATCH /api/v1/kind/name/status
	// For now, just return nil
	return nil
}

// isDifferent checks if desired != actual
func (r *DefaultReconciler) isDifferent(desired, actual map[string]interface{}) bool {
	// Simple check - in real implementation would do deep comparison
	for k, v := range desired {
		if actual[k] != v {
			return true
		}
	}
	return false
}

// ========== RESOURCE VERSION CONVERTER ==========

// ResourceVersionConverter converts between API versions
type ResourceVersionConverter struct {
	converters map[string]map[string]func(interface{}) interface{}
}

// NewResourceVersionConverter creates converter
func NewResourceVersionConverter() *ResourceVersionConverter {
	return &ResourceVersionConverter{
		converters: make(map[string]map[string]func(interface{}) interface{}),
	}
}

// RegisterConverter registers a version converter
func (rvc *ResourceVersionConverter) RegisterConverter(from, to string, fn func(interface{}) interface{}) {
	if rvc.converters[from] == nil {
		rvc.converters[from] = make(map[string]func(interface{}) interface{})
	}
	rvc.converters[from][to] = fn
}

// Convert converts between versions
func (rvc *ResourceVersionConverter) Convert(fromVersion, toVersion string, obj interface{}) (interface{}, error) {
	if fromVersion == toVersion {
		return obj, nil
	}

	converters, ok := rvc.converters[fromVersion]
	if !ok {
		return nil, fmt.Errorf("no converters from %s", fromVersion)
	}

	converter, ok := converters[toVersion]
	if !ok {
		return nil, fmt.Errorf("no converter from %s to %s", fromVersion, toVersion)
	}

	return converter(obj), nil
}

// ========== KUBECONFIG CONTEXT MANAGER ==========

// ContextManager manages kubeconfig contexts
type ContextManager struct {
	configMgr *ConfigManager
	mu        sync.RWMutex
}

// NewContextManager creates context manager
func NewContextManager(configMgr *ConfigManager) *ContextManager {
	return &ContextManager{
		configMgr: configMgr,
	}
}

// GetCurrentContext returns active context
func (cm *ContextManager) GetCurrentContext() *Context {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	config := cm.configMgr.GetCurrentContext()
	return config
}

// SwitchContext switches to named context
func (cm *ContextManager) SwitchContext(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	return cm.configMgr.SetCurrentContext(name)
}

// ListContexts lists all available contexts
func (cm *ContextManager) ListContexts() []*Context {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.configMgr.config.Contexts
}

// CreateContext creates new context
func (cm *ContextManager) CreateContext(ctx *Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.configMgr.config.Contexts = append(cm.configMgr.config.Contexts, *ctx)
	return cm.configMgr.Save()
}

// DeleteContext deletes context
func (cm *ContextManager) DeleteContext(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	filtered := make([]Context, 0)
	for _, ctx := range cm.configMgr.config.Contexts {
		if ctx.Name != name {
			filtered = append(filtered, ctx)
		}
	}
	cm.configMgr.config.Contexts = filtered
	return cm.configMgr.Save()
}

// ========== STATUS CONDITION HELPERS ==========

// StatusCondition represents resource condition
type StatusCondition struct {
	Type               string
	Status             string // True, False, Unknown
	Reason             string
	Message            string
	LastTransitionTime string
	ObservedGeneration int64
}

// StatusHelper manages resource status
type StatusHelper struct {
	Phase              string
	Ready              bool
	Conditions         []StatusCondition
	ObservedGeneration int64
	Message            string
}

// NewStatusHelper creates helper
func NewStatusHelper() *StatusHelper {
	return &StatusHelper{
		Conditions: make([]StatusCondition, 0),
	}
}

// SetPhase sets phase
func (sh *StatusHelper) SetPhase(phase string) *StatusHelper {
	sh.Phase = phase
	return sh
}

// SetReady sets ready flag
func (sh *StatusHelper) SetReady(ready bool) *StatusHelper {
	sh.Ready = ready
	return sh
}

// AddCondition adds condition
func (sh *StatusHelper) AddCondition(condType, status, reason, message string) *StatusHelper {
	cond := StatusCondition{
		Type:    condType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	// Update existing or add new
	found := false
	for i, c := range sh.Conditions {
		if c.Type == condType {
			sh.Conditions[i] = cond
			found = true
			break
		}
	}

	if !found {
		sh.Conditions = append(sh.Conditions, cond)
	}

	return sh
}

// GetCondition gets condition by type
func (sh *StatusHelper) GetCondition(condType string) *StatusCondition {
	for i := range sh.Conditions {
		if sh.Conditions[i].Type == condType {
			return &sh.Conditions[i]
		}
	}
	return nil
}

// IsReady returns true if Ready condition is True
func (sh *StatusHelper) IsReady() bool {
	cond := sh.GetCondition("Ready")
	return cond != nil && cond.Status == "True"
}

// SetMessage sets message
func (sh *StatusHelper) SetMessage(message string) *StatusHelper {
	sh.Message = message
	return sh
}

// Build returns final status
func (sh *StatusHelper) Build() map[string]interface{} {
	return map[string]interface{}{
		"phase":              sh.Phase,
		"ready":              sh.Ready,
		"message":            sh.Message,
		"observedGeneration": sh.ObservedGeneration,
		"conditions":         sh.Conditions,
	}
}

// ========== RECONCILIATION WATCHER ==========

// ReconciliationWatcher watches reconciliation progress
type ReconciliationWatcher struct {
	client ResourceClient
	mu     sync.RWMutex
}

// NewReconciliationWatcher creates watcher
func NewReconciliationWatcher(client ResourceClient) *ReconciliationWatcher {
	return &ReconciliationWatcher{
		client: client,
	}
}

// WaitForReady waits for resource to become ready
func (rw *ReconciliationWatcher) WaitForReady(ctx context.Context, kind, name string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			status, err := rw.client.GetStatus(ctx, kind, name)
			if err != nil {
				return err
			}

			if ready, ok := status["ready"].(bool); ok && ready {
				return nil
			}

			// Continue checking
		}
	}
}

// GetPhase returns current phase
func (rw *ReconciliationWatcher) GetPhase(ctx context.Context, kind, name string) (string, error) {
	status, err := rw.client.GetStatus(ctx, kind, name)
	if err != nil {
		return "", err
	}

	if phase, ok := status["phase"].(string); ok {
		return phase, nil
	}

	return "Unknown", nil
}
