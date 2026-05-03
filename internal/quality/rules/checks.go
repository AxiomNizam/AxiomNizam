package rules

// =====================================================
// WS-2.1 — Extended Quality Check Types
//
// Additional check types beyond the core 9 in engine.go:
// schema, regex, referential, distribution, timeliness,
// accepted_values.
// =====================================================

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// checkSchema validates that a table's column structure matches expectations.
func (e *RuleEngine) checkSchema(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.Schema == nil {
		return nil, fmt.Errorf("schema rule requires schema config")
	}

	// Query actual column list from information_schema.
	query := fmt.Sprintf(
		"SELECT column_name FROM information_schema.columns WHERE table_name = '%s' ORDER BY ordinal_position",
		rule.Spec.AssetRef,
	)
	_ = query // Schema validation uses catalog metadata when available.

	// For now, validate expected columns exist via scalar count queries.
	var missing []string
	for _, expected := range rule.Spec.Schema.ExpectedColumns {
		colQuery := fmt.Sprintf(
			"SELECT COUNT(*) FROM information_schema.columns WHERE table_name = '%s' AND column_name = '%s'",
			rule.Spec.AssetRef, expected.Name,
		)
		count, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, colQuery)
		if err != nil {
			return nil, fmt.Errorf("schema check query failed for column %s: %w", expected.Name, err)
		}
		if count == 0 {
			missing = append(missing, expected.Name)
		}
	}

	passed := len(missing) == 0
	msg := fmt.Sprintf("schema check: %d/%d expected columns present", len(rule.Spec.Schema.ExpectedColumns)-len(missing), len(rule.Spec.Schema.ExpectedColumns))
	if !passed {
		msg += fmt.Sprintf(" (missing: %s)", joinStrings(missing, ", "))
	}

	return &CheckOutput{
		Passed:      passed,
		FailCount:   int64(len(missing)),
		Message:     msg,
		ActualValue: fmt.Sprintf("%d missing", len(missing)),
	}, nil
}

// checkRegex validates that column values match a regular expression pattern.
func (e *RuleEngine) checkRegex(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.Regex == nil {
		return nil, fmt.Errorf("regex rule requires regex config")
	}

	// Validate the regex compiles.
	_, err := regexp.Compile(rule.Spec.Regex.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern %q: %w", rule.Spec.Regex.Pattern, err)
	}

	// Use SQL REGEXP or NOT REGEXP depending on database dialect.
	// This uses a portable approach: count rows that do NOT match.
	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM %s WHERE %s NOT REGEXP '%s'",
		rule.Spec.AssetRef, rule.Spec.Regex.Column,
		strings.ReplaceAll(rule.Spec.Regex.Pattern, "'", "''"),
	)
	failCount, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, query)
	if err != nil {
		return nil, fmt.Errorf("regex check query failed: %w", err)
	}

	return &CheckOutput{
		Passed:      failCount == 0,
		FailCount:   failCount,
		Message:     fmt.Sprintf("regex violations in %s: %d rows do not match /%s/", rule.Spec.Regex.Column, failCount, rule.Spec.Regex.Pattern),
		ActualValue: fmt.Sprintf("%d", failCount),
	}, nil
}

// checkReferential validates referential integrity between two tables.
func (e *RuleEngine) checkReferential(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.Referential == nil {
		return nil, fmt.Errorf("referential rule requires referential config")
	}

	// Count rows in the source where the FK column has no matching PK in the reference table.
	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM %s AS src LEFT JOIN %s AS ref ON src.%s = ref.%s WHERE ref.%s IS NULL AND src.%s IS NOT NULL",
		rule.Spec.AssetRef,
		rule.Spec.Referential.ReferenceTable,
		rule.Spec.Referential.SourceColumn,
		rule.Spec.Referential.ReferenceColumn,
		rule.Spec.Referential.ReferenceColumn,
		rule.Spec.Referential.SourceColumn,
	)
	orphanCount, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, query)
	if err != nil {
		return nil, fmt.Errorf("referential check query failed: %w", err)
	}

	return &CheckOutput{
		Passed:      orphanCount == 0,
		FailCount:   orphanCount,
		Message:     fmt.Sprintf("referential integrity: %d orphan rows in %s.%s (missing in %s.%s)", orphanCount, rule.Spec.AssetRef, rule.Spec.Referential.SourceColumn, rule.Spec.Referential.ReferenceTable, rule.Spec.Referential.ReferenceColumn),
		ActualValue: fmt.Sprintf("%d", orphanCount),
	}, nil
}

// checkDistribution validates that a column's value distribution meets expectations.
// Uses coefficient of variation (stddev / mean) to detect anomalies.
func (e *RuleEngine) checkDistribution(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.Distribution == nil {
		return nil, fmt.Errorf("distribution rule requires distribution config")
	}

	// Get mean and stddev for the column.
	meanQuery := fmt.Sprintf("SELECT AVG(%s) FROM %s", rule.Spec.Distribution.Column, rule.Spec.AssetRef)
	mean, err := e.querier.QueryFloat(ctx, rule.Spec.DataSourceRef, meanQuery)
	if err != nil {
		return nil, fmt.Errorf("distribution mean query failed: %w", err)
	}

	stddevQuery := fmt.Sprintf("SELECT STDDEV(%s) FROM %s", rule.Spec.Distribution.Column, rule.Spec.AssetRef)
	stddev, err := e.querier.QueryFloat(ctx, rule.Spec.DataSourceRef, stddevQuery)
	if err != nil {
		return nil, fmt.Errorf("distribution stddev query failed: %w", err)
	}

	// Coefficient of variation.
	var cv float64
	if mean != 0 {
		cv = stddev / mean
	}

	maxCV := rule.Spec.Distribution.MaxCoefficientOfVariation
	if maxCV == 0 {
		maxCV = 2.0 // Default: flag distributions with CV > 200%
	}

	passed := cv <= maxCV

	return &CheckOutput{
		Passed:      passed,
		Message:     fmt.Sprintf("distribution of %s: mean=%.2f, stddev=%.2f, CV=%.2f (max: %.2f)", rule.Spec.Distribution.Column, mean, stddev, cv, maxCV),
		ActualValue: fmt.Sprintf("%.4f", cv),
	}, nil
}

