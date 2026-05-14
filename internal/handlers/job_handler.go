package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"example.com/axiomnizam/internal/logging"

	"go.uber.org/zap"

	"example.com/axiomnizam/internal/workqueue"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// JobResource represents a job resource on the server
type JobResource struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   JobMetadata            `json:"metadata"`
	Spec       map[string]interface{} `json:"spec,omitempty"`
	Status     JobStatus              `json:"status,omitempty"`
	Schedule   *JobSchedule           `json:"schedule,omitempty"`
	Logs       []string               `json:"logs,omitempty"`
}

// JobSchedule holds scheduling metadata for recurring jobs.
type JobSchedule struct {
	Expression string `json:"expression"`
	Enabled    bool   `json:"enabled"`
	LastRun    string `json:"lastRun,omitempty"`
	NextRun    string `json:"nextRun,omitempty"`
}

// JobMetadata holds job metadata
type JobMetadata struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Namespace         string `json:"namespace,omitempty"`
	CreationTimestamp string `json:"creationTimestamp,omitempty"`
}

// JobStatus holds job execution status
type JobStatus struct {
	Phase     string `json:"phase"`
	Progress  int    `json:"progress"`
	LastRun   string `json:"lastRun,omitempty"`
	NextRun   string `json:"nextRun,omitempty"`
	LastError string `json:"lastError,omitempty"`
}

// JobHandler manages job resources
type JobHandler struct {
	mu              sync.RWMutex
	jobs            map[string]*JobResource // job ID -> job
	etcd            *clientv3.Client
	stateKey        string
	schedulerCtx    context.Context
	schedulerCancel context.CancelFunc
	schedulerWG     sync.WaitGroup

	// P1.5: execution is driven by a rate-limited workqueue + worker
	// pool rather than fire-and-forget `go h.executeJob(id)` goroutines.
	execQueue   workqueue.WorkQueue
	execWorkers int
}

// NewJobHandler creates a new job handler
func NewJobHandler(etcd ...*clientv3.Client) *JobHandler {
	var etcdClient *clientv3.Client
	if len(etcd) > 0 {
		etcdClient = etcd[0]
	}

	h := &JobHandler{
		jobs:        make(map[string]*JobResource),
		etcd:        etcdClient,
		stateKey:    "axiomnizam:jobs:state",
		execQueue:   workqueue.NewSimpleQueue(nil),
		execWorkers: 4,
	}
	h.loadState()
	h.startScheduler()
	return h
}

// Close stops the background scheduler loop.
func (h *JobHandler) Close() {
	h.mu.Lock()
	cancel := h.schedulerCancel
	h.schedulerCancel = nil
	h.schedulerCtx = nil
	q := h.execQueue
	h.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if q != nil {
		_ = q.Shutdown()
	}
	h.schedulerWG.Wait()
}

func (h *JobHandler) startScheduler() {
	h.mu.Lock()
	if h.schedulerCancel != nil {
		h.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	h.schedulerCtx = ctx
	h.schedulerCancel = cancel
	workers := h.execWorkers
	queue := h.execQueue
	h.mu.Unlock()

	// Cron / interval scanner: re-evaluates scheduled jobs once a second
	// and pushes due jobs onto the rate-limited workqueue.
	h.schedulerWG.Add(1)
	go func() {
		defer h.schedulerWG.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.processScheduledJobs()
			}
		}
	}()

	// P1.5: worker pool that drains the execution queue.  Replaces the
	// per-execution `go h.executeJob(id)` pattern.
	process := func(ctx context.Context, item *workqueue.Item) error {
		h.executeJob(item.Key)
		return nil
	}
	for i := 0; i < workers; i++ {
		h.schedulerWG.Add(1)
		go func() {
			defer h.schedulerWG.Done()
			w := workqueue.NewWorker(queue, process, 5)
			if err := w.Run(ctx); err != nil {
				logging.Z().Warn("jobs: exec worker exited", zap.Error(err))
			}
		}()
	}
}

