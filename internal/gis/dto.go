package gis

import "time"

// ─────────────────────────────────────────────────────────────────────────────
// API Response DTOs
// ─────────────────────────────────────────────────────────────────────────────

// LayerResponse is returned for single layer operations.
type LayerResponse struct {
	Layer   GISLayer `json:"layer"`
	Message string   `json:"message,omitempty"`
}

// LayerListResponse is returned by GET /api/v1/gis/layers.
type LayerListResponse struct {
	Layers []GISLayer `json:"layers"`
	Count  int        `json:"count"`
}

// RegionResponse is returned for single region operations.
type RegionResponse struct {
	Region  GISRegion `json:"region"`
	Message string    `json:"message,omitempty"`
}

// RegionListResponse is returned by GET /api/v1/gis/regions.
type RegionListResponse struct {
	Regions []GISRegion `json:"regions"`
	Count   int         `json:"count"`
}

// MarkerResponse is returned for single marker operations.
type MarkerResponse struct {
	Marker  GISMarker `json:"marker"`
	Message string    `json:"message,omitempty"`
}

// MarkerListResponse is returned by GET /api/v1/gis/markers.
type MarkerListResponse struct {
	Markers []GISMarker `json:"markers"`
	Count   int         `json:"count"`
}

// DatasetResponse is returned for single dataset operations.
type DatasetResponse struct {
	Dataset GISDataset `json:"dataset"`
	Message string     `json:"message,omitempty"`
}

// DatasetListResponse is returned by GET /api/v1/gis/datasets.
type DatasetListResponse struct {
	Datasets []GISDataset `json:"datasets"`
	Count    int          `json:"count"`
}

// SummaryResponse is returned by GET /api/v1/gis/summary.
type SummaryResponse struct {
	Summary GISSummary `json:"summary"`
}

// DashboardListResponse is returned by GET /api/v1/gis/dashboards.
type DashboardListResponse struct {
	Dashboards []DashboardTypeInfo `json:"dashboards"`
	Count      int                 `json:"count"`
}

// DashboardTypeInfo is a summary of a dashboard category.
type DashboardTypeInfo struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Layers      int    `json:"layers"`
	Regions     int    `json:"regions"`
	Markers     int    `json:"markers"`
	Datasets    int    `json:"datasets"`
}

// DashboardDataResponse is returned by GET /api/v1/gis/dashboards/:type.
type DashboardDataResponse struct {
	Data GISDashboardData `json:"data"`
}

// HealthResponse is the module health response.
type HealthResponse struct {
	Status      string `json:"status"`
	UptimeSec   int64  `json:"uptimeSeconds"`
	TotalLayers int    `json:"totalLayers"`
	TotalRegions int   `json:"totalRegions"`
	TotalMarkers int   `json:"totalMarkers"`
	Module      string `json:"module"`
}

// MetricsResponse holds GIS metrics.
type MetricsResponse struct {
	TotalLayers      int64 `json:"totalLayers"`
	TotalRegions     int64 `json:"totalRegions"`
	TotalMarkers     int64 `json:"totalMarkers"`
	TotalDatasets    int64 `json:"totalDatasets"`
	TotalViews       int64 `json:"totalViews"`
	TotalConversions int64 `json:"totalConversions"`
	UptimeSeconds    int64 `json:"uptimeSeconds"`
}

// AuditLogResponse is returned by GET /api/v1/gis/audit.
type AuditLogResponse struct {
	Events []AuditEventDTO `json:"events"`
	Count  int             `json:"count"`
}

// AuditEventDTO is a single audit event.
type AuditEventDTO struct {
	Timestamp time.Time   `json:"timestamp"`
	Severity  string      `json:"severity"`
	Category  string      `json:"category"`
	Action    string      `json:"action"`
	EntityID  string      `json:"entityId,omitempty"`
	Message   string      `json:"message"`
}

// ─────────────────────────────────────────────────────────────────────────────
// API Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// CreateLayerRequest is the body for POST /api/v1/gis/layers.
type CreateLayerRequest struct {
	Name    string     `json:"name" binding:"required"`
	Type    string     `json:"type" binding:"required"`
	Visible bool       `json:"visible"`
	Style   LayerStyle `json:"style,omitempty"`
	URL     string     `json:"url,omitempty"`
}

// CreateRegionRequest is the body for POST /api/v1/gis/regions.
type CreateRegionRequest struct {
	Name       string                 `json:"name" binding:"required"`
	Type       string                 `json:"type" binding:"required"`
	ParentID   string                 `json:"parentId,omitempty"`
	Center     [2]float64             `json:"center"`
	Bounds     [4]float64             `json:"bounds,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// CreateMarkerRequest is the body for POST /api/v1/gis/markers.
type CreateMarkerRequest struct {
	Name       string                 `json:"name" binding:"required"`
	Lat        float64                `json:"lat" binding:"required"`
	Lng        float64                `json:"lng" binding:"required"`
	Category   string                 `json:"category"`
	Icon       string                 `json:"icon,omitempty"`
	Color      string                 `json:"color,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// CreateDatasetRequest is the body for POST /api/v1/gis/datasets.
type CreateDatasetRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description,omitempty"`
	Unit        string          `json:"unit,omitempty"`
	Columns     []DatasetColumn `json:"columns" binding:"required"`
}
