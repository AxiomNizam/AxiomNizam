package sqlfilter

import (
	"regexp"
	"strings"
)

// InjectionRisk levels.
type InjectionRisk string

const (
	RiskNone     InjectionRisk = "none"
	RiskLow      InjectionRisk = "low"
	RiskMedium   InjectionRisk = "medium"
	RiskHigh     InjectionRisk = "high"
	RiskCritical InjectionRisk = "critical"
)

// InjectionResult holds the outcome of injection analysis.
type InjectionResult struct {
	Risk       InjectionRisk `json:"risk"`
	Score      int           `json:"score"` // 0-100
	Patterns   []InjectionPattern `json:"patterns,omitempty"`
	Suggestion string        `json:"suggestion,omitempty"`
}

// InjectionPattern describes a detected injection pattern.
type InjectionPattern struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // low, medium, high, critical
	Position    int    `json:"position,omitempty"`
}

// DetectInjection analyzes a query for SQL injection patterns.
// This is defense-in-depth — parameterized queries are the primary defense.
func DetectInjection(query string) InjectionResult {
	result := InjectionResult{Risk: RiskNone, Score: 0}
	upper := strings.ToUpper(query)

	// Check each pattern category.
	for _, check := range injectionChecks {
		if check.matcher(upper, query) {
			result.Patterns = append(result.Patterns, check.pattern)
			result.Score += check.score
		}
	}

	// Determine risk level from score.
	switch {
	case result.Score >= 80:
		result.Risk = RiskCritical
		result.Suggestion = "query is almost certainly an injection attempt — reject immediately"
	case result.Score >= 50:
		result.Risk = RiskHigh
		result.Suggestion = "query contains multiple injection indicators — reject or require manual review"
	case result.Score >= 30:
		result.Risk = RiskMedium
		result.Suggestion = "query contains suspicious patterns — log and monitor"
	case result.Score >= 10:
		result.Risk = RiskLow
		result.Suggestion = "query has minor suspicious patterns — likely safe but monitor"
	default:
		result.Risk = RiskNone
	}

	if result.Score > 100 {
		result.Score = 100
	}

	return result
}

type injectionCheck struct {
	pattern InjectionPattern
	score   int
	matcher func(upper, raw string) bool
}

var injectionChecks = []injectionCheck{
	// Tautology attacks: OR 1=1, OR 'a'='a', OR true
	{
		pattern: InjectionPattern{Name: "tautology", Description: "always-true condition (OR 1=1, OR true)", Severity: "high"},
		score:   40,
		matcher: func(upper, _ string) bool {
			return tautologyPattern.MatchString(upper)
		},
	},
	// UNION-based injection
	{
		pattern: InjectionPattern{Name: "union_select", Description: "UNION SELECT — data exfiltration attempt", Severity: "critical"},
		score:   50,
		matcher: func(upper, _ string) bool {
			return strings.Contains(upper, "UNION") && strings.Contains(upper, "SELECT") &&
				!strings.HasPrefix(strings.TrimSpace(upper), "SELECT") // legitimate UNION queries start with SELECT
		},
	},
	// Time-based blind injection
	{
		pattern: InjectionPattern{Name: "time_based", Description: "time-based blind injection (SLEEP, BENCHMARK, WAITFOR)", Severity: "critical"},
		score:   60,
		matcher: func(upper, _ string) bool {
			return strings.Contains(upper, "SLEEP(") ||
				strings.Contains(upper, "BENCHMARK(") ||
				strings.Contains(upper, "WAITFOR DELAY") ||
				strings.Contains(upper, "PG_SLEEP(")
		},
	},
	// Stacked queries (semicolon injection)
	{
		pattern: InjectionPattern{Name: "stacked_queries", Description: "semicolon followed by new statement", Severity: "high"},
		score:   40,
		matcher: func(_, raw string) bool {
			return CountStatements(raw) > 1
		},
	},
	// Comment-based evasion
	{
		pattern: InjectionPattern{Name: "comment_evasion", Description: "inline comment used to bypass filters (/**/)", Severity: "medium"},
		score:   20,
		matcher: func(_, raw string) bool {
			// Inline comments between keywords: SEL/**/ECT
			return strings.Contains(raw, "/**/") || strings.Contains(raw, "/*!")
		},
	},
	// Hex/char encoding evasion
	{
		pattern: InjectionPattern{Name: "encoding_evasion", Description: "hex or CHAR() encoding to bypass filters", Severity: "high"},
		score:   35,
		matcher: func(upper, _ string) bool {
			return strings.Contains(upper, "CHAR(") ||
				strings.Contains(upper, "0X") ||
				strings.Contains(upper, "UNHEX(") ||
				strings.Contains(upper, "CONVERT(")
		},
	},
	// System function access
	{
		pattern: InjectionPattern{Name: "system_access", Description: "system function or variable access", Severity: "high"},
		score:   30,
		matcher: func(upper, raw string) bool {
			return strings.Contains(upper, "SYSTEM_USER") ||
				strings.Contains(upper, "SESSION_USER") ||
				strings.Contains(upper, "CURRENT_USER") ||
				strings.Contains(raw, "@@") ||
				strings.Contains(upper, "VERSION(") ||
				strings.Contains(upper, "DATABASE(") ||
				strings.Contains(upper, "USER(")
		},
	},
	// File operations
	{
		pattern: InjectionPattern{Name: "file_operation", Description: "file read/write attempt", Severity: "critical"},
		score:   70,
		matcher: func(upper, _ string) bool {
			return strings.Contains(upper, "INTO OUTFILE") ||
				strings.Contains(upper, "INTO DUMPFILE") ||
				strings.Contains(upper, "LOAD_FILE(") ||
				strings.Contains(upper, "LOAD DATA")
		},
	},
	// Schema enumeration
	{
		pattern: InjectionPattern{Name: "schema_enum", Description: "information_schema or system catalog access", Severity: "medium"},
		score:   15,
		matcher: func(upper, _ string) bool {
			return strings.Contains(upper, "INFORMATION_SCHEMA") ||
				strings.Contains(upper, "PG_CATALOG") ||
				strings.Contains(upper, "SYS.") ||
				strings.Contains(upper, "MYSQL.USER") ||
				strings.Contains(upper, "ALL_TABLES")
		},
	},
	// Error-based injection
	{
		pattern: InjectionPattern{Name: "error_based", Description: "error-based injection (EXTRACTVALUE, UPDATEXML)", Severity: "high"},
		score:   45,
		matcher: func(upper, _ string) bool {
			return strings.Contains(upper, "EXTRACTVALUE(") ||
				strings.Contains(upper, "UPDATEXML(") ||
				strings.Contains(upper, "XMLTYPE(") ||
				strings.Contains(upper, "EXP(~")
		},
	},
	// Boolean-based blind
	{
		pattern: InjectionPattern{Name: "boolean_blind", Description: "boolean-based blind injection (AND 1=1, AND 1=2)", Severity: "medium"},
		score:   25,
		matcher: func(upper, _ string) bool {
			return booleanBlindPattern.MatchString(upper)
		},
	},
}

var tautologyPattern = regexp.MustCompile(`OR\s+['"]?\w+['"]?\s*=\s*['"]?\w+['"]?|OR\s+TRUE|OR\s+1\s*=\s*1`)
var booleanBlindPattern = regexp.MustCompile(`AND\s+\d+\s*=\s*\d+|AND\s+['"]?\w+['"]?\s*=\s*['"]?\w+['"]?`)
