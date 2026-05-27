package waitx

import "time"

// ─────────────────────────────────────────────────────────────────────────────
// API Response DTOs
// ─────────────────────────────────────────────────────────────────────────────

// CheckResponse is returned when a single check is executed.
type CheckResponse struct {
	CheckType  string `json:"checkType"`
	Target     string `json:"target"`
	Status     string `json:"status"`
	DurationMs int64  `json:"durationMs"`
	Message    string `json:"message,omitempty"`
	Attempts   int    `json:"attempts"`
}

// WaitResponse is returned when a wait completes.
type WaitResponse struct {
	CheckType  string `json:"checkType"`
	Target     string `json:"target"`
	Ready      bool   `json:"ready"`
	DurationMs int64  `json:"durationMs"`
	Attempts   int    `json:"attempts"`
	Message    string `json:"message,omitempty"`
}

// GroupWaitResponse is returned when a group wait completes.
type GroupWaitResponse struct {
	GroupName  string          `json:"groupName"`
	Ready      bool            `json:"ready"`
	DurationMs int64           `json:"durationMs"`
	Checks     []CheckResponse `json:"checks"`
	ReadyCount int             `json:"readyCount"`
	TotalCount int             `json:"totalCount"`
}

// CheckStatusResponse is the current status of a check.
type CheckStatusResponse struct {
	Name        string    `json:"name"`
	CheckType   string    `json:"checkType"`
	Target      string    `json:"target"`
	Status      string    `json:"status"`
	LastCheckAt time.Time `json:"lastCheckAt,omitempty"`
	Attempts    int       `json:"attempts"`
	LastError   string    `json:"lastError,omitempty"`
}

// HealthResponse is the module health response.
type HealthResponse struct {
	Status      string `json:"status"`
	UptimeSec   int64  `json:"uptimeSeconds"`
	TotalChecks int64  `json:"totalChecks"`
	SuccessRate string `json:"successRate"`
	Module      string `json:"module"`
}

// MetricsResponse holds waitx metrics.
type MetricsResponse struct {
	TotalChecks    int64                `json:"totalChecks"`
	TotalSuccesses int64                `json:"totalSuccesses"`
	TotalFailures  int64                `json:"totalFailures"`
	TotalTimeouts  int64                `json:"totalTimeouts"`
	SuccessRate    string               `json:"successRate"`
	UptimeSeconds  int64                `json:"uptimeSeconds"`
	ByCheckType    []CheckTypeStats     `json:"byCheckType"`
}

// CheckTypeStats holds per-type metrics.
type CheckTypeStats struct {
	CheckType  string  `json:"checkType"`
	Runs       int64   `json:"runs"`
	Successes  int64   `json:"successes"`
	Failures   int64   `json:"failures"`
	Timeouts   int64   `json:"timeouts"`
	TotalMs    int64   `json:"totalMs"`
	AvgMs      float64 `json:"avgMs"`
	SuccessRate string `json:"successRate"`
}

// ─────────────────────────────────────────────────────────────────────────────
// API Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// RunCheckRequest is the body for POST /api/v1/waitx/check.
type RunCheckRequest struct {
	CheckType    string            `json:"checkType" binding:"required"`
	Target       string            `json:"target" binding:"required"`
	Timeout      string            `json:"timeout,omitempty"`
	InvertCheck  bool              `json:"invertCheck,omitempty"`
	RetryPolicy  string            `json:"retryPolicy,omitempty"`
	MaxAttempts  int               `json:"maxAttempts,omitempty"`
	// HTTP-specific
	Method           string            `json:"method,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
	ExpectStatusCode int               `json:"expectStatusCode,omitempty"`
	InsecureSkipTLS  bool              `json:"insecureSkipTLS,omitempty"`
	// DNS-specific
	RecordType     string   `json:"recordType,omitempty"`
	ExpectedValues []string `json:"expectedValues,omitempty"`
	// Redis-specific
	ExpectedKey string `json:"expectedKey,omitempty"`
	// DB-specific
	DSN           string `json:"dsn,omitempty"`
	ExpectedTable string `json:"expectedTable,omitempty"`
	// Kafka-specific
	Brokers []string `json:"brokers,omitempty"`
}

// WaitRequest is the body for POST /api/v1/waitx/wait.
type WaitRequest struct {
	CheckType   string `json:"checkType" binding:"required"`
	Target      string `json:"target" binding:"required"`
	Timeout     string `json:"timeout,omitempty"`
	Interval    string `json:"interval,omitempty"`
	RetryPolicy string `json:"retryPolicy,omitempty"`
	InvertCheck bool   `json:"invertCheck,omitempty"`
}
