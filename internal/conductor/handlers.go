package conductor

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// isConfigError returns true if the error is a configuration/client problem, not a server error.
func isConfigError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "not configured") || strings.Contains(msg, "unsupported backend")
}

// Handler serves the conductor REST + WebSocket API.
type Handler struct {
	mgr      *Manager
	upgrader websocket.Upgrader
}

// NewHandler creates a new conductor handler.
func NewHandler(mgr *Manager) *Handler {
	return &Handler{
		mgr: mgr,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// ---------------------------------------------------------------
// Producers
// ---------------------------------------------------------------

// CreateProducer POST /api/v1/conductor/producers
func (h *Handler) CreateProducer(c *gin.Context) {
	var req CreateProducerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.mgr.CreateProducer(&req)
	if err != nil {
		status := http.StatusInternalServerError
		if isConfigError(err) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// ListProducers GET /api/v1/conductor/producers
func (h *Handler) ListProducers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"producers": h.mgr.ListProducers()})
}

// GetProducer GET /api/v1/conductor/producers/:id
func (h *Handler) GetProducer(c *gin.Context) {
	p, err := h.mgr.GetProducer(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// UpdateProducer PATCH /api/v1/conductor/producers/:id
func (h *Handler) UpdateProducer(c *gin.Context) {
	var req CreateProducerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.mgr.UpdateProducer(c.Param("id"), &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// DeleteProducer DELETE /api/v1/conductor/producers/:id
func (h *Handler) DeleteProducer(c *gin.Context) {
	if err := h.mgr.DeleteProducer(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "producer deleted"})
}

// PauseProducer POST /api/v1/conductor/producers/:id/pause
func (h *Handler) PauseProducer(c *gin.Context) {
	p, err := h.mgr.PauseProducer(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// ResumeProducer POST /api/v1/conductor/producers/:id/resume
func (h *Handler) ResumeProducer(c *gin.Context) {
	p, err := h.mgr.ResumeProducer(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// ---------------------------------------------------------------
// Consumers
// ---------------------------------------------------------------

// CreateConsumer POST /api/v1/conductor/consumers
func (h *Handler) CreateConsumer(c *gin.Context) {
	var req CreateConsumerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cons, err := h.mgr.CreateConsumer(&req)
	if err != nil {
		status := http.StatusInternalServerError
		if isConfigError(err) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cons)
}

// ListConsumers GET /api/v1/conductor/consumers
func (h *Handler) ListConsumers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"consumers": h.mgr.ListConsumers()})
}

// GetConsumer GET /api/v1/conductor/consumers/:id
func (h *Handler) GetConsumer(c *gin.Context) {
	cons, err := h.mgr.GetConsumer(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cons)
}

// DeleteConsumer DELETE /api/v1/conductor/consumers/:id
func (h *Handler) DeleteConsumer(c *gin.Context) {
	if err := h.mgr.DeleteConsumer(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "consumer deleted"})
}

// UpdateConsumer PATCH /api/v1/conductor/consumers/:id
func (h *Handler) UpdateConsumer(c *gin.Context) {
	var req CreateConsumerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cons, err := h.mgr.UpdateConsumer(c.Param("id"), &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cons)
}

// PauseConsumer POST /api/v1/conductor/consumers/:id/pause
func (h *Handler) PauseConsumer(c *gin.Context) {
	cons, err := h.mgr.PauseConsumer(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cons)
}

// ResumeConsumer POST /api/v1/conductor/consumers/:id/resume
func (h *Handler) ResumeConsumer(c *gin.Context) {
	cons, err := h.mgr.ResumeConsumer(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cons)
}

// ---------------------------------------------------------------
// Publish
// ---------------------------------------------------------------

// Publish POST /api/v1/conductor/publish
func (h *Handler) Publish(c *gin.Context) {
	var req PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	msg, err := h.mgr.Publish(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, msg)
}

// ---------------------------------------------------------------
// Messages & DLQ
// ---------------------------------------------------------------

// ListMessages GET /api/v1/conductor/messages
func (h *Handler) ListMessages(c *gin.Context) {
	limit := 100
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	c.JSON(http.StatusOK, gin.H{"messages": h.mgr.ListMessages(limit)})
}

// ListDLQ GET /api/v1/conductor/dlq
func (h *Handler) ListDLQ(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"dlq": h.mgr.ListDLQ()})
}

// ReplayDLQ POST /api/v1/conductor/dlq/:id/replay
func (h *Handler) ReplayDLQ(c *gin.Context) {
	msg, err := h.mgr.ReplayDLQ(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "replayed", "msg": msg})
}

// ---------------------------------------------------------------
// Stats
// ---------------------------------------------------------------

// GetStats GET /api/v1/conductor/stats
func (h *Handler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, h.mgr.GetStats())
}

