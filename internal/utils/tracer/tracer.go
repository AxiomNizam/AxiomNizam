package tracer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Span represents a traced span
type Span struct {
	span       trace.Span
	startTime  time.Time
	attributes map[string]interface{}
	mu         sync.RWMutex
}

// NewSpan creates a new span
func NewSpan(span trace.Span) *Span {
	return &Span{
		span:       span,
		startTime:  time.Now(),
		attributes: make(map[string]interface{}),
	}
}

// SetAttribute sets a span attribute
func (s *Span) SetAttribute(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attributes[key] = value

	// Set in OpenTelemetry span too
	// Would use otel attribute conversion here
}

// AddEvent adds an event to the span
func (s *Span) AddEvent(name string, attributes map[string]interface{}) {
	// Would add event to OpenTelemetry span
}

// SetStatus sets the span status
func (s *Span) SetStatus(code codes.Code, message string) {
	s.span.SetStatus(code, message)
}

// End ends the span
func (s *Span) End() {
	s.span.End()
}

// Duration returns span duration
func (s *Span) Duration() time.Duration {
	return time.Since(s.startTime)
}

// Tracer provides tracing functionality
type Tracer struct {
	tracer trace.Tracer
	spans  map[string]*Span
	mu     sync.RWMutex
}

// NewTracer creates a new tracer
func NewTracer(serviceName string) *Tracer {
	tracer := otel.Tracer(serviceName)
	return &Tracer{
		tracer: tracer,
		spans:  make(map[string]*Span),
	}
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, *Span) {
	ctx, span := t.tracer.Start(ctx, name)
	s := NewSpan(span)

	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans[name] = s

	return ctx, s
}

// StartSpanWithAttributes starts a span with attributes
func (t *Tracer) StartSpanWithAttributes(ctx context.Context, name string, attributes map[string]interface{}) (context.Context, *Span) {
	ctx, span := t.StartSpan(ctx, name)

	for key, value := range attributes {
		span.SetAttribute(key, value)
	}

	return ctx, span
}

// EndSpan ends a span
func (t *Tracer) EndSpan(span *Span) {
	span.End()
}

// GetSpan retrieves a span
func (t *Tracer) GetSpan(name string) *Span {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.spans[name]
}

// Trace executes a function with tracing
func (t *Tracer) Trace(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, span := t.StartSpan(ctx, name)
	defer span.End()

	if err := fn(ctx); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// TraceWithResult executes a function with tracing and result
func (t *Tracer) TraceWithResult(ctx context.Context, name string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	ctx, span := t.StartSpan(ctx, name)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return result, nil
}

// ChildTracer creates a child tracer
type ChildTracer struct {
	parent *Tracer
	span   *Span
}

// Start starts a child span
func (ct *ChildTracer) Start(ctx context.Context, name string) (context.Context, *Span) {
	return ct.parent.StartSpan(ctx, fmt.Sprintf("%s.%s", ct.span.span.SpanContext().SpanID(), name))
}

// Propagator handles trace context propagation
type Propagator struct {
	traceID string
	spanID  string
}

// NewPropagator creates a new propagator
func NewPropagator(traceID, spanID string) *Propagator {
	return &Propagator{
		traceID: traceID,
		spanID:  spanID,
	}
}

// TraceID returns the trace ID
func (p *Propagator) TraceID() string {
	return p.traceID
}

// SpanID returns the span ID
func (p *Propagator) SpanID() string {
	return p.spanID
}

// ToContext injects propagator into context
func (p *Propagator) ToContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "trace_id", p.traceID)
	ctx = context.WithValue(ctx, "span_id", p.spanID)
	return ctx
}

// SpanCollector collects span information
type SpanCollector struct {
	spans []*SpanInfo
	mu    sync.RWMutex
}

// SpanInfo contains span information
type SpanInfo struct {
	Name       string
	TraceID    string
	SpanID     string
	ParentID   string
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Status     string
	Message    string
	Attributes map[string]interface{}
}

// NewSpanCollector creates a new span collector
func NewSpanCollector() *SpanCollector {
	return &SpanCollector{
		spans: make([]*SpanInfo, 0),
	}
}

