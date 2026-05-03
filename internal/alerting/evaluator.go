package alerting

// =====================================================
// WS-4.1 — Alert Condition Evaluator
//
// Evaluates alert rule conditions against metric data sources.
// Supports threshold, anomaly, rate-of-change, and composite
// condition types with configurable aggregation windows.
// =====================================================

import (
	"context"
	"fmt"
	"math"
	"time"
)

// MetricQuerier abstracts metric data retrieval for alert evaluation.
type MetricQuerier interface {
	// QueryMetric returns the current value of a metric.
	QueryMetric(ctx context.Context, source, metric string) (float64, error)

	// QueryMetricRange returns metric values over a time range.
	QueryMetricRange(ctx context.Context, source, metric string, start, end time.Time) ([]MetricSample, error)
}

// MetricSample represents a single metric data point.
type MetricSample struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// EvaluationResult captures the outcome of evaluating a single alert condition.
type EvaluationResult struct {
	Firing       bool    `json:"firing"`
	CurrentValue float64 `json:"currentValue"`
	Threshold    float64 `json:"threshold"`
	Message      string  `json:"message"`
}

// ConditionType enumerates supported alert condition types.
type ConditionType string

const (
	ConditionThreshold    ConditionType = "threshold"     // value > / < / == threshold
	ConditionAnomaly      ConditionType = "anomaly"       // value deviates from rolling avg by N stddevs
	ConditionRateOfChange ConditionType = "rate_of_change" // value changes by > X% in window
	ConditionAbsent       ConditionType = "absent"        // no data for > duration
)

// AlertCondition defines a single evaluable condition.
type AlertCondition struct {
	Type          ConditionType `json:"type"`
	Metric        string        `json:"metric"`
	Source        string        `json:"source"`
	Operator      string        `json:"operator,omitempty"` // gt, lt, gte, lte, eq, ne
	Value         float64       `json:"value,omitempty"`    // Threshold value
	Window        string        `json:"window,omitempty"`   // Evaluation window ("5m", "1h")
	Sensitivity   float64       `json:"sensitivity,omitempty"` // Stddev multiplier for anomaly
	MaxChangePct  float64       `json:"maxChangePct,omitempty"`
	AbsentTimeout string        `json:"absentTimeout,omitempty"`
}

// Evaluator evaluates alert conditions against metric data.
type Evaluator struct {
	querier MetricQuerier
}

// NewEvaluator creates a new alert condition evaluator.
func NewEvaluator(querier MetricQuerier) *Evaluator {
	return &Evaluator{querier: querier}
}

// Evaluate checks a condition and returns whether it is firing.
func (e *Evaluator) Evaluate(ctx context.Context, cond AlertCondition) (*EvaluationResult, error) {
	if e.querier == nil {
		return &EvaluationResult{Firing: false, Message: "no metric querier configured"}, nil
	}

	switch cond.Type {
	case ConditionThreshold:
		return e.evalThreshold(ctx, cond)
	case ConditionAnomaly:
		return e.evalAnomaly(ctx, cond)
	case ConditionRateOfChange:
		return e.evalRateOfChange(ctx, cond)
	case ConditionAbsent:
		return e.evalAbsent(ctx, cond)
	default:
		return nil, fmt.Errorf("evaluator: unsupported condition type %q", cond.Type)
	}
}

// EvaluateAll checks multiple conditions with AND semantics (all must fire).
func (e *Evaluator) EvaluateAll(ctx context.Context, conditions []AlertCondition) (*EvaluationResult, error) {
	for _, cond := range conditions {
		result, err := e.Evaluate(ctx, cond)
		if err != nil {
			return nil, fmt.Errorf("evaluator: condition %s/%s failed: %w", cond.Source, cond.Metric, err)
		}
		if !result.Firing {
			return result, nil // Short-circuit: first non-firing condition returns
		}
	}
	return &EvaluationResult{Firing: true, Message: "all conditions met"}, nil
}

// --- Condition implementations ---

func (e *Evaluator) evalThreshold(ctx context.Context, cond AlertCondition) (*EvaluationResult, error) {
	value, err := e.querier.QueryMetric(ctx, cond.Source, cond.Metric)
	if err != nil {
		return nil, fmt.Errorf("threshold query failed: %w", err)
	}

	firing := compareValue(value, cond.Operator, cond.Value)

	return &EvaluationResult{
		Firing:       firing,
		CurrentValue: value,
		Threshold:    cond.Value,
		Message:      fmt.Sprintf("%s %s %.4f (threshold: %s %.4f)", cond.Metric, boolStr(firing, "is", "is not"), value, cond.Operator, cond.Value),
	}, nil
}

