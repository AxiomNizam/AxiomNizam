package utils

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig holds configuration for distributed tracing
type TracingConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
}

// MetricRecorder handles OpenTelemetry metrics
type MetricRecorder struct {
	tracer       trace.Tracer
	meter        metric.Meter
	requestCount metric.Int64Counter
	duration     metric.Float64Histogram
	errors       metric.Int64Counter
}

// NewMetricRecorder creates a new metric recorder
func NewMetricRecorder(serviceName string) (*MetricRecorder, error) {
	tracer := otel.Tracer("axiom-nizam/metrics")
	meter := otel.Meter("axiom-nizam/metrics")

	reqCount, err := meter.Int64Counter("http_requests_total",
		metric.WithDescription("Total HTTP requests"),
		metric.WithUnit("{requests}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request counter: %w", err)
	}

	dur, err := meter.Float64Histogram("http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create duration histogram: %w", err)
	}

	errCount, err := meter.Int64Counter("http_errors_total",
		metric.WithDescription("Total HTTP errors"),
		metric.WithUnit("{errors}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create error counter: %w", err)
	}

	return &MetricRecorder{
		tracer:       tracer,
		meter:        meter,
		requestCount: reqCount,
		duration:     dur,
		errors:       errCount,
	}, nil
}

// RecordRequest records HTTP request metrics
func (mr *MetricRecorder) RecordRequest(ctx context.Context, method, path, statusCode string, duration float64) {
	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.String("status_code", statusCode),
	}

	mr.requestCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	mr.duration.Record(ctx, duration, metric.WithAttributes(attrs...))

	if statusCode[0] >= '4' {
		mr.errors.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// SpanAttribute holds span attribute information
type SpanAttribute struct {
	Key   string
	Value interface{}
}

// StartSpan starts a new tracing span
func (mr *MetricRecorder) StartSpan(ctx context.Context, name string, attrs ...SpanAttribute) (context.Context, trace.Span) {
	ctx, span := mr.tracer.Start(ctx, name)

	for _, attr := range attrs {
		switch v := attr.Value.(type) {
		case string:
			span.SetAttributes(attribute.String(attr.Key, v))
		case int:
			span.SetAttributes(attribute.Int(attr.Key, v))
		case int64:
			span.SetAttributes(attribute.Int64(attr.Key, v))
		case float64:
			span.SetAttributes(attribute.Float64(attr.Key, v))
		case bool:
			span.SetAttributes(attribute.Bool(attr.Key, v))
		}
	}

	return ctx, span
}

// RecordError records an error in the current span
func RecordSpanError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// HealthChecker performs health checks for CNCF compliance
type HealthChecker struct {
	name    string
	timeout time.Duration
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(name string, timeout time.Duration) *HealthChecker {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HealthChecker{
		name:    name,
		timeout: timeout,
	}
}

// CheckHealth performs a health check
func (hc *HealthChecker) CheckHealth(ctx context.Context, checkFunc func(context.Context) error) (bool, string) {
	ctx, cancel := context.WithTimeout(ctx, hc.timeout)
	defer cancel()

	err := checkFunc(ctx)
	if err != nil {
		return false, fmt.Sprintf("%s health check failed: %v", hc.name, err)
	}

	return true, fmt.Sprintf("%s is healthy", hc.name)
}

// LivenessProbe checks if the application is alive
func LivenessProbe(checks map[string]func() bool) bool {
	for name, check := range checks {
		if !check() {
			fmt.Printf("Liveness check failed for %s\n", name)
			return false
		}
	}
	return true
}

// ReadinessProbe checks if the application is ready to serve
func ReadinessProbe(checks map[string]func() bool) bool {
	for name, check := range checks {
		if !check() {
			fmt.Printf("Readiness check failed for %s\n", name)
			return false
		}
	}
	return true
}

// StartupProbe checks if the application started successfully
func StartupProbe(checks map[string]func() bool) bool {
	for name, check := range checks {
		if !check() {
			fmt.Printf("Startup check failed for %s\n", name)
			return false
		}
	}
	return true
}

// ResourceLimit represents resource constraints
type ResourceLimit struct {
	CPUMillis    int64
	MemoryBytes  int64
	MaxOpenFiles int64
}

// ValidateResourceLimits validates if current resource usage is within limits
func ValidateResourceLimits(current, limit ResourceLimit) (bool, []string) {
	var violations []string

	if current.CPUMillis > limit.CPUMillis {
		violations = append(violations, fmt.Sprintf("CPU usage exceeded: %d > %d millicores", current.CPUMillis, limit.CPUMillis))
	}

	if current.MemoryBytes > limit.MemoryBytes {
		violations = append(violations, fmt.Sprintf("Memory usage exceeded: %d > %d bytes", current.MemoryBytes, limit.MemoryBytes))
	}

	if current.MaxOpenFiles > limit.MaxOpenFiles {
		violations = append(violations, fmt.Sprintf("Open files exceeded: %d > %d", current.MaxOpenFiles, limit.MaxOpenFiles))
	}

	return len(violations) == 0, violations
}

// GracefulShutdown handles graceful service shutdown
type GracefulShutdown struct {
	timeout time.Duration
}

// NewGracefulShutdown creates a new graceful shutdown handler
func NewGracefulShutdown(timeout time.Duration) *GracefulShutdown {
	return &GracefulShutdown{timeout: timeout}
}

// Shutdown performs graceful shutdown with timeout
func (gs *GracefulShutdown) Shutdown(ctx context.Context, shutdownFunc func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	return shutdownFunc(ctx)
}
