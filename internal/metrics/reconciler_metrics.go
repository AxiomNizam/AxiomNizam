package metrics

import (
	"sync"
	"time"
)

// ReconcilerStatus represents the operational state of a single reconciler.
type ReconcilerStatus struct {
	Module             string    `json:"module"`
	Initialized        bool      `json:"initialized"`
	Running            bool      `json:"running"`
	ShadowMode         bool      `json:"shadowMode"`
	LastReconcileTime  time.Time `json:"lastReconcileTime,omitempty"`
	LastSuccessTime    time.Time `json:"lastSuccessTime,omitempty"`
	TotalReconciles    int64     `json:"totalReconciles"`
	TotalSuccesses     int64     `json:"totalSuccesses"`
	TotalErrors        int64     `json:"totalErrors"`
	TotalRequeues      int64     `json:"totalRequeues"`
	LastError          string    `json:"lastError,omitempty"`
	AvgDurationMs      float64   `json:"avgDurationMs"`
	LastDurationMs     float64   `json:"lastDurationMs"`
	ConsecutiveErrors  int64     `json:"consecutiveErrors"`
}

// ReconcilerMetrics provides per-module reconciler observability.
// This is the Phase 0 observability layer from the migration plan.
type ReconcilerMetrics struct {
	mu       sync.RWMutex
	modules  map[string]*reconcilerModuleMetrics
}

type reconcilerModuleMetrics struct {
	status       ReconcilerStatus
	durationSum  float64
	durationCount int64
}

// NewReconcilerMetrics creates a new per-module metrics tracker.
func NewReconcilerMetrics() *ReconcilerMetrics {
	return &ReconcilerMetrics{
		modules: make(map[string]*reconcilerModuleMetrics),
	}
}

// Register marks a reconciler as initialized. Called once at startup.
func (rm *ReconcilerMetrics) Register(module string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if _, ok := rm.modules[module]; !ok {
		rm.modules[module] = &reconcilerModuleMetrics{
			status: ReconcilerStatus{
				Module:      module,
				Initialized: true,
			},
		}
	}
}

// SetRunning marks a reconciler as actively looping.
func (rm *ReconcilerMetrics) SetRunning(module string, running bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if m, ok := rm.modules[module]; ok {
		m.status.Running = running
	}
}

// SetShadowMode marks a reconciler as running in shadow mode.
func (rm *ReconcilerMetrics) SetShadowMode(module string, shadow bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if m, ok := rm.modules[module]; ok {
		m.status.ShadowMode = shadow
	}
}

// RecordReconcile records the result of a single reconcile call.
func (rm *ReconcilerMetrics) RecordReconcile(module string, duration time.Duration, success bool, requeued bool, errMsg string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	m, ok := rm.modules[module]
	if !ok {
		m = &reconcilerModuleMetrics{
			status: ReconcilerStatus{
				Module:      module,
				Initialized: true,
			},
		}
		rm.modules[module] = m
	}

	now := time.Now()
	durationMs := float64(duration.Milliseconds())

	m.status.TotalReconciles++
	m.status.LastReconcileTime = now
	m.status.LastDurationMs = durationMs
	m.durationSum += durationMs
	m.durationCount++
	m.status.AvgDurationMs = m.durationSum / float64(m.durationCount)

	if success {
		m.status.TotalSuccesses++
		m.status.LastSuccessTime = now
		m.status.ConsecutiveErrors = 0
		m.status.LastError = ""
	} else {
		m.status.TotalErrors++
		m.status.ConsecutiveErrors++
		m.status.LastError = errMsg
	}

	if requeued {
		m.status.TotalRequeues++
	}

	// Also feed global metrics
	GlobalMetrics.RecordReconciliation(duration, success)
}

// GetStatus returns the status of a single reconciler module.
func (rm *ReconcilerMetrics) GetStatus(module string) (ReconcilerStatus, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	m, ok := rm.modules[module]
	if !ok {
		return ReconcilerStatus{}, false
	}
	return m.status, true
}

// GetAllStatuses returns statuses for all registered reconcilers.
func (rm *ReconcilerMetrics) GetAllStatuses() []ReconcilerStatus {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	statuses := make([]ReconcilerStatus, 0, len(rm.modules))
	for _, m := range rm.modules {
		statuses = append(statuses, m.status)
	}
	return statuses
}

// HealthSummary returns a high-level health summary.
func (rm *ReconcilerMetrics) HealthSummary() ReconcilerHealthSummary {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	summary := ReconcilerHealthSummary{
		Total: len(rm.modules),
	}

	for _, m := range rm.modules {
		if m.status.Running {
			summary.Running++
		}
		if m.status.Initialized && !m.status.Running {
			summary.Initialized++
		}
		if m.status.ConsecutiveErrors > 5 {
			summary.Unhealthy++
		} else if m.status.Running {
			summary.Healthy++
		}
		if m.status.ShadowMode {
			summary.ShadowMode++
		}
	}

	summary.Status = "ok"
	if summary.Unhealthy > 0 {
		summary.Status = "degraded"
	}

	return summary
}

// ReconcilerHealthSummary is the response for /health/reconcilers.
type ReconcilerHealthSummary struct {
	Status      string `json:"status"` // "ok" or "degraded"
	Total       int    `json:"total"`
	Initialized int    `json:"initialized"` // registered but not looping
	Running     int    `json:"running"`
	Healthy     int    `json:"healthy"`
	Unhealthy   int    `json:"unhealthy"` // >5 consecutive errors
	ShadowMode  int    `json:"shadowMode"`
}

// GlobalReconcilerMetrics is the package-level instance.
var GlobalReconcilerMetrics = NewReconcilerMetrics()
