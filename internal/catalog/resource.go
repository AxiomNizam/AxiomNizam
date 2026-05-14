package catalog

// =====================================================
// WS-1.1 — Data Catalog as declarative resources
//
// CatalogAssetResource represents a discoverable data asset (table,
// view, topic, bucket, API, pipeline) in the unified metadata registry.
// The reconciler periodically scans connected datasources and updates
// asset metadata (row counts, column stats, freshness).
//
// CatalogCollectionResource groups assets into logical collections
// for organizational purposes (domains, projects, teams).
// =====================================================

import (
	"time"

	"example.com/axiomnizam/internal/resources"
)

// --- Constants ---

const (
	CatalogAssetKind       = "CatalogAsset"
	CatalogAssetAPIVersion = "catalog.axiomnizam.io/v1"

	CatalogCollectionKind       = "CatalogCollection"
	CatalogCollectionAPIVersion = "catalog.axiomnizam.io/v1"
)

// --- Asset Types ---

type AssetType string

const (
	AssetTypeTable    AssetType = "table"
	AssetTypeView     AssetType = "view"
	AssetTypeTopic    AssetType = "topic"
	AssetTypeBucket   AssetType = "bucket"
	AssetTypeAPI      AssetType = "api"
	AssetTypePipeline AssetType = "pipeline"
	AssetTypeModel    AssetType = "model"
	AssetTypeDataset  AssetType = "dataset"
)

// --- Refresh Policy ---

type RefreshPolicy struct {
	// Interval between automatic metadata refreshes (e.g. "1h", "6h", "24h")
	Interval string `json:"interval,omitempty"`

	// Schedule is a cron expression for refresh timing
	Schedule string `json:"schedule,omitempty"`

	// Enabled controls whether auto-refresh is active
	Enabled bool `json:"enabled"`
}

// --- Data Classification ---

type DataClassification struct {
	// Level: public, internal, confidential, restricted
	Level string `json:"level"`

	// Categories: PII, PHI, Financial, Credentials, etc.
	Categories []string `json:"categories,omitempty"`

	// PII indicates personally identifiable information is present
	PII bool `json:"pii"`

	// Sensitive indicates sensitive data is present
	Sensitive bool `json:"sensitive"`

	// Encrypted indicates data is encrypted at rest
	Encrypted bool `json:"encrypted"`

	// RetentionDays is the maximum retention period
	RetentionDays int `json:"retentionDays,omitempty"`
}

// --- Column Metadata ---

type CatalogColumn struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Nullable       bool     `json:"nullable"`
	Description    string   `json:"description,omitempty"`
	IsPrimaryKey   bool     `json:"isPrimaryKey,omitempty"`
	IsForeignKey   bool     `json:"isForeignKey,omitempty"`
	ForeignKeyRef  string   `json:"foreignKeyRef,omitempty"`
	DefaultValue   string   `json:"defaultValue,omitempty"`
	Classification string   `json:"classification,omitempty"`
	IsPII          bool     `json:"isPII,omitempty"`
	Tags           []string `json:"tags,omitempty"`

	// Profiling stats (populated by enrichment reconciler)
	Stats *ColumnStats `json:"stats,omitempty"`
}

type ColumnStats struct {
	NullCount     int64   `json:"nullCount"`
	NullPercent   float64 `json:"nullPercent"`
	DistinctCount int64   `json:"distinctCount"`
	MinValue      string  `json:"minValue,omitempty"`
	MaxValue      string  `json:"maxValue,omitempty"`
	AvgLength     float64 `json:"avgLength,omitempty"`
	SampleValues  []string `json:"sampleValues,omitempty"`
}

// --- CatalogAssetSpec ---

type CatalogAssetSpec struct {
	// AssetType: table, view, topic, bucket, api, pipeline, model, dataset
	AssetType AssetType `json:"assetType"`

	// DataSourceRef references a DataSource resource for connection details
	DataSourceRef string `json:"dataSourceRef"`

	// Database name within the datasource
	Database string `json:"database,omitempty"`

	// Schema (e.g. "public" in PostgreSQL)
	Schema string `json:"schema,omitempty"`

	// TableName is the actual table/view/topic name in the source system
	TableName string `json:"tableName,omitempty"`

	// Owner is the team or person responsible for this asset
	Owner string `json:"owner,omitempty"`

	// Domain is the business domain (sales, finance, engineering, etc.)
	Domain string `json:"domain,omitempty"`

	// Description is a human-readable description of the asset
	Description string `json:"description,omitempty"`

	// Tags for categorization and search
	Tags []string `json:"tags,omitempty"`

	// Classification describes data sensitivity
	Classification DataClassification `json:"classification,omitempty"`

	// Columns describes the schema of the asset
	Columns []CatalogColumn `json:"columns,omitempty"`

	// RefreshPolicy controls how often metadata is re-scanned
	RefreshPolicy RefreshPolicy `json:"refreshPolicy,omitempty"`

	// Upstream lists asset names this asset depends on (for lineage)
	Upstream []string `json:"upstream,omitempty"`

	// Downstream lists asset names that depend on this asset
	Downstream []string `json:"downstream,omitempty"`

	// Documentation is a URL or inline markdown documentation
	Documentation string `json:"documentation,omitempty"`

	// QualityRules references QualityRule resources attached to this asset
	QualityRules []string `json:"qualityRules,omitempty"`

	// ContractRef references a DataContract resource for this asset
	ContractRef string `json:"contractRef,omitempty"`
}

// --- CatalogAssetResourceStatus ---

type CatalogAssetResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	// RowCount is the number of rows (for tables/views)
	RowCount int64 `json:"rowCount"`

	// SizeBytes is the estimated storage size
	SizeBytes int64 `json:"sizeBytes"`

	// ColumnCount is the number of columns discovered
	ColumnCount int `json:"columnCount"`

	// IndexCount is the number of indexes on the table
	IndexCount int `json:"indexCount"`

	// LastScannedAt is when metadata was last refreshed from source
	LastScannedAt *time.Time `json:"lastScannedAt,omitempty"`

	// LastModifiedAt is when the source data was last modified
	LastModifiedAt *time.Time `json:"lastModifiedAt,omitempty"`

	// QualityScore is the overall data quality score (0-100)
	QualityScore float64 `json:"qualityScore"`

	// PopularityScore tracks query frequency (higher = more used)
	PopularityScore float64 `json:"popularityScore"`

	// FreshnessStatus: fresh, stale, unknown
	FreshnessStatus string `json:"freshnessStatus"`

	// ScanError records the last scan error if any
	ScanError string `json:"scanError,omitempty"`

	// ConsecutiveScanFailures tracks sequential scan failures for backoff
	ConsecutiveScanFailures int `json:"consecutiveScanFailures,omitempty"`

	// ProfiledAt is when column profiling was last run
	ProfiledAt *time.Time `json:"profiledAt,omitempty"`

	// ClassificationDetected is auto-detected classification
	ClassificationDetected *DataClassification `json:"classificationDetected,omitempty"`
}

// --- CatalogAssetResource ---

type CatalogAssetResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   CatalogAssetSpec           `json:"spec"`
	Status CatalogAssetResourceStatus `json:"status"`
}

// --- resources.Resource implementation ---

func (r *CatalogAssetResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *CatalogAssetResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *CatalogAssetResource) GetStatus() *resources.ObjectStatus   { return &r.Status.ObjectStatus }
func (r *CatalogAssetResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *CatalogAssetResource) DeepCopy() resources.Resource {
	cp := *r
	if len(r.Spec.Columns) > 0 {
		cp.Spec.Columns = make([]CatalogColumn, len(r.Spec.Columns))
		copy(cp.Spec.Columns, r.Spec.Columns)
	}
	if len(r.Spec.Tags) > 0 {
		cp.Spec.Tags = make([]string, len(r.Spec.Tags))
		copy(cp.Spec.Tags, r.Spec.Tags)
	}
	return &cp
}

// --- reconciler.Resource implementation ---

func (r *CatalogAssetResource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *CatalogAssetResource) GetGeneration() int64         { return r.Generation }
func (r *CatalogAssetResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }

// =====================================================
// CatalogCollectionResource — logical grouping of assets
// =====================================================

type CatalogCollectionSpec struct {
	// DisplayName is the human-readable collection name
	DisplayName string `json:"displayName"`

	// Description of the collection
	Description string `json:"description,omitempty"`

	// Domain is the business domain this collection belongs to
	Domain string `json:"domain,omitempty"`

	// Owner is the team or person responsible
	Owner string `json:"owner,omitempty"`

	// AssetSelector selects assets by labels
	AssetSelector map[string]string `json:"assetSelector,omitempty"`

	// AssetRefs explicitly lists asset names in this collection
	AssetRefs []string `json:"assetRefs,omitempty"`

	// Tags for categorization
	Tags []string `json:"tags,omitempty"`
}

type CatalogCollectionResourceStatus struct {
	resources.ObjectStatus `json:",inline"`

	// AssetCount is the number of assets in this collection
	AssetCount int `json:"assetCount"`

	// TotalSizeBytes is the combined size of all assets
	TotalSizeBytes int64 `json:"totalSizeBytes"`

	// AvgQualityScore is the average quality across assets
	AvgQualityScore float64 `json:"avgQualityScore"`

	// LastUpdatedAt is when the collection membership was last evaluated
	LastUpdatedAt *time.Time `json:"lastUpdatedAt,omitempty"`
}

type CatalogCollectionResource struct {
	resources.TypeMeta   `json:",inline"`
	resources.ObjectMeta `json:"metadata"`

	Spec   CatalogCollectionSpec           `json:"spec"`
	Status CatalogCollectionResourceStatus `json:"status"`
}

func (r *CatalogCollectionResource) GetObjectMeta() *resources.ObjectMeta { return &r.ObjectMeta }
func (r *CatalogCollectionResource) GetTypeMeta() *resources.TypeMeta     { return &r.TypeMeta }
func (r *CatalogCollectionResource) GetStatus() *resources.ObjectStatus {
	return &r.Status.ObjectStatus
}
func (r *CatalogCollectionResource) SetStatus(s *resources.ObjectStatus) {
	if s != nil {
		r.Status.ObjectStatus = *s
	}
}
func (r *CatalogCollectionResource) DeepCopy() resources.Resource {
	cp := *r
	if len(r.Spec.AssetRefs) > 0 {
		cp.Spec.AssetRefs = make([]string, len(r.Spec.AssetRefs))
		copy(cp.Spec.AssetRefs, r.Spec.AssetRefs)
	}
	return &cp
}

func (r *CatalogCollectionResource) GetKey() string {
	if r.Namespace == "" {
		return r.Name
	}
	return r.Namespace + "/" + r.Name
}
func (r *CatalogCollectionResource) GetGeneration() int64         { return r.Generation }
func (r *CatalogCollectionResource) GetObservedGeneration() int64 { return r.Status.ObservedGeneration }
