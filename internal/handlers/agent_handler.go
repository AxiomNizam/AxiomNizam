package handlers

import (
	"net/http"
	"time"

	"example.com/axiomnizam/internal/agents"
	"github.com/gin-gonic/gin"
)

// AgentHandler manages cloud agent REST API endpoints
type AgentHandler struct {
	registry   *agents.Registry
	executor   *agents.Executor
	httpClient *http.Client
}

// NewAgentHandler creates a new AgentHandler backed by the global registry and executor
func NewAgentHandler() *AgentHandler {
	return &AgentHandler{
		registry: agents.GlobalRegistry,
		executor: agents.GlobalExecutor,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// RegisterAgent godoc
// POST /api/v1/agents
func (h *AgentHandler) RegisterAgent(c *gin.Context) {
	var req struct {
		Name         string            `json:"name" binding:"required"`
		Description  string            `json:"description"`
		Endpoint     string            `json:"endpoint" binding:"required"`
		Capabilities []string          `json:"capabilities"`
		Labels       map[string]string `json:"labels"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent := &agents.CloudAgent{
		Name:         req.Name,
		Description:  req.Description,
		Endpoint:     req.Endpoint,
		Capabilities: req.Capabilities,
		Labels:       req.Labels,
	}

	if err := h.registry.Register(agent); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, agent)
}

// ListAgents godoc
// GET /api/v1/agents
func (h *AgentHandler) ListAgents(c *gin.Context) {
	c.JSON(http.StatusOK, h.registry.List())
}

// GetAgent godoc
// GET /api/v1/agents/:id
func (h *AgentHandler) GetAgent(c *gin.Context) {
	id := c.Param("id")
	agent, err := h.registry.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, agent)
}

// UnregisterAgent godoc
// DELETE /api/v1/agents/:id
func (h *AgentHandler) UnregisterAgent(c *gin.Context) {
	id := c.Param("id")
	if err := h.registry.Unregister(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "agent unregistered", "id": id})
}

// AgentHealth godoc
// GET /api/v1/agents/:id/health
// Performs a lightweight probe to the agent endpoint and updates its status.
func (h *AgentHandler) AgentHealth(c *gin.Context) {
	id := c.Param("id")
	agent, err := h.registry.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	client := h.httpClient
	resp, err := client.Get(agent.Endpoint + "/health")
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		_ = h.registry.UpdateStatus(id, agents.AgentStatusOffline)
		c.JSON(http.StatusOK, gin.H{"id": id, "status": string(agents.AgentStatusOffline)})
		return
	}
	if resp != nil {
		resp.Body.Close()
	}

	_ = h.registry.UpdateStatus(id, agents.AgentStatusOnline)
	c.JSON(http.StatusOK, gin.H{"id": id, "status": string(agents.AgentStatusOnline)})
}

// DelegateTask godoc
// POST /api/v1/agents/:id/tasks
// Delegates a task to the specified cloud agent (async – returns immediately).
func (h *AgentHandler) DelegateTask(c *gin.Context) {
	id := c.Param("id")

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.executor.DelegateAsync(c.Request.Context(), id, payload)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, task)
}

// GetTask godoc
// GET /api/v1/agents/:id/tasks/:taskId
// Returns the current status of a delegated task, refreshing it from the remote agent.
func (h *AgentHandler) GetTask(c *gin.Context) {
	taskID := c.Param("taskId")

	task, err := h.executor.PollTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		if err == agents.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// ListAgentTasks godoc
// GET /api/v1/agents/:id/tasks
// Returns all tasks that have been delegated to the specified agent.
func (h *AgentHandler) ListAgentTasks(c *gin.Context) {
	id := c.Param("id")

	if _, err := h.registry.Get(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.registry.ListTasksByAgent(id))
}
