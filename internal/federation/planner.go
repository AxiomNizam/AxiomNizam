package federation

// =====================================================
// WS-5.1 — Federated Query Planner
//
// Decomposes a cross-source SQL query into per-datasource sub-queries.
// Identifies table references, resolves them to datasources via the
// catalog, and builds an execution plan with push-down optimizations.
// =====================================================

import (
	"context"
	"fmt"
	"strings"
)

// CatalogResolver resolves table references to datasource connections.
type CatalogResolver interface {
	// ResolveTable maps a table reference (e.g. "postgres.customers") to a datasource.
	ResolveTable(ctx context.Context, tableRef string) (*ResolvedTable, error)
}

// ResolvedTable holds the datasource details for a table reference.
type ResolvedTable struct {
	DataSourceRef string `json:"dataSourceRef"`
	Database      string `json:"database"`
	Schema        string `json:"schema"`
	Table         string `json:"table"`
	EstimatedRows int64  `json:"estimatedRows"`
}

// SubQuery represents a query to be executed against a single datasource.
type SubQuery struct {
	DataSourceRef string `json:"dataSourceRef"`
	SQL           string `json:"sql"`
	Alias         string `json:"alias"`
}

// QueryPlanner decomposes federated queries into execution plans.
type QueryPlanner struct {
	resolver CatalogResolver
}

// NewQueryPlanner creates a new planner.
func NewQueryPlanner(resolver CatalogResolver) *QueryPlanner {
	return &QueryPlanner{resolver: resolver}
}

// Plan analyzes a SQL query and produces an execution plan.
// Returns the plan tree, list of datasources involved, and any error.
func (p *QueryPlanner) Plan(ctx context.Context, sql string) (*QueryPlanNode, []string, error) {
	if sql == "" {
		return nil, nil, fmt.Errorf("empty SQL query")
	}

	// Extract table references from the SQL.
	tableRefs := extractTableReferences(sql)
	if len(tableRefs) == 0 {
		return nil, nil, fmt.Errorf("no table references found in query")
	}

	// Resolve each table to a datasource.
	var sources []string
	sourceSet := make(map[string]bool)
	var scanNodes []QueryPlanNode

	for _, ref := range tableRefs {
		if p.resolver != nil {
			resolved, err := p.resolver.ResolveTable(ctx, ref)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot resolve table %q: %w", ref, err)
			}

			if !sourceSet[resolved.DataSourceRef] {
				sources = append(sources, resolved.DataSourceRef)
				sourceSet[resolved.DataSourceRef] = true
			}

			scanNodes = append(scanNodes, QueryPlanNode{
				Type:          "remote_scan",
				DataSource:    resolved.DataSourceRef,
				Table:         resolved.Table,
				EstimatedRows: resolved.EstimatedRows,
				EstimatedCost: float64(resolved.EstimatedRows) * 0.001,
			})
		} else {
			// No resolver — use the reference as-is.
			parts := strings.SplitN(ref, ".", 2)
			dsRef := parts[0]
			tableName := ref
			if len(parts) > 1 {
				tableName = parts[1]
			}

			if !sourceSet[dsRef] {
				sources = append(sources, dsRef)
				sourceSet[dsRef] = true
			}

			scanNodes = append(scanNodes, QueryPlanNode{
				Type:       "remote_scan",
				DataSource: dsRef,
				Table:      tableName,
			})
		}
	}

	// Build the plan tree.
	var root QueryPlanNode
	if len(scanNodes) == 1 {
		// Single source — just a remote scan with projection.
		root = QueryPlanNode{
			Type:     "project",
			Children: scanNodes,
		}
	} else {
		// Multiple sources — merge join.
		root = QueryPlanNode{
			Type:          "merge_join",
			EstimatedCost: estimateJoinCost(scanNodes),
			Children:      scanNodes,
		}
	}

	return &root, sources, nil
}

// extractTableReferences parses table references from SQL.
// Looks for patterns like "FROM table", "JOIN table", "datasource.table".
// This is a simplified parser — production would use a proper SQL AST.
func extractTableReferences(sql string) []string {
	normalized := strings.ToLower(sql)
	words := strings.Fields(normalized)

	var refs []string
	seen := make(map[string]bool)

	for i, word := range words {
		if (word == "from" || word == "join") && i+1 < len(words) {
			ref := cleanTableRef(words[i+1])
			if ref != "" && !seen[ref] && !isSQLKeyword(ref) {
				refs = append(refs, ref)
				seen[ref] = true
			}
		}
	}

	return refs
}

// cleanTableRef removes SQL noise from a table reference.
func cleanTableRef(ref string) string {
	ref = strings.TrimRight(ref, ",;()")
	ref = strings.TrimLeft(ref, "(")
	if ref == "" || ref == "(" || ref == ")" {
		return ""
	}
	return ref
}

// isSQLKeyword checks if a word is a SQL keyword (not a table name).
func isSQLKeyword(word string) bool {
	keywords := map[string]bool{
		"select": true, "where": true, "and": true, "or": true,
		"on": true, "as": true, "set": true, "values": true,
		"into": true, "group": true, "order": true, "having": true,
		"limit": true, "offset": true, "union": true, "except": true,
		"intersect": true, "case": true, "when": true, "then": true,
		"else": true, "end": true, "null": true, "not": true,
		"in": true, "between": true, "like": true, "is": true,
		"exists": true, "all": true, "any": true, "inner": true,
		"left": true, "right": true, "outer": true, "cross": true,
		"full": true, "natural": true, "lateral": true,
	}
	return keywords[word]
}

// estimateJoinCost estimates the cost of joining multiple scan nodes.
func estimateJoinCost(nodes []QueryPlanNode) float64 {
	var totalCost float64
	for _, n := range nodes {
		totalCost += n.EstimatedCost
		if n.EstimatedRows > 0 {
			totalCost += float64(n.EstimatedRows) * 0.0001
		}
	}
	return totalCost
}