func getScheduleExpression(spec map[string]interface{}) (string, bool) {
	if spec == nil {
		return "", false
	}
	for _, key := range []string{"schedule", "interval"} {
		if raw, ok := spec[key]; ok {
			expr := strings.TrimSpace(fmt.Sprintf("%v", raw))
			if expr != "" {
				return expr, true
			}
		}
	}
	return "", false
}

func calculateNextRun(expression string, from time.Time) (time.Time, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return time.Time{}, fmt.Errorf("schedule expression is required")
	}

	if duration, err := time.ParseDuration(expression); err == nil {
		if duration <= 0 {
			return time.Time{}, fmt.Errorf("duration must be positive")
		}
		return from.Add(duration), nil
	}

	schedule, err := cron.ParseStandard(expression)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid schedule expression: %w", err)
	}

	nextRun := schedule.Next(from)
	if nextRun.IsZero() {
		return time.Time{}, fmt.Errorf("could not determine next run for expression")
	}

	return nextRun, nil
}

func normalizeScheduleFromSpec(job *JobResource, now time.Time) error {
	if job == nil {
		return nil
	}
	if job.Schedule != nil {
		return nil
	}

	expression, ok := getScheduleExpression(job.Spec)
	if !ok {
		return nil
	}

	nextRun, err := calculateNextRun(expression, now)
	if err != nil {
		return err
	}

	job.Schedule = &JobSchedule{
		Expression: expression,
		Enabled:    true,
		NextRun:    nextRun.UTC().Format(time.RFC3339),
	}
	job.Status.NextRun = job.Schedule.NextRun
	return nil
}

func (h *JobHandler) processScheduledJobs() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now().UTC()
	for _, job := range h.jobs {
		if job.Schedule == nil || !job.Schedule.Enabled {
			continue
		}

		nextRun, err := time.Parse(time.RFC3339, strings.TrimSpace(job.Schedule.NextRun))
		if err != nil {
			nextRun, err = calculateNextRun(job.Schedule.Expression, now)
			if err != nil {
				job.Logs = append(job.Logs, fmt.Sprintf("[%s] Invalid schedule '%s': %v", now.Format(time.RFC3339), job.Schedule.Expression, err))
				continue
			}
			job.Schedule.NextRun = nextRun.UTC().Format(time.RFC3339)
			job.Status.NextRun = job.Schedule.NextRun
		}

		if now.Before(nextRun) {
			continue
		}

		if strings.EqualFold(job.Status.Phase, "Running") {
			continue
		}

		h.startJobExecutionLocked(job, "scheduled")
	}
}

func (h *JobHandler) startJobExecutionLocked(job *JobResource, trigger string) {
	now := time.Now().UTC()
	job.Status.Phase = "Running"
	job.Status.Progress = 0
	job.Status.LastRun = now.Format(time.RFC3339)
	if job.Schedule != nil {
		job.Schedule.LastRun = job.Status.LastRun
	}
	job.Logs = append(job.Logs, fmt.Sprintf("[%s] Job started (%s)", now.Format(time.RFC3339), trigger))
	h.persistStateLocked()

	jobID := job.Metadata.ID
	// P1.5: hand off to the rate-limited workqueue instead of spawning
	// an untracked `go h.executeJob(jobID)` goroutine.
	if h.execQueue != nil {
		_ = h.execQueue.Add(jobID)
	}
}

