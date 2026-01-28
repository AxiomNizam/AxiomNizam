package webhooks

import (
	"time"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID             string             `json:"id"`
	TenantID       string             `json:"tenantId"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	URL            string             `json:"url"`              // Target endpoint
	Secret         string             `json:"secret,omitempty"` // For HMAC validation
	Events         []WebhookEventType `json:"events"`           // Events to subscribe to
	Filters        WebhookFilter      `json:"filters"`          // Optional filtering
	Version        string             `json:"version"`          // Webhook API version
	Active         bool               `json:"active"`
	RetryPolicy    RetryPolicy        `json:"retryPolicy"`
	RateLimit      RateLimitConfig    `json:"rateLimit"`
	Timeout        int                `json:"timeout"`        // Seconds
	Headers        map[string]string  `json:"headers"`        // Custom headers
	Authentication WebhookAuth        `json:"authentication"` // Auth config
	SSL            SSLConfig          `json:"ssl"`            // TLS config
	CreatedBy      string             `json:"createdBy"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	LastTriggered  time.Time          `json:"lastTriggered"`
	SuccessCount   int64              `json:"successCount"`
	FailureCount   int64              `json:"failureCount"`
	Stats          WebhookStats       `json:"stats"`
	Metadata       map[string]string  `json:"metadata"`
	Tags           []string           `json:"tags"`
}

// WebhookEventType represents types of events
type WebhookEventType string

const (
	EventResourceCreated       WebhookEventType = "resource.created"
	EventResourceUpdated       WebhookEventType = "resource.updated"
	EventResourceDeleted       WebhookEventType = "resource.deleted"
	EventResourceStatusChanged WebhookEventType = "resource.status.changed"
	EventJobCompleted          WebhookEventType = "job.completed"
	EventJobFailed             WebhookEventType = "job.failed"
	EventQueryExecuted         WebhookEventType = "query.executed"
	EventPolicyViolation       WebhookEventType = "policy.violated"
	EventDataAnomalyDetected   WebhookEventType = "data.anomaly.detected"
	EventQuotaExceeded         WebhookEventType = "quota.exceeded"
	EventTenantCreated         WebhookEventType = "tenant.created"
	EventTenantDeleted         WebhookEventType = "tenant.deleted"
	EventUserAdded             WebhookEventType = "user.added"
	EventUserRemoved           WebhookEventType = "user.removed"
	EventAuditEvent            WebhookEventType = "audit.event"
	EventSystemAlert           WebhookEventType = "system.alert"
)

// WebhookFilter filters which events trigger webhook
type WebhookFilter struct {
	ResourceTypes []string          `json:"resourceTypes"` // Empty = all
	ResourceIDs   []string          `json:"resourceIds"`   // Empty = all
	Actions       []string          `json:"actions"`       // CREATE, UPDATE, DELETE
	Users         []string          `json:"users"`         // Empty = all users
	Tags          []string          `json:"tags"`          // Resource tags
	MatchAllTags  bool              `json:"matchAllTags"`
	TenantIDs     []string          `json:"tenantIds"`  // For cross-tenant webhooks
	Conditions    []FilterCondition `json:"conditions"` // Custom conditions
}

