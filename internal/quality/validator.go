package quality

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"
)

// ValidationRule defines a rule for data validation
type ValidationRule struct {
	ID          string
	FieldName   string
	FieldType   string
	MinValue    interface{}
	MaxValue    interface{}
	Required    bool
	Pattern     string
	CustomCheck func(interface{}) (bool, string)
	Metadata    map[string]string
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Rule    string
	Message string
	Value   interface{}
}

// DataQualityAnalyzer analyzes data quality and validates records
type DataQualityAnalyzer struct {
	mu                 sync.RWMutex
	rules              map[string][]*ValidationRule
	violations         []*ValidationViolation
	patterns           map[string]*regexp.Regexp
	anomalies          []*Anomaly
	baseline           map[string]*DataBaseline
	validationMetrics  *ValidationMetrics
	historyWindow      time.Duration
	anomalyThreshold   float64
	maxViolationsStore int
}

// ValidationViolation tracks validation failures
type ValidationViolation struct {
	ID        string
	Timestamp time.Time
	Table     string
	Row       interface{}
	Errors    []*ValidationError
	Severity  string // low, medium, high, critical
}

// DataBaseline stores baseline statistics for anomaly detection
type DataBaseline struct {
	FieldName  string
	Count      int64
	Mean       float64
	StdDev     float64
	Min        interface{}
	Max        interface{}
	LastUpdate time.Time
}

// Anomaly represents detected anomalous data
type Anomaly struct {
	ID          string
	Timestamp   time.Time
	Table       string
	Field       string
	Value       interface{}
	Deviation   float64 // standard deviations from mean
	Description string
	Score       float64 // 0-1 anomaly score
}

// ValidationMetrics tracks validation statistics
type ValidationMetrics struct {
	TotalChecks      int64
	PassedChecks     int64
	FailedChecks     int64
	AnomaliesFound   int64
	ViolationsByType map[string]int64
	LastUpdate       time.Time
}

// NewDataQualityAnalyzer creates a new analyzer
func NewDataQualityAnalyzer() *DataQualityAnalyzer {
	return &DataQualityAnalyzer{
		rules:              make(map[string][]*ValidationRule),
		violations:         make([]*ValidationViolation, 0),
		patterns:           make(map[string]*regexp.Regexp),
		anomalies:          make([]*Anomaly, 0),
		baseline:           make(map[string]*DataBaseline),
		validationMetrics:  &ValidationMetrics{ViolationsByType: make(map[string]int64)},
		historyWindow:      24 * time.Hour,
		anomalyThreshold:   3.0, // 3 standard deviations
		maxViolationsStore: 10000,
	}
}

// AddRule adds a validation rule for a table field
func (dqa *DataQualityAnalyzer) AddRule(tableField string, rule *ValidationRule) error {
	dqa.mu.Lock()
	defer dqa.mu.Unlock()

	if rule.Pattern != "" {
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
		dqa.patterns[rule.Pattern] = regexp.MustCompile(rule.Pattern)
	}

	if _, exists := dqa.rules[tableField]; !exists {
		dqa.rules[tableField] = make([]*ValidationRule, 0)
	}
	dqa.rules[tableField] = append(dqa.rules[tableField], rule)
	return nil
}

// ValidateRecord validates a record against defined rules
func (dqa *DataQualityAnalyzer) ValidateRecord(ctx context.Context, table string, record map[string]interface{}) ([]*ValidationError, error) {
	dqa.mu.Lock()
	errors := make([]*ValidationError, 0)
	dqa.validationMetrics.TotalChecks++
	dqa.mu.Unlock()

	for fieldName, value := range record {
		tableField := fmt.Sprintf("%s.%s", table, fieldName)
		rules := dqa.getRules(tableField)

		for _, rule := range rules {
			if err := dqa.validateField(value, rule); err != nil {
				errors = append(errors, &ValidationError{
					Field:   fieldName,
					Rule:    rule.ID,
					Message: err.Error(),
					Value:   value,
				})
			}
		}
	}

	dqa.mu.Lock()
	if len(errors) > 0 {
		dqa.validationMetrics.FailedChecks++
	} else {
		dqa.validationMetrics.PassedChecks++
	}
	dqa.validationMetrics.LastUpdate = time.Now()
	dqa.mu.Unlock()

	return errors, nil
}

