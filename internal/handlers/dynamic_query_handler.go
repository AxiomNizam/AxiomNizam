package handlers

import (
	"net/http"
	"strings"

	"example.com/axiomnizam/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DynamicQueryHandler handles dynamic SQL queries
type DynamicQueryHandler struct {
	db *gorm.DB
}

// NewDynamicQueryHandler creates a new dynamic query handler
func NewDynamicQueryHandler(db *gorm.DB) *DynamicQueryHandler {
	return &DynamicQueryHandler{db: db}
}

// QueryRequest represents a dynamic query request
type QueryRequest struct {
	Query  string        `json:"query" binding:"required"`
	Params []interface{} `json:"params"`
}

// DynamicQuery handles GET requests with dynamic SQL queries
// Query should be passed as URL parameter: /api/{db}/query?q=SELECT * FROM users WHERE id = ?&params=1
// Or with body: POST /api/{db}/query with JSON body
func (h *DynamicQueryHandler) DynamicQuery(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  "Database not connected",
		})
		return
	}

	var query string
	var params []interface{}

	// Try to get query from URL parameters (for GET requests)
	query = c.Query("q")
	if query == "" {
		// Try to get from body (for POST requests)
		var req QueryRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				Status: "error",
				Error:  "Missing query parameter. Use ?q=YOUR_QUERY or send JSON body with 'query' field",
			})
			return
		}
		query = req.Query
		params = req.Params
	} else {
		// Parse params from URL if provided
		paramStr := c.Query("params")
		if paramStr != "" {
			// Split params by comma and convert to interface{}
			paramParts := strings.Split(paramStr, ",")
			for _, p := range paramParts {
				params = append(params, strings.TrimSpace(p))
			}
		}
	}

	// Validate query - only allow SELECT, WITH (CTE), SHOW, DESCRIBE
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	if !isSelectQuery(upperQuery) {
		c.JSON(http.StatusForbidden, models.Response{
			Status: "error",
			Error:  "Only SELECT queries are allowed for GET requests. Use POST for INSERT/UPDATE/DELETE/CREATE",
		})
		return
	}

	// Execute the query
	var results []map[string]interface{}
	rows, err := h.db.Raw(query, params...).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Query execution failed: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  "Failed to get columns: " + err.Error(),
		})
		return
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  "Failed to scan row: " + err.Error(),
			})
			return
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		results = append(results, entry)
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "Query executed successfully",
		Data:    results,
	})
}

// DynamicQueryWithBody handles POST requests with dynamic SQL queries (supports all query types)
func (h *DynamicQueryHandler) DynamicQueryWithBody(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  "Database not connected",
		})
		return
	}

	var req QueryRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid request body. Expected: {\"query\": \"SQL_QUERY\", \"params\": []}",
		})
		return
	}

	if req.Query == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Query cannot be empty",
		})
		return
	}

	// For POST, allow SELECT/INSERT/UPDATE/DELETE/CREATE
	upperQuery := strings.ToUpper(strings.TrimSpace(req.Query))
	if !isWriteOrSelectQuery(upperQuery) {
		c.JSON(http.StatusForbidden, models.Response{
			Status: "error",
			Error:  "Query type not allowed. Allowed: SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, ALTER",
		})
		return
	}

	// Check if it's a SELECT query
	if isSelectQuery(upperQuery) {
		// Execute SELECT query
		var results []map[string]interface{}
		rows, err := h.db.Raw(req.Query, req.Params...).Rows()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  "Query execution failed: " + err.Error(),
			})
			return
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  "Failed to get columns: " + err.Error(),
			})
			return
		}

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				c.JSON(http.StatusInternalServerError, models.Response{
					Status: "error",
					Error:  "Failed to scan row: " + err.Error(),
				})
				return
			}

			entry := make(map[string]interface{})
			for i, col := range columns {
				var v interface{}
				val := values[i]
				b, ok := val.([]byte)
				if ok {
					v = string(b)
				} else {
					v = val
				}
				entry[col] = v
			}
			results = append(results, entry)
		}

		c.JSON(http.StatusOK, models.Response{
			Status:  "ok",
			Message: "Query executed successfully",
			Data:    results,
		})
	} else {
		// Execute non-SELECT query (INSERT, UPDATE, DELETE, CREATE, DROP, ALTER)
		result := h.db.Exec(req.Query, req.Params...)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, models.Response{
				Status: "error",
				Error:  "Query execution failed: " + result.Error.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, models.Response{
			Status:  "ok",
			Message: "Query executed successfully",
			Data: map[string]interface{}{
				"rows_affected": result.RowsAffected,
			},
		})
	}
}

