package handlers

import (
	"net/http"

	"example.com/axiomnizam/internal/docs"
	"example.com/axiomnizam/internal/models"
	"github.com/gin-gonic/gin"
)

// DocsHandler handles documentation endpoints
type DocsHandler struct {
	generator *docs.OpenAPIGenerator
}

// NewDocsHandler creates a new docs handler
func NewDocsHandler() *DocsHandler {
	generator := docs.NewOpenAPIGenerator(docs.OpenAPIInfo{
		Title:       "AxiomNizam API",
		Version:     "1.0.0",
		Description: "AxiomNizam Platform API",
	})
	return &DocsHandler{
		generator: generator,
	}
}

// GetOpenAPISpec handles GET /api/docs/openapi.json
func (dh *DocsHandler) GetOpenAPISpec(c *gin.Context) {
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]string{
			"title":   "AxiomNizam API",
			"version": "1.0.0",
		},
	}
	c.JSON(http.StatusOK, spec)
}

// GetSwaggerUI handles GET /api/docs/swagger
func (dh *DocsHandler) GetSwaggerUI(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>API Docs</title>
</head>
<body>
    <h1>API Documentation</h1>
    <p>Visit /api/docs/openapi.json for the OpenAPI spec.</p>
</body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// GetReDocUI handles GET /api/docs/redoc
func (dh *DocsHandler) GetReDocUI(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>ReDoc</title>
</head>
<body>
    <h1>API Documentation</h1>
    <p>Visit /api/docs/openapi.json for the OpenAPI spec.</p>
</body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// GetMarkdownDocs handles GET /api/docs/markdown
func (dh *DocsHandler) GetMarkdownDocs(c *gin.Context) {
	markdown := dh.generator.GetEndpointMarkdown()
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, markdown)
}

// ListEndpoints handles GET /api/docs/endpoints
func (dh *DocsHandler) ListEndpoints(c *gin.Context) {
	endpoints := make([]map[string]interface{}, 0)

	c.JSON(http.StatusOK, models.Response{
		Status: "ok",
		Data: map[string]interface{}{
			"endpoints": endpoints,
			"total":     len(endpoints),
		},
	})
}

// GetEndpointDetails handles GET /api/docs/endpoints/:id
func (dh *DocsHandler) GetEndpointDetails(c *gin.Context) {
	id := c.Param("id")

	if id == "" || id == "0" {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "endpoint id is required",
		})
		return
	}

	// Return detailed spec for endpoint
	c.JSON(http.StatusOK, models.Response{
		Status: "ok",
		Data:   dh.generator.BuildOpenAPI(),
	})
}
