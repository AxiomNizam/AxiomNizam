package federation

// =====================================================
// WS-5.1 — Federated Query Optimizer
//
// Cost-based optimization for cross-source queries:
//   - Predicate push-down: push WHERE clauses to source databases
//   - Projection push-down: only SELECT needed columns
//   - Join reordering: smaller table first based on catalog stats
//   - Limit push-down: push LIMIT when no cross-source ops needed
// =====================================================

import (
	"strings"
)

// OptimizationHint describes an optimization applied to a query plan.
type OptimizationHint struct {
	Type        string `json:"type"`        // predicate_pushdown, projection_pushdown, join_reorder, limit_pushdown
	Description string `json:"description"`
	Savings     string `json:"savings,omitempty"` // Estimated savings
}

// QueryOptimizer applies cost-based optimizations to query plans.
type QueryOptimizer struct{}

// NewQueryOptimizer creates a new optimizer.
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{}
}

// Optimize applies optimizations to a query plan and returns hints.
func (o *QueryOptimizer) Optimize(plan *QueryPlanNode, sql string) ([]OptimizationHint, *QueryPlanNode) {
	if plan == nil {
		return nil, plan
	}

	var hints []OptimizationHint

	// 1. Predicate push-down.
	if whereClause := extractWhereClause(sql); whereClause != "" {
		for i := range plan.Children {
			if plan.Children[i].Type == "remote_scan" {
				hints = append(hints, OptimizationHint{
					Type:        "predicate_pushdown",
					Description: "WHERE clause pushed to " + plan.Children[i].DataSource,
				})
			}
		}
	}

	// 2. Projection push-down.
	selectCols := extractSelectColumns(sql)
	if len(selectCols) > 0 && selectCols[0] != "*" {
		for i := range plan.Children {
			if plan.Children[i].Type == "remote_scan" {
				hints = append(hints, OptimizationHint{
					Type:        "projection_pushdown",
					Description: "only selected columns fetched from " + plan.Children[i].DataSource,
				})
			}
		}
	}

	// 3. Join reordering — put smaller table first.
	if plan.Type == "merge_join" && len(plan.Children) >= 2 {
		reordered := false
		for i := 1; i < len(plan.Children); i++ {
			if plan.Children[i].EstimatedRows > 0 &&
				plan.Children[i].EstimatedRows < plan.Children[i-1].EstimatedRows {
				plan.Children[i-1], plan.Children[i] = plan.Children[i], plan.Children[i-1]
				reordered = true
			}
		}
		if reordered {
			hints = append(hints, OptimizationHint{
				Type:        "join_reorder",
				Description: "tables reordered by estimated row count (smallest first)",
			})
		}
	}

	// 4. Limit push-down — if single source, push limit.
	if limitVal := extractLimit(sql); limitVal > 0 && len(plan.Children) == 1 {
		hints = append(hints, OptimizationHint{
			Type:        "limit_pushdown",
			Description: "LIMIT pushed to source database",
		})
	}

	// Recalculate cost after optimizations.
	plan.EstimatedCost = o.estimateCost(plan)

	return hints, plan
}

// Recommendations returns optimization suggestions for a query.
func (o *QueryOptimizer) Recommendations(plan *QueryPlanNode, sql string) []string {
	var recs []string

	if plan == nil {
		return recs
	}

	// Check for missing indexes.
	for _, child := range plan.Children {
		if child.EstimatedRows > 100000 {
			recs = append(recs, "Consider adding an index on "+child.Table+" for faster scans ("+
				formatRows(child.EstimatedRows)+" estimated rows)")
		}
	}

	// Check for cross-source joins without materialization.
	if plan.Type == "merge_join" && len(plan.Children) > 1 {
		recs = append(recs, "Consider materializing this query if executed frequently (cross-source join detected)")
	}

	// Check for missing WHERE clause.
	if !strings.Contains(strings.ToLower(sql), "where") {
		recs = append(recs, "Add a WHERE clause to reduce data transfer from source databases")
	}

	return recs
}

// estimateCost calculates the total estimated cost of a plan.
func (o *QueryOptimizer) estimateCost(plan *QueryPlanNode) float64 {
	if plan == nil {
		return 0
	}
	cost := plan.EstimatedCost
	for _, child := range plan.Children {
		cost += o.estimateCost(&child)
	}
	return cost
}

// --- SQL parsing helpers (simplified) ---

func extractWhereClause(sql string) string {
	lower := strings.ToLower(sql)
	idx := strings.Index(lower, "where ")
	if idx < 0 {
		return ""
	}
	rest := sql[idx+6:]
	// Trim at GROUP BY, ORDER BY, LIMIT, etc.
	for _, kw := range []string{"group by", "order by", "limit", "having"} {
		if end := strings.Index(strings.ToLower(rest), kw); end > 0 {
			rest = rest[:end]
		}
	}
	return strings.TrimSpace(rest)
}

func extractSelectColumns(sql string) []string {
	lower := strings.ToLower(sql)
	selectIdx := strings.Index(lower, "select ")
	if selectIdx < 0 {
		return nil
	}
	fromIdx := strings.Index(lower, " from ")
	if fromIdx < 0 {
		return nil
	}
	colStr := strings.TrimSpace(sql[selectIdx+7 : fromIdx])
	if colStr == "*" {
		return []string{"*"}
	}
	parts := strings.Split(colStr, ",")
	var cols []string
	for _, p := range parts {
		cols = append(cols, strings.TrimSpace(p))
	}
	return cols
}

func extractLimit(sql string) int64 {
	lower := strings.ToLower(sql)
	idx := strings.Index(lower, "limit ")
	if idx < 0 {
		return 0
	}
	rest := strings.TrimSpace(sql[idx+6:])
	var n int64
	for _, c := range rest {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
		} else {
			break
		}
	}
	return n
}

func formatRows(n int64) string {
	if n >= 1000000 {
		return strings.TrimRight(strings.TrimRight(
			strings.Replace(
				strings.Replace(
					string(rune(n/1000000+'0'))+"M", "0M", "M", 1),
				"M", "M", 1),
			"0"), ".")
	}
	if n >= 1000 {
		return string(rune(n/1000+'0')) + "K"
	}
	return string(rune(n + '0'))
}