// Collect collects span information
func (sc *SpanCollector) Collect(info SpanInfo) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.spans = append(sc.spans, &info)
}

// GetSpans returns all collected spans
func (sc *SpanCollector) GetSpans() []*SpanInfo {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	spans := make([]*SpanInfo, len(sc.spans))
	copy(spans, sc.spans)
	return spans
}

// Clear clears collected spans
func (sc *SpanCollector) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.spans = make([]*SpanInfo, 0)
}

// Filter filters spans by criteria
type SpanFilter struct {
	collector *SpanCollector
}

// NewSpanFilter creates a new span filter
func NewSpanFilter(collector *SpanCollector) *SpanFilter {
	return &SpanFilter{
		collector: collector,
	}
}

// ByName filters spans by name
func (sf *SpanFilter) ByName(name string) []*SpanInfo {
	var result []*SpanInfo
	for _, span := range sf.collector.GetSpans() {
		if span.Name == name {
			result = append(result, span)
		}
	}
	return result
}

// ByStatus filters spans by status
func (sf *SpanFilter) ByStatus(status string) []*SpanInfo {
	var result []*SpanInfo
	for _, span := range sf.collector.GetSpans() {
		if span.Status == status {
			result = append(result, span)
		}
	}
	return result
}

// ByTraceID filters spans by trace ID
func (sf *SpanFilter) ByTraceID(traceID string) []*SpanInfo {
	var result []*SpanInfo
	for _, span := range sf.collector.GetSpans() {
		if span.TraceID == traceID {
			result = append(result, span)
		}
	}
	return result
}

// SlowSpans returns spans slower than threshold
func (sf *SpanFilter) SlowSpans(threshold time.Duration) []*SpanInfo {
	var result []*SpanInfo
	for _, span := range sf.collector.GetSpans() {
		if span.Duration > threshold {
			result = append(result, span)
		}
	}
	return result
}

// ErrorSpans returns spans with errors
func (sf *SpanFilter) ErrorSpans() []*SpanInfo {
	return sf.ByStatus("error")
}

// SpanAnalyzer analyzes collected spans
type SpanAnalyzer struct {
	collector *SpanCollector
}

// NewSpanAnalyzer creates a new span analyzer
func NewSpanAnalyzer(collector *SpanCollector) *SpanAnalyzer {
	return &SpanAnalyzer{
		collector: collector,
	}
}

// TotalDuration returns total duration of all spans
func (sa *SpanAnalyzer) TotalDuration() time.Duration {
	var total time.Duration
	for _, span := range sa.collector.GetSpans() {
		total += span.Duration
	}
	return total
}

// AverageDuration returns average duration
func (sa *SpanAnalyzer) AverageDuration() time.Duration {
	spans := sa.collector.GetSpans()
	if len(spans) == 0 {
		return 0
	}

	total := sa.TotalDuration()
	return time.Duration(int64(total) / int64(len(spans)))
}

// SlowestSpan returns the slowest span
func (sa *SpanAnalyzer) SlowestSpan() *SpanInfo {
	spans := sa.collector.GetSpans()
	if len(spans) == 0 {
		return nil
	}

	slowest := spans[0]
	for _, span := range spans[1:] {
		if span.Duration > slowest.Duration {
			slowest = span
		}
	}
	return slowest
}

// FastestSpan returns the fastest span
func (sa *SpanAnalyzer) FastestSpan() *SpanInfo {
	spans := sa.collector.GetSpans()
	if len(spans) == 0 {
		return nil
	}

	fastest := spans[0]
	for _, span := range spans[1:] {
		if span.Duration < fastest.Duration {
			fastest = span
		}
	}
	return fastest
}

// ErrorCount returns count of error spans
func (sa *SpanAnalyzer) ErrorCount() int {
	count := 0
	for _, span := range sa.collector.GetSpans() {
		if span.Status == "error" {
			count++
		}
	}
	return count
}

// SpanCount returns total span count
func (sa *SpanAnalyzer) SpanCount() int {
	return len(sa.collector.GetSpans())
}
