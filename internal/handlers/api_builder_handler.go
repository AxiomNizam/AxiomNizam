package handlers

import (
	"encoding/csv"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// =====================================================================
// API Builder Handler — GUI-driven API & resource creation,
// CSV-to-Dashboard ingestion, and Dashboard <-> GIS conversion
// =====================================================================

// ---------- API Builder Models ----------

// CustomAPI represents an API endpoint created through the GUI builder
type CustomAPI struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Method         string            `json:"method"` // GET, POST, PUT, DELETE
	Path           string            `json:"path"`
	Description    string            `json:"description"`
	Category       string            `json:"category"`
	AuthRequired   bool              `json:"auth_required"`
	RateLimit      int               `json:"rate_limit"` // requests per minute, 0=unlimited
	RequestSchema  *SchemaDefinition `json:"request_schema,omitempty"`
	ResponseSchema *SchemaDefinition `json:"response_schema,omitempty"`
	MockResponse   interface{}       `json:"mock_response,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	QueryParams    []ParamDef        `json:"query_params,omitempty"`
	Status         string            `json:"status"` // active, inactive, draft
	CreatedBy      string            `json:"created_by"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	HitCount       int64             `json:"hit_count"`
}

type SchemaDefinition struct {
	Type       string                 `json:"type"` // object, array, string, number, boolean
	Properties map[string]SchemaField `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
	Example    interface{}            `json:"example,omitempty"`
}

type SchemaField struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
}

type ParamDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
}

// ---------- CSV Dashboard Models ----------

// CSVUpload tracks a CSV file upload that was converted to a dashboard
type CSVUpload struct {
	ID             string                   `json:"id"`
	Filename       string                   `json:"filename"`
	Rows           int                      `json:"rows"`
	Columns        int                      `json:"columns"`
	ColumnNames    []string                 `json:"column_names"`
	ColumnTypes    []string                 `json:"column_types"` // string, number, date, geo_lat, geo_lng, geo_name
	SampleData     []map[string]interface{} `json:"sample_data"`
	DashboardID    string                   `json:"dashboard_id,omitempty"`
	GISDashboardID string                   `json:"gis_dashboard_id,omitempty"`
	HasGeoData     bool                     `json:"has_geo_data"`
	Status         string                   `json:"status"` // uploaded, analyzed, dashboard_created, gis_created
	CreatedAt      time.Time                `json:"created_at"`
}

// ---------- Dashboard <-> GIS Conversion ----------

type ConversionResult struct {
	ID             string         `json:"id"`
	SourceType     string         `json:"source_type"` // dashboard, gis
	SourceID       string         `json:"source_id"`
	TargetType     string         `json:"target_type"` // gis, dashboard
	TargetID       string         `json:"target_id"`
	FieldMappings  []FieldMapping `json:"field_mappings"`
	GeoFieldsFound []string       `json:"geo_fields_found,omitempty"`
	Confidence     float64        `json:"confidence"` // 0-1 how good the conversion is
	Status         string         `json:"status"`     // pending, completed, failed
	CreatedAt      time.Time      `json:"created_at"`
}

type FieldMapping struct {
	SourceField string `json:"source_field"`
	TargetField string `json:"target_field"`
	MappingType string `json:"mapping_type"` // direct, inferred, geo_lat, geo_lng, geo_region
}

// ---------- Handler ----------

type APIBuilderHandler struct {
	mu          sync.RWMutex
	customAPIs  map[string]*CustomAPI
	csvUploads  map[string]*CSVUpload
	conversions map[string]*ConversionResult
	apiData     map[string][]map[string]interface{} // stores mock data per API
	csvData     map[string][][]string               // raw CSV data per upload
	// reference to analytics & GIS handlers for conversion
	analyticsHandler *AnalyticsHandler
	gisHandler       *GISHandler
}

func NewAPIBuilderHandler(ah *AnalyticsHandler, gh *GISHandler) *APIBuilderHandler {
	h := &APIBuilderHandler{
		customAPIs:       make(map[string]*CustomAPI),
		csvUploads:       make(map[string]*CSVUpload),
		conversions:      make(map[string]*ConversionResult),
		apiData:          make(map[string][]map[string]interface{}),
		csvData:          make(map[string][][]string),
		analyticsHandler: ah,
		gisHandler:       gh,
	}
	h.seedData()
	return h
}

// ===================================================================
// API Builder CRUD
// ===================================================================

func (h *APIBuilderHandler) ListAPIs(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	category := c.Query("category")
	status := c.Query("status")

	result := make([]*CustomAPI, 0, len(h.customAPIs))
	for _, api := range h.customAPIs {
		if category != "" && api.Category != category {
			continue
		}
		if status != "" && api.Status != status {
			continue
		}
		result = append(result, api)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })

	c.JSON(http.StatusOK, gin.H{"status": "success", "count": len(result), "apis": result})
}

func (h *APIBuilderHandler) GetAPI(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	id := c.Param("id")
	api, ok := h.customAPIs[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "api not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "api": api})
}

func (h *APIBuilderHandler) CreateAPI(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var req struct {
		Name           string            `json:"name" binding:"required"`
		Method         string            `json:"method" binding:"required"`
		Path           string            `json:"path" binding:"required"`
		Description    string            `json:"description"`
		Category       string            `json:"category"`
		AuthRequired   bool              `json:"auth_required"`
		RateLimit      int               `json:"rate_limit"`
		RequestSchema  *SchemaDefinition `json:"request_schema"`
		ResponseSchema *SchemaDefinition `json:"response_schema"`
		MockResponse   interface{}       `json:"mock_response"`
		Headers        map[string]string `json:"headers"`
		QueryParams    []ParamDef        `json:"query_params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	method := strings.ToUpper(req.Method)
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "method must be GET, POST, PUT, DELETE, or PATCH"})
		return
	}

	id := "api-" + uuid.New().String()[:8]
	now := time.Now()

	api := &CustomAPI{
		ID:             id,
		Name:           req.Name,
		Method:         method,
		Path:           req.Path,
		Description:    req.Description,
		Category:       req.Category,
		AuthRequired:   req.AuthRequired,
		RateLimit:      req.RateLimit,
		RequestSchema:  req.RequestSchema,
		ResponseSchema: req.ResponseSchema,
		MockResponse:   req.MockResponse,
		Headers:        req.Headers,
		QueryParams:    req.QueryParams,
		Status:         "active",
		CreatedBy:      "admin",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	h.customAPIs[id] = api
	h.apiData[id] = make([]map[string]interface{}, 0)

	c.JSON(http.StatusCreated, gin.H{"status": "success", "api": api})
}

