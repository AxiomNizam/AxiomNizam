package jobs

import "time"

// Job represents an async long-running operation
type Job struct {
	ID            string                 `json:"id" db:"id"`
	TenantID      string                 `json:"tenantId" db:"tenant_id"`
	UserID        string                 `json:"userId" db:"user_id"`
	Type          JobType                `json:"type" db:"type"`         // Export, Import, Transform, etc
	Status        JobStatus              `json:"status" db:"status"`     // Pending, Running, Succeeded, Failed
	Priority      int                    `json:"priority" db:"priority"` // 1-10, higher runs first
	Progress      int                    `json:"progress" db:"progress"` // 0-100
	Input         map[string]interface{} `json:"input" db:"-"`           // Job parameters
	Output        map[string]interface{} `json:"output" db:"-"`          // Job results
	Error         *JobError              `json:"error" db:"-"`           // Error details if failed
	StartedAt     *time.Time             `json:"startedAt" db:"started_at"`
	CompletedAt   *time.Time             `json:"completedAt" db:"completed_at"`
	CreatedAt     time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time              `json:"updatedAt" db:"updated_at"`
	Timeout       int                    `json:"timeout" db:"timeout"` // Seconds
	RetryCount    int                    `json:"retryCount" db:"retry_count"`
	MaxRetries    int                    `json:"maxRetries" db:"max_retries"`
	ParentJobID   *string                `json:"parentJobId" db:"parent_job_id"` // For dependent jobs
	DependsOn     []string               `json:"dependsOn" db:"-"`               // Job IDs this depends on
	ResultURL     string                 `json:"resultUrl" db:"result_url"`      // S3/blob URL for large results
	Tags          map[string]string      `json:"tags" db:"-"`                    // Custom labels
	ResourceQuota JobResourceQuota       `json:"resourceQuota" db:"-"`           // CPU/Memory limits
	Notifications []JobNotification      `json:"notifications" db:"-"`           // Webhooks/events on completion
}

// JobType represents type of job
type JobType string

const (
	JobTypeExport      JobType = "EXPORT"      // Data export
	JobTypeImport      JobType = "IMPORT"      // Data import
	JobTypeTransform   JobType = "TRANSFORM"   // Data transformation
	JobTypeQuery       JobType = "QUERY"       // Long-running query
	JobTypeBackup      JobType = "BACKUP"      // Database backup
	JobTypeRestore     JobType = "RESTORE"     // Database restore
	JobTypeAnalytics   JobType = "ANALYTICS"   // Analytics processing
	JobTypeMigration   JobType = "MIGRATION"   // Data migration
	JobTypeBulkDelete  JobType = "BULK_DELETE" // Bulk deletion
	JobTypeMaintenance JobType = "MAINTENANCE" // Maintenance task
)

// JobStatus represents job lifecycle status
type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusQueued    JobStatus = "QUEUED"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusSucceeded JobStatus = "SUCCEEDED"
	JobStatusFailed    JobStatus = "FAILED"
	JobStatusCancelled JobStatus = "CANCELLED"
	JobStatusPaused    JobStatus = "PAUSED"
	JobStatusRetrying  JobStatus = "RETRYING"
)

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
