package apibuilder

import "time"

// ─────────────────────────────────────────────────────────────────────────────
// API Response DTOs
// ─────────────────────────────────────────────────────────────────────────────

// APIListResponse is returned by GET /api/v1/apibuilder/apis.
type APIListResponse struct {
	Count int            `json:"count"`
	APIs  []CustomAPIInfo `json:"apis"`
}

// CustomAPIInfo is a summary view of a custom API.
type CustomAPIInfo struct {
	ID             string    `json:"id"`
	APIType        string    `json:"apiType"`
	Name           string    `json:"name"`
	Method         string    `json:"method"`
	Path           string    `json:"path"`
	Description    string    `json:"description,omitempty"`
	Category       string    `json:"category,omitempty"`
	SourceDatabase string    `json:"sourceDatabase,omitempty"`
	AuthRequired   bool      `json:"authRequired"`
	Status         string    `json:"status"`
	HitCount       int64     `json:"hitCount"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// APIActionResponse is returned by create/update/delete operations.
type APIActionResponse struct {
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

// UploadResponse is returned by POST /api/v1/apibuilder/upload.
type UploadResponse struct {
	ID          string   `json:"id"`
	Filename    string   `json:"filename"`
	Rows        int      `json:"rows"`
	Columns     int      `json:"columns"`
	ColumnNames []string `json:"columnNames,omitempty"`
	HasGeoData  bool     `json:"hasGeoData"`
	Message     string   `json:"message"`
}

// ScanResponse is returned by POST /api/v1/apibuilder/scan.
type ScanResponse struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	SHA256    string `json:"sha256"`
	Safe      bool   `json:"safe"`
	Findings  int    `json:"findings"`
	ScannedAt string `json:"scannedAt"`
}

// ConversionResponse is returned by POST /api/v1/apibuilder/convert.
type ConversionResponse struct {
	ID             string        `json:"id"`
	SourceType     string        `json:"sourceType"`
	TargetType     string        `json:"targetType"`
	TargetID       string        `json:"targetId"`
	FieldMappings  []FieldMapping `json:"fieldMappings"`
	Confidence     float64       `json:"confidence"`
	GeoFieldsFound int           `json:"geoFieldsFound"`
	Message        string        `json:"message"`
}

// DashboardListResponse is returned by GET /api/v1/apibuilder/dashboards.
type DashboardListResponse struct {
	Dashboards []DashboardInfo `json:"dashboards"`
	Count      int             `json:"count"`
}

// DashboardInfo is a summary view of an analytics dashboard.
type DashboardInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Category    string    `json:"category,omitempty"`
	WidgetCount int       `json:"widgetCount"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// HealthResponse is the module health response.
type HealthResponse struct {
	Status       string `json:"status"`
	UptimeSec    int64  `json:"uptimeSeconds"`
	ActiveAPIs   int    `json:"activeApis"`
	TotalHits    int64  `json:"totalHits"`
	TotalUploads int64  `json:"totalUploads"`
	TotalScans   int64  `json:"totalScans"`
	Module       string `json:"module"`
}

// MetricsResponse holds apibuilder metrics.
type MetricsResponse struct {
	TotalAPIs        int64  `json:"totalApis"`
	TotalHits        int64  `json:"totalHits"`
	TotalUploads     int64  `json:"totalUploads"`
	TotalScans       int64  `json:"totalScans"`
	SafeScans        int64  `json:"safeScans"`
	UnsafeScans      int64  `json:"unsafeScans"`
	TotalConversions int64  `json:"totalConversions"`
	UptimeSeconds    int64  `json:"uptimeSeconds"`
}

// ─────────────────────────────────────────────────────────────────────────────
// API Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// CreateAPIRequest is the body for POST /api/v1/apibuilder/apis.
type CreateAPIRequest struct {
	APIType        string            `json:"apiType"`
	Name           string            `json:"name" binding:"required"`
	Method         string            `json:"method" binding:"required"`
	Path           string            `json:"path" binding:"required"`
	SQLTemplate    string            `json:"sqlTemplate,omitempty"`
	SQLPolicyMode  string            `json:"sqlPolicyMode,omitempty"`
	GraphQLQuery   string            `json:"graphqlQuery,omitempty"`
	GraphQLOpName  string            `json:"graphqlOperationName,omitempty"`
	Description    string            `json:"description,omitempty"`
	Category       string            `json:"category,omitempty"`
	SourceDatabase string            `json:"sourceDatabase,omitempty"`
	SourceServer   string            `json:"sourceServer,omitempty"`
	AuthRequired   bool              `json:"authRequired"`
	RateLimit      int               `json:"rateLimit,omitempty"`
	CacheEnabled   bool              `json:"cacheEnabled"`
	CacheTTL       int               `json:"cacheTtl,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
}

// UpdateAPIRequest is the body for PUT /api/v1/apibuilder/apis/:id.
type UpdateAPIRequest struct {
	Name           string            `json:"name,omitempty"`
	Method         string            `json:"method,omitempty"`
	Path           string            `json:"path,omitempty"`
	SQLTemplate    string            `json:"sqlTemplate,omitempty"`
	Description    string            `json:"description,omitempty"`
	Category       string            `json:"category,omitempty"`
	Status         string            `json:"status,omitempty"`
	AuthRequired   *bool             `json:"authRequired,omitempty"`
	RateLimit      *int              `json:"rateLimit,omitempty"`
	CacheEnabled   *bool             `json:"cacheEnabled,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
}

// ConvertRequest is the body for POST /api/v1/apibuilder/convert.
type ConvertRequest struct {
	SourceType string `json:"sourceType" binding:"required"`
	SourceID   string `json:"sourceId" binding:"required"`
	TargetType string `json:"targetType" binding:"required"`
}

// SQLQueryRequest is the body for POST /api/v1/apibuilder/sql-query.
type SQLQueryRequest struct {
	SQL        string        `json:"sql" binding:"required"`
	Database   string        `json:"database,omitempty"`
	Params     []interface{} `json:"params,omitempty"`
}
