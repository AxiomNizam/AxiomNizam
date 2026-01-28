package jobs

import "time"

// JobStatus and JobType are defined in job.go
// Job struct and related models below

// Note: Job struct is defined in job.go, this file contains additional API models

// JobError represents error details
type JobError struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

// JobResourceQuota defines resource limits
type JobResourceQuota struct {
	CPUMillis  int64 `json:"cpuMillis"`  // CPU millicores
	MemoryMB   int64 `json:"memoryMb"`   // Memory in MB
	DiskMB     int64 `json:"diskMb"`     // Disk in MB
	TimeoutSec int64 `json:"timeoutSec"` // Max execution time
}

// JobNotification defines notification on job completion
type JobNotification struct {
	Type       string `json:"type"`     // "webhook", "email", "slack"
	Endpoint   string `json:"endpoint"` // URL or email
	OnSuccess  bool   `json:"onSuccess"`
	OnFailure  bool   `json:"onFailure"`
	OnProgress bool   `json:"onProgress"`
}

// JobLog represents job execution log
type JobLog struct {
	ID        string                 `json:"id"`
	JobID     string                 `json:"jobId"`
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"` // INFO, WARN, ERROR
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context"` // Additional data
}

// JobMetrics tracks job statistics
type JobMetrics struct {
	JobID            string        `json:"jobId"`
	TenantID         string        `json:"tenantId"`
	Type             JobType       `json:"type"`
	Duration         time.Duration `json:"duration"`
	RecordsProcessed int64         `json:"recordsProcessed"`
	RecordsFailed    int64         `json:"recordsFailed"`
	BytesRead        int64         `json:"bytesRead"`
	BytesWritten     int64         `json:"bytesWritten"`
	CPUUsage         float64       `json:"cpuUsage"` // Percentage
	MemoryPeakMB     int64         `json:"memoryPeakMb"`
}

// JobFilter for querying jobs
type JobFilter struct {
	TenantID  string
	UserID    string
	Type      JobType
	Status    JobStatus
	Priority  int
	Tags      map[string]string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Offset    int
	SortBy    string // "createdAt", "priority", "progress"
	SortOrder string // "asc", "desc"
}

// JobSubmitRequest for submitting jobs
type JobSubmitRequest struct {
	TenantID      string                 `json:"tenantId"`
	Type          JobType                `json:"type"`
	Priority      int                    `json:"priority"`
	Input         map[string]interface{} `json:"input"`
	Timeout       int                    `json:"timeout"`
	MaxRetries    int                    `json:"maxRetries"`
	DependsOn     []string               `json:"dependsOn"`
	Tags          map[string]string      `json:"tags"`
	ResourceQuota JobResourceQuota       `json:"resourceQuota"`
	Notifications []JobNotification      `json:"notifications"`
}

// JobResponse for API responses
type JobResponse struct {
	ID          string     `json:"id"`
	Status      JobStatus  `json:"status"`
	Progress    int        `json:"progress"`
	Error       *JobError  `json:"error"`
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// JobExecutionStats aggregate statistics
type JobExecutionStats struct {
	TotalJobs       int64               `json:"totalJobs"`
	CompletedJobs   int64               `json:"completedJobs"`
	FailedJobs      int64               `json:"failedJobs"`
	AverageDuration time.Duration       `json:"averageDuration"`
	SuccessRate     float64             `json:"successRate"`
	ByType          map[JobType]int64   `json:"byType"`
	ByStatus        map[JobStatus]int64 `json:"byStatus"`
}
