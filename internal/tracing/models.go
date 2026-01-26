package tracing

import (
	"time"
)

// Trace represents a complete request trace
type Trace struct {
	ID               string                 `json:"id"` // Trace ID (usually UUID)
	TenantID         string                 `json:"tenantId"`
	Spans            []Span                 `json:"spans"` // All spans in trace
	StartTime        time.Time              `json:"startTime"`
	EndTime          time.Time              `json:"endTime"`
	Duration         int64                  `json:"duration"` // Milliseconds
	TotalSpans       int                    `json:"totalSpans"`
	ErrorSpans       int                    `json:"errorSpans"`
	WarningSpans     int                    `json:"warningSpans"`
	Status           string                 `json:"status"` // "success", "error", "partial"
	ErrorMessage     string                 `json:"errorMessage,omitempty"`
	Services         []string               `json:"services"` // Services involved
	Root             Span                   `json:"root"`     // Root span
	Metadata         map[string]interface{} `json:"metadata"`
	SamplingDecision SamplingDecision       `json:"samplingDecision"`
	Tags             map[string]string      `json:"tags"`
}

// Span represents operation within trace
type Span struct {
	ID                 string                 `json:"id"` // Span ID
	TraceID            string                 `json:"traceId"`
	ParentSpanID       string                 `json:"parentSpanId,omitempty"`
	TenantID           string                 `json:"tenantId"`
	OperationName      string                 `json:"operationName"`
	Service            string                 `json:"service"` // Which service
	StartTime          time.Time              `json:"startTime"`
	EndTime            time.Time              `json:"endTime"`
	Duration           int64                  `json:"duration"` // Microseconds
	Status             SpanStatus             `json:"status"`
	Error              bool                   `json:"error"`
	ErrorMessage       string                 `json:"errorMessage,omitempty"`
	ErrorType          string                 `json:"errorType,omitempty"`
	StackTrace         string                 `json:"stackTrace,omitempty"`
	Tags               map[string]interface{} `json:"tags"`
	Logs               []LogEntry             `json:"logs"`
	Metrics            map[string]float64     `json:"metrics"` // e.g. db_queries, http_requests
	Attributes         map[string]interface{} `json:"attributes"`
	Kind               SpanKind               `json:"kind"`
	Links              []SpanLink             `json:"links,omitempty"`
	Events             []SpanEvent            `json:"events,omitempty"`
	Resource           ResourceInfo           `json:"resource"`
	InstrumentationLib string                 `json:"instrumentationLib"`
}

// SpanStatus represents span state
type SpanStatus string

const (
	SpanStatusUnset SpanStatus = "UNSET"
	SpanStatusOk    SpanStatus = "OK"
	SpanStatusError SpanStatus = "ERROR"
)

// SpanKind indicates span type
type SpanKind string

const (
	SpanKindInternal SpanKind = "INTERNAL"
	SpanKindServer   SpanKind = "SERVER"
	SpanKindClient   SpanKind = "CLIENT"
	SpanKindProducer SpanKind = "PRODUCER"
	SpanKindConsumer SpanKind = "CONSUMER"
)

// LogEntry represents log in span
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"` // "info", "warn", "error", "debug"
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields"`
}

// SpanLink links to another span
type SpanLink struct {
	TraceID    string                 `json:"traceId"`
	SpanID     string                 `json:"spanId"`
	Type       string                 `json:"type"` // "child", "parent", "follows", "related"
	Attributes map[string]interface{} `json:"attributes"`
}

// SpanEvent represents event during span
type SpanEvent struct {
	Name       string                 `json:"name"`
	Timestamp  time.Time              `json:"timestamp"`
	Attributes map[string]interface{} `json:"attributes"`
}

// ResourceInfo describes span resource
type ResourceInfo struct {
	ServiceName    string            `json:"serviceName"`
	ServiceVersion string            `json:"serviceVersion"`
	Hostname       string            `json:"hostname"`
	PodName        string            `json:"podName,omitempty"`
	Namespace      string            `json:"namespace,omitempty"`
	ClusterName    string            `json:"clusterName,omitempty"`
	Attributes     map[string]string `json:"attributes"`
}

// SamplingDecision indicates if trace was sampled
type SamplingDecision string

const (
	SamplingNotRecorded        SamplingDecision = "NOT_RECORDED"
	SamplingRecorded           SamplingDecision = "RECORDED"
	SamplingRecordedAndSampled SamplingDecision = "RECORDED_AND_SAMPLED"
)

// TracingConfig configures tracing behavior
type TracingConfig struct {
	Enabled            bool
	ExporterType       string // "jaeger", "zipkin", "datadog", "prometheus", "otlp"
	Endpoint           string
	ApiKey             string
	SamplingRate       float64 // 0.0-1.0
	SamplingStrategy   string  // "always_on", "always_off", "probability", "rate_limiting"
	MaxTraceSize       int
	MaxSpansPerTrace   int
	BatchSize          int
	BatchTimeout       int // Milliseconds
	ServiceName        string
	Environment        string
	Version            string
	Headers            map[string]string
	TLS                TLSConfig
	ProxyURL           string
	InsecureSkipVerify bool
}

// TLSConfig for tracing endpoint
type TLSConfig struct {
	Enabled            bool
	InsecureSkipVerify bool
	CertFile           string
	KeyFile            string
	CAFile             string
}

