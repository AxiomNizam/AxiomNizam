package rules

// =====================================================
// WS-2.1 — Quality Rules Evaluation Engine
//
// Executes quality checks against datasources. Supports built-in
// check types: freshness, volume, not_null, unique, custom_sql,
// range, completeness, row_count_change, statistical.
//
// The engine connects to datasources via the DataSourceQuerier
// interface and executes type-specific validation logic.
// =====================================================

import (
	"context"
	"fmt"
	"time"
)

// CheckType enumerates the built-in quality check types.
type CheckType string

const (
	CheckTypeFreshness      CheckType = "freshness"
	CheckTypeVolume         CheckType = "volume"
	CheckTypeSchema         CheckType = "schema"
	CheckTypeNotNull        CheckType = "not_null"
	CheckTypeUnique         CheckType = "unique"
	CheckTypeAcceptedValues CheckType = "accepted_values"
	CheckTypeRange          CheckType = "range"
	CheckTypeRegex          CheckType = "regex"
	CheckTypeReferential    CheckType = "referential"
	CheckTypeCustomSQL      CheckType = "custom_sql"
	CheckTypeStatistical    CheckType = "statistical"
	CheckTypeRowCountChange CheckType = "row_count_change"
	CheckTypeCompleteness   CheckType = "completeness"
	CheckTypeDistribution   CheckType = "distribution"
	CheckTypeTimeliness     CheckType = "timeliness"
)

// DataSourceQuerier abstracts query execution against a datasource.
type DataSourceQuerier interface {
	// QueryScalar executes a query and returns a single scalar value.
	QueryScalar(ctx context.Context, dsRef, query string) (interface{}, error)

	// QueryRows executes a query and returns the row count.
	QueryRows(ctx context.Context, dsRef, query string) (int64, error)

	// QueryFloat executes a query and returns a float64 result.
	QueryFloat(ctx context.Context, dsRef, query string) (float64, error)

	// QueryTimestamp executes a query and returns a timestamp result.
	QueryTimestamp(ctx context.Context, dsRef, query string) (*time.Time, error)
}

// RuleEngine evaluates quality rules against datasources.
type RuleEngine struct {
	querier DataSourceQuerier
}

// NewRuleEngine creates a new quality rule evaluation engine.
func NewRuleEngine(querier DataSourceQuerier) *RuleEngine {
	return &RuleEngine{querier: querier}
}

// Evaluate executes a quality rule and returns the check output.
func (e *RuleEngine) Evaluate(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if e.querier == nil {
		return nil, fmt.Errorf("quality engine: no datasource querier configured")
	}

	// Validate all user-provided identifiers before building SQL.
	if err := ValidateRuleInputs(rule); err != nil {
		return nil, fmt.Errorf("quality engine: input validation failed: %w", err)
	}

	switch CheckType(rule.Spec.RuleType) {
	case CheckTypeFreshness:
		return e.checkFreshness(ctx, rule)
	case CheckTypeVolume:
		return e.checkVolume(ctx, rule)
	case CheckTypeNotNull:
		return e.checkNotNull(ctx, rule)
	case CheckTypeUnique:
		return e.checkUnique(ctx, rule)
	case CheckTypeCustomSQL:
		return e.checkCustomSQL(ctx, rule)
	case CheckTypeRange:
		return e.checkRange(ctx, rule)
	case CheckTypeCompleteness:
		return e.checkCompleteness(ctx, rule)
	case CheckTypeRowCountChange:
		return e.checkRowCountChange(ctx, rule)
	case CheckTypeStatistical:
		return e.checkStatistical(ctx, rule)
	case CheckTypeSchema:
		return e.checkSchema(ctx, rule)
	case CheckTypeRegex:
		return e.checkRegex(ctx, rule)
	case CheckTypeReferential:
		return e.checkReferential(ctx, rule)
	case CheckTypeDistribution:
		return e.checkDistribution(ctx, rule)
	case CheckTypeTimeliness:
		return e.checkTimeliness(ctx, rule)
	case CheckTypeAcceptedValues:
		return e.checkAcceptedValues(ctx, rule)
	default:
		return nil, fmt.Errorf("quality engine: unsupported rule type %q", rule.Spec.RuleType)
	}
}

