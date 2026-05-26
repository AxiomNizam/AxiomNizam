package metrics

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Entity CRUD counters
	LayersCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_gis",
		Name:      "layers_created_total",
		Help:      "Total number of GIS layers created",
	})

	RegionsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_gis",
		Name:      "regions_created_total",
		Help:      "Total number of GIS regions created",
	})

	MarkersCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_gis",
		Name:      "markers_created_total",
		Help:      "Total number of GIS markers created",
	})

	DatasetsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_gis",
		Name:      "datasets_created_total",
		Help:      "Total number of GIS datasets created",
	})

	// Dashboard access counters
	DashboardViews = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_gis",
		Name:      "dashboard_views_total",
		Help:      "Total GIS dashboard views by category",
	}, []string{"category"})

	// Conversion counters
	ConversionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_gis",
		Name:      "conversions_total",
		Help:      "Total GIS conversion operations",
	}, []string{"direction", "outcome"})

	// Gauge for total entities
	TotalLayers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_gis",
		Name:      "total_layers",
		Help:      "Current number of GIS layers",
	})

	TotalRegions = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_gis",
		Name:      "total_regions",
		Help:      "Current number of GIS regions",
	})

	TotalMarkers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_gis",
		Name:      "total_markers",
		Help:      "Current number of GIS markers",
	})

	TotalDatasets = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_gis",
		Name:      "total_datasets",
		Help:      "Current number of GIS datasets",
	})

	// Collector holds in-memory metrics for API responses.
	Collector = &MetricsCollector{startTime: time.Now()}
)

// MetricsCollector holds in-memory GIS metrics.
type MetricsCollector struct {
	startTime       time.Time
	totalLayers     int64
	totalRegions    int64
	totalMarkers    int64
	totalDatasets   int64
	totalViews      int64
	totalConversions int64
}

// RecordEntityCreated records an entity creation.
func (m *MetricsCollector) RecordEntityCreated(entityType string) {
	switch entityType {
	case "layer":
		atomic.AddInt64(&m.totalLayers, 1)
		LayersCreated.Inc()
		TotalLayers.Inc()
	case "region":
		atomic.AddInt64(&m.totalRegions, 1)
		RegionsCreated.Inc()
		TotalRegions.Inc()
	case "marker":
		atomic.AddInt64(&m.totalMarkers, 1)
		MarkersCreated.Inc()
		TotalMarkers.Inc()
	case "dataset":
		atomic.AddInt64(&m.totalDatasets, 1)
		DatasetsCreated.Inc()
		TotalDatasets.Inc()
	}
}

// RecordDashboardView records a dashboard view.
func (m *MetricsCollector) RecordDashboardView(category string) {
	atomic.AddInt64(&m.totalViews, 1)
	DashboardViews.WithLabelValues(category).Inc()
}

// RecordConversion records a conversion operation.
func (m *MetricsCollector) RecordConversion(direction, outcome string) {
	atomic.AddInt64(&m.totalConversions, 1)
	ConversionsTotal.WithLabelValues(direction, outcome).Inc()
}

// Snapshot returns a point-in-time snapshot.
func (m *MetricsCollector) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		TotalLayers:      atomic.LoadInt64(&m.totalLayers),
		TotalRegions:     atomic.LoadInt64(&m.totalRegions),
		TotalMarkers:     atomic.LoadInt64(&m.totalMarkers),
		TotalDatasets:    atomic.LoadInt64(&m.totalDatasets),
		TotalViews:       atomic.LoadInt64(&m.totalViews),
		TotalConversions: atomic.LoadInt64(&m.totalConversions),
		UptimeSeconds:    int64(time.Since(m.startTime).Seconds()),
	}
}

// MetricsSnapshot is a point-in-time snapshot.
type MetricsSnapshot struct {
	TotalLayers      int64 `json:"total_layers"`
	TotalRegions     int64 `json:"total_regions"`
	TotalMarkers     int64 `json:"total_markers"`
	TotalDatasets    int64 `json:"total_datasets"`
	TotalViews       int64 `json:"total_views"`
	TotalConversions int64 `json:"total_conversions"`
	UptimeSeconds    int64 `json:"uptime_seconds"`
}