// TraceQuery filters traces
type TraceQuery struct {
	TenantID         string
	Service          string
	Operation        string
	MinDuration      time.Duration
	MaxDuration      time.Duration
	StartTime        time.Time
	EndTime          time.Time
	HasError         *bool
	Tags             map[string]string
	Status           SpanStatus
	SamplingDecision SamplingDecision
	Limit            int
	Offset           int
	SortBy           string // "duration", "timestamp", "error_count"
}

// TraceMetrics aggregates trace statistics
type TraceMetrics struct {
	TenantID            string
	Service             string
	Operation           string
	TraceCount          int64
	ErrorTraceCount     int64
	WarningTraceCount   int64
	AverageDuration     float64 // Milliseconds
	P50Duration         float64
	P95Duration         float64
	P99Duration         float64
	MaxDuration         int64
	MinDuration         int64
	ThroughputPerSecond float64
	ErrorRate           float64 // 0-1
	AverageSpanCount    float64
	TopErrorTypes       map[string]int
	Timestamp           time.Time
}

// SpanMetrics for individual span statistics
type SpanMetrics struct {
	TenantID        string
	Service         string
	Operation       string
	SpanCount       int64
	ErrorSpanCount  int64
	AverageDuration int64 // Microseconds
	P50Duration     int64
	P95Duration     int64
	P99Duration     int64
	MaxDuration     int64
	MinDuration     int64
}

// TraceContext carries trace information across process boundaries
type TraceContext struct {
	TraceID    string            `json:"traceId"`
	SpanID     string            `json:"spanId"`
	TraceFlags string            `json:"traceFlags"` // W3C trace flags
	TraceState string            `json:"traceState"` // W3C trace state
	Baggage    map[string]string `json:"baggage"`    // Cross-cutting concerns
}

// ContextPropagation for trace context propagation
type ContextPropagation struct {
	Format  string // "w3c", "jaeger", "zipkin", "otel"
	Headers map[string]string
}

// TraceSearchRequest for searching traces
type TraceSearchRequest struct {
	Query       string            `json:"query"` // Free text search
	Service     string            `json:"service"`
	Operation   string            `json:"operation"`
	Tags        map[string]string `json:"tags"`
	MinDuration int64             `json:"minDuration"`
	MaxDuration int64             `json:"maxDuration"`
	StartTime   time.Time         `json:"startTime"`
	EndTime     time.Time         `json:"endTime"`
	Limit       int               `json:"limit"`
}

// TraceSearchResult for search results
type TraceSearchResult struct {
	TraceID    string    `json:"traceId"`
	Service    string    `json:"service"`
	Operation  string    `json:"operation"`
	StartTime  time.Time `json:"startTime"`
	Duration   int64     `json:"duration"`
	SpanCount  int       `json:"spanCount"`
	ErrorCount int       `json:"errorCount"`
	Status     string    `json:"status"`
}

// DependencyMetrics for service dependencies
type DependencyMetrics struct {
	Source         string    `json:"source"`
	Destination    string    `json:"destination"`
	CallCount      int64     `json:"callCount"`
	ErrorCount     int64     `json:"errorCount"`
	ErrorRate      float64   `json:"errorRate"`
	AverageLatency float64   `json:"averageLatency"`
	P95Latency     float64   `json:"p95Latency"`
	P99Latency     float64   `json:"p99Latency"`
	MaxLatency     int64     `json:"maxLatency"`
	MinLatency     int64     `json:"minLatency"`
	Throughput     float64   `json:"throughput"` // Per second
	LastSeen       time.Time `json:"lastSeen"`
	Transitive     bool      `json:"transitive"` // Indirect dependency
}

// ServiceMap represents service dependencies
type ServiceMap struct {
	TenantID     string              `json:"tenantId"`
	Timestamp    time.Time           `json:"timestamp"`
	Services     []ServiceInfo       `json:"services"`
	Dependencies []DependencyMetrics `json:"dependencies"`
}

// ServiceInfo describes service in map
type ServiceInfo struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"` // "http", "database", "cache", "queue"
	RequestRate   float64           `json:"requestRate"`
	ErrorRate     float64           `json:"errorRate"`
	LatencyP99    float64           `json:"latencyP99"`
	LastSeen      time.Time         `json:"lastSeen"`
	Status        string            `json:"status"` // "healthy", "degraded", "unhealthy"
	InstanceCount int               `json:"instanceCount"`
	Attributes    map[string]string `json:"attributes"`
}

// LatencySummary for percentile-based latency analysis
type LatencySummary struct {
	Min    int64   `json:"min"`
	Max    int64   `json:"max"`
	Mean   float64 `json:"mean"`
	P50    int64   `json:"p50"`
	P75    int64   `json:"p75"`
	P90    int64   `json:"p90"`
	P95    int64   `json:"p95"`
	P99    int64   `json:"p99"`
	P999   int64   `json:"p999"`
	StdDev float64 `json:"stdDev"`
}

// ErrorAnalysis for error trends
type ErrorAnalysis struct {
	TenantID           string    `json:"tenantId"`
	ErrorType          string    `json:"errorType"`
	Count              int64     `json:"count"`
	FirstOccurrence    time.Time `json:"firstOccurrence"`
	LastOccurrence     time.Time `json:"lastOccurrence"`
	AffectedServices   []string  `json:"affectedServices"`
	AffectedOperations []string  `json:"affectedOperations"`
	TopTraces          []string  `json:"topTraces"` // Trace IDs with this error
	RootCausePatterns  []string  `json:"rootCausePatterns"`
	Resolution         string    `json:"resolution,omitempty"`
}
