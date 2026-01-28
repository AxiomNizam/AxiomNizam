package bulk

import (
	"time"
)

// BulkOperation represents batch operation
type BulkOperation struct {
	ID              string               `json:"id"`
	TenantID        string               `json:"tenantId"`
	UserID          string               `json:"userId"`
	Type            BulkOpType           `json:"type"`         // CREATE, UPDATE, DELETE, PATCH
	ResourceType    string               `json:"resourceType"` // User, APIResource, etc
	Status          BulkOpStatus         `json:"status"`       // Pending, Running, Completed, Failed
	Progress        int                  `json:"progress"`     // 0-100
	TotalItems      int64                `json:"totalItems"`
	SuccessCount    int64                `json:"successCount"`
	FailureCount    int64                `json:"failureCount"`
	SkippedCount    int64                `json:"skippedCount"`
	Items           []BulkItem           `json:"items"` // The actual operations
	StartedAt       *time.Time           `json:"startedAt"`
	CompletedAt     *time.Time           `json:"completedAt"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
	ErrorSummary    *BulkErrorSummary    `json:"errorSummary"`
	Options         BulkOperationOptions `json:"options"`
	Timeout         int                  `json:"timeout"`         // Seconds
	Atomic          bool                 `json:"atomic"`          // All or nothing
	RollbackOnError bool                 `json:"rollbackOnError"` // Rollback if any fails
}

// BulkOpType represents operation type
type BulkOpType string

const (
	BulkOpCreate  BulkOpType = "CREATE"
	BulkOpUpdate  BulkOpType = "UPDATE"
	BulkOpDelete  BulkOpType = "DELETE"
	BulkOpPatch   BulkOpType = "PATCH"
	BulkOpReplace BulkOpType = "REPLACE"
	BulkOpUpsert  BulkOpType = "UPSERT"
)

// BulkOpStatus represents status
type BulkOpStatus string

const (
	BulkOpPending   BulkOpStatus = "PENDING"
	BulkOpRunning   BulkOpStatus = "RUNNING"
	BulkOpCompleted BulkOpStatus = "COMPLETED"
	BulkOpFailed    BulkOpStatus = "FAILED"
	BulkOpCancelled BulkOpStatus = "CANCELLED"
	BulkOpPartial   BulkOpStatus = "PARTIAL" // Some succeeded, some failed
)

// BulkItem represents single operation in bulk
type BulkItem struct {
	ID           string                 `json:"id"`
	Index        int64                  `json:"index"`     // Position in batch
	Status       BulkItemStatus         `json:"status"`    // Pending, Success, Failed, Skipped
	Operation    string                 `json:"operation"` // The actual op (POST, PUT, DELETE, PATCH)
	ResourceType string                 `json:"resourceType"`
	ResourceID   string                 `json:"resourceId"` // ID for update/delete/patch
	Data         map[string]interface{} `json:"data"`       // Payload
	Result       *BulkItemResult        `json:"result"`
	Error        *BulkItemError         `json:"error"`
	Timestamp    time.Time              `json:"timestamp"`
}

// BulkItemStatus represents item status
type BulkItemStatus string

const (
	BulkItemPending BulkItemStatus = "PENDING"
	BulkItemSuccess BulkItemStatus = "SUCCESS"
	BulkItemFailed  BulkItemStatus = "FAILED"
	BulkItemSkipped BulkItemStatus = "SKIPPED"
	BulkItemRetry   BulkItemStatus = "RETRY"
)

// BulkItemResult successful result
type BulkItemResult struct {
	ResourceID string                 `json:"resourceId"`
	Created    bool                   `json:"created"`
	Modified   bool                   `json:"modified"`
	Data       map[string]interface{} `json:"data"`
	Version    string                 `json:"version"`
}

// BulkItemError represents error
type BulkItemError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details"`
	Retryable bool   `json:"retryable"`
}

// BulkErrorSummary aggregates errors
type BulkErrorSummary struct {
	TotalErrors  int64            `json:"totalErrors"`
	ByCode       map[string]int64 `json:"byCode"`
	ByMessage    map[string]int64 `json:"byMessage"`
	SampleErrors []BulkItemError  `json:"sampleErrors"` // First 5 errors
}

// BulkOperationOptions configures bulk operation
type BulkOperationOptions struct {
	ContinueOnError  bool        `json:"continueOnError"` // Keep processing on error
	Atomic           bool        `json:"atomic"`          // All or nothing
	BatchSize        int         `json:"batchSize"`       // Items per batch
	Parallelism      int         `json:"parallelism"`     // Concurrent operations
	RetryPolicy      RetryPolicy `json:"retryPolicy"`
	NotifyOnComplete bool        `json:"notifyOnComplete"`
	NotifyOnError    bool        `json:"notifyOnError"`
	ValidateOnly     bool        `json:"validateOnly"` // Dry run
	DryRun           bool        `json:"dryRun"`
	Timeout          int         `json:"timeout"` // Seconds
}

// RetryPolicy for failed items
type RetryPolicy struct {
	MaxRetries   int           `json:"maxRetries"`
	InitialDelay time.Duration `json:"initialDelay"`
	MaxDelay     time.Duration `json:"maxDelay"`
	Exponential  bool          `json:"exponential"`
}

// BulkOperationRequest submits bulk operation
type BulkOperationRequest struct {
	TenantID     string               `json:"tenantId"`
	Type         BulkOpType           `json:"type"`
	ResourceType string               `json:"resourceType"`
	Items        []BulkItem           `json:"items"`
	Options      BulkOperationOptions `json:"options"`
	Timeout      int                  `json:"timeout"`
}

// BulkOperationResponse for submission
type BulkOperationResponse struct {
	OperationID         string       `json:"operationId"`
	Status              BulkOpStatus `json:"status"`
	CreatedAt           time.Time    `json:"createdAt"`
	EstimatedCompletion *time.Time   `json:"estimatedCompletion"`
}

// BulkOperationProgress tracks progress
type BulkOperationProgress struct {
	OperationID  string       `json:"operationId"`
	Status       BulkOpStatus `json:"status"`
	Progress     int          `json:"progress"`
	TotalItems   int64        `json:"totalItems"`
	SuccessCount int64        `json:"successCount"`
	FailureCount int64        `json:"failureCount"`
	SkippedCount int64        `json:"skippedCount"`
	Rate         float64      `json:"rate"`        // Items/sec
	ElapsedTime  int64        `json:"elapsedTime"` // Milliseconds
	ETA          *time.Time   `json:"eta"`
}

// BulkImportFormat for importing data
type BulkImportFormat string

const (
	FormatJSON    BulkImportFormat = "JSON"
	FormatCSV     BulkImportFormat = "CSV"
	FormatParquet BulkImportFormat = "PARQUET"
	FormatXML     BulkImportFormat = "XML"
)

// BulkImportRequest for importing data
type BulkImportRequest struct {
	TenantID     string               `json:"tenantId"`
	ResourceType string               `json:"resourceType"`
	Format       BulkImportFormat     `json:"format"`
	Source       string               `json:"source"` // URL or file path
	Options      BulkOperationOptions `json:"options"`
}

// BulkExportRequest for exporting data
type BulkExportRequest struct {
	TenantID     string                 `json:"tenantId"`
	ResourceType string                 `json:"resourceType"`
	Format       BulkImportFormat       `json:"format"`
	Filters      map[string]interface{} `json:"filters"`
	Fields       []string               `json:"fields"`      // Column selection
	Destination  string                 `json:"destination"` // S3, GCS, etc
}
