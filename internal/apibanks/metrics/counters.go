package metrics

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Snapshot is a point-in-time metrics snapshot.
type Snapshot struct {
	TotalBanks       int   `json:"total_banks"`
	TotalAPIs        int   `json:"total_apis"`
	TotalCreated     int64 `json:"total_created"`
	TotalUpdated     int64 `json:"total_updated"`
	TotalDeleted     int64 `json:"total_deleted"`
	TotalAPIsAdded   int64 `json:"total_apis_added"`
	TotalAPIsRemoved int64 `json:"total_apis_removed"`
	UptimeSeconds    int64 `json:"uptime_seconds"`
}

var (
	BanksTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom",
		Subsystem: "apibanks",
		Name:      "banks_total",
		Help:      "Current number of API banks",
	})

	APIsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom",
		Subsystem: "apibanks",
		Name:      "apis_total",
		Help:      "Current total number of APIs across all banks",
	})

	OperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom",
		Subsystem: "apibanks",
		Name:      "operations_total",
		Help:      "Total apibank operations by type and outcome",
	}, []string{"operation", "outcome"})

	CatalogSearchesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom",
		Subsystem: "apibanks",
		Name:      "catalog_searches_total",
		Help:      "Total catalog search requests",
	})
)

// MetricsCollector tracks apibanks metrics with in-memory counters.
type MetricsCollector struct {
	startTime      time.Time
	totalCreated   int64
	totalUpdated   int64
	totalDeleted   int64
	totalAPIsAdded int64
	totalAPIsRemoved int64
}

// Collector is the global apibanks metrics collector.
var Collector = &MetricsCollector{startTime: time.Now()}

func (m *MetricsCollector) RecordCreated()  { atomic.AddInt64(&m.totalCreated, 1); OperationsTotal.WithLabelValues("create_bank", "success").Inc() }
func (m *MetricsCollector) RecordUpdated()  { atomic.AddInt64(&m.totalUpdated, 1); OperationsTotal.WithLabelValues("update_bank", "success").Inc() }
func (m *MetricsCollector) RecordDeleted()  { atomic.AddInt64(&m.totalDeleted, 1); OperationsTotal.WithLabelValues("delete_bank", "success").Inc() }
func (m *MetricsCollector) RecordAPIAdded() { atomic.AddInt64(&m.totalAPIsAdded, 1); OperationsTotal.WithLabelValues("add_api", "success").Inc() }
func (m *MetricsCollector) RecordAPIRemoved() { atomic.AddInt64(&m.totalAPIsRemoved, 1); OperationsTotal.WithLabelValues("remove_api", "success").Inc() }
func (m *MetricsCollector) RecordError(op string) { OperationsTotal.WithLabelValues(op, "error").Inc() }
func (m *MetricsCollector) RecordCatalogSearch() { CatalogSearchesTotal.Inc() }

// Snapshot returns a point-in-time metrics snapshot.
func (m *MetricsCollector) Snapshot() Snapshot {
	return Snapshot{
		TotalCreated:     atomic.LoadInt64(&m.totalCreated),
		TotalUpdated:     atomic.LoadInt64(&m.totalUpdated),
		TotalDeleted:     atomic.LoadInt64(&m.totalDeleted),
		TotalAPIsAdded:   atomic.LoadInt64(&m.totalAPIsAdded),
		TotalAPIsRemoved: atomic.LoadInt64(&m.totalAPIsRemoved),
		UptimeSeconds:    int64(time.Since(m.startTime).Seconds()),
	}
}
