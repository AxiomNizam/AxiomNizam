// Package sqlfilter provides SQL query validation, classification,
// and syntax error detection for the AxiomNizam platform.
//
// Used by:
//   - API Builder (custom API SQL templates)
//   - ETL pipelines (transformation queries)
//   - Dynamic query handler (ad-hoc queries)
//   - CDC pipelines (source/sink queries)
//
// Architecture: standalone module following IAM/storage pattern.
// No external dependencies — pure Go string analysis.
package sqlfilter

import "strings"

// PolicyMode controls how strict the SQL validation is.
type PolicyMode string

const (
	// PolicyCompat allows legacy patterns that look safe.
	PolicyCompat PolicyMode = "compat"
	// PolicyStrict rejects anything not explicitly classified as read-only.
	PolicyStrict PolicyMode = "strict"
)

// QueryClass categorizes a SQL statement.
type QueryClass string

const (
	ClassRead    QueryClass = "read"
	ClassWrite   QueryClass = "write"
	ClassDDL     QueryClass = "ddl"
	ClassControl QueryClass = "control"
	ClassExec    QueryClass = "exec"
	ClassUnknown QueryClass = "unknown"
)

// ValidationResult holds the outcome of SQL validation.
type ValidationResult struct {
	Valid       bool       `json:"valid"`
	ReadOnly   bool       `json:"readOnly"`
	Class      QueryClass `json:"class"`
	Keyword    string     `json:"keyword"`
	Error      string     `json:"error,omitempty"`
	Warnings   []string   `json:"warnings,omitempty"`
	Statements int        `json:"statements"`
}

// Filter is the main SQL filter engine.
type Filter struct {
	mode PolicyMode
}

// New creates a Filter with the given policy mode.
func New(mode PolicyMode) *Filter {
	if mode == "" {
		mode = PolicyCompat
	}
	return &Filter{mode: mode}
}

// Validate checks a SQL query for safety and classification.
func (f *Filter) Validate(query string) ValidationResult {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return ValidationResult{Valid: false, Error: "empty query"}
	}

	result := ValidationResult{
		Statements: CountStatements(trimmed),
	}

	// Multiple statements are always rejected.
	if result.Statements > 1 {
		result.Valid = false
		result.Error = "multiple statements not allowed"
		return result
	}

	// Classify the query.
	keyword := FirstKeyword(trimmed)
	result.Keyword = keyword
	result.Class = Classify(keyword)

	// Check read-only.
	result.ReadOnly = f.IsReadOnly(trimmed)
	result.Valid = true

	// Add warnings for suspicious patterns.
	result.Warnings = f.detectWarnings(trimmed)

	return result
}

// IsReadOnly returns true if the query is safe for read-only execution.
func (f *Filter) IsReadOnly(query string) bool {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return false
	}

	if CountStatements(trimmed) > 1 {
		return false
	}

	keyword := FirstKeyword(trimmed)
	class := Classify(keyword)

	switch class {
	case ClassRead:
		// continue to blocklist check
	case ClassWrite, ClassDDL, ClassControl, ClassExec:
		return false
	case ClassUnknown:
		if f.mode == PolicyStrict {
			return false
		}
	}

	// Blocklist check.
	if containsBlockedKeywords(trimmed, f.mode) {
		return false
	}

	if class == ClassUnknown {
		return legacyReadOnlyHeuristic(trimmed)
	}

	return class == ClassRead
}

// detectWarnings finds suspicious patterns that aren't blocking.
func (f *Filter) detectWarnings(query string) []string {
	var warnings []string
	upper := strings.ToUpper(query)

	if strings.Contains(upper, "SLEEP(") || strings.Contains(upper, "BENCHMARK(") {
		warnings = append(warnings, "potential time-based injection pattern detected")
	}
	if strings.Contains(upper, "UNION") && strings.Contains(upper, "SELECT") {
		warnings = append(warnings, "UNION SELECT detected — verify intent")
	}
	if strings.Contains(upper, "INTO OUTFILE") || strings.Contains(upper, "INTO DUMPFILE") {
		warnings = append(warnings, "file write attempt detected")
	}
	if strings.Contains(upper, "INFORMATION_SCHEMA") {
		warnings = append(warnings, "schema introspection detected")
	}
	if strings.Contains(upper, "@@") || strings.Contains(upper, "SYSTEM_USER") {
		warnings = append(warnings, "system variable access detected")
	}

	return warnings
}
