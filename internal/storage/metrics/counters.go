package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// BucketsTotal tracks total buckets by operation.
	BucketsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom",
		Subsystem: "storage",
		Name:      "bucket_operations_total",
		Help:      "Total bucket operations by type",
	}, []string{"operation"})

	// ObjectsTotal tracks total object operations.
	ObjectsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom",
		Subsystem: "storage",
		Name:      "object_operations_total",
		Help:      "Total object operations by type",
	}, []string{"operation"})

	// OperationsTotal tracks total storage operations with outcome.
	OperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom",
		Subsystem: "storage",
		Name:      "operations_total",
		Help:      "Total storage operations by type and outcome",
	}, []string{"operation", "outcome"})

	// AuditEventsTotal tracks total audit events recorded.
	AuditEventsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom",
		Subsystem: "storage",
		Name:      "audit_events_total",
		Help:      "Total audit events recorded",
	})
)
