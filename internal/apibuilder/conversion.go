package apibuilder

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
	h.persistStateLocked()
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
	h.generatedDashboards[dashID] = dashboard
	h.persistStateLocked()
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
	h.persistStateLocked()
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
// File Scanner (SafeGate Pipeline)
// ===================================================================

// ScanFile scans an uploaded file through the SafeGate security pipeline