func (h *APIBuilderHandler) UpdateAPI(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := c.Param("id")
	api, ok := h.customAPIs[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "api not found"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if v, ok := req["name"].(string); ok && v != "" {
		api.Name = v
	}
	if v, ok := req["description"].(string); ok {
		api.Description = v
	}
	if v, ok := req["status"].(string); ok {
		api.Status = v
	}
	if v, ok := req["category"].(string); ok {
		api.Category = v
	}
	if v, ok := req["auth_required"].(bool); ok {
		api.AuthRequired = v
	}
	if v, ok := req["rate_limit"].(float64); ok {
		api.RateLimit = int(v)
	}
	api.UpdatedAt = time.Now()

	c.JSON(http.StatusOK, gin.H{"status": "success", "api": api})
}

func (h *APIBuilderHandler) DeleteAPI(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := c.Param("id")
	if _, ok := h.customAPIs[id]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "api not found"})
		return
	}
	delete(h.customAPIs, id)
	delete(h.apiData, id)

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "api deleted"})
}

// TestAPI executes a mock call against a custom API and returns the mock response
func (h *APIBuilderHandler) TestAPI(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	id := c.Param("id")
	api, ok := h.customAPIs[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "api not found"})
		return
	}
	api.HitCount++

	if api.MockResponse != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":   "success",
			"api_id":   api.ID,
			"method":   api.Method,
			"path":     api.Path,
			"response": api.MockResponse,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"api_id":  api.ID,
		"method":  api.Method,
		"path":    api.Path,
		"message": "API endpoint active, no mock response configured",
		"data":    h.apiData[id],
	})
}

func (h *APIBuilderHandler) GetSummary(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	active, inactive, draft := 0, 0, 0
	byCategory := map[string]int{}
	byMethod := map[string]int{}
	var totalHits int64

	for _, api := range h.customAPIs {
		switch api.Status {
		case "active":
			active++
		case "inactive":
			inactive++
		case "draft":
			draft++
		}
		byCategory[api.Category]++
		byMethod[api.Method]++
		totalHits += api.HitCount
	}

	c.JSON(http.StatusOK, gin.H{
		"status":            "success",
		"total_apis":        len(h.customAPIs),
		"active":            active,
		"inactive":          inactive,
		"draft":             draft,
		"total_hits":        totalHits,
		"by_category":       byCategory,
		"by_method":         byMethod,
		"total_csv_uploads": len(h.csvUploads),
		"total_conversions": len(h.conversions),
	})
}

// ===================================================================
// CSV Upload -> Auto Dashboard
// ===================================================================

func (h *APIBuilderHandler) UploadCSV(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required: " + err.Error()})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid CSV: " + err.Error()})
		return
	}
	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV must have header row and at least one data row"})
		return
	}

	headers := records[0]
	dataRows := records[1:]

	// Analyze column types
	colTypes := analyzeColumnTypes(headers, dataRows)

	// Check for geo data
	hasGeo := false
	for _, ct := range colTypes {
		if ct == "geo_lat" || ct == "geo_lng" || ct == "geo_name" {
			hasGeo = true
			break
		}
	}

	// Build sample data (first 10 rows)
	sampleSize := 10
	if len(dataRows) < sampleSize {
		sampleSize = len(dataRows)
	}
	sampleData := make([]map[string]interface{}, sampleSize)
	for i := 0; i < sampleSize; i++ {
		row := map[string]interface{}{}
		for j, hdr := range headers {
			if j < len(dataRows[i]) {
				row[hdr] = dataRows[i][j]
			}
		}
		sampleData[i] = row
	}

	id := "csv-" + uuid.New().String()[:8]
	now := time.Now()

	upload := &CSVUpload{
		ID:          id,
		Filename:    header.Filename,
		Rows:        len(dataRows),
		Columns:     len(headers),
		ColumnNames: headers,
		ColumnTypes: colTypes,
		SampleData:  sampleData,
		HasGeoData:  hasGeo,
		Status:      "analyzed",
		CreatedAt:   now,
	}

	h.mu.Lock()
	h.csvUploads[id] = upload
	// Store raw data for dashboard generation
	h.csvData[id] = records
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"upload":          upload,
		"message":         "CSV analyzed. Call POST /generate-dashboard to create an analytics dashboard.",
		"can_convert_gis": hasGeo,
	})
}

func (h *APIBuilderHandler) ListCSVUploads(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*CSVUpload, 0, len(h.csvUploads))
	for _, u := range h.csvUploads {
		result = append(result, u)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })

	c.JSON(http.StatusOK, gin.H{"status": "success", "count": len(result), "uploads": result})
}

func (h *APIBuilderHandler) GetCSVUpload(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	id := c.Param("id")
	u, ok := h.csvUploads[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "upload": u})
}

