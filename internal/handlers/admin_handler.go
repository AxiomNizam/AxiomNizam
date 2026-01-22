package handlers

import (
	"fmt"
	"log"
	"net/http"

	"example.com/axiomnizam/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminHandler handles admin operations like database and table creation
type AdminHandler struct {
	connections map[string]*gorm.DB
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(connections map[string]*gorm.DB) *AdminHandler {
	return &AdminHandler{
		connections: connections,
	}
}

// CreateDatabaseRequest represents a request to create a database
type CreateDatabaseRequest struct {
	DatabaseName string `json:"database_name" binding:"required"`
	DBType       string `json:"db_type" binding:"required"` // mysql, postgres, mongodb, etc.
}

// CreateTableRequest represents a request to create a table
type CreateTableRequest struct {
	TableName string `json:"table_name" binding:"required"`
	DBType    string `json:"db_type" binding:"required"`
	Columns   []struct {
		Name     string `json:"name" binding:"required"`
		Type     string `json:"type" binding:"required"` // varchar, int, text, etc.
		Size     int    `json:"size"`
		Nullable bool   `json:"nullable"`
		Primary  bool   `json:"primary"`
	} `json:"columns" binding:"required,min=1"`
}

// CreateDatabase creates a new database on the specified DB type
func (h *AdminHandler) CreateDatabase(c *gin.Context) {
	var req CreateDatabaseRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Validate database type
	db, exists := h.connections[req.DBType]
	if !exists {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Database type '%s' not supported", req.DBType),
		})
		return
	}

	if db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Database '%s' is not connected", req.DBType),
		})
		return
	}

	// Create database based on type
	var createSQL string
	switch req.DBType {
	case "mysql", "mariadb", "percona":
		createSQL = fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", req.DatabaseName)
	case "postgres":
		createSQL = fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", req.DatabaseName)
	default:
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Database creation not supported for '%s'", req.DBType),
		})
		return
	}

	// Execute create database query
	if result := db.Exec(createSQL); result.Error != nil {
		log.Printf("❌ Failed to create database '%s' on %s: %v", req.DatabaseName, req.DBType, result.Error)
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Failed to create database: %v", result.Error),
		})
		return
	}

	log.Printf("✅ Database '%s' created successfully on %s", req.DatabaseName, req.DBType)
	c.JSON(http.StatusCreated, gin.H{
		"status":   "success",
		"message":  fmt.Sprintf("Database '%s' created successfully", req.DatabaseName),
		"database": req.DatabaseName,
		"db_type":  req.DBType,
	})
}

// CreateTable creates a new table on the specified database
func (h *AdminHandler) CreateTable(c *gin.Context) {
	var req CreateTableRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Validate database type
	db, exists := h.connections[req.DBType]
	if !exists {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Database type '%s' not supported", req.DBType),
		})
		return
	}

	if db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Database '%s' is not connected", req.DBType),
		})
		return
	}

	// Build CREATE TABLE SQL based on database type
	var createSQL string
	switch req.DBType {
	case "mysql", "mariadb", "percona":
		createSQL = h.buildMySQLCreateTable(req)
	case "postgres":
		createSQL = h.buildPostgresCreateTable(req)
	default:
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Table creation not supported for '%s'", req.DBType),
		})
		return
	}

	// Execute create table query
	if result := db.Exec(createSQL); result.Error != nil {
		log.Printf("❌ Failed to create table '%s' on %s: %v", req.TableName, req.DBType, result.Error)
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Failed to create table: %v", result.Error),
		})
		return
	}

	log.Printf("✅ Table '%s' created successfully on %s", req.TableName, req.DBType)
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Table '%s' created successfully", req.TableName),
		"table":   req.TableName,
		"db_type": req.DBType,
		"columns": len(req.Columns),
	})
}

// buildMySQLCreateTable builds MySQL CREATE TABLE statement
func (h *AdminHandler) buildMySQLCreateTable(req CreateTableRequest) string {
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n", req.TableName)

	for i, col := range req.Columns {
		sql += fmt.Sprintf("  `%s` %s", col.Name, col.Type)

		if col.Size > 0 {
			sql += fmt.Sprintf("(%d)", col.Size)
		}

		if col.Primary {
			sql += " PRIMARY KEY"
		}

		if !col.Nullable {
			sql += " NOT NULL"
		}

		if i < len(req.Columns)-1 {
			sql += ",\n"
		} else {
			sql += "\n"
		}
	}

	sql += ")"
	return sql
}

// buildPostgresCreateTable builds PostgreSQL CREATE TABLE statement
func (h *AdminHandler) buildPostgresCreateTable(req CreateTableRequest) string {
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", req.TableName)

	for i, col := range req.Columns {
		sql += fmt.Sprintf("  %s %s", col.Name, col.Type)

		if col.Size > 0 {
			sql += fmt.Sprintf("(%d)", col.Size)
		}

		if col.Primary {
			sql += " PRIMARY KEY"
		}

		if !col.Nullable {
			sql += " NOT NULL"
		}

		if i < len(req.Columns)-1 {
			sql += ",\n"
		} else {
			sql += "\n"
		}
	}

	sql += ")"
	return sql
}

// ListDatabases lists all databases on the specified DB type
func (h *AdminHandler) ListDatabases(c *gin.Context) {
	dbType := c.Query("db_type")
	if dbType == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "db_type query parameter is required",
		})
		return
	}

	db, exists := h.connections[dbType]
	if !exists {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Database type '%s' not supported", dbType),
		})
		return
	}

	if db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Database '%s' is not connected", dbType),
		})
		return
	}

	var databases []string

	switch dbType {
	case "mysql", "mariadb", "percona":
		// Query: SHOW DATABASES
		rows, err := db.Raw("SHOW DATABASES").Rows()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  fmt.Sprintf("Failed to list databases: %v", err),
			})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var dbName string
			if err := rows.Scan(&dbName); err != nil {
				continue
			}
			databases = append(databases, dbName)
		}

	case "postgres":
		// Query: SELECT datname FROM pg_database
		rows, err := db.Raw("SELECT datname FROM pg_database WHERE datistemplate = false").Rows()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  fmt.Sprintf("Failed to list databases: %v", err),
			})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var dbName string
			if err := rows.Scan(&dbName); err != nil {
				continue
			}
			databases = append(databases, dbName)
		}

	default:
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Listing databases not supported for '%s'", dbType),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"db_type":   dbType,
		"databases": databases,
		"count":     len(databases),
	})
}

// ListTables lists all tables in the specified database
func (h *AdminHandler) ListTables(c *gin.Context) {
	dbType := c.Query("db_type")
	if dbType == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "db_type query parameter is required",
		})
		return
	}

	db, exists := h.connections[dbType]
	if !exists {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Database type '%s' not supported", dbType),
		})
		return
	}

	if db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Database '%s' is not connected", dbType),
		})
		return
	}

	var tables []string

	switch dbType {
	case "mysql", "mariadb", "percona":
		// Query: SHOW TABLES
		rows, err := db.Raw("SHOW TABLES").Rows()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  fmt.Sprintf("Failed to list tables: %v", err),
			})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				continue
			}
			tables = append(tables, tableName)
		}

	case "postgres":
		// Query: SELECT table_name FROM information_schema
		rows, err := db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='public'").Rows()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  fmt.Sprintf("Failed to list tables: %v", err),
			})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				continue
			}
			tables = append(tables, tableName)
		}

	default:
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  fmt.Sprintf("Listing tables not supported for '%s'", dbType),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"db_type": dbType,
		"tables":  tables,
		"count":   len(tables),
	})
}
