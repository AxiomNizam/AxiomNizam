package eventbus

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type EventBusHandler struct {
	manager            EventBusManager
	topicDualWriteStore TopicDualWriteStore
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

	// Phase 3: reconciler-authoritative path
	if h.isAuthoritative() {
		topic := &EventTopic{Name: req.Name, Partitions: req.Partitions, IsActive: true}
		resource := h.buildTopicResource(topic)
		if h.topicDualWriteStore != nil {
			if err := h.topicDualWriteStore.Create(c.Request.Context(), resource); err != nil {
				_ = h.topicDualWriteStore.Update(c.Request.Context(), resource)
			}
		}
		c.JSON(http.StatusAccepted, gin.H{"name": resource.Name, "status": "Pending", "message": "topic resource created"})
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

	h.dualWriteTopic(created)
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

// AckEvent handles POST /api/v1/events/:id/ack
func (h *EventBusHandler) AckEvent(c *gin.Context) {
	eventID := strings.TrimSpace(c.Param("id"))
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event id is required"})
		return
	}

	var req struct {
		SubscriptionID string `json:"subscriptionId"`
		AcknowledgedBy string `json:"acknowledgedBy"`
		Message        string `json:"message"`
	}

	if c.Request != nil && c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	event, err := h.manager.AckEvent(eventID, strings.TrimSpace(req.SubscriptionID), strings.TrimSpace(req.AcknowledgedBy), strings.TrimSpace(req.Message))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event acknowledged", "event": event})
}

// ReplayDLQEvent handles POST /api/v1/dlq/:id/replay
func (h *EventBusHandler) ReplayDLQEvent(c *gin.Context) {
	dlqID := strings.TrimSpace(c.Param("id"))
	if dlqID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dlq id is required"})
		return
	}

	var req struct {
		ReplayToTopic string `json:"replayToTopic"`
		ReplayedBy    string `json:"replayedBy"`
	}

	if c.Request != nil && c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	resp, err := h.manager.ReplayDLQEvent(dlqID, strings.TrimSpace(req.ReplayToTopic), strings.TrimSpace(req.ReplayedBy))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "dlq event replayed", "replay": resp})
}

// RegisterEventBusRoutes registers all event bus routes
func RegisterEventBusRoutes(router *gin.Engine, manager EventBusManager) {
	handler := NewEventBusHandler(manager)

	group := router.Group("/api/v1")
	{
		group.POST("/events/publish", handler.PublishEvent)
		group.GET("/events", handler.ListEvents)
		group.POST("/events/:id/ack", handler.AckEvent)
		group.POST("/topics", handler.CreateTopic)
		group.GET("/topics", handler.ListTopics)
		group.POST("/subscriptions", handler.CreateSubscription)
		group.GET("/subscriptions/:id", handler.GetSubscription)
		group.GET("/subscriptions", handler.ListSubscriptions)
		group.GET("/dlq", handler.ListDLQ)
		group.POST("/dlq/:id/replay", handler.ReplayDLQEvent)
	}
}

// EventBusManager interface
type EventBusManager interface {
	PublishEvent(event *EventBusEvent) (*EventPublishResponse, error)
	ListEvents(tenantID, eventType, processed string) ([]*EventBusEvent, error)
	AckEvent(eventID, subscriptionID, acknowledgedBy, message string) (*EventBusEvent, error)
	CreateTopic(topic *EventTopic) (*EventTopic, error)
	ListTopics() ([]*EventTopic, error)
	CreateSubscription(sub *EventSubscription) (*EventSubscription, error)
	GetSubscription(id string) (*EventSubscription, error)
	ListSubscriptions(tenantID string) ([]*EventSubscription, error)
	ListDLQEvents(tenantID string) ([]*DLQEvent, error)
	ReplayDLQEvent(dlqID, replayToTopic, replayedBy string) (*EventPublishResponse, error)
}
