// Package quality — declarative dataset rule engine.
//
// The existing validator.go operates on individual records and rules
// expressed as Go struct literals.  This file adds a higher-level
// engine that takes a YAML/JSON rule document and evaluates it against
// a dataset expressed as []map[string]interface{}.  The rule surface
// is deliberately close to Great Expectations / Soda so that operators
// familiar with those tools can onboard quickly.
//
// Supported checks:
//
//   - not_null            column value must be non-nil and non-empty
//   - unique              all rows must have distinct value for column
//   - range               numeric value must fall within [min,max]
//   - regex               string value must match the given pattern
//   - allowed_values      value must be in the supplied list
//   - row_count_min       dataset row count must be >= value
//   - row_count_max       dataset row count must be <= value
//   - freshness           latest timestamp column must be within max_age
//
// Rules are evaluated independently; each produces a RuleResult
// containing pass/fail, violating-row sample, and diagnostic context.
// The engine never panics on bad input — malformed rules produce a
// RuleResult with Ok=false and Err set to the parse/eval error.
package quality

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// RuleKind enumerates the supported check types.
type RuleKind string

const (
	RuleNotNull       RuleKind = "not_null"
	RuleUnique        RuleKind = "unique"
	RuleRange         RuleKind = "range"
	RuleRegex         RuleKind = "regex"
	RuleAllowedValues RuleKind = "allowed_values"
	RuleRowCountMin   RuleKind = "row_count_min"
	RuleRowCountMax   RuleKind = "row_count_max"
	RuleFreshness     RuleKind = "freshness"
)

