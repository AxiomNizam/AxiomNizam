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

	// Dynamic Query handlers for each database
	mysqlDynamicHandler := handlers.NewDynamicQueryHandler(conns.MySQL, queryLogger)
	mariadbDynamicHandler := handlers.NewDynamicQueryHandler(conns.MariaDB, queryLogger)
	postgresDynamicHandler := handlers.NewDynamicQueryHandler(conns.PostgreSQL, queryLogger)
	perconaDynamicHandler := handlers.NewDynamicQueryHandler(conns.Percona, queryLogger)
	oracleDynamicHandler := handlers.NewDynamicQueryHandler(conns.Oracle, queryLogger)

	// Notification handler
	discordWebhookURL := cfg.Discord.WebhookURL
	notificationHandler := handlers.NewNotificationHandler(discordWebhookURL, dbConnections)

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
	if tokenValidator != nil {
		adminMiddleware = func(c *gin.Context) {
			authMiddleware(c)
			if !c.IsAborted() {
				auth.RequireAdmin()(c)
			}
		}
	} else {
		adminMiddleware = func(c *gin.Context) { c.Next() }
	}

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
	router.POST("/api/admin/database/create", adminMiddleware, adminHandler.CreateDatabase)
	router.GET("/api/admin/database/list", adminMiddleware, adminHandler.ListDatabases)

	// Table management endpoints (admin only)
	router.POST("/api/admin/table/create", adminMiddleware, adminHandler.CreateTable)
	router.GET("/api/admin/table/list", adminMiddleware, adminHandler.ListTables)

	// API Metrics endpoints (admin only)
	router.GET("/api/admin/metrics/all", adminMiddleware, apiMetricsTracker.GetAllAPIMetrics)
	router.GET("/api/admin/metrics/count", adminMiddleware, apiMetricsTracker.GetAPICount)
	router.GET("/api/admin/metrics/stats", adminMiddleware, apiMetricsTracker.GetAPIStats)

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
		nsAPI.POST("/:namespace/:kind", resourceHandler.CreateOrUpdate)
		nsAPI.GET("/:namespace/:kind", resourceHandler.List)
		nsAPI.GET("/:namespace/:kind/:name", resourceHandler.Get)
		nsAPI.PUT("/:namespace/:kind/:name", resourceHandler.Update)
		nsAPI.DELETE("/:namespace/:kind/:name", resourceHandler.Delete)
		nsAPI.GET("/:namespace/:kind/:name/status", resourceHandler.GetStatus)
		nsAPI.GET("/:namespace/:kind/:name/events", resourceHandler.Events)
	}

	// Non-namespaced resource endpoints: /api/v1/{kind}
	router.POST("/api/v1/apis", resourceHandler.CreateOrUpdate)
	router.GET("/api/v1/apis", resourceHandler.ListAll)
	router.POST("/api/v1/policies", resourceHandler.CreateOrUpdate)
	router.GET("/api/v1/policies", resourceHandler.ListAll)
	router.POST("/api/v1/workflows", resourceHandler.CreateOrUpdate)
	router.GET("/api/v1/workflows", resourceHandler.ListAll)
	router.POST("/api/v1/workflows/:name/run", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": fmt.Sprintf("Workflow '%s' started", c.Param("name")), "status": "Running"})
	})

	// DataSource endpoints
	dsHandler := handlers.NewDataSourceHandler()
	router.POST("/api/v1/datasources", dsHandler.Create)
	router.GET("/api/v1/datasources", dsHandler.List)
	router.GET("/api/v1/datasources/:name", dsHandler.Get)
	router.PUT("/api/v1/datasources/:name", dsHandler.Update)
	router.DELETE("/api/v1/datasources/:name", dsHandler.Delete)
	router.POST("/api/v1/datasources/:name/test", dsHandler.Test)

	// Job endpoints
	jobHandler := handlers.NewJobHandler()
	router.POST("/api/v1/jobs", jobHandler.Create)
	router.GET("/api/v1/jobs", jobHandler.List)
	router.GET("/api/v1/jobs/:id", jobHandler.Get)
	router.POST("/api/v1/jobs/:id/run", jobHandler.Run)
	router.GET("/api/v1/jobs/:id/logs", jobHandler.GetLogs)
	router.POST("/api/v1/jobs/:id/cancel", jobHandler.Cancel)
	router.DELETE("/api/v1/jobs/:id", jobHandler.Delete)

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
	fmt.Println("  POST /api/admin/database/create - Create a new database")
	fmt.Println("  GET  /api/admin/database/list   - List all databases")
	fmt.Println("  POST /api/admin/table/create    - Create a new table")
	fmt.Println("  GET  /api/admin/table/list      - List all tables")
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