// checkFreshness validates that data is not older than the configured max age.
func (e *RuleEngine) checkFreshness(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.Freshness == nil {
		return nil, fmt.Errorf("freshness rule requires freshness config")
	}

	maxAge, err := time.ParseDuration(rule.Spec.Freshness.MaxAge)
	if err != nil {
		return nil, fmt.Errorf("invalid maxAge %q: %w", rule.Spec.Freshness.MaxAge, err)
	}

	query := fmt.Sprintf("SELECT MAX(%s) FROM %s",
		rule.Spec.Freshness.TimestampColumn, rule.Spec.AssetRef)

	ts, err := e.querier.QueryTimestamp(ctx, rule.Spec.DataSourceRef, query)
	if err != nil {
		return nil, fmt.Errorf("freshness check query failed: %w", err)
	}

	if ts == nil {
		return &CheckOutput{Passed: false, Message: "no data found (NULL timestamp)"}, nil
	}

	age := time.Since(*ts)
	passed := age <= maxAge

	return &CheckOutput{
		Passed:      passed,
		Message:     fmt.Sprintf("data age: %s (max allowed: %s)", age.Truncate(time.Second), maxAge),
		ActualValue: age.Truncate(time.Second).String(),
	}, nil
}

// checkVolume validates that row count is within expected bounds.
func (e *RuleEngine) checkVolume(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.Volume == nil {
		return nil, fmt.Errorf("volume rule requires volume config")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", rule.Spec.AssetRef)
	count, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, query)
	if err != nil {
		return nil, fmt.Errorf("volume check query failed: %w", err)
	}

	passed := true
	if rule.Spec.Volume.MinRows > 0 && count < rule.Spec.Volume.MinRows {
		passed = false
	}
	if rule.Spec.Volume.MaxRows > 0 && count > rule.Spec.Volume.MaxRows {
		passed = false
	}

	return &CheckOutput{
		Passed:      passed,
		TotalRows:   count,
		Message:     fmt.Sprintf("row count: %d (min: %d, max: %d)", count, rule.Spec.Volume.MinRows, rule.Spec.Volume.MaxRows),
		ActualValue: fmt.Sprintf("%d", count),
	}, nil
}

// checkNotNull validates that a column has no (or few) null values.
func (e *RuleEngine) checkNotNull(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.NotNull == nil {
		return nil, fmt.Errorf("not_null rule requires notNull config")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s IS NULL",
		rule.Spec.AssetRef, rule.Spec.NotNull.Column)
	nullCount, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, query)
	if err != nil {
		return nil, fmt.Errorf("not_null check query failed: %w", err)
	}

	passed := nullCount == 0
	if rule.Spec.NotNull.Threshold > 0 {
		totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", rule.Spec.AssetRef)
		total, tErr := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, totalQuery)
		if tErr == nil && total > 0 {
			nullRatio := float64(nullCount) / float64(total)
			passed = nullRatio <= rule.Spec.NotNull.Threshold
		}
	}

	return &CheckOutput{
		Passed:      passed,
		FailCount:   nullCount,
		Message:     fmt.Sprintf("null count in %s: %d", rule.Spec.NotNull.Column, nullCount),
		ActualValue: fmt.Sprintf("%d", nullCount),
	}, nil
}

// checkUnique validates that a column has unique values.
func (e *RuleEngine) checkUnique(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.Unique == nil {
		return nil, fmt.Errorf("unique rule requires unique config")
	}

	query := fmt.Sprintf("SELECT COUNT(*) - COUNT(DISTINCT %s) FROM %s",
		rule.Spec.Unique.Column, rule.Spec.AssetRef)
	duplicates, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, query)
	if err != nil {
		return nil, fmt.Errorf("unique check query failed: %w", err)
	}

	return &CheckOutput{
		Passed:      duplicates == 0,
		FailCount:   duplicates,
		Message:     fmt.Sprintf("duplicate count in %s: %d", rule.Spec.Unique.Column, duplicates),
		ActualValue: fmt.Sprintf("%d", duplicates),
	}, nil
}

// checkCustomSQL executes a user-defined SQL query. Passes if result <= threshold.
func (e *RuleEngine) checkCustomSQL(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.CustomSQL == nil {
		return nil, fmt.Errorf("custom_sql rule requires customSQL config")
	}

	count, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, rule.Spec.CustomSQL.Query)
	if err != nil {
		return nil, fmt.Errorf("custom_sql check query failed: %w", err)
	}

	return &CheckOutput{
		Passed:      count <= rule.Spec.CustomSQL.Threshold,
		FailCount:   count,
		Message:     fmt.Sprintf("custom SQL returned %d failing rows (threshold: %d)", count, rule.Spec.CustomSQL.Threshold),
		ActualValue: fmt.Sprintf("%d", count),
	}, nil
}

// checkRange validates that numeric values fall within min/max bounds.
func (e *RuleEngine) checkRange(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.Range == nil {
		return nil, fmt.Errorf("range rule requires range config")
	}

	var conditions []string
	if rule.Spec.Range.MinValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s < %f", rule.Spec.Range.Column, *rule.Spec.Range.MinValue))
	}
	if rule.Spec.Range.MaxValue != nil {
		conditions = append(conditions, fmt.Sprintf("%s > %f", rule.Spec.Range.Column, *rule.Spec.Range.MaxValue))
	}
	if len(conditions) == 0 {
		return &CheckOutput{Passed: true, Message: "no range bounds specified"}, nil
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s",
		rule.Spec.AssetRef, joinStrings(conditions, " OR "))
	outOfRange, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, query)
	if err != nil {
		return nil, fmt.Errorf("range check query failed: %w", err)
	}

	return &CheckOutput{
		Passed:      outOfRange == 0,
		FailCount:   outOfRange,
		Message:     fmt.Sprintf("out-of-range values in %s: %d", rule.Spec.Range.Column, outOfRange),
		ActualValue: fmt.Sprintf("%d", outOfRange),
	}, nil
}

