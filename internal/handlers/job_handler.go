package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// JobResource represents a job resource on the server
type JobResource struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   JobMetadata            `json:"metadata"`
	Spec       map[string]interface{} `json:"spec,omitempty"`
	Status     JobStatus              `json:"status,omitempty"`
	Logs       []string               `json:"logs,omitempty"`
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
	mu       sync.RWMutex
	jobs     map[string]*JobResource // job ID -> job
	etcd     *clientv3.Client
	stateKey string
}

// NewJobHandler creates a new job handler
func NewJobHandler(etcd ...*clientv3.Client) *JobHandler {
	var etcdClient *clientv3.Client
	if len(etcd) > 0 {
		etcdClient = etcd[0]
	}

	h := &JobHandler{
		jobs:     make(map[string]*JobResource),
		etcd:     etcdClient,
		stateKey: "axiomnizam:jobs:state",
	}
	h.loadState()
	return h
}

func (h *JobHandler) loadState() {
	if h.etcd == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.etcd.Get(ctx, h.stateKey)
	if err != nil {
		log.Printf("jobs: failed to load persisted state from etcd: %v", err)
		return
	}
	if len(resp.Kvs) == 0 {
		return
	}

	var jobs map[string]*JobResource
	if err := json.Unmarshal(resp.Kvs[0].Value, &jobs); err != nil {
		log.Printf("jobs: failed to decode persisted state: %v", err)
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
		log.Printf("jobs: failed to encode state: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.etcd.Put(ctx, h.stateKey, string(payload)); err != nil {
		log.Printf("jobs: failed to persist state to etcd: %v", err)
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

	job.Status.Phase = "Running"
	job.Status.Progress = 0
	job.Status.LastRun = time.Now().UTC().Format(time.RFC3339)
	job.Logs = append(job.Logs, fmt.Sprintf("[%s] Job started", time.Now().UTC().Format(time.RFC3339)))
	h.persistStateLocked()
	h.mu.Unlock()

	// Simulate job execution
	go func() {
		for i := 1; i <= 10; i++ {
			time.Sleep(200 * time.Millisecond)
			h.mu.Lock()
			if j := h.findJob(id); j != nil {
				j.Status.Progress = i * 10
				j.Logs = append(j.Logs, fmt.Sprintf("[%s] Progress: %d%%", time.Now().UTC().Format(time.RFC3339), i*10))
				h.persistStateLocked()
			}
			h.mu.Unlock()
		}
		h.mu.Lock()
		if j := h.findJob(id); j != nil {
			j.Status.Phase = "Succeeded"
			j.Status.Progress = 100
			j.Status.NextRun = time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
			j.Logs = append(j.Logs, fmt.Sprintf("[%s] Job completed successfully", time.Now().UTC().Format(time.RFC3339)))
			h.persistStateLocked()
		}
		h.mu.Unlock()
	}()

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("job '%s' started", id), "status": "Running"})
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