// ---------------------------------------------------------------
// Live Stream (SSE)
// ---------------------------------------------------------------

// StreamSSE GET /api/v1/conductor/stream (Server-Sent Events)
func (h *Handler) StreamSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	lastLen := 0
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msgs := h.mgr.GetStream(50)
			if len(msgs) != lastLen {
				lastLen = len(msgs)
				data, _ := h.mgr.GetStreamJSON(50)
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}

// StreamWS GET /ws/conductor (WebSocket live stream)
func (h *Handler) StreamWS(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastLen := 0
	for {
		select {
		case <-ticker.C:
			msgs := h.mgr.GetStream(50)
			if len(msgs) != lastLen {
				lastLen = len(msgs)
				if err := conn.WriteJSON(gin.H{
					"type":     "conductor_stream",
					"messages": msgs,
					"stats":    h.mgr.GetStats(),
				}); err != nil {
					return
				}
			}
		}
	}
}

// ---------------------------------------------------------------
// Route registration
// ---------------------------------------------------------------

// RegisterRoutes registers all conductor routes on the given router.
func RegisterRoutes(router *gin.Engine, mgr *Manager, authMiddleware, adminMiddleware gin.HandlerFunc) {
	h := NewHandler(mgr)

	api := router.Group("/api/v1/conductor", authMiddleware)
	{
		// Producers
		api.POST("/producers", adminMiddleware, h.CreateProducer)
		api.GET("/producers", h.ListProducers)
		api.GET("/producers/:id", h.GetProducer)
		api.PATCH("/producers/:id", adminMiddleware, h.UpdateProducer)
		api.DELETE("/producers/:id", adminMiddleware, h.DeleteProducer)
		api.POST("/producers/:id/pause", adminMiddleware, h.PauseProducer)
		api.POST("/producers/:id/resume", adminMiddleware, h.ResumeProducer)

		// Consumers
		api.POST("/consumers", adminMiddleware, h.CreateConsumer)
		api.GET("/consumers", h.ListConsumers)
		api.GET("/consumers/:id", h.GetConsumer)
		api.PATCH("/consumers/:id", adminMiddleware, h.UpdateConsumer)
		api.DELETE("/consumers/:id", adminMiddleware, h.DeleteConsumer)
		api.POST("/consumers/:id/pause", adminMiddleware, h.PauseConsumer)
		api.POST("/consumers/:id/resume", adminMiddleware, h.ResumeConsumer)

		// Publish
		api.POST("/publish", adminMiddleware, h.Publish)

		// Messages & DLQ
		api.GET("/messages", h.ListMessages)
		api.GET("/dlq", h.ListDLQ)
		api.POST("/dlq/:id/replay", adminMiddleware, h.ReplayDLQ)

		// Stats & Stream
		api.GET("/stats", h.GetStats)
		api.GET("/stream", h.StreamSSE)
	}

	// WebSocket endpoint
	router.GET("/ws/conductor", authMiddleware, h.StreamWS)
}
