package jobs

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ObservabilityHandler provides HTTP endpoints for job and event observability
type ObservabilityHandler struct {
	manager          *JobManager
	metricsCollector *MetricsCollector
	healthMetrics    *HealthMetrics
	logger           *log.Logger
}

// NewObservabilityHandler creates a new observability handler
func NewObservabilityHandler(
	manager *JobManager,
	metricsCollector *MetricsCollector,
) *ObservabilityHandler {
	return &ObservabilityHandler{
		manager:          manager,
		metricsCollector: metricsCollector,
		healthMetrics:    NewHealthMetrics(metricsCollector, manager),
		logger:           log.New(log.Writer(), "[OBSERVABILITY] ", log.LstdFlags),
	}
}

// RegisterRoutes registers observability routes
func (oh *ObservabilityHandler) RegisterRoutes(router *gin.Engine) {
	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Job observability
	jobs := router.Group("/api/observability/jobs")
	{
		jobs.GET("/stats", oh.GetJobStats)
		jobs.GET("/health", oh.GetJobHealth)
		jobs.GET("/:id", oh.GetJobDetails)
		jobs.GET("/:id/metrics", oh.GetJobMetrics)
		jobs.GET("/status/:status", oh.GetJobsByStatus)
		jobs.GET("/type/:type", oh.GetJobsByType)
	}

	// Queue observability
	queue := router.Group("/api/observability/queue")
	{
		queue.GET("/stats", oh.GetQueueStats)
		queue.GET("/health", oh.GetQueueHealth)
		queue.GET("/depth", oh.GetQueueDepth)
	}

	// Processor observability
	processor := router.Group("/api/observability/processor")
	{
		processor.GET("/stats", oh.GetProcessorStats)
		processor.GET("/workers", oh.GetWorkerInfo)
		processor.GET("/health", oh.GetProcessorHealth)
	}

	// System observability
	system := router.Group("/api/observability/system")
	{
		system.GET("/health", oh.GetSystemHealth)
		system.GET("/info", oh.GetSystemInfo)
		system.GET("/report", oh.GetFullReport)
	}
}

// GetJobStats returns overall job statistics
func (oh *ObservabilityHandler) GetJobStats(c *gin.Context) {
	stats, err := oh.manager.GetJobStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get job stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     stats.Total,
		"pending":   stats.Pending,
		"running":   stats.Running,
		"completed": stats.Completed,
		"failed":    stats.Failed,
		"cancelled": stats.Cancelled,
	})
}

// GetJobHealth returns job system health status
func (oh *ObservabilityHandler) GetJobHealth(c *gin.Context) {
	err := oh.manager.Health()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// GetJobDetails returns detailed job information
func (oh *ObservabilityHandler) GetJobDetails(c *gin.Context) {
	jobID := c.Param("id")

	job, err := oh.manager.GetJob(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	c.JSON(http.StatusOK, job)
}

// GetJobMetrics returns metrics for a specific job
func (oh *ObservabilityHandler) GetJobMetrics(c *gin.Context) {
	jobID := c.Param("id")

	job, err := oh.manager.GetJob(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	var duration time.Duration
	if !job.CompletedAt.IsZero() && !job.CreatedAt.IsZero() {
		duration = job.CompletedAt.Sub(job.CreatedAt)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           job.ID,
		"status":       job.Status,
		"created_at":   job.CreatedAt,
		"started_at":   job.StartedAt,
		"completed_at": job.CompletedAt,
		"duration_ms":  duration.Milliseconds(),
		"retries":      job.Retries,
		"error":        job.Error,
	})
}

// GetJobsByStatus returns jobs by status
func (oh *ObservabilityHandler) GetJobsByStatus(c *gin.Context) {
	status := c.Param("status")

	jobs, err := oh.manager.queue.GetByStatus(c.Request.Context(), JobStatus(status), 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(jobs),
		"jobs":  jobs,
	})
}

// GetJobsByType returns jobs by type
func (oh *ObservabilityHandler) GetJobsByType(c *gin.Context) {
	jobType := c.Param("type")

	// Get from queue
	jobs, _ := oh.manager.queue.GetByStatus(c.Request.Context(), JobStatusPending, 1000)

	filteredJobs := make([]*Job, 0)
	for _, job := range jobs {
		if job.Type == JobType(jobType) {
			filteredJobs = append(filteredJobs, job)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"type":  jobType,
		"count": len(filteredJobs),
		"jobs":  filteredJobs,
	})
}

// GetQueueStats returns queue statistics
func (oh *ObservabilityHandler) GetQueueStats(c *gin.Context) {
	stats, err := oh.manager.GetJobStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get queue stats",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetQueueHealth returns queue health status
func (oh *ObservabilityHandler) GetQueueHealth(c *gin.Context) {
	stats, err := oh.manager.GetJobStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
		})
		return
	}

	status := "healthy"
	if stats.Pending > 5000 {
		status = "warning"
	}
	if stats.Pending > 10000 {
		status = "critical"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  status,
		"pending": stats.Pending,
	})
}

