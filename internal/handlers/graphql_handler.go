package handlers

import (
	"net/http"

	"example.com/axiomnizam/internal/graphql"
	"example.com/axiomnizam/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GraphQLHandler handles GraphQL requests
type GraphQLHandler struct {
	resolver *graphql.QueryResolver
}

// NewGraphQLHandler creates a new GraphQL handler
func NewGraphQLHandler(db *gorm.DB) *GraphQLHandler {
	return &GraphQLHandler{
		resolver: graphql.NewQueryResolver(db),
	}
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []interface{} `json:"errors,omitempty"`
}

// Query handles POST /api/graphql
func (h *GraphQLHandler) Query(c *gin.Context) {
	var req GraphQLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid request: " + err.Error(),
		})
		return
	}

	if req.Query == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Query is required",
		})
		return
	}

	data, err := h.resolver.ResolveQuery(c.Request.Context(), req.Query, req.Variables, req.OperationName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GraphQLResponse{
		Data: data,
	})
}

// GetSchema handles GET /api/graphql/schema
func (h *GraphQLHandler) GetSchema(c *gin.Context) {
	_, err := h.resolver.BuildSchema()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Status: "ok",
		Data: map[string]interface{}{
			"schema":         "GraphQL schema available",
			"query_endpoint": "/api/graphql",
			"playground":     "/api/graphql/playground",
		},
	})
}

// Playground handles GET /graphql (interactive playground)
func (h *GraphQLHandler) Playground(c *gin.Context) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>GraphQL Playground</title>
		<link rel="stylesheet" href="https://unpkg.com/graphql-playground-react/build/static/css/index.css"/>
	</head>
	<body>
		<div id="root"></div>
		<script src="https://unpkg.com/graphql-playground-react/build/umd/graphql-playground.js"></script>
		<script>
			window.addEventListener('load', function (event) {
				GraphQLPlayground.init(document.getElementById('root'), {
					endpoint: '/api/graphql',
				})
			})
		</script>
	</body>
	</html>
	`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