// BatchQueries handles batch queries via POST
func (h *DynamicQueryHandler) BatchQueries(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  "Database not connected",
		})
		return
	}

	var requests []QueryRequest
	if err := c.BindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid request body. Expected array of {\"query\": \"SQL_QUERY\", \"params\": []}",
		})
		return
	}

	if len(requests) == 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Requests array cannot be empty",
		})
		return
	}

	results := make([]map[string]interface{}, 0)
	for i, req := range requests {
		upperQuery := strings.ToUpper(strings.TrimSpace(req.Query))
		if !isWriteOrSelectQuery(upperQuery) {
			c.JSON(http.StatusForbidden, models.Response{
				Status: "error",
				Error:  "Query " + string(rune(i+1)) + " - Query type not allowed",
			})
			return
		}

		if isSelectQuery(upperQuery) {
			rows, err := h.db.Raw(req.Query, req.Params...).Rows()
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.Response{
					Status: "error",
					Error:  "Query " + string(rune(i+1)) + " failed: " + err.Error(),
				})
				return
			}
			defer rows.Close()

			columns, _ := rows.Columns()
			for rows.Next() {
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for j := range columns {
					valuePtrs[j] = &values[j]
				}
				rows.Scan(valuePtrs...)

				entry := make(map[string]interface{})
				for j, col := range columns {
					var v interface{}
					val := values[j]
					b, ok := val.([]byte)
					if ok {
						v = string(b)
					} else {
						v = val
					}
					entry[col] = v
				}
				results = append(results, entry)
			}
		} else {
			result := h.db.Exec(req.Query, req.Params...)
			if result.Error != nil {
				c.JSON(http.StatusInternalServerError, models.Response{
					Status: "error",
					Error:  "Query " + string(rune(i+1)) + " failed: " + result.Error.Error(),
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "All queries executed successfully",
		Data:    results,
	})
}

// TableSchema returns the schema of a table
func (h *DynamicQueryHandler) TableSchema(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Status: "error",
			Error:  "Database not connected",
		})
		return
	}

	tableName := c.Query("table")
	if tableName == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Table name required. Use ?table=table_name",
		})
		return
	}

	// Get table columns
	type Column struct {
		Field   string
		Type    string
		Null    string
		Key     string
		Default interface{}
		Extra   string
	}

	var columns []Column
	// This works for MySQL, MariaDB, Percona
	// For PostgreSQL, you would need different query
	// For Oracle, you would need different query

	dbType := h.db.Dialector.Name()
	var query string

	switch dbType {
	case "mysql":
		query = "SELECT COLUMN_NAME as Field, COLUMN_TYPE as Type, IS_NULLABLE as Null, COLUMN_KEY as Key, COLUMN_DEFAULT as Default, EXTRA as Extra FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = ? ORDER BY ORDINAL_POSITION"
		h.db.Raw(query, tableName).Scan(&columns)
	case "postgres":
		// PostgreSQL query
		query = "SELECT column_name as Field, data_type as Type, is_nullable as Null FROM information_schema.columns WHERE table_name = $1"
		h.db.Raw(query, tableName).Scan(&columns)
	default:
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Schema query not supported for this database type",
		})
		return
	}

	if len(columns) == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Status: "error",
			Error:  "Table not found or has no columns",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  "ok",
		Message: "Table schema retrieved successfully",
		Data:    columns,
	})
}

// Helper function to check if query is a SELECT query
func isSelectQuery(upperQuery string) bool {
	return strings.HasPrefix(upperQuery, "SELECT") ||
		strings.HasPrefix(upperQuery, "WITH") ||
		strings.HasPrefix(upperQuery, "SHOW") ||
		strings.HasPrefix(upperQuery, "DESCRIBE") ||
		strings.HasPrefix(upperQuery, "DESC") ||
		strings.HasPrefix(upperQuery, "EXPLAIN")
}

// Helper function to check if query is allowed (SELECT or write operations)
func isWriteOrSelectQuery(upperQuery string) bool {
	allowedPrefixes := []string{
		"SELECT", "WITH", "SHOW", "DESCRIBE", "DESC", "EXPLAIN",
		"INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER",
		"TRUNCATE", "REPLACE",
	}
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(upperQuery, prefix) {
			return true
		}
	}
	return false
}

// QueryValidator validates SQL query safety
func ValidateQuerySafety(query string) bool {
	// Basic validation - in production, use more sophisticated SQL parser
	dangerousKeywords := []string{
		"DROP DATABASE",
		"DROP SCHEMA",
	}
	upperQuery := strings.ToUpper(query)
	for _, keyword := range dangerousKeywords {
		if strings.Contains(upperQuery, keyword) {
			return false
		}
	}
	return true
}
