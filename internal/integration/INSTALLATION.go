package integration

const GoModRequirements = `
# Add this to your go.mod for Phase 1 features

require (
    github.com/graphql-go/graphql v0.8.1
)

# Installation command:
# go get github.com/graphql-go/graphql

# Then run:
# go mod tidy

# All other dependencies should already be in your go.mod:
# - github.com/gin-gonic/gin
# - gorm.io/gorm
# - github.com/spf13/cobra
# - go.uber.org/zap
# - gopkg.in/yaml.v3
# - etc.
`

const InstallationSteps = `
╔════════════════════════════════════════════════════════════════════════════════╗
║                    INSTALLATION STEPS - Phase 1 Features                      ║
╚════════════════════════════════════════════════════════════════════════════════╝

STEP 1: Install GraphQL Dependency
════════════════════════════════════════════════════════════════════════════════

$ go get github.com/graphql-go/graphql
$ go mod tidy


STEP 2: Update Your main.go
════════════════════════════════════════════════════════════════════════════════

Add this import:
  import "example.com/axiomnizam/internal/integration"

Add this in your main() after database initialization:
  
  // Initialize Phase 1 Features (GraphQL, Rate Limiting, Docs, Performance)
  phase1 := integration.NewPhase1Features(db)
  phase1.RegisterRoutes(router)
  phase1.ApplyRateLimitMiddleware(router)
  
  // Optional: Print setup instructions
  integration.PrintSetupGuide()


STEP 3: Build and Run
════════════════════════════════════════════════════════════════════════════════

$ go mod download
$ go build -o axiomnizam ./cmd/axiomnizam-server/
$ ./axiomnizam


STEP 4: Verify Installation
════════════════════════════════════════════════════════════════════════════════

Test GraphQL:
  $ curl -X POST http://localhost:8000/api/graphql \
      -H "Content-Type: application/json" \
      -d '{"query":"{ users(limit: 5) { id name email } }"}'

View Swagger UI:
  Open: http://localhost:8000/api/docs/swagger

Check Performance:
  $ curl http://localhost:8000/api/v1/performance/stats

Check User Quotas:
  $ curl http://localhost:8000/api/v1/quota/user1


TROUBLESHOOTING
════════════════════════════════════════════════════════════════════════════════

Problem: "cannot find package github.com/graphql-go/graphql"
Solution: Run: go get github.com/graphql-go/graphql

Problem: "undefined: integration.NewPhase1Features"
Solution: Make sure import path is correct in main.go

Problem: GraphQL returns "database schema error"
Solution: Ensure PostgreSQL has public schema with tables

Problem: Rate limiting not working
Solution: Check that middleware is registered with RegisterRoutes()

Problem: Documentation endpoints return 404
Solution: Verify DocsHandler is registered in RegisterRoutes()

Problem: Performance analytics empty
Solution: Need to record queries first with POST /api/v1/performance/record


DATABASE REQUIREMENTS
════════════════════════════════════════════════════════════════════════════════

For GraphQL to work:
  ✓ PostgreSQL 10+ (tested)
  ✓ MySQL 5.7+ (should work)
  ✓ Other GORM-supported databases (may need schema adjustments)

Required:
  ✓ information_schema access for schema introspection
  ✓ SELECT permission on all tables
  ✓ PostgreSQL connection string in environment


ENVIRONMENT VARIABLES (Optional)
════════════════════════════════════════════════════════════════════════════════

QUERY_SLOW_THRESHOLD=100              # milliseconds (default: 100)
PERFORMANCE_STORAGE_SIZE=10000        # max queries to store (default: 10000)
RATE_LIMIT_ENABLED=true               # enable rate limiting (default: true)
GRAPHQL_PLAYGROUND_ENABLED=true       # enable playground (default: true)
DOCS_UI_ENABLED=true                  # enable docs endpoints (default: true)


FILE STRUCTURE CREATED
════════════════════════════════════════════════════════════════════════════════

internal/
├── graphql/
│   ├── schema.go          - GraphQL schema builder
│   └── resolver.go        - Query resolver
├── ratelimit/
│   ├── quota_manager.go   - Quota management
│   └── middleware.go      - Gin middleware + handlers
├── docs/
│   └── openapi.go         - OpenAPI spec generation
├── performance/
│   └── analyzer.go        - Performance analysis
├── handlers/
│   ├── graphql_handler.go       - GraphQL HTTP handlers
│   ├── docs_handler.go          - Documentation handlers
│   └── performance_handler.go   - Performance endpoints
└── integration/
    ├── phase1_features.go       - Main integration
    ├── phase1_examples.go       - Usage examples
    ├── phase1_tests.go          - Test suite
    ├── setup_guide.go           - Setup instructions
    └── PHASE1_SUMMARY.go        - Implementation summary


FEATURE CHECKLIST
════════════════════════════════════════════════════════════════════════════════

GraphQL Support
  ☑ Schema generation from database
  ☑ Query execution
  ☑ Playground interface
  ☑ Schema introspection

Rate Limiting & Quotas
  ☑ Per-user daily quotas
  ☑ Per-endpoint limits
  ☑ Byte rate limiting
  ☑ Concurrent request limiting
  ☑ Admin management endpoints
  ☑ Response headers

API Documentation
  ☑ OpenAPI 3.0 generation
  ☑ Swagger UI
  ☑ ReDoc UI
  ☑ Markdown export
  ☑ Endpoint listing

Performance Monitoring
  ☑ Query timing
  ☑ Slow query detection
  ☑ Cache hit tracking
  ☑ Error rate monitoring
  ☑ Percentile analysis
  ☑ Optimization recommendations


NEXT STEPS
════════════════════════════════════════════════════════════════════════════════

1. Review setup_guide.go for detailed configuration
2. Check phase1_examples.go for API usage examples
3. Run tests: go test ./internal/integration/... -v
4. Deploy and monitor performance with dashboard
5. Plan Phase 2 features


GETTING HELP
════════════════════════════════════════════════════════════════════════════════

For issues:
  1. Check setup_guide.go troubleshooting section
  2. Review phase1_examples.go for correct usage
  3. Check test cases in phase1_tests.go
  4. Verify database connectivity

For features:
  1. GraphQL playground: /api/graphql/playground
  2. API docs: /api/docs/swagger
  3. Performance dashboard: /api/v1/performance/dashboard
  4. Quota management: /api/v1/quota/...
`

