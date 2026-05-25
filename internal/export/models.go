package export

import (
	"time"
)

// ExportJob represents an export operation
type ExportJob struct {
	ID            string                 `json:"id"`
	TenantID      string                 `json:"tenantId"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Status        ExportStatus           `json:"status"`
	Format        ExportFormat           `json:"format"`          // CSV, JSON, Parquet, etc
	Source        ExportSource           `json:"source"`          // What to export
	Query         string                 `json:"query,omitempty"` // Custom query
	Filters       map[string]interface{} `json:"filters"`         // What to filter
	Columns       []string               `json:"columns"`         // Which columns/fields
	SortBy        []SortOrder            `json:"sortBy"`
	Pagination    PaginationConfig       `json:"pagination"`
	Compression   CompressionType        `json:"compression"` // gzip, brotli, etc
	Encryption    EncryptionConfig       `json:"encryption"`
	Destination   ExportDestination      `json:"destination"`        // Where to save
	Schedule      ScheduleConfig         `json:"schedule,omitempty"` // Recurring exports
	RecordCount   int64                  `json:"recordCount"`
	FileSize      int64                  `json:"fileSize"`
	ProcessedRows int64                  `json:"processedRows"`
	SkippedRows   int64                  `json:"skippedRows"`
	ErrorRows     int64                  `json:"errorRows"`
	Progress      float64                `json:"progress"`      // 0-100%
	EstimatedTime int64                  `json:"estimatedTime"` // Seconds remaining
	CreatedBy     string                 `json:"createdBy"`
	CreatedAt     time.Time              `json:"createdAt"`
	StartedAt     time.Time              `json:"startedAt"`
	CompletedAt   time.Time              `json:"completedAt"`
	FailureReason string                 `json:"failureReason,omitempty"`
	RetryCount    int                    `json:"retryCount"`
	Metadata      map[string]string      `json:"metadata"`
	Tags          []string               `json:"tags"`
	Notification  NotificationConfig     `json:"notification"` // Notify when done
}

// SortOrder for export
type SortOrder struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc", "desc"
}

// PaginationConfig for limiting export
type PaginationConfig struct {
	Limit  int64 `json:"limit"` // 0 = unlimited
	Offset int64 `json:"offset"`
}

// ExportCreateRequest API request
type ExportCreateRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Format       ExportFormat           `json:"format"`
	Source       ExportSource           `json:"source"`
	Query        string                 `json:"query,omitempty"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
	Columns      []string               `json:"columns,omitempty"`
	Compression  CompressionType        `json:"compression,omitempty"`
	Encryption   EncryptionConfig       `json:"encryption,omitempty"`
	Destination  ExportDestination      `json:"destination"`
	Schedule     ScheduleConfig         `json:"schedule,omitempty"`
	Notification NotificationConfig     `json:"notification,omitempty"`
}

// ExportResponse returns export details
type ExportResponse struct {
	ID            string       `json:"id"`
	Status        ExportStatus `json:"status"`
	Progress      float64      `json:"progress"`
	RecordCount   int64        `json:"recordCount"`
	FileSize      int64        `json:"fileSize"`
	ProcessedRows int64        `json:"processedRows"`
	SkippedRows   int64        `json:"skippedRows"`
	ErrorRows     int64        `json:"errorRows"`
	DownloadURL   string       `json:"downloadUrl,omitempty"`
	FileLocation  string       `json:"fileLocation"`
	CompletedAt   time.Time    `json:"completedAt"`
	FailureReason string       `json:"failureReason,omitempty"`
}

// ExportQuery filters exports
type ExportQuery struct {
	TenantID  string
	Status    ExportStatus
	Format    ExportFormat
	CreatedBy string
	Tags      []string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Offset    int
	SortBy    string // "createdAt", "status"
}

// ExportTemplate represents reusable export configuration
type ExportTemplate struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenantId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Format      ExportFormat           `json:"format"`
	Source      ExportSource           `json:"source"`
	Filters     map[string]interface{} `json:"filters"`
	Columns     []string               `json:"columns"`
	Compression CompressionType        `json:"compression"`
	Encryption  EncryptionConfig       `json:"encryption"`
	Destination ExportDestination      `json:"destination"`
	CreatedBy   string                 `json:"createdBy"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	UsageCount  int64                  `json:"usageCount"`
	IsPublic    bool                   `json:"isPublic"`
	Tags        []string               `json:"tags"`
}

// ExportHistory tracks export executions
type ExportHistory struct {
	ID           string       `json:"id"`
	ExportJobID  string       `json:"exportJobId"`
	TemplateID   string       `json:"templateId,omitempty"`
	ExecutionNum int          `json:"executionNum"`
	Status       ExportStatus `json:"status"`
	StartTime    time.Time    `json:"startTime"`
	EndTime      time.Time    `json:"endTime"`
	Duration     int64        `json:"duration"` // Seconds
	RecordCount  int64        `json:"recordCount"`
	FileSize     int64        `json:"fileSize"`
	ErrorMessage string       `json:"errorMessage,omitempty"`
}

// BulkExportRequest for exporting multiple items
type BulkExportRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Format      ExportFormat       `json:"format"`
	Items       []ExportSourceItem `json:"items"`
	Compression CompressionType    `json:"compression,omitempty"`
	Encryption  EncryptionConfig   `json:"encryption,omitempty"`
	Destination ExportDestination  `json:"destination"`
	CreateZip   bool               `json:"createZip"` // Combine in ZIP
}

// ExportSourceItem represents item in bulk export
type ExportSourceItem struct {
	Type       string                 `json:"type"`
	ResourceID string                 `json:"resourceId,omitempty"`
	Database   string                 `json:"database,omitempty"`
	Table      string                 `json:"table,omitempty"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

// ExportStatistics aggregates export metrics
type ExportStatistics struct {
	TenantID            string
	TotalExports        int64
	SuccessfulCount     int64
	FailedCount         int64
	CancelledCount      int64
	PartialCount        int64
	AverageDuration     float64 // Seconds
	AverageFileSize     int64
	TotalDataExported   int64 // Bytes
	MostUsedFormat      ExportFormat
	MostUsedDestination string
	TopUsers            map[string]int64
	Timestamp           time.Time
}