// validateField validates a single field against a rule
func (dqa *DataQualityAnalyzer) validateField(value interface{}, rule *ValidationRule) error {
	// Required check
	if rule.Required && value == nil {
		return fmt.Errorf("field is required")
	}

	if value == nil {
		return nil
	}

	// Range check
	if rule.MinValue != nil {
		if !dqa.isGreaterOrEqual(value, rule.MinValue) {
			return fmt.Errorf("value must be >= %v", rule.MinValue)
		}
	}

	if rule.MaxValue != nil {
		if !dqa.isLessOrEqual(value, rule.MaxValue) {
			return fmt.Errorf("value must be <= %v", rule.MaxValue)
		}
	}

	// Pattern check
	if rule.Pattern != "" {
		strValue := fmt.Sprintf("%v", value)
		pattern := dqa.patterns[rule.Pattern]
		if !pattern.MatchString(strValue) {
			return fmt.Errorf("value does not match pattern: %s", rule.Pattern)
		}
	}

	// Custom check
	if rule.CustomCheck != nil {
		if ok, msg := rule.CustomCheck(value); !ok {
			return fmt.Errorf("%s", msg)
		}
	}

	return nil
}

// DetectAnomalies detects anomalous data points
func (dqa *DataQualityAnalyzer) DetectAnomalies(ctx context.Context, table string, field string, values []interface{}) ([]*Anomaly, error) {
	dqa.mu.Lock()
	defer dqa.mu.Unlock()

	baseline, exists := dqa.baseline[fmt.Sprintf("%s.%s", table, field)]
	if !exists {
		// Build baseline from current data
		baseline = dqa.buildBaseline(field, values)
		dqa.baseline[fmt.Sprintf("%s.%s", table, field)] = baseline
		return make([]*Anomaly, 0), nil
	}

	anomalies := make([]*Anomaly, 0)

	for _, value := range values {
		numValue := dqa.toFloat64(value)
		deviation := (numValue - baseline.Mean) / baseline.StdDev

		if deviation > dqa.anomalyThreshold || deviation < -dqa.anomalyThreshold {
			anomaly := &Anomaly{
				ID:          fmt.Sprintf("%s-%d", field, time.Now().UnixNano()),
				Timestamp:   time.Now(),
				Table:       table,
				Field:       field,
				Value:       value,
				Deviation:   deviation,
				Description: fmt.Sprintf("Value %.2f is %.2f std devs from mean", numValue, deviation),
				Score:       dqa.calculateAnomalyScore(deviation, dqa.anomalyThreshold),
			}
			anomalies = append(anomalies, anomaly)
			dqa.anomalies = append(dqa.anomalies, anomaly)
			dqa.validationMetrics.AnomaliesFound++

			if len(dqa.anomalies) > dqa.maxViolationsStore {
				dqa.anomalies = dqa.anomalies[1:]
			}
		}
	}

	return anomalies, nil
}

// buildBaseline creates baseline statistics from values
func (dqa *DataQualityAnalyzer) buildBaseline(field string, values []interface{}) *DataBaseline {
	if len(values) == 0 {
		return &DataBaseline{FieldName: field, Count: 0, LastUpdate: time.Now()}
	}

	baseline := &DataBaseline{
		FieldName:  field,
		Count:      int64(len(values)),
		LastUpdate: time.Now(),
	}

	floatValues := make([]float64, 0)
	var sum float64

	for _, v := range values {
		if v != nil {
			fv := dqa.toFloat64(v)
			floatValues = append(floatValues, fv)
			sum += fv
			if baseline.Min == nil || dqa.isLess(v, baseline.Min) {
				baseline.Min = v
			}
			if baseline.Max == nil || dqa.isGreater(v, baseline.Max) {
				baseline.Max = v
			}
		}
	}

	if len(floatValues) > 0 {
		baseline.Mean = sum / float64(len(floatValues))

		// Calculate standard deviation
		var variance float64
		for _, v := range floatValues {
			variance += (v - baseline.Mean) * (v - baseline.Mean)
		}
		baseline.StdDev = dqa.sqrt(variance / float64(len(floatValues)))
	}

	return baseline
}

