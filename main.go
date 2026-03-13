package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"example.com/axiomnizam/internal/auth"
	"example.com/axiomnizam/internal/config"
	"example.com/axiomnizam/internal/database"
	"example.com/axiomnizam/internal/handlers"
	"example.com/axiomnizam/internal/models"
	"example.com/axiomnizam/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}
	fmt.Println("🚀 Starting AxiomNizam with Kubernetes-style Runtime...\n")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize Runtime
	log.Println("📦 Initializing Kubernetes-style runtime...")
	rt := runtime.NewRuntime("1.0.0")

	if err := rt.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Keycloak token validator
	keycloakConfig := &auth.KeycloakConfig{
		ServerURL: cfg.GetKeycloakURL(),
		Realm:     cfg.Keycloak.Realm,
		ClientID:  cfg.Keycloak.ClientID,
	}
	tokenValidator, err := auth.NewTokenValidator(keycloakConfig)
	if err != nil {
		log.Printf("⚠️  Keycloak initialization failed: %v (running without auth)", err)
		tokenValidator = nil
	}

	// Initialize all connections
	conns := database.InitConnections(cfg)

	// Create tables
	createTables(conns)

	// Create Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-KEY")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Add API Metrics tracking middleware
	// Initialize first before adding middleware
	apiMetricsTracker := handlers.NewAPIMetricsTracker(conns.Valkey)
	router.Use(handlers.MetricsMiddleware(apiMetricsTracker))

	// Initialize Rate Limiter
	// Max calls and token validity from config (.env)
	rateLimiter := auth.NewRateLimiter(cfg.RateLimiting.MaxCallsPerToken, cfg.RateLimiting.TokenValidityMinutes)

	// Initialize Query Logger with Valkey/Redis
	queryLogger := handlers.NewQueryLogger(conns.Valkey, "/data/query_logs")

	// Initialize all handlers
	healthHandler := handlers.NewHealthHandler(conns)
	userHandler := handlers.NewUserHandler(conns.MySQL)
	mariadbHandler := handlers.NewUserHandler(conns.MariaDB)
	postgresHandler := handlers.NewUserHandler(conns.PostgreSQL)
	perconaHandler := handlers.NewUserHandler(conns.Percona)
	mongoHandler := handlers.NewMongoDBHandler(conns.MongoDB)
	firebaseHandler := handlers.NewFirebaseHandler("http://firebase:9000")
	oracleHandler := handlers.NewOracleHandler(conns.Oracle)

	// Admin handler for database and table creation
	// Only include SQL databases (MongoDB and Firebase don't support SQL DDL operations)
	dbConnections := map[string]*gorm.DB{
		"mysql":    conns.MySQL,
		"mariadb":  conns.MariaDB,
		"postgres": conns.PostgreSQL,
		"percona":  conns.Percona,
		"oracle":   conns.Oracle,
	}
	adminHandler := handlers.NewAdminHandler(dbConnections)

	// User management handler
	platformUserHandler := handlers.NewPlatformUserHandler()

	// Dynamic Query handlers for each database
	mysqlDynamicHandler := handlers.NewDynamicQueryHandler(conns.MySQL, queryLogger)
	mariadbDynamicHandler := handlers.NewDynamicQueryHandler(conns.MariaDB, queryLogger)
	postgresDynamicHandler := handlers.NewDynamicQueryHandler(conns.PostgreSQL, queryLogger)
	perconaDynamicHandler := handlers.NewDynamicQueryHandler(conns.Percona, queryLogger)
	oracleDynamicHandler := handlers.NewDynamicQueryHandler(conns.Oracle, queryLogger)

	// Notification handler
	discordWebhookURL := cfg.Discord.WebhookURL
	notificationHandler := handlers.NewNotificationHandler(discordWebhookURL, dbConnections)

	// GraphQL handler (prefer PostgreSQL for schema introspection; fallback to available SQL engines)
	graphQLDB := conns.PostgreSQL
	if graphQLDB == nil {
		graphQLDB = conns.MySQL
	}
	if graphQLDB == nil {
		graphQLDB = conns.MariaDB
	}
	if graphQLDB == nil {
		graphQLDB = conns.Percona
	}
	if graphQLDB == nil {
		graphQLDB = conns.Oracle
	}
	graphQLHandler := handlers.NewGraphQLHandler(graphQLDB)

	// Apply auth middleware to protected routes
	var authMiddleware gin.HandlerFunc
	if tokenValidator != nil {
		authMiddleware = auth.CombinedAuthMiddleware(tokenValidator, rateLimiter)
	} else {
		authMiddleware = func(c *gin.Context) { c.Next() }
	}

	// Health check endpoints (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/status", healthHandler.Status)
	router.GET("/distributed", healthHandler.Distributed)

	// Authentication endpoints (no auth required for login/refresh)
	authHandler := handlers.NewAuthHandler()
	authHandler.SetRateLimiter(rateLimiter)
	authHandler.SetPlatformUserHandler(platformUserHandler)
	router.POST("/auth/login", authHandler.Login)
	router.POST("/auth/refresh", authHandler.RefreshToken)
	router.GET("/auth/validate", authHandler.ValidateToken)

	// Token status endpoints (auth required)
	router.GET("/auth/token-status", authMiddleware, authHandler.GetTokenStatus)
	router.GET("/auth/admin/tokens-status", authMiddleware, auth.RequireAdmin(), authHandler.GetAllTokensStatus)

	// Context enrichment middleware - populates database name and user info for logging
	contextEnrichmentMiddleware := func(c *gin.Context) {
		// Extract database name from URL path (e.g., /api/mysql/query -> mysql)
		pathParts := strings.Split(c.Request.URL.Path, "/")
		if len(pathParts) >= 3 {
			dbName := pathParts[2]
			switch dbName {
			case "mysql", "mariadb", "postgres", "percona", "oracle":
				c.Set("database", dbName)
			}
		}

		// Extract user info from token claims if available
		if claims, exists := c.Get("claims"); exists {
			if tokenClaims, ok := claims.(jwt.MapClaims); ok {
				if userID, ok := tokenClaims["sub"]; ok {
					c.Set("user_id", userID)
				}
			}
		}

		c.Next()
	}

	// Apply context enrichment to auth middleware
	originalAuthMiddleware := authMiddleware
	authMiddleware = func(c *gin.Context) {
		originalAuthMiddleware(c)
		if !c.IsAborted() {
			contextEnrichmentMiddleware(c)
		}
	}

	// Get admin middleware (requires admin role)
	var adminMiddleware gin.HandlerFunc
	var adminOrSysMiddleware gin.HandlerFunc
	if tokenValidator != nil {
		adminMiddleware = func(c *gin.Context) {
			authMiddleware(c)
			if !c.IsAborted() {
				auth.RequireAdmin()(c)
			}
		}
		adminOrSysMiddleware = func(c *gin.Context) {
			authMiddleware(c)
			if !c.IsAborted() {
				auth.RequireAnyRole("admin", "system-manager", "sysadmin", "system_admin", "system-admin")(c)
			}
		}
	} else {
		adminMiddleware = func(c *gin.Context) { c.Next() }
		adminOrSysMiddleware = func(c *gin.Context) { c.Next() }
	}

	// GraphQL endpoints (auth required)
	router.POST("/api/graphql", authMiddleware, graphQLHandler.Query)
	router.GET("/api/graphql/schema", authMiddleware, graphQLHandler.GetSchema)
	router.GET("/api/graphql/playground", authMiddleware, graphQLHandler.Playground)

	// CRUD routes for MySQL
	// Read operations (allowed for all authenticated users)
	router.GET("/api/mysql/users", authMiddleware, userHandler.GetAllUsers)
	router.GET("/api/mysql/users/:id", authMiddleware, userHandler.GetUserByID)
	// Write operations (allowed only for admin users)
	router.POST("/api/mysql/users", adminMiddleware, userHandler.CreateUser)
	router.PUT("/api/mysql/users/:id", adminMiddleware, userHandler.UpdateUser)
	router.DELETE("/api/mysql/users/:id", adminMiddleware, userHandler.DeleteUser)

	// CRUD routes for MariaDB
	// Read operations (allowed for all authenticated users)
	router.GET("/api/mariadb/users", authMiddleware, mariadbHandler.GetAllUsers)
	router.GET("/api/mariadb/users/:id", authMiddleware, mariadbHandler.GetUserByID)
	// Write operations (allowed only for admin users)
	router.POST("/api/mariadb/users", adminMiddleware, mariadbHandler.CreateUser)
	router.PUT("/api/mariadb/users/:id", adminMiddleware, mariadbHandler.UpdateUser)
	router.DELETE("/api/mariadb/users/:id", adminMiddleware, mariadbHandler.DeleteUser)

	// CRUD routes for PostgreSQL
	// Read operations (allowed for all authenticated users)
	router.GET("/api/postgres/users", authMiddleware, postgresHandler.GetAllUsers)
	router.GET("/api/postgres/users/:id", authMiddleware, postgresHandler.GetUserByID)
	// Write operations (allowed only for admin users)
	router.POST("/api/postgres/users", adminMiddleware, postgresHandler.CreateUser)
	router.PUT("/api/postgres/users/:id", adminMiddleware, postgresHandler.UpdateUser)
	router.DELETE("/api/postgres/users/:id", adminMiddleware, postgresHandler.DeleteUser)

	// CRUD routes for Percona
	// Read operations (allowed for all authenticated users)
	router.GET("/api/percona/users", authMiddleware, perconaHandler.GetAllUsers)
	router.GET("/api/percona/users/:id", authMiddleware, perconaHandler.GetUserByID)
	// Write operations (allowed only for admin users)
	router.POST("/api/percona/users", adminMiddleware, perconaHandler.CreateUser)
	router.PUT("/api/percona/users/:id", adminMiddleware, perconaHandler.UpdateUser)
	router.DELETE("/api/percona/users/:id", adminMiddleware, perconaHandler.DeleteUser)

	// CRUD routes for MongoDB
	// Read operations (allowed for all authenticated users)
	router.GET("/api/mongodb/users", authMiddleware, mongoHandler.GetAllUsers)
	router.GET("/api/mongodb/users/:id", authMiddleware, mongoHandler.GetUserByID)
	// Write operations (allowed only for admin users)
	router.POST("/api/mongodb/users", adminMiddleware, mongoHandler.CreateUser)
	router.PUT("/api/mongodb/users/:id", adminMiddleware, mongoHandler.UpdateUser)
	router.DELETE("/api/mongodb/users/:id", adminMiddleware, mongoHandler.DeleteUser)

	// CRUD routes for Firebase
	// Read operations (allowed for all authenticated users)
	router.GET("/api/firebase/users", authMiddleware, firebaseHandler.GetAllUsers)
	router.GET("/api/firebase/users/:id", authMiddleware, firebaseHandler.GetUserByID)
	// Write operations (allowed only for admin users)
	router.POST("/api/firebase/users", adminMiddleware, firebaseHandler.CreateUser)
	router.PUT("/api/firebase/users/:id", adminMiddleware, firebaseHandler.UpdateUser)
	router.DELETE("/api/firebase/users/:id", adminMiddleware, firebaseHandler.DeleteUser)

	// CRUD routes for Oracle
	// Read operations (allowed for all authenticated users)
	router.GET("/api/oracle/users", authMiddleware, oracleHandler.GetAllUsers)
	router.GET("/api/oracle/users/:id", authMiddleware, oracleHandler.GetUserByID)
	// Write operations (allowed only for admin users)
	router.POST("/api/oracle/users", adminMiddleware, oracleHandler.CreateUser)
	router.PUT("/api/oracle/users/:id", adminMiddleware, oracleHandler.UpdateUser)
	router.DELETE("/api/oracle/users/:id", adminMiddleware, oracleHandler.DeleteUser)

	// ====================================
	// DYNAMIC QUERY ENDPOINTS (Auth Required)
	// ====================================
	// These endpoints allow dynamic SQL queries via Postman or any HTTP client
	// GET requests only support SELECT queries
	// POST requests support all query types (SELECT, INSERT, UPDATE, DELETE, CREATE, etc.)

	// MySQL Dynamic Queries
	router.GET("/api/mysql/query", authMiddleware, mysqlDynamicHandler.DynamicQuery)
	router.POST("/api/mysql/query", authMiddleware, mysqlDynamicHandler.DynamicQueryWithBody)
	router.POST("/api/mysql/query/batch", authMiddleware, mysqlDynamicHandler.BatchQueries)
	router.GET("/api/mysql/schema", authMiddleware, mysqlDynamicHandler.TableSchema)

	// MariaDB Dynamic Queries
	router.GET("/api/mariadb/query", authMiddleware, mariadbDynamicHandler.DynamicQuery)
	router.POST("/api/mariadb/query", authMiddleware, mariadbDynamicHandler.DynamicQueryWithBody)
	router.POST("/api/mariadb/query/batch", authMiddleware, mariadbDynamicHandler.BatchQueries)
	router.GET("/api/mariadb/schema", authMiddleware, mariadbDynamicHandler.TableSchema)

	// PostgreSQL Dynamic Queries
	router.GET("/api/postgres/query", authMiddleware, postgresDynamicHandler.DynamicQuery)
	router.POST("/api/postgres/query", authMiddleware, postgresDynamicHandler.DynamicQueryWithBody)
	router.POST("/api/postgres/query/batch", authMiddleware, postgresDynamicHandler.BatchQueries)
	router.GET("/api/postgres/schema", authMiddleware, postgresDynamicHandler.TableSchema)

	// Percona Dynamic Queries
	router.GET("/api/percona/query", authMiddleware, perconaDynamicHandler.DynamicQuery)
	router.POST("/api/percona/query", authMiddleware, perconaDynamicHandler.DynamicQueryWithBody)
	router.POST("/api/percona/query/batch", authMiddleware, perconaDynamicHandler.BatchQueries)
	router.GET("/api/percona/schema", authMiddleware, perconaDynamicHandler.TableSchema)

	// Oracle Dynamic Queries
	router.GET("/api/oracle/query", authMiddleware, oracleDynamicHandler.DynamicQuery)
	router.POST("/api/oracle/query", authMiddleware, oracleDynamicHandler.DynamicQueryWithBody)
	router.POST("/api/oracle/query/batch", authMiddleware, oracleDynamicHandler.BatchQueries)
	router.GET("/api/oracle/schema", authMiddleware, oracleDynamicHandler.TableSchema)

	// ====================================
	// QUERY LOGGING & STATISTICS
	// ====================================

	// MySQL Logging
	router.GET("/api/mysql/logs", authMiddleware, mysqlDynamicHandler.GetQueryLogs)
	router.GET("/api/mysql/stats", authMiddleware, mysqlDynamicHandler.GetQueryStats)

	// MariaDB Logging
	router.GET("/api/mariadb/logs", authMiddleware, mariadbDynamicHandler.GetQueryLogs)
	router.GET("/api/mariadb/stats", authMiddleware, mariadbDynamicHandler.GetQueryStats)

	// PostgreSQL Logging
	router.GET("/api/postgres/logs", authMiddleware, postgresDynamicHandler.GetQueryLogs)
	router.GET("/api/postgres/stats", authMiddleware, postgresDynamicHandler.GetQueryStats)

	// Percona Logging
	router.GET("/api/percona/logs", authMiddleware, perconaDynamicHandler.GetQueryLogs)
	router.GET("/api/percona/stats", authMiddleware, perconaDynamicHandler.GetQueryStats)

	// Oracle Logging
	router.GET("/api/oracle/logs", authMiddleware, oracleDynamicHandler.GetQueryLogs)
	router.GET("/api/oracle/stats", authMiddleware, oracleDynamicHandler.GetQueryStats)

	// ====================================
	// DATA TRANSFORMATION ENDPOINTS (Auth Required)
	// ====================================

	transformHandler := handlers.NewTransformationHandler()

	// Rule Management endpoints
	router.POST("/api/transform/rules", authMiddleware, transformHandler.RegisterRule)
	router.GET("/api/transform/rules", authMiddleware, transformHandler.ListRules)
	router.GET("/api/transform/rules/:name", authMiddleware, transformHandler.GetRule)
	router.DELETE("/api/transform/rules/:name", adminMiddleware, transformHandler.DeleteRule)

	// Transformation endpoints
	router.POST("/api/transform/apply", authMiddleware, transformHandler.Transform)
	router.POST("/api/transform/batch", authMiddleware, transformHandler.TransformBatch)
	router.POST("/api/transform/preview", authMiddleware, transformHandler.PreviewTransformation)

	// Feature Testing endpoints
	router.POST("/api/transform/test/rename", authMiddleware, transformHandler.TestFieldRename)
	router.POST("/api/transform/test/types", authMiddleware, transformHandler.TestTypeConversion)
	router.POST("/api/transform/test/flatten", authMiddleware, transformHandler.TestFlattening)

	// Import/Export endpoints
	router.GET("/api/transform/rules/export", authMiddleware, transformHandler.ExportRules)
	router.POST("/api/transform/rules/import", adminMiddleware, transformHandler.ImportRules)

	// ====================================
	// ADMIN OPERATIONS (Admin Only)
	// ====================================

	// Database management endpoints (admin only)
	router.POST("/api/admin/database/create", adminOrSysMiddleware, adminHandler.CreateDatabase)
	router.GET("/api/admin/database/list", adminOrSysMiddleware, adminHandler.ListDatabases)
	router.GET("/api/admin/database/servers", adminOrSysMiddleware, adminHandler.ListDatabaseServers)
	router.POST("/api/admin/database/connect", adminOrSysMiddleware, adminHandler.ConnectDatabaseServer)

	// Table management endpoints (admin only)
	router.POST("/api/admin/table/create", adminOrSysMiddleware, adminHandler.CreateTable)
	router.GET("/api/admin/table/list", adminOrSysMiddleware, adminHandler.ListTables)

	// User management endpoints (admin only)
	router.GET("/api/v1/users", adminOrSysMiddleware, platformUserHandler.ListPlatformUsers)
	router.GET("/api/v1/users/:id", adminOrSysMiddleware, platformUserHandler.GetPlatformUser)
	router.POST("/api/v1/users", adminOrSysMiddleware, platformUserHandler.CreatePlatformUser)
	router.PUT("/api/v1/users/:id", adminOrSysMiddleware, platformUserHandler.UpdatePlatformUser)
	router.DELETE("/api/v1/users/:id", adminOrSysMiddleware, platformUserHandler.DeletePlatformUser)

	// API Metrics endpoints (admin only)
	router.GET("/api/admin/metrics/all", adminOrSysMiddleware, apiMetricsTracker.GetAllAPIMetrics)
	router.GET("/api/admin/metrics/count", adminOrSysMiddleware, apiMetricsTracker.GetAPICount)
	router.GET("/api/admin/metrics/stats", adminOrSysMiddleware, apiMetricsTracker.GetAPIStats)

	// ====================================
	// NOTIFICATION ENDPOINTS (Auth Required)
	// ====================================

	// Notification endpoints (authenticated users)
	router.POST("/api/notifications/send", authMiddleware, notificationHandler.SendNotification)
	router.POST("/api/notifications/health", authMiddleware, notificationHandler.SendHealthNotification)
	router.POST("/api/notifications/status", authMiddleware, notificationHandler.SendStatusNotification)
	router.GET("/api/notifications/status", notificationHandler.GetNotificationStatus)

	// ====================================
	// CLI AUTHENTICATION ENDPOINTS
	// ====================================
	cliAuth := handlers.NewCLIAuthHandler()
	router.POST("/api/v1/auth/login", cliAuth.Login)
	router.GET("/api/v1/auth/verify", cliAuth.Verify)
	router.GET("/api/v1/auth/whoami", cliAuth.WhoAmI)

	// ====================================
	// KUBERNETES-STYLE RESOURCE ENDPOINTS
	// ====================================
	resourceHandler := handlers.NewResourceHandler()

	// Namespaced resource endpoints: /api/v1/namespaces/{namespace}/{kind}
	nsAPI := router.Group("/api/v1/namespaces")
	{
		nsAPI.POST("/:namespace/:kind", adminOrSysMiddleware, resourceHandler.CreateOrUpdate)
		nsAPI.GET("/:namespace/:kind", authMiddleware, resourceHandler.List)
		nsAPI.GET("/:namespace/:kind/:name", authMiddleware, resourceHandler.Get)
		nsAPI.PUT("/:namespace/:kind/:name", adminOrSysMiddleware, resourceHandler.Update)
		nsAPI.DELETE("/:namespace/:kind/:name", adminOrSysMiddleware, resourceHandler.Delete)
		nsAPI.GET("/:namespace/:kind/:name/status", authMiddleware, resourceHandler.GetStatus)
		nsAPI.GET("/:namespace/:kind/:name/events", authMiddleware, resourceHandler.Events)
	}

	// Non-namespaced resource endpoints: /api/v1/{kind}
	router.POST("/api/v1/apis", adminOrSysMiddleware, resourceHandler.CreateOrUpdate)
	router.GET("/api/v1/apis", authMiddleware, resourceHandler.ListAll)
	router.POST("/api/v1/policies", adminOrSysMiddleware, resourceHandler.CreateOrUpdate)
	router.GET("/api/v1/policies", authMiddleware, resourceHandler.ListAll)
	router.POST("/api/v1/workflows", adminOrSysMiddleware, resourceHandler.CreateOrUpdate)
	router.GET("/api/v1/workflows", authMiddleware, resourceHandler.ListAll)
	router.POST("/api/v1/workflows/:name/run", adminOrSysMiddleware, func(c *gin.Context) {
		c.JSON(200, gin.H{"message": fmt.Sprintf("Workflow '%s' started", c.Param("name")), "status": "Running"})
	})

	// DataSource endpoints
	dsHandler := handlers.NewDataSourceHandler()
	router.POST("/api/v1/datasources", adminOrSysMiddleware, dsHandler.Create)
	router.GET("/api/v1/datasources", authMiddleware, dsHandler.List)
	router.GET("/api/v1/datasources/:name", authMiddleware, dsHandler.Get)
	router.PUT("/api/v1/datasources/:name", adminOrSysMiddleware, dsHandler.Update)
	router.DELETE("/api/v1/datasources/:name", adminOrSysMiddleware, dsHandler.Delete)
	router.POST("/api/v1/datasources/:name/test", adminOrSysMiddleware, dsHandler.Test)

	// Job endpoints
	jobHandler := handlers.NewJobHandler()
	router.POST("/api/v1/jobs", adminOrSysMiddleware, jobHandler.Create)
	router.GET("/api/v1/jobs", authMiddleware, jobHandler.List)
	router.GET("/api/v1/jobs/:id", authMiddleware, jobHandler.Get)
	router.POST("/api/v1/jobs/:id/run", adminOrSysMiddleware, jobHandler.Run)
	router.GET("/api/v1/jobs/:id/logs", authMiddleware, jobHandler.GetLogs)
	router.POST("/api/v1/jobs/:id/cancel", adminOrSysMiddleware, jobHandler.Cancel)
	router.DELETE("/api/v1/jobs/:id", adminOrSysMiddleware, jobHandler.Delete)

	// ====================================
	// GIS DASHBOARD ENDPOINTS
	// ====================================
	gisHandler := handlers.NewGISHandler()
	gisAPI := router.Group("/api/v1/gis")
	{
		gisAPI.GET("/summary", gisHandler.GetSummary)

		gisAPI.GET("/layers", gisHandler.ListLayers)
		gisAPI.POST("/layers", gisHandler.CreateLayer)
		gisAPI.PUT("/layers/:id", gisHandler.UpdateLayer)
		gisAPI.DELETE("/layers/:id", gisHandler.DeleteLayer)

		gisAPI.GET("/regions", gisHandler.ListRegions)
		gisAPI.GET("/regions/:id", gisHandler.GetRegion)
		gisAPI.POST("/regions", gisHandler.CreateRegion)
		gisAPI.PUT("/regions/:id", gisHandler.UpdateRegion)
		gisAPI.DELETE("/regions/:id", gisHandler.DeleteRegion)

		gisAPI.GET("/markers", gisHandler.ListMarkers)
		gisAPI.POST("/markers", gisHandler.CreateMarker)
		gisAPI.DELETE("/markers/:id", gisHandler.DeleteMarker)

		gisAPI.GET("/datasets", gisHandler.ListDatasets)
		gisAPI.GET("/datasets/:id", gisHandler.GetDataset)
		gisAPI.POST("/datasets", gisHandler.CreateDataset)
		gisAPI.PUT("/datasets/:id", gisHandler.UpdateDataset)
		gisAPI.DELETE("/datasets/:id", gisHandler.DeleteDataset)
	}

	// Specialized GIS dashboards (agriculture, industries, medical, satellite, airplane, ship)
	gisSpecHandler := handlers.NewGISSpecializedHandler()
	gisSpecAPI := router.Group("/api/v1/gis/dashboards")
	{
		gisSpecAPI.GET("", gisSpecHandler.ListDashboardTypes)
		gisSpecAPI.GET("/:type", gisSpecHandler.GetDashboard)
		gisSpecAPI.GET("/:type/summary", gisSpecHandler.GetDashboardSummary)
	}

	// Analytics dashboards (charts, graphs, tables, KPI, heatmap, export)
	analyticsHandler := handlers.NewAnalyticsHandler()
	analyticsAPI := router.Group("/api/v1/analytics")
	{
		analyticsAPI.GET("/dashboards", analyticsHandler.ListDashboards)
		analyticsAPI.GET("/dashboards/:id", analyticsHandler.GetDashboard)
		analyticsAPI.PUT("/dashboards/:id/widgets/:widgetId", analyticsHandler.UpdateWidget)
		analyticsAPI.PUT("/dashboards/:id/layout", analyticsHandler.ReorderWidgets)
		analyticsAPI.GET("/dashboards/:id/widgets/:widgetId/export", analyticsHandler.ExportCSV)
		analyticsAPI.GET("/widget-types", analyticsHandler.GetWidgetTypes)
	}

	// ====================================
	// CDC & ETL DATA PLATFORM ENDPOINTS
	// ====================================
	cdcEtlHandler := handlers.NewCDCETLHandler()

	// ETL Pipeline Management
	etlAPI := router.Group("/api/v1/etl")
	{
		etlAPI.GET("/pipelines", cdcEtlHandler.ListETLPipelines)
		etlAPI.GET("/pipelines/:id", cdcEtlHandler.GetETLPipeline)
		etlAPI.POST("/pipelines", cdcEtlHandler.CreateETLPipeline)
		etlAPI.PUT("/pipelines/:id", cdcEtlHandler.UpdateETLPipeline)
		etlAPI.DELETE("/pipelines/:id", cdcEtlHandler.DeleteETLPipeline)
		etlAPI.POST("/pipelines/:id/run", cdcEtlHandler.RunETLPipeline)
		etlAPI.GET("/runs", cdcEtlHandler.ListETLRuns)
		etlAPI.GET("/runs/:id", cdcEtlHandler.GetETLRun)
		etlAPI.POST("/connectors", cdcEtlHandler.CreateETLConnector)
		etlAPI.GET("/connectors", cdcEtlHandler.GetETLConnectors)
		etlAPI.GET("/connectors/catalog", cdcEtlHandler.GetETLConnectorCatalog)
		etlAPI.GET("/orchestration/capabilities", cdcEtlHandler.GetETLOrchestrationCapabilities)
		etlAPI.GET("/blueprints", cdcEtlHandler.GetETLBlueprints)
		etlAPI.GET("/observability", cdcEtlHandler.GetETLObservability)
	}

	// CDC Pipeline Management
	cdcAPI := router.Group("/api/v1/cdc")
	{
		cdcAPI.GET("/pipelines", cdcEtlHandler.ListCDCPipelines)
		cdcAPI.GET("/pipelines/:id", cdcEtlHandler.GetCDCPipeline)
		cdcAPI.POST("/pipelines", cdcEtlHandler.CreateCDCPipeline)
		cdcAPI.PUT("/pipelines/:id", cdcEtlHandler.UpdateCDCPipeline)
		cdcAPI.DELETE("/pipelines/:id", cdcEtlHandler.DeleteCDCPipeline)
		cdcAPI.POST("/pipelines/:id/start", cdcEtlHandler.StartCDCPipeline)
		cdcAPI.POST("/pipelines/:id/pause", cdcEtlHandler.PauseCDCPipeline)
		cdcAPI.POST("/pipelines/:id/stop", cdcEtlHandler.StopCDCPipeline)
		cdcAPI.GET("/sources", cdcEtlHandler.GetCDCSourceTypes)
		cdcAPI.GET("/sinks", cdcEtlHandler.GetCDCSinkTypes)
		cdcAPI.GET("/observability", cdcEtlHandler.GetCDCObservability)
	}

	// Data Platform Overview
	router.GET("/api/v1/data-platform/overview", cdcEtlHandler.GetPlatformOverview)

	// ====================================
	// API BUILDER, CSV DASHBOARD & CONVERSION
	// ====================================
	apiBuilderHandler := handlers.NewAPIBuilderHandler(analyticsHandler, gisHandler)

	builderAPI := router.Group("/api/v1/builder")
	{
		// Summary
		builderAPI.GET("/summary", apiBuilderHandler.GetSummary)

		// Custom API CRUD
		builderAPI.GET("/apis", apiBuilderHandler.ListAPIs)
		builderAPI.GET("/apis/:id", apiBuilderHandler.GetAPI)
		builderAPI.POST("/apis", apiBuilderHandler.CreateAPI)
		builderAPI.PUT("/apis/:id", apiBuilderHandler.UpdateAPI)
		builderAPI.DELETE("/apis/:id", apiBuilderHandler.DeleteAPI)
		builderAPI.POST("/apis/:id/test", apiBuilderHandler.TestAPI)

		// CSV Upload & Dashboard Generation
		builderAPI.POST("/csv/upload", apiBuilderHandler.UploadCSV)
		builderAPI.GET("/csv/uploads", apiBuilderHandler.ListCSVUploads)
		builderAPI.GET("/csv/uploads/:id", apiBuilderHandler.GetCSVUpload)
		builderAPI.DELETE("/csv/uploads/:id", apiBuilderHandler.DeleteCSVUpload)
		builderAPI.POST("/csv/uploads/:id/generate-dashboard", apiBuilderHandler.GenerateDashboard)
		builderAPI.POST("/csv/uploads/:id/generate-gis", apiBuilderHandler.GenerateGISFromCSV)

		// Dashboard <-> GIS Conversion
		builderAPI.POST("/convert/analyze", apiBuilderHandler.AnalyzeConversion)
		builderAPI.POST("/convert/dashboard-to-gis", apiBuilderHandler.ConvertDashboardToGIS)
		builderAPI.POST("/convert/gis-to-dashboard", apiBuilderHandler.ConvertGISToDashboard)
		builderAPI.GET("/conversions", apiBuilderHandler.ListConversions)

		// File Scanner (SafeGate Pipeline)
		builderAPI.POST("/scanner/scan", apiBuilderHandler.ScanFile)
		builderAPI.GET("/scanner/scans", apiBuilderHandler.ListScans)
		builderAPI.GET("/scanner/health", apiBuilderHandler.GetScannerHealth)

		// Dashboard Deletion
		builderAPI.DELETE("/dashboards/:id", apiBuilderHandler.DeleteDashboard)
	}

	// ====================================
	// NETWORK INTELLIGENCE ENDPOINTS
	// ====================================
	netIntelHandler := handlers.NewNetIntelHandler()

	netintelAPI := router.Group("/api/v1/netintel")
	{
		// Summary & Observability
		netintelAPI.GET("/summary", netIntelHandler.GetSummary)
		netintelAPI.GET("/observability", netIntelHandler.GetObservability)
		netintelAPI.GET("/log-types", netIntelHandler.GetLogTypes)

		// Parser CRUD
		netintelAPI.GET("/parsers", netIntelHandler.ListParsers)
		netintelAPI.GET("/parsers/:id", netIntelHandler.GetParser)
		netintelAPI.POST("/parsers", netIntelHandler.CreateParser)
		netintelAPI.PUT("/parsers/:id", netIntelHandler.UpdateParser)
		netintelAPI.DELETE("/parsers/:id", netIntelHandler.DeleteParser)

		// Log Entries
		netintelAPI.GET("/logs", netIntelHandler.ListEntries)
		netintelAPI.POST("/logs", netIntelHandler.IngestLog)
		netintelAPI.GET("/logs/stats", netIntelHandler.GetEntryStats)

		// Topology
		netintelAPI.GET("/topology", netIntelHandler.GetTopology)
		netintelAPI.GET("/topology/nodes/:id", netIntelHandler.GetTopologyNode)
		netintelAPI.PUT("/topology/nodes/:id", netIntelHandler.UpdateTopologyNode)

		// Heatmaps & Trends
		netintelAPI.GET("/heatmap", netIntelHandler.GetHeatmap)
		netintelAPI.GET("/trends", netIntelHandler.GetTrends)

		// Predictions & Tracks
		netintelAPI.GET("/predictions", netIntelHandler.GetPredictions)
		netintelAPI.GET("/tracks", netIntelHandler.ListTracks)
		netintelAPI.GET("/tracks/:mac", netIntelHandler.GetTrack)

		// Anomalies
		netintelAPI.GET("/anomalies", netIntelHandler.ListAnomalies)
		netintelAPI.POST("/anomalies/:id/acknowledge", netIntelHandler.AcknowledgeAnomaly)
		netintelAPI.POST("/anomalies/:id/resolve", netIntelHandler.ResolveAnomaly)

		// Alerts
		netintelAPI.GET("/alerts", netIntelHandler.ListAlerts)
		netintelAPI.POST("/alerts/:id/acknowledge", netIntelHandler.AcknowledgeAlert)
		netintelAPI.POST("/alerts/:id/resolve", netIntelHandler.ResolveAlert)

		// Forecasts
		netintelAPI.GET("/forecasts", netIntelHandler.ListForecasts)
		netintelAPI.GET("/forecasts/:metric", netIntelHandler.GetForecast)
	}

	apiPort := cfg.API.Port
	apiHost := cfg.API.Host

	fmt.Printf("📡 API Server running on http://%s:%s\n", apiHost, apiPort)
	fmt.Println("\n🔐 RBAC Security Model:")
	fmt.Println("  ✅ READ  operations (GET)     - Allowed for all authenticated users")
	fmt.Println("  ❌ WRITE operations (POST/PUT/DELETE) - Allowed ONLY for users with 'admin' role\n")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /health                  - Health check (no auth)")
	fmt.Println("  GET  /status                  - Check all connections (no auth)")
	fmt.Println()
	fmt.Println("MySQL endpoints:")
	fmt.Println("  GET  /api/mysql/users         - List users (authenticated users)")
	fmt.Println("  GET  /api/mysql/users/:id     - Get user (authenticated users)")
	fmt.Println("  POST /api/mysql/users         - Create user (admin only)")
	fmt.Println("  PUT  /api/mysql/users/:id     - Update user (admin only)")
	fmt.Println("  DELETE /api/mysql/users/:id   - Delete user (admin only)")
	fmt.Println()
	fmt.Println("MariaDB endpoints:")
	fmt.Println("  GET  /api/mariadb/users       - List users (authenticated users)")
	fmt.Println("  GET  /api/mariadb/users/:id   - Get user (authenticated users)")
	fmt.Println("  POST /api/mariadb/users       - Create user (admin only)")
	fmt.Println("  PUT  /api/mariadb/users/:id   - Update user (admin only)")
	fmt.Println("  DELETE /api/mariadb/users/:id - Delete user (admin only)")
	fmt.Println()
	fmt.Println("PostgreSQL endpoints:")
	fmt.Println("  GET  /api/postgres/users      - List users (authenticated users)")
	fmt.Println("  GET  /api/postgres/users/:id  - Get user (authenticated users)")
	fmt.Println("  POST /api/postgres/users      - Create user (admin only)")
	fmt.Println("  PUT  /api/postgres/users/:id  - Update user (admin only)")
	fmt.Println("  DELETE /api/postgres/users/:id - Delete user (admin only)")
	fmt.Println()
	fmt.Println("Percona endpoints:")
	fmt.Println("  GET  /api/percona/users       - List users (authenticated users)")
	fmt.Println("  GET  /api/percona/users/:id   - Get user (authenticated users)")
	fmt.Println("  POST /api/percona/users       - Create user (admin only)")
	fmt.Println("  PUT  /api/percona/users/:id   - Update user (admin only)")
	fmt.Println("  DELETE /api/percona/users/:id - Delete user (admin only)")
	fmt.Println()
	fmt.Println("MongoDB endpoints:")
	fmt.Println("  GET  /api/mongodb/users       - List users (authenticated users)")
	fmt.Println("  GET  /api/mongodb/users/:id   - Get user (authenticated users)")
	fmt.Println("  POST /api/mongodb/users       - Create user (admin only)")
	fmt.Println("  PUT  /api/mongodb/users/:id   - Update user (admin only)")
	fmt.Println("  DELETE /api/mongodb/users/:id - Delete user (admin only)")
	fmt.Println()
	fmt.Println("Firebase endpoints:")
	fmt.Println("  GET  /api/firebase/users      - List users (authenticated users)")
	fmt.Println("  GET  /api/firebase/users/:id  - Get user (authenticated users)")
	fmt.Println("  POST /api/firebase/users      - Create user (admin only)")
	fmt.Println("  PUT  /api/firebase/users/:id  - Update user (admin only)")
	fmt.Println("  DELETE /api/firebase/users/:id - Delete user (admin only)")
	fmt.Println()
	fmt.Println("Oracle endpoints:")
	fmt.Println("  GET  /api/oracle/users        - List users (authenticated users)")
	fmt.Println("  GET  /api/oracle/users/:id    - Get user (authenticated users)")
	fmt.Println("  POST /api/oracle/users        - Create user (admin only)")
	fmt.Println("  PUT  /api/oracle/users/:id    - Update user (admin only)")
	fmt.Println("  DELETE /api/oracle/users/:id  - Delete user (admin only)")
	fmt.Println()
	fmt.Println("Admin endpoints (admin role required):")
	fmt.Println("  POST /api/admin/database/create  - Create a new database")
	fmt.Println("  GET  /api/admin/database/list    - List all databases")
	fmt.Println("  GET  /api/admin/database/servers - List default and connected DB servers")
	fmt.Println("  POST /api/admin/database/connect - Connect a new DB server")
	fmt.Println("  POST /api/admin/table/create     - Create a new table")
	fmt.Println("  GET  /api/admin/table/list       - List all tables")
	fmt.Println()
	fmt.Println("User Management endpoints (admin only):")
	fmt.Println("  GET    /api/v1/users            - List all platform users")
	fmt.Println("  GET    /api/v1/users/:id        - Get a platform user")
	fmt.Println("  POST   /api/v1/users            - Create a platform user")
	fmt.Println("  PUT    /api/v1/users/:id        - Update a platform user")
	fmt.Println("  DELETE /api/v1/users/:id        - Delete a platform user")
	fmt.Println()
	fmt.Println("Dynamic Query endpoints (authenticated users):")
	fmt.Println("  GET  /api/{db}/query            - Execute SELECT queries with parameters")
	fmt.Println("       Example: /api/mysql/query?q=SELECT * FROM users&params=1")
	fmt.Println("  POST /api/{db}/query            - Execute any query (SELECT/INSERT/UPDATE/DELETE/CREATE)")
	fmt.Println("       Body: {\"query\": \"SQL_QUERY\", \"params\": [\"value1\", \"value2\"]}")
	fmt.Println("  POST /api/{db}/query/batch      - Execute multiple queries at once")
	fmt.Println("       Body: [{\"query\": \"SQL_QUERY\", \"params\": []}]")
	fmt.Println("  GET  /api/{db}/schema           - Get table schema")
	fmt.Println("       Example: /api/mysql/schema?table=users")
	fmt.Println("  Available databases: mysql, mariadb, postgres, percona, oracle")
	fmt.Println()
	fmt.Println("Notification endpoints (authenticated users):")
	fmt.Println("  POST /api/notifications/send    - Send custom notification to Discord")
	fmt.Println("  POST /api/notifications/health  - Send health check notification")
	fmt.Println("  POST /api/notifications/status  - Send status report notification")
	fmt.Println("  GET  /api/notifications/status  - Get notification service status (no auth)")
	fmt.Println()
	fmt.Println("Kubernetes-style API endpoints:")
	fmt.Println("  POST /api/v1/{namespace}/{kind}              - Create resource")
	fmt.Println("  GET  /api/v1/{namespace}/{kind}/{name}       - Get resource")
	fmt.Println("  PUT  /api/v1/{namespace}/{kind}/{name}       - Update resource")
	fmt.Println("  DELETE /api/v1/{namespace}/{kind}/{name}     - Delete resource")
	fmt.Println("  GET  /api/v1/{namespace}/{kind}              - List resources")
	fmt.Println("  Supported kinds: workloads, pipelines, schedules")
	fmt.Println()

	// Start runtime in background
	go func() {
		if err := rt.Start(ctx, fmt.Sprintf("%s:%s", apiHost, apiPort)); err != nil {
			log.Printf("Failed to start runtime: %v", err)
			cancel()
		}
	}()

	// Start API server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", apiHost, apiPort),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("🛑 Shutting down gracefully...")

	// Give handlers 10 seconds to finish
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Stop runtime
	if err := rt.Stop(); err != nil {
		log.Printf("Runtime stop error: %v", err)
	}

	cancel()
	log.Println("✅ AxiomNizam stopped")
}

// Create tables on all databases
func createTables(conns *database.Connections) {
	if conns.MySQL != nil {
		conns.MySQL.AutoMigrate(&models.User{})
		log.Println("✅ MySQL table created/migrated")
	}
	if conns.MariaDB != nil {
		conns.MariaDB.AutoMigrate(&models.User{})
		log.Println("✅ MariaDB table created/migrated")
	}
	if conns.PostgreSQL != nil {
		conns.PostgreSQL.AutoMigrate(&models.User{})
		log.Println("✅ PostgreSQL table created/migrated")
	}
	if conns.Percona != nil {
		conns.Percona.AutoMigrate(&models.User{})
		log.Println("✅ Percona table created/migrated")
	}
	if conns.Oracle != nil {
		conns.Oracle.AutoMigrate(&models.User{})
		log.Println("✅ Oracle table created/migrated")
	}
}
