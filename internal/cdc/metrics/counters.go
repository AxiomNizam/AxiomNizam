package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Pipeline lifecycle counters
	PipelinesCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_cdc",
		Name:      "pipelines_created_total",
		Help:      "Total number of CDC pipelines created",
	})

	PipelinesDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_cdc",
		Name:      "pipelines_deleted_total",
		Help:      "Total number of CDC pipelines deleted",
	})

	// Event counters
	EventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_cdc",
		Name:      "events_total",
		Help:      "Total number of CDC change events processed",
	}, []string{"operation", "table"})

	EventsPerSecond = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_cdc",
		Name:      "events_per_second",
		Help:      "Current CDC events per second",
	})

	// Error counters
	ErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_cdc",
		Name:      "errors_total",
		Help:      "Total number of CDC pipeline errors",
	})

	// Pipeline gauges
	ActivePipelines = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_cdc",
		Name:      "active_pipelines",
		Help:      "Number of currently active CDC pipelines",
	})

	PausedPipelines = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_cdc",
		Name:      "paused_pipelines",
		Help:      "Number of currently paused CDC pipelines",
	})

	// Lag histogram
	PipelineLag = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_cdc",
		Name:      "pipeline_lag_seconds",
		Help:      "CDC pipeline lag in seconds",
		Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 30, 60, 300},
	})

	// ETL counters
	ETLPipelinesCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_cdc",
		Name:      "etl_pipelines_created_total",
		Help:      "Total number of ETL pipelines created",
	})

	ETLRunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_cdc",
		Name:      "etl_runs_total",
		Help:      "Total number of ETL pipeline runs",
	}, []string{"outcome"})

	ETLRowsRead = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_cdc",
		Name:      "etl_rows_read_total",
		Help:      "Total rows read by ETL pipelines",
	})

	ETLRowsWritten = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_cdc",
		Name:      "etl_rows_written_total",
		Help:      "Total rows written by ETL pipelines",
	})

	// Collector holds in-memory metrics for API responses.
	Collector = &MetricsCollector{startTime: time.Now()}
)

// MetricsCollector holds in-memory CDC/ETL metrics.
type MetricsCollector struct {
	mu              sync.RWMutex
	startTime       time.Time
	totalCDCEvents  int64
	totalErrors     int64
	totalETLRuns    int64
	successETLRuns  int64
	failedETLRuns   int64
	totalRowsRead   int64
	totalRowsWritten int64
}

// RecordCDCEvent records a CDC change event.
func (m *MetricsCollector) RecordCDCEvent(operation, table string) {
	atomic.AddInt64(&m.totalCDCEvents, 1)
	EventsTotal.WithLabelValues(operation, table).Inc()
}

// RecordCDCError records a CDC pipeline error.
func (m *MetricsCollector) RecordCDCError() {
	atomic.AddInt64(&m.totalErrors, 1)
	ErrorsTotal.Inc()
}

// RecordETLRun records an ETL pipeline run.
func (m *MetricsCollector) RecordETLRun(success bool, rowsRead, rowsWritten int64) {
	atomic.AddInt64(&m.totalETLRuns, 1)
	atomic.AddInt64(&m.totalRowsRead, rowsRead)
	atomic.AddInt64(&m.totalRowsWritten, rowsWritten)
	ETLRowsRead.Add(float64(rowsRead))
	ETLRowsWritten.Add(float64(rowsWritten))
	if success {
		atomic.AddInt64(&m.successETLRuns, 1)
		ETLRunsTotal.WithLabelValues("success").Inc()
	} else {
		atomic.AddInt64(&m.failedETLRuns, 1)
		ETLRunsTotal.WithLabelValues("failure").Inc()
	}
}

// Snapshot returns a point-in-time snapshot.
func (m *MetricsCollector) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		TotalCDCEvents:   atomic.LoadInt64(&m.totalCDCEvents),
		TotalErrors:      atomic.LoadInt64(&m.totalErrors),
		TotalETLRuns:     atomic.LoadInt64(&m.totalETLRuns),
		SuccessETLRuns:   atomic.LoadInt64(&m.successETLRuns),
		FailedETLRuns:    atomic.LoadInt64(&m.failedETLRuns),
		TotalRowsRead:    atomic.LoadInt64(&m.totalRowsRead),
		TotalRowsWritten: atomic.LoadInt64(&m.totalRowsWritten),
		UptimeSeconds:    int64(time.Since(m.startTime).Seconds()),
	}
}

// MetricsSnapshot is a point-in-time snapshot.
type MetricsSnapshot struct {
	TotalCDCEvents   int64 `json:"total_cdc_events"`
	TotalErrors      int64 `json:"total_errors"`
	TotalETLRuns     int64 `json:"total_etl_runs"`
	SuccessETLRuns   int64 `json:"success_etl_runs"`
	FailedETLRuns    int64 `json:"failed_etl_runs"`
	TotalRowsRead    int64 `json:"total_rows_read"`
	TotalRowsWritten int64 `json:"total_rows_written"`
	UptimeSeconds    int64 `json:"uptime_seconds"`
}