// checkCompleteness validates that a column has a minimum percentage of non-null values.
func (e *RuleEngine) checkCompleteness(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.Completeness == nil {
		return nil, fmt.Errorf("completeness rule requires completeness config")
	}

	totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", rule.Spec.AssetRef)
	total, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, totalQuery)
	if err != nil {
		return nil, fmt.Errorf("completeness total query failed: %w", err)
	}
	if total == 0 {
		return &CheckOutput{Passed: false, Message: "table is empty"}, nil
	}

	nonNullQuery := fmt.Sprintf("SELECT COUNT(%s) FROM %s",
		rule.Spec.Completeness.Column, rule.Spec.AssetRef)
	nonNull, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, nonNullQuery)
	if err != nil {
		return nil, fmt.Errorf("completeness non-null query failed: %w", err)
	}

	completeness := float64(nonNull) / float64(total)
	passed := completeness >= rule.Spec.Completeness.Threshold

	return &CheckOutput{
		Passed:      passed,
		TotalRows:   total,
		FailCount:   total - nonNull,
		Message:     fmt.Sprintf("completeness of %s: %.1f%% (minimum: %.1f%%)", rule.Spec.Completeness.Column, completeness*100, rule.Spec.Completeness.Threshold*100),
		ActualValue: fmt.Sprintf("%.2f%%", completeness*100),
	}, nil
}

// checkRowCountChange validates that day-over-day row count change is within threshold.
func (e *RuleEngine) checkRowCountChange(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.RowCountChange == nil {
		return nil, fmt.Errorf("row_count_change rule requires rowCountChange config")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", rule.Spec.AssetRef)
	currentCount, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, query)
	if err != nil {
		return nil, fmt.Errorf("row_count_change query failed: %w", err)
	}

	previousCount := rule.Spec.RowCountChange.PreviousCount
	if previousCount == 0 {
		return &CheckOutput{
			Passed:      true,
			TotalRows:   currentCount,
			Message:     fmt.Sprintf("first run, recorded baseline count: %d", currentCount),
			ActualValue: fmt.Sprintf("%d", currentCount),
		}, nil
	}

	changePct := float64(currentCount-previousCount) / float64(previousCount) * 100.0
	maxChange := rule.Spec.RowCountChange.MaxChangePct
	passed := changePct >= -maxChange && changePct <= maxChange

	return &CheckOutput{
		Passed:      passed,
		TotalRows:   currentCount,
		Message:     fmt.Sprintf("row count change: %.1f%% (threshold: ±%.1f%%)", changePct, maxChange),
		ActualValue: fmt.Sprintf("%.1f%%", changePct),
	}, nil
}

// checkStatistical validates that a metric is within historical bounds.
func (e *RuleEngine) checkStatistical(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.Statistical == nil {
		return nil, fmt.Errorf("statistical rule requires statistical config")
	}

	var aggFunc string
	switch rule.Spec.Statistical.Metric {
	case "mean", "avg":
		aggFunc = "AVG"
	case "min":
		aggFunc = "MIN"
	case "max":
		aggFunc = "MAX"
	case "sum":
		aggFunc = "SUM"
	default:
		aggFunc = "AVG"
	}

	query := fmt.Sprintf("SELECT %s(%s) FROM %s", aggFunc, rule.Spec.Statistical.Column, rule.Spec.AssetRef)
	value, err := e.querier.QueryFloat(ctx, rule.Spec.DataSourceRef, query)
	if err != nil {
		return nil, fmt.Errorf("statistical check query failed: %w", err)
	}

	passed := true
	if rule.Spec.Statistical.MinValue != 0 && value < rule.Spec.Statistical.MinValue {
		passed = false
	}
	if rule.Spec.Statistical.MaxValue != 0 && value > rule.Spec.Statistical.MaxValue {
		passed = false
	}

	return &CheckOutput{
		Passed:      passed,
		Message:     fmt.Sprintf("%s(%s) = %.4f (bounds: [%.4f, %.4f])", aggFunc, rule.Spec.Statistical.Column, value, rule.Spec.Statistical.MinValue, rule.Spec.Statistical.MaxValue),
		ActualValue: fmt.Sprintf("%.4f", value),
	}, nil
}

// --- Helpers ---

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += sep + p
	}
	return result
}
