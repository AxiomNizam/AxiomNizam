package streaming

import (
	"time"
)

// StreamMessage represents a message in streaming protocol
type StreamMessage struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"` // "query_result", "error", "progress", "close"
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`     // Actual payload
	Error     *StreamError           `json:"error"`    // If type=error
	Progress  *ProgressUpdate        `json:"progress"` // If type=progress
	Sequence  int64                  `json:"sequence"` // Message ordering
	TenantID  string                 `json:"tenantId"`
}

// MessageType for streaming
type MessageType string

const (
	MessageTypeQueryResult  MessageType = "QUERY_RESULT"
	MessageTypeError        MessageType = "ERROR"
	MessageTypeProgress     MessageType = "PROGRESS"
	MessageTypeClose        MessageType = "CLOSE"
	MessageTypeKeepAlive    MessageType = "KEEP_ALIVE"
	MessageTypeJobUpdate    MessageType = "JOB_UPDATE"
	MessageTypeNotification MessageType = "NOTIFICATION"
	MessageTypeMetrics      MessageType = "METRICS"
)

// StreamError represents error in stream
type StreamError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details"`
	Retryable bool   `json:"retryable"`
}

// ProgressUpdate represents progress update
type ProgressUpdate struct {
	Total       int64      `json:"total"`       // Total items
	Processed   int64      `json:"processed"`   // Items processed
	Percentage  int        `json:"percentage"`  // 0-100
	Rate        float64    `json:"rate"`        // Items/sec
	ETA         *time.Time `json:"eta"`         // Estimated completion
	CurrentStep string     `json:"currentStep"` // Current operation
}

// StreamRequest initiates a stream
type StreamRequest struct {
	RequestID string                 `json:"requestId"` // Unique stream ID
	Query     string                 `json:"query"`     // SQL query or GraphQL
	Format    OutputFormat           `json:"format"`    // JSON, CSV, Parquet
	TenantID  string                 `json:"tenantId"`
	Params    map[string]interface{} `json:"params"`    // Query parameters
	Filters   map[string]interface{} `json:"filters"`   // Additional filters
	ChunkSize int                    `json:"chunkSize"` // Rows per message
	Timeout   int                    `json:"timeout"`   // Seconds
	Track     bool                   `json:"track"`     // Enable progress tracking
}

// StreamSession manages a streaming session
type StreamSession struct {
	ID           string
	TenantID     string
	UserID       string
	CreatedAt    time.Time
	LastActivity time.Time
	RequestID    string
	Query        string
	Format       OutputFormat
	ChunkSize    int
	Timeout      time.Duration
	TotalRows    int64
	SentRows     int64
	Active       bool
	MessageCount int64
}

// StreamSubscription for real-time subscriptions
type StreamSubscription struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenantId"`
	UserID       string                 `json:"userId"`
	Topic        string                 `json:"topic"`      // Resource topic to subscribe
	EventTypes   []string               `json:"eventTypes"` // CREATE, UPDATE, DELETE, etc
	Filter       map[string]interface{} `json:"filter"`     // Match criteria
	CreatedAt    time.Time              `json:"createdAt"`
	Active       bool                   `json:"active"`
	LastActivity time.Time              `json:"lastActivity"`
	DeliveryMode DeliveryMode           `json:"deliveryMode"` // At-most-once, At-least-once
}

// RealtimeEvent represents event for subscriptions
type RealtimeEvent struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenantId"`
	Topic        string                 `json:"topic"`
	EventType    string                 `json:"eventType"` // CREATE, UPDATE, DELETE
	ResourceType string                 `json:"resourceType"`
	ResourceID   string                 `json:"resourceId"`
	Object       map[string]interface{} `json:"object"`    // The resource
	OldObject    map[string]interface{} `json:"oldObject"` // Previous state (for UPDATE)
	Timestamp    time.Time              `json:"timestamp"`
	Source       string                 `json:"source"` // Service that generated event
}

// StreamConfig configures streaming behavior
type StreamConfig struct {
	Enabled              bool
	MaxConcurrentStreams int
	MaxChunkSize         int           // Maximum rows per chunk
	DefaultChunkSize     int           // Default rows per chunk
	SessionTimeout       time.Duration // How long to keep session open
	HeartbeatInterval    time.Duration // Keep-alive ping interval
	MaxRetries           int
	CompressionEnabled   bool
	BufferSize           int // Message buffer size
}