func (h *JobHandler) executeJob(jobID string) {
	for i := 1; i <= 10; i++ {
		time.Sleep(200 * time.Millisecond)
		h.mu.Lock()
		job := h.findJob(jobID)
		if job == nil {
			h.mu.Unlock()
			return
		}
		if strings.EqualFold(job.Status.Phase, "Cancelled") {
			h.persistStateLocked()
			h.mu.Unlock()
			return
		}
		job.Status.Progress = i * 10
		job.Logs = append(job.Logs, fmt.Sprintf("[%s] Progress: %d%%", time.Now().UTC().Format(time.RFC3339), i*10))
		h.persistStateLocked()
		h.mu.Unlock()
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	job := h.findJob(jobID)
	if job == nil {
		return
	}

	job.Status.Phase = "Succeeded"
	job.Status.Progress = 100
	job.Logs = append(job.Logs, fmt.Sprintf("[%s] Job completed successfully", time.Now().UTC().Format(time.RFC3339)))

	if job.Schedule != nil && job.Schedule.Enabled {
		nextRun, err := calculateNextRun(job.Schedule.Expression, time.Now().UTC())
		if err != nil {
			job.Status.LastError = err.Error()
			job.Logs = append(job.Logs, fmt.Sprintf("[%s] Failed to compute next run: %v", time.Now().UTC().Format(time.RFC3339), err))
			job.Status.NextRun = ""
			job.Schedule.NextRun = ""
		} else {
			job.Schedule.NextRun = nextRun.UTC().Format(time.RFC3339)
			job.Status.NextRun = job.Schedule.NextRun
		}
	} else {
		job.Status.NextRun = ""
	}

	h.persistStateLocked()
}

func (h *JobHandler) loadState() {
	if h.etcd == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.etcd.Get(ctx, h.stateKey)
	if err != nil {
		logging.Z().Warn("jobs: failed to load persisted state", zap.Error(err))
		return
	}
	if len(resp.Kvs) == 0 {
		return
	}

	var jobs map[string]*JobResource
	if err := json.Unmarshal(resp.Kvs[0].Value, &jobs); err != nil {
		logging.Z().Warn("jobs: failed to decode persisted state", zap.Error(err))
		return
	}
	if jobs == nil {
		jobs = make(map[string]*JobResource)
	}
	h.jobs = jobs
}

func (h *JobHandler) persistStateLocked() {
	if h.etcd == nil {
		return
	}

	payload, err := json.Marshal(h.jobs)
	if err != nil {
		logging.Z().Warn("jobs: failed to encode state", zap.Error(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.etcd.Put(ctx, h.stateKey, string(payload)); err != nil {
		logging.Z().Warn("jobs: failed to persist state", zap.Error(err))
	}
}

// Create creates a new job
func (h *JobHandler) Create(c *gin.Context) {
	var job JobResource
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	id := uuid.New().String()[:8]
	job.Metadata.ID = id
	job.Kind = "Job"
	job.APIVersion = "axiom-nizam.io/v1"
	job.Metadata.CreationTimestamp = time.Now().UTC().Format(time.RFC3339)
	job.Status = JobStatus{
		Phase:    "Pending",
		Progress: 0,
	}

	if job.Metadata.Name == "" {
		job.Metadata.Name = fmt.Sprintf("job-%s", id)
	}

	if err := normalizeScheduleFromSpec(&job, time.Now().UTC()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.jobs[id] = &job
	h.persistStateLocked()

	c.JSON(http.StatusCreated, &job)
}

// List lists all jobs
func (h *JobHandler) List(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	items := make([]*JobResource, 0, len(h.jobs))
	for _, j := range h.jobs {
		items = append(items, j)
	}

	c.JSON(http.StatusOK, items)
}

// ListSchedules lists all job schedules.
func (h *JobHandler) ListSchedules(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	items := make([]gin.H, 0)
	for _, job := range h.jobs {
		if job.Schedule == nil {
			continue
		}
		items = append(items, gin.H{
			"id":         job.Metadata.ID,
			"name":       job.Metadata.Name,
			"expression": job.Schedule.Expression,
			"enabled":    job.Schedule.Enabled,
			"lastRun":    job.Schedule.LastRun,
			"nextRun":    job.Schedule.NextRun,
			"phase":      job.Status.Phase,
		})
	}

	c.JSON(http.StatusOK, gin.H{"schedules": items, "count": len(items)})
}

// Get returns a job by ID
func (h *JobHandler) Get(c *gin.Context) {
	id := c.Param("id")

	h.mu.RLock()
	defer h.mu.RUnlock()

	// Search by ID or name
	job := h.findJob(id)
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("job '%s' not found", id)})
		return
	}

	c.JSON(http.StatusOK, job)
}

// Run runs a job
func (h *JobHandler) Run(c *gin.Context) {
	id := c.Param("id")

	h.mu.Lock()

	job := h.findJob(id)
	if job == nil {
		h.mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("job '%s' not found", id)})
		return
	}
	if strings.EqualFold(job.Status.Phase, "Running") {
		h.mu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("job '%s' is already running", id)})
		return
	}

	h.startJobExecutionLocked(job, "manual")
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("job '%s' started", id), "status": "Running"})
}

