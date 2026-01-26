package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Counter tracks a monotonically increasing value
type Counter struct {
	value int64
	name  string
}

// Gauge tracks a value that can go up or down
type Gauge struct {
	mu    sync.RWMutex
	value float64
	name  string
}

// Histogram tracks distribution of values
type Histogram struct {
	mu    sync.RWMutex
	name  string
	sum   float64
	count int64
	min   float64
	max   float64
	// buckets for percentile tracking
	buckets []float64
}

// Metrics provides metrics collection
type Metrics struct {
	mu sync.RWMutex

	// Reconciliation metrics
	ReconcileDurations Histogram
	ReconcileCount     Counter
	ReconcileErrors    Counter
	ReconcileFailed    Counter

	// Policy metrics
	PolicyEvaluations Counter
	PolicyDenied      Counter
	PolicyErrors      Counter
	PolicyDuration    Histogram

	// Queue metrics
	QueueDepth Gauge
	QueueAdds  Counter

	// API metrics
	APIRequests Counter
	APIErrors   Counter
	APIDuration Histogram

	// Error tracking
	ErrorCount Counter
	ErrorRate  Gauge

	// Resource metrics
	ResourcesCreated Counter
	ResourcesUpdated Counter
	ResourcesDeleted Counter
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		ReconcileDurations: Histogram{name: "reconcile_duration_seconds"},
		ReconcileCount:     Counter{name: "reconcile_total"},
		ReconcileErrors:    Counter{name: "reconcile_errors"},
		ReconcileFailed:    Counter{name: "reconcile_failed"},
		PolicyEvaluations:  Counter{name: "policy_evaluations_total"},
		PolicyDenied:       Counter{name: "policy_denied_total"},
		PolicyErrors:       Counter{name: "policy_errors"},
		PolicyDuration:     Histogram{name: "policy_duration_seconds"},
		QueueDepth:         Gauge{name: "queue_depth"},
		QueueAdds:          Counter{name: "queue_adds"},
		APIRequests:        Counter{name: "api_requests_total"},
		APIErrors:          Counter{name: "api_errors"},
		APIDuration:        Histogram{name: "api_duration_seconds"},
		ErrorCount:         Counter{name: "error_total"},
		ErrorRate:          Gauge{name: "error_rate"},
		ResourcesCreated:   Counter{name: "resources_created_total"},
		ResourcesUpdated:   Counter{name: "resources_updated_total"},
		ResourcesDeleted:   Counter{name: "resources_deleted_total"},
	}
}

// RecordReconciliation records a reconciliation event
func (m *Metrics) RecordReconciliation(duration time.Duration, success bool) {
	m.ReconcileDurations.Add(duration.Seconds())
	m.ReconcileCount.Inc()

	if !success {
		m.ReconcileFailed.Inc()
	}
}

// RecordPolicyEvaluation records a policy evaluation
func (m *Metrics) RecordPolicyEvaluation(duration time.Duration, denied bool, err bool) {
	m.PolicyEvaluations.Inc()
	m.PolicyDuration.Add(duration.Seconds())

	if denied {
		m.PolicyDenied.Inc()
	}
	if err {
		m.PolicyErrors.Inc()
	}
}

// RecordQueueAdd records an item being added to queue
func (m *Metrics) RecordQueueAdd(depth int) {
	m.QueueAdds.Inc()
	m.QueueDepth.Set(float64(depth))
}

// RecordAPICall records an API call
func (m *Metrics) RecordAPICall(duration time.Duration, err bool) {
	m.APIRequests.Inc()
	m.APIDuration.Add(duration.Seconds())

	if err {
		m.APIErrors.Inc()
		m.ErrorCount.Inc()
	}
}

// RecordResourceChange records resource creation/update/deletion
func (m *Metrics) RecordResourceChange(action string) {
	switch action {
	case "created":
		m.ResourcesCreated.Inc()
	case "updated":
		m.ResourcesUpdated.Inc()
	case "deleted":
		m.ResourcesDeleted.Inc()
	}
}

// Counter methods

// Inc increments the counter
func (c *Counter) Inc() {
	atomic.AddInt64(&c.value, 1)
}

// Add adds to the counter
func (c *Counter) Add(delta int64) {
	atomic.AddInt64(&c.value, delta)
}

// Value returns the counter value
func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

// Gauge methods

// Set sets the gauge value
func (g *Gauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

// Get gets the gauge value
func (g *Gauge) Get() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

// Histogram methods

// Add adds a value to the histogram
func (h *Histogram) Add(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sum += value
	h.count++

	if h.min == 0 || value < h.min {
		h.min = value
	}
	if value > h.max {
		h.max = value
	}

	h.buckets = append(h.buckets, value)
}

// Mean returns the mean value
func (h *Histogram) Mean() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.count == 0 {
		return 0
	}
	return h.sum / float64(h.count)
}

// Min returns the minimum value
func (h *Histogram) Min() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.min
}

// Max returns the maximum value
func (h *Histogram) Max() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.max
}

// Count returns the number of samples
func (h *Histogram) Count() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

// Stats represents metrics statistics
type Stats struct {
	ReconcileDuration struct {
		Mean  float64
		Min   float64
		Max   float64
		Count int64
	}
	ReconcileCount  int64
	ReconcileFailed int64
	PolicyDenied    int64
	QueueDepth      float64
	ErrorRate       float64
	APIErrorRate    float64
}

// GetStats returns current metrics statistics
func (m *Metrics) GetStats() Stats {
	return Stats{
		ReconcileDuration: struct {
			Mean  float64
			Min   float64
			Max   float64
			Count int64
		}{
			Mean:  m.ReconcileDurations.Mean(),
			Min:   m.ReconcileDurations.Min(),
			Max:   m.ReconcileDurations.Max(),
			Count: m.ReconcileDurations.Count(),
		},
		ReconcileCount:  m.ReconcileCount.Value(),
		ReconcileFailed: m.ReconcileFailed.Value(),
		PolicyDenied:    m.PolicyDenied.Value(),
		QueueDepth:      m.QueueDepth.Get(),
		ErrorRate:       m.ErrorRate.Get(),
		APIErrorRate:    float64(m.APIErrors.Value()) / float64(m.APIRequests.Value()),
	}
}

// GlobalMetrics is the package-level metrics instance
var GlobalMetrics = NewMetrics()

// RecordReconciliation records reconciliation via global metrics
func RecordReconciliation(duration time.Duration, success bool) {
	GlobalMetrics.RecordReconciliation(duration, success)
}

// RecordPolicyEvaluation records policy evaluation via global metrics
func RecordPolicyEvaluation(duration time.Duration, denied bool, err bool) {
	GlobalMetrics.RecordPolicyEvaluation(duration, denied, err)
}

// RecordAPICall records API call via global metrics
func RecordAPICall(duration time.Duration, err bool) {
	GlobalMetrics.RecordAPICall(duration, err)
}

// RecordQueueAdd records queue add via global metrics
func RecordQueueAdd(depth int) {
	GlobalMetrics.RecordQueueAdd(depth)
}

// RecordResourceChange records resource change via global metrics
func RecordResourceChange(action string) {
	GlobalMetrics.RecordResourceChange(action)
}

// GetStats returns global metrics stats
func GetStats() Stats {
	return GlobalMetrics.GetStats()
}
