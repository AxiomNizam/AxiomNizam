package lineage

import (
	"time"
)

// LineageNode represents entity in data lineage
type LineageNode struct {
	ID             string                 `json:"id"`
	TenantID       string                 `json:"tenantId"`
	NodeType       NodeType               `json:"nodeType"` // Source, Process, Sink
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	ResourceType   string                 `json:"resourceType"` // Table, Dataset, Query, etc
	ResourceID     string                 `json:"resourceId"`
	System         string                 `json:"system"` // Database, DataWarehouse, API
	Schema         string                 `json:"schema"`
	Location       string                 `json:"location"` // Path, URL, table location
	Format         string                 `json:"format"`   // JSON, CSV, Parquet, etc
	Columns        []ColumnInfo           `json:"columns"`  // Column-level lineage
	RecordCount    int64                  `json:"recordCount"`
	SizeBytes      int64                  `json:"sizeBytes"`
	LastModified   time.Time              `json:"lastModified"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	Owner          string                 `json:"owner"`
	Classification DataClassification     `json:"classification"` // PII, Sensitive, etc
	Quality        DataQuality            `json:"quality"`
	Metadata       map[string]interface{} `json:"metadata"`
	Tags           []string               `json:"tags"`
	IsActive       bool                   `json:"isActive"`
	VersionNumber  int64                  `json:"versionNumber"` // If dataset versioning
}

// NodeType represents type of lineage node
type NodeType string

const (
	NodeTypeSource    NodeType = "SOURCE"
	NodeTypeProcess   NodeType = "PROCESS"
	NodeTypeSink      NodeType = "SINK"
	NodeTypeTransform NodeType = "TRANSFORM"
	NodeTypeCache     NodeType = "CACHE"
	NodeTypeView      NodeType = "VIEW"
)

// ColumnInfo tracks column-level lineage
type ColumnInfo struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"` // int, string, bool, etc
	Description    string   `json:"description"`
	Nullable       bool     `json:"nullable"`
	Precision      int      `json:"precision,omitempty"` // For numeric
	Scale          int      `json:"scale,omitempty"`
	Classification string   `json:"classification"`
	IsKey          bool     `json:"isKey"`
	IsPII          bool     `json:"isPII"`
	Tags           []string `json:"tags"`
	LineageID      string   `json:"lineageId,omitempty"` // Link to column lineage
}

// DataClassification represents data sensitivity
type DataClassification struct {
	Level         string   `json:"level"`      // "Public", "Internal", "Confidential", "Restricted"
	Categories    []string `json:"categories"` // "PII", "PHI", "Financial", etc
	PII           bool     `json:"pii"`
	Sensitive     bool     `json:"sensitive"`
	Encrypted     bool     `json:"encrypted"`
	Masked        bool     `json:"masked"`
	Policy        string   `json:"policy"`
	RetentionDays int      `json:"retentionDays"`
}

// DataQuality represents data quality metrics
type DataQuality struct {
	Completeness   float64        `json:"completeness"` // 0-1
	Accuracy       float64        `json:"accuracy"`
	Consistency    float64        `json:"consistency"`
	Validity       float64        `json:"validity"`
	Timeliness     float64        `json:"timeliness"`
	UniquenesScore float64        `json:"uniquenesScore"`
	LastCheckedAt  time.Time      `json:"lastCheckedAt"`
	Issues         []QualityIssue `json:"issues"`
	OverallScore   float64        `json:"overallScore"` // 0-100
}

// QualityIssue represents quality problem
type QualityIssue struct {
	Type        string    `json:"type"`     // "missing_values", "duplicates", "outliers", "format_error"
	Severity    string    `json:"severity"` // "low", "medium", "high"
	Count       int64     `json:"count"`
	Percentage  float64   `json:"percentage"`
	Description string    `json:"description"`
	DetectedAt  time.Time `json:"detectedAt"`
	Resolution  string    `json:"resolution,omitempty"`
}

