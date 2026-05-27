package gc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	GCPassesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom",
		Subsystem: "gc",
		Name:      "passes_total",
		Help:      "Total number of GC passes completed",
	})

	ResourcesDeletedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom",
		Subsystem: "gc",
		Name:      "resources_deleted_total",
		Help:      "Total resources hard-deleted by the GC",
	})

	ResourcesCascadedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom",
		Subsystem: "gc",
		Name:      "resources_cascaded_total",
		Help:      "Total child resources soft-deleted via owner-reference cascade",
	})

	ResourcesSkippedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom",
		Subsystem: "gc",
		Name:      "resources_skipped_total",
		Help:      "Total resources skipped because finalizers are pending",
	})

	GCPassDurationSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom",
		Subsystem: "gc",
		Name:      "pass_duration_seconds",
		Help:      "Duration of each GC pass in seconds",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30},
	})
)
