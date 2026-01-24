package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SQLInjectionProtection provides SQL injection prevention utilities
type SQLInjectionProtection struct {
	// Forbidden SQL keywords and patterns that indicate injection attempts
	forbiddenKeywords []string
	// Maximum allowed query length
	maxQueryLength int
}

// NewSQLInjectionProtection creates a new SQL injection protection instance
func NewSQLInjectionProtection() *SQLInjectionProtection {
	return &SQLInjectionProtection{
		forbiddenKeywords: []string{
			"DROP", "DELETE", "TRUNCATE", "INSERT", "UPDATE", "ALTER",
			"EXEC", "EXECUTE", "UNION", "SELECT", "REPLACE",
			"--", "/*", "*/", ";", "xp_", "sp_", "|", "&",
		},
		maxQueryLength: 10000,
	}
}

// SanitizeIdentifier sanitizes table and column names
// Only allows alphanumeric characters and underscores
func (s *SQLInjectionProtection) SanitizeIdentifier(identifier string) (string, error) {
	if IsEmpty(identifier) {
		return "", fmt.Errorf("identifier cannot be empty")
	}

	// Only allow alphanumeric characters and underscores
	pattern := `^[a-zA-Z_][a-zA-Z0-9_]*$`
	matched, _ := regexp.MatchString(pattern, identifier)

	if !matched {
		return "", fmt.Errorf("invalid identifier format: %s", identifier)
	}

	return identifier, nil
}

// SanitizeTableName validates and returns a safe table name
func (s *SQLInjectionProtection) SanitizeTableName(tableName string) (string, error) {
	return s.SanitizeIdentifier(tableName)
}

// SanitizeColumnName validates and returns a safe column name
func (s *SQLInjectionProtection) SanitizeColumnName(columnName string) (string, error) {
	return s.SanitizeIdentifier(columnName)
}

// SanitizeValue sanitizes string values for SQL queries
// Returns a properly escaped value for SQL injection prevention
func (s *SQLInjectionProtection) SanitizeValue(value string) string {
	// Trim whitespace
	value = strings.TrimSpace(value)

	// Escape single quotes by doubling them
	value = strings.ReplaceAll(value, "'", "''")

	// Remove dangerous characters commonly used in SQL injection
	value = removeInjectablePatterns(value)

	return value
}

// ValidateSQLInput validates and checks for common SQL injection patterns
func (s *SQLInjectionProtection) ValidateSQLInput(input string) error {
	if IsEmpty(input) {
		return fmt.Errorf("input cannot be empty")
	}

	// Check length
	if len(input) > s.maxQueryLength {
		return fmt.Errorf("input exceeds maximum length of %d characters", s.maxQueryLength)
	}

	// Check for forbidden keywords in input
	upperInput := strings.ToUpper(input)
	for _, keyword := range s.forbiddenKeywords {
		if strings.Contains(upperInput, keyword) {
			// Check if it's not just part of a legitimate string
			if isRiskyKeyword(upperInput, keyword) {
				return fmt.Errorf("potentially dangerous SQL keyword detected: %s", keyword)
			}
		}
	}

	// Check for common SQL injection patterns
	if detectSQLInjectionPatterns(input) {
		return fmt.Errorf("potential SQL injection pattern detected")
	}

	return nil
}

// SanitizeWhereClause sanitizes WHERE clause conditions
func (s *SQLInjectionProtection) SanitizeWhereClause(clause string) error {
	if IsEmpty(clause) {
		return fmt.Errorf("WHERE clause cannot be empty")
	}

	return s.ValidateSQLInput(clause)
}

// SanitizeOrderBy validates ORDER BY clause
func (s *SQLInjectionProtection) SanitizeOrderBy(orderBy string) (string, error) {
	if IsEmpty(orderBy) {
		return "", fmt.Errorf("ORDER BY clause cannot be empty")
	}

	// Only allow column names, ASC, DESC
	pattern := `^[a-zA-Z_][a-zA-Z0-9_]*(\s+(ASC|DESC))?(\s*,\s*[a-zA-Z_][a-zA-Z0-9_]*(\s+(ASC|DESC))?)*$`
	matched, _ := regexp.MatchString(pattern, orderBy)

	if !matched {
		return "", fmt.Errorf("invalid ORDER BY clause format")
	}

	return orderBy, nil
}