// FilterCondition represents custom filter logic
type FilterCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, regex
	Value    interface{} `json:"value"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries           int     `json:"maxRetries"`
	InitialDelay         int     `json:"initialDelay"` // Milliseconds
	MaxDelay             int     `json:"maxDelay"`
	BackoffFactor        float64 `json:"backoffFactor"`
	ExponentialBackoff   bool    `json:"exponentialBackoff"`
	TimeoutExceededRetry bool    `json:"timeoutExceededRetry"`
	StatusCodes          []int   `json:"statusCodes"` // Retry on these codes
}

// RateLimitConfig limits webhook delivery rate
type RateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond int
	BurstSize         int
	TimeWindow        string // "minute", "hour"
	MaxBurstRequests  int
}

// WebhookAuth configures authentication
type WebhookAuth struct {
	Type     string       `json:"type"` // "none", "basic", "bearer", "oauth2", "api_key"
	Username string       `json:"username,omitempty"`
	Password string       `json:"password,omitempty"`
	Token    string       `json:"token,omitempty"`
	ApiKey   string       `json:"apiKey,omitempty"`
	OAuth2   OAuth2Config `json:"oauth2,omitempty"`
}

// OAuth2Config for OAuth2 authentication
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	Scopes       []string
}

// SSLConfig for TLS/HTTPS
type SSLConfig struct {
	Enabled             bool
	VerifySSL           bool
	CertFile            string
	KeyFile             string
	CABundle            string
	SkipSSLVerification bool
	MinTLSVersion       string // "1.0", "1.1", "1.2", "1.3"
}

// WebhookStats tracks webhook statistics
type WebhookStats struct {
	TotalDeliveries int64     `json:"totalDeliveries"`
	SuccessfulCount int64     `json:"successfulCount"`
	FailedCount     int64     `json:"failedCount"`
	RetryCount      int64     `json:"retryCount"`
	AverageLatency  float64   `json:"averageLatency"` // Milliseconds
	P95Latency      float64   `json:"p95Latency"`
	P99Latency      float64   `json:"p99Latency"`
	LastError       string    `json:"lastError"`
	LastErrorTime   time.Time `json:"lastErrorTime"`
	SuccessRate     float64   `json:"successRate"` // 0-1
	LastHealthCheck time.Time `json:"lastHealthCheck"`
	HealthStatus    string    `json:"healthStatus"` // "healthy", "degraded", "unhealthy"
}

// WebhookEvent represents event to deliver
type WebhookEvent struct {
	ID               string                 `json:"id"`
	TenantID         string                 `json:"tenantId"`
	WebhookID        string                 `json:"webhookId"`
	EventType        WebhookEventType       `json:"eventType"`
	Timestamp        time.Time              `json:"timestamp"`
	ResourceType     string                 `json:"resourceType"`
	ResourceID       string                 `json:"resourceId"`
	Action           string                 `json:"action"`  // CREATE, UPDATE, DELETE
	Actor            string                 `json:"actor"`   // User ID
	Changes          map[string]interface{} `json:"changes"` // Old/new values
	Data             map[string]interface{} `json:"data"`    // Full resource
	Metadata         map[string]string      `json:"metadata"`
	DeliveryStatus   DeliveryStatus         `json:"deliveryStatus"`
	DeliveryAttempts int                    `json:"deliveryAttempts"`
	NextRetryTime    time.Time              `json:"nextRetryTime"`
	LastError        string                 `json:"lastError"`
	Request          *WebhookRequest        `json:"request,omitempty"`  // What was sent
	Response         *WebhookResponse       `json:"response,omitempty"` // What was received
}

// DeliveryStatus tracks delivery state
type DeliveryStatus string

const (
	DeliveryPending    DeliveryStatus = "PENDING"
	DeliveryInProgress DeliveryStatus = "IN_PROGRESS"
	DeliverySucceeded  DeliveryStatus = "SUCCEEDED"
	DeliveryFailed     DeliveryStatus = "FAILED"
	DeliveryRejected   DeliveryStatus = "REJECTED"
	DeliveryThrottled  DeliveryStatus = "THROTTLED"
	DeliveryTimeout    DeliveryStatus = "TIMEOUT"
)

// WebhookRequest captures the HTTP request sent
type WebhookRequest struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	ContentType string            `json:"contentType"`
	Signature   string            `json:"signature"` // HMAC signature if using secret
	SentAt      time.Time         `json:"sentAt"`
	BodySize    int64             `json:"bodySize"`
}

// WebhookResponse captures the HTTP response received
type WebhookResponse struct {
	StatusCode   int               `json:"statusCode"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	ContentType  string            `json:"contentType"`
	BodySize     int64             `json:"bodySize"`
	ReceivedAt   time.Time         `json:"receivedAt"`
	ResponseTime int64             `json:"responseTime"` // Milliseconds
}

// WebhookCreateRequest API request
type WebhookCreateRequest struct {
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	URL            string             `json:"url"`
	Events         []WebhookEventType `json:"events"`
	Secret         string             `json:"secret,omitempty"`
	Filters        WebhookFilter      `json:"filters,omitempty"`
	RetryPolicy    RetryPolicy        `json:"retryPolicy,omitempty"`
	Authentication WebhookAuth        `json:"authentication,omitempty"`
	Headers        map[string]string  `json:"headers,omitempty"`
}

// WebhookQuery filters webhooks
type WebhookQuery struct {
	TenantID     string
	Active       *bool
	EventType    WebhookEventType
	ResourceType string
	CreatedBy    string
	Tags         []string
	HealthStatus string
	Limit        int
	Offset       int
	SortBy       string
}

// WebhookDeliveryLog tracks delivery attempts
type WebhookDeliveryLog struct {
	ID        string           `json:"id"`
	WebhookID string           `json:"webhookId"`
	EventID   string           `json:"eventId"`
	TenantID  string           `json:"tenantId"`
	Status    DeliveryStatus   `json:"status"`
	Attempt   int              `json:"attempt"`
	Request   WebhookRequest   `json:"request"`
	Response  *WebhookResponse `json:"response,omitempty"`
	Error     string           `json:"error,omitempty"`
	Latency   int64            `json:"latency"` // Milliseconds
	StartTime time.Time        `json:"startTime"`
	EndTime   time.Time        `json:"endTime"`
	NextRetry time.Time        `json:"nextRetry,omitempty"`
}
