package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_waitx",
		Name:      "checks_total",
		Help:      "Total number of waitx checks executed",
	}, []string{"check_type", "outcome"})

	ChecksRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_waitx",
		Name:      "checks_running",
		Help:      "Number of checks currently running",
	})

	CheckDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "axiom_waitx",
		Name:      "check_duration_seconds",
		Help:      "Duration of waitx checks in seconds",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
	}, []string{"check_type"})

	RetryAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_waitx",
		Name:      "retry_attempts_total",
		Help:      "Total retry attempts by check type",
	}, []string{"check_type"})

	GroupChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_waitx",
		Name:      "group_checks_total",
		Help:      "Total group check executions",
	}, []string{"outcome"})

	// Collector holds additional counters tracked in-memory for the API.
	Collector = &MetricsCollector{startTime: time.Now()}
)

// MetricsCollector holds in-memory waitx metrics for API responses.
type MetricsCollector struct {
	mu             sync.RWMutex
	startTime      time.Time
	totalChecks    int64
	totalSuccesses int64
	totalFailures  int64
	totalTimeouts  int64
	byCheckType    map[string]*checkTypeStats
}

type checkTypeStats struct {
	runs     int64
	successes int64
	failures  int64
	timeouts  int64
	totalMs   int64
}

// RecordCheck records a completed check.
func (m *MetricsCollector) RecordCheck(checkType string, success bool, timedOut bool, durationMs int64) {
	atomic.AddInt64(&m.totalChecks, 1)

	outcome := "success"
	if !success {
		if timedOut {
			outcome = "timeout"
			atomic.AddInt64(&m.totalTimeouts, 1)
		} else {
			outcome = "failure"
		}
		atomic.AddInt64(&m.totalFailures, 1)
	} else {
		atomic.AddInt64(&m.totalSuccesses, 1)
	}

	ChecksTotal.WithLabelValues(checkType, outcome).Inc()
	CheckDuration.WithLabelValues(checkType).Observe(float64(durationMs) / 1000.0)

	m.mu.Lock()
	if m.byCheckType == nil {
		m.byCheckType = make(map[string]*checkTypeStats)
	}
	stats := m.byCheckType[checkType]
	if stats == nil {
		stats = &checkTypeStats{}
		m.byCheckType[checkType] = stats
	}
	stats.runs++
	if success {
		stats.successes++
	} else if timedOut {
		stats.timeouts++
	} else {
		stats.failures++
	}
	stats.totalMs += durationMs
	m.mu.Unlock()
}

// RecordRetry increments retry counter.
func (m *MetricsCollector) RecordRetry(checkType string) {
	RetryAttemptsTotal.WithLabelValues(checkType).Inc()
}

// Snapshot returns a point-in-time snapshot for API responses.
func (m *MetricsCollector) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snap := MetricsSnapshot{
		TotalChecks:    atomic.LoadInt64(&m.totalChecks),
		TotalSuccesses: atomic.LoadInt64(&m.totalSuccesses),
		TotalFailures:  atomic.LoadInt64(&m.totalFailures),
		TotalTimeouts:  atomic.LoadInt64(&m.totalTimeouts),
		UptimeSeconds:  int64(time.Since(m.startTime).Seconds()),
		ByCheckType:    make([]CheckTypeSnapshot, 0, len(m.byCheckType)),
	}

	for name, stats := range m.byCheckType {
		avgMs := float64(0)
		if stats.runs > 0 {
			avgMs = float64(stats.totalMs) / float64(stats.runs)
		}
		snap.ByCheckType = append(snap.ByCheckType, CheckTypeSnapshot{
			CheckType:  name,
			Runs:       stats.runs,
			Successes:  stats.successes,
			Failures:   stats.failures,
			Timeouts:   stats.timeouts,
			TotalMs:    stats.totalMs,
			AvgMs:      avgMs,
		})
	}

	if snap.TotalChecks > 0 {
		snap.SuccessRate = float64(snap.TotalSuccesses) / float64(snap.TotalChecks) * 100
	}

	return snap
}

// MetricsSnapshot is a point-in-time snapshot of all waitx metrics.
type MetricsSnapshot struct {
	TotalChecks    int64               `json:"total_checks"`
	TotalSuccesses int64               `json:"total_successes"`
	TotalFailures  int64               `json:"total_failures"`
	TotalTimeouts  int64               `json:"total_timeouts"`
	SuccessRate    float64             `json:"success_rate"`
	UptimeSeconds  int64               `json:"uptime_seconds"`
	ByCheckType    []CheckTypeSnapshot `json:"by_check_type"`
}

// CheckTypeSnapshot holds metrics for a single check type.
type CheckTypeSnapshot struct {
	CheckType string  `json:"check_type"`
	Runs      int64   `json:"runs"`
	Successes int64   `json:"successes"`
	Failures  int64   `json:"failures"`
	Timeouts  int64   `json:"timeouts"`
	TotalMs   int64   `json:"total_ms"`
	AvgMs     float64 `json:"avg_ms"`
}