// checkTimeliness validates that data arrives within the expected SLA window.
func (e *RuleEngine) checkTimeliness(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.Timeliness == nil {
		return nil, fmt.Errorf("timeliness rule requires timeliness config")
	}

	sla, err := time.ParseDuration(rule.Spec.Timeliness.ExpectedDelay)
	if err != nil {
		return nil, fmt.Errorf("invalid expected delay %q: %w", rule.Spec.Timeliness.ExpectedDelay, err)
	}

	// Measure the lag between the ingest timestamp column and the event timestamp column.
	query := fmt.Sprintf(
		"SELECT AVG(TIMESTAMPDIFF(SECOND, %s, %s)) FROM %s WHERE %s >= NOW() - INTERVAL 1 DAY",
		rule.Spec.Timeliness.EventColumn,
		rule.Spec.Timeliness.IngestColumn,
		rule.Spec.AssetRef,
		rule.Spec.Timeliness.IngestColumn,
	)
	avgLagSec, err := e.querier.QueryFloat(ctx, rule.Spec.DataSourceRef, query)
	if err != nil {
		return nil, fmt.Errorf("timeliness check query failed: %w", err)
	}

	avgLag := time.Duration(int64(avgLagSec)) * time.Second
	passed := avgLag <= sla

	return &CheckOutput{
		Passed:      passed,
		Message:     fmt.Sprintf("timeliness: avg ingest lag %s (SLA: %s)", avgLag.Truncate(time.Second), sla),
		ActualValue: avgLag.Truncate(time.Second).String(),
	}, nil
}

// checkAcceptedValues validates that a column contains only values from an allowed set.
func (e *RuleEngine) checkAcceptedValues(ctx context.Context, rule *QualityRuleResource) (*CheckOutput, error) {
	if rule.Spec.AcceptedValues == nil {
		return nil, fmt.Errorf("accepted_values rule requires acceptedValues config")
	}

	// Build the NOT IN clause.
	var quoted []string
	for _, v := range rule.Spec.AcceptedValues.Values {
		quoted = append(quoted, fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''")))
	}

	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM %s WHERE %s NOT IN (%s)",
		rule.Spec.AssetRef,
		rule.Spec.AcceptedValues.Column,
		joinStrings(quoted, ", "),
	)
	failCount, err := e.querier.QueryRows(ctx, rule.Spec.DataSourceRef, query)
	if err != nil {
		return nil, fmt.Errorf("accepted_values check query failed: %w", err)
	}

	return &CheckOutput{
		Passed:      failCount == 0,
		FailCount:   failCount,
		Message:     fmt.Sprintf("accepted values in %s: %d rows have unexpected values", rule.Spec.AcceptedValues.Column, failCount),
		ActualValue: fmt.Sprintf("%d", failCount),
	}, nil
}

// --- Spec types for extended checks ---

// SchemaCheckConfig defines expected columns for schema validation.
type SchemaCheckConfig struct {
	ExpectedColumns []ExpectedColumn `json:"expectedColumns"`
	StrictMode      bool             `json:"strictMode,omitempty"` // Fail if extra columns exist
}

// ExpectedColumn defines a single expected column.
type ExpectedColumn struct {
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`     // Expected data type
	Nullable bool   `json:"nullable,omitempty"` // Whether NULL is allowed
}

// RegexCheckConfig defines regex pattern validation for a column.
type RegexCheckConfig struct {
	Column  string `json:"column"`
	Pattern string `json:"pattern"` // Go-compatible regex
}

// ReferentialCheckConfig defines referential integrity check parameters.
type ReferentialCheckConfig struct {
	SourceColumn    string `json:"sourceColumn"`
	ReferenceTable  string `json:"referenceTable"`
	ReferenceColumn string `json:"referenceColumn"`
}

// DistributionCheckConfig defines statistical distribution check parameters.
type DistributionCheckConfig struct {
	Column                      string  `json:"column"`
	MaxCoefficientOfVariation   float64 `json:"maxCoefficientOfVariation,omitempty"`
}

// TimelinessCheckConfig defines data timeliness SLA check parameters.
type TimelinessCheckConfig struct {
	EventColumn   string `json:"eventColumn"`   // When the event happened
	IngestColumn  string `json:"ingestColumn"`  // When it was ingested
	ExpectedDelay string `json:"expectedDelay"` // Max acceptable lag ("5m", "1h")
}

// AcceptedValuesCheckConfig defines allowed values for a column.
type AcceptedValuesCheckConfig struct {
	Column string   `json:"column"`
	Values []string `json:"values"`
}
