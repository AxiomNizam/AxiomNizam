package jobs

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HTTPJobHandler handles job HTTP endpoints
type HTTPJobHandler struct {
	manager JobManager
}

// NewJobHandler creates handler
func NewJobHandler(manager JobManager) *HTTPJobHandler {
	return &HTTPJobHandler{manager: manager}
}

// SubmitJob handles POST /api/v1/jobs
func (h *HTTPJobHandler) SubmitJob(c *gin.Context) {
	var req struct {
		TenantID    string                 `json:"tenantId" binding:"required"`
		Type        JobType                `json:"type" binding:"required"`
		Params      map[string]interface{} `json:"params"`
		Priority    int                    `json:"priority"`
		SubmittedBy string                 `json:"submittedBy" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := &Job{
		ID:        "job-" + uuid.New().String()[:8],
		Type:      JobType(req.Type),
		Status:    JobStatus("Pending"),
		Data:      req.Params,
		CreatedAt: time.Now(),
	}

	created, err := h.manager.SubmitJob(job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, created)
}

// GetJob handles GET /api/v1/jobs/:id
func (h *HTTPJobHandler) GetJob(c *gin.Context) {
	id := c.Param("id")
	job, err := h.manager.GetJob(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListJobs handles GET /api/v1/jobs
func (h *HTTPJobHandler) ListJobs(c *gin.Context) {
	tenantID := c.Query("tenantId")
	status := c.Query("status")
	jobType := c.Query("type")

	filter := &JobFilter{
		TenantID: tenantID,
		Status:   JobStatus(status),
		Type:     JobType(jobType),
		Limit:    100,
	}

	jobs, err := h.manager.ListJobs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs, "count": len(jobs)})
}

// CancelJob handles DELETE /api/v1/jobs/:id
func (h *HTTPJobHandler) CancelJob(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.CancelJob(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "job cancelled"})
}

// GetJobProgress handles GET /api/v1/jobs/:id/progress
func (h *HTTPJobHandler) GetJobProgress(c *gin.Context) {
	id := c.Param("id")
	job, err := h.manager.GetJob(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     job.ID,
		"status": job.Status,
		"data":   job.Data,
	})
}

// RetryJob handles POST /api/v1/jobs/:id/retry
func (h *HTTPJobHandler) RetryJob(c *gin.Context) {
	id := c.Param("id")
	retried, err := h.manager.RetryJob(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, retried)
}

// GetJobLogs handles GET /api/v1/jobs/:id/logs
func (h *HTTPJobHandler) GetJobLogs(c *gin.Context) {
	id := c.Param("id")
	logs, err := h.manager.GetJobLogs(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "logs not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// RegisterJobRoutes registers all job routes
func RegisterJobRoutes(router *gin.Engine, manager JobManager) {
	handler := NewJobHandler(manager)

	group := router.Group("/api/v1/jobs")
	{
		group.POST("", handler.SubmitJob)
		group.GET("", handler.ListJobs)
		group.GET("/:id", handler.GetJob)
		group.DELETE("/:id", handler.CancelJob)
		group.GET("/:id/progress", handler.GetJobProgress)
		group.POST("/:id/retry", handler.RetryJob)
		group.GET("/:id/logs", handler.GetJobLogs)
	}
}

// JobManager interface
type JobManager interface {
	SubmitJob(job *Job) (*Job, error)
	GetJob(id string) (*Job, error)
	ListJobs(filter *JobFilter) ([]*Job, error)
	CancelJob(id string) error
	RetryJob(id string) (*Job, error)
	GetJobLogs(id string) ([]*JobLog, error)
	ScheduleJob(jobType JobType, interval string, data map[string]interface{}) error
	GetJobStats(ctx context.Context) (*QueueStats, error)
	GetProcessorStats() *ProcessorStats
	Health() error
}
