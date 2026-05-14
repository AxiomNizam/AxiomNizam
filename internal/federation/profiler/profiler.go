package profiler

// =====================================================
// WS-5.3 — Query Profiler
//
// Profiles federated query execution: timing breakdown per
// sub-query, bottleneck detection, data transfer analysis,
// and optimization recommendations.
// =====================================================

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// QueryProfile holds the full performance profile of a federated query execution.
type QueryProfile struct {
	QueryID         string              `json:"queryId"`
	SQL             string              `json:"sql"`
	TotalDuration   time.Duration       `json:"totalDuration"`
	PlanningTime    time.Duration       `json:"planningTime"`
	ExecutionTime   time.Duration       `json:"executionTime"`
	MergingTime     time.Duration       `json:"mergingTime"`
	SubQueryProfiles []SubQueryProfile  `json:"subQueryProfiles"`
	DataTransfer    DataTransferProfile `json:"dataTransfer"`
	Bottleneck      *Bottleneck         `json:"bottleneck,omitempty"`
	Recommendations []Recommendation    `json:"recommendations"`
	ProfiledAt      time.Time           `json:"profiledAt"`
}

// SubQueryProfile captures timing for a single sub-query.
type SubQueryProfile struct {
	DataSource   string        `json:"dataSource"`
	SQL          string        `json:"sql"`
	Duration     time.Duration `json:"duration"`
	RowsReturned int64         `json:"rowsReturned"`
	BytesRead    int64         `json:"bytesRead"`
	WaitTime     time.Duration `json:"waitTime"` // Time waiting for semaphore
}

// DataTransferProfile summarizes cross-source data movement.
type DataTransferProfile struct {
	TotalBytesTransferred int64 `json:"totalBytesTransferred"`
	TotalRowsTransferred  int64 `json:"totalRowsTransferred"`
	SourceCount           int   `json:"sourceCount"`
	LargestSourceRows     int64 `json:"largestSourceRows"`
	LargestSourceName     string `json:"largestSourceName"`
}

// Bottleneck identifies the performance bottleneck in the query.
type Bottleneck struct {
	Type        string `json:"type"`        // slow_source, large_transfer, missing_pushdown, cross_join
	Source      string `json:"source"`
	Description string `json:"description"`
	Impact      string `json:"impact"` // low, medium, high, critical
}

// Recommendation suggests an optimization for the query.
type Recommendation struct {
	Type        string `json:"type"`        // index, materialized_view, pushdown, cache, partition
	Priority    string `json:"priority"`    // low, medium, high
	Description string `json:"description"`
	Savings     string `json:"savings,omitempty"` // Estimated improvement
}

// Profiler analyzes federated query execution performance.
type Profiler struct{}

// NewProfiler creates a new query profiler.
func NewProfiler() *Profiler {
	return &Profiler{}
}

// Profile analyzes sub-query profiles and generates a full query profile with recommendations.
func (p *Profiler) Profile(queryID, sql string, planningTime, executionTime, mergingTime time.Duration, subProfiles []SubQueryProfile) *QueryProfile {
	profile := &QueryProfile{
		QueryID:          queryID,
		SQL:              sql,
		TotalDuration:    planningTime + executionTime + mergingTime,
		PlanningTime:     planningTime,
		ExecutionTime:    executionTime,
		MergingTime:      mergingTime,
		SubQueryProfiles: subProfiles,
		ProfiledAt:       time.Now(),
	}

	// Compute data transfer profile.
	profile.DataTransfer = p.computeTransferProfile(subProfiles)

	// Detect bottleneck.
	profile.Bottleneck = p.detectBottleneck(profile)

	// Generate recommendations.
	profile.Recommendations = p.generateRecommendations(profile)

	return profile
}

// computeTransferProfile calculates data transfer statistics.
func (p *Profiler) computeTransferProfile(subProfiles []SubQueryProfile) DataTransferProfile {
	dtp := DataTransferProfile{
		SourceCount: len(subProfiles),
	}

	for _, sp := range subProfiles {
		dtp.TotalBytesTransferred += sp.BytesRead
		dtp.TotalRowsTransferred += sp.RowsReturned
		if sp.RowsReturned > dtp.LargestSourceRows {
			dtp.LargestSourceRows = sp.RowsReturned
			dtp.LargestSourceName = sp.DataSource
		}
	}

	return dtp
}

