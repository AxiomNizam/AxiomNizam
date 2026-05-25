package lineage

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LineageHandler handles lineage endpoints
type LineageHandler struct {
	manager        LineageManager
	dualWriteStore LineageDualWriteStore
}

// NewLineageHandler creates handler
func NewLineageHandler(manager LineageManager) *LineageHandler {
	return &LineageHandler{manager: manager}
}

// GetNode handles GET /api/v1/lineage/nodes/:id
func (h *LineageHandler) GetNode(c *gin.Context) {
	id := c.Param("id")
	node, err := h.manager.GetNode(id)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: "node not found"})
		return
	}

	c.JSON(http.StatusOK, node)
}

// ListNodes handles GET /api/v1/lineage/nodes
func (h *LineageHandler) ListNodes(c *gin.Context) {
	tenantID := c.Query("tenantId")
	resourceType := c.Query("resourceType")

	nodes, err := h.manager.ListNodes(tenantID, resourceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, NodeListResponse{Nodes: nodes, Count: len(nodes)})
}

// GetLineageGraph handles GET /api/v1/lineage/:resourceType/:resourceId
func (h *LineageHandler) GetLineageGraph(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")

	graph, err := h.manager.GetLineageGraph(resourceType, resourceID)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: "lineage not found"})
		return
	}

	c.JSON(http.StatusOK, graph)
}

// GetUpstreamLineage handles GET /api/v1/lineage/upstream/:resourceType/:resourceId
func (h *LineageHandler) GetUpstreamLineage(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")

	paths, err := h.manager.GetUpstreamLineage(resourceType, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, PathListResponse{Paths: paths, Count: len(paths)})
}

// GetDownstreamLineage handles GET /api/v1/lineage/downstream/:resourceType/:resourceId
func (h *LineageHandler) GetDownstreamLineage(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")

	paths, err := h.manager.GetDownstreamLineage(resourceType, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, PathListResponse{Paths: paths, Count: len(paths)})
}

// GetImpactAnalysis handles GET /api/v1/lineage/impact/:resourceType/:resourceId
func (h *LineageHandler) GetImpactAnalysis(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")

	impact, err := h.manager.GetImpactAnalysis(resourceType, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, impact)
}

// GetColumnLineage handles GET /api/v1/lineage/columns
func (h *LineageHandler) GetColumnLineage(c *gin.Context) {
	sourceCol := c.Query("sourceColumn")
	targetCol := c.Query("targetColumn")

	lineage, err := h.manager.GetColumnLineage(sourceCol, targetCol)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: "lineage not found"})
		return
	}

	c.JSON(http.StatusOK, lineage)
}

// TraceDataFlow handles GET /api/v1/lineage/trace
func (h *LineageHandler) TraceDataFlow(c *gin.Context) {
	sourceID := c.Query("sourceId")
	targetID := c.Query("targetId")

	path, err := h.manager.TraceDataFlow(sourceID, targetID)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: "no data flow path found"})
		return
	}

	c.JSON(http.StatusOK, path)
}

// GetStatistics handles GET /api/v1/lineage/statistics
func (h *LineageHandler) GetStatistics(c *gin.Context) {
	tenantID := c.Query("tenantId")

	stats, err := h.manager.GetStatistics(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// RegisterLineageRoutes registers all lineage routes
func RegisterLineageRoutes(router *gin.Engine, manager LineageManager) {
	handler := NewLineageHandler(manager)

	group := router.Group("/api/v1/lineage")
	{
		group.GET("/nodes/:id", handler.GetNode)
		group.GET("/nodes", handler.ListNodes)
		group.GET("/:resourceType/:resourceId", handler.GetLineageGraph)
		group.GET("/upstream/:resourceType/:resourceId", handler.GetUpstreamLineage)
		group.GET("/downstream/:resourceType/:resourceId", handler.GetDownstreamLineage)
		group.GET("/impact/:resourceType/:resourceId", handler.GetImpactAnalysis)
		group.GET("/columns", handler.GetColumnLineage)
		group.GET("/trace", handler.TraceDataFlow)
		group.GET("/statistics", handler.GetStatistics)
	}
}

// LineageManager interface
type LineageManager interface {
	GetNode(id string) (*LineageNode, error)
	ListNodes(tenantID, resourceType string) ([]*LineageNode, error)
	GetLineageGraph(resourceType, resourceID string) (*LineageGraph, error)
	GetUpstreamLineage(resourceType, resourceID string) ([]*LineagePath, error)
	GetDownstreamLineage(resourceType, resourceID string) ([]*LineagePath, error)
	GetImpactAnalysis(resourceType, resourceID string) (*ImpactAnalysis, error)
	GetColumnLineage(sourceCol, targetCol string) (*ColumnLineage, error)
	TraceDataFlow(sourceID, targetID string) (*LineagePath, error)
	GetStatistics(tenantID string) (*LineageStatistics, error)
}
