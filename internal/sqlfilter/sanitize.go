package sqlfilter

import (
	"strings"
	"unicode"
)

// SanitizeResult holds the outcome of query sanitization.
type SanitizeResult struct {
	Original  string   `json:"original"`
	Sanitized string   `json:"sanitized"`
	Changes   []string `json:"changes,omitempty"`
	Safe      bool     `json:"safe"`
}

// Sanitize cleans a SQL query by removing dangerous patterns while
// preserving the query's intent. This is a last-resort defense —
// parameterized queries are always preferred.
func Sanitize(query string) SanitizeResult {
	result := SanitizeResult{Original: query, Safe: true}
	sanitized := query

	// Remove null bytes.
	if strings.ContainsRune(sanitized, 0) {
		sanitized = strings.ReplaceAll(sanitized, "\x00", "")
		result.Changes = append(result.Changes, "removed null bytes")
	}

	// Remove inline comments used for evasion.
	if strings.Contains(sanitized, "/**/") {
		sanitized = strings.ReplaceAll(sanitized, "/**/", " ")
		result.Changes = append(result.Changes, "removed inline comment evasion (/**/)")
		result.Safe = false
	}

	// Remove MySQL version-specific comments /*!...*/
	for strings.Contains(sanitized, "/*!") {
		start := strings.Index(sanitized, "/*!")
		end := strings.Index(sanitized[start:], "*/")
		if end < 0 {
			break
		}
		// Extract the content between /*! and */
		content := sanitized[start+3 : start+end]
		// Remove version number prefix if present.
		content = strings.TrimLeftFunc(content, unicode.IsDigit)
		sanitized = sanitized[:start] + " " + strings.TrimSpace(content) + " " + sanitized[start+end+2:]
		result.Changes = append(result.Changes, "expanded MySQL version comment")
		result.Safe = false
	}

	// Collapse multiple spaces.
	sanitized = collapseWhitespace(sanitized)

	// Trim trailing semicolons (prevent stacked queries).
	trimmed := strings.TrimRight(strings.TrimSpace(sanitized), ";")
	if trimmed != strings.TrimSpace(sanitized) {
		sanitized = trimmed
		result.Changes = append(result.Changes, "removed trailing semicolons")
	}

	// Remove stacked queries (keep only the first statement).
	if CountStatements(sanitized) > 1 {
		// Find the first semicolon outside of strings.
		firstEnd := findFirstStatementEnd(sanitized)
		if firstEnd > 0 {
			sanitized = strings.TrimSpace(sanitized[:firstEnd])
			result.Changes = append(result.Changes, "removed stacked queries (kept first statement only)")
			result.Safe = false
		}
	}

	result.Sanitized = strings.TrimSpace(sanitized)
	return result
}

// EscapeIdentifier escapes a SQL identifier for safe use in queries.
// Uses the appropriate quoting for the dialect.
func EscapeIdentifier(name string, dialect Dialect) string {
	// Remove any existing quotes.
	name = strings.Trim(name, "`\"[]")

	// Remove dangerous characters.
	var clean strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' {
			clean.WriteRune(r)
		}
	}
	escaped := clean.String()

	switch dialect {
	case DialectMySQL, DialectMariaDB, DialectPercona:
		return "`" + strings.ReplaceAll(escaped, "`", "``") + "`"
	case DialectOracle, DialectPostgreSQL:
		return "\"" + strings.ReplaceAll(escaped, "\"", "\"\"") + "\""
	default:
		return "\"" + strings.ReplaceAll(escaped, "\"", "\"\"") + "\""
	}
}

// EscapeString escapes a string value for safe use in SQL.
// WARNING: Always prefer parameterized queries. This is for edge cases only.
func EscapeString(value string) string {
	var result strings.Builder
	result.WriteByte('\'')
	for _, r := range value {
		switch r {
		case '\'':
			result.WriteString("''")
		case '\\':
			result.WriteString("\\\\")
		case 0:
			// Skip null bytes.
		default:
			result.WriteRune(r)
		}
	}
	result.WriteByte('\'')
	return result.String()
}

// findFirstStatementEnd finds the position of the first statement-ending
// semicolon that's not inside a string literal.
func findFirstStatementEnd(query string) int {
	inSingle := false
	inDouble := false
	prev := byte(0)

	for i := 0; i < len(query); i++ {
		ch := query[i]
		if ch == '\'' && !inDouble && prev != '\\' {
			inSingle = !inSingle
		}
		if ch == '"' && !inSingle && prev != '\\' {
			inDouble = !inDouble
		}
		if !inSingle && !inDouble && ch == ';' {
			return i
		}
		prev = ch
	}
	return -1
}