// SanitizeLimitOffset validates LIMIT and OFFSET values
func (s *SQLInjectionProtection) SanitizeLimitOffset(limit, offset interface{}) (int, int, error) {
	limitInt := 0
	offsetInt := 0

	// Validate limit
	switch v := limit.(type) {
	case int:
		limitInt = v
	case int64:
		limitInt = int(v)
	case string:
		val, err := strconv.Atoi(v)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid LIMIT value: %v", limit)
		}
		limitInt = val
	default:
		return 0, 0, fmt.Errorf("unsupported LIMIT type: %T", limit)
	}

	// Validate offset
	switch v := offset.(type) {
	case int:
		offsetInt = v
	case int64:
		offsetInt = int(v)
	case string:
		val, err := strconv.Atoi(v)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid OFFSET value: %v", offset)
		}
		offsetInt = val
	default:
		return 0, 0, fmt.Errorf("unsupported OFFSET type: %T", offset)
	}

	// Validate ranges
	if limitInt < 0 || limitInt > 10000 {
		return 0, 0, fmt.Errorf("LIMIT must be between 0 and 10000, got %d", limitInt)
	}

	if offsetInt < 0 || offsetInt > 1000000 {
		return 0, 0, fmt.Errorf("OFFSET must be between 0 and 1000000, got %d", offsetInt)
	}

	return limitInt, offsetInt, nil
}

// ValidateTableColumns validates a list of column names
func (s *SQLInjectionProtection) ValidateTableColumns(columns []string) error {
	if len(columns) == 0 {
		return fmt.Errorf("columns list cannot be empty")
	}

	for _, col := range columns {
		if _, err := s.SanitizeColumnName(col); err != nil {
			return err
		}
	}

	return nil
}

// removeInjectablePatterns removes characters commonly used in SQL injection
func removeInjectablePatterns(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove newlines and carriage returns
	input = strings.ReplaceAll(input, "\n", "")
	input = strings.ReplaceAll(input, "\r", "")

	// Remove tabs
	input = strings.ReplaceAll(input, "\t", " ")

	return input
}

// isRiskyKeyword determines if a keyword in the input is actually risky
func isRiskyKeyword(input, keyword string) bool {
	// These contexts don't necessarily indicate injection
	safeContexts := []string{
		"'%SELECT%'", // Inside a string literal
		"'%UNION%'",  // Inside a string literal
		"'%UPDATE%'", // Inside a string literal
		"'%DELETE%'", // Inside a string literal
	}

	for _, context := range safeContexts {
		if strings.Contains(input, context) {
			return false
		}
	}

	return true
}

// detectSQLInjectionPatterns detects common SQL injection patterns
func detectSQLInjectionPatterns(input string) bool {
	patterns := []string{
		`(?i)('\s*(OR|AND)\s*'?[^']*'?\s*=)`,           // ' OR '1'='1
		`(?i)(;\s*(DROP|DELETE|INSERT|UPDATE|CREATE))`, // ; DROP TABLE
		`(?i)(UNION\s+SELECT)`,                         // UNION SELECT
		`(?i)(--\s|#\s)`,                               // SQL comments
		`(?i)(\/\*.*?\*\/)`,                            // Block comments
		`(?i)(xp_|sp_)`,                                // Extended/system procedures
		`(?i)(CAST\s*\()`,                              // CAST injection
		`(?i)(CONVERT\s*\()`,                           // CONVERT injection
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}

	return false
}

// BuildSafeQuery builds a safe SQL query using parameterized queries
// This should be used instead of string concatenation
func BuildSafeQuery(baseQuery string, params []interface{}) (string, []interface{}, error) {
	if IsEmpty(baseQuery) {
		return "", nil, fmt.Errorf("base query cannot be empty")
	}

	// Validate the structure
	paramCount := strings.Count(baseQuery, "?")
	if paramCount != len(params) {
		return "", nil, fmt.Errorf("parameter count mismatch: query has %d placeholders but got %d params", paramCount, len(params))
	}

	return baseQuery, params, nil
}

// SanitizeSearchInput sanitizes user input for LIKE queries
func (s *SQLInjectionProtection) SanitizeSearchInput(input string) (string, error) {
	if IsEmpty(input) {
		return "", fmt.Errorf("search input cannot be empty")
	}

	// Remove SQL special characters
	sanitized := s.SanitizeValue(input)

	// Escape LIKE wildcards
	sanitized = strings.ReplaceAll(sanitized, "%", "\\%")
	sanitized = strings.ReplaceAll(sanitized, "_", "\\_")

	return sanitized, nil
}

// ValidateColumnFilter validates filter specifications for SELECT queries
func (s *SQLInjectionProtection) ValidateColumnFilter(filters map[string]interface{}) error {
	if len(filters) == 0 {
		return fmt.Errorf("filters cannot be empty")
	}

	for col := range filters {
		if _, err := s.SanitizeColumnName(col); err != nil {
			return err
		}
	}

	return nil
}
