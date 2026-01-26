package audit

import "time"

// AuditAction represents the type of action audited
type AuditAction string

const (
	ActionCreate       AuditAction = "CREATE"
	ActionRead         AuditAction = "READ"
	ActionUpdate       AuditAction = "UPDATE"
	ActionDelete       AuditAction = "DELETE"
	ActionPatch        AuditAction = "PATCH"
	ActionList         AuditAction = "LIST"
	ActionLogin        AuditAction = "LOGIN"
	ActionLogout       AuditAction = "LOGOUT"
	ActionAccessDenied AuditAction = "ACCESS_DENIED"
	ActionValidation   AuditAction = "VALIDATION"
	ActionExport       AuditAction = "EXPORT"
	ActionImport       AuditAction = "IMPORT"
	ActionPolicyChange AuditAction = "POLICY_CHANGE"
)

// AuditResult represents outcome of audited action
type AuditResult string

const (
	ResultSuccess AuditResult = "SUCCESS"
	ResultFailure AuditResult = "FAILURE"
	ResultDenied  AuditResult = "DENIED"
	ResultWarning AuditResult = "WARNING"
)

// AuditLog represents immutable audit record
type AuditLog struct {
	ID            string                 `json:"id" db:"id"`                        // Unique audit log ID
	Timestamp     time.Time              `json:"timestamp" db:"timestamp"`          // When action occurred
	TenantID      string                 `json:"tenantId" db:"tenant_id"`           // Multi-tenant isolation
	UserID        string                 `json:"userId" db:"user_id"`               // Who performed action
	Username      string                 `json:"username" db:"username"`            // Username for easy filtering
	Action        AuditAction            `json:"action" db:"action"`                // What action (CRUD)
	Result        AuditResult            `json:"result" db:"result"`                // Success/Failure/Denied
	ResourceType  string                 `json:"resourceType" db:"resource_type"`   // What resource (User, APIResource, etc)
	ResourceID    string                 `json:"resourceId" db:"resource_id"`       // Specific resource ID
	ResourceName  string                 `json:"resourceName" db:"resource_name"`   // Resource name
	Namespace     string                 `json:"namespace" db:"namespace"`          // Resource namespace
	OldValues     map[string]interface{} `json:"oldValues" db:"-"`                  // Previous values (for updates)
	NewValues     map[string]interface{} `json:"newValues" db:"-"`                  // New values
	Changes       []Change               `json:"changes" db:"-"`                    // Detailed field changes
	SourceIP      string                 `json:"sourceIp" db:"source_ip"`           // Request source IP
	UserAgent     string                 `json:"userAgent" db:"user_agent"`         // User agent
	Method        string                 `json:"method" db:"method"`                // HTTP method
	Path          string                 `json:"path" db:"path"`                    // Request path
	RequestID     string                 `json:"requestId" db:"request_id"`         // Correlation ID
	StatusCode    int                    `json:"statusCode" db:"status_code"`       // HTTP status
	ErrorMessage  string                 `json:"errorMessage" db:"error_message"`   // Error details if failed
	Duration      int64                  `json:"duration" db:"duration"`            // Milliseconds taken
	Labels        map[string]string      `json:"labels" db:"-"`                     // Custom labels for filtering
	Reason        string                 `json:"reason" db:"reason"`                // Human readable reason
	ImmutableHash string                 `json:"immutableHash" db:"immutable_hash"` // SHA256 for immutability verification
}

// Change represents a single field change
type Change struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
}

// AuditFilter allows filtering audit logs
type AuditFilter struct {
	TenantID     string
	UserID       string
	Username     string
	Action       AuditAction
	Result       AuditResult
	ResourceType string
	ResourceID   string
	ResourceName string
	Namespace    string
	StartTime    time.Time
	EndTime      time.Time
	SourceIP     string
	StatusCode   int
	Limit        int
	Offset       int
	Sort         string // "timestamp:desc" or "timestamp:asc"
}

// AuditReport represents audit statistics
type AuditReport struct {
	TotalRecords      int64            `json:"totalRecords"`
	DateRange         DateRange        `json:"dateRange"`
	ActionBreakdown   map[string]int64 `json:"actionBreakdown"`   // By action type
	ResultBreakdown   map[string]int64 `json:"resultBreakdown"`   // By result
	UserBreakdown     map[string]int64 `json:"userBreakdown"`     // By user
	ResourceBreakdown map[string]int64 `json:"resourceBreakdown"` // By resource type
	TopUsers          []UserActivity   `json:"topUsers"`
	FailureRate       float64          `json:"failureRate"` // Percentage
	AccessDenialCount int64            `json:"accessDenialCount"`
	HighRiskActions   []AuditLog       `json:"highRiskActions"` // DELETE, POLICY_CHANGE, etc
}

// DateRange for audit reports
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// UserActivity tracks user actions
type UserActivity struct {
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	ActionCount  int64  `json:"actionCount"`
	SuccessCount int64  `json:"successCount"`
	FailureCount int64  `json:"failureCount"`
}

// ComplianceSnapshot captures state at point in time
type ComplianceSnapshot struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	TenantID      string                 `json:"tenantId"`
	ResourceType  string                 `json:"resourceType"`
	ResourceID    string                 `json:"resourceId"`
	State         map[string]interface{} `json:"state"`
	Hash          string                 `json:"hash"`
	CreatedBy     string                 `json:"createdBy"`
	Reason        string                 `json:"reason"`
	RetentionDays int                    `json:"retentionDays"`
}

// AuditConfig configures audit behavior
type AuditConfig struct {
	Enabled              bool     `json:"enabled"`
	LogActions           []string `json:"logActions"`           // Actions to log
	IgnoreActions        []string `json:"ignoreActions"`        // Actions to skip
	LogRequestBody       bool     `json:"logRequestBody"`       // Include request body
	LogResponseBody      bool     `json:"logResponseBody"`      // Include response body
	SensitiveFields      []string `json:"sensitiveFields"`      // Fields to redact
	RetentionDays        int      `json:"retentionDays"`        // How long to keep logs
	StorageBackend       string   `json:"storageBackend"`       // "database", "elasticsearch", "s3"
	AsyncWrite           bool     `json:"asyncWrite"`           // Write logs asynchronously
	HighRiskActions      []string `json:"highRiskActions"`      // Actions needing special attention
	ComplianceMode       bool     `json:"complianceMode"`       // Immutable mode for compliance
	EncryptSensitiveData bool     `json:"encryptSensitiveData"` // Encrypt sensitive values
}
