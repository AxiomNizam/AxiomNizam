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

// ExportStatus represents export state
type ExportStatus string

const (
	ExportPending               ExportStatus = "PENDING"
	ExportValidating            ExportStatus = "VALIDATING"
	ExportQueued                ExportStatus = "QUEUED"
	ExportRunning               ExportStatus = "RUNNING"
	ExportProcessing            ExportStatus = "PROCESSING"
	ExportCompressing           ExportStatus = "COMPRESSING"
	ExportEncrypting            ExportStatus = "ENCRYPTING"
	ExportUploadingExportStatus              = "UPLOADING"
	ExportCompleted             ExportStatus = "COMPLETED"
	ExportFailed                ExportStatus = "FAILED"
	ExportCancelled             ExportStatus = "CANCELLED"
	ExportPartial               ExportStatus = "PARTIAL"
)

// ExportFormat defines output format
type ExportFormat string

const (
	FormatJSON     ExportFormat = "JSON"
	FormatCSV      ExportFormat = "CSV"
	FormatXML      ExportFormat = "XML"
	FormatParquet  ExportFormat = "PARQUET"
	FormatAvro     ExportFormat = "AVRO"
	FormatProtobuf ExportFormat = "PROTOBUF"
	FormatExcel    ExportFormat = "EXCEL"
	FormatPDF      ExportFormat = "PDF"
	FormatSQL      ExportFormat = "SQL"
	FormatNDJSON   ExportFormat = "NDJSON"
)

// ExportSource defines what to export
type ExportSource struct {
	Type           string `json:"type"` // "resource", "query", "database", "table"
	ResourceType   string `json:"resourceType,omitempty"`
	ResourceID     string `json:"resourceId,omitempty"`
	Database       string `json:"database,omitempty"`
	Table          string `json:"table,omitempty"`
	Query          string `json:"query,omitempty"`
	IncludeRelated bool   `json:"includeRelated"`
	IncludeAudit   bool   `json:"includeAudit"`
	IncludeHistory bool   `json:"includeHistory"`
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

// CompressionType for export compression
type CompressionType string

const (
	CompressionNone   CompressionType = "NONE"
	CompressionGzip   CompressionType = "GZIP"
	CompressionBrotli CompressionType = "BROTLI"
	CompressionLZ4    CompressionType = "LZ4"
	CompressionZstd   CompressionType = "ZSTD"
)

// EncryptionConfig for export encryption
type EncryptionConfig struct {
	Enabled     bool   `json:"enabled"`
	Algorithm   string `json:"algorithm"`   // "AES-256-GCM", "AES-256-CBC"
	KeyID       string `json:"keyId"`       // Reference to encryption key
	KeyProvider string `json:"keyProvider"` // "kms", "vault", "local"
	Passphrase  string `json:"passphrase"`  // Client-side encryption key
}

// ExportDestination where to store export
type ExportDestination struct {
	Type              string             `json:"type"`             // "local", "s3", "gcs", "azure", "ftp", "sftp"
	Path              string             `json:"path"`             // Local path or remote path
	Bucket            string             `json:"bucket,omitempty"` // S3/GCS bucket
	Region            string             `json:"region,omitempty"` // AWS region
	AccessKey         string             `json:"accessKey,omitempty"`
	SecretKey         string             `json:"secretKey,omitempty"`
	Endpoint          string             `json:"endpoint,omitempty"` // Custom endpoint
	StorageClass      string             `json:"storageClass"`       // S3 storage class
	ACL               string             `json:"acl"`                // S3 ACL
	MakePublic        bool               `json:"makePublic"`
	CreateBackup      bool               `json:"createBackup"` // Keep backup copy
	OverwriteExisting bool               `json:"overwriteExisting"`
	Notification      NotificationConfig `json:"notification"`
}

// ScheduleConfig for recurring exports
type ScheduleConfig struct {
	Enabled    bool   `json:"enabled"`
	Cron       string `json:"cron"` // Cron expression
	Timezone   string `json:"timezone"`
	Frequency  string `json:"frequency"`           // "hourly", "daily", "weekly", "monthly"
	Time       string `json:"time"`                // HH:MM
	DayOfWeek  string `json:"dayOfWeek,omitempty"` // "monday", "tuesday", etc
	DayOfMonth int    `json:"dayOfMonth,omitempty"`
	MaxRuns    int    `json:"maxRuns"`   // 0 = unlimited
	Retention  int    `json:"retention"` // Days to keep scheduled exports
}

// NotificationConfig for export notifications
type NotificationConfig struct {
	Enabled      bool
	OnSuccess    bool
	OnFailure    bool
	Webhooks     []string // Webhook URLs
	Email        []string // Email addresses
	SlackWebhook string   // Slack webhook
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
