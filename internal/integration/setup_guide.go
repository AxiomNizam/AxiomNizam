package integration

import (
	"fmt"
)

// SetupGuide provides step-by-step setup instructions
func PrintSetupGuide() {
	fmt.Println(`
╔════════════════════════════════════════════════════════════════════════════════╗
║                    PHASE 1 FEATURES SETUP GUIDE                               ║
║           GraphQL | Rate Limiting | API Docs | Performance Monitor            ║
╚════════════════════════════════════════════════════════════════════════════════╝

STEP 1: UPDATE main.go
════════════════════════════════════════════════════════════════════════════════

Add these imports:
  "example.com/axiomnizam/internal/integration"

In your main() function, after creating the database connection:

  // Initialize Phase 1 features
  phase1 := integration.NewPhase1Features(db)
  phase1.RegisterRoutes(router)
  phase1.ApplyRateLimitMiddleware(router)


STEP 2: ADD DEPENDENCIES TO go.mod
════════════════════════════════════════════════════════════════════════════════

Run these commands:
  go get github.com/graphql-go/graphql
  go mod tidy


STEP 3: CREATE API ENDPOINTS
════════════════════════════════════════════════════════════════════════════════

All endpoints are automatically registered. Here's what's available:

GRAPHQL ENDPOINTS:
  POST   /api/graphql           - Execute GraphQL queries
  GET    /api/graphql/schema    - Get schema definition
  GET    /api/graphql/playground - Interactive playground

RATE LIMITING & QUOTAS:
  GET    /api/v1/quota/:user_id - Get user quota status
  PUT    /api/v1/quota/:user_id - Set user daily quota
  POST   /api/v1/quota/:user_id/reset - Reset user quota
  GET    /api/v1/quotas         - List all user quotas
  POST   /api/v1/endpoints/:endpoint/limit - Set endpoint rate limit

API DOCUMENTATION:
  GET    /api/docs/openapi.json - OpenAPI 3.0 specification
  GET    /api/docs/swagger      - Swagger UI
  GET    /api/docs/redoc        - ReDoc UI
  GET    /api/docs/markdown     - Markdown docs
  GET    /api/docs/endpoints    - List all endpoints
  GET    /api/docs/endpoints/:id - Get endpoint details

PERFORMANCE MONITORING:
  GET    /api/v1/performance/stats         - Overall statistics
  GET    /api/v1/performance/slow-queries  - Queries > threshold
  GET    /api/v1/performance/query-types   - Stats by query type
  GET    /api/v1/performance/user-stats    - Stats by user
  GET    /api/v1/performance/recommendations - Optimization suggestions
  GET    /api/v1/performance/percentile/:value - Duration at percentile
  POST   /api/v1/performance/record        - Record query performance
  GET    /api/v1/performance/dashboard     - Complete dashboard


STEP 4: INTEGRATE WITH EXISTING CODE
════════════════════════════════════════════════════════════════════════════════

A. Wrap database queries to track performance:

  import "example.com/axiomnizam/internal/performance"
  
  start := time.Now()
  rows, err := db.Query(query)
  duration := time.Since(start).Milliseconds()
  
  phase1.Analyzer.RecordQuery(performance.QueryPerformance{
    Query: query,
    Duration: duration,
    ExecutedBy: userID,
    Status: "success",
  })

B. Apply rate limiting to specific routes:

  protectedAPI := router.Group("/api/protected")
  protectedAPI.Use(phase1.RateLimitMiddleware.Handler())
  {
    protectedAPI.GET("/users", getUsersHandler)
    protectedAPI.POST("/users", createUserHandler)
  }

C. Set endpoint-specific quotas:

  phase1.QuotaManager.SetEndpointLimit("/api/users", ratelimit.QuotaLimit{
    RequestsPerSecond: 100,
    RequestsPerMinute: 6000,
    BytesPerMinute: 10485760, // 10MB
    MaxConcurrent: 100,
  })

D. Set user quotas:

  phase1.QuotaManager.SetUserDailyQuota("user123", 1000000)


STEP 5: VERIFY INSTALLATION
════════════════════════════════════════════════════════════════════════════════

1. Start the server:
   go run ./cmd/axiomnizam-server/main.go

2. Test GraphQL:
   curl -X POST http://localhost:8000/api/graphql \\
     -H "Content-Type: application/json" \\
     -d '{"query":"{ users(limit: 10) { id name email } }"}'

3. View Swagger UI:
   Open browser: http://localhost:8000/api/docs/swagger

4. Check Performance:
   curl http://localhost:8000/api/v1/performance/stats

5. Check User Quota:
   curl http://localhost:8000/api/v1/quota/user1


STEP 6: OPTIONAL CONFIGURATIONS
════════════════════════════════════════════════════════════════════════════════

A. Customize Query Performance Threshold:

  // Default: 100ms threshold
  analyzer := performance.NewQueryPerformanceAnalyzer(200, 10000)

B. Add Custom Documentation:

  generator.AddEndpoint(docs.OpenAPIEndpoint{
    Path: "/api/custom",
    Method: "GET",
    Summary: "Custom endpoint",
    Description: "Does something special",
    Tags: []string{"custom"},
  })

C. Configure Default Rate Limits:

  qm := ratelimit.NewQuotaManager()
  // Customize via qm.defaultLimit


TROUBLESHOOTING
════════════════════════════════════════════════════════════════════════════════

Q: GraphQL endpoint returns error about database schema
A: Make sure PostgreSQL information_schema is accessible
   Run: SELECT * FROM information_schema.tables;

Q: Rate limiting not working
A: Ensure middleware is registered with router.Use()
   Check that user_id is set in context

Q: No recommendations in performance analysis
A: Need at least 10 queries recorded with varied performance
   Performance data is kept for last 10,000 queries

Q: Documentation not showing all endpoints
A: Manually register endpoints with generator.AddEndpoint()


FEATURES IMPLEMENTED
════════════════════════════════════════════════════════════════════════════════

✓ GraphQL Schema Generation from Database
✓ GraphQL Query Execution
✓ Per-User Daily Quotas
✓ Per-Endpoint Rate Limiting
✓ Concurrent Request Limiting
✓ Automatic OpenAPI Spec Generation
✓ Swagger UI Integration
✓ ReDoc Integration
✓ Query Performance Tracking
✓ Slow Query Detection
✓ Performance Recommendations
✓ Percentile Analysis (P50, P95, P99)
✓ User and Query Type Statistics
✓ Cache Hit Rate Tracking
✓ Request Size Tracking


NEXT STEPS (Phase 2)
════════════════════════════════════════════════════════════════════════════════

1. Data Quality & Validation
2. Row-Level Security
3. Real-time Data Sync (CDC)
4. API Versioning

Estimated Timeline: 2-3 weeks


SUPPORT & RESOURCES
════════════════════════════════════════════════════════════════════════════════

Files Created:
  internal/graphql/schema.go
  internal/graphql/resolver.go
  internal/handlers/graphql_handler.go
  internal/ratelimit/quota_manager.go
  internal/ratelimit/middleware.go
  internal/docs/openapi.go
  internal/handlers/docs_handler.go
  internal/performance/analyzer.go
  internal/handlers/performance_handler.go
  internal/integration/phase1_features.go
  internal/integration/phase1_examples.go

Test Queries: See phase1_examples.go for complete examples

`)
}

// QuickStart provides quick start instructions
func PrintQuickStart() {
	fmt.Println(`
╔════════════════════════════════════════════════════════════════════════════════╗
║                         QUICK START (2 MINUTES)                               ║
╚════════════════════════════════════════════════════════════════════════════════╝

1. Add to main.go:
   ───────────────────────────────────────────────────────────────────────────
   phase1 := integration.NewPhase1Features(db)
   phase1.RegisterRoutes(router)

2. Install dependency:
   ───────────────────────────────────────────────────────────────────────────
   go get github.com/graphql-go/graphql

3. Run server:
   ───────────────────────────────────────────────────────────────────────────
   go run ./cmd/axiomnizam-server/main.go

4. Try features:
   ───────────────────────────────────────────────────────────────────────────
   
   GraphQL Query:
   curl -X POST http://localhost:8000/api/graphql \
     -H "Content-Type: application/json" \
     -d '{"query":"{ users(limit: 5) { id name } }"}'

   Check Quota:
   curl http://localhost:8000/api/v1/quota/user1

   View Docs:
   http://localhost:8000/api/docs/swagger

   Performance Dashboard:
   curl http://localhost:8000/api/v1/performance/dashboard

Done! ✓
`)
}
