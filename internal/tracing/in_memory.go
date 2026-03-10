package tracing

import (
	"fmt"
	"sync"
	"time"
)

// InMemoryTracingManager in-memory tracing implementation
type InMemoryTracingManager struct {
	mu       sync.RWMutex
	traces   map[string]*Trace
	spans    map[string]*Span
	services map[string]*TraceMetrics
}

// NewInMemoryTracingManager creates manager
func NewInMemoryTracingManager() *InMemoryTracingManager {
	return &InMemoryTracingManager{
		traces:   make(map[string]*Trace),
		spans:    make(map[string]*Span),
		services: make(map[string]*TraceMetrics),
	}
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
