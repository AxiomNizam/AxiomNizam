package metrics

import (
	"context"
	"sync"
	"time"
)

// Metric represents a metric
type Metric interface {
	Name() string
	Type() string
	Value() interface{}
	Labels() map[string]string
}

// Counter represents a counter metric
type Counter struct {
	name   string
	value  int64
	labels map[string]string
	mu     sync.RWMutex
}

// NewCounter creates a new counter
func NewCounter(name string) *Counter {
	return &Counter{
		name:   name,
		value:  0,
		labels: make(map[string]string),
	}
}

// Increment increments the counter
func (c *Counter) Increment() {
	c.Add(1)
}

// Add adds value to counter
func (c *Counter) Add(value int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += value
}

// Value returns the counter value
func (c *Counter) Value() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// Name returns the metric name
func (c *Counter) Name() string {
	return c.name
}

// Type returns the metric type
func (c *Counter) Type() string {
	return "counter"
}

// Gauge represents a gauge metric
type Gauge struct {
	name   string
	value  float64
	labels map[string]string
	mu     sync.RWMutex
}

// NewGauge creates a new gauge
func NewGauge(name string) *Gauge {
	return &Gauge{
		name:   name,
		value:  0,
		labels: make(map[string]string),
	}
}

// Set sets the gauge value
func (g *Gauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

// Increment increments the gauge
func (g *Gauge) Increment(delta float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value += delta
}

// Decrement decrements the gauge
func (g *Gauge) Decrement(delta float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value -= delta
}

// Value returns the gauge value
func (g *Gauge) Value() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

// Name returns the metric name
func (g *Gauge) Name() string {
	return g.name
}

// Type returns the metric type
func (g *Gauge) Type() string {
	return "gauge"
}

// Histogram represents a histogram metric
type Histogram struct {
	name    string
	buckets map[float64]int64
	sum     float64
	count   int64
	labels  map[string]string
	mu      sync.RWMutex
}

// NewHistogram creates a new histogram
func NewHistogram(name string) *Histogram {
	return &Histogram{
		name:    name,
		buckets: make(map[float64]int64),
		labels:  make(map[string]string),
	}
}

// Record records a value in the histogram
func (h *Histogram) Record(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sum += value
	h.count++

	// Find appropriate bucket
	h.buckets[value]++
}

// Count returns the total count
func (h *Histogram) Count() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

// Sum returns the sum of values
func (h *Histogram) Sum() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sum
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

// Name returns the metric name
func (h *Histogram) Name() string {
	return h.name
}

// Type returns the metric type
func (h *Histogram) Type() string {
	return "histogram"
}

// Timer measures duration
type Timer struct {
	name      string
	histogram *Histogram
	start     time.Time
	labels    map[string]string
}

// NewTimer creates a new timer
func NewTimer(name string) *Timer {
	return &Timer{
		name:      name,
		histogram: NewHistogram(name),
		labels:    make(map[string]string),
	}
}

// Start starts the timer
func (t *Timer) Start() {
	t.start = time.Now()
}

// Stop stops the timer and records duration
func (t *Timer) Stop() time.Duration {
	duration := time.Since(t.start)
	t.histogram.Record(duration.Seconds())
	return duration
}

// StopAndRecord stops timer and records in milliseconds
func (t *Timer) StopAndRecord() int64 {
	duration := time.Since(t.start)
	millis := duration.Milliseconds()
	t.histogram.Record(float64(millis))
	return millis
}

// Name returns the metric name
func (t *Timer) Name() string {
	return t.name
}

// Registry holds all metrics
type Registry struct {
	metrics map[string]Metric
	mu      sync.RWMutex
}

// NewRegistry creates a new metric registry
func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]Metric),
	}
}

// Register registers a metric
func (r *Registry) Register(metric Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics[metric.Name()] = metric
}

// Get gets a metric
func (r *Registry) Get(name string) (Metric, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.metrics[name]
	return m, ok
}

// All returns all metrics
func (r *Registry) All() map[string]Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make(map[string]Metric)
	for k, v := range r.metrics {
		all[k] = v
	}
	return all
}

// Reset resets all metrics
func (r *Registry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics = make(map[string]Metric)
}

// Collector collects metrics
type Collector struct {
	registry *Registry
	interval time.Duration
	stopCh   chan struct{}
}

// NewCollector creates a new metric collector
func NewCollector(interval time.Duration) *Collector {
	return &Collector{
		registry: NewRegistry(),
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// GetRegistry returns the metric registry
func (c *Collector) GetRegistry() *Registry {
	return c.registry
}

// Start starts metric collection
func (c *Collector) Start(ctx context.Context, fn func(*Registry)) {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fn(c.registry)
			case <-c.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops metric collection
func (c *Collector) Stop() {
	close(c.stopCh)
}

// Exporter exports metrics
type Exporter interface {
	Export(metrics map[string]Metric) error
}

// PrometheusExporter exports metrics in Prometheus format
type PrometheusExporter struct {
	registry *Registry
}

// NewPrometheusExporter creates a new Prometheus exporter
func NewPrometheusExporter(registry *Registry) *PrometheusExporter {
	return &PrometheusExporter{
		registry: registry,
	}
}

// Export exports metrics
func (pe *PrometheusExporter) Export(metrics map[string]Metric) error {
	// Basic implementation - would format as Prometheus metrics
	return nil
}

// Meter provides metric operations
type Meter struct {
	registry *Registry
}

// NewMeter creates a new meter
func NewMeter() *Meter {
	return &Meter{
		registry: NewRegistry(),
	}
}

// Counter creates a counter
func (m *Meter) Counter(name string) *Counter {
	counter := NewCounter(name)
	m.registry.Register(counter)
	return counter
}

// Gauge creates a gauge
func (m *Meter) Gauge(name string) *Gauge {
	gauge := NewGauge(name)
	m.registry.Register(gauge)
	return gauge
}

// Histogram creates a histogram
func (m *Meter) Histogram(name string) *Histogram {
	histogram := NewHistogram(name)
	m.registry.Register(histogram)
	return histogram
}

// Timer creates a timer
func (m *Meter) Timer(name string) *Timer {
	timer := NewTimer(name)
	m.registry.Register(timer.histogram)
	return timer
}

// Registry returns the registry
func (m *Meter) Registry() *Registry {
	return m.registry
}

// MetricSnapshot represents a snapshot of metrics
type MetricSnapshot struct {
	Timestamp time.Time
	Metrics   map[string]interface{}
}

// Snapshot takes a snapshot of metrics
func (m *Meter) Snapshot() MetricSnapshot {
	snapshot := MetricSnapshot{
		Timestamp: time.Now(),
		Metrics:   make(map[string]interface{}),
	}

	for _, metric := range m.registry.All() {
		snapshot.Metrics[metric.Name()] = metric.Value()
	}

	return snapshot
}
