package main

import (
	"fmt"
	"log"

	"example.com/axiomnizam/internal/auth"
	"example.com/axiomnizam/internal/config"
	"example.com/axiomnizam/internal/database"
	"example.com/axiomnizam/internal/handlers"
	"example.com/axiomnizam/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}
	fmt.Println("🚀 Starting AxiomNizam API Server...\n")

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

	// Health check endpoints (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/status", healthHandler.Status)

	// Apply auth middleware to protected routes
	var authMiddleware gin.HandlerFunc
	if tokenValidator != nil {
		authMiddleware = auth.Middleware(tokenValidator)
	} else {
		authMiddleware = func(c *gin.Context) { c.Next() }
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
	// ADMIN OPERATIONS (Admin Only)
	// ====================================

	// Database management endpoints (admin only)
	router.POST("/api/admin/database/create", adminMiddleware, adminHandler.CreateDatabase)
	router.GET("/api/admin/database/list", adminMiddleware, adminHandler.ListDatabases)

	// Table management endpoints (admin only)
	router.POST("/api/admin/table/create", adminMiddleware, adminHandler.CreateTable)
	router.GET("/api/admin/table/list", adminMiddleware, adminHandler.ListTables)

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

	router.Run(fmt.Sprintf("%s:%s", apiHost, apiPort))
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