func (h *APIBuilderHandler) DeleteCSVUpload(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := c.Param("id")
	if _, ok := h.csvUploads[id]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
		return
	}
	delete(h.csvUploads, id)
	delete(h.csvData, id)

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "upload deleted"})
}

// GenerateDashboard creates an analytics dashboard from an uploaded CSV
func (h *APIBuilderHandler) GenerateDashboard(c *gin.Context) {
	id := c.Param("id")

	h.mu.RLock()
	upload, ok := h.csvUploads[id]
	rawData, hasRaw := h.csvData[id]
	h.mu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
		return
	}
	if !hasRaw || len(rawData) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no data available for dashboard generation"})
		return
	}

	headers := rawData[0]
	dataRows := rawData[1:]

	// Generate dashboard with auto-detected widgets
	dashID := "csv-dash-" + uuid.New().String()[:8]
	now := time.Now()
	widgets := generateWidgetsFromCSV(headers, upload.ColumnTypes, dataRows)

	dashboard := &AnalyticsDashboard{
		ID:          dashID,
		Name:        "CSV: " + upload.Filename,
		Description: fmt.Sprintf("Auto-generated from %s (%d rows, %d columns)", upload.Filename, upload.Rows, upload.Columns),
		Category:    "csv-import",
		Widgets:     widgets,
		Filters:     generateFiltersFromCSV(headers, upload.ColumnTypes),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Register in analytics handler
	h.analyticsHandler.mu.Lock()
	h.analyticsHandler.dashboards[dashID] = dashboard
	h.analyticsHandler.mu.Unlock()

	// Update upload record
	h.mu.Lock()
	upload.DashboardID = dashID
	upload.Status = "dashboard_created"
	h.mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"status":       "success",
		"dashboard_id": dashID,
		"dashboard":    dashboard,
		"message":      fmt.Sprintf("Dashboard created with %d auto-generated widgets", len(widgets)),
	})
}

// ===================================================================
// Dashboard <-> GIS Conversion
// ===================================================================

// AnalyzeConversion checks if a dashboard can be converted to GIS or vice versa
func (h *APIBuilderHandler) AnalyzeConversion(c *gin.Context) {
	var req struct {
		SourceType string `json:"source_type" binding:"required"` // dashboard or gis
		SourceID   string `json:"source_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SourceType == "dashboard" {
		h.analyzeDashboardToGIS(c, req.SourceID)
	} else if req.SourceType == "gis" {
		h.analyzeGISToDashboard(c, req.SourceID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source_type must be 'dashboard' or 'gis'"})
	}
}

func (h *APIBuilderHandler) analyzeDashboardToGIS(c *gin.Context, dashID string) {
	h.analyticsHandler.mu.RLock()
	dash, ok := h.analyticsHandler.dashboards[dashID]
	h.analyticsHandler.mu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}

	// Scan widgets for geo data
	geoFields := []string{}
	mappings := []FieldMapping{}
	confidence := 0.0

	for _, w := range dash.Widgets {
		if w.Type == "table" && len(w.Data.Columns) > 0 {
			for _, col := range w.Data.Columns {
				lower := strings.ToLower(col.Key)
				if strings.Contains(lower, "lat") {
					geoFields = append(geoFields, col.Key)
					mappings = append(mappings, FieldMapping{SourceField: col.Key, TargetField: "lat", MappingType: "geo_lat"})
					confidence += 0.3
				} else if strings.Contains(lower, "lng") || strings.Contains(lower, "lon") {
					geoFields = append(geoFields, col.Key)
					mappings = append(mappings, FieldMapping{SourceField: col.Key, TargetField: "lng", MappingType: "geo_lng"})
					confidence += 0.3
				} else if strings.Contains(lower, "region") || strings.Contains(lower, "district") || strings.Contains(lower, "city") || strings.Contains(lower, "country") || strings.Contains(lower, "location") || strings.Contains(lower, "area") || strings.Contains(lower, "zone") {
					geoFields = append(geoFields, col.Key)
					mappings = append(mappings, FieldMapping{SourceField: col.Key, TargetField: "region_name", MappingType: "geo_region"})
					confidence += 0.2
				} else {
					mappings = append(mappings, FieldMapping{SourceField: col.Key, TargetField: col.Key, MappingType: "direct"})
				}
			}
		}
		// Check chart labels for geographic info
		if len(w.Data.Labels) > 0 {
			for _, label := range w.Data.Labels {
				lower := strings.ToLower(label)
				if isLikelyGeoLabel(lower) {
					confidence += 0.1
				}
			}
		}
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	canConvert := confidence >= 0.3

	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"can_convert":      canConvert,
		"confidence":       math.Round(confidence*100) / 100,
		"geo_fields_found": geoFields,
		"field_mappings":   mappings,
		"source":           gin.H{"type": "dashboard", "id": dashID, "name": dash.Name},
		"target_type":      "gis",
		"suggestion":       fmt.Sprintf("Found %d geo-capable fields. Confidence: %.0f%%", len(geoFields), confidence*100),
	})
}

func (h *APIBuilderHandler) analyzeGISToDashboard(c *gin.Context, datasetID string) {
	h.gisHandler.mu.RLock()
	var dataset *GISDataset
	for i := range h.gisHandler.datasets {
		if h.gisHandler.datasets[i].ID == datasetID {
			dataset = &h.gisHandler.datasets[i]
			break
		}
	}
	h.gisHandler.mu.RUnlock()

	if dataset == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "GIS dataset not found"})
		return
	}

	mappings := []FieldMapping{}
	for _, col := range dataset.Columns {
		mappings = append(mappings, FieldMapping{
			SourceField: col.Key,
			TargetField: col.Key,
			MappingType: "direct",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"can_convert":    true,
		"confidence":     0.95,
		"field_mappings": mappings,
		"source":         gin.H{"type": "gis", "id": datasetID, "name": dataset.Name},
		"target_type":    "dashboard",
		"suggestion":     fmt.Sprintf("GIS dataset '%s' with %d columns and %d rows can be fully converted to an analytics dashboard.", dataset.Name, len(dataset.Columns), len(dataset.Rows)),
	})
}

// ConvertDashboardToGIS converts an analytics dashboard into a GIS dataset + markers
func (h *APIBuilderHandler) ConvertDashboardToGIS(c *gin.Context) {
	var req struct {
		DashboardID string         `json:"dashboard_id" binding:"required"`
		Mappings    []FieldMapping `json:"field_mappings"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.analyticsHandler.mu.RLock()
	dash, ok := h.analyticsHandler.dashboards[req.DashboardID]
	h.analyticsHandler.mu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}

	// Extract table data from dashboard widgets
	var rows []map[string]interface{}
	var columns []DatasetColumn
	for _, w := range dash.Widgets {
		if w.Type == "table" && len(w.Data.Rows) > 0 {
			rows = w.Data.Rows
			for _, col := range w.Data.Columns {
				columns = append(columns, DatasetColumn{Key: col.Key, Label: col.Label, Type: col.Type})
			}
			break
		}
	}

	// Create GIS dataset
	dsID := "gis-conv-" + uuid.New().String()[:8]
	now := time.Now()

	gisDataset := GISDataset{
		ID:          dsID,
		Name:        "Converted: " + dash.Name,
		Description: fmt.Sprintf("Converted from analytics dashboard '%s'", dash.Name),
		Columns:     columns,
		Rows:        rows,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Try to create markers from lat/lng mappings
	markers := extractMarkersFromRows(rows, req.Mappings)

	h.gisHandler.mu.Lock()
	h.gisHandler.datasets = append(h.gisHandler.datasets, gisDataset)
	h.gisHandler.markers = append(h.gisHandler.markers, markers...)
	h.gisHandler.mu.Unlock()

	// Record conversion
	convID := "conv-" + uuid.New().String()[:8]
	conv := &ConversionResult{
		ID:            convID,
		SourceType:    "dashboard",
		SourceID:      req.DashboardID,
		TargetType:    "gis",
		TargetID:      dsID,
		FieldMappings: req.Mappings,
		Confidence:    0.9,
		Status:        "completed",
		CreatedAt:     now,
	}

	h.mu.Lock()
	h.conversions[convID] = conv
	h.mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"status":          "success",
		"conversion":      conv,
		"dataset_id":      dsID,
		"markers_created": len(markers),
		"message":         fmt.Sprintf("Dashboard '%s' converted to GIS dataset '%s' with %d markers", dash.Name, gisDataset.Name, len(markers)),
	})
}

