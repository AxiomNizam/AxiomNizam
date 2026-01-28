package export

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ExportHandler handles export endpoints
type ExportHandler struct {
	manager ExportManager
}

// NewExportHandler creates handler
func NewExportHandler(manager ExportManager) *ExportHandler {
	return &ExportHandler{manager: manager}
}

// SubmitExport handles POST /api/v1/exports
func (h *ExportHandler) SubmitExport(c *gin.Context) {
	var req ExportCreateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := &ExportJob{
		Name:        req.Name,
		Description: req.Description,
		Format:      req.Format,
		Source:      req.Source,
		Query:       req.Query,
		Filters:     req.Filters,
		Columns:     req.Columns,
		Compression: req.Compression,
		Encryption:  req.Encryption,
		Destination: req.Destination,
		Status:      "Pending",
		CreatedAt:   time.Now(),
	}

	created, err := h.manager.SubmitExport(job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, created)
}

// GetExport handles GET /api/v1/exports/:id
func (h *ExportHandler) GetExport(c *gin.Context) {
	id := c.Param("id")
	job, err := h.manager.GetExport(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListExports handles GET /api/v1/exports
func (h *ExportHandler) ListExports(c *gin.Context) {
	tenantID := c.Query("tenantId")
	status := c.Query("status")
	format := c.Query("format")

	exports, err := h.manager.ListExports(tenantID, status, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exports": exports, "count": len(exports)})
}

// CancelExport handles DELETE /api/v1/exports/:id
func (h *ExportHandler) CancelExport(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.CancelExport(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "export cancelled"})
}

// GetExportProgress handles GET /api/v1/exports/:id/progress
func (h *ExportHandler) GetExportProgress(c *gin.Context) {
	id := c.Param("id")
	job, err := h.manager.GetExport(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        job.ID,
		"status":    job.Status,
		"progress":  job.Progress,
		"processed": job.ProcessedRows,
		"total":     job.RecordCount,
		"skipped":   job.SkippedRows,
		"errors":    job.ErrorRows,
	})
}

// DownloadExport handles GET /api/v1/exports/:id/download
func (h *ExportHandler) DownloadExport(c *gin.Context) {
	id := c.Param("id")
	job, err := h.manager.GetExport(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export not found"})
		return
	}

	if job.Status != "Completed" {
		c.JSON(http.StatusForbidden, gin.H{"error": "export not ready"})
		return
	}

	// Return download URL or file
	c.JSON(http.StatusOK, gin.H{
		"downloadUrl": "/files/" + job.ID,
		"fileSize":    job.FileSize,
		"contentType": "application/" + string(job.Format),
	})
}

// CreateTemplate handles POST /api/v1/export-templates
func (h *ExportHandler) CreateTemplate(c *gin.Context) {
	var req struct {
		Name        string            `json:"name" binding:"required"`
		Format      ExportFormat      `json:"format" binding:"required"`
		Source      ExportSource      `json:"source" binding:"required"`
		Destination ExportDestination `json:"destination" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := &ExportTemplate{
		Name:        req.Name,
		Format:      req.Format,
		Source:      req.Source,
		Destination: req.Destination,
		CreatedAt:   time.Now(),
	}

	created, err := h.manager.CreateTemplate(template)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// ListTemplates handles GET /api/v1/export-templates
func (h *ExportHandler) ListTemplates(c *gin.Context) {
	tenantID := c.Query("tenantId")
	templates, err := h.manager.ListTemplates(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates, "count": len(templates)})
}

// RegisterExportRoutes registers all export routes
func RegisterExportRoutes(router *gin.Engine, manager ExportManager) {
	handler := NewExportHandler(manager)

	group := router.Group("/api/v1")
	{
		group.POST("/exports", handler.SubmitExport)
		group.GET("/exports", handler.ListExports)
		group.GET("/exports/:id", handler.GetExport)
		group.GET("/exports/:id/progress", handler.GetExportProgress)
		group.GET("/exports/:id/download", handler.DownloadExport)
		group.DELETE("/exports/:id", handler.CancelExport)
		group.POST("/export-templates", handler.CreateTemplate)
		group.GET("/export-templates", handler.ListTemplates)
	}
}

// ExportManager interface
type ExportManager interface {
	SubmitExport(job *ExportJob) (*ExportJob, error)
	GetExport(id string) (*ExportJob, error)
	ListExports(tenantID, status, format string) ([]*ExportJob, error)
	CancelExport(id string) error
	CreateTemplate(template *ExportTemplate) (*ExportTemplate, error)
	ListTemplates(tenantID string) ([]*ExportTemplate, error)
}