// detectBottleneck identifies the primary performance bottleneck.
func (p *Profiler) detectBottleneck(profile *QueryProfile) *Bottleneck {
	if len(profile.SubQueryProfiles) == 0 {
		return nil
	}

	// Sort by duration descending.
	sorted := make([]SubQueryProfile, len(profile.SubQueryProfiles))
	copy(sorted, profile.SubQueryProfiles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Duration > sorted[j].Duration
	})

	slowest := sorted[0]

	// Check if one source is dramatically slower than others.
	if len(sorted) > 1 && slowest.Duration > sorted[1].Duration*3 {
		return &Bottleneck{
			Type:        "slow_source",
			Source:      slowest.DataSource,
			Description: fmt.Sprintf("Source '%s' took %s — 3x+ slower than the next source (%s)", slowest.DataSource, slowest.Duration.Truncate(time.Millisecond), sorted[1].Duration.Truncate(time.Millisecond)),
			Impact:      "high",
		}
	}

	// Check if data transfer is very large.
	if profile.DataTransfer.TotalRowsTransferred > 1000000 {
		return &Bottleneck{
			Type:        "large_transfer",
			Source:      profile.DataTransfer.LargestSourceName,
			Description: fmt.Sprintf("Transferred %d rows across %d sources (largest: %s with %d rows)", profile.DataTransfer.TotalRowsTransferred, profile.DataTransfer.SourceCount, profile.DataTransfer.LargestSourceName, profile.DataTransfer.LargestSourceRows),
			Impact:      "high",
		}
	}

	// Check if merging dominates.
	if profile.MergingTime > profile.ExecutionTime {
		return &Bottleneck{
			Type:        "cross_join",
			Description: fmt.Sprintf("Merge time (%s) exceeds execution time (%s) — consider adding join predicates or pre-filtering", profile.MergingTime.Truncate(time.Millisecond), profile.ExecutionTime.Truncate(time.Millisecond)),
			Impact:      "medium",
		}
	}

	return nil
}

// generateRecommendations produces optimization suggestions.
func (p *Profiler) generateRecommendations(profile *QueryProfile) []Recommendation {
	var recs []Recommendation

	// Large row scan recommendation.
	for _, sp := range profile.SubQueryProfiles {
		if sp.RowsReturned > 100000 {
			recs = append(recs, Recommendation{
				Type:        "index",
				Priority:    "high",
				Description: fmt.Sprintf("Source '%s' returned %d rows — add an index on filter columns to reduce scan", sp.DataSource, sp.RowsReturned),
				Savings:     "50-80% scan reduction",
			})
		}
	}

	// Materialized view recommendation for multi-source queries.
	if len(profile.SubQueryProfiles) > 2 {
		recs = append(recs, Recommendation{
			Type:        "materialized_view",
			Priority:    "medium",
			Description: fmt.Sprintf("Query touches %d sources — consider creating a materialized view for repeated execution", len(profile.SubQueryProfiles)),
			Savings:     "90%+ latency reduction on subsequent runs",
		})
	}

	// Caching recommendation for fast queries.
	if profile.TotalDuration > 500*time.Millisecond {
		recs = append(recs, Recommendation{
			Type:        "cache",
			Priority:    "medium",
			Description: fmt.Sprintf("Query takes %s — enable result caching if data freshness allows", profile.TotalDuration.Truncate(time.Millisecond)),
			Savings:     "Near-zero latency on cache hit",
		})
	}

	// Missing WHERE clause.
	if !strings.Contains(strings.ToLower(profile.SQL), "where") {
		recs = append(recs, Recommendation{
			Type:        "pushdown",
			Priority:    "high",
			Description: "No WHERE clause detected — add filters to reduce data transfer from sources",
			Savings:     "Proportional to filter selectivity",
		})
	}

	return recs
}
