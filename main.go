package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"example.com/axiomnizam/internal/auth"
	"example.com/axiomnizam/internal/bootstrapsecrets"
	"example.com/axiomnizam/internal/bulk"
	"example.com/axiomnizam/internal/conductor"
	"example.com/axiomnizam/internal/config"
	"example.com/axiomnizam/internal/database"
	"example.com/axiomnizam/internal/eventbus"
	exportpkg "example.com/axiomnizam/internal/export"
	"example.com/axiomnizam/internal/handlers"
	iampkg "example.com/axiomnizam/internal/iam"
	"example.com/axiomnizam/internal/integration"
	"example.com/axiomnizam/internal/kubeplus/admission"
	"example.com/axiomnizam/internal/kubeplus/crd"
	"example.com/axiomnizam/internal/kubeplus/scheduler"
	"example.com/axiomnizam/internal/lineage"
	"example.com/axiomnizam/internal/models"
	"example.com/axiomnizam/internal/netintel/modes"
	"example.com/axiomnizam/internal/platform"
	"example.com/axiomnizam/internal/rbac"
	"example.com/axiomnizam/internal/reviewflow"
	"example.com/axiomnizam/internal/runtime"
	"example.com/axiomnizam/internal/streaming"
	"example.com/axiomnizam/internal/tenant"
	"example.com/axiomnizam/internal/tracing"
	"example.com/axiomnizam/internal/vectorplus"
	"example.com/axiomnizam/internal/versioning"
	"example.com/axiomnizam/internal/webhooks"
	"example.com/axiomnizam/internal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}
	fmt.Println("🚀 Starting AxiomNizam with Kubernetes-style Runtime...")
	fmt.Println()

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
	applySecurityGuardrails(cfg)
	iamOnlyAuthRaw := strings.TrimSpace(os.Getenv("IAM_ONLY_AUTH"))
	iamOnlyAuth := true
	if iamOnlyAuthRaw != "" {
		iamOnlyAuth = strings.EqualFold(iamOnlyAuthRaw, "true")
	}

	// Initialize IAM token validator
	iamIssuerURL := strings.TrimSpace(os.Getenv("IAM_ISSUER_URL"))
	if iamIssuerURL == "" {
		iamIssuerURL = cfg.GetIAMURL()
	}
	normalizeBaseURL := func(raw string) string {
		candidate := strings.TrimSpace(raw)
		if candidate == "" {
			return ""
		}

		if parsed, parseErr := url.Parse(candidate); parseErr == nil && parsed.Scheme != "" && parsed.Host != "" {
			return strings.TrimRight(parsed.Scheme+"://"+parsed.Host+parsed.Path, "/")
		}

		return strings.TrimRight(candidate, "/")
	}

	validatorJWKSBases := make([]string, 0, 4)
	seenValidatorBase := make(map[string]struct{}, 4)
	addValidatorJWKSBase := func(raw string) {
		base := normalizeBaseURL(raw)
		if base == "" {
			return
		}
		if _, exists := seenValidatorBase[base]; exists {
			return
		}
		seenValidatorBase[base] = struct{}{}
		validatorJWKSBases = append(validatorJWKSBases, base)
	}

	addValidatorJWKSBase(os.Getenv("IAM_INTERNAL_BASE_URL"))
	addValidatorJWKSBase(iamIssuerURL)
	addValidatorJWKSBase(cfg.GetIAMURL())
	addValidatorJWKSBase("http://localhost:8000")

	buildValidatorConfig := func(jwksBase string) *auth.TokenValidatorConfig {
		validatorConfig := &auth.TokenValidatorConfig{
			IssuerURL: iamIssuerURL,
		}
		if jwksBase != "" {
			validatorConfig.JWKSURL = strings.TrimRight(jwksBase, "/") + "/.well-known/jwks.json"
		}
		return validatorConfig
	}

	initializeTokenValidator := func() (*auth.TokenValidator, error) {
		if len(validatorJWKSBases) == 0 {
			return nil, fmt.Errorf("no IAM JWKS base URLs configured")
		}

		initErrors := make([]string, 0, len(validatorJWKSBases))
		for _, jwksBase := range validatorJWKSBases {
			candidateConfig := buildValidatorConfig(jwksBase)
			initializedValidator, initErr := auth.NewTokenValidator(candidateConfig)
			if initErr == nil {
				log.Printf("✅ IAM token validator JWKS source: %s/.well-known/jwks.json", strings.TrimRight(jwksBase, "/"))
				return initializedValidator, nil
			}
			initErrors = append(initErrors, fmt.Sprintf("%s: %v", jwksBase, initErr))
		}

		return nil, fmt.Errorf("all IAM JWKS endpoints failed: %s", strings.Join(initErrors, " | "))
	}

	var tokenValidatorMu sync.RWMutex
	tokenValidator, err := initializeTokenValidator()
	if err != nil {
		log.Printf("⚠️  IAM token validator initialization failed at startup: %v", err)
		log.Printf("⚠️  Auth-protected APIs will return 503 until IAM JWKS becomes reachable")
		tokenValidator = nil
	}

	getOrInitTokenValidator := func() *auth.TokenValidator {
		tokenValidatorMu.RLock()
		tv := tokenValidator
		tokenValidatorMu.RUnlock()
		if tv != nil {
			return tv
		}

		tokenValidatorMu.Lock()
		defer tokenValidatorMu.Unlock()
		if tokenValidator != nil {
			return tokenValidator
		}

		initializedValidator, initErr := initializeTokenValidator()
		if initErr != nil {
			log.Printf("⚠️  IAM token validator still unavailable: %v", initErr)
			return nil
		}

		tokenValidator = initializedValidator
		log.Printf("✅ IAM token validator initialized after startup")
		return tokenValidator
	}

	// Initialize all connections
	conns := database.InitConnections(cfg)
	if _, secretErr := ensureSharedDemoJWTSecret(conns.PostgreSQL, conns.Etcd); secretErr != nil {
		log.Printf("⚠️  DEMO_JWT_SECRET synchronization failed: %v", secretErr)
	} else {
		log.Println("✅ DEMO_JWT_SECRET synchronized for replica-safe token validation")
	}
	workflows.ConfigureGlobalPersistence(conns.Etcd)
	modes.ConfigureGlobalPersistence(conns.Etcd)
	vectorplus.ConfigureGlobalPersistence(conns.Etcd)
	reviewflow.ConfigureGlobalPersistence(conns.Etcd)
	integration.ConfigureGlobalPersistence(conns.Etcd)

	// Create tables
	createTables(conns)

	// ====================================
	// IAM SYSTEM INITIALIZATION
	// ====================================
	iamSystem, iamErr := iampkg.NewSystem(conns.PostgreSQL, conns.Etcd, iampkg.Config{
		IssuerURL: strings.TrimSpace(os.Getenv("IAM_ISSUER_URL")),
	})
	if iamErr != nil {
		log.Printf("⚠️  IAM system initialization failed: %v", iamErr)
		log.Println("⚠️  IAM endpoints will not be available. Ensure PostgreSQL and etcd are connected.")
	} else {
		log.Println("✅ IAM system initialized")
	}

	// Create Gin router
	router := gin.Default()

	allowedOriginSet := make(map[string]struct{})
	addAllowedOrigin := func(raw string) {
		candidate := strings.TrimSpace(raw)
		if candidate == "" {
			return
		}
		if parsed, err := url.Parse(candidate); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			candidate = parsed.Scheme + "://" + parsed.Host
		}
		allowedOriginSet[candidate] = struct{}{}
	}

	// Always include canonical frontend URL when provided.
	addAllowedOrigin(os.Getenv("PUBLIC_FRONTEND_URL"))

	for _, candidate := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
		addAllowedOrigin(candidate)
	}
	if len(allowedOriginSet) == 0 {
		addAllowedOrigin("https://axiomnizam.bitbd.net")
		addAllowedOrigin("http://localhost:7000")
		addAllowedOrigin("http://127.0.0.1:7000")
	}

	isAllowedOrigin := func(origin string) bool {
		_, ok := allowedOriginSet[origin]
		return ok
	}

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-KEY")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-RateLimit-Limit, X-RateLimit-Remaining, X-Token-Expires-At")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		if origin != "" && isAllowedOrigin(origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			if origin != "" && !isAllowedOrigin(origin) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
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

	// Admin handler for database and table creation
	// Only include SQL databases (MongoDB and Firebase don't support SQL DDL operations)
	dbConnections := map[string]*gorm.DB{
		"mysql":    conns.MySQL,
		"mariadb":  conns.MariaDB,
		"postgres": conns.PostgreSQL,
		"percona":  conns.Percona,
		"oracle":   conns.Oracle,
	}
	adminHandler := handlers.NewAdminHandler(dbConnections, conns.PostgreSQL)

	// User management handler
	platformUserHandler := handlers.NewPlatformUserHandler(conns.Etcd)

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

	// Context enrichment helper - populates database name and user info for logging
	enrichRequestContext := func(c *gin.Context) {
		// Extract database name from URL path (e.g., /api/mysql/query -> mysql)
		pathParts := strings.Split(c.Request.URL.Path, "/")
		if len(pathParts) >= 3 {
			dbName := pathParts[2]
			switch dbName {
			case "mysql", "mariadb", "postgres", "percona", "oracle":
				c.Set("database", dbName)
			}
		}

		// Extract user info from validated claims if available
		if claims := auth.GetUser(c); claims != nil && claims.Sub != "" {
			c.Set("user_id", claims.Sub)
		}
	}

	// authenticateRequest validates token + rate limits and sets auth context without advancing handlers.
	authenticateRequest := func(c *gin.Context) bool {
		activeTokenValidator := getOrInitTokenValidator()
		if activeTokenValidator == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "authentication unavailable",
				"message": "token validation is not available because IAM token validator initialization failed",
			})
			c.Abort()
			return false
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return false
		}

		token, err := auth.ExtractBearerToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid authorization header: %v", err)})
			c.Abort()
			return false
		}

		claims, err := activeTokenValidator.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid token: %v", err)})
			c.Abort()
			return false
		}

		if claims != nil && len(claims.RolesList()) == 0 {
			configuredSysadminEmail := strings.ToLower(strings.TrimSpace(os.Getenv("IAM_SYSADMIN_EMAIL")))
			claimEmail := strings.ToLower(strings.TrimSpace(claims.Email))
			if configuredSysadminEmail != "" && claimEmail != "" && claimEmail == configuredSysadminEmail {
				fallbackRoles := []string{"sysadmin", "system-manager", "admin"}
				claims.Roles = append([]string{}, fallbackRoles...)
				claims.RealmAccess.Roles = append([]string{}, fallbackRoles...)
				log.Printf("⚠️  Applied bootstrap sysadmin role fallback for token subject %s", claimEmail)
			}
		}

		principal := strings.TrimSpace(claims.PreferredUsername)
		if principal == "" {
			principal = strings.TrimSpace(claims.Email)
		}
		if principal == "" {
			principal = strings.TrimSpace(claims.Sub)
		}
		if principal == "" {
			principal = "token-user"
		}

		defaultMaxCalls, defaultValidity := rateLimiter.DefaultPolicy()
		callsLimit := defaultMaxCalls

		allowed, callsRemaining, expiresAt, limitErr := rateLimiter.CheckRateLimit(token)
		if !allowed && limitErr != nil && limitErr.Error() == "token not tracked or invalid" {
			// Accept valid IAM/JWT tokens even if they were not issued through /auth/login.
			policyCalls := defaultMaxCalls
			policyValidity := defaultValidity

			if claims != nil && strings.TrimSpace(claims.ClientID) != "" && iamSystem != nil && iamSystem.Clients != nil {
				if clientCfg, clientErr := iamSystem.Clients.GetClient(strings.TrimSpace(claims.ClientID)); clientErr == nil && clientCfg != nil {
					if clientCfg.RateLimitMaxCalls > 0 {
						policyCalls = clientCfg.RateLimitMaxCalls
					}
					if clientCfg.TokenValidityMinutes > 0 {
						policyValidity = time.Duration(clientCfg.TokenValidityMinutes) * time.Minute
					}
				}
			}

			rateLimiter.RegisterTokenWithPolicy(token, principal, policyCalls, policyValidity)
			callsLimit = policyCalls
			allowed, callsRemaining, expiresAt, limitErr = rateLimiter.CheckRateLimit(token)
		}

		if trackedLimit, _, tracked := rateLimiter.GetTokenPolicy(token); tracked && trackedLimit > 0 {
			callsLimit = trackedLimit
		}

		if !allowed {
			if limitErr != nil && limitErr.Error() == "token expired" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":      "token expired",
					"message":    "your token is no longer valid. please login again to get a new token",
					"expired_at": expiresAt.Format("2006-01-02 15:04:05"),
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":           "api call limit exceeded",
					"message":         fmt.Sprintf("you have used all %d api calls allowed for this token", callsLimit),
					"calls_limit":     callsLimit,
					"expires_at":      expiresAt.Format("2006-01-02 15:04:05"),
					"action_required": fmt.Sprintf("use a new token to continue with a fresh %d-call quota", callsLimit),
					"action_endpoint": "/auth/login",
				})
			}
			c.Abort()
			return false
		}

		if err := rateLimiter.IncrementCallCount(token); err != nil {
			log.Printf("⚠️  Failed to increment call count: %v", err)
		}

		c.Set("user", claims)
		c.Set("username", principal)
		c.Set("email", claims.Email)
		c.Set("roles", claims.RolesList())
		c.Set("calls_remaining", callsRemaining)
		c.Set("token_expires_at", expiresAt.Format("2006-01-02 15:04:05"))
		c.Set("token", token)

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", callsLimit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", callsRemaining))
		c.Header("X-Token-Expires-At", expiresAt.Format("2006-01-02 15:04:05"))

		log.Printf("✅ Token validated & rate limit OK for user: %s (calls remaining: %d)", principal, callsRemaining)
		return true
	}

	// Apply auth middleware to protected routes.
	authMiddleware := func(c *gin.Context) {
		if !authenticateRequest(c) {
			return
		}
		enrichRequestContext(c)
		c.Next()
	}

	// Health check endpoints (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/status", healthHandler.Status)
	router.GET("/distributed", healthHandler.Distributed)

	// Authentication endpoints (no auth required for login/refresh)
	authHandler := handlers.NewAuthHandler()
	authHandler.SetRateLimiter(rateLimiter)
	authHandler.SetPlatformUserHandler(platformUserHandler)
	if iamSystem != nil && iamSystem.PGStore != nil {
		authHandler.SetIdentityProviderStore(iamSystem.PGStore)
	}
	if iamSystem != nil && iamSystem.Users != nil {
		authHandler.SetIAMUserRepository(iamSystem.Users)
	}
	if iamSystem != nil && iamSystem.Authorizer != nil {
		authHandler.SetIAMAuthorizer(iamSystem.Authorizer)
	}
	router.POST("/auth/login", authHandler.Login)
	router.POST("/auth/refresh", authHandler.RefreshToken)
	router.GET("/auth/validate", authHandler.ValidateToken)
	router.GET("/auth/oauth/start", authHandler.OAuthStart)
	router.GET("/auth/oauth/callback", authHandler.OAuthCallback)

	// Protected auth endpoints (auth required)
	router.POST("/auth/logout", authMiddleware, authHandler.Logout)
	router.GET("/auth/token-status", authMiddleware, authHandler.GetTokenStatus)
	router.GET("/auth/admin/tokens-status", authMiddleware, auth.RequireAdmin(), authHandler.GetAllTokensStatus)

	// Get admin middleware (requires admin role)
	var adminMiddleware gin.HandlerFunc
	var adminOrSysMiddleware gin.HandlerFunc
	adminMiddleware = func(c *gin.Context) {
		if !authenticateRequest(c) {
			return
		}
		enrichRequestContext(c)
		claims := auth.GetUser(c)
		if claims == nil || !claims.HasRole("admin") {
			roles := []string{}
			if claims != nil {
				roles = claims.RolesList()
			}
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "forbidden: user does not have 'admin' role",
				"user_roles": roles,
				"required":   "admin",
			})
			c.Abort()
			return
		}
		c.Next()
	}
	adminOrSysMiddleware = func(c *gin.Context) {
		if !authenticateRequest(c) {
			return
		}
		enrichRequestContext(c)
		claims := auth.GetUser(c)
		if claims == nil || !(claims.HasRole("admin") || claims.HasRole("system-manager") || claims.HasRole("sysadmin") || claims.HasRole("system_admin") || claims.HasRole("system-admin")) {
			roles := []string{}
			if claims != nil {
				roles = claims.RolesList()
			}
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "forbidden: user must have one of roles [admin system-manager sysadmin system_admin system-admin]",
				"user_roles": roles,
				"required":   []string{"admin", "system-manager", "sysadmin", "system_admin", "system-admin"},
			})
			c.Abort()
			return
		}
		c.Next()
	}

	// GraphQL endpoints (auth required)
	router.POST("/api/graphql", authMiddleware, graphQLHandler.Query)
	router.GET("/api/graphql/schema", authMiddleware, graphQLHandler.GetSchema)
	router.GET("/api/graphql/playground", authMiddleware, graphQLHandler.Playground)

	// ====================================
	// DYNAMIC QUERY ENDPOINTS (Auth Required)
	// ====================================
	// These endpoints allow dynamic SQL queries via Postman or any HTTP client
	// GET requests only support SELECT queries
	// POST requests are restricted to admin/system-manager roles.

	// MySQL Dynamic Queries
	router.GET("/api/mysql/query", authMiddleware, mysqlDynamicHandler.DynamicQuery)
	router.POST("/api/mysql/query", adminOrSysMiddleware, mysqlDynamicHandler.DynamicQueryWithBody)
	router.POST("/api/mysql/query/batch", adminOrSysMiddleware, mysqlDynamicHandler.BatchQueries)
	router.GET("/api/mysql/schema", authMiddleware, mysqlDynamicHandler.TableSchema)

	// MariaDB Dynamic Queries
	router.GET("/api/mariadb/query", authMiddleware, mariadbDynamicHandler.DynamicQuery)
	router.POST("/api/mariadb/query", adminOrSysMiddleware, mariadbDynamicHandler.DynamicQueryWithBody)
	router.POST("/api/mariadb/query/batch", adminOrSysMiddleware, mariadbDynamicHandler.BatchQueries)
	router.GET("/api/mariadb/schema", authMiddleware, mariadbDynamicHandler.TableSchema)

	// PostgreSQL Dynamic Queries
	router.GET("/api/postgres/query", authMiddleware, postgresDynamicHandler.DynamicQuery)
	router.POST("/api/postgres/query", adminOrSysMiddleware, postgresDynamicHandler.DynamicQueryWithBody)
	router.POST("/api/postgres/query/batch", adminOrSysMiddleware, postgresDynamicHandler.BatchQueries)
	router.GET("/api/postgres/schema", authMiddleware, postgresDynamicHandler.TableSchema)

	// Percona Dynamic Queries
	router.GET("/api/percona/query", authMiddleware, perconaDynamicHandler.DynamicQuery)
	router.POST("/api/percona/query", adminOrSysMiddleware, perconaDynamicHandler.DynamicQueryWithBody)
	router.POST("/api/percona/query/batch", adminOrSysMiddleware, perconaDynamicHandler.BatchQueries)
	router.GET("/api/percona/schema", authMiddleware, perconaDynamicHandler.TableSchema)

	// Oracle Dynamic Queries
	router.GET("/api/oracle/query", authMiddleware, oracleDynamicHandler.DynamicQuery)
	router.POST("/api/oracle/query", adminOrSysMiddleware, oracleDynamicHandler.DynamicQueryWithBody)
	router.POST("/api/oracle/query/batch", adminOrSysMiddleware, oracleDynamicHandler.BatchQueries)
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
	certificateHandler := handlers.NewCertificateHandler()

	// Database management endpoints (admin only)
	router.POST("/api/admin/database/create", adminOrSysMiddleware, adminHandler.CreateDatabase)
	router.GET("/api/admin/database/list", adminOrSysMiddleware, adminHandler.ListDatabases)
	router.GET("/api/admin/database/servers", adminOrSysMiddleware, adminHandler.ListDatabaseServers)
	router.POST("/api/admin/database/connect", adminOrSysMiddleware, adminHandler.ConnectDatabaseServer)
	router.PUT("/api/admin/database/servers/:key", adminOrSysMiddleware, adminHandler.UpdateDatabaseServer)
	router.DELETE("/api/admin/database/servers/:key", adminOrSysMiddleware, adminHandler.DeleteDatabaseServer)
	router.GET("/api/admin/certificates/status", adminOrSysMiddleware, certificateHandler.GetCertificateStatus)
	router.POST("/api/admin/certificates/renew", adminOrSysMiddleware, certificateHandler.RenewCertificate)

	// Table management endpoints (admin only)
	router.POST("/api/admin/table/create", adminOrSysMiddleware, adminHandler.CreateTable)
	router.GET("/api/admin/table/list", adminOrSysMiddleware, adminHandler.ListTables)

	// Legacy platform user management endpoints (admin only)
	if !iamOnlyAuth {
		router.GET("/api/v1/users", adminOrSysMiddleware, platformUserHandler.ListPlatformUsers)
		router.GET("/api/v1/users/:id", adminOrSysMiddleware, platformUserHandler.GetPlatformUser)
		router.POST("/api/v1/users", adminOrSysMiddleware, platformUserHandler.CreatePlatformUser)
		router.PUT("/api/v1/users/:id", adminOrSysMiddleware, platformUserHandler.UpdatePlatformUser)
		router.DELETE("/api/v1/users/:id", adminOrSysMiddleware, platformUserHandler.DeletePlatformUser)
	} else {
		log.Println("ℹ️  IAM_ONLY_AUTH=true: legacy /api/v1/users endpoints are disabled; use /iam/admin/users")
	}

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

	// Backward-compatible notification aliases restored under /api/v1.
	router.POST("/api/v1/notifications/send", authMiddleware, notificationHandler.SendNotification)
	router.POST("/api/v1/notifications/health", authMiddleware, notificationHandler.SendHealthNotification)
	router.POST("/api/v1/notifications/status", authMiddleware, notificationHandler.SendStatusNotification)
	router.GET("/api/v1/notifications/status", notificationHandler.GetNotificationStatus)

	// ====================================
	// CLI AUTHENTICATION ENDPOINTS
	// ====================================
	cliAuth := handlers.NewCLIAuthHandler()
	router.POST("/api/v1/auth/login", cliAuth.Login)
	router.POST("/api/v1/auth/logout", authHandler.Logout)
	router.GET("/api/v1/auth/verify", cliAuth.Verify)
	router.GET("/api/v1/auth/whoami", cliAuth.WhoAmI)

	// ====================================
	// KUBERNETES-STYLE RESOURCE ENDPOINTS
	// ====================================
	resourceHandler := handlers.NewResourceHandler(conns.Etcd)

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
		workflowName := strings.TrimSpace(c.Param("name"))
		if workflowName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workflow name is required"})
			return
		}

		var req struct {
			TriggerContext map[string]interface{} `json:"triggerContext"`
		}
		if c.Request.ContentLength > 0 {
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		if err := ensureWorkflowRegistered(c.Request.Context(), resourceHandler, workflowName); err != nil {
			if workflows.GlobalWorkflowEngine.GetWorkflow(workflowName) == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			log.Printf("⚠️  workflow run using previously-registered definition for %s: %v", workflowName, err)
		}

		triggerContext := req.TriggerContext
		if triggerContext == nil {
			triggerContext = make(map[string]interface{})
		}
		if username := strings.TrimSpace(auth.GetUsername(c)); username != "" {
			if _, exists := triggerContext["requestedBy"]; !exists {
				triggerContext["requestedBy"] = username
			}
		}
		if _, exists := triggerContext["triggeredAt"]; !exists {
			triggerContext["triggeredAt"] = time.Now().UTC().Format(time.RFC3339)
		}

		execution, err := workflows.Execute(c.Request.Context(), workflowName, triggerContext)
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			switch {
			case strings.Contains(errMsg, "not found"):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			case strings.Contains(errMsg, "disabled"):
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   fmt.Sprintf("Workflow '%s' executed", workflowName),
			"execution": execution,
			"status":    execution.Status,
		})
	})
	router.GET("/api/v1/workflows/:name/executions", authMiddleware, func(c *gin.Context) {
		workflowName := strings.TrimSpace(c.Param("name"))
		if workflowName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workflow name is required"})
			return
		}

		executions := workflows.GlobalWorkflowEngine.ListExecutions(workflowName)
		c.JSON(http.StatusOK, gin.H{
			"workflow":   workflowName,
			"executions": executions,
			"count":      len(executions),
		})
	})
	router.GET("/api/v1/workflows/executions/:id", authMiddleware, func(c *gin.Context) {
		executionID := strings.TrimSpace(c.Param("id"))
		if executionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "execution id is required"})
			return
		}

		execution := workflows.GlobalWorkflowEngine.GetExecution(executionID)
		if execution == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			return
		}

		c.JSON(http.StatusOK, execution)
	})

	// DataSource endpoints
	dsHandler := handlers.NewDataSourceHandler(conns.Etcd)
	router.POST("/api/v1/datasources", adminOrSysMiddleware, dsHandler.Create)
	router.GET("/api/v1/datasources", authMiddleware, dsHandler.List)
	router.GET("/api/v1/datasources/:name", authMiddleware, dsHandler.Get)
	router.PUT("/api/v1/datasources/:name", adminOrSysMiddleware, dsHandler.Update)
	router.DELETE("/api/v1/datasources/:name", adminOrSysMiddleware, dsHandler.Delete)
	router.POST("/api/v1/datasources/:name/test", adminOrSysMiddleware, dsHandler.Test)

	// Job endpoints
	jobHandler := handlers.NewJobHandler(conns.Etcd)
	router.POST("/api/v1/jobs", adminOrSysMiddleware, jobHandler.Create)
	router.GET("/api/v1/jobs", authMiddleware, jobHandler.List)
	router.GET("/api/v1/jobs/schedules", authMiddleware, jobHandler.ListSchedules)
	router.GET("/api/v1/jobs/:id", authMiddleware, jobHandler.Get)
	router.POST("/api/v1/jobs/:id/schedule", adminOrSysMiddleware, jobHandler.SetSchedule)
	router.DELETE("/api/v1/jobs/:id/schedule", adminOrSysMiddleware, jobHandler.RemoveSchedule)
	router.POST("/api/v1/jobs/:id/run", adminOrSysMiddleware, jobHandler.Run)
	router.GET("/api/v1/jobs/:id/logs", authMiddleware, jobHandler.GetLogs)
	router.POST("/api/v1/jobs/:id/cancel", adminOrSysMiddleware, jobHandler.Cancel)
	router.DELETE("/api/v1/jobs/:id", adminOrSysMiddleware, jobHandler.Delete)

	// ====================================
	// PLATFORM FEATURE APIs (PHASE 1)
	// ====================================
	platformManagers, err := platform.NewManagers(conns)
	if err != nil {
		log.Fatalf("failed to initialize etcd-backed platform managers: %v", err)
	}

	bulkHandler := bulk.NewBulkHandler(platformManagers.Bulk)
	eventBusHandler := eventbus.NewEventBusHandler(platformManagers.EventBus)
	exportHandler := exportpkg.NewExportHandler(platformManagers.Export)
	streamHandler := streaming.NewStreamHandler(platformManagers.Stream)
	webhookHandler := webhooks.NewWebhookHandler(platformManagers.Webhook)
	tenantHandler := tenant.NewTenantHandler(platformManagers.Tenant)
	rbacHandler := rbac.NewRBACHandler(platformManagers.RBAC)
	versionHandler := versioning.NewVersionHandler(platformManagers.Version)
	lineageHandler := lineage.NewLineageHandler(platformManagers.Lineage)
	tracingHandler := tracing.NewTracingHandler(platformManagers.Tracing)

	// Bulk operations
	bulkAPI := router.Group("/api/v1/bulk/operations", authMiddleware)
	{
		bulkAPI.POST("", adminOrSysMiddleware, bulkHandler.SubmitBulkOperation)
		bulkAPI.GET("", bulkHandler.ListOperations)
		bulkAPI.GET("/:id", bulkHandler.GetOperation)
		bulkAPI.GET("/:id/progress", bulkHandler.GetProgress)
		bulkAPI.DELETE("/:id", adminOrSysMiddleware, bulkHandler.CancelOperation)
		bulkAPI.POST("/:id/retry-failed", adminOrSysMiddleware, bulkHandler.RetryFailed)
		bulkAPI.GET("/:id/results", bulkHandler.GetResults)
	}

	// Event bus
	eventBusAPI := router.Group("/api/v1/eventbus", authMiddleware)
	{
		eventBusAPI.POST("/events/publish", adminOrSysMiddleware, eventBusHandler.PublishEvent)
		eventBusAPI.GET("/events", eventBusHandler.ListEvents)
		eventBusAPI.POST("/events/:id/ack", adminOrSysMiddleware, eventBusHandler.AckEvent)
		eventBusAPI.POST("/topics", adminOrSysMiddleware, eventBusHandler.CreateTopic)
		eventBusAPI.GET("/topics", eventBusHandler.ListTopics)
		eventBusAPI.POST("/subscriptions", adminOrSysMiddleware, eventBusHandler.CreateSubscription)
		eventBusAPI.GET("/subscriptions/:id", eventBusHandler.GetSubscription)
		eventBusAPI.GET("/subscriptions", eventBusHandler.ListSubscriptions)
		eventBusAPI.GET("/dlq", eventBusHandler.ListDLQ)
		eventBusAPI.POST("/dlq/:id/replay", adminOrSysMiddleware, eventBusHandler.ReplayDLQEvent)
	}

	// Exports
	exportAPI := router.Group("/api/v1/exports", authMiddleware)
	{
		exportAPI.POST("", adminOrSysMiddleware, exportHandler.SubmitExport)
		exportAPI.GET("", exportHandler.ListExports)
		exportAPI.GET("/:id", exportHandler.GetExport)
		exportAPI.GET("/:id/progress", exportHandler.GetExportProgress)
		exportAPI.GET("/:id/download", exportHandler.DownloadExport)
		exportAPI.DELETE("/:id", adminOrSysMiddleware, exportHandler.CancelExport)
	}
	router.POST("/api/v1/export-templates", authMiddleware, adminOrSysMiddleware, exportHandler.CreateTemplate)
	router.GET("/api/v1/export-templates", authMiddleware, exportHandler.ListTemplates)

	// Webhooks
	webhookAPI := router.Group("/api/v1/webhooks", authMiddleware)
	{
		webhookAPI.POST("", adminOrSysMiddleware, webhookHandler.CreateWebhook)
		webhookAPI.GET("", webhookHandler.ListWebhooks)
		webhookAPI.GET("/:id", webhookHandler.GetWebhook)
		webhookAPI.PATCH("/:id", adminOrSysMiddleware, webhookHandler.UpdateWebhook)
		webhookAPI.DELETE("/:id", adminOrSysMiddleware, webhookHandler.DeleteWebhook)
		webhookAPI.POST("/:id/test", adminOrSysMiddleware, webhookHandler.TestWebhook)
		webhookAPI.GET("/:id/deliveries", webhookHandler.GetDeliveryLogs)
	}

	// Streaming
	router.GET("/ws/stream", authMiddleware, streamHandler.HandleStream)
	streamsAPI := router.Group("/api/v1/streams", authMiddleware)
	{
		streamsAPI.POST("", adminOrSysMiddleware, streamHandler.CreateStreamRequest)
		streamsAPI.GET("", streamHandler.ListStreams)
		streamsAPI.GET("/:id", streamHandler.GetStreamStatus)
		streamsAPI.DELETE("/:id", adminOrSysMiddleware, streamHandler.CancelStream)
	}
	streamSubscriptionsAPI := router.Group("/api/v1/streaming/subscriptions", authMiddleware)
	{
		streamSubscriptionsAPI.POST("", adminOrSysMiddleware, streamHandler.Subscribe)
		streamSubscriptionsAPI.DELETE("/:id", adminOrSysMiddleware, streamHandler.Unsubscribe)
	}

	// Conductor (RabbitMQ / Kafka producer & consumer management)
	conductorCfg := conductor.Config{
		RabbitMQURL:  os.Getenv("RABBITMQ_URL"),
		KafkaBrokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
	}
	if conductorCfg.KafkaBrokers[0] == "" {
		conductorCfg.KafkaBrokers = nil
	}
	conductorMgr := conductor.NewManager(conductorCfg)
	conductor.RegisterRoutes(router, conductorMgr, authMiddleware, adminOrSysMiddleware)

	// Tenants
	tenantAPI := router.Group("/api/v1/tenants", authMiddleware)
	{
		tenantAPI.POST("", adminOrSysMiddleware, tenantHandler.CreateTenant)
		tenantAPI.GET("", tenantHandler.ListTenants)
		tenantAPI.GET("/:id", tenantHandler.GetTenant)
		tenantAPI.PATCH("/:id", adminOrSysMiddleware, tenantHandler.UpdateTenant)
		tenantAPI.DELETE("/:id", adminOrSysMiddleware, tenantHandler.DeleteTenant)
		tenantAPI.POST("/:id/members", adminOrSysMiddleware, tenantHandler.AddMember)
		tenantAPI.DELETE("/:id/members/:userId", adminOrSysMiddleware, tenantHandler.RemoveMember)
		tenantAPI.GET("/:id/quota", tenantHandler.GetQuota)
		tenantAPI.POST("/:id/quota/check", tenantHandler.CheckQuota)
	}

	// RBAC
	rbacAPI := router.Group("/api/v1/rbac", authMiddleware)
	{
		rbacAPI.POST("/roles", adminOrSysMiddleware, rbacHandler.CreateRole)
		rbacAPI.GET("/roles", rbacHandler.ListRoles)
		rbacAPI.GET("/roles/:id", rbacHandler.GetRole)
		rbacAPI.PATCH("/roles/:id", adminOrSysMiddleware, rbacHandler.UpdateRole)
		rbacAPI.DELETE("/roles/:id", adminOrSysMiddleware, rbacHandler.DeleteRole)

		rbacAPI.POST("/role-bindings", adminOrSysMiddleware, rbacHandler.BindRole)
		rbacAPI.GET("/role-bindings", rbacHandler.ListBindings)
		rbacAPI.DELETE("/role-bindings/:id", adminOrSysMiddleware, rbacHandler.DeleteBinding)

		rbacAPI.GET("/permissions", rbacHandler.ListPermissions)
		rbacAPI.POST("/permissions/check", rbacHandler.CheckPermission)

		rbacAPI.POST("/access-requests", rbacHandler.CreateAccessRequest)
		rbacAPI.GET("/access-requests", rbacHandler.ListAccessRequests)
		rbacAPI.POST("/access-requests/:id/approve", adminOrSysMiddleware, rbacHandler.ApproveAccessRequest)
		rbacAPI.POST("/access-requests/:id/reject", adminOrSysMiddleware, rbacHandler.RejectAccessRequest)
	}

	// Versioning
	versionAPI := router.Group("/api/v1/versioning", authMiddleware)
	{
		versionAPI.GET("/versions/:resourceType/:resourceId/:version", versionHandler.GetVersion)
		versionAPI.GET("/versions/:resourceType/:resourceId", versionHandler.ListVersions)
		versionAPI.GET("/history/:resourceType/:resourceId", versionHandler.GetHistory)
		versionAPI.GET("/diff/:resourceType/:resourceId", versionHandler.GetDiff)
		versionAPI.POST("/snapshots/:resourceType/:resourceId", adminOrSysMiddleware, versionHandler.CreateSnapshot)
		versionAPI.POST("/versions/:resourceType/:resourceId/rollback", adminOrSysMiddleware, versionHandler.Rollback)
	}

	// Lineage
	lineageAPI := router.Group("/api/v1/lineage", authMiddleware)
	{
		lineageAPI.GET("/nodes/:id", lineageHandler.GetNode)
		lineageAPI.GET("/nodes", lineageHandler.ListNodes)
		lineageAPI.GET("/:resourceType/:resourceId", lineageHandler.GetLineageGraph)
		lineageAPI.GET("/upstream/:resourceType/:resourceId", lineageHandler.GetUpstreamLineage)
		lineageAPI.GET("/downstream/:resourceType/:resourceId", lineageHandler.GetDownstreamLineage)
		lineageAPI.GET("/impact/:resourceType/:resourceId", lineageHandler.GetImpactAnalysis)
		lineageAPI.GET("/columns", lineageHandler.GetColumnLineage)
		lineageAPI.GET("/trace", lineageHandler.TraceDataFlow)
		lineageAPI.GET("/statistics", lineageHandler.GetStatistics)
	}

	// Tracing
	tracingAPI := router.Group("/api/v1/tracing", authMiddleware)
	{
		tracingAPI.POST("/traces", adminOrSysMiddleware, tracingHandler.IngestTrace)
		tracingAPI.GET("/traces/:traceId", tracingHandler.GetTrace)
		tracingAPI.GET("/traces/search", tracingHandler.SearchTraces)
		tracingAPI.POST("/spans", adminOrSysMiddleware, tracingHandler.IngestSpan)
		tracingAPI.GET("/spans/:spanId", tracingHandler.GetSpan)
		tracingAPI.GET("/service-map", tracingHandler.GetServiceMap)
		tracingAPI.GET("/services", tracingHandler.ListServices)
		tracingAPI.GET("/services/:service/metrics", tracingHandler.GetServiceMetrics)
		tracingAPI.GET("/services/:service/operations/:operation/metrics", tracingHandler.GetOperationMetrics)
		tracingAPI.GET("/errors/analysis", tracingHandler.GetErrorAnalysis)
		tracingAPI.GET("/ingestion/audit", adminOrSysMiddleware, tracingHandler.ListIngestionAudits)
	}

	// ====================================
	// GIS DASHBOARD ENDPOINTS
	// ====================================
	gisHandler := handlers.NewGISHandler()
	gisAPI := router.Group("/api/v1/gis", authMiddleware)
	{
		gisAPI.GET("/summary", gisHandler.GetSummary)

		gisAPI.GET("/layers", gisHandler.ListLayers)
		gisAPI.POST("/layers", adminOrSysMiddleware, gisHandler.CreateLayer)
		gisAPI.PUT("/layers/:id", adminOrSysMiddleware, gisHandler.UpdateLayer)
		gisAPI.DELETE("/layers/:id", adminOrSysMiddleware, gisHandler.DeleteLayer)

		gisAPI.GET("/regions", gisHandler.ListRegions)
		gisAPI.GET("/regions/:id", gisHandler.GetRegion)
		gisAPI.POST("/regions", adminOrSysMiddleware, gisHandler.CreateRegion)
		gisAPI.PUT("/regions/:id", adminOrSysMiddleware, gisHandler.UpdateRegion)
		gisAPI.DELETE("/regions/:id", adminOrSysMiddleware, gisHandler.DeleteRegion)

		gisAPI.GET("/markers", gisHandler.ListMarkers)
		gisAPI.POST("/markers", adminOrSysMiddleware, gisHandler.CreateMarker)
		gisAPI.DELETE("/markers/:id", adminOrSysMiddleware, gisHandler.DeleteMarker)

		gisAPI.GET("/datasets", gisHandler.ListDatasets)
		gisAPI.GET("/datasets/:id", gisHandler.GetDataset)
		gisAPI.POST("/datasets", adminOrSysMiddleware, gisHandler.CreateDataset)
		gisAPI.PUT("/datasets/:id", adminOrSysMiddleware, gisHandler.UpdateDataset)
		gisAPI.DELETE("/datasets/:id", adminOrSysMiddleware, gisHandler.DeleteDataset)
	}

	// Specialized GIS dashboards (agriculture, industries, medical, satellite, airplane, ship)
	gisSpecHandler := handlers.NewGISSpecializedHandler()
	gisSpecAPI := router.Group("/api/v1/gis/dashboards", authMiddleware)
	{
		gisSpecAPI.GET("", gisSpecHandler.ListDashboardTypes)
		gisSpecAPI.GET("/:type", gisSpecHandler.GetDashboard)
		gisSpecAPI.GET("/:type/summary", gisSpecHandler.GetDashboardSummary)
	}

	// Analytics dashboards (charts, graphs, tables, KPI, heatmap, export)
	analyticsHandler := handlers.NewAnalyticsHandler()
	analyticsAPI := router.Group("/api/v1/analytics", authMiddleware)
	{
		analyticsAPI.GET("/dashboards", analyticsHandler.ListDashboards)
		analyticsAPI.GET("/dashboards/:id", analyticsHandler.GetDashboard)
		analyticsAPI.PUT("/dashboards/:id/widgets/:widgetId", adminOrSysMiddleware, analyticsHandler.UpdateWidget)
		analyticsAPI.PUT("/dashboards/:id/layout", adminOrSysMiddleware, analyticsHandler.ReorderWidgets)
		analyticsAPI.GET("/dashboards/:id/widgets/:widgetId/export", analyticsHandler.ExportCSV)
		analyticsAPI.GET("/widget-types", analyticsHandler.GetWidgetTypes)
	}

	// ====================================
	// CDC & ETL DATA PLATFORM ENDPOINTS
	// ====================================
	cdcEtlHandler := handlers.NewCDCETLHandler(conns.Etcd)

	// ETL Pipeline Management
	etlAPI := router.Group("/api/v1/etl", authMiddleware)
	{
		etlAPI.GET("/pipelines", cdcEtlHandler.ListETLPipelines)
		etlAPI.GET("/pipelines/:id", cdcEtlHandler.GetETLPipeline)
		etlAPI.POST("/pipelines", adminOrSysMiddleware, cdcEtlHandler.CreateETLPipeline)
		etlAPI.PUT("/pipelines/:id", adminOrSysMiddleware, cdcEtlHandler.UpdateETLPipeline)
		etlAPI.DELETE("/pipelines/:id", adminOrSysMiddleware, cdcEtlHandler.DeleteETLPipeline)
		etlAPI.POST("/pipelines/:id/run", adminOrSysMiddleware, cdcEtlHandler.RunETLPipeline)
		etlAPI.GET("/runs", cdcEtlHandler.ListETLRuns)
		etlAPI.GET("/runs/:id", cdcEtlHandler.GetETLRun)
		etlAPI.POST("/connectors", adminOrSysMiddleware, cdcEtlHandler.CreateETLConnector)
		etlAPI.GET("/connectors", cdcEtlHandler.GetETLConnectors)
		etlAPI.GET("/connectors/catalog", cdcEtlHandler.GetETLConnectorCatalog)
		etlAPI.GET("/orchestration/capabilities", cdcEtlHandler.GetETLOrchestrationCapabilities)
		etlAPI.GET("/blueprints", cdcEtlHandler.GetETLBlueprints)
		etlAPI.GET("/observability", cdcEtlHandler.GetETLObservability)
	}

	// CDC Pipeline Management
	cdcAPI := router.Group("/api/v1/cdc", authMiddleware)
	{
		cdcAPI.GET("/pipelines", cdcEtlHandler.ListCDCPipelines)
		cdcAPI.GET("/pipelines/:id", cdcEtlHandler.GetCDCPipeline)
		cdcAPI.POST("/pipelines", adminOrSysMiddleware, cdcEtlHandler.CreateCDCPipeline)
		cdcAPI.PUT("/pipelines/:id", adminOrSysMiddleware, cdcEtlHandler.UpdateCDCPipeline)
		cdcAPI.DELETE("/pipelines/:id", adminOrSysMiddleware, cdcEtlHandler.DeleteCDCPipeline)
		cdcAPI.POST("/pipelines/:id/start", adminOrSysMiddleware, cdcEtlHandler.StartCDCPipeline)
		cdcAPI.POST("/pipelines/:id/pause", adminOrSysMiddleware, cdcEtlHandler.PauseCDCPipeline)
		cdcAPI.POST("/pipelines/:id/stop", adminOrSysMiddleware, cdcEtlHandler.StopCDCPipeline)
		cdcAPI.GET("/sources", cdcEtlHandler.GetCDCSourceTypes)
		cdcAPI.GET("/sinks", cdcEtlHandler.GetCDCSinkTypes)
		cdcAPI.GET("/observability", cdcEtlHandler.GetCDCObservability)
	}

	// Data Platform Overview
	router.GET("/api/v1/data-platform/overview", authMiddleware, cdcEtlHandler.GetPlatformOverview)

	// ====================================
	// API BUILDER, CSV DASHBOARD & CONVERSION
	// ====================================
	apiBuilderHandler := handlers.NewAPIBuilderHandler(analyticsHandler, gisHandler, dbConnections, conns.Etcd)

	builderAPI := router.Group("/api/v1/builder", authMiddleware)
	{
		// Summary
		builderAPI.GET("/summary", apiBuilderHandler.GetSummary)

		// Custom API CRUD
		builderAPI.GET("/apis", apiBuilderHandler.ListAPIs)
		builderAPI.GET("/apis/:id", apiBuilderHandler.GetAPI)
		builderAPI.POST("/apis", adminOrSysMiddleware, apiBuilderHandler.CreateAPI)
		builderAPI.PUT("/apis/:id", adminOrSysMiddleware, apiBuilderHandler.UpdateAPI)
		builderAPI.DELETE("/apis/:id", adminOrSysMiddleware, apiBuilderHandler.DeleteAPI)
		builderAPI.POST("/apis/:id/test", adminOrSysMiddleware, apiBuilderHandler.TestAPI)

		// CSV Upload & Dashboard Generation
		builderAPI.POST("/csv/upload", adminOrSysMiddleware, apiBuilderHandler.UploadCSV)
		builderAPI.GET("/csv/uploads", apiBuilderHandler.ListCSVUploads)
		builderAPI.GET("/csv/uploads/:id", apiBuilderHandler.GetCSVUpload)
		builderAPI.DELETE("/csv/uploads/:id", adminOrSysMiddleware, apiBuilderHandler.DeleteCSVUpload)
		builderAPI.POST("/csv/uploads/:id/generate-dashboard", adminOrSysMiddleware, apiBuilderHandler.GenerateDashboard)
		builderAPI.POST("/csv/uploads/:id/generate-gis", adminOrSysMiddleware, apiBuilderHandler.GenerateGISFromCSV)

		// Dashboard <-> GIS Conversion
		builderAPI.POST("/convert/analyze", adminOrSysMiddleware, apiBuilderHandler.AnalyzeConversion)
		builderAPI.POST("/convert/dashboard-to-gis", adminOrSysMiddleware, apiBuilderHandler.ConvertDashboardToGIS)
		builderAPI.POST("/convert/gis-to-dashboard", adminOrSysMiddleware, apiBuilderHandler.ConvertGISToDashboard)
		builderAPI.GET("/conversions", apiBuilderHandler.ListConversions)

		// File Scanner (SafeGate Pipeline)
		builderAPI.POST("/scanner/scan", adminOrSysMiddleware, apiBuilderHandler.ScanFile)
		builderAPI.GET("/scanner/scans", apiBuilderHandler.ListScans)
		builderAPI.GET("/scanner/health", apiBuilderHandler.GetScannerHealth)

		// API Scanner Reports
		builderAPI.POST("/api-scanner/scan", adminOrSysMiddleware, apiBuilderHandler.ScanAPI)
		builderAPI.GET("/api-scanner/reports", apiBuilderHandler.ListAPIScanReports)
		builderAPI.POST("/api-scanner/reports/bulk-delete", adminOrSysMiddleware, apiBuilderHandler.BulkDeleteAPIScanReports)
		builderAPI.GET("/api-scanner/reports/:id", apiBuilderHandler.GetAPIScanReport)
		builderAPI.DELETE("/api-scanner/reports/:id", adminOrSysMiddleware, apiBuilderHandler.DeleteAPIScanReport)

		// SQL Assistant for API Builder
		builderAPI.POST("/sql-assistant/chat", adminOrSysMiddleware, apiBuilderHandler.ChatSQLAssistant)

		// Dashboard Deletion
		builderAPI.DELETE("/dashboards/:id", adminOrSysMiddleware, apiBuilderHandler.DeleteDashboard)
	}

	// Runtime execution routes for REST APIs created via API Builder.
	router.Any("/api/custom", authMiddleware, apiBuilderHandler.InvokeCustomAPI)
	router.Any("/api/custom/*path", authMiddleware, apiBuilderHandler.InvokeCustomAPI)

	// ====================================
	// NETWORK INTELLIGENCE ENDPOINTS
	// ====================================
	netIntelHandler := handlers.NewNetIntelHandler()
	modeManager := modes.NewManager()

	// ====================================
	// NEWLY ADDED FEATURE MODULES
	// ====================================
	admissionEngine := admission.NewEngine()
	admissionEngine.RegisterPolicy("template-001", 100, admission.PolicyTemplate001)
	admissionEngine.RegisterPolicy("template-002", 90, admission.PolicyTemplate002)
	admissionEngine.RegisterPolicy("template-003", 80, admission.PolicyTemplate003)

	kubeScheduler := scheduler.NewScheduler()
	crdRegistry := crd.NewRegistry()
	vectorIndex := vectorplus.NewIndex(4)
	reviewPipeline := reviewflow.NewPipeline()

	netintelAPI := router.Group("/api/v1/netintel", authMiddleware)
	{
		// Summary & Observability
		netintelAPI.GET("/summary", netIntelHandler.GetSummary)
		netintelAPI.GET("/observability", netIntelHandler.GetObservability)
		netintelAPI.GET("/log-types", netIntelHandler.GetLogTypes)

		// Parser CRUD
		netintelAPI.GET("/parsers", netIntelHandler.ListParsers)
		netintelAPI.GET("/parsers/:id", netIntelHandler.GetParser)
		netintelAPI.POST("/parsers", adminOrSysMiddleware, netIntelHandler.CreateParser)
		netintelAPI.PUT("/parsers/:id", adminOrSysMiddleware, netIntelHandler.UpdateParser)
		netintelAPI.DELETE("/parsers/:id", adminOrSysMiddleware, netIntelHandler.DeleteParser)

		// Log Entries
		netintelAPI.GET("/logs", netIntelHandler.ListEntries)
		netintelAPI.POST("/logs", adminOrSysMiddleware, netIntelHandler.IngestLog)
		netintelAPI.GET("/logs/stats", netIntelHandler.GetEntryStats)

		// Topology
		netintelAPI.GET("/topology", netIntelHandler.GetTopology)
		netintelAPI.GET("/topology/nodes/:id", netIntelHandler.GetTopologyNode)
		netintelAPI.PUT("/topology/nodes/:id", adminOrSysMiddleware, netIntelHandler.UpdateTopologyNode)

		// Heatmaps & Trends
		netintelAPI.GET("/heatmap", netIntelHandler.GetHeatmap)
		netintelAPI.GET("/trends", netIntelHandler.GetTrends)

		// Predictions & Tracks
		netintelAPI.GET("/predictions", netIntelHandler.GetPredictions)
		netintelAPI.GET("/tracks", netIntelHandler.ListTracks)
		netintelAPI.GET("/tracks/:mac", netIntelHandler.GetTrack)

		// Anomalies
		netintelAPI.GET("/anomalies", netIntelHandler.ListAnomalies)
		netintelAPI.POST("/anomalies/:id/acknowledge", adminOrSysMiddleware, netIntelHandler.AcknowledgeAnomaly)
		netintelAPI.POST("/anomalies/:id/resolve", adminOrSysMiddleware, netIntelHandler.ResolveAnomaly)

		// Alerts
		netintelAPI.GET("/alerts", netIntelHandler.ListAlerts)
		netintelAPI.POST("/alerts/:id/acknowledge", adminOrSysMiddleware, netIntelHandler.AcknowledgeAlert)
		netintelAPI.POST("/alerts/:id/resolve", adminOrSysMiddleware, netIntelHandler.ResolveAlert)

		// Forecasts
		netintelAPI.GET("/forecasts", netIntelHandler.ListForecasts)
		netintelAPI.GET("/forecasts/:metric", netIntelHandler.GetForecast)

		// Modes (new module-backed endpoints)
		netintelAPI.GET("/modes", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"modes": modeManager.List()})
		})
		netintelAPI.PUT("/modes/:name", adminOrSysMiddleware, func(c *gin.Context) {
			var cfg modes.ModeConfig
			if err := c.ShouldBindJSON(&cfg); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			cfg.Name = modes.Mode(strings.ToLower(strings.TrimSpace(c.Param("name"))))
			if cfg.Name == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "mode name is required"})
				return
			}
			modeManager.Upsert(cfg)
			c.JSON(http.StatusOK, gin.H{"message": "mode upserted", "mode": cfg})
		})
		netintelAPI.POST("/modes/events", adminOrSysMiddleware, func(c *gin.Context) {
			var ev modes.ModeEvent
			if err := c.ShouldBindJSON(&ev); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if ev.Timestamp.IsZero() {
				ev.Timestamp = time.Now().UTC()
			}
			modeManager.Record(ev)
			c.JSON(http.StatusOK, gin.H{"message": "event recorded"})
		})
		netintelAPI.GET("/modes/:name/events", func(c *gin.Context) {
			name := modes.Mode(strings.ToLower(strings.TrimSpace(c.Param("name"))))
			c.JSON(http.StatusOK, gin.H{"events": modeManager.FindByMode(name)})
		})
		netintelAPI.POST("/modes/detect", func(c *gin.Context) {
			var body struct {
				Detector int       `json:"detector"`
				Samples  []float64 `json:"samples"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			var score float64
			switch body.Detector {
			case 2:
				score = modes.Detector002(body.Samples)
			case 3:
				score = modes.Detector003(body.Samples)
			case 4:
				score = modes.Detector004(body.Samples)
			case 5:
				score = modes.Detector005(body.Samples)
			default:
				score = modes.Detector001(body.Samples)
			}
			c.JSON(http.StatusOK, gin.H{"detector": body.Detector, "score": score})
		})
	}

	kubeplusAPI := router.Group("/api/v1/kubeplus", authMiddleware)
	{
		kubeplusAPI.GET("/admission/policies", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"policies": admissionEngine.ListPolicies()})
		})
		kubeplusAPI.POST("/admission/evaluate", func(c *gin.Context) {
			var req admission.AdmissionRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if req.Timestamp.IsZero() {
				req.Timestamp = time.Now().UTC()
			}
			c.JSON(http.StatusOK, admissionEngine.Evaluate(req))
		})

		kubeplusAPI.PUT("/scheduler/nodes/:name", adminOrSysMiddleware, func(c *gin.Context) {
			var node scheduler.Node
			if err := c.ShouldBindJSON(&node); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			node.Name = c.Param("name")
			if strings.TrimSpace(node.Name) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "node name is required"})
				return
			}
			kubeScheduler.UpsertNode(node)
			c.JSON(http.StatusOK, gin.H{"message": "node upserted", "node": node.Name})
		})
		kubeplusAPI.GET("/scheduler/nodes", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"nodes": kubeScheduler.ListNodes()})
		})
		kubeplusAPI.POST("/scheduler/score", func(c *gin.Context) {
			var w scheduler.Workload
			if err := c.ShouldBindJSON(&w); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"decisions": kubeScheduler.Score(w)})
		})
		kubeplusAPI.POST("/scheduler/pick", func(c *gin.Context) {
			var w scheduler.Workload
			if err := c.ShouldBindJSON(&w); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			best, ok := kubeScheduler.PickBest(w)
			if !ok {
				c.JSON(http.StatusNotFound, gin.H{"error": "no suitable node found"})
				return
			}
			c.JSON(http.StatusOK, best)
		})

		kubeplusAPI.POST("/crd/definitions", adminOrSysMiddleware, func(c *gin.Context) {
			var def crd.Definition
			if err := c.ShouldBindJSON(&def); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := crdRegistry.Register(def); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "definition registered"})
		})
		kubeplusAPI.GET("/crd/definitions", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"definitions": crdRegistry.List()})
		})
		kubeplusAPI.POST("/crd/validate", func(c *gin.Context) {
			var body struct {
				Group   string         `json:"group"`
				Kind    string         `json:"kind"`
				Version string         `json:"version"`
				Spec    map[string]any `json:"spec"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			def, ok := crdRegistry.Get(body.Group, body.Kind)
			if !ok {
				c.JSON(http.StatusNotFound, gin.H{"error": "definition not found"})
				return
			}
			if len(def.Versions) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "definition has no versions"})
				return
			}
			fields := def.Versions[0].Fields
			if strings.TrimSpace(body.Version) != "" {
				for _, v := range def.Versions {
					if v.Version == body.Version {
						fields = v.Fields
						break
					}
				}
			}
			c.JSON(http.StatusOK, crd.ValidateSpec(fields, body.Spec))
		})
	}

	vectorAPI := router.Group("/api/v1/vectorplus", authMiddleware)
	{
		vectorAPI.PUT("/records/:id", adminOrSysMiddleware, func(c *gin.Context) {
			var body struct {
				Vec    []float64         `json:"vec"`
				Labels map[string]string `json:"labels"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			rec := vectorplus.Record{ID: c.Param("id"), Vec: vectorplus.Vector(body.Vec), Labels: body.Labels}
			if !vectorIndex.Upsert(rec) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vector size or id"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "record upserted", "id": rec.ID})
		})
		vectorAPI.DELETE("/records/:id", adminOrSysMiddleware, func(c *gin.Context) {
			vectorIndex.Delete(c.Param("id"))
			c.JSON(http.StatusOK, gin.H{"message": "record deleted", "id": c.Param("id")})
		})
		vectorAPI.POST("/search", func(c *gin.Context) {
			var body struct {
				Query []float64 `json:"query"`
				K     int       `json:"k"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if body.K < 1 {
				body.K = 5
			}
			c.JSON(http.StatusOK, gin.H{"results": vectorIndex.Search(vectorplus.Vector(body.Query), body.K)})
		})
		vectorAPI.POST("/similarity", func(c *gin.Context) {
			var body struct {
				A      []float64 `json:"a"`
				B      []float64 `json:"b"`
				Metric int       `json:"metric"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			a := vectorplus.Vector(body.A)
			b := vectorplus.Vector(body.B)
			var score float64
			switch body.Metric {
			case 2:
				score = vectorplus.SimilarityMetric002(a, b)
			case 3:
				score = vectorplus.SimilarityMetric003(a, b)
			case 4:
				score = vectorplus.SimilarityMetric004(a, b)
			case 5:
				score = vectorplus.SimilarityMetric005(a, b)
			default:
				score = vectorplus.SimilarityMetric001(a, b)
			}
			c.JSON(http.StatusOK, gin.H{"metric": body.Metric, "score": score})
		})
	}

	reviewAPI := router.Group("/api/v1/reviewflow", authMiddleware)
	{
		reviewAPI.PUT("/items/:id", adminOrSysMiddleware, func(c *gin.Context) {
			var item reviewflow.ReviewItem
			if err := c.ShouldBindJSON(&item); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			item.ID = c.Param("id")
			if strings.TrimSpace(item.ID) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "item id is required"})
				return
			}
			if item.Score == 0 {
				item.Score = reviewflow.ScoreBySignals(item.Title, item.Description, item.Tags)
			}
			reviewPipeline.Upsert(item)
			c.JSON(http.StatusOK, gin.H{"message": "item upserted", "id": item.ID})
		})
		reviewAPI.GET("/items/:id", func(c *gin.Context) {
			item, ok := reviewPipeline.Get(c.Param("id"))
			if !ok {
				c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
				return
			}
			c.JSON(http.StatusOK, item)
		})
		reviewAPI.GET("/items", func(c *gin.Context) {
			stage := reviewflow.Stage(strings.TrimSpace(c.Query("stage")))
			c.JSON(http.StatusOK, gin.H{"items": reviewPipeline.ListByStage(stage)})
		})
		reviewAPI.POST("/items/:id/stage", adminOrSysMiddleware, func(c *gin.Context) {
			var body struct {
				Stage reviewflow.Stage `json:"stage"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if !reviewPipeline.Advance(c.Param("id"), body.Stage) {
				c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "stage updated", "id": c.Param("id"), "stage": body.Stage})
		})
		reviewAPI.POST("/score", func(c *gin.Context) {
			var body struct {
				Title       string   `json:"title"`
				Description string   `json:"description"`
				Tags        []string `json:"tags"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"score": reviewflow.ScoreBySignals(body.Title, body.Description, body.Tags)})
		})
		reviewAPI.POST("/quality", func(c *gin.Context) {
			var body struct {
				Item  reviewflow.ReviewItem `json:"item"`
				Check int                   `json:"check"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			var score float64
			switch body.Check {
			case 2:
				score = reviewflow.QualityCheck002(body.Item)
			case 3:
				score = reviewflow.QualityCheck003(body.Item)
			case 4:
				score = reviewflow.QualityCheck004(body.Item)
			case 5:
				score = reviewflow.QualityCheck005(body.Item)
			default:
				score = reviewflow.QualityCheck001(body.Item)
			}
			c.JSON(http.StatusOK, gin.H{"check": body.Check, "score": score})
		})
	}

	// ====================================
	// IAM ROUTES
	// ====================================
	if iamSystem != nil {
		iamSystem.RegisterRoutes(router)
		log.Println("✅ IAM routes registered")
	}

	apiPort := cfg.API.Port
	apiHost := cfg.API.Host

	fmt.Printf("📡 API Server running on http://%s:%s\n", apiHost, apiPort)
	fmt.Println("\n🔐 RBAC Security Model:")
	fmt.Println("  ✅ READ  operations (GET)     - Allowed for all authenticated users")
	fmt.Println("  ❌ WRITE operations (POST/PUT/DELETE) - Allowed ONLY for users with 'admin' role")
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /health                  - Health check (no auth)")
	fmt.Println("  GET  /status                  - Check all connections (no auth)")
	fmt.Println("  ANY  /api/custom/*path        - Execute API Builder runtime APIs")
	fmt.Println()
	fmt.Println("Admin endpoints (admin role required):")
	fmt.Println("  POST /api/admin/database/create       - Create a new database")
	fmt.Println("  GET  /api/admin/database/list         - List all databases")
	fmt.Println("  GET  /api/admin/database/servers      - List default and connected DB servers")
	fmt.Println("  POST /api/admin/database/connect      - Connect a new DB server")
	fmt.Println("  PUT  /api/admin/database/servers/:key - Update a custom DB server")
	fmt.Println("  DELETE /api/admin/database/servers/:key - Delete a custom DB server")
	fmt.Println("  POST /api/admin/table/create          - Create a new table")
	fmt.Println("  GET  /api/admin/table/list            - List all tables")
	fmt.Println()
	if !iamOnlyAuth {
		fmt.Println("User Management endpoints (admin only):")
		fmt.Println("  GET    /api/v1/users            - List all platform users")
		fmt.Println("  GET    /api/v1/users/:id        - Get a platform user")
		fmt.Println("  POST   /api/v1/users            - Create a platform user")
		fmt.Println("  PUT    /api/v1/users/:id        - Update a platform user")
		fmt.Println("  DELETE /api/v1/users/:id        - Delete a platform user")
		fmt.Println()
	}
	fmt.Println("Dynamic Query endpoints:")
	fmt.Println("  GET  /api/{db}/query            - Execute SELECT queries with parameters")
	fmt.Println("       Example: /api/mysql/query?q=SELECT * FROM users&params=1")
	fmt.Println("  POST /api/{db}/query            - Execute query body (admin/system-manager only)")
	fmt.Println("       Body: {\"query\": \"SQL_QUERY\", \"params\": [\"value1\", \"value2\"]}")
	fmt.Println("  POST /api/{db}/query/batch      - Execute multiple queries at once (admin/system-manager only)")
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

	runtimeHost := strings.TrimSpace(os.Getenv("RUNTIME_HOST"))
	if runtimeHost == "" {
		runtimeHost = apiHost
	}
	runtimePort := strings.TrimSpace(os.Getenv("RUNTIME_PORT"))
	if runtimePort == "" {
		runtimePort = "8001"
	}
	if runtimeHost == apiHost && runtimePort == apiPort {
		runtimePort = "8001"
	}
	runtimeAddr := fmt.Sprintf("%s:%s", runtimeHost, runtimePort)

	// Start runtime in background on a dedicated port to avoid router conflicts.
	go func() {
		if err := rt.Start(ctx, runtimeAddr); err != nil {
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

const (
	demoJWTSecretStoreKey = "demo-jwt-secret"
	demoJWTSecretEtcdKey  = "iam:bootstrap:demo-jwt-secret"
)

func ensureSharedDemoJWTSecret(pg *gorm.DB, etcd *clientv3.Client) (string, error) {
	if configured := strings.TrimSpace(os.Getenv("DEMO_JWT_SECRET")); configured != "" {
		if pg != nil {
			resolved, err := bootstrapsecrets.Ensure(pg, demoJWTSecretStoreKey, func() (string, error) {
				return configured, nil
			})
			if err != nil {
				log.Printf("⚠️  failed to seed DEMO_JWT_SECRET into postgres bootstrap store: %v", err)
			} else if resolved != configured {
				log.Printf("⚠️  postgres bootstrap DEMO_JWT_SECRET differs from env value; keeping env for current runtime")
			}
		}
		auth.SetDemoJWTSecret(configured)
		return configured, nil
	}

	if pg != nil {
		resolved, err := bootstrapsecrets.Ensure(pg, demoJWTSecretStoreKey, func() (string, error) {
			return generateBootstrapSecret(48)
		})
		if err == nil {
			if err := os.Setenv("DEMO_JWT_SECRET", resolved); err != nil {
				return "", fmt.Errorf("setting DEMO_JWT_SECRET from postgres: %w", err)
			}
			auth.SetDemoJWTSecret(resolved)
			return resolved, nil
		}
		log.Printf("⚠️  postgres bootstrap for DEMO_JWT_SECRET failed, falling back to etcd: %v", err)
	}

	resolved, err := ensureSharedDemoJWTSecretFromEtcd(etcd)
	if err != nil {
		return "", err
	}
	if err := os.Setenv("DEMO_JWT_SECRET", resolved); err != nil {
		return "", fmt.Errorf("setting DEMO_JWT_SECRET from etcd: %w", err)
	}
	auth.SetDemoJWTSecret(resolved)
	return resolved, nil
}

func ensureSharedDemoJWTSecretFromEtcd(etcd *clientv3.Client) (string, error) {

	if etcd == nil {
		return "", fmt.Errorf("DEMO_JWT_SECRET is not set and neither postgres nor etcd bootstrap store is available")
	}

	getCtx, getCancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := etcd.Get(getCtx, demoJWTSecretEtcdKey)
	getCancel()
	if err != nil {
		return "", fmt.Errorf("reading demo token secret from etcd: %w", err)
	}
	if len(resp.Kvs) > 0 {
		secret := strings.TrimSpace(string(resp.Kvs[0].Value))
		if secret != "" {
			return secret, nil
		}
	}

	candidate, err := generateBootstrapSecret(48)
	if err != nil {
		return "", err
	}

	txnCtx, txnCancel := context.WithTimeout(context.Background(), 5*time.Second)
	txnResp, err := etcd.Txn(txnCtx).
		If(clientv3.Compare(clientv3.Version(demoJWTSecretEtcdKey), "=", 0)).
		Then(clientv3.OpPut(demoJWTSecretEtcdKey, candidate)).
		Else(clientv3.OpGet(demoJWTSecretEtcdKey)).
		Commit()
	txnCancel()
	if err != nil {
		return "", fmt.Errorf("persisting demo token secret in etcd: %w", err)
	}

	resolved := candidate
	if !txnResp.Succeeded {
		resolved = ""
		if len(txnResp.Responses) > 0 {
			rangeResp := txnResp.Responses[0].GetResponseRange()
			if rangeResp != nil && len(rangeResp.Kvs) > 0 {
				resolved = strings.TrimSpace(string(rangeResp.Kvs[0].Value))
			}
		}
		if resolved == "" {
			return "", fmt.Errorf("demo token secret exists in etcd but value is empty")
		}
	}

	return resolved, nil
}

func generateBootstrapSecret(size int) (string, error) {
	if size <= 0 {
		size = 48
	}
	random := make([]byte, size)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generating bootstrap secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(random), nil
}

func resolveSecurityEnvironment() string {
	candidates := []string{
		strings.TrimSpace(os.Getenv("AXIOMNIZAM_ENV")),
		strings.TrimSpace(os.Getenv("APP_ENV")),
		strings.TrimSpace(os.Getenv("ENVIRONMENT")),
		strings.TrimSpace(os.Getenv("GO_ENV")),
	}
	for _, c := range candidates {
		if c != "" {
			return strings.ToLower(c)
		}
	}
	return "development"
}

func isProductionEnvironment(env string) bool {
	normalized := strings.ToLower(strings.TrimSpace(env))
	return normalized == "production" || normalized == "prod"
}

func resolveSecurityGuardrailMode(env string) string {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("SECURITY_GUARDRAILS_MODE")))
	switch mode {
	case "off", "audit", "enforce":
		return mode
	case "":
		if isProductionEnvironment(env) {
			return "audit"
		}
		return "off"
	default:
		log.Printf("⚠️  Unknown SECURITY_GUARDRAILS_MODE=%q, defaulting to audit", mode)
		return "audit"
	}
}

func applySecurityGuardrails(cfg *config.Config) {
	if cfg == nil {
		return
	}

	env := resolveSecurityEnvironment()
	mode := resolveSecurityGuardrailMode(env)
	if mode == "off" {
		return
	}

	blocking := make([]string, 0)
	warnings := make([]string, 0)
	addBlocking := func(msg string) {
		if strings.TrimSpace(msg) != "" {
			blocking = append(blocking, msg)
		}
	}
	addWarning := func(msg string) {
		if strings.TrimSpace(msg) != "" {
			warnings = append(warnings, msg)
		}
	}

	isDefault := func(value string, defaults ...string) bool {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		for _, d := range defaults {
			if trimmed == strings.ToLower(strings.TrimSpace(d)) {
				return true
			}
		}
		return false
	}

	if isDefault(cfg.IAM.SysadminPassword, "", "change-me", "changeme", "default", "password", "admin") {
		addBlocking("IAM_SYSADMIN_PASSWORD is empty or default-like")
	}
	if strings.TrimSpace(os.Getenv("DEMO_JWT_SECRET")) == "" {
		addBlocking("DEMO_JWT_SECRET is not set")
	}
	if strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")) == "" {
		addBlocking("CORS_ALLOWED_ORIGINS is empty")
	}

	if isDefault(cfg.MySQL.Password, "root", "password") {
		addWarning("MYSQL_PASSWORD appears to be a default credential")
	}
	if isDefault(cfg.PostgreSQL.Password, "postgres", "password") {
		addWarning("POSTGRES_PASSWORD appears to be a default credential")
	}
	if isDefault(cfg.Oracle.Password, "oracle123", "password") {
		addWarning("ORACLE_PASSWORD appears to be a default credential")
	}

	for _, w := range warnings {
		log.Printf("⚠️  Security guardrail warning: %s", w)
	}

	if len(blocking) == 0 {
		if mode == "audit" {
			log.Printf("✅ Security guardrails check passed (env=%s, mode=%s)", env, mode)
		}
		return
	}

	for _, b := range blocking {
		log.Printf("🚨 Security guardrail issue: %s", b)
	}

	if mode == "enforce" && isProductionEnvironment(env) {
		log.Fatalf("❌ Security guardrails blocked startup in production (mode=%s)", mode)
		return
	}

	log.Printf("⚠️  Security guardrails detected %d blocking issue(s) but startup continues (env=%s, mode=%s)", len(blocking), env, mode)
}

func ensureWorkflowRegistered(ctx context.Context, resourceHandler *handlers.ResourceHandler, workflowName string) error {
	if resourceHandler == nil {
		if workflows.GlobalWorkflowEngine.GetWorkflow(workflowName) != nil {
			return nil
		}
		return fmt.Errorf("workflow %q not found", workflowName)
	}

	resource, found := resourceHandler.FindResourceByKindAndName("workflow", workflowName)
	if !found {
		if workflows.GlobalWorkflowEngine.GetWorkflow(workflowName) != nil {
			return nil
		}
		return fmt.Errorf("workflow %q not found", workflowName)
	}

	workflowDef, err := workflowFromResource(resource)
	if err != nil {
		return err
	}

	return workflows.AddWorkflow(ctx, workflowDef)
}

func workflowFromResource(resource *handlers.GenericResource) (*workflows.Workflow, error) {
	if resource == nil {
		return nil, fmt.Errorf("workflow definition is nil")
	}

	name := strings.TrimSpace(resource.Metadata.Name)
	if name == "" {
		return nil, fmt.Errorf("workflow metadata.name is required")
	}

	steps, err := workflowStepsFromSpec(name, resource.Spec)
	if err != nil {
		return nil, err
	}

	enabled := true
	if v, ok := boolFromAny(resource.Spec["enabled"]); ok {
		enabled = v
	}
	if schedule, ok := resource.Spec["schedule"].(map[string]interface{}); ok {
		if v, ok := boolFromAny(schedule["enabled"]); ok {
			enabled = v
		}
	}

	version := strings.TrimSpace(stringFromAny(resource.Spec["version"]))
	if version == "" {
		version = "v1"
	}

	namespace := strings.TrimSpace(resource.Metadata.Namespace)
	if namespace == "" {
		namespace = "default"
	}

	return &workflows.Workflow{
		Name:        name,
		Namespace:   namespace,
		Version:     version,
		Description: stringFromAny(resource.Spec["description"]),
		Triggers:    workflowTriggersFromSpec(resource.Spec),
		Steps:       steps,
		Enabled:     enabled,
		Labels:      resource.Metadata.Labels,
		Annotations: resource.Metadata.Annotations,
	}, nil
}

func workflowTriggersFromSpec(spec map[string]interface{}) []workflows.WorkflowTrigger {
	triggers := make([]workflows.WorkflowTrigger, 0)

	if raw, ok := spec["triggers"].([]interface{}); ok {
		for _, item := range raw {
			triggerMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			triggerType := strings.TrimSpace(stringFromAny(triggerMap["type"]))
			if triggerType == "" {
				continue
			}

			condition := make(map[string]interface{})
			if condMap, ok := triggerMap["condition"].(map[string]interface{}); ok {
				for k, v := range condMap {
					condition[k] = v
				}
			}

			triggers = append(triggers, workflows.WorkflowTrigger{
				Type:      triggerType,
				Condition: condition,
			})
		}
	}

	if schedule, ok := spec["schedule"].(map[string]interface{}); ok {
		condition := make(map[string]interface{}, len(schedule))
		for k, v := range schedule {
			condition[k] = v
		}
		triggers = append(triggers, workflows.WorkflowTrigger{Type: "schedule", Condition: condition})
	}

	if len(triggers) == 0 {
		triggers = append(triggers, workflows.WorkflowTrigger{Type: "manual", Condition: map[string]interface{}{"source": "api"}})
	}

	return triggers
}

func workflowStepsFromSpec(workflowName string, spec map[string]interface{}) ([]workflows.WorkflowStep, error) {
	rawSteps, ok := spec["steps"].([]interface{})
	if !ok || len(rawSteps) == 0 {
		return nil, fmt.Errorf("workflow %q must define at least one step", workflowName)
	}

	steps := make([]workflows.WorkflowStep, 0, len(rawSteps))
	for i, rawStep := range rawSteps {
		stepMap, ok := rawStep.(map[string]interface{})
		if !ok {
			continue
		}

		stepID := strings.TrimSpace(stringFromAny(stepMap["id"]))
		if stepID == "" {
			stepID = fmt.Sprintf("%s-step-%d", workflowName, i+1)
		}

		stepName := strings.TrimSpace(stringFromAny(stepMap["name"]))
		if stepName == "" {
			stepName = stepID
		}

		stepType := strings.TrimSpace(stringFromAny(stepMap["type"]))
		if stepType == "" {
			stepType = "http"
		}

		action := strings.TrimSpace(stringFromAny(stepMap["action"]))
		if action == "" {
			action = stepType
		}

		config := make(map[string]interface{})
		if rawConfig, ok := stepMap["config"].(map[string]interface{}); ok {
			for k, v := range rawConfig {
				config[k] = v
			}
		}

		for k, v := range stepMap {
			switch k {
			case "id", "name", "type", "action", "retry", "timeout", "config":
				continue
			default:
				config[k] = v
			}
		}

		if _, exists := config["action"]; !exists && action != "" {
			config["action"] = action
		}
		if stepType == "http" {
			if _, exists := config["method"]; !exists {
				method := strings.ToUpper(strings.TrimSpace(stringFromAny(stepMap["method"])))
				if method == "" {
					method = "GET"
				}
				config["method"] = method
			}
		}

		steps = append(steps, workflows.WorkflowStep{
			ID:      stepID,
			Name:    stepName,
			Type:    stepType,
			Action:  action,
			Config:  config,
			Timeout: durationFromAny(stepMap["timeout"]),
			Retry:   intFromAny(stepMap["retry"]),
		})
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("workflow %q has invalid steps", workflowName)
	}

	return steps, nil
}

func stringFromAny(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

func boolFromAny(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(v))
		if err == nil {
			return parsed, true
		}
	}
	return false, false
}

func intFromAny(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil {
			return parsed
		}
	}
	return 0
}

func durationFromAny(value interface{}) time.Duration {
	switch v := value.(type) {
	case time.Duration:
		return v
	case string:
		parsed, err := time.ParseDuration(strings.TrimSpace(v))
		if err == nil {
			return parsed
		}
	case int:
		if v > 0 {
			return time.Duration(v) * time.Second
		}
	case int64:
		if v > 0 {
			return time.Duration(v) * time.Second
		}
	case float64:
		if v > 0 {
			return time.Duration(v * float64(time.Second))
		}
	}
	return 0
}
