package performance

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// QueryPerformance tracks individual query performance
type QueryPerformance struct {
	Query        string
	Duration     int64 // milliseconds
	RowsScanned  int64
	RowsReturned int64
	Timestamp    time.Time
	ExecutedBy   string
	Database     string
	Status       string // success, error, timeout
	IndexUsed    bool
	CacheHit     bool
	QueryType    string // SELECT, INSERT, UPDATE, DELETE, etc
	Indexed      bool
}

// QueryPerformanceAnalyzer analyzes query performance
type QueryPerformanceAnalyzer struct {
	queries       []QueryPerformance
	mu            sync.RWMutex
	slowThreshold int64 // milliseconds
	maxSize       int
}

// NewQueryPerformanceAnalyzer creates a new analyzer
func NewQueryPerformanceAnalyzer(slowThreshold int64, maxSize int) *QueryPerformanceAnalyzer {
	return &QueryPerformanceAnalyzer{
		queries:       make([]QueryPerformance, 0),
		slowThreshold: slowThreshold,
		maxSize:       maxSize,
	}
}

// RecordQuery records a query execution
func (qpa *QueryPerformanceAnalyzer) RecordQuery(qp QueryPerformance) {
	qpa.mu.Lock()
	defer qpa.mu.Unlock()

	qp.Timestamp = time.Now()
	qpa.queries = append(qpa.queries, qp)

	// Keep only recent queries
	if len(qpa.queries) > qpa.maxSize {
		qpa.queries = qpa.queries[len(qpa.queries)-qpa.maxSize:]
	}
}

// GetSlowQueries returns queries slower than threshold
func (qpa *QueryPerformanceAnalyzer) GetSlowQueries() []QueryPerformance {
	qpa.mu.RLock()
	defer qpa.mu.RUnlock()

	slow := make([]QueryPerformance, 0)
	for _, q := range qpa.queries {
		if q.Duration >= qpa.slowThreshold {
			slow = append(slow, q)
		}
	}

	sort.Slice(slow, func(i, j int) bool {
		return slow[i].Duration > slow[j].Duration
	})

	return slow
}

// GetQueryStats returns statistics for all queries
func (qpa *QueryPerformanceAnalyzer) GetQueryStats() map[string]interface{} {
	qpa.mu.RLock()
	defer qpa.mu.RUnlock()

	if len(qpa.queries) == 0 {
		return map[string]interface{}{
			"total_queries": 0,
			"avg_duration":  0,
			"min_duration":  0,
			"max_duration":  0,
		}
	}

	totalDuration := int64(0)
	minDuration := qpa.queries[0].Duration
	maxDuration := qpa.queries[0].Duration
	successCount := 0
	errorCount := 0
	cacheHits := 0

	for _, q := range qpa.queries {
		totalDuration += q.Duration
		if q.Duration < minDuration {
			minDuration = q.Duration
		}
		if q.Duration > maxDuration {
			maxDuration = q.Duration
		}
		if q.Status == "success" {
			successCount++
		} else if q.Status == "error" {
			errorCount++
		}
		if q.CacheHit {
			cacheHits++
		}
	}

	avgDuration := totalDuration / int64(len(qpa.queries))

	return map[string]interface{}{
		"total_queries":   len(qpa.queries),
		"avg_duration_ms": avgDuration,
		"min_duration_ms": minDuration,
		"max_duration_ms": maxDuration,
		"total_duration":  totalDuration,
		"success_count":   successCount,
		"error_count":     errorCount,
		"error_rate":      float64(errorCount) / float64(len(qpa.queries)),
		"cache_hits":      cacheHits,
		"cache_hit_rate":  float64(cacheHits) / float64(len(qpa.queries)),
	}
}

// GetQueryTypeStats returns stats grouped by query type
func (qpa *QueryPerformanceAnalyzer) GetQueryTypeStats() map[string]interface{} {
	qpa.mu.RLock()
	defer qpa.mu.RUnlock()

	typeStats := make(map[string]map[string]interface{})

	for _, q := range qpa.queries {
		if _, exists := typeStats[q.QueryType]; !exists {
			typeStats[q.QueryType] = map[string]interface{}{
				"count":        0,
				"total_ms":     int64(0),
				"avg_ms":       int64(0),
				"max_ms":       int64(0),
				"rows_scanned": int64(0),
			}
		}

		stats := typeStats[q.QueryType]
		stats["count"] = stats["count"].(int) + 1
		stats["total_ms"] = stats["total_ms"].(int64) + q.Duration
		if q.Duration > stats["max_ms"].(int64) {
			stats["max_ms"] = q.Duration
		}
		stats["rows_scanned"] = stats["rows_scanned"].(int64) + q.RowsScanned
	}

	// Calculate averages
	for queryType, stats := range typeStats {
		count := stats["count"].(int)
		total := stats["total_ms"].(int64)
		stats["avg_ms"] = total / int64(count)
		typeStats[queryType] = stats
	}

	// Convert to map[string]interface{}
	result := make(map[string]interface{})
	for k, v := range typeStats {
		result[k] = v
	}
	return result
}