// SetSchedule creates or updates a job schedule.
func (h *JobHandler) SetSchedule(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Schedule   string `json:"schedule"`
		Expression string `json:"expression"`
		Interval   string `json:"interval"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expression := strings.TrimSpace(req.Expression)
	if expression == "" {
		expression = strings.TrimSpace(req.Schedule)
	}
	if expression == "" {
		expression = strings.TrimSpace(req.Interval)
	}
	if expression == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schedule expression is required"})
		return
	}

	nextRun, err := calculateNextRun(expression, time.Now().UTC())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	job := h.findJob(id)
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("job '%s' not found", id)})
		return
	}

	if job.Spec == nil {
		job.Spec = map[string]interface{}{}
	}
	job.Spec["schedule"] = expression

	if job.Schedule == nil {
		job.Schedule = &JobSchedule{}
	}
	job.Schedule.Expression = expression
	job.Schedule.Enabled = true
	job.Schedule.NextRun = nextRun.UTC().Format(time.RFC3339)
	job.Status.NextRun = job.Schedule.NextRun
	job.Logs = append(job.Logs, fmt.Sprintf("[%s] Schedule set to '%s'", time.Now().UTC().Format(time.RFC3339), expression))
	h.persistStateLocked()

	c.JSON(http.StatusOK, gin.H{"message": "schedule updated", "job": job})
}

// RemoveSchedule removes a job schedule.
func (h *JobHandler) RemoveSchedule(c *gin.Context) {
	id := c.Param("id")

	h.mu.Lock()
	defer h.mu.Unlock()

	job := h.findJob(id)
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("job '%s' not found", id)})
		return
	}
	if job.Schedule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("job '%s' has no schedule", id)})
		return
	}

	job.Schedule = nil
	job.Status.NextRun = ""
	if job.Spec != nil {
		delete(job.Spec, "schedule")
	}
	job.Logs = append(job.Logs, fmt.Sprintf("[%s] Schedule removed", time.Now().UTC().Format(time.RFC3339)))
	h.persistStateLocked()

	c.JSON(http.StatusOK, gin.H{"message": "schedule removed", "job": job})
}

// GetLogs returns logs for a job
func (h *JobHandler) GetLogs(c *gin.Context) {
	id := c.Param("id")

	h.mu.RLock()
	defer h.mu.RUnlock()

	job := h.findJob(id)
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("job '%s' not found", id)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jobId": job.Metadata.ID,
		"name":  job.Metadata.Name,
		"logs":  job.Logs,
	})
}

// Cancel cancels a running job
func (h *JobHandler) Cancel(c *gin.Context) {
	id := c.Param("id")

	h.mu.Lock()
	defer h.mu.Unlock()

	job := h.findJob(id)
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("job '%s' not found", id)})
		return
	}

	job.Status.Phase = "Cancelled"
	job.Logs = append(job.Logs, fmt.Sprintf("[%s] Job cancelled", time.Now().UTC().Format(time.RFC3339)))
	h.persistStateLocked()

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("job '%s' cancelled", id)})
}

// Delete removes a job
func (h *JobHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	h.mu.Lock()
	defer h.mu.Unlock()

	// Find by ID or name
	found := ""
	for jobID, job := range h.jobs {
		if jobID == id || job.Metadata.Name == id {
			found = jobID
			break
		}
	}

	if found == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("job '%s' not found", id)})
		return
	}

	delete(h.jobs, found)
	h.persistStateLocked()
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("job '%s' deleted", id)})
}

// findJob finds a job by ID or name
func (h *JobHandler) findJob(idOrName string) *JobResource {
	if j, ok := h.jobs[idOrName]; ok {
		return j
	}
	for _, j := range h.jobs {
		if j.Metadata.Name == idOrName {
			return j
		}
	}
	return nil
}