func (e *Evaluator) evalAnomaly(ctx context.Context, cond AlertCondition) (*EvaluationResult, error) {
	window := parseDurationDefault(cond.Window, 1*time.Hour)
	now := time.Now()

	samples, err := e.querier.QueryMetricRange(ctx, cond.Source, cond.Metric, now.Add(-window), now)
	if err != nil {
		return nil, fmt.Errorf("anomaly query failed: %w", err)
	}
	if len(samples) < 2 {
		return &EvaluationResult{Firing: false, Message: "insufficient data for anomaly detection"}, nil
	}

	// Compute mean and standard deviation.
	mean, stddev := computeStats(samples)

	// Current value is the last sample.
	current := samples[len(samples)-1].Value

	sensitivity := cond.Sensitivity
	if sensitivity == 0 {
		sensitivity = 3.0 // Default: 3-sigma
	}

	deviation := math.Abs(current - mean)
	firing := deviation > sensitivity*stddev

	return &EvaluationResult{
		Firing:       firing,
		CurrentValue: current,
		Threshold:    mean + sensitivity*stddev,
		Message:      fmt.Sprintf("%s: current=%.2f, mean=%.2f, stddev=%.2f, deviation=%.1fσ (threshold: %.1fσ)", cond.Metric, current, mean, stddev, deviation/maxFloat(stddev, 0.001), sensitivity),
	}, nil
}

func (e *Evaluator) evalRateOfChange(ctx context.Context, cond AlertCondition) (*EvaluationResult, error) {
	window := parseDurationDefault(cond.Window, 5*time.Minute)
	now := time.Now()

	samples, err := e.querier.QueryMetricRange(ctx, cond.Source, cond.Metric, now.Add(-window), now)
	if err != nil {
		return nil, fmt.Errorf("rate_of_change query failed: %w", err)
	}
	if len(samples) < 2 {
		return &EvaluationResult{Firing: false, Message: "insufficient data for rate-of-change"}, nil
	}

	first := samples[0].Value
	last := samples[len(samples)-1].Value

	var changePct float64
	if first != 0 {
		changePct = ((last - first) / math.Abs(first)) * 100.0
	}

	maxChange := cond.MaxChangePct
	if maxChange == 0 {
		maxChange = 50.0
	}

	firing := math.Abs(changePct) > maxChange

	return &EvaluationResult{
		Firing:       firing,
		CurrentValue: changePct,
		Threshold:    maxChange,
		Message:      fmt.Sprintf("%s: changed %.1f%% over %s (threshold: ±%.1f%%)", cond.Metric, changePct, window, maxChange),
	}, nil
}

func (e *Evaluator) evalAbsent(ctx context.Context, cond AlertCondition) (*EvaluationResult, error) {
	timeout := parseDurationDefault(cond.AbsentTimeout, 5*time.Minute)
	now := time.Now()

	samples, err := e.querier.QueryMetricRange(ctx, cond.Source, cond.Metric, now.Add(-timeout), now)
	if err != nil {
		return nil, fmt.Errorf("absent query failed: %w", err)
	}

	firing := len(samples) == 0

	return &EvaluationResult{
		Firing:       firing,
		CurrentValue: float64(len(samples)),
		Message:      fmt.Sprintf("%s: %d samples in last %s (absent alert fires on 0)", cond.Metric, len(samples), timeout),
	}, nil
}

// --- Helpers ---

func compareValue(value float64, operator string, threshold float64) bool {
	switch operator {
	case "gt", ">":
		return value > threshold
	case "gte", ">=":
		return value >= threshold
	case "lt", "<":
		return value < threshold
	case "lte", "<=":
		return value <= threshold
	case "eq", "==":
		return value == threshold
	case "ne", "!=":
		return value != threshold
	default:
		return value > threshold
	}
}

func computeStats(samples []MetricSample) (mean, stddev float64) {
	n := float64(len(samples))
	if n == 0 {
		return 0, 0
	}
	var sum float64
	for _, s := range samples {
		sum += s.Value
	}
	mean = sum / n

	var variance float64
	for _, s := range samples {
		diff := s.Value - mean
		variance += diff * diff
	}
	stddev = math.Sqrt(variance / n)
	return
}

func parseDurationDefault(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return d
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func boolStr(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}
