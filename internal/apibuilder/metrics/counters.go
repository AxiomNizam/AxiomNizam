package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// API operation counters
	APIsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_apibuilder",
		Name:      "apis_created_total",
		Help:      "Total number of custom APIs created",
	})

	APIsDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_apibuilder",
		Name:      "apis_deleted_total",
		Help:      "Total number of custom APIs deleted",
	})

	APIHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_apibuilder",
		Name:      "api_hits_total",
		Help:      "Total number of custom API invocations",
	}, []string{"method", "status"})

	// CSV upload counters
	CSVUploads = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_apibuilder",
		Name:      "csv_uploads_total",
		Help:      "Total number of CSV file uploads",
	})

	// File scan counters
	FileScans = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_apibuilder",
		Name:      "file_scans_total",
		Help:      "Total number of file scans",
	}, []string{"verdict"})

	// Dashboard conversion counters
	Conversions = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_apibuilder",
		Name:      "conversions_total",
		Help:      "Total number of dashboard↔GIS conversions",
	}, []string{"direction", "outcome"})

	// Gauge for active APIs
	ActiveAPIs = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_apibuilder",
		Name:      "active_apis",
		Help:      "Number of currently active custom APIs",
	})

	// Collector holds additional counters for API responses.
	Collector = &MetricsCollector{startTime: time.Now()}
)

// MetricsCollector holds in-memory apibuilder metrics.
type MetricsCollector struct {
	mu            sync.RWMutex
	startTime     time.Time
	totalAPIs     int64
	totalHits     int64
	totalUploads  int64
	totalScans    int64
	safeScans     int64
	unsafeScans   int64
	totalConversions int64
}

// RecordAPICreation increments the API creation counter.
func (m *MetricsCollector) RecordAPICreation() {
	atomic.AddInt64(&m.totalAPIs, 1)
	APIsCreated.Inc()
	ActiveAPIs.Inc()
}

// RecordAPIDeletion increments the API deletion counter.
func (m *MetricsCollector) RecordAPIDeletion() {
	APIsDeleted.Inc()
	ActiveAPIs.Dec()
}

// RecordAPIHit records an API invocation.
func (m *MetricsCollector) RecordAPIHit(method, status string) {
	atomic.AddInt64(&m.totalHits, 1)
	APIHits.WithLabelValues(method, status).Inc()
}

// RecordUpload records a CSV upload.
func (m *MetricsCollector) RecordUpload() {
	atomic.AddInt64(&m.totalUploads, 1)
	CSVUploads.Inc()
}

// RecordScan records a file scan.
func (m *MetricsCollector) RecordScan(safe bool) {
	atomic.AddInt64(&m.totalScans, 1)
	if safe {
		atomic.AddInt64(&m.safeScans, 1)
		FileScans.WithLabelValues("safe").Inc()
	} else {
		atomic.AddInt64(&m.unsafeScans, 1)
		FileScans.WithLabelValues("unsafe").Inc()
	}
}

// RecordConversion records a dashboard↔GIS conversion.
func (m *MetricsCollector) RecordConversion(direction, outcome string) {
	atomic.AddInt64(&m.totalConversions, 1)
	Conversions.WithLabelValues(direction, outcome).Inc()
}

// Snapshot returns a point-in-time snapshot.
func (m *MetricsCollector) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		TotalAPIs:        atomic.LoadInt64(&m.totalAPIs),
		TotalHits:        atomic.LoadInt64(&m.totalHits),
		TotalUploads:     atomic.LoadInt64(&m.totalUploads),
		TotalScans:       atomic.LoadInt64(&m.totalScans),
		SafeScans:        atomic.LoadInt64(&m.safeScans),
		UnsafeScans:      atomic.LoadInt64(&m.unsafeScans),
		TotalConversions: atomic.LoadInt64(&m.totalConversions),
		UptimeSeconds:    int64(time.Since(m.startTime).Seconds()),
	}
}

// MetricsSnapshot is a point-in-time snapshot.
type MetricsSnapshot struct {
	TotalAPIs        int64 `json:"total_apis"`
	TotalHits        int64 `json:"total_hits"`
	TotalUploads     int64 `json:"total_uploads"`
	TotalScans       int64 `json:"total_scans"`
	SafeScans        int64 `json:"safe_scans"`
	UnsafeScans      int64 `json:"unsafe_scans"`
	TotalConversions int64 `json:"total_conversions"`
	UptimeSeconds    int64 `json:"uptime_seconds"`
}
