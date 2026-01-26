package streaming

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// StreamHandler handles streaming endpoints
type StreamHandler struct {
	manager StreamManager
	upgrader websocket.Upgrader
}

// NewStreamHandler creates handler
func NewStreamHandler(manager StreamManager) *StreamHandler {
	return &StreamHandler{
		manager: manager,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Configure properly in production
			},
		},
	}
}

// HandleStream handles WebSocket /ws/stream
func (h *StreamHandler) HandleStream(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	session := &StreamSession{
		ID:        fmt.Sprintf("session-%d", time.Now().UnixNano()),
		StartTime: time.Now(),
		IsActive:  true,
	}

	for {
		var req StreamRequest
		if err := conn.ReadJSON(&req); err != nil {
			break
		}

		stream, err := h.manager.CreateStream(&req)
		if err != nil {
			conn.WriteJSON(StreamMessage{
				Type: "error",
				Error: &StreamError{
					Code:    "STREAM_CREATE_FAILED",
					Message: err.Error(),
				},
			})
			continue
		}

		// Send stream messages
		for msg := range stream.Messages {
			if err := conn.WriteJSON(msg); err != nil {
				break
			}
		}
	}

	session.IsActive = false
	session.EndTime = time.Now()
}

// CreateStreamRequest handles POST /api/v1/streams
func (h *StreamHandler) CreateStreamRequest(c *gin.Context) {
	var req StreamRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stream, err := h.manager.CreateStream(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"streamId": stream.ID})
}

// GetStreamStatus handles GET /api/v1/streams/:id
func (h *StreamHandler) GetStreamStatus(c *gin.Context) {
	id := c.Param("id")
	stream, err := h.manager.GetStream(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
		return
	}

	c.JSON(http.StatusOK, stream)
}

// ListStreams handles GET /api/v1/streams
func (h *StreamHandler) ListStreams(c *gin.Context) {
	tenantID := c.Query("tenantId")
	status := c.Query("status")

	streams, err := h.manager.ListStreams(tenantID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"streams": streams, "count": len(streams)})
}

// CancelStream handles DELETE /api/v1/streams/:id
func (h *StreamHandler) CancelStream(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.CancelStream(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stream cancelled"})
}

// Subscribe handles POST /api/v1/subscriptions
func (h *StreamHandler) Subscribe(c *gin.Context) {
	var req struct {
		TenantID  string   `json:"tenantId" binding:"required"`
		Topic     string   `json:"topic" binding:"required"`
		EventTypes []string `json:"eventTypes"`
		Filters   map[string]interface{} `json:"filters"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub := &StreamSubscription{
		TenantID:   req.TenantID,
		Topic:      req.Topic,
		EventTypes: req.EventTypes,
		Filters:    req.Filters,
		CreatedAt:  time.Now(),
	}

	created, err := h.manager.Subscribe(sub)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// Unsubscribe handles DELETE /api/v1/subscriptions/:id
func (h *StreamHandler) Unsubscribe(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.Unsubscribe(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unsubscribed"})
}

// RegisterStreamRoutes registers all streaming routes
func RegisterStreamRoutes(router *gin.Engine, manager StreamManager) {
	handler := NewStreamHandler(manager)
	
	// WebSocket endpoint
	router.GET("/ws/stream", handler.HandleStream)
	
	// REST endpoints
	group := router.Group("/api/v1")
	{
		group.POST("/streams", handler.CreateStreamRequest)
		group.GET("/streams/:id", handler.GetStreamStatus)
		group.GET("/streams", handler.ListStreams)
		group.DELETE("/streams/:id", handler.CancelStream)
		group.POST("/subscriptions", handler.Subscribe)
		group.DELETE("/subscriptions/:id", handler.Unsubscribe)
	}
}

// StreamManager interface
type StreamManager interface {
	CreateStream(req *StreamRequest) (*StreamSession, error)
	GetStream(id string) (*StreamSession, error)
	ListStreams(tenantID, status string) ([]*StreamSession, error)
	CancelStream(id string) error
	Subscribe(sub *StreamSubscription) (*StreamSubscription, error)
	Unsubscribe(id string) error
}
