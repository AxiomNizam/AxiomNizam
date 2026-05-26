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

	"example.com/axiomnizam/internal/alerting"
	alertingmodels "example.com/axiomnizam/internal/alerting/models"
	"example.com/axiomnizam/internal/anonymization"
	"example.com/axiomnizam/internal/antivirus"
	"example.com/axiomnizam/internal/apibanks"
	"example.com/axiomnizam/internal/apiscanner"
	"example.com/axiomnizam/internal/audit"
	"example.com/axiomnizam/internal/auth"
	"example.com/axiomnizam/internal/autopilot"
	"example.com/axiomnizam/internal/bootstrapsecrets"
	"example.com/axiomnizam/internal/bulk"
	"example.com/axiomnizam/internal/catalog"
	"example.com/axiomnizam/internal/cdc"
	"example.com/axiomnizam/internal/conductor"
	"example.com/axiomnizam/internal/config"
	"example.com/axiomnizam/internal/contracts"
	"example.com/axiomnizam/internal/costing"
	"example.com/axiomnizam/internal/database"
	datasourceresource "example.com/axiomnizam/internal/datasource"
	"example.com/axiomnizam/internal/deployment"
	"example.com/axiomnizam/internal/encryption"
	"example.com/axiomnizam/internal/etl"
	"example.com/axiomnizam/internal/eventbus"
	exportpkg "example.com/axiomnizam/internal/export"
	"example.com/axiomnizam/internal/federation"
	"example.com/axiomnizam/internal/featurestore"
	"example.com/axiomnizam/internal/gatekeeper"
	analyticspkg "example.com/axiomnizam/internal/analytics"
	gispkg "example.com/axiomnizam/internal/gis"
	"example.com/axiomnizam/internal/governance"
	governancemodels "example.com/axiomnizam/internal/governance/models"
	graphqlpkg "example.com/axiomnizam/internal/graphql"
	healthpkg "example.com/axiomnizam/internal/health"
	apibuilder "example.com/axiomnizam/internal/apibuilder"
	authn "example.com/axiomnizam/internal/iam/authn"
	netintelpkg "example.com/axiomnizam/internal/netintel"
	notificationpkg "example.com/axiomnizam/internal/notification"
	transformpkg "example.com/axiomnizam/internal/transform"
	"example.com/axiomnizam/internal/heartbeat"
	iampkg "example.com/axiomnizam/internal/iam"
	iamstorage "example.com/axiomnizam/internal/iam/storage"
	iamtoken "example.com/axiomnizam/internal/iam/token"
	iamusers "example.com/axiomnizam/internal/iam/users"
	"example.com/axiomnizam/internal/integration"
	"example.com/axiomnizam/internal/jobs"
	"example.com/axiomnizam/internal/kubeplus/admission"
	"example.com/axiomnizam/internal/kubeplus/crd"
	"example.com/axiomnizam/internal/kubeplus/scheduler"
	"example.com/axiomnizam/internal/lineage"
	"example.com/axiomnizam/internal/metrics"
	"example.com/axiomnizam/internal/migrations"
	"example.com/axiomnizam/internal/mlpipeline"
	"example.com/axiomnizam/internal/models"
	"example.com/axiomnizam/internal/netintel/modes"
	"example.com/axiomnizam/internal/platform"
	genericctrl "example.com/axiomnizam/internal/platform/controller"
	querypkg "example.com/axiomnizam/internal/query"
	resourcespkg "example.com/axiomnizam/internal/resources"
	"example.com/axiomnizam/internal/platform/featureflags"
	platformstore "example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/policies"
	"example.com/axiomnizam/internal/ratelimit"
	"example.com/axiomnizam/internal/rbac"
	reconcilerpkg "example.com/axiomnizam/internal/reconciler"
	"example.com/axiomnizam/internal/reviewflow"
	"example.com/axiomnizam/internal/runtime"
	securitypkg "example.com/axiomnizam/internal/security"
	"example.com/axiomnizam/internal/schemaregistry"
	"example.com/axiomnizam/internal/serviceregistry"
	"example.com/axiomnizam/internal/slo"
	"example.com/axiomnizam/internal/storage"
	"example.com/axiomnizam/internal/stream"
	"example.com/axiomnizam/internal/streamanalytics"
	"example.com/axiomnizam/internal/streaming"
	"example.com/axiomnizam/internal/tenant"
	"example.com/axiomnizam/internal/tracing"
	"example.com/axiomnizam/internal/trivy"
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

	// Module registry — collects all modules for uniform lifecycle management.
	var modules []contracts.Module

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

	// JWT secret and module persistence configuration.
	// When using Raft backend, these are deferred until BackendManager is ready.
	// When using etcd (default), they run immediately.
	earlyStorageBackend := strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_BACKEND")))
	if earlyStorageBackend != "raft" {
		if _, secretErr := ensureSharedDemoJWTSecret(conns.PostgreSQL, conns.Etcd, nil); secretErr != nil {
			log.Printf("⚠️  DEMO_JWT_SECRET synchronization failed: %v", secretErr)
		} else {
			log.Println("✅ DEMO_JWT_SECRET synchronized for replica-safe token validation")
		}
		workflows.ConfigureGlobalPersistence(conns.Etcd)
		modes.ConfigureGlobalPersistence(conns.Etcd)
		vectorplus.ConfigureGlobalPersistence(conns.Etcd)
		reviewflow.ConfigureGlobalPersistence(conns.Etcd)
		integration.ConfigureGlobalPersistence(conns.Etcd)
	} else {
		// Raft mode: JWT secret from env or postgres only (KVStore not ready yet).
		if _, secretErr := ensureSharedDemoJWTSecret(conns.PostgreSQL, nil, nil); secretErr != nil {
			log.Printf("ℹ️  DEMO_JWT_SECRET deferred to Raft backend init: %v", secretErr)
		} else {
			log.Println("✅ DEMO_JWT_SECRET synchronized (postgres/env)")
		}
		// Module persistence will be configured after BackendManager init.
	}

	// Create tables
	createTables(conns)

	// NOTE: Legacy train / bd-train GIS PostgreSQL connections and their
	// handlers (GISTrainHandler, GISBDTrainHandler) were removed as part of
	// the Kubernetes-style control-plane refactor. GIS APIs are now authored
	// via the API Builder, which persists artifacts through the
	// ResourceStore -> Controller -> Reconciler pipeline. External railway
	// datasets should be exposed through DataSource resources and reached
	// only from reconcilers, never from HTTP handlers.

	// ====================================
	// IAM SYSTEM INITIALIZATION
	// ====================================
	var iamSystem *iampkg.System
	var iamErr error

	if earlyStorageBackend != "raft" {
		// etcd mode: initialize IAM immediately.
		iamSystem, iamErr = iampkg.NewSystem(conns.PostgreSQL, conns.Etcd, iampkg.Config{
			IssuerURL: strings.TrimSpace(os.Getenv("IAM_ISSUER_URL")),
		})
	} else {
		// Raft mode: IAM will be initialized after BackendManager is ready.
		log.Println("ℹ️  IAM initialization deferred (STORAGE_BACKEND=raft — waiting for Raft KV)")
	}
	if iamErr != nil {
		log.Printf("⚠️  IAM system initialization failed: %v", iamErr)
		log.Println("⚠️  IAM endpoints will not be available. Ensure PostgreSQL is connected.")
	} else if iamSystem != nil {
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
	apiMetricsTracker := metrics.NewAPIMetricsTracker(conns.Valkey)
	router.Use(metrics.MetricsMiddleware(apiMetricsTracker))

	// Initialize Rate Limiter
	// Max calls and token validity from config (.env)
	rateLimiter := auth.NewRateLimiter(cfg.RateLimiting.MaxCallsPerToken, cfg.RateLimiting.TokenValidityMinutes)

	// Initialize Query Logger with Valkey/Redis
	queryLogger := querypkg.NewQueryLogger(conns.Valkey, "/data/query_logs")

	// Initialize all handlers
	healthHandler := healthpkg.NewHandler(conns)

	// Admin handler for database and table creation
	// Only include SQL databases (MongoDB and Firebase don't support SQL DDL operations)
	dbConnections := map[string]*gorm.DB{
		"mysql":    conns.MySQL,
		"mariadb":  conns.MariaDB,
		"postgres": conns.PostgreSQL,
		"percona":  conns.Percona,
		"oracle":   conns.Oracle,
	}
	adminHandler := database.NewHandler(dbConnections, conns.PostgreSQL)

	// User management handler
	platformUserHandler := iamusers.NewPlatformUserHandler(conns.Etcd)

	// Dynamic Query handlers for each database
	mysqlDynamicHandler := querypkg.NewHandler(conns.MySQL, queryLogger)
	mariadbDynamicHandler := querypkg.NewHandler(conns.MariaDB, queryLogger)
	postgresDynamicHandler := querypkg.NewHandler(conns.PostgreSQL, queryLogger)
	perconaDynamicHandler := querypkg.NewHandler(conns.Percona, queryLogger)
	oracleDynamicHandler := querypkg.NewHandler(conns.Oracle, queryLogger)

	// Notification handler
	discordWebhookURL := cfg.Discord.WebhookURL
	notificationHandler := notificationpkg.NewHandler(discordWebhookURL, dbConnections)

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
	graphQLHandler := graphqlpkg.NewHandler(graphQLDB)

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
			// WebSocket connections cannot send custom headers from browsers;
			// accept token as a query parameter for upgrade requests.
			if qToken := strings.TrimSpace(c.Query("token")); qToken != "" {
				authHeader = "Bearer " + qToken
			}
		}
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

	// Phase 0: Reconciler health endpoint (no auth — ops visibility)
	var keySpaceMonitorRef *metrics.EtcdKeySpaceMonitor // set later when etcd is available
	router.GET("/health/reconcilers", func(c *gin.Context) {
		summary := metrics.GlobalReconcilerMetrics.HealthSummary()
		statuses := metrics.GlobalReconcilerMetrics.GetAllStatuses()
		status := http.StatusOK
		if summary.Status == "degraded" {
			status = http.StatusServiceUnavailable
		}
		response := gin.H{"summary": summary, "reconcilers": statuses}
		if keySpaceMonitorRef != nil {
			response["etcdKeySpace"] = keySpaceMonitorRef.GetStats()
		}
		c.JSON(status, response)
	})

	// Authentication endpoints (no auth required for login/refresh)
	authHandler := authn.NewAuthHandler()
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

	transformHandler := transformpkg.NewHandler()

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
	certificateHandler := securitypkg.NewHandler()

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
	cliAuth := authn.NewCLIAuthHandler()
	router.POST("/api/v1/auth/login", cliAuth.Login)
	router.POST("/api/v1/auth/logout", authHandler.Logout)
	router.GET("/api/v1/auth/verify", cliAuth.Verify)
	router.GET("/api/v1/auth/whoami", cliAuth.WhoAmI)

	// ====================================
	// KUBERNETES-STYLE RESOURCE ENDPOINTS
	// ====================================
	resourceHandler := resourcespkg.NewGenericResourceHandler(conns.Etcd)

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
	dsHandler := datasourceresource.NewDataSourceHandler(conns.Etcd)
	router.POST("/api/v1/datasources", adminOrSysMiddleware, dsHandler.Create)
	router.GET("/api/v1/datasources", authMiddleware, dsHandler.List)
	router.GET("/api/v1/datasources/:name", authMiddleware, dsHandler.Get)
	router.PUT("/api/v1/datasources/:name", adminOrSysMiddleware, dsHandler.Update)
	router.DELETE("/api/v1/datasources/:name", adminOrSysMiddleware, dsHandler.Delete)
	router.POST("/api/v1/datasources/:name/test", adminOrSysMiddleware, dsHandler.Test)

	// Job endpoints
	jobHandler := jobs.NewLegacyJobHandler(conns.Etcd)
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
		log.Printf("⚠️  platform managers initialization failed: %v", err)
		log.Println("  Platform service APIs may have limited functionality")
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
	conductorCfg := conductor.LoadConfigFromEnv()
	conductorMgr := conductor.NewManager(conductorCfg)
	conductorMgr.InitPersistence(conns.PostgreSQL)
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
	gisHandler := apibuilder.NewGISHandler()
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
	gisSpecHandler := gispkg.NewGISSpecializedHandler()
	gisSpecAPI := router.Group("/api/v1/gis/dashboards", authMiddleware)
	{
		gisSpecAPI.GET("", gisSpecHandler.ListDashboardTypes)
		gisSpecAPI.GET("/:type", gisSpecHandler.GetDashboard)
		gisSpecAPI.GET("/:type/summary", gisSpecHandler.GetDashboardSummary)
	}

	// GIS Train/Railway handlers (Indian + Bangladesh Railways) have been
	// removed. These previously held *gorm.DB directly and bypassed the
	// control plane entirely. Equivalent endpoints must now be authored in
	// the API Builder using DataSource resources.

	// Analytics dashboards (charts, graphs, tables, KPI, heatmap, export)
	analyticsHandler := apibuilder.NewAnalyticsHandler()
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
	cdcEtlHandler := cdc.NewHandler(conns.Etcd)

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
		etlAPI.PUT("/connectors/:id", adminOrSysMiddleware, cdcEtlHandler.UpdateETLConnector)
		etlAPI.DELETE("/connectors/:id", adminOrSysMiddleware, cdcEtlHandler.DeleteETLConnector)
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
	apiBuilderHandler := apibuilder.NewAPIBuilderHandler(analyticsHandler, gisHandler, dbConnections, conns.Etcd, nil)

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
		builderAPI.GET("/scanner/scans/:id", apiBuilderHandler.GetScan)
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
	netIntelHandler := netintelpkg.NewHandler()
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

	// ====================================
	// OBJECT STORAGE MODULE (Native S3)
	// ====================================
	storageCfg := storage.DefaultConfig()
	var storageIssuer *iamtoken.Issuer
	var storageRevokedStore *iamstorage.EtcdRevokedTokenStore
	if iamSystem != nil {
		storageIssuer = iamSystem.Issuer
		storageRevokedStore = iamSystem.RevokedStore
	}
	storageSys, storageErr := storage.NewSystem(storageCfg, storageIssuer, storageRevokedStore, conns.Etcd)
	if storageErr != nil {
		log.Printf("⚠️  Object storage module initialization failed: %v — storage API will be unavailable", storageErr)
	} else {
		storageAPI := router.Group("/api/v1")
		storageSys.RegisterRoutes(storageAPI)
		modules = append(modules, storageSys)
		log.Println("✅ Object Storage module registered (native backend, data:", storageCfg.DataDir, ")")

		// Wire antivirus engine to API builder scanner pipeline.
		if storageSys.AVEngine != nil {
			apiBuilderHandler.SetAVEngine(storageSys.AVEngine)
		}
	}

	// ====================================
	// GATEKEEPER 2FA MODULE
	// ====================================
	gkSystem, gkErr := gatekeeper.NewSystem(conns.PostgreSQL)
	if gkErr != nil {
		log.Printf("⚠️  Gatekeeper 2FA module initialization failed: %v — 2FA endpoints will be unavailable", gkErr)
	} else {
		mfaAPI := router.Group("/api/v1/mfa", authMiddleware)
		gkSystem.RegisterRoutes(mfaAPI)
		modules = append(modules, gkSystem)
		log.Println("✅ Gatekeeper 2FA module registered")
	}

	// Start all registered modules.
	for _, m := range modules {
		if err := m.Start(ctx); err != nil {
			log.Printf("⚠️  Module %s failed to start: %v", m.Name(), err)
		} else {
			log.Printf("✅ Module %s started", m.Name())
		}
	}

	// ====================================
	// AUDIT ENDPOINTS (previously unwired)
	// ====================================
	auditHandler := audit.NewAuditHandler(nil) // AuditLogger impl wired when available
	auditAPI := router.Group("/api/v1/audit", authMiddleware)
	{
		auditAPI.POST("/logs", adminOrSysMiddleware, auditHandler.LogAction)
		auditAPI.GET("/logs", auditHandler.QueryLogs)
		auditAPI.GET("/report", auditHandler.GetReport)
		auditAPI.DELETE("/logs", adminOrSysMiddleware, auditHandler.DeleteOldLogs)
	}
	log.Println("✅ Audit routes registered")

	// ====================================
	// ENCRYPTION ENDPOINTS (previously unwired)
	// ====================================
	// Note: InMemorySecretsManager is used directly; the SecretsManager
	// interface has a signature mismatch that will be unified in a follow-up.
	encryptionHandler := encryption.NewEncryptionHandler(nil)
	encryptionAPI := router.Group("/api/v1/encryption", authMiddleware)
	{
		encryptionAPI.POST("/keys", adminOrSysMiddleware, encryptionHandler.CreateKey)
		encryptionAPI.GET("/keys", encryptionHandler.ListKeys)
		encryptionAPI.GET("/keys/:id", encryptionHandler.GetKey)
		encryptionAPI.POST("/keys/:id/rotate", adminOrSysMiddleware, encryptionHandler.RotateKey)
		encryptionAPI.DELETE("/keys/:id", adminOrSysMiddleware, encryptionHandler.DeleteKey)
		encryptionAPI.POST("/encrypt", authMiddleware, encryptionHandler.Encrypt)
		encryptionAPI.POST("/decrypt", authMiddleware, encryptionHandler.Decrypt)
		encryptionAPI.POST("/policies", adminOrSysMiddleware, encryptionHandler.CreatePolicy)
		encryptionAPI.GET("/policies", encryptionHandler.ListPolicies)
	}
	log.Println("✅ Encryption routes registered")

	// ====================================
	// RECONCILER CONTROLLERS (P2 — AxiomNizam architecture)
	// ====================================
	// Initialize ResourceStore-backed reconcilers for all migrated modules.
	// Each reconciler is started in a background goroutine that periodically
	// reconciles resources from the store.
	//
	// Storage backend is selected by STORAGE_BACKEND env var:
	//   "etcd" (default) — uses EtcdStore[T] backed by external etcd
	//   "raft"           — uses RaftStore[T] backed by embedded Raft + go-memdb
	storageBackend := featureflags.StorageBackend()
	var backendMgr *platformstore.BackendManager

	if storageBackend == "raft" || conns.Etcd != nil {
		// Initialize the backend manager.
		var bmErr error
		backendMgr, bmErr = platformstore.NewBackendManager(platformstore.AllResourceTables())
		if bmErr != nil {
			log.Fatalf("Failed to initialize storage backend: %v", bmErr)
		}
		if backendMgr.IsEtcd() {
			backendMgr.SetEtcdClient(conns.Etcd)
		}
		defer backendMgr.Close()

		// Raft mode: complete deferred initialization now that BackendManager is ready.
		if backendMgr.IsRaft() {
			// Wait for Raft leader election before writing (single-node
			// election typically completes in ~1-2 seconds).
			log.Println("  ⏳ Waiting for Raft leader election...")
			for i := 0; i < 20; i++ {
				if backendMgr.RaftServer.IsLeader() {
					break
				}
				time.Sleep(250 * time.Millisecond)
			}
			if backendMgr.RaftServer.IsLeader() {
				log.Println("  ✅ Raft node is leader")
			} else {
				log.Println("  ⚠️  Raft node is not leader yet (writes may fail until election completes)")
			}

			// Re-attempt JWT secret with KVStore.
			if _, secretErr := ensureSharedDemoJWTSecret(conns.PostgreSQL, nil, backendMgr.KV()); secretErr != nil {
				log.Printf("⚠️  DEMO_JWT_SECRET synchronization via Raft KV failed: %v", secretErr)
			} else {
				log.Println("✅ DEMO_JWT_SECRET synchronized via Raft KV store")
			}

			// Initialize IAM with KVStore backend (if not already initialized).
			if iamSystem == nil {
				iamSystem, iamErr = iampkg.NewSystem(conns.PostgreSQL, nil, iampkg.Config{
					IssuerURL: strings.TrimSpace(os.Getenv("IAM_ISSUER_URL")),
				}, backendMgr.KV())
				if iamErr != nil {
					log.Printf("⚠️  IAM system initialization via Raft KV failed: %v", iamErr)
				} else {
					log.Println("✅ IAM system initialized via Raft KV store")
					// Register IAM routes now (deferred from early init).
					iamSystem.RegisterRoutes(router)
					log.Println("✅ IAM routes registered (deferred, Raft KV backend)")

					// Wire IAM into auth handler.
					if iamSystem.PGStore != nil {
						authHandler.SetIdentityProviderStore(iamSystem.PGStore)
					}
					if iamSystem.Users != nil {
						authHandler.SetIAMUserRepository(iamSystem.Users)
					}
					if iamSystem.Authorizer != nil {
						authHandler.SetIAMAuthorizer(iamSystem.Authorizer)
					}
				}
			}

			// Wire components into storage system (deferred).
			// This must happen even if IAM failed, to ensure bucket persistence works.
			if storageSys != nil {
				// Wire KV persistence first so buckets can be loaded.
				storageSys.SetKVStore(backendMgr.KV())
				log.Println("✅ Storage: Raft KV persistence wired (deferred)")

				// Wire IAM middleware if available.
				if iamSystem != nil && iamSystem.Issuer != nil {
					storageSys.SetIAM(iamSystem.Issuer, iamSystem.RevokedStore)
					log.Println("✅ Storage: IAM middleware attached (deferred, Raft KV backend)")
				}
			}

			// Wire Gatekeeper 2FA module KV persistence.
			if gkSystem != nil {
				gkSystem.SetKVStore(backendMgr.KV())
				log.Println("✅ Gatekeeper: Raft KV persistence wired (deferred)")
			}

			// Wire remaining modules to KV persistence in Raft mode.
			workflows.ConfigureGlobalKVPersistence(backendMgr.KV())
			workflows.GlobalWorkflowEngine.RegisterBuiltinHandlers()
			modes.ConfigureGlobalKVPersistence(backendMgr.KV())
			vectorplus.ConfigureGlobalKVPersistence(backendMgr.KV())
			reviewflow.ConfigureGlobalKVPersistence(backendMgr.KV())
			integration.ConfigureGlobalKVPersistence(backendMgr.KV())
			log.Println("✅ Workflows/Modes/VectorPlus/ReviewFlow/Integration: Raft KV persistence wired")

			log.Println("  ℹ️  Module persistence: Raft KV available via backendMgr.KV()")
		}

		// Wire backend manager to health handler for /distributed Raft status.
		healthHandler.SetBackendManager(backendMgr)

		// Raft cluster management API.
		// Supports two auth modes:
		//   1. Normal JWT admin auth (authMiddleware + adminMiddleware)
		//   2. ADMIN_TOKEN env-var bearer token (for cluster bootstrap when IAM isn't ready)
		if backendMgr.IsRaft() {
			adminToken := os.Getenv("ADMIN_TOKEN")
			raftAuthMiddleware := func(c *gin.Context) {
				// Check ADMIN_TOKEN first (bootstrap mode).
				if adminToken != "" {
					bearer := c.GetHeader("Authorization")
					if bearer == "Bearer "+adminToken {
						c.Next()
						return
					}
				}
				// Fall back to normal JWT admin auth.
				if !authenticateRequest(c) {
					return
				}
				enrichRequestContext(c)
				claims := auth.GetUser(c)
				if claims == nil || !claims.HasRole("admin") {
					c.JSON(http.StatusForbidden, models.Response{Status: "error", Error: "admin role or ADMIN_TOKEN required"})
					c.Abort()
					return
				}
				c.Next()
			}
			raftAPI := router.Group("/api/v1/raft")
			raftAPI.Use(raftAuthMiddleware)
			raftAPI.POST("/peers", func(c *gin.Context) {
				var req struct {
					ID   string `json:"id" binding:"required"`
					Addr string `json:"addr" binding:"required"`
				}
				if err := c.BindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, models.Response{Status: "error", Error: "invalid request"})
					return
				}
				if err := backendMgr.AddRaftPeer(req.ID, req.Addr); err != nil {
					c.JSON(http.StatusInternalServerError, models.Response{Status: "error", Error: fmt.Sprintf("failed to add peer: %v", err)})
					return
				}
				c.JSON(http.StatusOK, models.Response{Status: "ok", Message: fmt.Sprintf("peer %s added at %s", req.ID, req.Addr)})
			})
			raftAPI.DELETE("/peers/:id", func(c *gin.Context) {
				id := c.Param("id")
				if err := backendMgr.RemoveRaftPeer(id); err != nil {
					c.JSON(http.StatusInternalServerError, models.Response{Status: "error", Error: fmt.Sprintf("failed to remove peer: %v", err)})
					return
				}
				c.JSON(http.StatusOK, models.Response{Status: "ok", Message: fmt.Sprintf("peer %s removed", id)})
			})
			log.Println("  ✅ Raft cluster management API registered (/api/v1/raft/peers)")
		}

		log.Printf("🔄 Initializing reconciler controllers (backend=%s)...", storageBackend)
		reconcilerMetrics := metrics.GlobalReconcilerMetrics

		// Phase 1: Shadow mode — reconcilers run but don't affect production.
		shadowMode := true
		if strings.EqualFold(strings.TrimSpace(os.Getenv("RECONCILER_SHADOW_MODE")), "false") {
			shadowMode = false
		}
		if shadowMode {
			log.Println("  ℹ️  Shadow mode ON (set RECONCILER_SHADOW_MODE=false to disable)")
		} else {
			log.Println("  ⚠️  Shadow mode OFF — reconcilers will drive managers")
		}

		// Bulk Operation reconciler
		bulkStore := platformstore.NewStore[*bulk.BulkOperationResource](backendMgr, "bulkoperations", func() *bulk.BulkOperationResource { return &bulk.BulkOperationResource{} })
		bulkReconciler := reconcilerpkg.NewInstrumented("bulk",
			bulk.NewBulkOperationReconciler(bulkStore, platformManagers.Bulk), reconcilerMetrics)
		reconcilerMetrics.Register("bulk")
		go genericctrl.NewGenericController("bulk", bulkStore, bulkReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		bulkHandler.SetDualWriteStore(bulkStore)
		log.Println("  ✅ BulkOperation controller started (dual-write enabled)")

		// EventBus Topic reconciler
		topicStore := platformstore.NewStore[*eventbus.TopicResource](backendMgr, "eventbus-topics", func() *eventbus.TopicResource { return &eventbus.TopicResource{} })
		topicReconciler := reconcilerpkg.NewInstrumented("eventbus-topic",
			eventbus.NewTopicReconciler(topicStore, platformManagers.EventBus), reconcilerMetrics)
		reconcilerMetrics.Register("eventbus-topic")
		go genericctrl.NewGenericController("eventbus-topic", topicStore, topicReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		eventBusHandler.SetTopicDualWriteStore(topicStore)
		log.Println("  ✅ EventBusTopic controller started (dual-write enabled)")

		// EventBus Subscription reconciler
		subscriptionStore := platformstore.NewStore[*eventbus.SubscriptionResource](backendMgr, "eventbus-subscriptions", func() *eventbus.SubscriptionResource { return &eventbus.SubscriptionResource{} })
		subscriptionReconciler := reconcilerpkg.NewInstrumented("eventbus-subscription",
			eventbus.NewSubscriptionReconciler(subscriptionStore, platformManagers.EventBus), reconcilerMetrics)
		reconcilerMetrics.Register("eventbus-subscription")
		go genericctrl.NewGenericController("eventbus-subscription", subscriptionStore, subscriptionReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ EventBusSubscription controller started")

		// Export Job reconciler
		exportStore := platformstore.NewStore[*exportpkg.ExportJobResource](backendMgr, "exportjobs", func() *exportpkg.ExportJobResource { return &exportpkg.ExportJobResource{} })
		exportReconciler := reconcilerpkg.NewInstrumented("export",
			exportpkg.NewExportJobReconciler(exportStore, platformManagers.Export), reconcilerMetrics)
		reconcilerMetrics.Register("export")
		go genericctrl.NewGenericController("export", exportStore, exportReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		exportHandler.SetDualWriteStore(exportStore)
		log.Println("  ✅ ExportJob controller started (dual-write enabled)")

		// Streaming reconciler
		streamStore := platformstore.NewStore[*streaming.StreamResource](backendMgr, "streams", func() *streaming.StreamResource { return &streaming.StreamResource{} })
		streamReconciler := reconcilerpkg.NewInstrumented("streaming",
			streaming.NewStreamReconciler(streamStore), reconcilerMetrics)
		reconcilerMetrics.Register("streaming")
		go genericctrl.NewGenericController("streaming", streamStore, streamReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		streamHandler.SetDualWriteStore(streamStore)
		log.Println("  ✅ Stream controller started (dual-write enabled)")

		// RBAC Role reconciler
		roleStore := platformstore.NewStore[*rbac.RoleResource](backendMgr, "rbac-roles", func() *rbac.RoleResource { return &rbac.RoleResource{} })
		roleReconciler := reconcilerpkg.NewInstrumented("rbac-role",
			rbac.NewRoleReconciler(roleStore, platformManagers.RBAC), reconcilerMetrics)
		reconcilerMetrics.Register("rbac-role")
		go genericctrl.NewGenericController("rbac-role", roleStore, roleReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		rbacHandler.SetRoleDualWriteStore(roleStore)
		log.Println("  ✅ RBAC Role controller started (dual-write enabled)")

		// RBAC RoleBinding reconciler
		roleBindingStore := platformstore.NewStore[*rbac.RoleBindingResource](backendMgr, "rbac-rolebindings", func() *rbac.RoleBindingResource { return &rbac.RoleBindingResource{} })
		roleBindingReconciler := reconcilerpkg.NewInstrumented("rbac-rolebinding",
			rbac.NewRoleBindingReconciler(roleBindingStore, platformManagers.RBAC), reconcilerMetrics)
		reconcilerMetrics.Register("rbac-rolebinding")
		go genericctrl.NewGenericController("rbac-rolebinding", roleBindingStore, roleBindingReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ RBAC RoleBinding controller started")

		// Versioning Policy reconciler
		versionPolicyStore := platformstore.NewStore[*versioning.VersionPolicyResource](backendMgr, "version-policies", func() *versioning.VersionPolicyResource { return &versioning.VersionPolicyResource{} })
		versionPolicyReconciler := reconcilerpkg.NewInstrumented("versioning",
			versioning.NewVersionPolicyReconciler(versionPolicyStore), reconcilerMetrics)
		reconcilerMetrics.Register("versioning")
		go genericctrl.NewGenericController("versioning", versionPolicyStore, versionPolicyReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		versionHandler.SetDualWriteStore(versionPolicyStore) // Phase 2: dual-write
		log.Println("  ✅ VersionPolicy controller started (dual-write enabled)")

		// Tracing Config reconciler
		tracingConfigStore := platformstore.NewStore[*tracing.TracingConfigResource](backendMgr, "tracing-configs", func() *tracing.TracingConfigResource { return &tracing.TracingConfigResource{} })
		tracingConfigReconciler := reconcilerpkg.NewInstrumented("tracing",
			tracing.NewTracingConfigReconciler(tracingConfigStore), reconcilerMetrics)
		reconcilerMetrics.Register("tracing")
		go genericctrl.NewGenericController("tracing", tracingConfigStore, tracingConfigReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		tracingHandler.SetDualWriteStore(tracingConfigStore)
		log.Println("  ✅ TracingConfig controller started (dual-write enabled)")

		// Lineage Node reconciler
		lineageNodeStore := platformstore.NewStore[*lineage.LineageNodeResource](backendMgr, "lineage-nodes", func() *lineage.LineageNodeResource { return &lineage.LineageNodeResource{} })
		lineageNodeReconciler := reconcilerpkg.NewInstrumented("lineage",
			lineage.NewLineageNodeReconciler(lineageNodeStore, platformManagers.Lineage), reconcilerMetrics)
		reconcilerMetrics.Register("lineage")
		go genericctrl.NewGenericController("lineage", lineageNodeStore, lineageNodeReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		lineageHandler.SetDualWriteStore(lineageNodeStore)
		log.Println("  ✅ LineageNode controller started (dual-write enabled)")

		// Audit Policy reconciler
		auditPolicyStore := platformstore.NewStore[*audit.AuditPolicyResource](backendMgr, "audit-policies", func() *audit.AuditPolicyResource { return &audit.AuditPolicyResource{} })
		auditPolicyReconciler := reconcilerpkg.NewInstrumented("audit",
			audit.NewAuditPolicyReconciler(auditPolicyStore), reconcilerMetrics)
		reconcilerMetrics.Register("audit")
		go genericctrl.NewGenericController("audit", auditPolicyStore, auditPolicyReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		auditHandler.SetDualWriteStore(auditPolicyStore)
		log.Println("  ✅ AuditPolicy controller started (dual-write enabled)")

		// Encryption Key reconciler
		encryptionKeyStore := platformstore.NewStore[*encryption.EncryptionKeyResource](backendMgr, "encryption-keys", func() *encryption.EncryptionKeyResource { return &encryption.EncryptionKeyResource{} })
		encryptionKeyReconciler := reconcilerpkg.NewInstrumented("encryption-key",
			encryption.NewEncryptionKeyReconciler(encryptionKeyStore, nil), reconcilerMetrics)
		reconcilerMetrics.Register("encryption-key")
		go genericctrl.NewGenericController("encryption-key", encryptionKeyStore, encryptionKeyReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		encryptionHandler.SetKeyDualWriteStore(encryptionKeyStore)
		log.Println("  ✅ EncryptionKey controller started (dual-write enabled)")

		// Encryption Policy reconciler
		encryptionPolicyStore := platformstore.NewStore[*encryption.EncryptionPolicyResource](backendMgr, "encryption-policies", func() *encryption.EncryptionPolicyResource { return &encryption.EncryptionPolicyResource{} })
		encryptionPolicyReconciler := reconcilerpkg.NewInstrumented("encryption-policy",
			encryption.NewEncryptionPolicyReconciler(encryptionPolicyStore), reconcilerMetrics)
		reconcilerMetrics.Register("encryption-policy")
		go genericctrl.NewGenericController("encryption-policy", encryptionPolicyStore, encryptionPolicyReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ EncryptionPolicy controller started")

		// Conductor Producer reconciler
		producerStore := platformstore.NewStore[*conductor.ProducerResource](backendMgr, "conductor-producers", func() *conductor.ProducerResource { return &conductor.ProducerResource{} })
		producerReconciler := reconcilerpkg.NewInstrumented("conductor-producer",
			conductor.NewProducerReconciler(producerStore, conductorMgr), reconcilerMetrics)
		reconcilerMetrics.Register("conductor-producer")
		go genericctrl.NewGenericController("conductor-producer", producerStore, producerReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		// Note: conductor Handler is created inside RegisterRoutes — dual-write store
		// will be wired when conductor handler is refactored to accept store injection.
		log.Println("  ✅ ConductorProducer controller started (dual-write pending handler refactor)")

		// Conductor Consumer reconciler
		consumerStore := platformstore.NewStore[*conductor.ConsumerResource](backendMgr, "conductor-consumers", func() *conductor.ConsumerResource { return &conductor.ConsumerResource{} })
		consumerReconciler := reconcilerpkg.NewInstrumented("conductor-consumer",
			conductor.NewConsumerReconciler(consumerStore, conductorMgr), reconcilerMetrics)
		reconcilerMetrics.Register("conductor-consumer")
		go genericctrl.NewGenericController("conductor-consumer", consumerStore, consumerReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ ConductorConsumer controller started")

		// Webhook reconciler
		webhookStore := platformstore.NewStore[*webhooks.WebhookResource](backendMgr, "webhooks", func() *webhooks.WebhookResource { return &webhooks.WebhookResource{} })
		webhookReconciler := reconcilerpkg.NewInstrumented("webhook",
			webhooks.NewWebhookReconciler(webhookStore), reconcilerMetrics)
		reconcilerMetrics.Register("webhook")
		go genericctrl.NewGenericController("webhook", webhookStore, webhookReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		webhookHandler.SetDualWriteStore(webhookStore)
		log.Println("  ✅ Webhook controller started (dual-write enabled)")

		// Tenant reconciler
		tenantStore := platformstore.NewStore[*tenant.TenantV1Resource](backendMgr, "tenants", func() *tenant.TenantV1Resource { return &tenant.TenantV1Resource{} })
		tenantReconciler := reconcilerpkg.NewInstrumented("tenant",
			tenant.NewTenantReconciler(tenantStore), reconcilerMetrics)
		reconcilerMetrics.Register("tenant")
		go genericctrl.NewGenericController("tenant", tenantStore, tenantReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		tenantHandler.SetDualWriteStore(tenantStore)
		log.Println("  ✅ Tenant controller started (dual-write enabled)")

		log.Printf("🔄 All 17 reconciler controllers RUNNING in %d goroutines (shadow=%v) — Phase 1 active", 17, shadowMode)

		// ====================================
		// PHASE 5: Wire remaining reconcilers
		// ====================================

		// Jobs reconciler
		jobsStore := platformstore.NewStore[*jobs.JobResource](backendMgr, "jobs", func() *jobs.JobResource { return &jobs.JobResource{} })
		jobsReconciler := reconcilerpkg.NewInstrumented("jobs",
			jobs.NewJobController(nil, nil), reconcilerMetrics)
		reconcilerMetrics.Register("jobs")
		go genericctrl.NewGenericController("jobs", jobsStore, jobsReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ Jobs controller started")

		// ETL Pipeline reconciler
		etlStore := platformstore.NewStore[*etl.PipelineResource](backendMgr, "etl-pipelines", func() *etl.PipelineResource { return &etl.PipelineResource{} })
		etlReconciler := reconcilerpkg.NewInstrumented("etl",
			etl.NewPipelineController(nil, nil), reconcilerMetrics)
		reconcilerMetrics.Register("etl")
		go genericctrl.NewGenericController("etl", etlStore, etlReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ ETL Pipeline controller started")

		// CDC Pipeline reconciler
		cdcStore := platformstore.NewStore[*cdc.CDCPipelineResource](backendMgr, "cdc-pipelines", func() *cdc.CDCPipelineResource { return &cdc.CDCPipelineResource{} })
		cdcReconciler := reconcilerpkg.NewInstrumented("cdc",
			cdc.NewCDCPipelineController(nil, nil), reconcilerMetrics)
		reconcilerMetrics.Register("cdc")
		go genericctrl.NewGenericController("cdc", cdcStore, cdcReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ CDC Pipeline controller started")

		// Policies reconciler
		policiesStore := platformstore.NewStore[*policies.PolicyResource](backendMgr, "policies", func() *policies.PolicyResource { return &policies.PolicyResource{} })
		policiesReconciler := reconcilerpkg.NewInstrumented("policies",
			policies.NewPolicyReconciler(policiesStore, nil), reconcilerMetrics)
		reconcilerMetrics.Register("policies")
		go genericctrl.NewGenericController("policies", policiesStore, policiesReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ Policies controller started")

		// DataSource reconciler
		datasourceStore := platformstore.NewStore[*datasourceresource.DataSourceV1Resource](backendMgr, "datasources", func() *datasourceresource.DataSourceV1Resource { return &datasourceresource.DataSourceV1Resource{} })
		datasourceReconciler := reconcilerpkg.NewInstrumented("datasource",
			datasourceresource.NewDataSourceReconciler(datasourceStore, nil), reconcilerMetrics)
		reconcilerMetrics.Register("datasource")
		go genericctrl.NewGenericController("datasource", datasourceStore, datasourceReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ DataSource controller started")

		// IAM Users reconciler
		iamUsersStore := platformstore.NewStore[*iamusers.UserResource](backendMgr, "iam-users", func() *iamusers.UserResource { return &iamusers.UserResource{} })
		iamUsersReconciler := reconcilerpkg.NewInstrumented("iam-users",
			iamusers.NewUserReconciler(iamUsersStore), reconcilerMetrics)
		reconcilerMetrics.Register("iam-users")
		go genericctrl.NewGenericController("iam-users", iamUsersStore, iamUsersReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ IAM Users controller started")

		// API Scanner reconciler
		apiScannerStore := platformstore.NewStore[*apiscanner.APIScanResource](backendMgr, "api-scans", func() *apiscanner.APIScanResource { return &apiscanner.APIScanResource{} })
		apiScannerReconciler := reconcilerpkg.NewInstrumented("apiscanner",
			apiscanner.NewAPIScanReconciler(apiScannerStore, nil), reconcilerMetrics)
		reconcilerMetrics.Register("apiscanner")
		go genericctrl.NewGenericController("apiscanner", apiScannerStore, apiScannerReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ API Scanner controller started")

		log.Printf("🔄 Phase 5: +7 reconciler controllers started (total: 24 controllers, shadow=%v)", shadowMode)

		// ====================================
		// PHASE 6 P2: GIS resource controller
		// ====================================
		gisStore := platformstore.NewStore[*gispkg.GISResource](backendMgr, "gis", func() *gispkg.GISResource { return &gispkg.GISResource{} })
		gisReconciler := reconcilerpkg.NewInstrumented("gis",
			gispkg.NewGISReconciler(gisStore), reconcilerMetrics)
		reconcilerMetrics.Register("gis")
		go genericctrl.NewGenericController("gis", gisStore, gisReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ GIS controller started (Phase 6 P2)")

		// Analytics Dashboard controller
		analyticsStore := platformstore.NewStore[*analyticspkg.DashboardResource](backendMgr, "analytics-dashboards", func() *analyticspkg.DashboardResource { return &analyticspkg.DashboardResource{} })
		analyticsReconciler := reconcilerpkg.NewInstrumented("analytics",
			analyticspkg.NewDashboardReconciler(analyticsStore), reconcilerMetrics)
		reconcilerMetrics.Register("analytics")
		go genericctrl.NewGenericController("analytics", analyticsStore, analyticsReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ Analytics Dashboard controller started (Phase 6 P2)")

		// Transform Rule controller
		transformStore := platformstore.NewStore[*transformpkg.RuleResource](backendMgr, "transform-rules", func() *transformpkg.RuleResource { return &transformpkg.RuleResource{} })
		transformReconciler := reconcilerpkg.NewInstrumented("transform",
			transformpkg.NewRuleReconciler(transformStore), reconcilerMetrics)
		reconcilerMetrics.Register("transform")
		go genericctrl.NewGenericController("transform", transformStore, transformReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ Transform Rule controller started (Phase 6 P2)")

		// Notification Channel controller
		notificationStore := platformstore.NewStore[*notificationpkg.ChannelResource](backendMgr, "notification-channels", func() *notificationpkg.ChannelResource { return &notificationpkg.ChannelResource{} })
		notificationReconciler := reconcilerpkg.NewInstrumented("notification",
			notificationpkg.NewChannelReconciler(notificationStore), reconcilerMetrics)
		reconcilerMetrics.Register("notification")
		go genericctrl.NewGenericController("notification", notificationStore, notificationReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ Notification Channel controller started (Phase 6 P2)")

		// NetIntel Config controller
		netintelStore := platformstore.NewStore[*netintelpkg.ConfigResource](backendMgr, "netintel-configs", func() *netintelpkg.ConfigResource { return &netintelpkg.ConfigResource{} })
		netintelReconciler := reconcilerpkg.NewInstrumented("netintel",
			netintelpkg.NewConfigReconciler(netintelStore), reconcilerMetrics)
		reconcilerMetrics.Register("netintel")
		go genericctrl.NewGenericController("netintel", netintelStore, netintelReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ NetIntel Config controller started (Phase 6 P2)")

		log.Println("🔄 Phase 6 P2: +5 controllers started (gis, analytics, transform, notification, netintel)")

		// APIBank reconciler
		apiBankReconcilerStore := platformstore.NewStore[*apibanks.APIBankResource](backendMgr, "apibanks", func() *apibanks.APIBankResource { return &apibanks.APIBankResource{} })
		apiBankReconciler := reconcilerpkg.NewInstrumented("apibanks",
			apibanks.NewAPIBankReconciler(apiBankReconcilerStore, nil), reconcilerMetrics)
		reconcilerMetrics.Register("apibanks")
		go genericctrl.NewGenericController("apibanks", apiBankReconcilerStore, apiBankReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
		log.Println("  ✅ APIBank controller started")

		// Phase 0.4: etcd key-space monitoring
		etcdPrefixes := []string{
			"/axiomnizam/bulkoperations/",
			"/axiomnizam/eventbus-topics/",
			"/axiomnizam/eventbus-subscriptions/",
			"/axiomnizam/exportjobs/",
			"/axiomnizam/streams/",
			"/axiomnizam/rbac-roles/",
			"/axiomnizam/rbac-rolebindings/",
			"/axiomnizam/version-policies/",
			"/axiomnizam/tracing-configs/",
			"/axiomnizam/lineage-nodes/",
			"/axiomnizam/audit-policies/",
			"/axiomnizam/encryption-keys/",
			"/axiomnizam/encryption-policies/",
			"/axiomnizam/conductor-producers/",
			"/axiomnizam/conductor-consumers/",
			"/axiomnizam/webhooks/",
			"/axiomnizam/tenants/",
			"/axiomnizam/apibanks/",
			"/axiomnizam/jobs/",
			"/axiomnizam/etl-pipelines/",
			"/axiomnizam/cdc-pipelines/",
			"/axiomnizam/policies/",
			"/axiomnizam/datasources/",
			"/axiomnizam/iam-users/",
			"/axiomnizam/api-scans/",
			"/axiomnizam/gis/",
			"/axiomnizam/analytics-dashboards/",
			"/axiomnizam/transform-rules/",
			"/axiomnizam/notification-channels/",
			"/axiomnizam/netintel-configs/",
		}
		keySpaceMonitor := metrics.NewEtcdKeySpaceMonitor(conns.Etcd, etcdPrefixes, 30*time.Second)
		keySpaceMonitor.Start(ctx)
		keySpaceMonitorRef = keySpaceMonitor
		log.Println("  ✅ etcd key-space monitor started (18 prefixes, 30s interval)")
	} else {
		log.Println("⚠️  etcd not available — reconciler controllers skipped")
	}

	// ====================================
	// MIGRATIONS (previously unwired)
	// ====================================
	if conns.PostgreSQL != nil {
		if migrationErr := migrations.RunMigrations(conns.PostgreSQL); migrationErr != nil {
			log.Printf("⚠️  Database migrations failed: %v", migrationErr)
		} else {
			log.Println("✅ Database migrations completed successfully")
		}
	}

	// ====================================
	// HEARTBEAT TRACKER (previously unwired)
	// ====================================
	heartbeatTracker := heartbeat.New(func(id string) {
		log.Printf("⚠️  Heartbeat expired for entity: %s", id)
	})
	heartbeatTracker.ReapInterval = 5 * time.Second
	heartbeatTracker.Start()
	log.Println("✅ Heartbeat tracker started")

	// ====================================
	// SERVICE REGISTRY (previously unwired)
	// ====================================
	svcRegistry := serviceregistry.New()
	log.Println("✅ Service registry started")

	// ====================================
	// AUTOPILOT (previously unwired)
	// ====================================
	autopilotInstance := autopilot.New(autopilot.Config{
		MaxTrailingLogs:      250,
		LastContactThreshold: 200 * time.Millisecond,
		DeadServerCleanup:    true,
		MinQuorum:            3,
	})
	_ = autopilotInstance // available for cluster health evaluation
	log.Println("✅ Autopilot initialized")

	// ====================================
	// TRIVY VULNERABILITY SCANNER (previously unwired)
	// ====================================
	trivyBinaryPath := strings.TrimSpace(os.Getenv("TRIVY_BINARY_PATH"))
	if trivyBinaryPath == "" {
		trivyBinaryPath = "trivy"
	}
	trivyEngine := trivy.NewEngine(trivyBinaryPath)
	log.Printf("✅ Trivy vulnerability scanner initialized (binary: %s)", trivyBinaryPath)

	// ====================================
	// API BANKS (previously unwired)
	// ====================================
	apiBankManager := apibanks.NewAPIBankManager()
	apiBankCatalog := apibanks.NewAPIBankCatalog(apiBankManager)
	_ = apiBankCatalog // available for API discovery

	log.Println("✅ API Banks module initialized")

	// ====================================
	// API BANKS ROUTES
	// ====================================
	apiBankAPI := router.Group("/api/v1/apibanks", authMiddleware)
	{
		apiBankAPI.POST("", adminOrSysMiddleware, func(c *gin.Context) {
			var bank apibanks.APIBank
			if err := c.ShouldBindJSON(&bank); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := apiBankManager.CreateBank(c.Request.Context(), &bank); err != nil {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"message": "bank created", "bank": bank})
		})
		apiBankAPI.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"banks": apiBankManager.ListBanks()})
		})
		apiBankAPI.GET("/:name", func(c *gin.Context) {
			bank := apiBankManager.GetBank(strings.TrimSpace(c.Param("name")))
			if bank == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "bank not found"})
				return
			}
			c.JSON(http.StatusOK, bank)
		})
		apiBankAPI.POST("/:name/apis", adminOrSysMiddleware, func(c *gin.Context) {
			var api apibanks.APIReference
			if err := c.ShouldBindJSON(&api); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := apiBankManager.AddAPIToBank(c.Request.Context(), strings.TrimSpace(c.Param("name")), api); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "API added to bank"})
		})
		apiBankAPI.DELETE("/:name/apis/:apiName", adminOrSysMiddleware, func(c *gin.Context) {
			if err := apiBankManager.RemoveAPIFromBank(c.Request.Context(), strings.TrimSpace(c.Param("name")), strings.TrimSpace(c.Param("apiName"))); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "API removed from bank"})
		})
		apiBankAPI.GET("/search/data-class", func(c *gin.Context) {
			dataClass := strings.TrimSpace(c.Query("class"))
			c.JSON(http.StatusOK, gin.H{"apis": apiBankCatalog.SearchByDataClass(dataClass)})
		})
		apiBankAPI.GET("/search/owner", func(c *gin.Context) {
			owner := strings.TrimSpace(c.Query("owner"))
			c.JSON(http.StatusOK, gin.H{"banks": apiBankCatalog.SearchByOwner(owner)})
		})
		apiBankAPI.GET("/search/tag", func(c *gin.Context) {
			tag := strings.TrimSpace(c.Query("tag"))
			c.JSON(http.StatusOK, gin.H{"banks": apiBankCatalog.SearchByTag(tag)})
		})
	}

	// ====================================
	// TRIVY SCANNER ROUTES
	// ====================================
	trivyAPI := router.Group("/api/v1/trivy", authMiddleware)
	{
		trivyAPI.POST("/scan", adminOrSysMiddleware, func(c *gin.Context) {
			var req trivy.ScanRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			req.UseExternal = true
			result, err := trivyEngine.Scan(c.Request.Context(), req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
		})
	}

	// ====================================
	// DEPLOYMENT CONTROLLER ROUTES
	// ====================================
	deploymentControllers := make(map[string]*deployment.Controller)
	var deploymentMu sync.Mutex

	deploymentAPI := router.Group("/api/v1/deployments", authMiddleware)
	{
		deploymentAPI.POST("", adminOrSysMiddleware, func(c *gin.Context) {
			var spec deployment.Spec
			if err := c.ShouldBindJSON(&spec); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if strings.TrimSpace(spec.JobID) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "jobId is required"})
				return
			}
			deploymentMu.Lock()
			ctrl := deployment.NewController(spec)
			deploymentControllers[spec.JobID] = ctrl
			deploymentMu.Unlock()
			c.JSON(http.StatusCreated, gin.H{"message": "deployment created", "jobId": spec.JobID, "state": ctrl.State()})
		})
		deploymentAPI.GET("/:jobId", func(c *gin.Context) {
			jobID := strings.TrimSpace(c.Param("jobId"))
			deploymentMu.Lock()
			ctrl, ok := deploymentControllers[jobID]
			deploymentMu.Unlock()
			if !ok {
				c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
				return
			}
			c.JSON(http.StatusOK, ctrl.State())
		})
		deploymentAPI.POST("/:jobId/promote", adminOrSysMiddleware, func(c *gin.Context) {
			jobID := strings.TrimSpace(c.Param("jobId"))
			deploymentMu.Lock()
			ctrl, ok := deploymentControllers[jobID]
			deploymentMu.Unlock()
			if !ok {
				c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
				return
			}
			if !ctrl.Promote() {
				c.JSON(http.StatusConflict, gin.H{"error": "promotion not available — canaries may not be healthy or deployment not in running phase"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "deployment promoted", "state": ctrl.State()})
		})
		deploymentAPI.POST("/:jobId/fail", adminOrSysMiddleware, func(c *gin.Context) {
			jobID := strings.TrimSpace(c.Param("jobId"))
			var body struct {
				Reason string `json:"reason"`
			}
			_ = c.ShouldBindJSON(&body)
			if strings.TrimSpace(body.Reason) == "" {
				body.Reason = "manual rollback"
			}
			deploymentMu.Lock()
			ctrl, ok := deploymentControllers[jobID]
			deploymentMu.Unlock()
			if !ok {
				c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
				return
			}
			decision := ctrl.Fail(body.Reason)
			c.JSON(http.StatusOK, gin.H{"message": "deployment failed", "decision": decision, "state": ctrl.State()})
		})
	}

	// ====================================
	// SERVICE REGISTRY ROUTES
	// ====================================
	svcRegistryAPI := router.Group("/api/v1/service-registry", authMiddleware)
	{
		svcRegistryAPI.POST("/services", adminOrSysMiddleware, func(c *gin.Context) {
			var svc serviceregistry.Service
			if err := c.ShouldBindJSON(&svc); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if strings.TrimSpace(svc.ID) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "service id is required"})
				return
			}
			if svc.Checks == nil {
				svc.Checks = make(map[string]*serviceregistry.Check)
			}
			svcRegistry.Register(&svc)
			c.JSON(http.StatusCreated, gin.H{"message": "service registered", "id": svc.ID})
		})
		svcRegistryAPI.DELETE("/services/:id", adminOrSysMiddleware, func(c *gin.Context) {
			svcRegistry.Deregister(strings.TrimSpace(c.Param("id")))
			c.JSON(http.StatusOK, gin.H{"message": "service deregistered"})
		})
		svcRegistryAPI.GET("/services/:id", func(c *gin.Context) {
			svc, ok := svcRegistry.Get(strings.TrimSpace(c.Param("id")))
			if !ok {
				c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"service": svc, "status": svc.Rollup()})
		})
		svcRegistryAPI.GET("/services", func(c *gin.Context) {
			name := strings.TrimSpace(c.Query("name"))
			if name != "" {
				c.JSON(http.StatusOK, gin.H{"services": svcRegistry.ByName(name)})
				return
			}
			c.JSON(http.StatusOK, gin.H{"services": svcRegistry.ByName("")})
		})
		svcRegistryAPI.PUT("/services/:id/checks/:checkId", adminOrSysMiddleware, func(c *gin.Context) {
			var body struct {
				Status string `json:"status"`
				Notes  string `json:"notes"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := svcRegistry.UpdateCheck(strings.TrimSpace(c.Param("id")), strings.TrimSpace(c.Param("checkId")), serviceregistry.Status(body.Status), body.Notes); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "check updated"})
		})
	}

	// ====================================
	// HEARTBEAT ROUTES
	// ====================================
	heartbeatAPI := router.Group("/api/v1/heartbeat", authMiddleware)
	{
		heartbeatAPI.POST("/beat", func(c *gin.Context) {
			var body struct {
				ID  string `json:"id"`
				TTL int    `json:"ttl"` // seconds
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if strings.TrimSpace(body.ID) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
				return
			}
			ttl := time.Duration(body.TTL) * time.Second
			if ttl <= 0 {
				ttl = 30 * time.Second
			}
			heartbeatTracker.Beat(body.ID, ttl)
			c.JSON(http.StatusOK, gin.H{"message": "heartbeat recorded", "id": body.ID, "ttl_seconds": int(ttl.Seconds())})
		})
		heartbeatAPI.GET("/alive/:id", func(c *gin.Context) {
			id := strings.TrimSpace(c.Param("id"))
			c.JSON(http.StatusOK, gin.H{"id": id, "alive": heartbeatTracker.IsAlive(id)})
		})
		heartbeatAPI.GET("/expired", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"expired": heartbeatTracker.Expired()})
		})
		heartbeatAPI.DELETE("/:id", adminOrSysMiddleware, func(c *gin.Context) {
			heartbeatTracker.Delete(strings.TrimSpace(c.Param("id")))
			c.JSON(http.StatusOK, gin.H{"message": "heartbeat entry deleted"})
		})
	}

	// ====================================
	// AUTOPILOT ROUTES
	// ====================================
	router.POST("/api/v1/autopilot/evaluate", adminOrSysMiddleware, func(c *gin.Context) {
		var body struct {
			Peers       []autopilot.Server `json:"peers"`
			LeaderIndex uint64             `json:"leaderIndex"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		decisions := autopilotInstance.Evaluate(c.Request.Context(), body.Peers, body.LeaderIndex)
		c.JSON(http.StatusOK, gin.H{"decisions": decisions})
	})

	// ====================================
	// ANTIVIRUS MANAGEMENT API (uses existing engine from storage module)
	// ====================================
	if storageSys != nil && storageSys.AVEngine != nil {
		avHandler := antivirus.NewAPIHandler(storageSys.AVEngine)
		avHandler.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ Antivirus management API registered (reusing storage engine)")
	} else {
		log.Println("⚠️  Antivirus engine not available — management API skipped")
	}

	// ====================================
	// RATE LIMITING (previously unwired)
	// ====================================
	quotaMgr := ratelimit.NewQuotaManager()
	rlMiddleware := ratelimit.NewRateLimitMiddleware(quotaMgr)
	quotaHandler := ratelimit.NewQuotaHandler(quotaMgr)
	router.Use(rlMiddleware.Handler())
	quotaAPI := router.Group("/api/v1/quotas", authMiddleware)
	{
		quotaAPI.GET("", quotaHandler.ListQuotas)
		quotaAPI.GET("/:userID", quotaHandler.GetQuota)
		quotaAPI.POST("/:userID", adminOrSysMiddleware, quotaHandler.SetUserQuota)
		quotaAPI.DELETE("/:userID", adminOrSysMiddleware, quotaHandler.ResetQuota)
		quotaAPI.POST("/endpoint", adminOrSysMiddleware, quotaHandler.SetEndpointLimit)
	}
	log.Println("✅ Rate limiting module started")

	// ====================================
	// STREAM BROKER (previously unwired)
	// ====================================
	streamBroker := stream.NewBroker(10000)
	streamHTTP := stream.HTTPHandler(streamBroker)
	router.Any("/api/v1/stream", func(c *gin.Context) {
		streamHTTP.ServeHTTP(c.Writer, c.Request)
	})
	log.Println("✅ Stream broker started")

	// ====================================
	// FEATURE MODULES (storage-backed)
	// ====================================
	if backendMgr != nil {
		// Alerting
		alertRuleStore := platformstore.NewStore[*alertingmodels.AlertRuleResource](backendMgr, "alert-rules", func() *alertingmodels.AlertRuleResource { return &alertingmodels.AlertRuleResource{} })
		alertIncidentStore := platformstore.NewStore[*alertingmodels.AlertIncidentResource](backendMgr, "alert-incidents", func() *alertingmodels.AlertIncidentResource { return &alertingmodels.AlertIncidentResource{} })
		alertChannelStore := platformstore.NewStore[*alertingmodels.NotificationChannelResource](backendMgr, "alert-channels", func() *alertingmodels.NotificationChannelResource { return &alertingmodels.NotificationChannelResource{} })
		alertHandlers := alerting.NewAlertHandlers(alertRuleStore, alertIncidentStore, alertChannelStore)
		alertHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ Alerting module started")

		// Governance
		complianceStore := platformstore.NewStore[*governancemodels.CompliancePolicyResource](backendMgr, "compliance-policies", func() *governancemodels.CompliancePolicyResource { return &governancemodels.CompliancePolicyResource{} })
		retentionStore := platformstore.NewStore[*governancemodels.RetentionPolicyResource](backendMgr, "retention-policies", func() *governancemodels.RetentionPolicyResource { return &governancemodels.RetentionPolicyResource{} })
		accessReqStore := platformstore.NewStore[*governancemodels.AccessRequestResource](backendMgr, "access-requests", func() *governancemodels.AccessRequestResource { return &governancemodels.AccessRequestResource{} })
		govHandlers := governance.NewGovernanceHandlers(complianceStore, retentionStore, accessReqStore)
		govHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ Governance module started")

		// SLO
		sloStore := platformstore.NewStore[*slo.SLOResource](backendMgr, "slos", func() *slo.SLOResource { return &slo.SLOResource{} })
		sloHandlers := slo.NewSLOHandlers(sloStore)
		sloHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ SLO module started")

		// Catalog
		assetStore := platformstore.NewStore[*catalog.CatalogAssetResource](backendMgr, "catalog-assets", func() *catalog.CatalogAssetResource { return &catalog.CatalogAssetResource{} })
		collectionStore := platformstore.NewStore[*catalog.CatalogCollectionResource](backendMgr, "catalog-collections", func() *catalog.CatalogCollectionResource { return &catalog.CatalogCollectionResource{} })
		catalogHandlers := catalog.NewCatalogHandlers(assetStore, collectionStore, nil)
		catalogHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ Catalog module started")

		// Costing
		costPolicyStore := platformstore.NewStore[*costing.CostPolicyResource](backendMgr, "cost-policies", func() *costing.CostPolicyResource { return &costing.CostPolicyResource{} })
		usageStore := platformstore.NewStore[*costing.UsageRecordResource](backendMgr, "usage-records", func() *costing.UsageRecordResource { return &costing.UsageRecordResource{} })
		costHandlers := costing.NewCostHandlers(costPolicyStore, usageStore)
		costHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ Costing module started")

		// Contracts
		contractStore := platformstore.NewStore[*contracts.DataContractResource](backendMgr, "data-contracts", func() *contracts.DataContractResource { return &contracts.DataContractResource{} })
		contractHandlers := contracts.NewContractHandlers(contractStore)
		contractHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ Contracts module started")

		// Schema Registry
		schemaStore := platformstore.NewStore[*schemaregistry.SchemaResource](backendMgr, "schemas", func() *schemaregistry.SchemaResource { return &schemaregistry.SchemaResource{} })
		subjectStore := platformstore.NewStore[*schemaregistry.SchemaSubjectResource](backendMgr, "schema-subjects", func() *schemaregistry.SchemaSubjectResource { return &schemaregistry.SchemaSubjectResource{} })
		schemaChecker := schemaregistry.NewJSONSchemaCompatibilityChecker()
		schemaHandlers := schemaregistry.NewSchemaRegistryHandlers(schemaStore, subjectStore, schemaChecker)
		schemaHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ Schema Registry module started")

		// Anonymization
		anonymPolicyStore := platformstore.NewStore[*anonymization.AnonymizationPolicyResource](backendMgr, "anonymization-policies", func() *anonymization.AnonymizationPolicyResource { return &anonymization.AnonymizationPolicyResource{} })
		anonymHandlers := anonymization.NewAnonymizationHandlers(anonymPolicyStore)
		anonymHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ Anonymization module started")

		// Stream Analytics
		streamJobStore := platformstore.NewStore[*streamanalytics.StreamJobResource](backendMgr, "stream-jobs", func() *streamanalytics.StreamJobResource { return &streamanalytics.StreamJobResource{} })
		streamHandlers := streamanalytics.NewStreamAnalyticsHandlers(streamJobStore)
		streamHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ Stream Analytics module started")

		// Feature Store
		featureGroupStore := platformstore.NewStore[*featurestore.FeatureGroupResource](backendMgr, "feature-groups", func() *featurestore.FeatureGroupResource { return &featurestore.FeatureGroupResource{} })
		featureHandlers := featurestore.NewFeatureStoreHandlers(featureGroupStore)
		featureHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ Feature Store module started")

		// Federation
		vtStore := platformstore.NewStore[*federation.VirtualTableResource](backendMgr, "virtual-tables", func() *federation.VirtualTableResource { return &federation.VirtualTableResource{} })
		fedQueryStore := platformstore.NewStore[*federation.FederatedQueryResource](backendMgr, "federated-queries", func() *federation.FederatedQueryResource { return &federation.FederatedQueryResource{} })
		fedHandlers := federation.NewFederationHandlers(vtStore, fedQueryStore, nil)
		fedHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ Federation module started")

		// ML Pipeline
		mlPipelineStore := platformstore.NewStore[*mlpipeline.MLPipelineResource](backendMgr, "ml-pipelines", func() *mlpipeline.MLPipelineResource { return &mlpipeline.MLPipelineResource{} })
		modelDeployStore := platformstore.NewStore[*mlpipeline.ModelDeploymentResource](backendMgr, "model-deployments", func() *mlpipeline.ModelDeploymentResource { return &mlpipeline.ModelDeploymentResource{} })
		mlHandlers := mlpipeline.NewMLPipelineHandlers(mlPipelineStore, modelDeployStore)
		mlHandlers.RegisterRoutes(router.Group("/api/v1", authMiddleware))
		log.Println("✅ ML Pipeline module started")
	} else {
		log.Println("⚠️  Storage backend not available — feature modules (alerting, governance, slo, catalog, costing, contracts, schemaregistry, anonymization, streamanalytics, featurestore, federation, mlpipeline) skipped")
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

	// Stop all registered modules (reverse order).
	for i := len(modules) - 1; i >= 0; i-- {
		if err := modules[i].Stop(); err != nil {
			log.Printf("⚠️  Module %s stop error: %v", modules[i].Name(), err)
		} else {
			log.Printf("✅ Module %s stopped", modules[i].Name())
		}
	}

	// Flush conductor stats to DB before exit
	conductorMgr.Close()

	// Stop heartbeat tracker
	heartbeatTracker.Stop()

	// Stop service registry
	svcRegistry.Close()

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

func ensureSharedDemoJWTSecret(pg *gorm.DB, etcd *clientv3.Client, kv platformstore.KVStore) (string, error) {
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
		log.Printf("⚠️  postgres bootstrap for DEMO_JWT_SECRET failed, falling back to KV store: %v", err)
	}

	resolved, err := ensureSharedDemoJWTSecretFromKV(kv, etcd)
	if err != nil {
		return "", err
	}
	if err := os.Setenv("DEMO_JWT_SECRET", resolved); err != nil {
		return "", fmt.Errorf("setting DEMO_JWT_SECRET from KV store: %w", err)
	}
	auth.SetDemoJWTSecret(resolved)
	return resolved, nil
}

// ensureSharedDemoJWTSecretFromKV uses the KVStore abstraction (works
// with both etcd and Raft backends).  Falls back to raw etcd if KV is nil.
func ensureSharedDemoJWTSecretFromKV(kv platformstore.KVStore, etcd *clientv3.Client) (string, error) {
	if kv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		candidate, err := generateBootstrapSecret(48)
		if err != nil {
			return "", err
		}

		resolved, _, err := kv.CAS(ctx, demoJWTSecretEtcdKey, candidate)
		if err != nil {
			return "", fmt.Errorf("persisting demo token secret via KV store: %w", err)
		}
		if resolved == "" {
			return "", fmt.Errorf("demo token secret CAS returned empty value")
		}
		return resolved, nil
	}

	// Fallback to raw etcd client.
	return ensureSharedDemoJWTSecretFromEtcd(etcd)
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

func ensureWorkflowRegistered(ctx context.Context, resourceHandler *resourcespkg.GenericResourceHandler, workflowName string) error {
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

func workflowFromResource(resource *resourcespkg.GenericResource) (*workflows.Workflow, error) {
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
