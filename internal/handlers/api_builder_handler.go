package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"

	"example.com/axiomnizam/internal/scanner"
)

// =====================================================================
// API Builder Handler — GUI-driven API & resource creation,
// CSV-to-Dashboard ingestion, and Dashboard <-> GIS conversion
// =====================================================================

// ---------- API Builder Models ----------

// CustomAPI represents an API endpoint created through the GUI builder
type CustomAPI struct {
	ID             string            `json:"id"`
	APIType        string            `json:"api_type,omitempty"` // rest, graphql
	Name           string            `json:"name"`
	Method         string            `json:"method"` // GET, POST, PUT, DELETE
	Path           string            `json:"path"`
	SQLTemplate    string            `json:"sql_template,omitempty"`
	GraphQLQuery   string            `json:"graphql_query,omitempty"`
	GraphQLOpName  string            `json:"graphql_operation_name,omitempty"`
	Description    string            `json:"description"`
	Category       string            `json:"category"`
	SourceDatabase string            `json:"source_database,omitempty"`
	SourceServer   string            `json:"source_server,omitempty"`
	AuthRequired   bool              `json:"auth_required"`
	RateLimit      int               `json:"rate_limit"` // requests per minute, 0=unlimited
	CacheEnabled   bool              `json:"cache_enabled"`
	CacheTTL       int               `json:"cache_ttl"` // seconds, 0=default(300)
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
	cachedResult   interface{}       // in-memory cache of last test result
	cachedAt       time.Time         // when the cache was last set
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

// CSVUpload tracks a file upload that was converted to a dashboard
type CSVUpload struct {
	ID             string                   `json:"id"`
	Filename       string                   `json:"filename"`
	FileType       string                   `json:"file_type"` // csv, json, xlsx
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

// ScanResult tracks file scan results
type FileScanRecord struct {
	ID        string            `json:"id"`
	Filename  string            `json:"filename"`
	FileSize  int64             `json:"file_size"`
	SHA256    string            `json:"sha256"`
	Safe      bool              `json:"safe"`
	Findings  []scanner.Finding `json:"findings"`
	ScannedAt time.Time         `json:"scanned_at"`
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
	scanRecords map[string]*FileScanRecord
	apiData     map[string][]map[string]interface{} // stores mock data per API
	csvData     map[string][][]string               // raw CSV data per upload
	db          map[string]*gorm.DB
	etcd        *clientv3.Client
	stateKey    string
	scanOrch    *scanner.Orchestrator
	// reference to analytics & GIS handlers for conversion
	analyticsHandler *AnalyticsHandler
	gisHandler       *GISHandler
}

type apiBuilderState struct {
	CustomAPIs map[string]*CustomAPI               `json:"custom_apis"`
	APIData    map[string][]map[string]interface{} `json:"api_data"`
}

func NewAPIBuilderHandler(ah *AnalyticsHandler, gh *GISHandler, db map[string]*gorm.DB, etcd *clientv3.Client) *APIBuilderHandler {
	// Build scanner pipeline
	orchestrator := scanner.NewOrchestrator(
		scanner.NewMetadataScanner(100*1024*1024),
		scanner.NewMIMEScanner([]string{
			"text/plain", "text/csv", "text/html", "text/xml",
			"application/json", "application/xml", "application/pdf",
			"application/zip", "application/gzip",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.ms-excel",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/msword",
			"image/png", "image/jpeg", "image/gif", "image/svg+xml", "image/webp",
			"audio/mpeg", "video/mp4",
		}),
		&scanner.SVGScanner{},
		&scanner.MacroScanner{},
		scanner.NewArchiveScanner(5, 1024*1024*1024),
		scanner.NewClamAVScanner(getClamAVAddr()),
	)
	h := &APIBuilderHandler{
		customAPIs:       make(map[string]*CustomAPI),
		csvUploads:       make(map[string]*CSVUpload),
		conversions:      make(map[string]*ConversionResult),
		scanRecords:      make(map[string]*FileScanRecord),
		apiData:          make(map[string][]map[string]interface{}),
		csvData:          make(map[string][][]string),
		db:               db,
		etcd:             etcd,
		stateKey:         "axiomnizam:builder:custom_apis",
		scanOrch:         orchestrator,
		analyticsHandler: ah,
		gisHandler:       gh,
	}
	if !h.loadState() {
		h.seedData()
		h.persistStateLocked()
	}
	return h
}

func (h *APIBuilderHandler) loadState() bool {
	if h.etcd == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.etcd.Get(ctx, h.stateKey)
	if err != nil {
		log.Printf("api-builder: failed to load persisted state from etcd: %v", err)
		return false
	}
	if len(resp.Kvs) == 0 {
		return false
	}

	var state apiBuilderState
	if err := json.Unmarshal(resp.Kvs[0].Value, &state); err != nil {
		log.Printf("api-builder: failed to decode persisted state: %v", err)
		return false
	}

	if state.CustomAPIs == nil {
		state.CustomAPIs = make(map[string]*CustomAPI)
	}
	if state.APIData == nil {
		state.APIData = make(map[string][]map[string]interface{})
	}

	h.customAPIs = state.CustomAPIs
	h.apiData = state.APIData
	return true
}

func (h *APIBuilderHandler) persistStateLocked() {
	if h.etcd == nil {
		return
	}

	state := apiBuilderState{
		CustomAPIs: h.customAPIs,
		APIData:    h.apiData,
	}
	payload, err := json.Marshal(state)
	if err != nil {
		log.Printf("api-builder: failed to encode state: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.etcd.Put(ctx, h.stateKey, string(payload)); err != nil {
		log.Printf("api-builder: failed to persist state to etcd: %v", err)
	}
}

func getClamAVAddr() string {
	if v := strings.TrimSpace(os.Getenv("SAFEGATE_CLAMAV_ADDR")); v != "" {
		return v
	}
	return "clamav:3310"
}

// ===================================================================
// API Builder CRUD
// ===================================================================

func (h *APIBuilderHandler) ListAPIs(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	category := c.Query("category")
	status := c.Query("status")
	apiType := strings.ToLower(strings.TrimSpace(c.Query("api_type")))

	result := make([]*CustomAPI, 0, len(h.customAPIs))
	for _, api := range h.customAPIs {
		currentType := strings.ToLower(strings.TrimSpace(api.APIType))
		if currentType == "" {
			currentType = "rest"
		}
		if apiType != "" && currentType != apiType {
			continue
		}
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
		APIType        string            `json:"api_type"`
		Name           string            `json:"name" binding:"required"`
		Method         string            `json:"method"`
		Path           string            `json:"path"`
		SQLTemplate    string            `json:"sql_template"`
		GraphQLQuery   string            `json:"graphql_query"`
		GraphQLOpName  string            `json:"graphql_operation_name"`
		Description    string            `json:"description"`
		Category       string            `json:"category"`
		SourceDatabase string            `json:"source_database"`
		SourceServer   string            `json:"source_server"`
		AuthRequired   bool              `json:"auth_required"`
		RateLimit      int               `json:"rate_limit"`
		RequestSchema  *SchemaDefinition `json:"request_schema"`
		ResponseSchema *SchemaDefinition `json:"response_schema"`
		MockResponse   interface{}       `json:"mock_response"`
		Headers        map[string]string `json:"headers"`
		QueryParams    []ParamDef        `json:"query_params"`
		CacheEnabled   bool              `json:"cache_enabled"`
		CacheTTL       int               `json:"cache_ttl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiType := strings.ToLower(strings.TrimSpace(req.APIType))
	if apiType == "" {
		apiType = "rest"
	}
	if apiType != "rest" && apiType != "graphql" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "api_type must be rest or graphql"})
		return
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	path := strings.TrimSpace(req.Path)

	if apiType == "rest" {
		if method == "" || path == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "method and path are required for REST APIs"})
			return
		}
		if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "method must be GET, POST, PUT, DELETE, or PATCH"})
			return
		}

		sqlTemplate := strings.TrimSpace(req.SQLTemplate)
		if sqlTemplate == "" && strings.TrimSpace(req.SourceDatabase) != "" {
			legacyTemplate, _ := extractQueryFromMock(req.MockResponse)
			sqlTemplate = strings.TrimSpace(legacyTemplate)
		}
		if strings.TrimSpace(req.SourceDatabase) != "" && sqlTemplate == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sql_template is required when source_database is set"})
			return
		}
		if sqlTemplate != "" && !isStrictReadOnlyQuery(sqlTemplate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sql_template must be read-only (SELECT/WITH/SHOW/DESCRIBE/EXPLAIN)"})
			return
		}
		req.SQLTemplate = sqlTemplate
	} else {
		if method == "" {
			method = "POST"
		}
		if path == "" {
			path = "/api/graphql"
		}
		if strings.TrimSpace(req.GraphQLQuery) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "graphql_query is required for GraphQL APIs"})
			return
		}
	}

	id := "api-" + uuid.New().String()[:8]
	now := time.Now()

	ttl := req.CacheTTL
	if req.CacheEnabled && ttl <= 0 {
		ttl = 300 // default 5 minutes
	}

	api := &CustomAPI{
		ID:             id,
		APIType:        apiType,
		Name:           req.Name,
		Method:         method,
		Path:           path,
		SQLTemplate:    strings.TrimSpace(req.SQLTemplate),
		GraphQLQuery:   strings.TrimSpace(req.GraphQLQuery),
		GraphQLOpName:  strings.TrimSpace(req.GraphQLOpName),
		Description:    req.Description,
		Category:       req.Category,
		SourceDatabase: strings.TrimSpace(req.SourceDatabase),
		SourceServer:   strings.TrimSpace(req.SourceServer),
		AuthRequired:   req.AuthRequired,
		RateLimit:      req.RateLimit,
		CacheEnabled:   req.CacheEnabled,
		CacheTTL:       ttl,
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
	h.persistStateLocked()

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
	if strings.TrimSpace(api.APIType) == "" {
		api.APIType = "rest"
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if v, ok := req["name"].(string); ok && v != "" {
		api.Name = v
	}
	if v, ok := req["api_type"].(string); ok {
		vt := strings.ToLower(strings.TrimSpace(v))
		if vt != "" && vt != "rest" && vt != "graphql" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "api_type must be rest or graphql"})
			return
		}
		if vt != "" {
			api.APIType = vt
		}
	}
	if v, ok := req["method"].(string); ok && strings.TrimSpace(v) != "" {
		method := strings.ToUpper(strings.TrimSpace(v))
		if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "method must be GET, POST, PUT, DELETE, or PATCH"})
			return
		}
		api.Method = method
	}
	if v, ok := req["path"].(string); ok && strings.TrimSpace(v) != "" {
		api.Path = strings.TrimSpace(v)
	}
	if v, ok := req["endpoint"].(string); ok && strings.TrimSpace(v) != "" {
		api.Path = strings.TrimSpace(v)
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
	if v, ok := req["source_database"].(string); ok {
		api.SourceDatabase = strings.TrimSpace(v)
	}
	if v, ok := req["source_server"].(string); ok {
		api.SourceServer = strings.TrimSpace(v)
	}
	if v, ok := req["graphql_query"].(string); ok {
		api.GraphQLQuery = strings.TrimSpace(v)
	}
	if v, ok := req["graphql_operation_name"].(string); ok {
		api.GraphQLOpName = strings.TrimSpace(v)
	}
	if v, ok := req["sql_template"].(string); ok {
		template := strings.TrimSpace(v)
		if template != "" && !isStrictReadOnlyQuery(template) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sql_template must be read-only (SELECT/WITH/SHOW/DESCRIBE/EXPLAIN)"})
			return
		}
		api.SQLTemplate = template
	}
	if v, ok := req["auth_required"].(bool); ok {
		api.AuthRequired = v
	}
	if v, ok := req["rate_limit"].(float64); ok {
		api.RateLimit = int(v)
	}
	if v, ok := req["cache_enabled"].(bool); ok {
		api.CacheEnabled = v
		if !v {
			api.cachedResult = nil
		}
	}
	if v, ok := req["cache_ttl"].(float64); ok {
		api.CacheTTL = int(v)
	}
	if api.APIType == "graphql" {
		if strings.TrimSpace(api.Path) == "" {
			api.Path = "/api/graphql"
		}
		if strings.TrimSpace(api.Method) == "" {
			api.Method = "POST"
		}
	} else if strings.TrimSpace(api.SourceDatabase) != "" {
		template := strings.TrimSpace(api.SQLTemplate)
		if template == "" {
			legacyTemplate, _ := extractQueryFromMock(api.MockResponse)
			template = strings.TrimSpace(legacyTemplate)
		}
		if template == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sql_template is required when source_database is set"})
			return
		}
		if !isStrictReadOnlyQuery(template) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sql_template must be read-only (SELECT/WITH/SHOW/DESCRIBE/EXPLAIN)"})
			return
		}
		api.SQLTemplate = template
	}
	api.UpdatedAt = time.Now()
	h.persistStateLocked()

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
	h.persistStateLocked()

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "api deleted"})
}

// TestAPI executes a mock call against a custom API and returns the mock response
func (h *APIBuilderHandler) TestAPI(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := c.Param("id")
	api, ok := h.customAPIs[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "api not found"})
		return
	}
	api.HitCount++

	// Cache check
	if api.CacheEnabled && api.cachedResult != nil {
		ttl := time.Duration(api.CacheTTL) * time.Second
		if ttl <= 0 {
			ttl = 300 * time.Second
		}
		if time.Since(api.cachedAt) < ttl {
			c.JSON(http.StatusOK, gin.H{
				"status":    "success",
				"api_id":    api.ID,
				"api_type":  api.APIType,
				"method":    api.Method,
				"path":      api.Path,
				"operation": api.GraphQLOpName,
				"cached":    true,
				"cache_ttl": api.CacheTTL,
				"response":  api.cachedResult,
			})
			return
		}
	}

	var response interface{}
	if api.MockResponse != nil {
		response = api.MockResponse
	} else {
		response = gin.H{
			"message": "API endpoint active, no mock response configured",
			"data":    h.apiData[id],
		}
	}

	// Store in cache if enabled
	if api.CacheEnabled {
		api.cachedResult = response
		api.cachedAt = time.Now()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"api_id":    api.ID,
		"api_type":  api.APIType,
		"method":    api.Method,
		"path":      api.Path,
		"operation": api.GraphQLOpName,
		"cached":    false,
		"cache_ttl": api.CacheTTL,
		"response":  response,
	})
}

// InvokeCustomAPI executes runtime calls for builder-created REST APIs.
// Routes are mounted under /api/custom/* and resolved against saved API definitions.
func (h *APIBuilderHandler) InvokeCustomAPI(c *gin.Context) {
	method := strings.ToUpper(strings.TrimSpace(c.Request.Method))
	path := strings.TrimSpace(c.Request.URL.Path)

	h.mu.Lock()
	defer h.mu.Unlock()

	pathMatches := make([]*CustomAPI, 0)
	for _, candidate := range h.customAPIs {
		if strings.ToLower(strings.TrimSpace(candidate.APIType)) != "rest" {
			continue
		}
		if !matchBuilderPath(candidate.Path, path) {
			continue
		}
		pathMatches = append(pathMatches, candidate)
	}

	if len(pathMatches) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":            "custom api not found",
			"requested_path":   path,
			"requested_method": method,
		})
		return
	}

	var api *CustomAPI
	allowedMethods := make([]string, 0)
	for _, candidate := range pathMatches {
		candidateMethod := strings.ToUpper(strings.TrimSpace(candidate.Method))
		allowedMethods = append(allowedMethods, candidateMethod)
		if candidateMethod == method {
			api = candidate
			break
		}
	}

	if api == nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error":            "method not allowed for custom api",
			"requested_path":   path,
			"requested_method": method,
			"allowed_methods":  allowedMethods,
		})
		return
	}

	if strings.ToLower(strings.TrimSpace(api.Status)) != "active" {
		c.JSON(http.StatusForbidden, gin.H{"error": "custom api is not active"})
		return
	}

	api.HitCount++

	if api.CacheEnabled && api.cachedResult != nil {
		ttl := time.Duration(api.CacheTTL) * time.Second
		if ttl <= 0 {
			ttl = 300 * time.Second
		}
		if time.Since(api.cachedAt) < ttl {
			c.JSON(http.StatusOK, gin.H{
				"status":    "success",
				"api_id":    api.ID,
				"api_type":  api.APIType,
				"method":    api.Method,
				"path":      api.Path,
				"cached":    true,
				"cache_ttl": api.CacheTTL,
				"params":    c.Request.URL.Query(),
				"response":  api.cachedResult,
			})
			return
		}
	}

	query := resolveStoredSQLTemplate(api)
	params := make([]interface{}, 0)
	if query != "" {
		if !isStrictReadOnlyQuery(query) {
			c.JSON(http.StatusForbidden, gin.H{"error": "custom api sql_template must be read-only"})
			return
		}

		placeholderCount := countSQLPlaceholders(query)
		if placeholderCount > 0 {
			paramDefs := api.QueryParams
			if len(paramDefs) > placeholderCount {
				paramDefs = paramDefs[:placeholderCount]
			}
			if len(paramDefs) > 0 && len(paramDefs) < placeholderCount {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":                    "insufficient query_params definitions for sql_template placeholders",
					"required_placeholders":    placeholderCount,
					"defined_query_parameters": len(paramDefs),
				})
				return
			}

			extractedParams, paramErr := extractRuntimeParamsFromRequest(c, method, paramDefs)
			if paramErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": paramErr.Error()})
				return
			}
			params = extractedParams
			if len(params) != placeholderCount {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":               "parameter count does not match sql_template placeholders",
					"expected_parameters": placeholderCount,
					"received_parameters": len(params),
				})
				return
			}
		}
	}

	var response interface{}
	if query != "" {
		dbType := strings.ToLower(strings.TrimSpace(api.SourceDatabase))
		dbConn := h.db[dbType]
		if dbType == "" || dbConn == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":           "source database is not configured for this custom api",
				"source_database": api.SourceDatabase,
			})
			return
		}

		rows, err := dbConn.Raw(query, params...).Rows()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "query execution failed",
				"detail": err.Error(),
			})
			return
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "failed to get result columns",
				"detail": err.Error(),
			})
			return
		}

		result := make([]map[string]interface{}, 0)
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":  "failed to scan result row",
					"detail": err.Error(),
				})
				return
			}

			entry := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					entry[col] = string(b)
				} else {
					entry[col] = val
				}
			}
			result = append(result, entry)
		}

		response = gin.H{
			"source_database": dbType,
			"query":           query,
			"params":          params,
			"count":           len(result),
			"data":            result,
		}
	} else if api.MockResponse != nil {
		response = api.MockResponse
	} else {
		response = gin.H{
			"message": "API endpoint active, no mock response configured",
			"data":    h.apiData[api.ID],
		}
	}

	if api.CacheEnabled {
		api.cachedResult = response
		api.cachedAt = time.Now()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"api_id":    api.ID,
		"api_type":  api.APIType,
		"method":    api.Method,
		"path":      api.Path,
		"cached":    false,
		"cache_ttl": api.CacheTTL,
		"params":    c.Request.URL.Query(),
		"response":  response,
	})
}

func matchBuilderPath(pattern, actual string) bool {
	pattern = normalizeBuilderRuntimePath(pattern)
	actual = normalizeBuilderRuntimePath(actual)

	if pattern == actual {
		return true
	}

	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	actualParts := strings.Split(strings.Trim(actual, "/"), "/")

	if len(patternParts) != len(actualParts) {
		return false
	}

	for i := 0; i < len(patternParts); i++ {
		segment := strings.TrimSpace(patternParts[i])
		if strings.HasPrefix(segment, ":") {
			continue
		}
		if segment != strings.TrimSpace(actualParts[i]) {
			return false
		}
	}

	return true
}

func normalizeBuilderRuntimePath(path string) string {
	normalized := "/" + strings.Trim(strings.TrimSpace(path), "/")
	if normalized == "/api/custom" {
		return "/"
	}
	if strings.HasPrefix(normalized, "/api/custom/") {
		trimmed := strings.TrimPrefix(normalized, "/api/custom")
		if trimmed == "" {
			return "/"
		}
		return "/" + strings.Trim(strings.TrimSpace(trimmed), "/")
	}
	return normalized
}

func extractQueryFromMock(mock interface{}) (string, []interface{}) {
	m, ok := mock.(map[string]interface{})
	if !ok {
		return "", nil
	}

	query, _ := m["query"].(string)
	paramsRaw, _ := m["params"].([]interface{})
	if paramsRaw == nil {
		return query, nil
	}

	params := make([]interface{}, 0, len(paramsRaw))
	for _, p := range paramsRaw {
		params = append(params, p)
	}
	return query, params
}

func resolveStoredSQLTemplate(api *CustomAPI) string {
	if api == nil {
		return ""
	}
	template := strings.TrimSpace(api.SQLTemplate)
	if template != "" {
		return template
	}
	legacyTemplate, _ := extractQueryFromMock(api.MockResponse)
	return strings.TrimSpace(legacyTemplate)
}

func extractRuntimeParamsFromRequest(c *gin.Context, method string, defs []ParamDef) ([]interface{}, error) {
	bodyParamsByName, bodyParamsList, err := extractRuntimeBodyParams(c, method)
	if err != nil {
		return nil, err
	}

	if len(defs) == 0 {
		if len(bodyParamsList) > 0 {
			return bodyParamsList, nil
		}
		return make([]interface{}, 0), nil
	}

	params := make([]interface{}, 0, len(defs))
	missing := make([]string, 0)

	for _, def := range defs {
		name := strings.TrimSpace(def.Name)
		if name == "" {
			continue
		}

		rawFromQuery := strings.TrimSpace(c.Query(name))
		if rawFromQuery != "" {
			parsed, parseErr := parseParamValue(def.Type, rawFromQuery)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid value for query parameter %q: %v", name, parseErr)
			}
			params = append(params, parsed)
			continue
		}

		if bodyParamsByName != nil {
			if raw, ok := bodyParamsByName[name]; ok {
				parsed, parseErr := parseParamValue(def.Type, raw)
				if parseErr != nil {
					return nil, fmt.Errorf("invalid value for parameter %q: %v", name, parseErr)
				}
				params = append(params, parsed)
				continue
			}
		}

		if strings.TrimSpace(def.Default) != "" {
			parsed, parseErr := parseParamValue(def.Type, def.Default)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid default value for parameter %q: %v", name, parseErr)
			}
			params = append(params, parsed)
			continue
		}

		if def.Required {
			missing = append(missing, name)
			continue
		}

		params = append(params, nil)
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required query parameters: %s", strings.Join(missing, ", "))
	}

	return params, nil
}

func extractRuntimeBodyParams(c *gin.Context, method string) (map[string]interface{}, []interface{}, error) {
	if method == "GET" || c.Request.ContentLength == 0 {
		return nil, nil, nil
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		return nil, nil, fmt.Errorf("invalid request body: expected JSON parameter object")
	}

	if rawParams, ok := body["params"]; ok {
		switch typed := rawParams.(type) {
		case map[string]interface{}:
			return typed, nil, nil
		case []interface{}:
			return nil, typed, nil
		default:
			return nil, nil, fmt.Errorf("invalid params field: expected object or array")
		}
	}

	return body, nil, nil
}

func parseParamValue(paramType string, raw interface{}) (interface{}, error) {
	t := strings.ToLower(strings.TrimSpace(paramType))
	if t == "" || t == "string" {
		return fmt.Sprintf("%v", raw), nil
	}

	switch t {
	case "int", "integer":
		switch v := raw.(type) {
		case float64:
			return int64(v), nil
		case int:
			return int64(v), nil
		case int32:
			return int64(v), nil
		case int64:
			return v, nil
		case string:
			parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err != nil {
				return nil, err
			}
			return parsed, nil
		default:
			parsed, err := strconv.ParseInt(strings.TrimSpace(fmt.Sprintf("%v", raw)), 10, 64)
			if err != nil {
				return nil, err
			}
			return parsed, nil
		}
	case "number", "float", "decimal":
		switch v := raw.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int32:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil {
				return nil, err
			}
			return parsed, nil
		default:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(fmt.Sprintf("%v", raw)), 64)
			if err != nil {
				return nil, err
			}
			return parsed, nil
		}
	case "bool", "boolean":
		switch v := raw.(type) {
		case bool:
			return v, nil
		case string:
			parsed, err := strconv.ParseBool(strings.TrimSpace(v))
			if err != nil {
				return nil, err
			}
			return parsed, nil
		default:
			parsed, err := strconv.ParseBool(strings.TrimSpace(fmt.Sprintf("%v", raw)))
			if err != nil {
				return nil, err
			}
			return parsed, nil
		}
	default:
		return raw, nil
	}
}

func countSQLPlaceholders(query string) int {
	return strings.Count(query, "?")
}

func isStrictReadOnlyQuery(query string) bool {
	normalized := strings.ToUpper(strings.Join(strings.Fields(strings.TrimSpace(query)), " "))
	if normalized == "" {
		return false
	}

	if !(strings.HasPrefix(normalized, "SELECT") ||
		strings.HasPrefix(normalized, "WITH") ||
		strings.HasPrefix(normalized, "SHOW") ||
		strings.HasPrefix(normalized, "DESCRIBE") ||
		strings.HasPrefix(normalized, "EXPLAIN")) {
		return false
	}

	blockedKeywords := []string{
		" INSERT ", " UPDATE ", " DELETE ", " DROP ", " ALTER ", " TRUNCATE ",
		" CREATE ", " REPLACE ", " MERGE ", " GRANT ", " REVOKE ", " CALL ",
		" EXEC ", " EXECUTE ",
	}
	padded := " " + normalized + " "
	for _, blocked := range blockedKeywords {
		if strings.Contains(padded, blocked) {
			return false
		}
	}

	return true
}

func (h *APIBuilderHandler) GetSummary(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	apiTypeFilter := strings.ToLower(strings.TrimSpace(c.Query("api_type")))
	active, inactive, draft := 0, 0, 0
	byCategory := map[string]int{}
	byMethod := map[string]int{}
	byAPIType := map[string]int{}
	var totalHits int64

	for _, api := range h.customAPIs {
		currentType := strings.ToLower(strings.TrimSpace(api.APIType))
		if currentType == "" {
			currentType = "rest"
		}
		if apiTypeFilter != "" && currentType != apiTypeFilter {
			continue
		}
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
		byAPIType[currentType]++
		totalHits += api.HitCount
	}

	totalAPIs := active + inactive + draft

	c.JSON(http.StatusOK, gin.H{
		"status":            "success",
		"total_apis":        totalAPIs,
		"active":            active,
		"inactive":          inactive,
		"draft":             draft,
		"total_hits":        totalHits,
		"by_category":       byCategory,
		"by_method":         byMethod,
		"by_api_type":       byAPIType,
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

	// Read file bytes for multi-format parsing
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file: " + err.Error()})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))

	// Run SafeGate scanner pipeline on the uploaded file
	claimedType := header.Header.Get("Content-Type")
	scanInfo := &scanner.FileInfo{
		Filename:  header.Filename,
		Extension: ext,
		MIMEType:  claimedType,
		Size:      int64(len(fileBytes)),
		SHA256:    fmt.Sprintf("%x", sha256.Sum256(fileBytes)),
		Content:   fileBytes,
	}
	scanResult := h.scanOrch.Scan(scanInfo)
	if !scanResult.Safe {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "File failed security scan — upload rejected",
			"safe":     false,
			"findings": scanResult.Findings,
		})
		return
	}

	var headers []string
	var dataRows [][]string

	switch ext {
	case ".csv":
		reader := csv.NewReader(bytes.NewReader(fileBytes))
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
		headers = records[0]
		dataRows = records[1:]

	case ".json":
		var jsonData []map[string]interface{}
		if err := json.Unmarshal(fileBytes, &jsonData); err != nil {
			// Try object with data array
			var wrapper map[string]interface{}
			if err2 := json.Unmarshal(fileBytes, &wrapper); err2 != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: must be an array of objects or an object with a data array"})
				return
			}
			// Look for array fields
			found := false
			for _, v := range wrapper {
				if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
					for _, item := range arr {
						if obj, ok := item.(map[string]interface{}); ok {
							jsonData = append(jsonData, obj)
						}
					}
					found = true
					break
				}
			}
			if !found || len(jsonData) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "JSON must contain an array of objects"})
				return
			}
		}
		if len(jsonData) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON array is empty"})
			return
		}
		// Extract headers from all keys
		keySet := map[string]bool{}
		for _, obj := range jsonData {
			for k := range obj {
				keySet[k] = true
			}
		}
		for k := range keySet {
			headers = append(headers, k)
		}
		sort.Strings(headers)
		// Convert to string rows
		for _, obj := range jsonData {
			row := make([]string, len(headers))
			for i, h := range headers {
				if v, ok := obj[h]; ok {
					row[i] = fmt.Sprintf("%v", v)
				}
			}
			dataRows = append(dataRows, row)
		}

	case ".xlsx", ".xls":
		f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Excel file: " + err.Error()})
			return
		}
		defer f.Close()
		sheetName := f.GetSheetName(0)
		if sheetName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Excel file has no sheets"})
			return
		}
		rows, err := f.GetRows(sheetName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read Excel sheet: " + err.Error()})
			return
		}
		if len(rows) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Excel sheet must have header row and at least one data row"})
			return
		}
		headers = rows[0]
		dataRows = rows[1:]

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type: " + ext + ". Supported: .csv, .json, .xlsx, .xls"})
		return
	}

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

	fileType := "csv"
	if ext == ".json" {
		fileType = "json"
	} else if ext == ".xlsx" || ext == ".xls" {
		fileType = "xlsx"
	}

	upload := &CSVUpload{
		ID:          id,
		Filename:    header.Filename,
		FileType:    fileType,
		Rows:        len(dataRows),
		Columns:     len(headers),
		ColumnNames: headers,
		ColumnTypes: colTypes,
		SampleData:  sampleData,
		HasGeoData:  hasGeo,
		Status:      "analyzed",
		CreatedAt:   now,
	}

	// Reconstruct records in CSV format for internal storage
	records := make([][]string, 0, len(dataRows)+1)
	records = append(records, headers)
	records = append(records, dataRows...)

	h.mu.Lock()
	h.csvUploads[id] = upload
	h.csvData[id] = records
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"upload":          upload,
		"message":         fmt.Sprintf("%s file analyzed. Call POST /generate-dashboard to create an analytics dashboard.", strings.ToUpper(fileType)),
		"can_convert_gis": hasGeo,
		"scan_safe":       scanResult.Safe,
		"scan_findings":   len(scanResult.Findings),
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
// File Scanner (SafeGate Pipeline)
// ===================================================================

// ScanFile scans an uploaded file through the SafeGate security pipeline
func (h *APIBuilderHandler) ScanFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required: " + err.Error()})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file: " + err.Error()})
		return
	}

	// Build SHA256
	hash := sha256.Sum256(fileBytes)
	sha := fmt.Sprintf("%x", hash)

	// Build FileInfo for scanner
	claimedType := header.Header.Get("Content-Type")
	info := &scanner.FileInfo{
		Filename:  header.Filename,
		Extension: strings.ToLower(filepath.Ext(header.Filename)),
		MIMEType:  claimedType,
		Size:      int64(len(fileBytes)),
		SHA256:    sha,
		Content:   fileBytes,
	}

	result := h.scanOrch.Scan(info)

	record := &FileScanRecord{
		ID:        "scan-" + uuid.New().String()[:8],
		Filename:  header.Filename,
		FileSize:  int64(len(fileBytes)),
		SHA256:    sha,
		Safe:      result.Safe,
		Findings:  result.Findings,
		ScannedAt: time.Now(),
	}

	h.mu.Lock()
	h.scanRecords[record.ID] = record
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"scan":     record,
		"safe":     result.Safe,
		"findings": len(result.Findings),
	})
}

// ListScans returns all file scan records
func (h *APIBuilderHandler) ListScans(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*FileScanRecord, 0, len(h.scanRecords))
	for _, r := range h.scanRecords {
		result = append(result, r)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ScannedAt.After(result[j].ScannedAt) })

	c.JSON(http.StatusOK, gin.H{"status": "success", "count": len(result), "scans": result})
}

// GetScannerHealth returns the scanner pipeline status
func (h *APIBuilderHandler) GetScannerHealth(c *gin.Context) {
	scanners := h.scanOrch.ScannerNames()
	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"scanner_count": len(scanners),
		"scanners":      scanners,
		"total_scans":   len(h.scanRecords),
	})
}

// ===================================================================
// Dashboard Deletion
// ===================================================================

// DeleteDashboard removes a generated analytics dashboard
func (h *APIBuilderHandler) DeleteDashboard(c *gin.Context) {
	dashID := c.Param("id")

	h.analyticsHandler.mu.Lock()
	_, ok := h.analyticsHandler.dashboards[dashID]
	if !ok {
		h.analyticsHandler.mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}
	delete(h.analyticsHandler.dashboards, dashID)
	h.analyticsHandler.mu.Unlock()

	// Clear references in CSV uploads
	h.mu.Lock()
	for _, u := range h.csvUploads {
		if u.DashboardID == dashID {
			u.DashboardID = ""
			if u.Status == "dashboard_created" {
				u.Status = "analyzed"
			}
		}
	}
	h.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "dashboard deleted", "id": dashID})
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
		if strings.TrimSpace(api.APIType) == "" {
			api.APIType = "rest"
		}
		h.customAPIs[api.ID] = api
		h.apiData[api.ID] = []map[string]interface{}{}
	}

	// Seed a sample CSV upload
	h.csvUploads["csv-demo-001"] = &CSVUpload{
		ID:          "csv-demo-001",
		Filename:    "sales_data_2025.csv",
		FileType:    "csv",
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
