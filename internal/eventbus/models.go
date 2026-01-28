package eventbus

import (
	"time"
)

// EventBusEvent represents an event in the system
type EventBusEvent struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenantId"`
	Type            string                 `json:"type"`    // Topic/type of event
	Source          string                 `json:"source"`  // Origin service
	Version         string                 `json:"version"` // Schema version
	Timestamp       time.Time              `json:"timestamp"`
	Subject         string                 `json:"subject"`         // What it's about
	DataContentType string                 `json:"dataContentType"` // "application/json"
	Data            map[string]interface{} `json:"data"`            // Event payload
	Metadata        map[string]string      `json:"metadata"`
	CorrelationID   string                 `json:"correlationId"` // For tracing
	CausationID     string                 `json:"causationId"`   // What caused this
	AggregateID     string                 `json:"aggregateId"`   // Root entity ID
	AggregateType   string                 `json:"aggregateType"` // Resource type
	EventSequence   int64                  `json:"eventSequence"` // Order in stream
	IsProcessed     bool                   `json:"isProcessed"`
	ProcessedAt     time.Time              `json:"processedAt,omitempty"`
	RetryCount      int                    `json:"retryCount"`
	DeadLettered    bool                   `json:"deadLettered"`
	Priority        int                    `json:"priority"` // 1-10, higher = urgent
	TTL             time.Duration          `json:"ttl"`      // Time to live
	ExpiresAt       time.Time              `json:"expiresAt"`
	Headers         map[string]string      `json:"headers"`
}

// EventTopic represents an event channel
type EventTopic struct {
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Schema            EventSchema     `json:"schema"`            // Expected event format
	Partitions        int             `json:"partitions"`        // For parallelism
	ReplicationFactor int             `json:"replicationFactor"` // For HA
	Retention         RetentionConfig `json:"retention"`
	Config            TopicConfig     `json:"config"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	MessageCount      int64           `json:"messageCount"`
	IsActive          bool            `json:"isActive"`
}

// EventSchema validates event structure
type EventSchema struct {
	Version  string                   `json:"version"`
	Fields   map[string]FieldSchema   `json:"fields"`
	Required []string                 `json:"required"`
	Examples []map[string]interface{} `json:"examples"`
}

// FieldSchema describes event field
type FieldSchema struct {
	Type        string        `json:"type"` // string, int, bool, object, array
	Description string        `json:"description"`
	Required    bool          `json:"required"`
	Pattern     string        `json:"pattern,omitempty"` // Regex validation
	Enum        []interface{} `json:"enum,omitempty"`    // Allowed values
	Min         interface{}   `json:"min,omitempty"`
	Max         interface{}   `json:"max,omitempty"`
}

// RetentionConfig defines data retention
type RetentionConfig struct {
	Type            string `json:"type"` // "time", "size", "both"
	TimeMs          int64  `json:"timeMs"`
	SizeBytes       int64  `json:"sizeBytes"`
	DeletePolicy    string `json:"deletePolicy"` // "delete", "compact"
	CompactDeleteMs int64  `json:"compactDeleteMs"`
}

// TopicConfig configures topic behavior
type TopicConfig struct {
	CompressionType   string `json:"compressionType"` // "none", "gzip", "snappy", "lz4"
	MinInSyncReplicas int    `json:"minInSyncReplicas"`
	MaxMessageBytes   int    `json:"maxMessageBytes"`
}

// EventSubscription represents consumer subscription
type EventSubscription struct {
	ID             string             `json:"id"`
	TenantID       string             `json:"tenantId"`
	Name           string             `json:"name"`
	Topics         []string           `json:"topics"`        // Topics subscribed to
	ConsumerGroup  string             `json:"consumerGroup"` // For grouping
	Handler        string             `json:"handler"`       // Handler function/service
	Filter         EventFilter        `json:"filter"`        // Optional filtering
	Config         SubscriptionConfig `json:"config"`
	Status         string             `json:"status"` // "active", "paused", "stopped"
	Offset         int64              `json:"offset"` // Current position
	Lag            int64              `json:"lag"`    // Messages behind
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	LastProcessed  time.Time          `json:"lastProcessed"`
	ProcessedCount int64              `json:"processedCount"`
	FailedCount    int64              `json:"failedCount"`
	Metadata       map[string]string  `json:"metadata"`
}

// EventFilter filters which events trigger handler
type EventFilter struct {
	Types          []string          `json:"types"`          // Event types
	Sources        []string          `json:"sources"`        // From these services
	Subjects       []string          `json:"subjects"`       // About these subjects
	AggregateTypes []string          `json:"aggregateTypes"` // Resource types
	Conditions     []FilterCondition `json:"conditions"`     // Custom logic
	MinPriority    int               `json:"minPriority"`    // Only important events
}

// FilterCondition represents filter condition
type FilterCondition struct {
	Path     string      `json:"path"`     // JSON path in data
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, exists, matches
	Value    interface{} `json:"value"`
}

// SubscriptionConfig configures subscription behavior
type SubscriptionConfig struct {
	ProcessingMode  string      `json:"processingMode"` // "auto", "manual"
	DeliveryMode    string      `json:"deliveryMode"`   // "at_most_once", "at_least_once", "exactly_once"
	Timeout         int         `json:"timeout"`        // Seconds
	MaxConcurrency  int         `json:"maxConcurrency"` // Parallel handlers
	RetryPolicy     RetryPolicy `json:"retryPolicy"`
	DeadLetterTopic string      `json:"deadLetterTopic"` // Failed events go here
	DLQ             bool        `json:"dlq"`             // Enable DLQ
	AckTimeout      int         `json:"ackTimeout"`      // Seconds before redelivery
	OrderGuarantee  string      `json:"orderGuarantee"`  // "none", "key", "partition"
	StartFromOffset int64       `json:"startFromOffset"`
	StartFromTime   time.Time   `json:"startFromTime"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries    int     `json:"maxRetries"`
	InitialDelay  int     `json:"initialDelay"` // Milliseconds
	MaxDelay      int     `json:"maxDelay"`
	BackoffFactor float64 `json:"backoffFactor"`
}

