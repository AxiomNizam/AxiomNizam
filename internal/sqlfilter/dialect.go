package sqlfilter

import "strings"

// Dialect represents a SQL database dialect.
type Dialect string

const (
	DialectMySQL      Dialect = "mysql"
	DialectPostgreSQL Dialect = "postgresql"
	DialectMariaDB    Dialect = "mariadb"
	DialectOracle     Dialect = "oracle"
	DialectPercona    Dialect = "percona"
	DialectGeneric    Dialect = "generic"
)

// DialectConfig holds dialect-specific validation rules.
type DialectConfig struct {
	Dialect           Dialect  `json:"dialect"`
	AllowedFunctions  []string `json:"allowedFunctions,omitempty"`
	BlockedFunctions  []string `json:"blockedFunctions,omitempty"`
	MaxQueryLength    int      `json:"maxQueryLength,omitempty"`
	AllowCTE          bool     `json:"allowCTE"`
	AllowWindowFuncs  bool     `json:"allowWindowFuncs"`
	AllowLateralJoin  bool     `json:"allowLateralJoin"`
	AllowRecursiveCTE bool     `json:"allowRecursiveCTE"`
	IdentifierQuote   string   `json:"identifierQuote"` // ` for MySQL, " for PostgreSQL
}

// DefaultDialectConfig returns the default config for a dialect.
func DefaultDialectConfig(dialect Dialect) DialectConfig {
	switch dialect {
	case DialectMySQL, DialectMariaDB, DialectPercona:
		return DialectConfig{
			Dialect:           dialect,
			AllowCTE:          true,
			AllowWindowFuncs:  true,
			AllowLateralJoin:  true,
			AllowRecursiveCTE: true,
			MaxQueryLength:    1048576, // 1MB
			IdentifierQuote:   "`",
			BlockedFunctions: []string{
				"SLEEP", "BENCHMARK", "LOAD_FILE", "INTO OUTFILE",
				"INTO DUMPFILE", "SYSTEM", "EXEC",
			},
		}
	case DialectPostgreSQL:
		return DialectConfig{
			Dialect:           dialect,
			AllowCTE:          true,
			AllowWindowFuncs:  true,
			AllowLateralJoin:  true,
			AllowRecursiveCTE: true,
			MaxQueryLength:    1048576,
			IdentifierQuote:   "\"",
			BlockedFunctions: []string{
				"PG_SLEEP", "PG_READ_FILE", "PG_LS_DIR",
				"LO_IMPORT", "LO_EXPORT", "COPY",
			},
		}
	case DialectOracle:
		return DialectConfig{
			Dialect:           dialect,
			AllowCTE:          true,
			AllowWindowFuncs:  true,
			AllowLateralJoin:  false,
			AllowRecursiveCTE: true,
			MaxQueryLength:    1048576,
			IdentifierQuote:   "\"",
			BlockedFunctions: []string{
				"DBMS_PIPE", "UTL_HTTP", "UTL_FILE", "UTL_TCP",
				"DBMS_SCHEDULER", "DBMS_JOB",
			},
		}
	default:
		return DialectConfig{
			Dialect:           DialectGeneric,
			AllowCTE:          true,
			AllowWindowFuncs:  true,
			AllowLateralJoin:  true,
			AllowRecursiveCTE: true,
			MaxQueryLength:    1048576,
			IdentifierQuote:   "\"",
		}
	}
}

// DialectValidationResult holds dialect-specific validation issues.
type DialectValidationResult struct {
	Valid    bool     `json:"valid"`
	Dialect Dialect  `json:"dialect"`
	Errors  []string `json:"errors,omitempty"`
}

// ValidateForDialect checks a query against dialect-specific rules.
func ValidateForDialect(query string, dialect Dialect) DialectValidationResult {
	cfg := DefaultDialectConfig(dialect)
	result := DialectValidationResult{Valid: true, Dialect: dialect}
	upper := strings.ToUpper(query)

	// Check max length.
	if cfg.MaxQueryLength > 0 && len(query) > cfg.MaxQueryLength {
		result.Valid = false
		result.Errors = append(result.Errors, "query exceeds maximum length")
	}

	// Check blocked functions.
	for _, fn := range cfg.BlockedFunctions {
		if strings.Contains(upper, strings.ToUpper(fn)) {
			result.Valid = false
			result.Errors = append(result.Errors, "blocked function: "+fn)
		}
	}

	// Check CTE support.
	if !cfg.AllowCTE && strings.Contains(upper, "WITH ") && strings.Contains(upper, " AS (") {
		result.Valid = false
		result.Errors = append(result.Errors, "CTEs not allowed for this dialect")
	}

	// Check window function support.
	if !cfg.AllowWindowFuncs && (strings.Contains(upper, " OVER (") || strings.Contains(upper, " OVER(")) {
		result.Valid = false
		result.Errors = append(result.Errors, "window functions not allowed for this dialect")
	}

	// Check LATERAL join support.
	if !cfg.AllowLateralJoin && strings.Contains(upper, " LATERAL ") {
		result.Valid = false
		result.Errors = append(result.Errors, "LATERAL joins not allowed for this dialect")
	}

	return result
}

// DetectDialect attempts to identify the SQL dialect from query syntax.
func DetectDialect(query string) Dialect {
	upper := strings.ToUpper(query)

	// PostgreSQL indicators.
	if strings.Contains(upper, "::") || // type cast
		strings.Contains(upper, "ILIKE") ||
		strings.Contains(upper, "RETURNING") ||
		strings.Contains(upper, "SERIAL") ||
		strings.Contains(upper, "JSONB") {
		return DialectPostgreSQL
	}

	// MySQL/MariaDB indicators.
	if strings.Contains(upper, "LIMIT ") && strings.Contains(upper, " OFFSET ") && !strings.Contains(upper, "FETCH") ||
		strings.Contains(query, "`") || // backtick identifiers
		strings.Contains(upper, "AUTO_INCREMENT") ||
		strings.Contains(upper, "ENGINE=") ||
		strings.Contains(upper, "IFNULL(") {
		return DialectMySQL
	}

	// Oracle indicators.
	if strings.Contains(upper, "ROWNUM") ||
		strings.Contains(upper, "NVL(") ||
		strings.Contains(upper, "SYSDATE") ||
		strings.Contains(upper, "DUAL") ||
		strings.Contains(upper, "CONNECT BY") {
		return DialectOracle
	}

	return DialectGeneric
}