const DeploymentChecklist = `
╔════════════════════════════════════════════════════════════════════════════════╗
║                        DEPLOYMENT CHECKLIST                                   ║
╚════════════════════════════════════════════════════════════════════════════════╝

Pre-Deployment
  ☐ All tests passing: go test ./... -v
  ☐ No compilation warnings: go build -v
  ☐ Dependencies resolved: go mod tidy
  ☐ Database backup created
  ☐ Performance baseline established

During Deployment
  ☐ Backup current configuration
  ☐ Update binary
  ☐ Run database migrations (if any)
  ☐ Restart services
  ☐ Monitor logs for errors

Post-Deployment
  ☐ Verify GraphQL endpoint responds
  ☐ Check Swagger UI loads
  ☐ Confirm rate limiting works
  ☐ Test quota management
  ☐ Validate performance dashboard
  ☐ Monitor error rates
  ☐ Confirm cache hit rates acceptable

Health Checks
  ☐ GET /health returns 200
  ☐ GET /api/docs/swagger loads
  ☐ POST /api/graphql executes queries
  ☐ GET /api/v1/performance/stats returns data
  ☐ GET /api/v1/quota/test-user returns quota
`

// PrintInstallationGuide prints installation instructions
func PrintInstallationGuide() {
	println(InstallationSteps)
}

// PrintDeploymentChecklist prints deployment checklist
func PrintDeploymentChecklist() {
	println(DeploymentChecklist)
}