// EventPublishRequest API request
type EventPublishRequest struct {
	Type          string                 `json:"type"`
	Subject       string                 `json:"subject,omitempty"`
	Data          map[string]interface{} `json:"data"`
	Source        string                 `json:"source,omitempty"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
	Priority      int                    `json:"priority,omitempty"`
	CorrelationID string                 `json:"correlationId,omitempty"`
	CausationID   string                 `json:"causationId,omitempty"`
}

// EventPublishResponse returns created event
type EventPublishResponse struct {
	EventID   string    `json:"eventId"`
	Timestamp time.Time `json:"timestamp"`
	Topic     string    `json:"topic"`
	Partition int       `json:"partition"`
	Offset    int64     `json:"offset"`
}

// EventBusConfig configures the event bus
type EventBusConfig struct {
	Type              string         `json:"type"` // "kafka", "rabbitmq", "redis", "memory"
	Brokers           []string       `json:"brokers"`
	GroupID           string         `json:"groupId"`
	SchemaRegistry    string         `json:"schemaRegistry"`
	Compression       string         `json:"compression"`
	MaxMessageSize    int            `json:"maxMessageSize"`
	ConnectionTimeout int            `json:"connectionTimeout"`
	RequestTimeout    int            `json:"requestTimeout"`
	SecurityConfig    SecurityConfig `json:"securityConfig"`
}

// SecurityConfig for event bus connection
type SecurityConfig struct {
	Type           string     `json:"type"` // "plaintext", "ssl", "sasl_ssl"
	SASL           SASLConfig `json:"sasl,omitempty"`
	CACertPath     string     `json:"caCertPath,omitempty"`
	ClientCertPath string     `json:"clientCertPath,omitempty"`
	ClientKeyPath  string     `json:"clientKeyPath,omitempty"`
}

// SASLConfig for SASL authentication
type SASLConfig struct {
	Mechanism string `json:"mechanism"` // "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"
	Username  string `json:"username"`
	Password  string `json:"password"`
}

// EventQuery filters events
type EventQuery struct {
	TenantID      string
	Type          string
	Source        string
	AggregateID   string
	AggregateType string
	CorrelationID string
	StartTime     time.Time
	EndTime       time.Time
	IsProcessed   *bool
	DeadLettered  *bool
	MinPriority   int
	Limit         int
	Offset        int
	SortBy        string // "timestamp"
}

// EventStreamConsumer represents active consumer
type EventStreamConsumer struct {
	ID                string    `json:"id"`
	SubscriptionID    string    `json:"subscriptionId"`
	TenantID          string    `json:"tenantId"`
	ConsumerGroup     string    `json:"consumerGroup"`
	CurrentOffset     int64     `json:"currentOffset"`
	CommittedOffset   int64     `json:"committedOffset"`
	Lag               int64     `json:"lag"`
	BytesConsumed     int64     `json:"bytesConsumed"`
	MessagesConsumed  int64     `json:"messagesConsumed"`
	LastProcessedTime time.Time `json:"lastProcessedTime"`
	StartTime         time.Time `json:"startTime"`
	Status            string    `json:"status"` // "active", "paused", "stopped"
	Host              string    `json:"host"`   // Hostname
	SessionID         string    `json:"sessionId"`
}

// EventBusMetrics tracks event bus statistics
type EventBusMetrics struct {
	TopicsCount         int64
	PartitionsCount     int64
	ConsumerGroupsCount int64
	TotalEvents         int64
	EventsPerSecond     float64
	AverageLatency      float64 // Milliseconds
	P95Latency          float64
	P99Latency          float64
	ConsumedCount       int64
	FailedCount         int64
	DeadLetteredCount   int64
	RetriedCount        int64
	StorageSize         int64
	ReplicationLag      int64
	UnhealthyPartitions int
	Timestamp           time.Time
}

// DLQEvent represents dead-lettered event
type DLQEvent struct {
	ID               string        `json:"id"`
	OriginalEventID  string        `json:"originalEventId"`
	TenantID         string        `json:"tenantId"`
	Topic            string        `json:"topic"`
	SubscriptionID   string        `json:"subscriptionId"`
	Error            string        `json:"error"`
	FailureCount     int           `json:"failureCount"`
	LastFailureTime  time.Time     `json:"lastFailureTime"`
	Event            EventBusEvent `json:"event"`
	ManuallyResolved bool          `json:"manuallyResolved"`
	ResolutionAction string        `json:"resolutionAction"` // "discard", "retry", "manual"
	ResolutionTime   time.Time     `json:"resolutionTime,omitempty"`
	ReplayToTopic    string        `json:"replayToTopic,omitempty"`
}
