package tests

import (
	"encoding/json"
	"testing"

	"example.com/axiomnizam/internal/docs"
	"example.com/axiomnizam/internal/performance"
	"example.com/axiomnizam/internal/ratelimit"
)

// TestQuotaManager tests quota functionality
func TestQuotaManager(t *testing.T) {
	qm := ratelimit.NewQuotaManager()

	// Test: Set daily quota
	qm.SetUserDailyQuota("user1", 1000000)

	// Test: Check quota
	allowed, remaining, err := qm.CheckQuota("user1", "/api/users", 100000)
	if !allowed || err != nil {
		t.Fatalf("Expected quota to be allowed, got: %v", err)
	}

	// Test: Get quota status
	status := qm.GetQuotaStatus("user1")
	if status["daily_limit"] != int64(1000000) {
		t.Fatalf("Expected daily_limit 1000000, got: %v", status["daily_limit"])
	}

	// Test: Exceed quota
	for i := 0; i < 10; i++ {
		qm.CheckQuota("user1", "/api/users", 100000)
	}

	allowed, _, err = qm.CheckQuota("user1", "/api/users", 100000)
	if allowed && err == nil {
		t.Fatalf("Expected quota to be exceeded")
	}

	t.Log("✓ QuotaManager tests passed")
}

// TestQueryPerformanceAnalyzer tests performance analysis
func TestQueryPerformanceAnalyzer(t *testing.T) {
	analyzer := performance.NewQueryPerformanceAnalyzer(100, 10000)

	// Test: Record queries
	for i := 0; i < 100; i++ {
		analyzer.RecordQuery(performance.QueryPerformance{
			Query:        "SELECT * FROM users",
			Duration:     int64(i * 10),
			RowsScanned:  int64(i * 1000),
			RowsReturned: int64(i * 100),
			ExecutedBy:   "user1",
			Status:       "success",
			QueryType:    "SELECT",
			IndexUsed:    i%2 == 0,
		})
	}

	// Test: Get stats
	stats := analyzer.GetQueryStats()
	if stats["total_queries"] != 100 {
		t.Fatalf("Expected 100 queries, got: %v", stats["total_queries"])
	}

	// Test: Get slow queries
	slow := analyzer.GetSlowQueries()
	if len(slow) == 0 {
		t.Fatalf("Expected some slow queries")
	}

	// Test: Get percentiles
	p99 := analyzer.GetPercentile(99)
	if p99 <= 0 {
		t.Fatalf("Expected p99 > 0, got: %v", p99)
	}

	// Test: Get recommendations
	recommendations := analyzer.GetRecommendations()
	if len(recommendations) == 0 {
		t.Log("✓ No recommendations (queries well optimized)")
	}

	t.Log("✓ QueryPerformanceAnalyzer tests passed")
}

// TestOpenAPIGenerator tests API documentation generation
func TestOpenAPIGenerator(t *testing.T) {
	generator := docs.NewOpenAPIGenerator(docs.OpenAPIInfo{
		Title:       "Test API",
		Version:     "1.0.0",
		Description: "Test Description",
	})

	// Test: Add endpoint
	generator.AddEndpoint(docs.OpenAPIEndpoint{
		Path:        "/api/users",
		Method:      "GET",
		Summary:     "Get all users",
		Description: "Retrieve list of all users",
		Tags:        []string{"users"},
		Parameters: []docs.OpenAPIParameter{
			{
				Name:        "limit",
				In:          "query",
				Required:    false,
				Type:        "integer",
				Description: "Max number of results",
			},
		},
		Response: docs.OpenAPIResponse{
			Status:      200,
			Description: "Success",
		},
	})

	// Test: Build OpenAPI spec
	spec := generator.BuildOpenAPI()
	if spec["openapi"] != "3.0.0" {
		t.Fatalf("Expected OpenAPI 3.0.0")
	}

	// Test: Get markdown
	markdown := generator.GetEndpointMarkdown()
	if !contains(markdown, "/api/users") {
		t.Fatalf("Expected markdown to contain endpoint")
	}

	t.Log("✓ OpenAPIGenerator tests passed")
}

// TestRateLimitMiddleware tests middleware functionality
func TestRateLimitMiddleware(t *testing.T) {
	qm := ratelimit.NewQuotaManager()
	middleware := ratelimit.NewRateLimitMiddleware(qm)

	// Middleware should be created successfully
	if middleware == nil {
		t.Fatalf("Failed to create middleware")
	}

	// Handler should be callable
	handler := middleware.Handler()
	if handler == nil {
		t.Fatalf("Handler is nil")
	}

	t.Log("✓ RateLimitMiddleware tests passed")
}

// TestPercentileCalculation tests percentile calculation
func TestPercentileCalculation(t *testing.T) {
	analyzer := performance.NewQueryPerformanceAnalyzer(100, 10000)

	// Record 1000 queries with varying durations
	for i := 0; i < 1000; i++ {
		analyzer.RecordQuery(performance.QueryPerformance{
			Query:        "SELECT * FROM data",
			Duration:     int64((i % 200) * 5), // 0-995ms
			RowsScanned:  100,
			RowsReturned: 50,
			ExecutedBy:   "user1",
			Status:       "success",
			QueryType:    "SELECT",
		})
	}

	// Test percentiles
	p50 := analyzer.GetPercentile(50)
	p95 := analyzer.GetPercentile(95)
	p99 := analyzer.GetPercentile(99)

	if p50 <= 0 || p95 <= p50 || p99 <= p95 {
		t.Fatalf("Invalid percentiles: p50=%v, p95=%v, p99=%v", p50, p95, p99)
	}

	t.Logf("✓ Percentiles: P50=%vms, P95=%vms, P99=%vms", p50, p95, p99)
}

