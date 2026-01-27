package integration

import (
	"example.com/axiomnizam/internal/docs"
	"example.com/axiomnizam/internal/handlers"
	"example.com/axiomnizam/internal/performance"
	"example.com/axiomnizam/internal/ratelimit"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Phase1Features integrates all Phase 1 features
type Phase1Features struct {
	GraphQLHandler      *handlers.GraphQLHandler
	QuotaHandler        *ratelimit.QuotaHandler
	RateLimitMiddleware *ratelimit.RateLimitMiddleware
	DocsHandler         *docs.DocsHandler
	PerformanceHandler  *handlers.PerformanceHandler
	QuotaManager        *ratelimit.QuotaManager
	Analyzer            *performance.QueryPerformanceAnalyzer
}

// NewPhase1Features initializes all Phase 1 features
func NewPhase1Features(db *gorm.DB) *Phase1Features {
	// Initialize GraphQL
	graphqlHandler := handlers.NewGraphQLHandler(db)

	// Initialize Rate Limiting & Quotas
	quotaManager := ratelimit.NewQuotaManager()
	rateLimitMiddleware := ratelimit.NewRateLimitMiddleware(quotaManager)
	quotaHandler := ratelimit.NewQuotaHandler(quotaManager)

	// Initialize API Documentation
	apiInfo := docs.OpenAPIInfo{
		Title:       "AxiomNizam API",
		Version:     "1.0.0",
		Description: "Enterprise Kubernetes-style data control plane",
	}
	docsGenerator := docs.NewOpenAPIGenerator(apiInfo)
	docsHandler := docs.NewDocsHandler(docsGenerator)

	// Initialize Query Performance Analyzer
	analyzer := performance.NewQueryPerformanceAnalyzer(100, 10000) // 100ms threshold
	performanceHandler := handlers.NewPerformanceHandler(analyzer)

	return &Phase1Features{
		GraphQLHandler:      graphqlHandler,
		QuotaHandler:        quotaHandler,
		RateLimitMiddleware: rateLimitMiddleware,
		DocsHandler:         docsHandler,
		PerformanceHandler:  performanceHandler,
		QuotaManager:        quotaManager,
		Analyzer:            analyzer,
	}
}

// RegisterRoutes registers all Phase 1 feature routes
func (pf *Phase1Features) RegisterRoutes(router *gin.Engine) {
	// GraphQL endpoints
	graphql := router.Group("/api/graphql")
	{
		graphql.POST("", pf.GraphQLHandler.Query)
		graphql.GET("/schema", pf.GraphQLHandler.GetSchema)
		graphql.GET("/playground", pf.GraphQLHandler.Playground)
	}

	// Rate Limiting & Quotas endpoints
	quota := router.Group("/api/v1/quota")
	quota.Use(pf.RateLimitMiddleware.Handler())
	{
		quota.GET("/:user_id", pf.QuotaHandler.GetQuota)
		quota.PUT("/:user_id", pf.QuotaHandler.SetUserQuota)
		quota.POST("/:user_id/reset", pf.QuotaHandler.ResetQuota)
		quota.GET("", pf.QuotaHandler.ListQuotas)
	}

	endpoints := router.Group("/api/v1/endpoints")
	{
		endpoints.POST("/:endpoint/limit", pf.QuotaHandler.SetEndpointLimit)
	}

	// API Documentation endpoints
	docs := router.Group("/api/docs")
	{
		docs.GET("/openapi.json", pf.DocsHandler.GetOpenAPISpec)
		docs.GET("/swagger", pf.DocsHandler.GetSwaggerUI)
		docs.GET("/redoc", pf.DocsHandler.GetReDocUI)
		docs.GET("/markdown", pf.DocsHandler.GetMarkdownDocs)
		docs.GET("/endpoints", pf.DocsHandler.ListEndpoints)
		docs.GET("/endpoints/:id", pf.DocsHandler.GetEndpointDetails)
	}

	// Performance Monitoring endpoints
	perf := router.Group("/api/v1/performance")
	{
		perf.GET("/stats", pf.PerformanceHandler.GetStats)
		perf.GET("/slow-queries", pf.PerformanceHandler.GetSlowQueries)
		perf.GET("/query-types", pf.PerformanceHandler.GetQueryTypeStats)
		perf.GET("/user-stats", pf.PerformanceHandler.GetUserStats)
		perf.GET("/recommendations", pf.PerformanceHandler.GetRecommendations)
		perf.GET("/percentile/:value", pf.PerformanceHandler.GetPercentile)
		perf.POST("/record", pf.PerformanceHandler.RecordQuery)
		perf.GET("/dashboard", pf.PerformanceHandler.GetDashboard)
	}
}

// ApplyRateLimitMiddleware applies rate limiting to all routes
func (pf *Phase1Features) ApplyRateLimitMiddleware(router *gin.Engine) {
	router.Use(pf.RateLimitMiddleware.Handler())
}

// RegisterQuotaMiddleware registers quota tracking on specific routes
func (pf *Phase1Features) RegisterQuotaMiddleware(group *gin.RouterGroup) {
	group.Use(pf.RateLimitMiddleware.Handler())
}
