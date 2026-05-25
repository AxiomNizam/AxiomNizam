package versioning

import (
	"time"
)

// ResourceVersion represents a version of a resource
type ResourceVersion struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenantId"`
	ResourceType    string                 `json:"resourceType"`
	ResourceID      string                 `json:"resourceId"`
	Version         int64                  `json:"version"`    // Incremental version number
	Generation      int64                  `json:"generation"` // Kubernetes-style generation
	Content         map[string]interface{} `json:"content"`    // Full resource snapshot
	Changes         []FieldChange          `json:"changes"`    // What changed from previous
	CreatedBy       string                 `json:"createdBy"`  // User ID
	CreatedAt       time.Time              `json:"createdAt"`
	Reason          string                 `json:"reason"` // Why changed
	Action          VersionAction          `json:"action"` // CREATE, UPDATE, DELETE
	SizeBytes       int64                  `json:"sizeBytes"`
	Hash            string                 `json:"hash"`       // Content hash for dedup
	Checksum        string                 `json:"checksum"`   // Integrity check
	Compressed      bool                   `json:"compressed"` // Whether stored compressed
	Archived        bool                   `json:"archived"`   // In cold storage
	Metadata        map[string]string      `json:"metadata"`
	Tags            []string               `json:"tags"`            // Release tags, milestones
	RetentionPolicy RetentionPolicy        `json:"retentionPolicy"` // When to delete
	IsLatest        bool                   `json:"isLatest"`
	CurrentVersion  bool                   `json:"currentVersion"`
}

// FieldChange represents change to single field
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
	Type     ChangeType  `json:"type"` // ADDED, REMOVED, MODIFIED
}

// ChangeType represents type of change
type ChangeType string

const (
	ChangeAdded    ChangeType = "ADDED"
	ChangeRemoved  ChangeType = "REMOVED"
	ChangeModified ChangeType = "MODIFIED"
)

// VersionAction represents operation type
type VersionAction string

const (
	ActionCreate  VersionAction = "CREATE"
	ActionUpdate  VersionAction = "UPDATE"
	ActionDelete  VersionAction = "DELETE"
	ActionRestore VersionAction = "RESTORE"
)

// VersionHistory aggregates version data
type VersionHistory struct {
	TenantID       string            `json:"tenantId"`
	ResourceType   string            `json:"resourceType"`
	ResourceID     string            `json:"resourceId"`
	CurrentVersion int64             `json:"currentVersion"`
	LatestVersion  int64             `json:"latestVersion"`
	TotalVersions  int64             `json:"totalVersions"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
	Versions       []ResourceVersion `json:"versions"`  // All versions
	Snapshots      []Snapshot        `json:"snapshots"` // Named snapshots
	Branches       []VersionBranch   `json:"branches"`  // Version branches
}

// Snapshot represents named point-in-time
type Snapshot struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId"`
	Name         string    `json:"name"`
	ResourceType string    `json:"resourceType"`
	ResourceID   string    `json:"resourceId"`
	Version      int64     `json:"version"`
	Description  string    `json:"description"`
	CreatedBy    string    `json:"createdBy"`
	CreatedAt    time.Time `json:"createdAt"`
	Tags         []string  `json:"tags"`
	Pinned       bool      `json:"pinned"` // Don't auto-clean
	ReadOnly     bool      `json:"readOnly"`
}

// VersionBranch represents alternative version line
type VersionBranch struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"` // Feature branch, release, etc
	ResourceType  string    `json:"resourceType"`
	ResourceID    string    `json:"resourceId"`
	BaseVersion   int64     `json:"baseVersion"` // Where it forked from
	HeadVersion   int64     `json:"headVersion"` // Current version on branch
	CreatedBy     string    `json:"createdBy"`
	CreatedAt     time.Time `json:"createdAt"`
	Merged        bool      `json:"merged"` // If merged back
	MergedVersion int64     `json:"mergedVersion"`
	Tags          []string  `json:"tags"`
}

// VersionDiff compares two versions
type VersionDiff struct {
	FromVersion   int64         `json:"fromVersion"`
	ToVersion     int64         `json:"toVersion"`
	FromTimestamp time.Time     `json:"fromTimestamp"`
	ToTimestamp   time.Time     `json:"toTimestamp"`
	Changes       []FieldChange `json:"changes"`
	ChangeCount   int           `json:"changeCount"`
	AddedCount    int           `json:"addedCount"`
	RemovedCount  int           `json:"removedCount"`
	ModifiedCount int           `json:"modifiedCount"`
	SizeDiff      int64         `json:"sizeDiff"` // Bytes difference
}

// RollbackRequest requests rollback to version
type RollbackRequest struct {
	TenantID      string `json:"tenantId"`
	ResourceType  string `json:"resourceType"`
	ResourceID    string `json:"resourceId"`
	TargetVersion int64  `json:"targetVersion"`
	Reason        string `json:"reason"`
}

// RollbackResult represents rollback outcome
type RollbackResult struct {
	Success     bool          `json:"success"`
	FromVersion int64         `json:"fromVersion"`
	ToVersion   int64         `json:"toVersion"`
	Changes     []FieldChange `json:"changes"`
	CreatedAt   time.Time     `json:"createdAt"`
}

// VersionQuery filters versions
type VersionQuery struct {
	TenantID     string
	ResourceType string
	ResourceID   string
	Version      int64
	StartVersion int64
	EndVersion   int64
	StartTime    time.Time
	EndTime      time.Time
	CreatedBy    string
	Action       VersionAction
	Tags         []string
	Limit        int
	Offset       int
	SortBy       string // "version", "timestamp"
	SortOrder    string // "asc", "desc"
}

// VersionStats tracks versioning metrics
type VersionStats struct {
	TenantID          string
	ResourceType      string
	ResourceID        string
	TotalVersions     int64
	OldestVersion     int64
	LatestVersion     int64
	AverageChangeSize float64
	TotalStorageBytes int64
	ArchivedVersions  int64
	DeletedVersions   int64
}

// VersioningConfig configures version behavior
type VersioningConfig struct {
	Enabled              bool
	MaxVersions          int
	RetentionDays        int
	ArchiveAfterDays     int
	DeleteAfterDays      int
	CompressionEnabled   bool
	DeduplicationEnabled bool
	AutoSnapshot         bool // Auto-snapshot on major changes
	SnapshotInterval     time.Duration
	BranchingEnabled     bool
	DiffEnabled          bool
	RollbackEnabled      bool
	AuditVersions        bool // Log all version operations
}
