package webhooks

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	manager        WebhookManager
	dualWriteStore WebhookDualWriteStore
}

// NewWebhookHandler creates handler
func NewWebhookHandler(manager WebhookManager) *WebhookHandler {
	return &WebhookHandler{manager: manager}
}

// CreateWebhook handles POST /api/v1/webhooks
func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	var req WebhookCreateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Phase 3: reconciler-authoritative path
	if h.isAuthoritative() {
		wh := &Webhook{Name: req.Name, Description: req.Description, URL: req.URL, Secret: req.Secret, Events: req.Events, Active: true}
		resource := h.buildWebhookResource(wh)
		if h.dualWriteStore != nil {
			if err := h.dualWriteStore.Create(c.Request.Context(), resource); err != nil {
				_ = h.dualWriteStore.Update(c.Request.Context(), resource)
			}
		}
		c.JSON(http.StatusAccepted, gin.H{"name": resource.Name, "status": "Pending", "message": "webhook resource created"})
		return
	}

	webhook := &Webhook{
		Name: req.Name, Description: req.Description, URL: req.URL, Secret: req.Secret,
		Events: req.Events, Filters: req.Filters, Active: true, CreatedAt: time.Now(),
	}
	created, err := h.manager.CreateWebhook(webhook)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.dualWriteWebhook(created)
	c.JSON(http.StatusCreated, created)
}

// GetWebhook handles GET /api/v1/webhooks/:id
func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	id := c.Param("id")
	webhook, err := h.manager.GetWebhook(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// ListWebhooks handles GET /api/v1/webhooks
func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	tenantID := c.Query("tenantId")
	eventType := c.Query("eventType")

	webhooks, err := h.manager.ListWebhooks(tenantID, eventType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"webhooks": webhooks, "count": len(webhooks)})
}

// UpdateWebhook handles PATCH /api/v1/webhooks/:id
func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook, err := h.manager.GetWebhook(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}

	webhook.UpdatedAt = time.Now()
	updated, err := h.manager.UpdateWebhook(webhook)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteWebhook handles DELETE /api/v1/webhooks/:id
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.DeleteWebhook(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook deleted"})
}

// TestWebhook handles POST /api/v1/webhooks/:id/test
func (h *WebhookHandler) TestWebhook(c *gin.Context) {
	id := c.Param("id")
	webhook, err := h.manager.GetWebhook(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}

	result, err := h.manager.TestWebhook(webhook)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetDeliveryLogs handles GET /api/v1/webhooks/:id/deliveries
func (h *WebhookHandler) GetDeliveryLogs(c *gin.Context) {
	id := c.Param("id")
	logs, err := h.manager.GetDeliveryLogs(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "logs not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deliveries": logs})
}

// RegisterWebhookRoutes registers all webhook routes
func RegisterWebhookRoutes(router *gin.Engine, manager WebhookManager) {
	handler := NewWebhookHandler(manager)

	group := router.Group("/api/v1/webhooks")
	{
		group.POST("", handler.CreateWebhook)
		group.GET("", handler.ListWebhooks)
		group.GET("/:id", handler.GetWebhook)
		group.PATCH("/:id", handler.UpdateWebhook)
		group.DELETE("/:id", handler.DeleteWebhook)
		group.POST("/:id/test", handler.TestWebhook)
		group.GET("/:id/deliveries", handler.GetDeliveryLogs)
	}
}

// WebhookManager interface
type WebhookManager interface {
	CreateWebhook(webhook *Webhook) (*Webhook, error)
	GetWebhook(id string) (*Webhook, error)
	ListWebhooks(tenantID, eventType string) ([]*Webhook, error)
	UpdateWebhook(webhook *Webhook) (*Webhook, error)
	DeleteWebhook(id string) error
	TestWebhook(webhook *Webhook) (interface{}, error)
	GetDeliveryLogs(webhookID string) ([]*WebhookDeliveryLog, error)
}
