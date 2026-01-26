package bulk

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// BulkHandler handles bulk operation endpoints
type BulkHandler struct {
	manager BulkManager
}

// NewBulkHandler creates handler
func NewBulkHandler(manager BulkManager) *BulkHandler {
	return &BulkHandler{manager: manager}
}

// SubmitBulkOperation handles POST /api/v1/bulk/operations
func (h *BulkHandler) SubmitBulkOperation(c *gin.Context) {
	var req BulkOperationRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	op := &BulkOperation{
		TenantID:  req.TenantID,
		Type:      req.Type,
		Status:    "Pending",
		Items:     req.Items,
		Options:   req.Options,
		CreatedAt: time.Now(),
	}

	created, err := h.manager.SubmitOperation(op)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, created)
}

// GetOperation handles GET /api/v1/bulk/operations/:id
func (h *BulkHandler) GetOperation(c *gin.Context) {
	id := c.Param("id")
	op, err := h.manager.GetOperation(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "operation not found"})
		return
	}

	c.JSON(http.StatusOK, op)
}

// ListOperations handles GET /api/v1/bulk/operations
func (h *BulkHandler) ListOperations(c *gin.Context) {
	tenantID := c.Query("tenantId")
	status := c.Query("status")

	ops, err := h.manager.ListOperations(tenantID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"operations": ops, "count": len(ops)})
}

// GetProgress handles GET /api/v1/bulk/operations/:id/progress
func (h *BulkHandler) GetProgress(c *gin.Context) {
	id := c.Param("id")
	op, err := h.manager.GetOperation(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "operation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       op.ID,
		"status":   op.Status,
		"progress": float64(op.SuccessCount+op.FailureCount) / float64(op.TotalItems) * 100,
		"total":    op.TotalItems,
		"success":  op.SuccessCount,
		"failed":   op.FailureCount,
		"skipped":  op.SkippedCount,
	})
}

// CancelOperation handles DELETE /api/v1/bulk/operations/:id
func (h *BulkHandler) CancelOperation(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.CancelOperation(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "operation cancelled"})
}

// RetryFailed handles POST /api/v1/bulk/operations/:id/retry-failed
func (h *BulkHandler) RetryFailed(c *gin.Context) {
	id := c.Param("id")
	retried, err := h.manager.RetryFailed(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, retried)
}

// GetResults handles GET /api/v1/bulk/operations/:id/results
func (h *BulkHandler) GetResults(c *gin.Context) {
	id := c.Param("id")
	results, err := h.manager.GetResults(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "results not found"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// RegisterBulkRoutes registers all bulk operation routes
func RegisterBulkRoutes(router *gin.Engine, manager BulkManager) {
	handler := NewBulkHandler(manager)

	group := router.Group("/api/v1/bulk/operations")
	{
		group.POST("", handler.SubmitBulkOperation)
		group.GET("", handler.ListOperations)
		group.GET("/:id", handler.GetOperation)
		group.GET("/:id/progress", handler.GetProgress)
		group.DELETE("/:id", handler.CancelOperation)
		group.POST("/:id/retry-failed", handler.RetryFailed)
		group.GET("/:id/results", handler.GetResults)
	}
}

// BulkManager interface
type BulkManager interface {
	SubmitOperation(op *BulkOperation) (*BulkOperation, error)
	GetOperation(id string) (*BulkOperation, error)
	ListOperations(tenantID, status string) ([]*BulkOperation, error)
	CancelOperation(id string) error
	RetryFailed(id string) (*BulkOperation, error)
	GetResults(id string) (*BulkOperationResponse, error)
}