// GetUserStats returns stats by user
func (qpa *QueryPerformanceAnalyzer) GetUserStats() map[string]interface{} {
	qpa.mu.RLock()
	defer qpa.mu.RUnlock()

	userStats := make(map[string]map[string]interface{})

	for _, q := range qpa.queries {
		if _, exists := userStats[q.ExecutedBy]; !exists {
			userStats[q.ExecutedBy] = map[string]interface{}{
				"queries":  0,
				"total_ms": int64(0),
				"avg_ms":   int64(0),
				"errors":   0,
				"last_run": q.Timestamp,
			}
		}

		stats := userStats[q.ExecutedBy]
		stats["queries"] = stats["queries"].(int) + 1
		stats["total_ms"] = stats["total_ms"].(int64) + q.Duration
		if q.Status == "error" {
			stats["errors"] = stats["errors"].(int) + 1
		}
		stats["last_run"] = q.Timestamp
		userStats[q.ExecutedBy] = stats
	}

	// Calculate averages
	for user, stats := range userStats {
		count := stats["queries"].(int)
		total := stats["total_ms"].(int64)
		stats["avg_ms"] = total / int64(count)
		userStats[user] = stats
	}

	// Convert to map[string]interface{}
	result := make(map[string]interface{})
	for k, v := range userStats {
		result[k] = v
	}
	return result
}

// GetRecommendations returns optimization recommendations
func (qpa *QueryPerformanceAnalyzer) GetRecommendations() []map[string]interface{} {
	qpa.mu.RLock()
	defer qpa.mu.RUnlock()

	recommendations := make([]map[string]interface{}, 0)

	// Check for unindexed queries
	unindexedCount := 0
	for _, q := range qpa.queries {
		if !q.IndexUsed && q.RowsScanned > 1000 {
			unindexedCount++
		}
	}
	if unindexedCount > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "missing_index",
			"count":       unindexedCount,
			"description": fmt.Sprintf("Found %d queries scanning >1000 rows without index", unindexedCount),
			"priority":    "high",
		})
	}

	// Check for slow queries
	slowCount := 0
	for _, q := range qpa.queries {
		if q.Duration >= qpa.slowThreshold {
			slowCount++
		}
	}
	if slowCount > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "slow_queries",
			"count":       slowCount,
			"description": fmt.Sprintf("Found %d queries slower than %dms", slowCount, qpa.slowThreshold),
			"priority":    "medium",
		})
	}

	// Check for low cache hit rate
	cacheHitRate := qpa.calculateCacheHitRate()
	if cacheHitRate < 0.5 && len(qpa.queries) > 10 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "low_cache_hit",
			"rate":        cacheHitRate,
			"description": "Cache hit rate is below 50%, consider caching more queries",
			"priority":    "low",
		})
	}

	return recommendations
}

// calculateCacheHitRate calculates overall cache hit rate
func (qpa *QueryPerformanceAnalyzer) calculateCacheHitRate() float64 {
	if len(qpa.queries) == 0 {
		return 0
	}

	hits := 0
	for _, q := range qpa.queries {
		if q.CacheHit {
			hits++
		}
	}

	return float64(hits) / float64(len(qpa.queries))
}

// GetPercentile returns query duration at given percentile
func (qpa *QueryPerformanceAnalyzer) GetPercentile(percentile float64) int64 {
	qpa.mu.RLock()
	defer qpa.mu.RUnlock()

	if len(qpa.queries) == 0 {
		return 0
	}

	durations := make([]int64, len(qpa.queries))
	for i, q := range qpa.queries {
		durations[i] = q.Duration
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	index := int(float64(len(durations)) * percentile / 100)
	if index >= len(durations) {
		index = len(durations) - 1
	}

	return durations[index]
}

// ClearOldQueries removes queries older than given duration
func (qpa *QueryPerformanceAnalyzer) ClearOldQueries(duration time.Duration) {
	qpa.mu.Lock()
	defer qpa.mu.Unlock()

	cutoffTime := time.Now().Add(-duration)
	newQueries := make([]QueryPerformance, 0)

	for _, q := range qpa.queries {
		if q.Timestamp.After(cutoffTime) {
			newQueries = append(newQueries, q)
		}
	}

	qpa.queries = newQueries
}