// ConvertGISToDashboard converts a GIS dataset into an analytics dashboard
func (h *APIBuilderHandler) ConvertGISToDashboard(c *gin.Context) {
	var req struct {
		DatasetID string `json:"dataset_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.gisHandler.mu.RLock()
	var dataset *GISDataset
	for i := range h.gisHandler.datasets {
		if h.gisHandler.datasets[i].ID == req.DatasetID {
			dataset = &h.gisHandler.datasets[i]
			break
		}
	}
	h.gisHandler.mu.RUnlock()

	if dataset == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "GIS dataset not found"})
		return
	}

	// Generate widgets from GIS dataset
	dashID := "gis-dash-" + uuid.New().String()[:8]
	now := time.Now()
	widgets := generateWidgetsFromGISDataset(dataset)

	dashboard := &AnalyticsDashboard{
		ID:          dashID,
		Name:        "GIS: " + dataset.Name,
		Description: fmt.Sprintf("Converted from GIS dataset '%s'", dataset.Name),
		Category:    "gis-conversion",
		Widgets:     widgets,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	h.analyticsHandler.mu.Lock()
	h.analyticsHandler.dashboards[dashID] = dashboard
	h.analyticsHandler.mu.Unlock()

	convID := "conv-" + uuid.New().String()[:8]
	conv := &ConversionResult{
		ID:         convID,
		SourceType: "gis",
		SourceID:   req.DatasetID,
		TargetType: "dashboard",
		TargetID:   dashID,
		Confidence: 0.95,
		Status:     "completed",
		CreatedAt:  now,
	}

	h.mu.Lock()
	h.conversions[convID] = conv
	h.mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"status":       "success",
		"conversion":   conv,
		"dashboard_id": dashID,
		"widget_count": len(widgets),
		"message":      fmt.Sprintf("GIS dataset '%s' converted to dashboard '%s' with %d widgets", dataset.Name, dashboard.Name, len(widgets)),
	})
}

// GenerateGISFromCSV directly converts a CSV upload to a GIS dashboard (requires geo data)
func (h *APIBuilderHandler) GenerateGISFromCSV(c *gin.Context) {
	id := c.Param("id")

	h.mu.RLock()
	upload, ok := h.csvUploads[id]
	rawData, hasRaw := h.csvData[id]
	h.mu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
		return
	}
	if !hasRaw || len(rawData) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no data available"})
		return
	}
	if !upload.HasGeoData {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV does not contain geographic data (lat/lng columns). Cannot create GIS dashboard."})
		return
	}

	headers := rawData[0]
	dataRows := rawData[1:]

	// Find lat/lng column indices
	latIdx, lngIdx := -1, -1
	nameIdx := -1
	for i, h := range headers {
		lower := strings.ToLower(h)
		if strings.Contains(lower, "lat") {
			latIdx = i
		} else if strings.Contains(lower, "lng") || strings.Contains(lower, "lon") {
			lngIdx = i
		} else if strings.Contains(lower, "name") || strings.Contains(lower, "label") || strings.Contains(lower, "title") {
			nameIdx = i
		}
	}

	// Create GIS dataset
	dsID := "csv-gis-" + uuid.New().String()[:8]
	now := time.Now()

	columns := make([]DatasetColumn, len(headers))
	for i, h := range headers {
		columns[i] = DatasetColumn{Key: h, Label: h, Type: upload.ColumnTypes[i]}
	}

	gisRows := make([]map[string]interface{}, 0, len(dataRows))
	markers := make([]GISMarker, 0)

	for ri, row := range dataRows {
		rowMap := map[string]interface{}{}
		for j, hdr := range headers {
			if j < len(row) {
				rowMap[hdr] = row[j]
			}
		}
		gisRows = append(gisRows, rowMap)

		// Create marker if lat/lng available
		if latIdx >= 0 && lngIdx >= 0 && latIdx < len(row) && lngIdx < len(row) {
			lat, errLat := strconv.ParseFloat(strings.TrimSpace(row[latIdx]), 64)
			lng, errLng := strconv.ParseFloat(strings.TrimSpace(row[lngIdx]), 64)
			if errLat == nil && errLng == nil {
				mName := fmt.Sprintf("Point %d", ri+1)
				if nameIdx >= 0 && nameIdx < len(row) {
					mName = row[nameIdx]
				}
				markers = append(markers, GISMarker{
					ID:         fmt.Sprintf("csv-mkr-%d", ri+1),
					Name:       mName,
					Lat:        lat,
					Lng:        lng,
					Category:   "csv-import",
					Icon:       "📍",
					Color:      "#3b82f6",
					Properties: rowMap,
				})
			}
		}
	}

	gisDataset := GISDataset{
		ID:          dsID,
		Name:        "CSV GIS: " + upload.Filename,
		Description: fmt.Sprintf("GIS dataset from %s", upload.Filename),
		Columns:     columns,
		Rows:        gisRows,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	h.gisHandler.mu.Lock()
	h.gisHandler.datasets = append(h.gisHandler.datasets, gisDataset)
	h.gisHandler.markers = append(h.gisHandler.markers, markers...)
	h.gisHandler.mu.Unlock()

	h.mu.Lock()
	upload.GISDashboardID = dsID
	upload.Status = "gis_created"
	h.mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"status":          "success",
		"dataset_id":      dsID,
		"markers_created": len(markers),
		"rows":            len(gisRows),
		"message":         fmt.Sprintf("GIS dataset created with %d rows and %d markers from %s", len(gisRows), len(markers), upload.Filename),
	})
}

func (h *APIBuilderHandler) ListConversions(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*ConversionResult, 0, len(h.conversions))
	for _, conv := range h.conversions {
		result = append(result, conv)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })

	c.JSON(http.StatusOK, gin.H{"status": "success", "count": len(result), "conversions": result})
}

// ===================================================================
// Helper Functions
// ===================================================================

func analyzeColumnTypes(headers []string, rows [][]string) []string {
	types := make([]string, len(headers))
	for i, hdr := range headers {
		lower := strings.ToLower(hdr)

		// Check header name for geo hints
		if strings.Contains(lower, "lat") || strings.Contains(lower, "latitude") {
			types[i] = "geo_lat"
			continue
		}
		if strings.Contains(lower, "lng") || strings.Contains(lower, "lon") || strings.Contains(lower, "longitude") {
			types[i] = "geo_lng"
			continue
		}
		if lower == "region" || lower == "district" || lower == "city" || lower == "country" || lower == "state" || lower == "province" || lower == "area" || lower == "zone" || lower == "location" {
			types[i] = "geo_name"
			continue
		}

		// Sample values to determine type
		numCount, dateCount := 0, 0
		sampleSize := len(rows)
		if sampleSize > 50 {
			sampleSize = 50
		}
		for j := 0; j < sampleSize; j++ {
			if i >= len(rows[j]) {
				continue
			}
			val := strings.TrimSpace(rows[j][i])
			if val == "" {
				continue
			}
			if _, err := strconv.ParseFloat(val, 64); err == nil {
				numCount++
			}
			if len(val) >= 8 && (strings.Contains(val, "-") || strings.Contains(val, "/")) {
				dateCount++
			}
		}

		threshold := sampleSize / 2
		if threshold < 1 {
			threshold = 1
		}
		if numCount >= threshold {
			types[i] = "number"
		} else if dateCount >= threshold {
			types[i] = "date"
		} else {
			types[i] = "string"
		}
	}
	return types
}

func generateWidgetsFromCSV(headers []string, colTypes []string, rows [][]string) []AnalyticsWidget {
	widgets := []AnalyticsWidget{}
	order := 1

	// Identify numeric and string columns
	numericCols := []int{}
	stringCols := []int{}
	dateCols := []int{}
	for i, ct := range colTypes {
		switch ct {
		case "number", "geo_lat", "geo_lng":
			numericCols = append(numericCols, i)
		case "string", "geo_name":
			stringCols = append(stringCols, i)
		case "date":
			dateCols = append(dateCols, i)
		}
	}

	// 1) KPI widgets for each numeric column (first 4)
	maxKPI := 4
	if len(numericCols) < maxKPI {
		maxKPI = len(numericCols)
	}
	for k := 0; k < maxKPI; k++ {
		col := numericCols[k]
		sum := 0.0
		count := 0
		for _, row := range rows {
			if col < len(row) {
				if v, err := strconv.ParseFloat(strings.TrimSpace(row[col]), 64); err == nil {
					sum += v
					count++
				}
			}
		}
		avg := 0.0
		if count > 0 {
			avg = math.Round(sum/float64(count)*100) / 100
		}

		widgets = append(widgets, AnalyticsWidget{
			ID:     fmt.Sprintf("w-kpi-%d", order),
			Type:   "kpi",
			Title:  fmt.Sprintf("Avg %s", headers[col]),
			Width:  3,
			Height: 1,
			Order:  order,
			Config: WidgetConfig{ShowLegend: false},
			Data:   WidgetData{Value: avg},
		})
		order++
	}

	// 2) Bar chart: first string column vs first numeric column
	if len(stringCols) > 0 && len(numericCols) > 0 {
		sc := stringCols[0]
		nc := numericCols[0]

		// Aggregate by string value
		agg := map[string]float64{}
		for _, row := range rows {
			if sc < len(row) && nc < len(row) {
				key := row[sc]
				if v, err := strconv.ParseFloat(strings.TrimSpace(row[nc]), 64); err == nil {
					agg[key] += v
				}
			}
		}

		labels := make([]string, 0, len(agg))
		values := make([]float64, 0, len(agg))
		for k, v := range agg {
			labels = append(labels, k)
			values = append(values, math.Round(v*100)/100)
		}
		// Limit to top 15
		if len(labels) > 15 {
			labels = labels[:15]
			values = values[:15]
		}

		colors := generateColors(len(labels))
		widgets = append(widgets, AnalyticsWidget{
			ID:    fmt.Sprintf("w-bar-%d", order),
			Type:  "bar",
			Title: fmt.Sprintf("%s by %s", headers[nc], headers[sc]),
			Width: 6, Height: 2, Order: order,
			Config: WidgetConfig{XAxis: headers[sc], YAxis: headers[nc], Colors: colors, ShowLegend: true, ShowGrid: true, Animation: true},
			Data: WidgetData{
				Labels:   labels,
				Datasets: []ChartDataset{{Label: headers[nc], Data: values, BackgroundColor: colors, BorderColor: colors[0], BorderWidth: 1}},
			},
		})
		order++
	}

	// 3) Pie/doughnut for first string column distribution
	if len(stringCols) > 0 {
		sc := stringCols[0]
		freq := map[string]int{}
		for _, row := range rows {
			if sc < len(row) {
				freq[row[sc]]++
			}
		}
		labels := make([]string, 0, len(freq))
		values := make([]float64, 0, len(freq))
		for k, v := range freq {
			labels = append(labels, k)
			values = append(values, float64(v))
		}
		if len(labels) > 10 {
			labels = labels[:10]
			values = values[:10]
		}
		colors := generateColors(len(labels))
		widgets = append(widgets, AnalyticsWidget{
			ID:    fmt.Sprintf("w-pie-%d", order),
			Type:  "doughnut",
			Title: fmt.Sprintf("%s Distribution", headers[sc]),
			Width: 6, Height: 2, Order: order,
			Config: WidgetConfig{Colors: colors, ShowLegend: true, Animation: true},
			Data: WidgetData{
				Labels:   labels,
				Datasets: []ChartDataset{{Label: headers[sc], Data: values, BackgroundColor: colors}},
			},
		})
		order++
	}

	// 4) Line chart if there's a date column and numeric column
	if len(dateCols) > 0 && len(numericCols) > 0 {
		dc := dateCols[0]
		nc := numericCols[0]
		labels := make([]string, 0)
		values := make([]float64, 0)
		maxPts := 30
		step := 1
		if len(rows) > maxPts {
			step = len(rows) / maxPts
		}
		for i := 0; i < len(rows); i += step {
			if dc < len(rows[i]) && nc < len(rows[i]) {
				labels = append(labels, rows[i][dc])
				if v, err := strconv.ParseFloat(strings.TrimSpace(rows[i][nc]), 64); err == nil {
					values = append(values, v)
				} else {
					values = append(values, 0)
				}
			}
		}
		widgets = append(widgets, AnalyticsWidget{
			ID:    fmt.Sprintf("w-line-%d", order),
			Type:  "line",
			Title: fmt.Sprintf("%s Over Time", headers[nc]),
			Width: 12, Height: 2, Order: order,
			Config: WidgetConfig{XAxis: headers[dc], YAxis: headers[nc], Colors: []string{"#3b82f6"}, ShowLegend: true, ShowGrid: true, Animation: true},
			Data: WidgetData{
				Labels:   labels,
				Datasets: []ChartDataset{{Label: headers[nc], Data: values, BorderColor: "#3b82f6", Fill: false, Tension: 0.3}},
			},
		})
		order++
	}

	// 5) Full data table
	tableCols := make([]TableColumn, len(headers))
	for i, h := range headers {
		colType := "string"
		if i < len(colTypes) {
			switch colTypes[i] {
			case "number", "geo_lat", "geo_lng":
				colType = "number"
			case "date":
				colType = "date"
			}
		}
		tableCols[i] = TableColumn{Key: h, Label: h, Type: colType, Sortable: true}
	}

	tableRows := make([]map[string]interface{}, 0)
	maxRows := 100
	if len(rows) < maxRows {
		maxRows = len(rows)
	}
	for i := 0; i < maxRows; i++ {
		rm := map[string]interface{}{}
		for j, hdr := range headers {
			if j < len(rows[i]) {
				rm[hdr] = rows[i][j]
			}
		}
		tableRows = append(tableRows, rm)
	}

	widgets = append(widgets, AnalyticsWidget{
		ID:    fmt.Sprintf("w-table-%d", order),
		Type:  "table",
		Title: "Data Table",
		Width: 12, Height: 3, Order: order,
		Config: WidgetConfig{ShowGrid: true},
		Data: WidgetData{
			Columns: tableCols,
			Rows:    tableRows,
		},
	})

	return widgets
}

func generateFiltersFromCSV(headers []string, colTypes []string) []DashboardFilter {
	filters := []DashboardFilter{}
	for i, ct := range colTypes {
		if ct == "string" || ct == "geo_name" {
			filters = append(filters, DashboardFilter{
				ID:    fmt.Sprintf("f-%d", i),
				Label: headers[i],
				Type:  "select",
				Key:   headers[i],
			})
		} else if ct == "date" {
			filters = append(filters, DashboardFilter{
				ID:    fmt.Sprintf("f-%d", i),
				Label: headers[i],
				Type:  "date-range",
				Key:   headers[i],
			})
		}
		if len(filters) >= 4 {
			break
		}
	}
	return filters
}

func generateWidgetsFromGISDataset(ds *GISDataset) []AnalyticsWidget {
	widgets := []AnalyticsWidget{}
	order := 1

	// KPI: row count
	widgets = append(widgets, AnalyticsWidget{
		ID: "w-kpi-1", Type: "kpi", Title: "Total Records",
		Width: 3, Height: 1, Order: order,
		Data: WidgetData{Value: len(ds.Rows)},
	})
	order++

	// KPI: column count
	widgets = append(widgets, AnalyticsWidget{
		ID: "w-kpi-2", Type: "kpi", Title: "Data Fields",
		Width: 3, Height: 1, Order: order,
		Data: WidgetData{Value: len(ds.Columns)},
	})
	order++

	// Find numeric columns for charts
	numCols := []DatasetColumn{}
	strCols := []DatasetColumn{}
	for _, col := range ds.Columns {
		if col.Type == "number" {
			numCols = append(numCols, col)
		} else {
			strCols = append(strCols, col)
		}
	}

	// Bar chart from first string + numeric cols
	if len(strCols) > 0 && len(numCols) > 0 {
		sc := strCols[0]
		nc := numCols[0]
		labels := []string{}
		values := []float64{}
		for _, row := range ds.Rows {
			if len(labels) >= 15 {
				break
			}
			if l, ok := row[sc.Key]; ok {
				labels = append(labels, fmt.Sprintf("%v", l))
			}
			if v, ok := row[nc.Key]; ok {
				if fv, err := toFloat64(v); err == nil {
					values = append(values, fv)
				}
			}
		}
		colors := generateColors(len(labels))
		widgets = append(widgets, AnalyticsWidget{
			ID: fmt.Sprintf("w-bar-%d", order), Type: "bar",
			Title: fmt.Sprintf("%s by %s", nc.Label, sc.Label),
			Width: 6, Height: 2, Order: order,
			Config: WidgetConfig{Colors: colors, ShowLegend: true, ShowGrid: true, Animation: true},
			Data: WidgetData{
				Labels:   labels,
				Datasets: []ChartDataset{{Label: nc.Label, Data: values, BackgroundColor: colors}},
			},
		})
		order++
	}

	// Full table
	tableCols := make([]TableColumn, len(ds.Columns))
	for i, c := range ds.Columns {
		tableCols[i] = TableColumn{Key: c.Key, Label: c.Label, Type: c.Type, Sortable: true}
	}
	maxRows := 100
	if len(ds.Rows) < maxRows {
		maxRows = len(ds.Rows)
	}
	widgets = append(widgets, AnalyticsWidget{
		ID: fmt.Sprintf("w-table-%d", order), Type: "table",
		Title: "GIS Data Table", Width: 12, Height: 3, Order: order,
		Config: WidgetConfig{ShowGrid: true},
		Data:   WidgetData{Columns: tableCols, Rows: ds.Rows[:maxRows]},
	})

	return widgets
}

func extractMarkersFromRows(rows []map[string]interface{}, mappings []FieldMapping) []GISMarker {
	latField, lngField, nameField := "", "", ""
	for _, m := range mappings {
		switch m.MappingType {
		case "geo_lat":
			latField = m.SourceField
		case "geo_lng":
			lngField = m.SourceField
		case "geo_region":
			nameField = m.SourceField
		}
	}
	if latField == "" || lngField == "" {
		return nil
	}

	markers := make([]GISMarker, 0)
	for i, row := range rows {
		latV, okLat := row[latField]
		lngV, okLng := row[lngField]
		if !okLat || !okLng {
			continue
		}
		lat, errLat := toFloat64(latV)
		lng, errLng := toFloat64(lngV)
		if errLat != nil || errLng != nil {
			continue
		}
		mName := fmt.Sprintf("Point %d", i+1)
		if nameField != "" {
			if v, ok := row[nameField]; ok {
				mName = fmt.Sprintf("%v", v)
			}
		}
		markers = append(markers, GISMarker{
			ID:         fmt.Sprintf("conv-mkr-%d", i+1),
			Name:       mName,
			Lat:        lat,
			Lng:        lng,
			Category:   "conversion",
			Icon:       "📍",
			Color:      "#ef4444",
			Properties: row,
		})
	}
	return markers
}

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(val), 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func isLikelyGeoLabel(lower string) bool {
	geoTerms := []string{"dhaka", "chittagong", "sylhet", "rajshahi", "khulna", "barisal", "rangpur", "mymensingh",
		"north", "south", "east", "west", "zone", "area", "district", "division", "region", "city", "country"}
	for _, term := range geoTerms {
		if strings.Contains(lower, term) {
			return true
		}
	}
	return false
}

func generateColors(n int) []string {
	palette := []string{
		"#3b82f6", "#ef4444", "#10b981", "#f59e0b", "#8b5cf6",
		"#ec4899", "#06b6d4", "#84cc16", "#f97316", "#6366f1",
		"#14b8a6", "#e11d48", "#a855f7", "#0ea5e9", "#22c55e",
	}
	colors := make([]string, n)
	for i := 0; i < n; i++ {
		colors[i] = palette[i%len(palette)]
	}
	return colors
}

// ===================================================================
// Seed Data
// ===================================================================

func (h *APIBuilderHandler) seedData() {
	now := time.Now()

	// Seed sample custom APIs
	sampleAPIs := []*CustomAPI{
		{
			ID: "api-products", Name: "List Products", Method: "GET", Path: "/api/custom/products",
			Description: "Returns paginated product list with filters", Category: "e-commerce",
			AuthRequired: true, RateLimit: 100, Status: "active", CreatedBy: "admin",
			QueryParams: []ParamDef{
				{Name: "page", Type: "number", Required: false, Default: "1"},
				{Name: "limit", Type: "number", Required: false, Default: "20"},
				{Name: "category", Type: "string", Required: false},
			},
			MockResponse: map[string]interface{}{
				"products": []map[string]interface{}{
					{"id": 1, "name": "Widget A", "price": 29.99, "category": "electronics"},
					{"id": 2, "name": "Gadget B", "price": 49.99, "category": "electronics"},
					{"id": 3, "name": "Tool C", "price": 19.99, "category": "tools"},
				},
				"total": 3, "page": 1,
			},
			CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour), HitCount: 142,
		},
		{
			ID: "api-create-order", Name: "Create Order", Method: "POST", Path: "/api/custom/orders",
			Description: "Create a new customer order", Category: "e-commerce",
			AuthRequired: true, RateLimit: 30, Status: "active", CreatedBy: "admin",
			RequestSchema: &SchemaDefinition{
				Type: "object",
				Properties: map[string]SchemaField{
					"customer_id": {Type: "string", Description: "Customer ID"},
					"items":       {Type: "array", Description: "Order items"},
					"total":       {Type: "number", Description: "Order total"},
				},
				Required: []string{"customer_id", "items"},
			},
			MockResponse: map[string]interface{}{"order_id": "ORD-001", "status": "created"},
			CreatedAt:    now.Add(-36 * time.Hour), UpdatedAt: now.Add(-12 * time.Hour), HitCount: 67,
		},
		{
			ID: "api-weather", Name: "Get Weather", Method: "GET", Path: "/api/custom/weather",
			Description: "Weather data for a given city", Category: "external",
			AuthRequired: false, RateLimit: 60, Status: "active", CreatedBy: "admin",
			QueryParams: []ParamDef{
				{Name: "city", Type: "string", Required: true},
			},
			MockResponse: map[string]interface{}{"city": "Dhaka", "temp": 32, "humidity": 78, "condition": "Partly Cloudy"},
			CreatedAt:    now.Add(-24 * time.Hour), UpdatedAt: now, HitCount: 203,
		},
		{
			ID: "api-inventory", Name: "Update Inventory", Method: "PUT", Path: "/api/custom/inventory/:id",
			Description: "Update stock quantity for a product", Category: "warehouse",
			AuthRequired: true, RateLimit: 50, Status: "active", CreatedBy: "admin",
			RequestSchema: &SchemaDefinition{
				Type: "object",
				Properties: map[string]SchemaField{
					"quantity": {Type: "number", Description: "New stock quantity"},
					"location": {Type: "string", Description: "Warehouse location"},
				},
				Required: []string{"quantity"},
			},
			MockResponse: map[string]interface{}{"id": "INV-001", "quantity": 150, "updated": true},
			CreatedAt:    now.Add(-20 * time.Hour), UpdatedAt: now, HitCount: 89,
		},
		{
			ID: "api-user-analytics", Name: "User Analytics", Method: "GET", Path: "/api/custom/analytics/users",
			Description: "User behavior analytics and aggregated metrics", Category: "analytics",
			AuthRequired: true, RateLimit: 20, Status: "draft", CreatedBy: "admin",
			MockResponse: map[string]interface{}{
				"active_users": 1247, "avg_session_min": 12.5, "bounce_rate": 0.32,
				"top_pages": []string{"/dashboard", "/analytics", "/gis"},
			},
			CreatedAt: now.Add(-10 * time.Hour), UpdatedAt: now, HitCount: 0,
		},
	}

	for _, api := range sampleAPIs {
		h.customAPIs[api.ID] = api
		h.apiData[api.ID] = []map[string]interface{}{}
	}

	// Seed a sample CSV upload
	h.csvUploads["csv-demo-001"] = &CSVUpload{
		ID:          "csv-demo-001",
		Filename:    "sales_data_2025.csv",
		Rows:        250,
		Columns:     6,
		ColumnNames: []string{"Region", "Product", "Sales", "Revenue", "Latitude", "Longitude"},
		ColumnTypes: []string{"geo_name", "string", "number", "number", "geo_lat", "geo_lng"},
		SampleData: []map[string]interface{}{
			{"Region": "Dhaka", "Product": "Widget A", "Sales": "150", "Revenue": "4500.00", "Latitude": "23.8103", "Longitude": "90.4125"},
			{"Region": "Chittagong", "Product": "Gadget B", "Sales": "98", "Revenue": "4802.00", "Latitude": "22.3569", "Longitude": "91.7832"},
			{"Region": "Sylhet", "Product": "Tool C", "Sales": "65", "Revenue": "1300.00", "Latitude": "24.8949", "Longitude": "91.8687"},
		},
		HasGeoData: true,
		Status:     "analyzed",
		CreatedAt:  now.Add(-5 * time.Hour),
	}
}