// GetViolations returns recent violations
func (dqa *DataQualityAnalyzer) GetViolations(limit int) []*ValidationViolation {
	dqa.mu.RLock()
	defer dqa.mu.RUnlock()

	if limit > len(dqa.violations) {
		limit = len(dqa.violations)
	}
	if limit == 0 {
		return make([]*ValidationViolation, 0)
	}

	return dqa.violations[len(dqa.violations)-limit:]
}

// GetAnomalies returns detected anomalies
func (dqa *DataQualityAnalyzer) GetAnomalies(limit int) []*Anomaly {
	dqa.mu.RLock()
	defer dqa.mu.RUnlock()

	if limit > len(dqa.anomalies) {
		limit = len(dqa.anomalies)
	}
	if limit == 0 {
		return make([]*Anomaly, 0)
	}

	return dqa.anomalies[len(dqa.anomalies)-limit:]
}

// RecordViolation records a validation violation
func (dqa *DataQualityAnalyzer) RecordViolation(violation *ValidationViolation) {
	dqa.mu.Lock()
	defer dqa.mu.Unlock()

	violation.ID = fmt.Sprintf("vio-%d", time.Now().UnixNano())
	violation.Timestamp = time.Now()

	dqa.violations = append(dqa.violations, violation)
	dqa.validationMetrics.ViolationsByType[violation.Severity]++

	if len(dqa.violations) > dqa.maxViolationsStore {
		dqa.violations = dqa.violations[1:]
	}
}

// GetMetrics returns validation metrics
func (dqa *DataQualityAnalyzer) GetMetrics() *ValidationMetrics {
	dqa.mu.RLock()
	defer dqa.mu.RUnlock()

	return &ValidationMetrics{
		TotalChecks:      dqa.validationMetrics.TotalChecks,
		PassedChecks:     dqa.validationMetrics.PassedChecks,
		FailedChecks:     dqa.validationMetrics.FailedChecks,
		AnomaliesFound:   dqa.validationMetrics.AnomaliesFound,
		ViolationsByType: dqa.validationMetrics.ViolationsByType,
		LastUpdate:       dqa.validationMetrics.LastUpdate,
	}
}

// GetDataQualityScore calculates data quality score 0-100
func (dqa *DataQualityAnalyzer) GetDataQualityScore() float64 {
	dqa.mu.RLock()
	defer dqa.mu.RUnlock()

	if dqa.validationMetrics.TotalChecks == 0 {
		return 100.0
	}

	passRate := float64(dqa.validationMetrics.PassedChecks) / float64(dqa.validationMetrics.TotalChecks) * 100
	anomalyPenalty := (float64(dqa.validationMetrics.AnomaliesFound) / float64(dqa.validationMetrics.TotalChecks)) * 10

	score := passRate - anomalyPenalty
	if score < 0 {
		score = 0
	}
	return score
}

// Helper methods

func (dqa *DataQualityAnalyzer) getRules(tableField string) []*ValidationRule {
	dqa.mu.RLock()
	defer dqa.mu.RUnlock()

	if rules, exists := dqa.rules[tableField]; exists {
		return rules
	}
	return make([]*ValidationRule, 0)
}

func (dqa *DataQualityAnalyzer) toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if len(val) > 0 {
			return float64(len(val))
		}
	}
	return 0
}

func (dqa *DataQualityAnalyzer) sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 100; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func (dqa *DataQualityAnalyzer) isGreater(a, b interface{}) bool {
	return dqa.toFloat64(a) > dqa.toFloat64(b)
}

func (dqa *DataQualityAnalyzer) isLess(a, b interface{}) bool {
	return dqa.toFloat64(a) < dqa.toFloat64(b)
}

func (dqa *DataQualityAnalyzer) isGreaterOrEqual(a, b interface{}) bool {
	return dqa.toFloat64(a) >= dqa.toFloat64(b)
}

func (dqa *DataQualityAnalyzer) isLessOrEqual(a, b interface{}) bool {
	return dqa.toFloat64(a) <= dqa.toFloat64(b)
}

func (dqa *DataQualityAnalyzer) calculateAnomalyScore(deviation, threshold float64) float64 {
	absDeviation := deviation
	if absDeviation < 0 {
		absDeviation = -absDeviation
	}

	if absDeviation <= threshold {
		return 0.0
	}

	score := (absDeviation - threshold) / (threshold * 10)
	if score > 1.0 {
		score = 1.0
	}
	return score
}
