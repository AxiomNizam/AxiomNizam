package bulk

import "time"

// BulkOpResponse is the API response for a bulk operation.
type BulkOpResponse struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId,omitempty"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	TotalItems   int64     `json:"totalItems"`
	SuccessCount int64     `json:"successCount"`
	FailureCount int64     `json:"failureCount"`
	SkippedCount int64     `json:"skippedCount"`
	Progress     float64   `json:"progress,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
}

// ProgressResponse is the API response for operation progress.
type ProgressResponse struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	Progress float64 `json:"progress"`
	Total    int64   `json:"total"`
	Success  int64   `json:"success"`
	Failed   int64   `json:"failed"`
	Skipped  int64   `json:"skipped"`
}

// BulkOpListResponse is the API response for listing operations.
type BulkOpListResponse struct {
	Operations []BulkOpResponse `json:"operations"`
	Count      int              `json:"count"`
}

// ResourceCreatedResponse is the accepted response for reconciler-authoritative creation.
type ResourceCreatedResponse struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// MessageResponse is a generic action acknowledgment.
type MessageResponse struct {
	Message string `json:"message"`
}