// Rule is the declarative unit evaluated by the engine.  Only the
// fields relevant to the chosen Kind are consulted; the others may be
// left zero.
type Rule struct {
	// ID is a stable identifier used in results and dashboards.  Two
	// rules with the same ID overwrite each other in the engine; if
	// left blank a deterministic ID is derived from Kind+Column.
	ID string `json:"id" yaml:"id"`

	// Kind selects the evaluator.
	Kind RuleKind `json:"kind" yaml:"kind"`

	// Column is the dotted column path evaluated by record-level rules.
	// Ignored by row_count_* rules.
	Column string `json:"column" yaml:"column"`

	// Severity controls how violations are reported upstream
	// ("info" | "warn" | "error" | "critical").  Defaults to "error".
	Severity string `json:"severity,omitempty" yaml:"severity,omitempty"`

	// Range rule parameters.
	Min *float64 `json:"min,omitempty" yaml:"min,omitempty"`
	Max *float64 `json:"max,omitempty" yaml:"max,omitempty"`

	// Regex / AllowedValues parameters.
	Pattern       string        `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	AllowedValues []interface{} `json:"allowed_values,omitempty" yaml:"allowed_values,omitempty"`

	// RowCount* parameters.
	RowCount int `json:"row_count,omitempty" yaml:"row_count,omitempty"`

	// Freshness parameters.
	MaxAge time.Duration `json:"max_age,omitempty" yaml:"max_age,omitempty"`
}

// RuleResult is the outcome of evaluating a single rule against a
// dataset.
type RuleResult struct {
	RuleID     string                 `json:"rule_id"`
	Kind       RuleKind               `json:"kind"`
	Column     string                 `json:"column,omitempty"`
	Severity   string                 `json:"severity"`
	Ok         bool                   `json:"ok"`
	Violations int                    `json:"violations"`
	SampleRow  map[string]interface{} `json:"sample_row,omitempty"`
	DurationMs int64                  `json:"duration_ms"`
	Err        string                 `json:"error,omitempty"`
}

// Report aggregates per-rule results.
type Report struct {
	Dataset     string       `json:"dataset"`
	Rows        int          `json:"rows"`
	EvaluatedAt time.Time    `json:"evaluated_at"`
	Results     []RuleResult `json:"results"`
}

// OverallOk reports whether every rule in the report passed.
func (r *Report) OverallOk() bool {
	for _, res := range r.Results {
		if !res.Ok {
			return false
		}
	}
	return true
}

// Engine evaluates rule sets against datasets.  Rule registration is
// mutex-protected so rules may be added/removed at runtime; Evaluate
// itself takes a snapshot and releases the lock before any per-row work.
type Engine struct {
	mu    sync.RWMutex
	rules map[string]Rule
}

// NewEngine creates an empty engine.
func NewEngine() *Engine {
	return &Engine{rules: make(map[string]Rule)}
}

// Register adds or replaces a rule.  An empty ID is populated from
// Kind+Column so operators can write terse YAML.
func (e *Engine) Register(r Rule) {
	if r.ID == "" {
		r.ID = fmt.Sprintf("%s:%s", r.Kind, r.Column)
	}
	if r.Severity == "" {
		r.Severity = "error"
	}
	e.mu.Lock()
	e.rules[r.ID] = r
	e.mu.Unlock()
}

// Remove deletes a rule by ID.  Unknown IDs are a no-op.
func (e *Engine) Remove(id string) {
	e.mu.Lock()
	delete(e.rules, id)
	e.mu.Unlock()
}

// Evaluate runs every registered rule against rows.  The dataset name
// is included verbatim in the Report for downstream routing.
func (e *Engine) Evaluate(dataset string, rows []map[string]interface{}) Report {
	e.mu.RLock()
	snapshot := make([]Rule, 0, len(e.rules))
	for _, r := range e.rules {
		snapshot = append(snapshot, r)
	}
	e.mu.RUnlock()

	// Stable rule ordering for reproducibility.
	sort.Slice(snapshot, func(i, j int) bool { return snapshot[i].ID < snapshot[j].ID })

	report := Report{
		Dataset:     dataset,
		Rows:        len(rows),
		EvaluatedAt: time.Now().UTC(),
		Results:     make([]RuleResult, 0, len(snapshot)),
	}
	for _, r := range snapshot {
		report.Results = append(report.Results, evaluate(r, rows))
	}
	return report
}

// evaluate runs one rule; returns a RuleResult whose Err captures any
// malformed-rule condition.  The function never allocates a goroutine
// — callers that want parallelism should shard the rule set externally.
func evaluate(r Rule, rows []map[string]interface{}) RuleResult {
	start := time.Now()
	res := RuleResult{RuleID: r.ID, Kind: r.Kind, Column: r.Column, Severity: r.Severity, Ok: true}
	defer func() { res.DurationMs = time.Since(start).Milliseconds() }()

	switch r.Kind {
	case RuleNotNull:
		for _, row := range rows {
			v, ok := row[r.Column]
			if !ok || v == nil || isEmptyString(v) {
				res.Ok = false
				res.Violations++
				if res.SampleRow == nil {
					res.SampleRow = row
				}
			}
		}

	case RuleUnique:
		seen := make(map[string]struct{}, len(rows))
		for _, row := range rows {
			key := fmt.Sprint(row[r.Column])
			if _, dup := seen[key]; dup {
				res.Ok = false
				res.Violations++
				if res.SampleRow == nil {
					res.SampleRow = row
				}
				continue
			}
			seen[key] = struct{}{}
		}

	case RuleRange:
		if r.Min == nil && r.Max == nil {
			res.Ok, res.Err = false, "range rule requires at least one of min / max"
			return res
		}
		for _, row := range rows {
			f, ok := toFloat(row[r.Column])
			if !ok {
				res.Ok = false
				res.Violations++
				if res.SampleRow == nil {
					res.SampleRow = row
				}
				continue
			}
			if (r.Min != nil && f < *r.Min) || (r.Max != nil && f > *r.Max) {
				res.Ok = false
				res.Violations++
				if res.SampleRow == nil {
					res.SampleRow = row
				}
			}
		}

	case RuleRegex:
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			res.Ok, res.Err = false, fmt.Sprintf("invalid regex: %v", err)
			return res
		}
		for _, row := range rows {
			s, ok := row[r.Column].(string)
			if !ok || !re.MatchString(s) {
				res.Ok = false
				res.Violations++
				if res.SampleRow == nil {
					res.SampleRow = row
				}
			}
		}

	case RuleAllowedValues:
		if len(r.AllowedValues) == 0 {
			res.Ok, res.Err = false, "allowed_values rule requires a non-empty list"
			return res
		}
		allowed := make(map[string]struct{}, len(r.AllowedValues))
		for _, v := range r.AllowedValues {
			allowed[fmt.Sprint(v)] = struct{}{}
		}
		for _, row := range rows {
			key := fmt.Sprint(row[r.Column])
			if _, ok := allowed[key]; !ok {
				res.Ok = false
				res.Violations++
				if res.SampleRow == nil {
					res.SampleRow = row
				}
			}
		}

	case RuleRowCountMin:
		if len(rows) < r.RowCount {
			res.Ok = false
			res.Violations = r.RowCount - len(rows)
		}

	case RuleRowCountMax:
		if len(rows) > r.RowCount {
			res.Ok = false
			res.Violations = len(rows) - r.RowCount
		}

	case RuleFreshness:
		if r.MaxAge <= 0 {
			res.Ok, res.Err = false, "freshness rule requires max_age > 0"
			return res
		}
		var latest time.Time
		for _, row := range rows {
			t, ok := toTime(row[r.Column])
			if !ok {
				continue
			}
			if t.After(latest) {
				latest = t
			}
		}
		if latest.IsZero() || time.Since(latest) > r.MaxAge {
			res.Ok = false
			res.Violations = 1
		}

	default:
		res.Ok, res.Err = false, fmt.Sprintf("unsupported rule kind %q", r.Kind)
	}
	return res
}

// isEmptyString returns true for "" and strings consisting only of
// whitespace; other types are treated as non-empty.
func isEmptyString(v interface{}) bool {
	s, ok := v.(string)
	return ok && strings.TrimSpace(s) == ""
}

// toFloat attempts to coerce v to float64.  Accepts int/int64/float32/
// float64/string; returns ok=false for anything else.
func toFloat(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case string:
		var f float64
		if _, err := fmt.Sscanf(x, "%g", &f); err == nil {
			return f, true
		}
	}
	return 0, false
}

// toTime parses v as an RFC3339 timestamp.  Accepts time.Time values
// directly and string forms.
func toTime(v interface{}) (time.Time, bool) {
	switch x := v.(type) {
	case time.Time:
		return x, true
	case string:
		if t, err := time.Parse(time.RFC3339Nano, x); err == nil {
			return t, true
		}
		if t, err := time.Parse(time.RFC3339, x); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
