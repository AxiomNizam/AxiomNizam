package docs

import (
	"net/http"

	"example.com/axiomnizam/internal/models"
	"github.com/gin-gonic/gin"
)

// DocsHandler handles documentation endpoints
type DocsHandler struct {
	generator *OpenAPIGenerator
}

// NewDocsHandler creates a new docs handler
func NewDocsHandler(generator *OpenAPIGenerator) *DocsHandler {
	return &DocsHandler{
		generator: generator,
	}
}

// GetOpenAPISpec handles GET /api/docs/openapi.json
func (dh *DocsHandler) GetOpenAPISpec(c *gin.Context) {
	spec := dh.generator.BuildOpenAPI()
	c.JSON(http.StatusOK, spec)
}

// GetSwaggerUI handles GET /api/docs/swagger
func (dh *DocsHandler) GetSwaggerUI(c *gin.Context) {
	html := GetSwaggerUIHTML("/api/docs/openapi.json")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// GetReDocUI handles GET /api/docs/redoc
func (dh *DocsHandler) GetReDocUI(c *gin.Context) {
	html := GetSwaggerUIHTML("/api/docs/openapi.json")
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

	for _, endpoint := range dh.generator.endpoints {
		endpoints = append(endpoints, map[string]interface{}{
			"path":        endpoint.Path,
			"method":      endpoint.Method,
			"summary":     endpoint.Summary,
			"description": endpoint.Description,
			"tags":        endpoint.Tags,
		})
	}

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
		Data: map[string]interface{}{
			"endpoint": dh.generator.endpoints,
		},
	})
}
