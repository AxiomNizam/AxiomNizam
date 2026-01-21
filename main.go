package main

import (
	"fmt"
	"log"

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

	// Initialize all connections
	conns := database.InitConnections(cfg)

	// Create tables
	createTables(conns)

	// Create Gin router
	router := gin.Default()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(conns)
	userHandler := handlers.NewUserHandler(conns.MySQL)

	// Health check endpoints
	router.GET("/health", healthHandler.Health)
	router.GET("/status", healthHandler.Status)

	// CRUD routes for MySQL
	router.POST("/api/mysql/users", userHandler.CreateUser)
	router.GET("/api/mysql/users", userHandler.GetAllUsers)
	router.GET("/api/mysql/users/:id", userHandler.GetUserByID)
	router.PUT("/api/mysql/users/:id", userHandler.UpdateUser)
	router.DELETE("/api/mysql/users/:id", userHandler.DeleteUser)

	// CRUD routes for MariaDB
	mariadbHandler := handlers.NewUserHandler(conns.MariaDB)
	router.POST("/api/mariadb/users", mariadbHandler.CreateUser)
	router.GET("/api/mariadb/users", mariadbHandler.GetAllUsers)
	router.GET("/api/mariadb/users/:id", mariadbHandler.GetUserByID)
	router.PUT("/api/mariadb/users/:id", mariadbHandler.UpdateUser)
	router.DELETE("/api/mariadb/users/:id", mariadbHandler.DeleteUser)

	// CRUD routes for PostgreSQL
	postgresHandler := handlers.NewUserHandler(conns.PostgreSQL)
	router.POST("/api/postgres/users", postgresHandler.CreateUser)
	router.GET("/api/postgres/users", postgresHandler.GetAllUsers)
	router.GET("/api/postgres/users/:id", postgresHandler.GetUserByID)
	router.PUT("/api/postgres/users/:id", postgresHandler.UpdateUser)
	router.DELETE("/api/postgres/users/:id", postgresHandler.DeleteUser)

	// CRUD routes for Firebase
	firebaseHandler := handlers.NewFirebaseHandler("http://firebase:9000")
	router.POST("/api/firebase/users", firebaseHandler.CreateUser)
	router.GET("/api/firebase/users", firebaseHandler.GetAllUsers)
	router.GET("/api/firebase/users/:id", firebaseHandler.GetUserByID)
	router.PUT("/api/firebase/users/:id", firebaseHandler.UpdateUser)
	router.DELETE("/api/firebase/users/:id", firebaseHandler.DeleteUser)

	apiPort := cfg.API.Port
	apiHost := cfg.API.Host

	fmt.Printf("📡 API Server running on http://%s:%s\n", apiHost, apiPort)
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  /health              - Health check")
	fmt.Println("  GET  /status              - Check all connections")
	fmt.Println("  POST /api/mysql/users     - Create user (MySQL)")
	fmt.Println("  GET  /api/mysql/users     - Get all users (MySQL)")
	fmt.Println("  GET  /api/mysql/users/:id - Get user by ID (MySQL)")
	fmt.Println("  PUT  /api/mysql/users/:id - Update user (MySQL)")
	fmt.Println("  DELETE /api/mysql/users/:id - Delete user (MySQL)")
	fmt.Println("  POST /api/mariadb/users   - Create user (MariaDB)")
	fmt.Println("  GET  /api/mariadb/users   - Get all users (MariaDB)")
	fmt.Println("  GET  /api/mariadb/users/:id - Get user by ID (MariaDB)")
	fmt.Println("  PUT  /api/mariadb/users/:id - Update user (MariaDB)")
	fmt.Println("  DELETE /api/mariadb/users/:id - Delete user (MariaDB)")
	fmt.Println("  POST /api/postgres/users  - Create user (PostgreSQL)")
	fmt.Println("  GET  /api/postgres/users  - Get all users (PostgreSQL)")
	fmt.Println("  GET  /api/postgres/users/:id - Get user by ID (PostgreSQL)")
	fmt.Println("  PUT  /api/postgres/users/:id - Update user (PostgreSQL)")
	fmt.Println("  DELETE /api/postgres/users/:id - Delete user (PostgreSQL)")
	fmt.Println("  POST /api/firebase/users  - Create user (Firebase)")
	fmt.Println("  GET  /api/firebase/users  - Get all users (Firebase)")
	fmt.Println("  GET  /api/firebase/users/:id - Get user by ID (Firebase)")
	fmt.Println("  PUT  /api/firebase/users/:id - Update user (Firebase)")
	fmt.Println("  DELETE /api/firebase/users/:id - Delete user (Firebase)")
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
}
