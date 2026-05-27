package bulk

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type BulkHandler struct {
	manager        BulkManager
	dualWriteStore BulkDualWriteStore
}

// NewBulkHandler creates handler
func NewBulkHandler(manager BulkManager) *BulkHandler {
	return &BulkHandler{manager: manager}
}

// SubmitBulkOperation handles POST /api/v1/bulk/operations
func (h *BulkHandler) SubmitBulkOperation(c *gin.Context) {
	var req BulkOperationRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: err.Error()})
		return
	}

	// Phase 3: reconciler-authoritative path
	if h.isAuthoritative() {
		op := &BulkOperation{TenantID: req.TenantID, Type: req.Type, Items: req.Items, Options: req.Options}
		resource := h.buildOperationResource(op)
		if h.dualWriteStore != nil {
			if err := h.dualWriteStore.Create(c.Request.Context(), resource); err != nil {
				_ = h.dualWriteStore.Update(c.Request.Context(), resource)
			}
		}
		c.JSON(http.StatusAccepted, ResourceCreatedResponse{Name: resource.Name, Status: "Pending", Message: "bulk operation resource created"})
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
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	h.dualWriteOperation(created)
	c.JSON(http.StatusAccepted, BulkOpToResponse(created))
}

// GetOperation handles GET /api/v1/bulk/operations/:id
func (h *BulkHandler) GetOperation(c *gin.Context) {
	id := c.Param("id")
	op, err := h.manager.GetOperation(id)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Message: "operation not found"})
		return
	}

	c.JSON(http.StatusOK, BulkOpToResponse(op))
}

// ListOperations handles GET /api/v1/bulk/operations
func (h *BulkHandler) ListOperations(c *gin.Context) {
	tenantID := c.Query("tenantId")
	status := c.Query("status")

	ops, err := h.manager.ListOperations(tenantID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	list := make([]BulkOpResponse, len(ops))
	for i, op := range ops {
		list[i] = BulkOpToResponse(op)
	}
	c.JSON(http.StatusOK, BulkOpListResponse{Operations: list, Count: len(list)})
}

// GetProgress handles GET /api/v1/bulk/operations/:id/progress
func (h *BulkHandler) GetProgress(c *gin.Context) {
	id := c.Param("id")
	op, err := h.manager.GetOperation(id)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Message: "operation not found"})
		return
	}

	c.JSON(http.StatusOK, BulkOpToProgressResponse(op))
}

// CancelOperation handles DELETE /api/v1/bulk/operations/:id
func (h *BulkHandler) CancelOperation(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.CancelOperation(id); err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "operation cancelled"})
}

// RetryFailed handles POST /api/v1/bulk/operations/:id/retry-failed
func (h *BulkHandler) RetryFailed(c *gin.Context) {
	id := c.Param("id")
	retried, err := h.manager.RetryFailed(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, BulkOpToResponse(retried))
}

// GetResults handles GET /api/v1/bulk/operations/:id/results
func (h *BulkHandler) GetResults(c *gin.Context) {
	id := c.Param("id")
	results, err := h.manager.GetResults(id)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Message: "results not found"})
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
