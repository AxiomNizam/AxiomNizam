package eventbus

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// EventBusHandler handles event bus endpoints
type EventBusHandler struct {
	manager EventBusManager
}

// NewEventBusHandler creates handler
func NewEventBusHandler(manager EventBusManager) *EventBusHandler {
	return &EventBusHandler{manager: manager}
}

// PublishEvent handles POST /api/v1/events/publish
func (h *EventBusHandler) PublishEvent(c *gin.Context) {
	var req EventPublishRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := &EventBusEvent{
		Type:          req.Type,
		Subject:       req.Subject,
		Data:          req.Data,
		Source:        req.Source,
		Timestamp:     time.Now(),
		Metadata:      req.Metadata,
		CorrelationID: req.CorrelationID,
		CausationID:   req.CausationID,
		Priority:      req.Priority,
	}

	resp, err := h.manager.PublishEvent(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// ListEvents handles GET /api/v1/events
func (h *EventBusHandler) ListEvents(c *gin.Context) {
	tenantID := c.Query("tenantId")
	eventType := c.Query("type")
	processed := c.Query("processed")

	events, err := h.manager.ListEvents(tenantID, eventType, processed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events, "count": len(events)})
}

// CreateTopic handles POST /api/v1/topics
func (h *EventBusHandler) CreateTopic(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		Partitions int    `json:"partitions"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic := &EventTopic{
		Name:       req.Name,
		Partitions: req.Partitions,
		CreatedAt:  time.Now(),
		IsActive:   true,
	}

	created, err := h.manager.CreateTopic(topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// ListTopics handles GET /api/v1/topics
func (h *EventBusHandler) ListTopics(c *gin.Context) {
	topics, err := h.manager.ListTopics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"topics": topics, "count": len(topics)})
}

// CreateSubscription handles POST /api/v1/subscriptions
func (h *EventBusHandler) CreateSubscription(c *gin.Context) {
	var req struct {
		TenantID string      `json:"tenantId" binding:"required"`
		Name     string      `json:"name" binding:"required"`
		Topics   []string    `json:"topics" binding:"required"`
		Handler  string      `json:"handler" binding:"required"`
		Filter   EventFilter `json:"filter"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub := &EventSubscription{
		TenantID:  req.TenantID,
		Name:      req.Name,
		Topics:    req.Topics,
		Handler:   req.Handler,
		Filter:    req.Filter,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	created, err := h.manager.CreateSubscription(sub)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GetSubscription handles GET /api/v1/subscriptions/:id
func (h *EventBusHandler) GetSubscription(c *gin.Context) {
	id := c.Param("id")
	sub, err := h.manager.GetSubscription(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// ListSubscriptions handles GET /api/v1/subscriptions
func (h *EventBusHandler) ListSubscriptions(c *gin.Context) {
	tenantID := c.Query("tenantId")
	subs, err := h.manager.ListSubscriptions(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscriptions": subs, "count": len(subs)})
}

// ListDLQ handles GET /api/v1/dlq
func (h *EventBusHandler) ListDLQ(c *gin.Context) {
	tenantID := c.Query("tenantId")
	events, err := h.manager.ListDLQEvents(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events, "count": len(events)})
}

// RegisterEventBusRoutes registers all event bus routes
func RegisterEventBusRoutes(router *gin.Engine, manager EventBusManager) {
	handler := NewEventBusHandler(manager)

	group := router.Group("/api/v1")
	{
		group.POST("/events/publish", handler.PublishEvent)
		group.GET("/events", handler.ListEvents)
		group.POST("/topics", handler.CreateTopic)
		group.GET("/topics", handler.ListTopics)
		group.POST("/subscriptions", handler.CreateSubscription)
		group.GET("/subscriptions/:id", handler.GetSubscription)
		group.GET("/subscriptions", handler.ListSubscriptions)
		group.GET("/dlq", handler.ListDLQ)
	}
}

// EventBusManager interface
type EventBusManager interface {
	PublishEvent(event *EventBusEvent) (*EventPublishResponse, error)
	ListEvents(tenantID, eventType, processed string) ([]*EventBusEvent, error)
	CreateTopic(topic *EventTopic) (*EventTopic, error)
	ListTopics() ([]*EventTopic, error)
	CreateSubscription(sub *EventSubscription) (*EventSubscription, error)
	GetSubscription(id string) (*EventSubscription, error)
	ListSubscriptions(tenantID string) ([]*EventSubscription, error)
	ListDLQEvents(tenantID string) ([]*DLQEvent, error)
}
