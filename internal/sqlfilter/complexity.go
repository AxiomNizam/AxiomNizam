package sqlfilter

import (
	"strings"
)

// ComplexityLevel indicates how complex a query is.
type ComplexityLevel string

const (
	ComplexitySimple   ComplexityLevel = "simple"   // single table, no joins
	ComplexityModerate ComplexityLevel = "moderate" // 1-2 joins, subqueries
	ComplexityComplex  ComplexityLevel = "complex"  // 3+ joins, nested subqueries
	ComplexityExtreme  ComplexityLevel = "extreme"  // deeply nested, CTEs, window functions
)

// ComplexityResult holds the analysis of query complexity.
type ComplexityResult struct {
	Level          ComplexityLevel `json:"level"`
	Score          int             `json:"score"` // 0-100
	JoinCount      int             `json:"joinCount"`
	SubqueryDepth  int             `json:"subqueryDepth"`
	CTECount       int             `json:"cteCount"`
	UnionCount     int             `json:"unionCount"`
	WindowFunctions int            `json:"windowFunctions"`
	AggregateCount int             `json:"aggregateCount"`
	TableCount     int             `json:"tableCount"`
	ConditionCount int             `json:"conditionCount"`
	HasGroupBy     bool            `json:"hasGroupBy"`
	HasOrderBy     bool            `json:"hasOrderBy"`
	HasHaving      bool            `json:"hasHaving"`
	HasDistinct    bool            `json:"hasDistinct"`
	HasLimit       bool            `json:"hasLimit"`
	EstimatedCost  string          `json:"estimatedCost"` // "low", "medium", "high", "very_high"
}

// AnalyzeComplexity evaluates the structural complexity of a SQL query.
// Useful for rate limiting, timeout estimation, and query plan warnings.
func AnalyzeComplexity(query string) ComplexityResult {
	upper := strings.ToUpper(query)
	padded := " " + strings.Join(strings.Fields(upper), " ") + " "

	result := ComplexityResult{}

	// Count JOINs.
	result.JoinCount = strings.Count(padded, " JOIN ") +
		strings.Count(padded, " INNER JOIN ") +
		strings.Count(padded, " LEFT JOIN ") +
		strings.Count(padded, " RIGHT JOIN ") +
		strings.Count(padded, " FULL JOIN ") +
		strings.Count(padded, " CROSS JOIN ") +
		strings.Count(padded, " NATURAL JOIN ")
	// Avoid double-counting "LEFT JOIN" as both "JOIN" and "LEFT JOIN".
	result.JoinCount = strings.Count(padded, " JOIN ")

	// Count subquery depth (nested parentheses with SELECT inside).
	result.SubqueryDepth = countSubqueryDepth(query)

	// Count CTEs (WITH ... AS).
	result.CTECount = strings.Count(padded, " AS (")

	// Count UNIONs.
	result.UnionCount = strings.Count(padded, " UNION ")

	// Count window functions.
	result.WindowFunctions = strings.Count(padded, " OVER (") + strings.Count(padded, " OVER(")

	// Count aggregates.
	aggregates := []string{"COUNT(", "SUM(", "AVG(", "MIN(", "MAX(", "GROUP_CONCAT(", "STRING_AGG(", "ARRAY_AGG("}
	for _, agg := range aggregates {
		result.AggregateCount += strings.Count(upper, agg)
	}

	// Count tables (rough: count FROM + JOIN occurrences).
	result.TableCount = 1 + result.JoinCount
	if strings.Contains(padded, " FROM ") {
		// Multiple tables in FROM clause (comma-separated).
		fromIdx := strings.Index(padded, " FROM ")
		if fromIdx >= 0 {
			afterFrom := padded[fromIdx+6:]
			// Find the next clause keyword.
			endIdx := len(afterFrom)
			for _, kw := range []string{" WHERE ", " GROUP ", " ORDER ", " LIMIT ", " HAVING ", " UNION "} {
				if idx := strings.Index(afterFrom, kw); idx >= 0 && idx < endIdx {
					endIdx = idx
				}
			}
			fromClause := afterFrom[:endIdx]
			result.TableCount = strings.Count(fromClause, ",") + 1 + result.JoinCount
		}
	}

	// Count conditions (WHERE/AND/OR).
	result.ConditionCount = strings.Count(padded, " AND ") + strings.Count(padded, " OR ") + 1
	if !strings.Contains(padded, " WHERE ") {
		result.ConditionCount = 0
	}

	// Clause presence.
	result.HasGroupBy = strings.Contains(padded, " GROUP BY ")
	result.HasOrderBy = strings.Contains(padded, " ORDER BY ")
	result.HasHaving = strings.Contains(padded, " HAVING ")
	result.HasDistinct = strings.Contains(padded, " DISTINCT ")
	result.HasLimit = strings.Contains(padded, " LIMIT ") || strings.Contains(padded, " FETCH ")

	// Calculate score.
	result.Score = calculateComplexityScore(result)

	// Determine level.
	switch {
	case result.Score >= 75:
		result.Level = ComplexityExtreme
		result.EstimatedCost = "very_high"
	case result.Score >= 50:
		result.Level = ComplexityComplex
		result.EstimatedCost = "high"
	case result.Score >= 25:
		result.Level = ComplexityModerate
		result.EstimatedCost = "medium"
	default:
		result.Level = ComplexitySimple
		result.EstimatedCost = "low"
	}

	return result
}

func calculateComplexityScore(r ComplexityResult) int {
	score := 0

	// Joins: 8 points each.
	score += r.JoinCount * 8

	// Subquery depth: 15 points per level.
	score += r.SubqueryDepth * 15

	// CTEs: 10 points each.
	score += r.CTECount * 10

	// Unions: 12 points each.
	score += r.UnionCount * 12

	// Window functions: 10 points each.
	score += r.WindowFunctions * 10

	// Aggregates: 3 points each.
	score += r.AggregateCount * 3

	// Tables: 2 points each beyond the first.
	if r.TableCount > 1 {
		score += (r.TableCount - 1) * 2
	}

	// Conditions: 1 point each beyond 3.
	if r.ConditionCount > 3 {
		score += (r.ConditionCount - 3)
	}

	// Clause bonuses.
	if r.HasGroupBy {
		score += 5
	}
	if r.HasHaving {
		score += 5
	}
	if r.HasDistinct {
		score += 3
	}
	if !r.HasLimit && r.JoinCount > 0 {
		score += 5 // no LIMIT on a join query is risky
	}

	if score > 100 {
		score = 100
	}
	return score
}

func countSubqueryDepth(query string) int {
	maxDepth := 0
	currentDepth := 0
	upper := strings.ToUpper(query)
	inString := false
	prev := byte(0)

	for i := 0; i < len(query); i++ {
		ch := query[i]
		if ch == '\'' && prev != '\\' {
			inString = !inString
		}
		if inString {
			prev = ch
			continue
		}

		if ch == '(' {
			currentDepth++
			// Check if this paren contains a SELECT (subquery).
			remaining := strings.TrimSpace(upper[i+1:])
			if strings.HasPrefix(remaining, "SELECT") || strings.HasPrefix(remaining, "WITH") {
				if currentDepth > maxDepth {
					maxDepth = currentDepth
				}
			}
		}
		if ch == ')' {
			currentDepth--
			if currentDepth < 0 {
				currentDepth = 0
			}
		}
		prev = ch
	}

	return maxDepth
}