// LineageEdge represents data flow between nodes
type LineageEdge struct {
	ID             string                 `json:"id"`
	TenantID       string                 `json:"tenantId"`
	SourceNodeID   string                 `json:"sourceNodeId"`
	TargetNodeID   string                 `json:"targetNodeId"`
	EdgeType       EdgeType               `json:"edgeType"`       // Direct, Derived, Aggregated
	TransformType  string                 `json:"transformType"`  // copy, aggregate, join, etc
	ColumnMappings []ColumnMapping        `json:"columnMappings"` // Column-level mappings
	Transformation string                 `json:"transformation"` // Transformation logic/SQL
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	ProcessID      string                 `json:"processId,omitempty"` // Link to process/job
	Frequency      string                 `json:"frequency,omitempty"` // How often data flows
	Latency        int                    `json:"latency,omitempty"`   // Seconds
	Metadata       map[string]interface{} `json:"metadata"`
	IsActive       bool                   `json:"isActive"`
}

// EdgeType represents type of relationship
type EdgeType string

const (
	EdgeTypeDirect     EdgeType = "DIRECT"     // 1:1 copy
	EdgeTypeDerived    EdgeType = "DERIVED"    // Derived/computed
	EdgeTypeAggregated EdgeType = "AGGREGATED" // Aggregated
	EdgeTypeJoined     EdgeType = "JOINED"     // Result of join
	EdgeTypeUnioned    EdgeType = "UNIONED"    // Result of union
)

// ColumnMapping maps source to target columns
type ColumnMapping struct {
	SourceColumn string   `json:"sourceColumn"`
	TargetColumn string   `json:"targetColumn"`
	Transform    string   `json:"transform,omitempty"`   // Transformation function
	Filters      []string `json:"filters,omitempty"`     // Applied filters
	Aggregation  string   `json:"aggregation,omitempty"` // COUNT, SUM, AVG, etc
}

// LineageProcess represents data transformation process
type LineageProcess struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenantId"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"` // Job, Query, Pipeline, ETL, Workflow
	Description     string                 `json:"description"`
	ProcessType     ProcessType            `json:"processType"`
	SourceNodes     []string               `json:"sourceNodes"` // Node IDs
	TargetNodes     []string               `json:"targetNodes"`
	Owner           string                 `json:"owner"`
	Status          string                 `json:"status"` // Active, Inactive, Archived
	Schedule        ScheduleInfo           `json:"schedule,omitempty"`
	LastRun         time.Time              `json:"lastRun,omitempty"`
	NextRun         time.Time              `json:"nextRun,omitempty"`
	AverageDuration int                    `json:"averageDuration"` // Seconds
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
	Metrics         ProcessMetrics         `json:"metrics"`
	Documentation   string                 `json:"documentation"`
	Code            string                 `json:"code,omitempty"` // SQL, Python, etc
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ProcessType represents process type
type ProcessType string

const (
	ProcessTypeJob      ProcessType = "JOB"
	ProcessTypeQuery    ProcessType = "QUERY"
	ProcessTypePipeline ProcessType = "PIPELINE"
	ProcessTypeETL      ProcessType = "ETL"
	ProcessTypeWorkflow ProcessType = "WORKFLOW"
	ProcessTypeScript   ProcessType = "SCRIPT"
)

// ScheduleInfo for scheduled processes
type ScheduleInfo struct {
	Frequency      string // "hourly", "daily", "weekly", "monthly"
	CronExpression string
	Timezone       string
	NextRun        time.Time
	LastRun        time.Time
}

// ProcessMetrics tracks process statistics
type ProcessMetrics struct {
	TotalRuns        int64
	SuccessfulRuns   int64
	FailedRuns       int64
	AverageDuration  int
	MaxDuration      int
	MinDuration      int
	RecordsProcessed int64
	RecordsFailed    int64
	SuccessRate      float64
	LastRunDuration  int
}

// LineageGraph represents complete data lineage
type LineageGraph struct {
	ID           string           `json:"id"`
	TenantID     string           `json:"tenantId"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Nodes        []LineageNode    `json:"nodes"`
	Edges        []LineageEdge    `json:"edges"`
	Processes    []LineageProcess `json:"processes"`
	RootNodes    []string         `json:"rootNodes"` // Source nodes
	LeafNodes    []string         `json:"leafNodes"` // Sink nodes
	Depth        int              `json:"depth"`     // Max path length
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
	LastAnalyzed time.Time        `json:"lastAnalyzed"`
}

