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

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(conns)
	userHandler := handlers.NewUserHandler(conns.MySQL)

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

	// CRUD routes for MySQL (protected by auth)
	router.POST("/api/mysql/users", authMiddleware, userHandler.CreateUser)
	router.GET("/api/mysql/users", authMiddleware, userHandler.GetAllUsers)
	router.GET("/api/mysql/users/:id", authMiddleware, userHandler.GetUserByID)
	router.PUT("/api/mysql/users/:id", authMiddleware, userHandler.UpdateUser)
	router.DELETE("/api/mysql/users/:id", authMiddleware, userHandler.DeleteUser)

	// CRUD routes for MariaDB (protected by auth)
	mariadbHandler := handlers.NewUserHandler(conns.MariaDB)
	router.POST("/api/mariadb/users", authMiddleware, mariadbHandler.CreateUser)
	router.GET("/api/mariadb/users", authMiddleware, mariadbHandler.GetAllUsers)
	router.GET("/api/mariadb/users/:id", authMiddleware, mariadbHandler.GetUserByID)
	router.PUT("/api/mariadb/users/:id", authMiddleware, mariadbHandler.UpdateUser)
	router.DELETE("/api/mariadb/users/:id", authMiddleware, mariadbHandler.DeleteUser)

	// CRUD routes for PostgreSQL (protected by auth)
	postgresHandler := handlers.NewUserHandler(conns.PostgreSQL)
	router.POST("/api/postgres/users", authMiddleware, postgresHandler.CreateUser)
	router.GET("/api/postgres/users", authMiddleware, postgresHandler.GetAllUsers)
	router.GET("/api/postgres/users/:id", authMiddleware, postgresHandler.GetUserByID)
	router.PUT("/api/postgres/users/:id", authMiddleware, postgresHandler.UpdateUser)
	router.DELETE("/api/postgres/users/:id", authMiddleware, postgresHandler.DeleteUser)

	// CRUD routes for Percona (protected by auth)
	perconaHandler := handlers.NewUserHandler(conns.Percona)
	router.POST("/api/percona/users", authMiddleware, perconaHandler.CreateUser)
	router.GET("/api/percona/users", authMiddleware, perconaHandler.GetAllUsers)
	router.GET("/api/percona/users/:id", authMiddleware, perconaHandler.GetUserByID)
	router.PUT("/api/percona/users/:id", authMiddleware, perconaHandler.UpdateUser)
	router.DELETE("/api/percona/users/:id", authMiddleware, perconaHandler.DeleteUser)

	// CRUD routes for MongoDB (protected by auth)
	mongoHandler := handlers.NewMongoDBHandler(conns.MongoDB)
	router.POST("/api/mongodb/users", authMiddleware, mongoHandler.CreateUser)
	router.GET("/api/mongodb/users", authMiddleware, mongoHandler.GetAllUsers)
	router.GET("/api/mongodb/users/:id", authMiddleware, mongoHandler.GetUserByID)
	router.PUT("/api/mongodb/users/:id", authMiddleware, mongoHandler.UpdateUser)
	router.DELETE("/api/mongodb/users/:id", authMiddleware, mongoHandler.DeleteUser)

	// CRUD routes for Firebase (protected by auth)
	firebaseHandler := handlers.NewFirebaseHandler("http://firebase:9000")
	router.POST("/api/firebase/users", authMiddleware, firebaseHandler.CreateUser)
	router.GET("/api/firebase/users", authMiddleware, firebaseHandler.GetAllUsers)
	router.GET("/api/firebase/users/:id", authMiddleware, firebaseHandler.GetUserByID)
	router.PUT("/api/firebase/users/:id", authMiddleware, firebaseHandler.UpdateUser)
	router.DELETE("/api/firebase/users/:id", authMiddleware, firebaseHandler.DeleteUser)

	// CRUD routes for Oracle (protected by auth)
	oracleHandler := handlers.NewOracleHandler(conns.Oracle)
	router.POST("/api/oracle/users", authMiddleware, oracleHandler.CreateUser)
	router.GET("/api/oracle/users", authMiddleware, oracleHandler.GetAllUsers)
	router.GET("/api/oracle/users/:id", authMiddleware, oracleHandler.GetUserByID)
	router.PUT("/api/oracle/users/:id", authMiddleware, oracleHandler.UpdateUser)
	router.DELETE("/api/oracle/users/:id", authMiddleware, oracleHandler.DeleteUser)

	apiPort := cfg.API.Port
	apiHost := cfg.API.Host

	fmt.Printf("📡 API Server running on http://%s:%s\n", apiHost, apiPort)
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  /health                  - Health check")
	fmt.Println("  GET  /status                  - Check all connections")
	fmt.Println("  POST /api/mysql/users         - Create user (MySQL)")
	fmt.Println("  GET  /api/mysql/users         - Get all users (MySQL)")
	fmt.Println("  GET  /api/mysql/users/:id     - Get user by ID (MySQL)")
	fmt.Println("  PUT  /api/mysql/users/:id     - Update user (MySQL)")
	fmt.Println("  DELETE /api/mysql/users/:id   - Delete user (MySQL)")
	fmt.Println("  POST /api/mariadb/users       - Create user (MariaDB)")
	fmt.Println("  GET  /api/mariadb/users       - Get all users (MariaDB)")
	fmt.Println("  GET  /api/mariadb/users/:id   - Get user by ID (MariaDB)")
	fmt.Println("  PUT  /api/mariadb/users/:id   - Update user (MariaDB)")
	fmt.Println("  DELETE /api/mariadb/users/:id - Delete user (MariaDB)")
	fmt.Println("  POST /api/postgres/users      - Create user (PostgreSQL)")
	fmt.Println("  GET  /api/postgres/users      - Get all users (PostgreSQL)")
	fmt.Println("  GET  /api/postgres/users/:id  - Get user by ID (PostgreSQL)")
	fmt.Println("  PUT  /api/postgres/users/:id  - Update user (PostgreSQL)")
	fmt.Println("  DELETE /api/postgres/users/:id - Delete user (PostgreSQL)")
	fmt.Println("  POST /api/percona/users       - Create user (Percona)")
	fmt.Println("  GET  /api/percona/users       - Get all users (Percona)")
	fmt.Println("  GET  /api/percona/users/:id   - Get user by ID (Percona)")
	fmt.Println("  PUT  /api/percona/users/:id   - Update user (Percona)")
	fmt.Println("  DELETE /api/percona/users/:id - Delete user (Percona)")
	fmt.Println("  POST /api/mongodb/users       - Create user (MongoDB)")
	fmt.Println("  GET  /api/mongodb/users       - Get all users (MongoDB)")
	fmt.Println("  GET  /api/mongodb/users/:id   - Get user by ID (MongoDB)")
	fmt.Println("  PUT  /api/mongodb/users/:id   - Update user (MongoDB)")
	fmt.Println("  DELETE /api/mongodb/users/:id - Delete user (MongoDB)")
	fmt.Println("  POST /api/firebase/users      - Create user (Firebase)")
	fmt.Println("  GET  /api/firebase/users      - Get all users (Firebase)")
	fmt.Println("  GET  /api/firebase/users/:id  - Get user by ID (Firebase)")
	fmt.Println("  PUT  /api/firebase/users/:id  - Update user (Firebase)")
	fmt.Println("  DELETE /api/firebase/users/:id - Delete user (Firebase)")
	fmt.Println("  POST /api/oracle/users        - Create user (Oracle)")
	fmt.Println("  GET  /api/oracle/users        - Get all users (Oracle)")
	fmt.Println("  GET  /api/oracle/users/:id    - Get user by ID (Oracle)")
	fmt.Println("  PUT  /api/oracle/users/:id    - Update user (Oracle)")
	fmt.Println("  DELETE /api/oracle/users/:id  - Delete user (Oracle)")
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
