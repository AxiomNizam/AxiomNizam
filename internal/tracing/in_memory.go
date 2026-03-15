package tracing

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// InMemoryTracingManager in-memory tracing implementation
type InMemoryTracingManager struct {
	mu       sync.RWMutex
	traces   map[string]*Trace
	spans    map[string]*Span
	services map[string]*TraceMetrics
	audits   []*TraceIngestionAuditLog
}

// NewInMemoryTracingManager creates manager
func NewInMemoryTracingManager() *InMemoryTracingManager {
	return &InMemoryTracingManager{
		traces:   make(map[string]*Trace),
		spans:    make(map[string]*Span),
		services: make(map[string]*TraceMetrics),
		audits:   make([]*TraceIngestionAuditLog, 0, 256),
	}
}

func normalizeTraceForStorage(trace *Trace) {
	if trace == nil {
		return
	}

	now := time.Now().UTC()
	if trace.ID == "" {
		trace.ID = fmt.Sprintf("trace-%d", now.UnixNano())
	}
	if trace.StartTime.IsZero() {
		trace.StartTime = now
	}
	if trace.EndTime.IsZero() {
		trace.EndTime = trace.StartTime
	}
	if trace.Duration <= 0 {
		trace.Duration = trace.EndTime.Sub(trace.StartTime).Milliseconds()
	}

	if trace.TotalSpans == 0 {
		trace.TotalSpans = len(trace.Spans)
	}

	serviceSet := make(map[string]bool)
	errorSpans := 0
	for i := range trace.Spans {
		span := &trace.Spans[i]
		if span.ID == "" {
			span.ID = fmt.Sprintf("span-%d", time.Now().UnixNano())
		}
		if span.TraceID == "" {
			span.TraceID = trace.ID
		}
		if span.TenantID == "" {
			span.TenantID = trace.TenantID
		}
		if span.Service != "" {
			serviceSet[span.Service] = true
		}
		if span.Error || span.Status == SpanStatusError {
			errorSpans++
		}
	}

	if len(trace.Services) == 0 && len(serviceSet) > 0 {
		trace.Services = make([]string, 0, len(serviceSet))
		for svc := range serviceSet {
			trace.Services = append(trace.Services, svc)
		}
	}

	trace.ErrorSpans = errorSpans
	if trace.Status == "" {
		if errorSpans > 0 {
			trace.Status = "error"
		} else {
			trace.Status = "success"
		}
	}
}

func (m *InMemoryTracingManager) recalculateTrace(trace *Trace) {
	if trace == nil {
		return
	}

	serviceSet := make(map[string]bool)
	errorSpans := 0
	for _, span := range trace.Spans {
		if span.Service != "" {
			serviceSet[span.Service] = true
		}
		if span.Error || span.Status == SpanStatusError {
			errorSpans++
		}
	}

	trace.TotalSpans = len(trace.Spans)
	trace.ErrorSpans = errorSpans
	if errorSpans > 0 {
		trace.Status = "error"
	} else if trace.Status == "" || strings.EqualFold(trace.Status, "error") {
		trace.Status = "success"
	}

	trace.Services = trace.Services[:0]
	for svc := range serviceSet {
		trace.Services = append(trace.Services, svc)
	}

	if len(trace.Spans) > 0 {
		start := trace.Spans[0].StartTime
		end := trace.Spans[0].EndTime
		for _, span := range trace.Spans[1:] {
			if !span.StartTime.IsZero() && (start.IsZero() || span.StartTime.Before(start)) {
				start = span.StartTime
			}
			if span.EndTime.After(end) {
				end = span.EndTime
			}
		}
		if !start.IsZero() {
			trace.StartTime = start
		}
		if !end.IsZero() {
			trace.EndTime = end
		}
		if !trace.EndTime.IsZero() && !trace.StartTime.IsZero() {
			trace.Duration = trace.EndTime.Sub(trace.StartTime).Milliseconds()
		}
	}
}

