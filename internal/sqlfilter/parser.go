package sqlfilter

import "strings"

// FirstKeyword extracts the first SQL keyword from a query,
// skipping comments and leading parentheses.
func FirstKeyword(query string) string {
	remaining := StripLeadingComments(strings.TrimSpace(query))

	// Skip leading parentheses (subquery wrappers).
	for strings.HasPrefix(remaining, "(") {
		remaining = strings.TrimSpace(remaining[1:])
		remaining = StripLeadingComments(remaining)
	}
	if remaining == "" {
		return ""
	}

	// Extract the first word (letters + underscore).
	idx := 0
	for idx < len(remaining) {
		ch := remaining[idx]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
			idx++
			continue
		}
		break
	}
	if idx == 0 {
		return ""
	}
	return strings.ToUpper(remaining[:idx])
}

// StripLeadingComments removes SQL comments from the beginning of a query.
// Handles --, #, and /* */ comment styles.
func StripLeadingComments(query string) string {
	s := strings.TrimSpace(query)
	for {
		s = strings.TrimSpace(s)
		if s == "" {
			return ""
		}
		// Single-line comment: --
		if strings.HasPrefix(s, "--") {
			if i := strings.IndexByte(s, '\n'); i >= 0 {
				s = s[i+1:]
				continue
			}
			return ""
		}
		// MySQL single-line comment: #
		if strings.HasPrefix(s, "#") {
			if i := strings.IndexByte(s, '\n'); i >= 0 {
				s = s[i+1:]
				continue
			}
			return ""
		}
		// Block comment: /* */
		if strings.HasPrefix(s, "/*") {
			if i := strings.Index(s, "*/"); i >= 0 {
				s = s[i+2:]
				continue
			}
			return ""
		}
		return s
	}
}

// CountStatements counts the number of SQL statements in a query
// by tracking semicolons outside of string literals.
func CountStatements(query string) int {
	if strings.TrimSpace(query) == "" {
		return 0
	}

	count := 0
	tokenSeen := false
	inSingle := false
	inDouble := false
	inBacktick := false
	prev := byte(0)

	for i := 0; i < len(query); i++ {
		ch := query[i]

		// Track string literals.
		switch {
		case ch == '\'' && !inDouble && !inBacktick && prev != '\\':
			inSingle = !inSingle
		case ch == '"' && !inSingle && !inBacktick && prev != '\\':
			inDouble = !inDouble
		case ch == '`' && !inSingle && !inDouble:
			inBacktick = !inBacktick
		}

		if inSingle || inDouble || inBacktick {
			prev = ch
			continue
		}

		if ch == ';' {
			if tokenSeen {
				count++
				tokenSeen = false
			}
		} else if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			tokenSeen = true
		}

		prev = ch
	}

	// Count the last statement if it doesn't end with semicolon.
	if tokenSeen {
		count++
	}

	return count
}

// CountPlaceholders counts ? placeholders in a SQL query.
func CountPlaceholders(query string) int {
	return strings.Count(query, "?")
}
