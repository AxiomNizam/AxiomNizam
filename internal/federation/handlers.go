package federation

// =====================================================
// WS-5.1 — Federation REST API Handlers
//
// Provides virtual table CRUD, federated query execution,
// and EXPLAIN plan endpoints.
// =====================================================

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"github.com/gin-gonic/gin"
)

// FederationHandlers provides REST API handlers for federation.
type FederationHandlers struct {
	vtStore    store.ResourceStore[*VirtualTableResource]
	queryStore store.ResourceStore[*FederatedQueryResource]
	planner    *QueryPlanner
}

// NewFederationHandlers creates new handlers.
func NewFederationHandlers(
	vtStore store.ResourceStore[*VirtualTableResource],
	queryStore store.ResourceStore[*FederatedQueryResource],
	planner *QueryPlanner,
) *FederationHandlers {
	return &FederationHandlers{
		vtStore:    vtStore,
		queryStore: queryStore,
		planner:    planner,
	}
}

// RegisterRoutes registers federation API routes.
func (h *FederationHandlers) RegisterRoutes(rg *gin.RouterGroup) {
	fed := rg.Group("/federation")
	{
		// Virtual tables
		fed.GET("/virtual-tables", h.ListVirtualTables)
		fed.GET("/virtual-tables/:name", h.GetVirtualTable)
		fed.POST("/virtual-tables", h.CreateVirtualTable)
		fed.PUT("/virtual-tables/:name", h.UpdateVirtualTable)
		fed.DELETE("/virtual-tables/:name", h.DeleteVirtualTable)

		// Query execution
		fed.POST("/query", h.ExecuteQuery)
		fed.POST("/explain", h.ExplainQuery)

		// Query history
		fed.GET("/queries", h.ListQueries)
		fed.GET("/queries/:name", h.GetQuery)
	}
}

// --- Virtual Table Handlers ---

func (h *FederationHandlers) ListVirtualTables(c *gin.Context) {
	tables, err := h.vtStore.List(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"virtualTables": tables, "count": len(tables)})
}

func (h *FederationHandlers) GetVirtualTable(c *gin.Context) {
	name := c.Param("name")
	vt, err := h.vtStore.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "virtual table not found", "name": name})
		return
	}
	c.JSON(http.StatusOK, vt)
}

func (h *FederationHandlers) CreateVirtualTable(c *gin.Context) {
	var vt VirtualTableResource
	if err := c.ShouldBindJSON(&vt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vt.Kind = VirtualTableKind
	vt.APIVersion = VirtualTableAPIVersion
	now := time.Now()
	vt.CreatedAt = now
	vt.Generation = 1
	vt.Status.Phase = "Pending"

	if err := h.vtStore.Create(c.Request.Context(), &vt); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, vt)
}

func (h *FederationHandlers) UpdateVirtualTable(c *gin.Context) {
	name := c.Param("name")
	existing, err := h.vtStore.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "virtual table not found", "name": name})
		return
	}

	var updated VirtualTableResource
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated.ObjectMeta = existing.ObjectMeta
	updated.Generation = existing.Generation + 1
	updated.Status = existing.Status

	if err := h.vtStore.Update(c.Request.Context(), &updated); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *FederationHandlers) DeleteVirtualTable(c *gin.Context) {
	name := c.Param("name")
	if err := h.vtStore.Delete(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "virtual table not found", "name": name})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": name})
}

// --- Query Handlers ---

func (h *FederationHandlers) ExecuteQuery(c *gin.Context) {
	var req FederatedQuerySpec
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SQL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sql is required"})
		return
	}
	if req.Format == "" {
		req.Format = QueryFormatJSON
	}
	if req.MaxRows <= 0 {
		req.MaxRows = 10000
	}

	// Create a query resource for tracking.
	query := &FederatedQueryResource{
		Spec: req,
	}
	query.Kind = FederatedQueryKind
	query.APIVersion = FederatedQueryAPIVersion
	now := time.Now()
	query.CreatedAt = now
	query.Generation = 1
	query.Name = fmt.Sprintf("query-%d", now.UnixNano())
	query.Status.Phase = "Running"
	query.Status.QueryStatus = "running"
	query.Status.StartedAt = &now

	// Plan the query.
	if h.planner != nil {
		plan, sources, err := h.planner.Plan(ctx(c), req.SQL)
		if err != nil {
			query.Status.Phase = "Failed"
			query.Status.QueryStatus = "failed"
			query.Status.ErrorMessage = err.Error()
			completedAt := time.Now()
			query.Status.CompletedAt = &completedAt
			query.Status.DurationMs = completedAt.Sub(now).Milliseconds()

			if h.queryStore != nil {
				_ = h.queryStore.Create(c.Request.Context(), query)
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "queryId": query.Name})
			return
		}
		query.Status.Plan = plan
		query.Status.SourcesQueried = sources
	}

	// For now, return the plan. Full execution would stream results.
	completedAt := time.Now()
	query.Status.Phase = "Completed"
	query.Status.QueryStatus = "completed"
	query.Status.CompletedAt = &completedAt
	query.Status.DurationMs = completedAt.Sub(now).Milliseconds()

	if h.queryStore != nil {
		_ = h.queryStore.Create(c.Request.Context(), query)
	}

	c.JSON(http.StatusOK, gin.H{
		"queryId":  query.Name,
		"status":   query.Status.QueryStatus,
		"plan":     query.Status.Plan,
		"duration": fmt.Sprintf("%dms", query.Status.DurationMs),
		"sources":  query.Status.SourcesQueried,
	})
}

func (h *FederationHandlers) ExplainQuery(c *gin.Context) {
	var req struct {
		SQL     string `json:"sql"`
		Analyze bool   `json:"analyze"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SQL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sql is required"})
		return
	}

	if h.planner == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "query planner not available"})
		return
	}

	plan, sources, err := h.planner.Plan(ctx(c), req.SQL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plan":    plan,
		"sources": sources,
	})
}

func (h *FederationHandlers) ListQueries(c *gin.Context) {
	queries, err := h.queryStore.List(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"queries": queries, "count": len(queries)})
}

func (h *FederationHandlers) GetQuery(c *gin.Context) {
	name := c.Param("name")
	query, err := h.queryStore.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "query not found", "name": name})
		return
	}
	c.JSON(http.StatusOK, query)
}

// ctx extracts context from gin.
func ctx(c *gin.Context) context.Context { return c.Request.Context() }
