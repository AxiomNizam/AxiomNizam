package sqlfilter

import (
	"fmt"
	"strings"
	"unicode"
)

// SyntaxError represents a SQL syntax issue found during validation.
type SyntaxError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("line %d, col %d: %s", e.Line, e.Column, e.Message)
}

// SyntaxCheckResult holds all syntax issues found.
type SyntaxCheckResult struct {
	Valid  bool          `json:"valid"`
	Errors []SyntaxError `json:"errors,omitempty"`
}

// CheckSyntax performs lightweight SQL syntax validation.
// This is NOT a full parser — it catches common mistakes that would
// cause runtime errors in any SQL engine.
func CheckSyntax(query string) SyntaxCheckResult {
	result := SyntaxCheckResult{Valid: true}

	if strings.TrimSpace(query) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, SyntaxError{
			Line: 1, Column: 1, Message: "empty query",
		})
		return result
	}

	// Check unbalanced quotes.
	if err := checkUnbalancedQuotes(query); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, *err)
	}

	// Check unbalanced parentheses.
	if err := checkUnbalancedParens(query); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, *err)
	}

	// Check trailing comma before FROM/WHERE/GROUP/ORDER/LIMIT.
	if errs := checkTrailingCommas(query); len(errs) > 0 {
		result.Valid = false
		result.Errors = append(result.Errors, errs...)
	}

	// Check common typos.
	if errs := checkCommonTypos(query); len(errs) > 0 {
		// Typos are warnings, not errors — don't set Valid=false.
		result.Errors = append(result.Errors, errs...)
	}

	return result
}

func checkUnbalancedQuotes(query string) *SyntaxError {
	singleCount := 0
	doubleCount := 0
	line, col := 1, 0

	for i, ch := range query {
		col++
		if ch == '\n' {
			line++
			col = 0
			continue
		}
		if ch == '\'' && (i == 0 || query[i-1] != '\\') {
			singleCount++
		}
		if ch == '"' && (i == 0 || query[i-1] != '\\') {
			doubleCount++
		}
	}

	if singleCount%2 != 0 {
		return &SyntaxError{
			Line: line, Column: col,
			Message: "unbalanced single quote (')",
			Hint:    "check for missing closing quote in string literal",
		}
	}
	if doubleCount%2 != 0 {
		return &SyntaxError{
			Line: line, Column: col,
			Message: "unbalanced double quote (\")",
			Hint:    "check for missing closing quote in identifier",
		}
	}
	return nil
}

func checkUnbalancedParens(query string) *SyntaxError {
	depth := 0
	line, col := 1, 0
	inSingle := false
	inDouble := false

	for i, ch := range query {
		col++
		if ch == '\n' {
			line++
			col = 0
			continue
		}

		if ch == '\'' && !inDouble && (i == 0 || query[i-1] != '\\') {
			inSingle = !inSingle
		}
		if ch == '"' && !inSingle && (i == 0 || query[i-1] != '\\') {
			inDouble = !inDouble
		}
		if inSingle || inDouble {
			continue
		}

		if ch == '(' {
			depth++
		}
		if ch == ')' {
			depth--
			if depth < 0 {
				return &SyntaxError{
					Line: line, Column: col,
					Message: "unexpected closing parenthesis",
					Hint:    "extra ')' without matching '('",
				}
			}
		}
	}

	if depth > 0 {
		return &SyntaxError{
			Line: line, Column: col,
			Message: fmt.Sprintf("unclosed parenthesis (%d open)", depth),
			Hint:    "missing ')' to close subquery or function call",
		}
	}
	return nil
}

func checkTrailingCommas(query string) []SyntaxError {
	var errors []SyntaxError
	upper := strings.ToUpper(query)

	// Pattern: comma followed by a clause keyword.
	clauseKeywords := []string{" FROM ", " WHERE ", " GROUP ", " ORDER ", " LIMIT ", " HAVING ", " UNION "}
	for _, kw := range clauseKeywords {
		idx := strings.Index(upper, kw)
		if idx <= 0 {
			continue
		}
		// Check if there's a comma just before the keyword (ignoring whitespace).
		before := strings.TrimRight(query[:idx], " \t\n\r")
		if len(before) > 0 && before[len(before)-1] == ',' {
			line := strings.Count(query[:idx], "\n") + 1
			errors = append(errors, SyntaxError{
				Line:    line,
				Column:  len(before),
				Message: fmt.Sprintf("trailing comma before %s", strings.TrimSpace(kw)),
				Hint:    "remove the comma before the clause keyword",
			})
		}
	}
	return errors
}

func checkCommonTypos(query string) []SyntaxError {
	var errors []SyntaxError
	words := strings.Fields(query)

	typos := map[string]string{
		"SELCT":   "SELECT",
		"SLECT":   "SELECT",
		"SELET":   "SELECT",
		"FORM":    "FROM",
		"FROME":   "FROM",
		"WEHRE":   "WHERE",
		"WHRE":    "WHERE",
		"GRUOP":   "GROUP",
		"GROPU":   "GROUP",
		"ODER":    "ORDER",
		"ORDR":    "ORDER",
		"LIMT":    "LIMIT",
		"INSRT":   "INSERT",
		"UDPATE":  "UPDATE",
		"DELTE":   "DELETE",
		"DELEETE": "DELETE",
	}

	for i, word := range words {
		upper := strings.ToUpper(word)
		// Strip trailing punctuation for matching.
		cleaned := strings.TrimRightFunc(upper, func(r rune) bool {
			return !unicode.IsLetter(r)
		})
		if suggestion, ok := typos[cleaned]; ok {
			errors = append(errors, SyntaxError{
				Line:    1,
				Column:  i + 1,
				Message: fmt.Sprintf("possible typo: %q — did you mean %s?", word, suggestion),
				Hint:    fmt.Sprintf("replace %q with %q", word, suggestion),
			})
		}
	}
	return errors
}
