package integration

import (
	"fmt"
)

// GraphQLExample shows how to use GraphQL
func GraphQLExample() {
	fmt.Println(`
=== GraphQL EXAMPLES ===

1. Query Users:
{
  users(limit: 10) {
    id
    name
    email
  }
}

2. Query with Filter:
{
  users(where: "age > 18") {
    id
    name
    age
  }
}

3. Get User by ID:
{
  usersById(id: "123") {
    id
    name
    email
    age
  }
}

API Endpoints:
- POST /api/graphql - Execute GraphQL query
- GET /api/graphql/schema - Get schema
- GET /api/graphql/playground - Interactive playground
`)
}

// RateLimitExample shows how to use rate limiting
func RateLimitExample() {
	fmt.Println(`
=== RATE LIMITING & QUOTAS EXAMPLES ===

1. Check User Quota:
GET /api/v1/quota/user123
Response:
{
  "user_id": "user123",
  "daily_limit": 1000000,
  "daily_used": 250000,
  "daily_remaining": 750000,
  "bytes_consumed": 512000,
  "concurrent_requests": 5
}

2. Set User Daily Quota:
PUT /api/v1/quota/user123
{
  "daily_limit": 5000000
}

3. List All User Quotas:
GET /api/v1/quotas
Response:
{
  "user123": {...},
  "user456": {...}
}

4. Set Endpoint Rate Limit:
POST /api/v1/endpoints/api/users/limit
{
  "requests_per_second": 100,
  "requests_per_minute": 6000,
  "requests_per_hour": 360000,
  "bytes_per_minute": 10485760,
  "max_concurrent": 100
}

5. Reset User Quota:
POST /api/v1/quota/user123/reset

Rate Limit Response Headers:
- X-RateLimit-Limit: 10000
- X-RateLimit-Remaining: 9500
- X-RateLimit-Reset: 1234567890
`)
}

// APIDocumentationExample shows how to use API docs
func APIDocumentationExample() {
	fmt.Println(`
=== API DOCUMENTATION EXAMPLES ===

1. Get OpenAPI Spec (JSON):
GET /api/docs/openapi.json
Returns complete OpenAPI 3.0 specification

2. Swagger UI:
GET /api/docs/swagger
Interactive API explorer with test functionality

3. ReDoc UI:
GET /api/docs/redoc
Beautiful API documentation

4. Markdown Documentation:
GET /api/docs/markdown
Plain text markdown documentation

5. List All Endpoints:
GET /api/docs/endpoints
{
  "endpoints": [
    {
      "path": "/api/users",
      "method": "GET",
      "summary": "Get all users",
      "description": "Retrieve list of all users"
    }
  ],
  "total": 45
}

6. Get Endpoint Details:
GET /api/docs/endpoints/0

Integration:
- Auto-generates from code
- Updates on deployment
- Supports authentication docs
- Shows usage examples
`)
}

// PerformanceExample shows how to use performance monitoring
func PerformanceExample() {
	fmt.Println(`
=== QUERY PERFORMANCE MONITORING EXAMPLES ===

1. Get Overall Query Stats:
GET /api/v1/performance/stats
{
  "total_queries": 15420,
  "avg_duration_ms": 45,
  "min_duration_ms": 2,
  "max_duration_ms": 8932,
  "success_count": 15100,
  "error_count": 320,
  "error_rate": 0.021,
  "cache_hits": 7200,
  "cache_hit_rate": 0.467
}

2. Get Slow Queries:
GET /api/v1/performance/slow-queries
{
  "queries": [
    {
      "query": "SELECT * FROM users WHERE created_at > NOW() - INTERVAL 1 YEAR",
      "duration": 8932,
      "rows_scanned": 1000000,
      "rows_returned": 45000,
      "index_used": false,
      "cache_hit": false
    }
  ],
  "count": 42
}

3. Get Stats by Query Type:
GET /api/v1/performance/query-types
{
  "SELECT": {
    "count": 10000,
    "total_ms": 450000,
    "avg_ms": 45,
    "max_ms": 8932,
    "rows_scanned": 50000000
  },
  "INSERT": {...},
  "UPDATE": {...}
}

4. Get User Statistics:
GET /api/v1/performance/user-stats
{
  "user123": {
    "queries": 1500,
    "total_ms": 67500,
    "avg_ms": 45,
    "errors": 12,
    "last_run": "2024-01-27T10:30:00Z"
  }
}

5. Get Performance Recommendations:
GET /api/v1/performance/recommendations
{
  "recommendations": [
    {
      "type": "missing_index",
      "count": 15,
      "description": "Found 15 queries scanning >1000 rows without index",
      "priority": "high"
    },
    {
      "type": "slow_queries",
      "count": 42,
      "description": "Found 42 queries slower than 100ms",
      "priority": "medium"
    }
  ]
}

6. Get Percentiles:
GET /api/v1/performance/percentile/95
{
  "percentile": 95,
  "duration": 215
}

7. Record Query:
POST /api/v1/performance/record
{
  "query": "SELECT * FROM users",
  "duration": 45,
  "rows_scanned": 100000,
  "rows_returned": 50,
  "executed_by": "user123",
  "database": "postgresql",
  "status": "success",
  "index_used": true,
  "cache_hit": false,
  "query_type": "SELECT"
}

8. Get Performance Dashboard:
GET /api/v1/performance/dashboard
{
  "overall_stats": {...},
  "query_type_stats": {...},
  "user_stats": {...},
  "recommendations": [...],
  "p50_duration": 32,
  "p95_duration": 215,
  "p99_duration": 1250
}
`)
}

// ConfigurationExample shows configuration
func ConfigurationExample() {
	fmt.Println(`
=== CONFIGURATION EXAMPLES ===

1. QuotaManager Setup:
qm := ratelimit.NewQuotaManager()
qm.SetEndpointLimit("/api/users", ratelimit.QuotaLimit{
    RequestsPerSecond: 100,
    RequestsPerMinute: 6000,
    BytesPerMinute: 10485760, // 10MB
    MaxConcurrent: 100,
})

2. QueryPerformanceAnalyzer Setup:
analyzer := performance.NewQueryPerformanceAnalyzer(100, 10000) // 100ms threshold

3. Recording Query Performance:
analyzer.RecordQuery(performance.QueryPerformance{
    Query: "SELECT * FROM users",
    Duration: 45,
    RowsScanned: 100000,
    RowsReturned: 50,
    ExecutedBy: "user123",
    Status: "success",
    IndexUsed: true,
})

4. API Documentation Setup:
generator := docs.NewOpenAPIGenerator(docs.OpenAPIInfo{
    Title: "My API",
    Version: "1.0.0",
    Description: "API Description",
})

5. Phase 1 Features Integration:
features := integration.NewPhase1Features(db)
features.RegisterRoutes(router)
features.ApplyRateLimitMiddleware(router)
`)
}

// RunAllExamples runs all examples
func RunAllExamples() {
	GraphQLExample()
	fmt.Println("\n")
	RateLimitExample()
	fmt.Println("\n")
	APIDocumentationExample()
	fmt.Println("\n")
	PerformanceExample()
	fmt.Println("\n")
	ConfigurationExample()
}