// GetQueueDepth returns current queue depth
func (oh *ObservabilityHandler) GetQueueDepth(c *gin.Context) {
	stats, err := oh.manager.GetJobStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get queue depth",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"depth":   stats.Total,
		"pending": stats.Pending,
	})
}

// GetProcessorStats returns processor statistics
func (oh *ObservabilityHandler) GetProcessorStats(c *gin.Context) {
	stats := oh.manager.GetProcessorStats()

	successRate := 0.0
	if stats.JobsProcessed > 0 {
		successRate = float64(stats.JobsSucceeded) / float64(stats.JobsProcessed)
	}

	c.JSON(http.StatusOK, gin.H{
		"workers_active": stats.WorkersActive,
		"workers_total":  stats.WorkersTotal,
		"jobs_processed": stats.JobsProcessed,
		"jobs_succeeded": stats.JobsSucceeded,
		"jobs_failed":    stats.JobsFailed,
		"success_rate":   successRate,
	})
}

// GetWorkerInfo returns detailed worker information
func (oh *ObservabilityHandler) GetWorkerInfo(c *gin.Context) {
	stats := oh.manager.GetProcessorStats()

	c.JSON(http.StatusOK, gin.H{
		"active":              stats.WorkersActive,
		"total":               stats.WorkersTotal,
		"utilization_percent": (float64(stats.WorkersActive) / float64(stats.WorkersTotal)) * 100,
	})
}

// GetProcessorHealth returns processor health
func (oh *ObservabilityHandler) GetProcessorHealth(c *gin.Context) {
	stats := oh.manager.GetProcessorStats()

	if stats.WorkersActive == 0 && stats.WorkersTotal > 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"reason": "No active workers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// GetSystemHealth returns overall system health
func (oh *ObservabilityHandler) GetSystemHealth(c *gin.Context) {
	health := oh.healthMetrics.CheckHealth(c.Request.Context())

	status := health["status"].(string)
	statusCode := http.StatusOK

	if status == "degraded" || status == "warning" {
		statusCode = http.StatusAccepted
	} else if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// GetSystemInfo returns system information
func (oh *ObservabilityHandler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":      "Axiom Nizam Job System",
		"version":   "4.0.0",
		"timestamp": time.Now(),
	})
}

// GetFullReport returns comprehensive system report
func (oh *ObservabilityHandler) GetFullReport(c *gin.Context) {
	ctx := c.Request.Context()

	jobStats, _ := oh.manager.GetJobStats(ctx)
	procStats := oh.manager.GetProcessorStats()
	health := oh.healthMetrics.CheckHealth(ctx)

	report := map[string]interface{}{
		"timestamp": time.Now(),
		"status":    health["status"],
		"jobs": map[string]interface{}{
			"total":     jobStats.Total,
			"pending":   jobStats.Pending,
			"running":   jobStats.Running,
			"completed": jobStats.Completed,
			"failed":    jobStats.Failed,
		},
		"processor": map[string]interface{}{
			"workers_active": procStats.WorkersActive,
			"workers_total":  procStats.WorkersTotal,
			"jobs_processed": procStats.JobsProcessed,
			"success_rate":   procStats.SuccessRate,
		},
	}

	c.JSON(http.StatusOK, report)
}

// MetricsMiddleware for Gin
type MetricsGinMiddleware struct {
	collector *MetricsCollector
}

// NewMetricsGinMiddleware creates a new Gin metrics middleware
func NewMetricsGinMiddleware(collector *MetricsCollector) *MetricsGinMiddleware {
	return &MetricsGinMiddleware{
		collector: collector,
	}
}

// Middleware returns a Gin middleware function
func (mgm *MetricsGinMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)

		// Record request metrics
		statusCode := c.Writer.Status()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Log metrics
		log.Printf("[METRICS] %s %s %d took %dms",
			method, path, statusCode, duration.Milliseconds(),
		)
	}
}
