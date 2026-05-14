package federation

// =====================================================
// WS-5.1 — Federated Query Result Merger
//
// Merges result sets from multiple datasources. Supports:
//   - UNION: concatenate rows from multiple sources
//   - JOIN: hash-join on matching keys
//   - SORT: order merged results
//   - LIMIT: cap output row count
// =====================================================

import (
	"fmt"
	"sort"
	"strings"
)

// MergeStrategy defines how to combine result sets.
type MergeStrategy string

const (
	MergeUnion     MergeStrategy = "union"
	MergeInnerJoin MergeStrategy = "inner_join"
	MergeLeftJoin  MergeStrategy = "left_join"
	MergeCrossJoin MergeStrategy = "cross_join"
)

// MergeOptions controls merge behavior.
type MergeOptions struct {
	Strategy   MergeStrategy
	JoinKey    string // Column name to join on
	SortColumn string
	SortDesc   bool
	Limit      int64
}

// ResultMerger combines result sets from parallel sub-query execution.
type ResultMerger struct{}

// NewResultMerger creates a new merger.
func NewResultMerger() *ResultMerger {
	return &ResultMerger{}
}

// Merge combines multiple result sets according to the given options.
func (m *ResultMerger) Merge(results []SubQueryResult, opts MergeOptions) (*ResultSet, error) {
	if len(results) == 0 {
		return &ResultSet{}, nil
	}

	// Filter out errored results.
	var valid []SubQueryResult
	for _, r := range results {
		if r.Error == nil && r.Result != nil {
			valid = append(valid, r)
		}
	}

	if len(valid) == 0 {
		return nil, fmt.Errorf("merger: no valid result sets to merge")
	}

	switch opts.Strategy {
	case MergeUnion:
		return m.mergeUnion(valid, opts)
	case MergeInnerJoin:
		return m.mergeInnerJoin(valid, opts)
	case MergeLeftJoin:
		return m.mergeLeftJoin(valid, opts)
	default:
		return m.mergeUnion(valid, opts)
	}
}

// mergeUnion concatenates all rows from all result sets.
func (m *ResultMerger) mergeUnion(results []SubQueryResult, opts MergeOptions) (*ResultSet, error) {
	merged := &ResultSet{}

	// Use columns from the first result.
	if len(results) > 0 && results[0].Result != nil {
		merged.Columns = results[0].Result.Columns
	}

	for _, r := range results {
		merged.Rows = append(merged.Rows, r.Result.Rows...)
	}
	merged.RowCount = int64(len(merged.Rows))

	// Apply sort.
	if opts.SortColumn != "" {
		m.sortRows(merged.Rows, opts.SortColumn, opts.SortDesc)
	}

	// Apply limit.
	if opts.Limit > 0 && int64(len(merged.Rows)) > opts.Limit {
		merged.Rows = merged.Rows[:opts.Limit]
		merged.RowCount = opts.Limit
	}

	return merged, nil
}

// mergeInnerJoin performs a hash join on the specified key.
func (m *ResultMerger) mergeInnerJoin(results []SubQueryResult, opts MergeOptions) (*ResultSet, error) {
	if len(results) < 2 {
		return m.mergeUnion(results, opts)
	}
	if opts.JoinKey == "" {
		return nil, fmt.Errorf("merger: inner join requires a join key")
	}

	// Build hash index from the first result set.
	left := results[0].Result
	if left == nil {
		return nil, fmt.Errorf("merger: left result set is nil")
	}
	if results[1].Result == nil {
		return nil, fmt.Errorf("merger: right result set is nil")
	}
	index := make(map[string][]map[string]interface{})
	for _, row := range left.Rows {
		key := fmt.Sprintf("%v", row[opts.JoinKey])
		index[key] = append(index[key], row)
	}

	// Probe with subsequent result sets.
	merged := &ResultSet{}
	merged.Columns = mergeColumns(left.Columns, results[1].Result.Columns)

	for _, r := range results[1:] {
		for _, rightRow := range r.Result.Rows {
			key := fmt.Sprintf("%v", rightRow[opts.JoinKey])
			if leftRows, ok := index[key]; ok {
				for _, leftRow := range leftRows {
					combined := mergeRow(leftRow, rightRow)
					merged.Rows = append(merged.Rows, combined)
				}
			}
		}
	}

	merged.RowCount = int64(len(merged.Rows))

	if opts.SortColumn != "" {
		m.sortRows(merged.Rows, opts.SortColumn, opts.SortDesc)
	}
	if opts.Limit > 0 && int64(len(merged.Rows)) > opts.Limit {
		merged.Rows = merged.Rows[:opts.Limit]
		merged.RowCount = opts.Limit
	}

	return merged, nil
}

// mergeLeftJoin keeps all rows from the left and matches from the right.
func (m *ResultMerger) mergeLeftJoin(results []SubQueryResult, opts MergeOptions) (*ResultSet, error) {
	if len(results) < 2 {
		return m.mergeUnion(results, opts)
	}
	if opts.JoinKey == "" {
		return nil, fmt.Errorf("merger: left join requires a join key")
	}

	left := results[0].Result

	// Build hash index from right result sets.
	rightIndex := make(map[string][]map[string]interface{})
	for _, r := range results[1:] {
		for _, row := range r.Result.Rows {
			key := fmt.Sprintf("%v", row[opts.JoinKey])
			rightIndex[key] = append(rightIndex[key], row)
		}
	}

	merged := &ResultSet{}
	merged.Columns = mergeColumns(left.Columns, results[1].Result.Columns)

	for _, leftRow := range left.Rows {
		key := fmt.Sprintf("%v", leftRow[opts.JoinKey])
		if rightRows, ok := rightIndex[key]; ok {
			for _, rightRow := range rightRows {
				merged.Rows = append(merged.Rows, mergeRow(leftRow, rightRow))
			}
		} else {
			// No match — include left row with nulls for right columns.
			merged.Rows = append(merged.Rows, leftRow)
		}
	}

	merged.RowCount = int64(len(merged.Rows))

	if opts.Limit > 0 && int64(len(merged.Rows)) > opts.Limit {
		merged.Rows = merged.Rows[:opts.Limit]
		merged.RowCount = opts.Limit
	}

	return merged, nil
}

// sortRows sorts rows by a column.
func (m *ResultMerger) sortRows(rows []map[string]interface{}, column string, desc bool) {
	sort.SliceStable(rows, func(i, j int) bool {
		vi := fmt.Sprintf("%v", rows[i][column])
		vj := fmt.Sprintf("%v", rows[j][column])
		if desc {
			return strings.Compare(vi, vj) > 0
		}
		return strings.Compare(vi, vj) < 0
	})
}

// mergeColumns combines column lists, deduplicating by name.
func mergeColumns(left, right []ColumnMeta) []ColumnMeta {
	seen := make(map[string]bool)
	var merged []ColumnMeta
	for _, c := range left {
		merged = append(merged, c)
		seen[c.Name] = true
	}
	for _, c := range right {
		if !seen[c.Name] {
			merged = append(merged, c)
		}
	}
	return merged
}

// mergeRow combines two rows into one, with right values overriding on conflict.
func mergeRow(left, right map[string]interface{}) map[string]interface{} {
	combined := make(map[string]interface{}, len(left)+len(right))
	for k, v := range left {
		combined[k] = v
	}
	for k, v := range right {
		if _, exists := combined[k]; !exists {
			combined[k] = v
		}
	}
	return combined
}
