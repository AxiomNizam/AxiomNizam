package streamanalytics

// =====================================================
// WS-7.2 — Streaming Aggregation Engine
//
// Processes events through windows and computes aggregations.
// Supports count, sum, avg, min, max, percentile, and
// distinct_count functions with group-by support.
// =====================================================

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

// AggregationEngine computes aggregations over event batches.
type AggregationEngine struct {
	mu sync.Mutex
}

// NewAggregationEngine creates a new aggregation engine.
func NewAggregationEngine() *AggregationEngine {
	return &AggregationEngine{}
}

// AggregateResult holds the output of an aggregation computation.
type AggregateResult struct {
	Field    string  `json:"field"`
	Function string  `json:"function"`
	Value    float64 `json:"value"`
}

// ComputeAggregations processes a batch of events through the given aggregation specs.
func (e *AggregationEngine) ComputeAggregations(events []Event, specs []AggregationSpec) []AggregateResult {
	e.mu.Lock()
	defer e.mu.Unlock()

	var results []AggregateResult
	for _, spec := range specs {
		value := e.compute(events, spec)
		results = append(results, AggregateResult{
			Field:    spec.OutputField,
			Function: spec.Function,
			Value:    value,
		})
	}
	return results
}

// ComputeGroupedAggregations processes events with group-by support.
func (e *AggregationEngine) ComputeGroupedAggregations(events []Event, specs []AggregationSpec, groupByFields []string) map[string][]AggregateResult {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Group events by the group-by key.
	groups := make(map[string][]Event)
	for _, event := range events {
		key := e.buildGroupKey(event, groupByFields)
		groups[key] = append(groups[key], event)
	}

	results := make(map[string][]AggregateResult)
	for key, groupEvents := range groups {
		var groupResults []AggregateResult
		for _, spec := range specs {
			value := e.compute(groupEvents, spec)
			groupResults = append(groupResults, AggregateResult{
				Field:    spec.OutputField,
				Function: spec.Function,
				Value:    value,
			})
		}
		results[key] = groupResults
	}

	return results
}

// --- Aggregation functions ---

func (e *AggregationEngine) compute(events []Event, spec AggregationSpec) float64 {
	switch spec.Function {
	case "count":
		return float64(len(events))
	case "sum":
		return e.sum(events, spec.InputField)
	case "avg":
		return e.avg(events, spec.InputField)
	case "min":
		return e.min(events, spec.InputField)
	case "max":
		return e.max(events, spec.InputField)
	case "p50":
		return e.percentile(events, spec.InputField, 50)
	case "p95":
		return e.percentile(events, spec.InputField, 95)
	case "p99":
		return e.percentile(events, spec.InputField, 99)
	case "distinct_count":
		return float64(e.distinctCount(events, spec.InputField))
	case "variance":
		return e.variance(events, spec.InputField)
	case "stddev":
		return math.Sqrt(e.variance(events, spec.InputField))
	case "rate":
		return e.rate(events, spec.InputField)
	default:
		return 0
	}
}

func (e *AggregationEngine) sum(events []Event, field string) float64 {
	var s float64
	for _, ev := range events {
		s += toFloat64(ev.Data[field])
	}
	return s
}

func (e *AggregationEngine) avg(events []Event, field string) float64 {
	if len(events) == 0 {
		return 0
	}
	return e.sum(events, field) / float64(len(events))
}

func (e *AggregationEngine) min(events []Event, field string) float64 {
	if len(events) == 0 {
		return 0
	}
	m := toFloat64(events[0].Data[field])
	for _, ev := range events[1:] {
		v := toFloat64(ev.Data[field])
		if v < m {
			m = v
		}
	}
	return m
}

func (e *AggregationEngine) max(events []Event, field string) float64 {
	if len(events) == 0 {
		return 0
	}
	m := toFloat64(events[0].Data[field])
	for _, ev := range events[1:] {
		v := toFloat64(ev.Data[field])
		if v > m {
			m = v
		}
	}
	return m
}

func (e *AggregationEngine) percentile(events []Event, field string, pct float64) float64 {
	if len(events) == 0 {
		return 0
	}

	values := make([]float64, len(events))
	for i, ev := range events {
		values[i] = toFloat64(ev.Data[field])
	}
	sort.Float64s(values)

	idx := int(math.Ceil(pct/100.0*float64(len(values)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(values) {
		idx = len(values) - 1
	}
	return values[idx]
}

func (e *AggregationEngine) distinctCount(events []Event, field string) int {
	seen := make(map[interface{}]bool)
	for _, ev := range events {
		if v, ok := ev.Data[field]; ok {
			seen[v] = true
		}
	}
	return len(seen)
}

func (e *AggregationEngine) variance(events []Event, field string) float64 {
	if len(events) == 0 {
		return 0
	}
	mean := e.avg(events, field)
	var sumSqDiff float64
	for _, ev := range events {
		diff := toFloat64(ev.Data[field]) - mean
		sumSqDiff += diff * diff
	}
	return sumSqDiff / float64(len(events))
}

func (e *AggregationEngine) rate(events []Event, field string) float64 {
	if len(events) < 2 {
		return 0
	}
	first := events[0].Timestamp
	last := events[len(events)-1].Timestamp
	duration := last.Sub(first).Seconds()
	if duration == 0 {
		return 0
	}
	return e.sum(events, field) / duration
}

func (e *AggregationEngine) buildGroupKey(event Event, groupByFields []string) string {
	if len(groupByFields) == 0 {
		return "__all__"
	}
	var parts []string
	for _, field := range groupByFields {
		val := fmt.Sprintf("%v", event.Data[field])
		parts = append(parts, field+"="+val)
	}
	return joinParts(parts, "|")
}

func joinParts(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += sep + p
	}
	return result
}