// LineagePath represents path through lineage
type LineageePath struct {
	ID               string   `json:"id"`
	SourceNodeID     string   `json:"sourceNodeId"`
	TargetNodeID     string   `json:"targetNodeId"`
	Path             []string `json:"path"` // Node IDs in order
	Length           int      `json:"length"`
	Transformations  []string `json:"transformations"` // Transform types
	EstimatedLatency int      `json:"estimatedLatency"`
	IsActive         bool     `json:"isActive"`
}

// ImpactAnalysis analyzes impact of changes
type ImpactAnalysis struct {
	TenantID          string        `json:"tenantId"`
	SourceNodeID      string        `json:"sourceNodeId"`
	AffectedNodeCount int           `json:"affectedNodeCount"`
	AffectedNodes     []string      `json:"affectedNodes"`
	CriticalNodes     []string      `json:"criticalNodes"` // High impact
	Paths             []LineagePath `json:"paths"`
	DownstreamAssets  []string      `json:"downstreamAssets"` // All dependent assets
	EstimatedImpact   string        `json:"estimatedImpact"`  // "low", "medium", "high"
	Recommendations   []string      `json:"recommendations"`
}

// LineageQuery filters lineage data
type LineageQuery struct {
	TenantID         string
	NodeType         NodeType
	ResourceType     string
	Owner            string
	Classification   string
	HasQualityIssues *bool
	StartTime        time.Time
	EndTime          time.Time
	Tags             []string
	Limit            int
	Offset           int
}

// LineageStatistics aggregates lineage metrics
type LineageStatistics struct {
	TenantID        string
	TotalNodes      int64
	SourceNodes     int64
	ProcessNodes    int64
	SinkNodes       int64
	TotalEdges      int64
	TotalProcesses  int64
	AverageDepth    float64
	ComplexityScore float64 // 0-100
	DataFlowRate    float64 // Records per second
	TopDataSources  []string
	TopDataSinks    []string
	LastUpdated     time.Time
}

// ColumnLineage tracks column-level transformations
type ColumnLineage struct {
	ID             string             `json:"id"`
	TenantID       string             `json:"tenantId"`
	SourceColumn   string             `json:"sourceColumn"`  // In source format: "table.column"
	TargetColumn   string             `json:"targetColumn"`  // In target format
	TransformPath  []ColumnTransform  `json:"transformPath"` // Steps in transformation
	Classification DataClassification `json:"classification"`
	Complexity     int                `json:"complexity"` // How complex transform is
	LastModified   time.Time          `json:"lastModified"`
}

// ColumnTransform represents one transformation step
type ColumnTransform struct {
	Step       int                    `json:"step"`
	Type       string                 `json:"type"`     // cast, aggregate, filter, join, etc
	Function   string                 `json:"function"` // Specific function
	Input      []string               `json:"input"`    // Input columns
	Output     string                 `json:"output"`   // Output column
	Parameters map[string]interface{} `json:"parameters"`
}

// LineageNotification notifies of changes
type LineageNotification struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId"`
	Type         string    `json:"type"` // "node_added", "edge_added", "process_failed"
	NodeID       string    `json:"nodeId,omitempty"`
	EdgeID       string    `json:"edgeId,omitempty"`
	Severity     string    `json:"severity"` // "info", "warning", "critical"
	Message      string    `json:"message"`
	ImpactNodes  []string  `json:"impactNodes"`
	CreatedAt    time.Time `json:"createdAt"`
	Acknowledged bool      `json:"acknowledged"`
}