// IngestTrace ingests an entire trace and its spans.
func (m *InMemoryTracingManager) IngestTrace(trace *Trace) (*Trace, error) {
	if trace == nil {
		return nil, fmt.Errorf("trace payload is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	normalizeTraceForStorage(trace)
	m.traces[trace.ID] = trace

	for i := range trace.Spans {
		span := trace.Spans[i]
		if span.ID == "" {
			span.ID = fmt.Sprintf("span-%d", time.Now().UnixNano())
		}
		if span.TraceID == "" {
			span.TraceID = trace.ID
		}
		if span.TenantID == "" {
			span.TenantID = trace.TenantID
		}
		trace.Spans[i] = span
		spanCopy := span
		m.spans[span.ID] = &spanCopy
	}

	// Update service metrics for the primary service
	svc := traceService(trace)
	if svc != "" {
		if m.services[svc] == nil {
			m.services[svc] = &TraceMetrics{Service: svc}
		}
		m.services[svc].TraceCount++
		if trace.ErrorSpans > 0 || strings.EqualFold(trace.Status, "error") {
			m.services[svc].ErrorTraceCount++
		}
	}

	return trace, nil
}

// IngestSpan ingests a span into an existing trace.
func (m *InMemoryTracingManager) IngestSpan(span *Span) (*Span, error) {
	if span == nil {
		return nil, fmt.Errorf("span payload is required")
	}
	if strings.TrimSpace(span.TraceID) == "" {
		return nil, fmt.Errorf("traceId is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	trace, exists := m.traces[span.TraceID]
	if !exists {
		return nil, fmt.Errorf("trace not found")
	}

	now := time.Now().UTC()
	if span.ID == "" {
		span.ID = fmt.Sprintf("span-%d", now.UnixNano())
	}
	if span.TenantID == "" {
		span.TenantID = trace.TenantID
	}
	if span.StartTime.IsZero() {
		span.StartTime = now
	}
	if span.EndTime.IsZero() {
		span.EndTime = span.StartTime
	}
	if span.Duration <= 0 {
		span.Duration = span.EndTime.Sub(span.StartTime).Microseconds()
	}

	spanCopy := *span
	m.spans[spanCopy.ID] = &spanCopy

	updated := false
	for i := range trace.Spans {
		if trace.Spans[i].ID == spanCopy.ID {
			trace.Spans[i] = spanCopy
			updated = true
			break
		}
	}
	if !updated {
		trace.Spans = append(trace.Spans, spanCopy)
	}

	m.recalculateTrace(trace)
	return &spanCopy, nil
}

// RecordIngestionAudit stores a tracing ingestion audit event.
func (m *InMemoryTracingManager) RecordIngestionAudit(entry *TraceIngestionAuditLog) error {
	if entry == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("trace-audit-%d", time.Now().UnixNano())
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	m.audits = append(m.audits, entry)
	if len(m.audits) > 5000 {
		m.audits = append([]*TraceIngestionAuditLog(nil), m.audits[len(m.audits)-5000:]...)
	}

	return nil
}

// ListIngestionAudits lists tracing ingestion audit logs.
func (m *InMemoryTracingManager) ListIngestionAudits(filter *TraceIngestionAuditFilter) ([]*TraceIngestionAuditLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if filter == nil {
		filter = &TraceIngestionAuditFilter{}
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	results := make([]*TraceIngestionAuditLog, 0, limit)
	for i := len(m.audits) - 1; i >= 0; i-- {
		entry := m.audits[i]
		if filter.TenantID != "" && !strings.EqualFold(strings.TrimSpace(entry.TenantID), strings.TrimSpace(filter.TenantID)) {
			continue
		}
		if filter.Username != "" && !strings.EqualFold(strings.TrimSpace(entry.Username), strings.TrimSpace(filter.Username)) {
			continue
		}
		if filter.ResourceType != "" && !strings.EqualFold(strings.TrimSpace(entry.ResourceType), strings.TrimSpace(filter.ResourceType)) {
			continue
		}
		if filter.Result != "" && !strings.EqualFold(strings.TrimSpace(entry.Result), strings.TrimSpace(filter.Result)) {
			continue
		}

		results = append(results, entry)
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// GetTrace retrieves trace
func (m *InMemoryTracingManager) GetTrace(id string) (*Trace, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	trace, exists := m.traces[id]
	if !exists {
		return nil, fmt.Errorf("trace not found")
	}
	return trace, nil
}

// traceService returns the primary service name from a trace.
func traceService(t *Trace) string {
	if len(t.Services) > 0 {
		return t.Services[0]
	}
	return ""
}

// SearchTraces searches traces by criteria
func (m *InMemoryTracingManager) SearchTraces(req *TraceSearchRequest) ([]*Trace, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Trace
	for _, t := range m.traces {
		if req.Service != "" && traceService(t) != req.Service {
			continue
		}
		if req.MinDuration > 0 && t.Duration < req.MinDuration {
			continue
		}
		result = append(result, t)
		if req.Limit > 0 && len(result) >= req.Limit {
			break
		}
	}
	return result, nil
}

// GetSpan retrieves span
func (m *InMemoryTracingManager) GetSpan(id string) (*Span, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	span, exists := m.spans[id]
	if !exists {
		return nil, fmt.Errorf("span not found")
	}
	return span, nil
}

// ListSpans lists spans for trace
func (m *InMemoryTracingManager) ListSpans(traceID string) ([]*Span, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Span
	for _, s := range m.spans {
		if s.TraceID == traceID {
			result = append(result, s)
		}
	}
	return result, nil
}

// RecordTrace records trace
func (m *InMemoryTracingManager) RecordTrace(trace *Trace) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if trace.ID == "" {
		trace.ID = fmt.Sprintf("trace-%d", time.Now().UnixNano())
	}
	if trace.StartTime.IsZero() {
		trace.StartTime = time.Now()
	}

	m.traces[trace.ID] = trace

	// Update service metrics for the primary service
	svc := traceService(trace)
	if svc != "" {
		if m.services[svc] == nil {
			m.services[svc] = &TraceMetrics{
				Service: svc,
			}
		}
		m.services[svc].TraceCount++
	}

	return nil
}

// RecordSpan records span
func (m *InMemoryTracingManager) RecordSpan(span *Span) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if span.ID == "" {
		span.ID = fmt.Sprintf("span-%d", time.Now().UnixNano())
	}

	m.spans[span.ID] = span
	return nil
}

// GetServiceMap builds service dependency map
func (m *InMemoryTracingManager) GetServiceMap() (map[string][]*DependencyMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dependencies := make(map[string][]*DependencyMetrics)
	for _, span := range m.spans {
		if span.ParentSpanID != "" {
			parent := m.spans[span.ParentSpanID]
			if parent != nil {
				dep := &DependencyMetrics{
					Source:      parent.Service,
					Destination: span.Service,
				}
				dependencies[parent.Service] = append(dependencies[parent.Service], dep)
			}
		}
	}
	return dependencies, nil
}

// GetServiceMetrics retrieves service metrics
func (m *InMemoryTracingManager) GetServiceMetrics(service string) (*TraceMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics, exists := m.services[service]
	if !exists {
		return nil, fmt.Errorf("service metrics not found")
	}
	return metrics, nil
}

// GetOperationMetrics retrieves operation metrics
func (m *InMemoryTracingManager) GetOperationMetrics(service, operation string) (*SpanMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int64
	var totalDuration int64
	var errorCount int64

	for _, span := range m.spans {
		if span.Service == service && span.OperationName == operation {
			count++
			totalDuration += span.Duration
			if span.Status == SpanStatusError {
				errorCount++
			}
		}
	}

	if count == 0 {
		return nil, fmt.Errorf("no metrics found")
	}

	return &SpanMetrics{
		Service:         service,
		Operation:       operation,
		SpanCount:       count,
		ErrorSpanCount:  errorCount,
		AverageDuration: totalDuration / count,
	}, nil
}

// AnalyzeErrors analyzes errors
func (m *InMemoryTracingManager) AnalyzeErrors(service string) ([]*ErrorAnalysis, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	errorMap := make(map[string]int64)
	for _, span := range m.spans {
		if span.Service == service && span.Error {
			errorMap[span.ErrorMessage]++
		}
	}

	var result []*ErrorAnalysis
	for errMsg, count := range errorMap {
		result = append(result, &ErrorAnalysis{
			ErrorType: errMsg,
			Count:     count,
		})
	}
	return result, nil
}
