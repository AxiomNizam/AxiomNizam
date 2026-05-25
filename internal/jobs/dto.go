package jobs

import "time"

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
	Status  string `json:"status,omitempty"`
}

// --- Observability Response DTOs ---

type JobStatsResponse struct {
	Total     int64 `json:"total"`
	Pending   int64 `json:"pending"`
	Running   int64 `json:"running"`
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
	Cancelled int64 `json:"cancelled"`
}

type HealthStatusResponse struct {
	Status string `json:"status"`
}

type JobMetricsResponse struct {
	ID          string    `json:"id"`
	Status      JobStatus `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	DurationMs  int64     `json:"duration_ms"`
	Retries     int       `json:"retries"`
	Error       string    `json:"error,omitempty"`
}

type JobsByStatusResponse struct {
	Count int    `json:"count"`
	Jobs  []*Job `json:"jobs"`
}

type JobsByTypeResponse struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
	Jobs  []*Job `json:"jobs"`
}

type QueueHealthResponse struct {
	Status  string `json:"status"`
	Pending int    `json:"pending"`
}

type QueueDepthResponse struct {
	Total   int `json:"depth"`
	Pending int `json:"pending"`
}

type ProcessorStatsResponse struct {
	WorkersActive int     `json:"workers_active"`
	WorkersTotal  int     `json:"workers_total"`
	JobsProcessed int64   `json:"jobs_processed"`
	JobsSucceeded int64   `json:"jobs_succeeded"`
	JobsFailed    int64   `json:"jobs_failed"`
	SuccessRate   float64 `json:"success_rate"`
}

type WorkerInfoResponse struct {
	Active      int     `json:"active"`
	Total       int     `json:"total"`
	Utilization float64 `json:"utilization_percent"`
}

type ProcessorHealthResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type SystemInfoResponse struct {
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// JobListResponse is the API response for listing jobs.
type JobListResponse struct {
	Jobs  []*Job `json:"jobs"`
	Count int    `json:"count"`
}

// JobProgressResponse is the API response for job progress.
type JobProgressResponse struct {
	ID     string                 `json:"id"`
	Status JobStatus              `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// JobLogsResponse is the API response for job logs.
type JobLogsResponse struct {
	Logs []*JobLog `json:"logs"`
}

// ScheduleItem is a single schedule entry.
type ScheduleItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Expression string `json:"expression"`
	Enabled    bool   `json:"enabled"`
	LastRun    string `json:"lastRun,omitempty"`
	NextRun    string `json:"nextRun,omitempty"`
	Phase      string `json:"phase"`
}

// ScheduleListResponse is the API response for listing schedules.
type ScheduleListResponse struct {
	Schedules []ScheduleItem `json:"schedules"`
	Count     int            `json:"count"`
}

// LegacyJobLogsResponse is the API response for legacy job logs.
type LegacyJobLogsResponse struct {
	JobID string   `json:"jobId"`
	Name  string   `json:"name"`
	Logs  []string `json:"logs"`
}
