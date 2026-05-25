package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ─────────────────────────────────────────────────────────────────────────────
// Counter metrics
// ─────────────────────────────────────────────────────────────────────────────

var (
	MessagesSent = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "messages_sent_total",
		Help:      "Total number of messages sent to backends",
	})

	MessagesReceived = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "messages_received_total",
		Help:      "Total number of messages received from backends",
	})

	MessagesAcked = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "messages_acked_total",
		Help:      "Total number of messages acknowledged",
	})

	MessagesFailed = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "messages_failed_total",
		Help:      "Total number of messages that failed processing",
	})

	MessagesDLQ = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "messages_dlq_total",
		Help:      "Total number of messages sent to dead-letter queue",
	})

	BackendConnections = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "backend_connections_total",
		Help:      "Total number of backend connection attempts",
	})

	BackendErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "backend_errors_total",
		Help:      "Total errors per backend type",
	}, []string{"backend"})

	WorkflowsStarted = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "workflows_started_total",
		Help:      "Total number of workflows started",
	})

	WorkflowsCompleted = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "workflows_completed_total",
		Help:      "Total number of workflows completed successfully",
	})

	WorkflowsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "workflows_failed_total",
		Help:      "Total number of workflows that failed",
	})

	StepsExecuted = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "steps_executed_total",
		Help:      "Total number of workflow steps executed",
	})

	StepsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "conductor",
		Name:      "steps_failed_total",
		Help:      "Total number of workflow steps that failed",
	})
)

// ─────────────────────────────────────────────────────────────────────────────
// Gauge metrics
// ─────────────────────────────────────────────────────────────────────────────

var (
	ActiveProducers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "conductor",
		Name:      "active_producers",
		Help:      "Current number of active producers",
	})

	ActiveConsumers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "conductor",
		Name:      "active_consumers",
		Help:      "Current number of active consumers",
	})

	DLQSize = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "conductor",
		Name:      "dlq_size",
		Help:      "Current number of messages in the dead-letter queue",
	})

	ActiveMessages = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "conductor",
		Name:      "active_messages",
		Help:      "Current number of in-flight messages",
	})

	ActiveWorkflows = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "conductor",
		Name:      "active_workflows",
		Help:      "Current number of running workflows",
	})
)

// ─────────────────────────────────────────────────────────────────────────────
// Histogram metrics
// ─────────────────────────────────────────────────────────────────────────────

var (
	MessageLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "conductor",
		Name:      "message_latency_seconds",
		Help:      "End-to-end message processing latency",
		Buckets:   prometheus.DefBuckets,
	})

	WorkflowDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "conductor",
		Name:      "workflow_duration_seconds",
		Help:      "Total workflow execution time",
		Buckets:   prometheus.DefBuckets,
	})

	StepDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "conductor",
		Name:      "step_duration_seconds",
		Help:      "Duration per workflow step type",
		Buckets:   prometheus.DefBuckets,
	}, []string{"step_type"})
)
