package jobs

// MessageResponse is a generic error/ack response.
type MessageResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
	Status  string `json:"status,omitempty"`
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
