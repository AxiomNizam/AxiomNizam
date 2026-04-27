package sqlfilter

import (
	"strings"
	"unicode"
)

// NormalizeResult holds the normalized query and metadata.
type NormalizeResult struct {
	Original    string `json:"original"`
	Normalized  string `json:"normalized"`
	Fingerprint string `json:"fingerprint"` // query shape for grouping
	Parameterized string `json:"parameterized"` // literals replaced with ?
}

// Normalize cleans up a SQL query for consistent comparison and logging.
// Removes extra whitespace, normalizes case for keywords, strips comments.
func Normalize(query string) NormalizeResult {
	result := NormalizeResult{Original: query}

	// Strip comments.
	cleaned := stripAllComments(query)

	// Collapse whitespace.
	cleaned = collapseWhitespace(cleaned)

	// Trim.
	cleaned = strings.TrimSpace(cleaned)

	result.Normalized = cleaned
	result.Fingerprint = fingerprint(cleaned)
	result.Parameterized = parameterize(cleaned)

	return result
}

// stripAllComments removes all SQL comments (not just leading ones).
func stripAllComments(query string) string {
	var result strings.Builder
	i := 0
	inSingle := false
	inDouble := false

	for i < len(query) {
		ch := query[i]

		// Track string literals.
		if ch == '\'' && !inDouble && (i == 0 || query[i-1] != '\\') {
			inSingle = !inSingle
			result.WriteByte(ch)
			i++
			continue
		}
		if ch == '"' && !inSingle && (i == 0 || query[i-1] != '\\') {
			inDouble = !inDouble
			result.WriteByte(ch)
			i++
			continue
		}
		if inSingle || inDouble {
			result.WriteByte(ch)
			i++
			continue
		}

		// Block comment.
		if i+1 < len(query) && ch == '/' && query[i+1] == '*' {
			end := strings.Index(query[i+2:], "*/")
			if end >= 0 {
				i = i + 2 + end + 2
				result.WriteByte(' ')
				continue
			}
			// Unclosed comment — skip rest.
			break
		}

		// Single-line comment: --
		if i+1 < len(query) && ch == '-' && query[i+1] == '-' {
			end := strings.IndexByte(query[i:], '\n')
			if end >= 0 {
				i = i + end + 1
				result.WriteByte(' ')
				continue
			}
			break
		}

		// MySQL comment: #
		if ch == '#' {
			end := strings.IndexByte(query[i:], '\n')
			if end >= 0 {
				i = i + end + 1
				result.WriteByte(' ')
				continue
			}
			break
		}

		result.WriteByte(ch)
		i++
	}

	return result.String()
}

// collapseWhitespace replaces runs of whitespace with a single space.
func collapseWhitespace(s string) string {
	var result strings.Builder
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				result.WriteByte(' ')
				prevSpace = true
			}
		} else {
			result.WriteRune(r)
			prevSpace = false
		}
	}
	return result.String()
}

// fingerprint creates a query shape by replacing all literals with ?.
// Useful for grouping similar queries in logs/metrics.
func fingerprint(query string) string {
	return parameterize(strings.ToUpper(query))
}

// parameterize replaces string literals and numbers with ? placeholders.
func parameterize(query string) string {
	var result strings.Builder
	i := 0

	for i < len(query) {
		ch := query[i]

		// String literal (single quote).
		if ch == '\'' {
			result.WriteByte('?')
			i++
			for i < len(query) {
				if query[i] == '\'' {
					if i+1 < len(query) && query[i+1] == '\'' {
						i += 2 // escaped quote
						continue
					}
					i++
					break
				}
				if query[i] == '\\' && i+1 < len(query) {
					i += 2
					continue
				}
				i++
			}
			continue
		}

		// Number literal.
		if (ch >= '0' && ch <= '9') || (ch == '.' && i+1 < len(query) && query[i+1] >= '0' && query[i+1] <= '9') {
			// Check it's not part of an identifier.
			if i > 0 && (isIdentChar(query[i-1])) {
				result.WriteByte(ch)
				i++
				continue
			}
			result.WriteByte('?')
			for i < len(query) && (query[i] >= '0' && query[i] <= '9' || query[i] == '.' || query[i] == 'e' || query[i] == 'E' || query[i] == '+' || query[i] == '-') {
				i++
			}
			continue
		}

		result.WriteByte(ch)
		i++
	}

	return result.String()
}

func isIdentChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || (ch >= '0' && ch <= '9')
}
