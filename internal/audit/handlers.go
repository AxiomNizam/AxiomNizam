package audit

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditHandler handles audit endpoints
type AuditHandler struct {
	logger AuditLogger
}

// NewAuditHandler creates handler
func NewAuditHandler(logger AuditLogger) *AuditHandler {
	return &AuditHandler{logger: logger}
}

// LogAction handles POST /api/v1/audit/logs
func (h *AuditHandler) LogAction(c *gin.Context) {
	var req struct {
		TenantID   string      `json:"tenantId" binding:"required"`
		UserID     string      `json:"userId" binding:"required"`
		Username   string      `json:"username"`
		Action     AuditAction `json:"action" binding:"required"`
		ResourceID string      `json:"resourceId"`
		Result     AuditResult `json:"result" binding:"required"`
		SourceIP   string      `json:"sourceIp"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log := &AuditLog{
		TenantID:   req.TenantID,
		UserID:     req.UserID,
		Username:   req.Username,
		Action:     req.Action,
		ResourceID: req.ResourceID,
		Result:     req.Result,
		SourceIP:   req.SourceIP,
		Timestamp:  time.Now(),
	}

	if err := h.logger.LogAction(c.Request.Context(), log); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": log.ID, "timestamp": log.Timestamp})
}

// QueryLogs handles GET /api/v1/audit/logs
func (h *AuditHandler) QueryLogs(c *gin.Context) {
	filter := &AuditFilter{
		TenantID: c.Query("tenantId"),
		UserID:   c.Query("userId"),
		Username: c.Query("username"),
		Limit:    100,
	}

	logs, err := h.logger.QueryLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs, "count": len(logs)})
}

// GetReport handles GET /api/v1/audit/report
func (h *AuditHandler) GetReport(c *gin.Context) {
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenantId required"})
		return
	}

	filter := &AuditFilter{TenantID: tenantID}
	report, err := h.logger.GetReport(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// DeleteOldLogs handles DELETE /api/v1/audit/logs
func (h *AuditHandler) DeleteOldLogs(c *gin.Context) {
	var olderThanDays int = 90

	if err := h.logger.DeleteOldLogs(c.Request.Context(), olderThanDays); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "old logs deleted"})
}

// RegisterAuditRoutes registers all audit routes
func RegisterAuditRoutes(router *gin.Engine, logger AuditLogger) {
	handler := NewAuditHandler(logger)

	group := router.Group("/api/v1/audit")
	{
		group.POST("/logs", handler.LogAction)
		group.GET("/logs", handler.QueryLogs)
		group.GET("/report", handler.GetReport)
		group.DELETE("/logs", handler.DeleteOldLogs)
	}
}