// TestQuotaReset tests quota reset functionality
func TestQuotaReset(t *testing.T) {
	qm := ratelimit.NewQuotaManager()

	// Set and use quota
	qm.SetUserDailyQuota("user1", 1000000)
	qm.CheckQuota("user1", "/api/users", 500000)

	// Check usage before reset
	status1 := qm.GetQuotaStatus("user1")
	dailyUsed1 := status1["daily_used"]

	// Reset
	qm.ResetUserQuota("user1")

	// Check usage after reset
	status2 := qm.GetQuotaStatus("user1")
	if status2["total_requests"] != 0 {
		t.Fatalf("Expected total_requests to be 0 after reset")
	}

	t.Log("✓ QuotaReset tests passed")
}

// TestUserStatistics tests user statistics tracking
func TestUserStatistics(t *testing.T) {
	analyzer := performance.NewQueryPerformanceAnalyzer(100, 10000)

	// Record queries from different users
	for user := 1; user <= 5; user++ {
		for i := 0; i < 20; i++ {
			analyzer.RecordQuery(performance.QueryPerformance{
				Query:        "SELECT * FROM users",
				Duration:     int64(i * 10),
				RowsScanned:  100,
				RowsReturned: 50,
				ExecutedBy:   "user" + string(rune('0'+user)),
				Status:       "success",
				QueryType:    "SELECT",
			})
		}
	}

	// Get user stats
	userStats := analyzer.GetUserStats()
	if len(userStats) != 5 {
		t.Fatalf("Expected 5 users, got: %v", len(userStats))
	}

	t.Log("✓ UserStatistics tests passed")
}

// TestCacheHitTracking tests cache hit tracking
func TestCacheHitTracking(t *testing.T) {
	analyzer := performance.NewQueryPerformanceAnalyzer(100, 10000)

	// Record with and without cache hits
	for i := 0; i < 100; i++ {
		analyzer.RecordQuery(performance.QueryPerformance{
			Query:        "SELECT * FROM users",
			Duration:     int64(i * 5),
			RowsScanned:  100,
			RowsReturned: 50,
			ExecutedBy:   "user1",
			Status:       "success",
			QueryType:    "SELECT",
			CacheHit:     i%2 == 0, // 50% cache hit rate
		})
	}

	// Check stats
	stats := analyzer.GetQueryStats()
	cacheHitRate := stats["cache_hit_rate"].(float64)

	if cacheHitRate < 0.45 || cacheHitRate > 0.55 {
		t.Fatalf("Expected ~50%% cache hit rate, got: %v%%", cacheHitRate*100)
	}

	t.Log("✓ CacheHitTracking tests passed")
}

// BenchmarkQuotaCheck benchmarks quota checking
func BenchmarkQuotaCheck(b *testing.B) {
	qm := ratelimit.NewQuotaManager()
	qm.SetUserDailyQuota("bench_user", 10000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qm.CheckQuota("bench_user", "/api/users", 1000)
	}
}

// BenchmarkQueryRecording benchmarks query recording
func BenchmarkQueryRecording(b *testing.B) {
	analyzer := performance.NewQueryPerformanceAnalyzer(100, 100000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.RecordQuery(performance.QueryPerformance{
			Query:        "SELECT * FROM users",
			Duration:     int64(i % 1000),
			RowsScanned:  int64(i % 10000),
			RowsReturned: int64(i % 1000),
			ExecutedBy:   "bench_user",
			Status:       "success",
			QueryType:    "SELECT",
		})
	}
}

// TestSpecMarshal tests OpenAPI spec marshaling to JSON
func TestSpecMarshal(t *testing.T) {
	generator := docs.NewOpenAPIGenerator(docs.OpenAPIInfo{
		Title:       "Test API",
		Version:     "1.0.0",
		Description: "Test",
	})

	spec := generator.BuildOpenAPI()

	// Should be JSON marshallable
	jsonData, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal spec to JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Fatalf("Empty JSON output")
	}

	t.Log("✓ SpecMarshal tests passed")
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RunAllTests runs all tests
func RunAllTests(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"QuotaManager", TestQuotaManager},
		{"QueryPerformanceAnalyzer", TestQueryPerformanceAnalyzer},
		{"OpenAPIGenerator", TestOpenAPIGenerator},
		{"RateLimitMiddleware", TestRateLimitMiddleware},
		{"PercentileCalculation", TestPercentileCalculation},
		{"QuotaReset", TestQuotaReset},
		{"UserStatistics", TestUserStatistics},
		{"CacheHitTracking", TestCacheHitTracking},
		{"SpecMarshal", TestSpecMarshal},
	}

	for _, test := range tests {
		t.Run(test.name, test.fn)
	}

	t.Log("\n✓ All Phase 1 tests passed!")
}
