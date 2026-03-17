package handlers

import (
	"net/http"
	"strings"

	"example.com/axiomnizam/internal/cdc"
	"example.com/axiomnizam/internal/etl"
	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// ==============================================
// CDC & ETL Handler — Full REST API
// Dynamic pipeline management with observability
// ==============================================

type CDCETLHandler struct {
	etlEngine *etl.Engine
	cdcEngine *cdc.PipelineEngine
}

func NewCDCETLHandler(etcd ...*clientv3.Client) *CDCETLHandler {
	cdcCore := cdc.NewChangeDataCapture(etcd...)
	return &CDCETLHandler{
		etlEngine: etl.NewEngine(etcd...),
		cdcEngine: cdc.NewPipelineEngine(cdcCore, etcd...),
	}
}

// ================
// ETL Endpoints
// ================

// ListETLPipelines GET /api/v1/etl/pipelines
func (h *CDCETLHandler) ListETLPipelines(c *gin.Context) {
	pipelines := h.etlEngine.ListPipelines()
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"pipelines": pipelines,
		"total":     len(pipelines),
	})
}

// GetETLPipeline GET /api/v1/etl/pipelines/:id
func (h *CDCETLHandler) GetETLPipeline(c *gin.Context) {
	id := c.Param("id")
	p, ok := h.etlEngine.GetPipeline(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "pipeline not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "pipeline": p})
}

// CreateETLPipeline POST /api/v1/etl/pipelines
func (h *CDCETLHandler) CreateETLPipeline(c *gin.Context) {
	var p etl.Pipeline
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if p.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "name is required"})
		return
	}
	if err := h.etlEngine.CreatePipeline(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "pipeline": p})
}

// UpdateETLPipeline PUT /api/v1/etl/pipelines/:id
func (h *CDCETLHandler) UpdateETLPipeline(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if err := h.etlEngine.UpdatePipeline(id, updates); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	p, _ := h.etlEngine.GetPipeline(id)
	c.JSON(http.StatusOK, gin.H{"status": "success", "pipeline": p})
}

// DeleteETLPipeline DELETE /api/v1/etl/pipelines/:id
func (h *CDCETLHandler) DeleteETLPipeline(c *gin.Context) {
	id := c.Param("id")
	if err := h.etlEngine.DeletePipeline(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "pipeline deleted"})
}

// RunETLPipeline POST /api/v1/etl/pipelines/:id/run
func (h *CDCETLHandler) RunETLPipeline(c *gin.Context) {
	id := c.Param("id")
	run, err := h.etlEngine.RunPipeline(c.Request.Context(), id, "manual")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "run": run})
}

// ListETLRuns GET /api/v1/etl/runs
func (h *CDCETLHandler) ListETLRuns(c *gin.Context) {
	pipelineID := c.Query("pipeline_id")
	runs := h.etlEngine.ListRuns(pipelineID)
	c.JSON(http.StatusOK, gin.H{"status": "success", "runs": runs, "total": len(runs)})
}

// GetETLRun GET /api/v1/etl/runs/:id
func (h *CDCETLHandler) GetETLRun(c *gin.Context) {
	id := c.Param("id")
	run, ok := h.etlEngine.GetRun(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "run not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "run": run})
}

// GetETLConnectors GET /api/v1/etl/connectors
func (h *CDCETLHandler) GetETLConnectors(c *gin.Context) {
	connectors := h.etlEngine.GetConnectors()
	q := strings.TrimSpace(strings.ToLower(c.Query("q")))
	category := strings.TrimSpace(strings.ToLower(c.Query("category")))

	filtered := make([]etl.ConnectorType, 0, len(connectors))
	for _, connector := range connectors {
		if category != "" && strings.ToLower(connector.Category) != category {
			continue
		}
		if q != "" {
			name := strings.ToLower(connector.Name)
			id := strings.ToLower(connector.ID)
			desc := strings.ToLower(connector.Description)
			if !strings.Contains(name, q) && !strings.Contains(id, q) && !strings.Contains(desc, q) {
				continue
			}
		}
		filtered = append(filtered, connector)
	}

	categories := map[string]int{}
	for _, connector := range connectors {
		categories[connector.Category]++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"connectors": filtered,
		"total":      len(filtered),
		"categories": categories,
	})
}

// CreateETLConnector POST /api/v1/etl/connectors
func (h *CDCETLHandler) CreateETLConnector(c *gin.Context) {
	var connector etl.ConnectorType
	if err := c.ShouldBindJSON(&connector); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	if err := h.etlEngine.AddConnector(connector); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "connector": connector})
}

// GetETLConnectorCatalog GET /api/v1/etl/connectors/catalog
func (h *CDCETLHandler) GetETLConnectorCatalog(c *gin.Context) {
	connectors := h.etlEngine.GetConnectors()
	byCategory := map[string][]etl.ConnectorType{}
	for _, connector := range connectors {
		byCategory[connector.Category] = append(byCategory[connector.Category], connector)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"connectors":  connectors,
		"by_category": byCategory,
		"total":       len(connectors),
	})
}

// GetETLOrchestrationCapabilities GET /api/v1/etl/orchestration/capabilities
func (h *CDCETLHandler) GetETLOrchestrationCapabilities(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"capabilities": h.etlEngine.GetOrchestrationCapabilities(),
	})
}

// GetETLBlueprints GET /api/v1/etl/blueprints
func (h *CDCETLHandler) GetETLBlueprints(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"blueprints": h.etlEngine.GetPipelineBlueprints(),
	})
}

// GetETLObservability GET /api/v1/etl/observability
func (h *CDCETLHandler) GetETLObservability(c *gin.Context) {
	obs := h.etlEngine.GetObservability()
	c.JSON(http.StatusOK, gin.H{"status": "success", "observability": obs})
}

// ================
// CDC Endpoints
// ================

// ListCDCPipelines GET /api/v1/cdc/pipelines
func (h *CDCETLHandler) ListCDCPipelines(c *gin.Context) {
	pipelines := h.cdcEngine.ListPipelines()
	c.JSON(http.StatusOK, gin.H{"status": "success", "pipelines": pipelines, "total": len(pipelines)})
}

// GetCDCPipeline GET /api/v1/cdc/pipelines/:id
func (h *CDCETLHandler) GetCDCPipeline(c *gin.Context) {
	id := c.Param("id")
	p, ok := h.cdcEngine.GetPipeline(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "pipeline not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "pipeline": p})
}

// CreateCDCPipeline POST /api/v1/cdc/pipelines
func (h *CDCETLHandler) CreateCDCPipeline(c *gin.Context) {
	var p cdc.CDCPipeline
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if p.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "name is required"})
		return
	}
	if err := h.cdcEngine.CreatePipeline(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "pipeline": p})
}

// UpdateCDCPipeline PUT /api/v1/cdc/pipelines/:id
func (h *CDCETLHandler) UpdateCDCPipeline(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if err := h.cdcEngine.UpdatePipeline(id, updates); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	p, _ := h.cdcEngine.GetPipeline(id)
	c.JSON(http.StatusOK, gin.H{"status": "success", "pipeline": p})
}

// DeleteCDCPipeline DELETE /api/v1/cdc/pipelines/:id
func (h *CDCETLHandler) DeleteCDCPipeline(c *gin.Context) {
	id := c.Param("id")
	if err := h.cdcEngine.DeletePipeline(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "pipeline deleted"})
}

// StartCDCPipeline POST /api/v1/cdc/pipelines/:id/start
func (h *CDCETLHandler) StartCDCPipeline(c *gin.Context) {
	id := c.Param("id")
	if err := h.cdcEngine.StartPipeline(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	p, _ := h.cdcEngine.GetPipeline(id)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "pipeline started", "pipeline": p})
}

// PauseCDCPipeline POST /api/v1/cdc/pipelines/:id/pause
func (h *CDCETLHandler) PauseCDCPipeline(c *gin.Context) {
	id := c.Param("id")
	if err := h.cdcEngine.PausePipeline(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	p, _ := h.cdcEngine.GetPipeline(id)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "pipeline paused", "pipeline": p})
}

// StopCDCPipeline POST /api/v1/cdc/pipelines/:id/stop
func (h *CDCETLHandler) StopCDCPipeline(c *gin.Context) {
	id := c.Param("id")
	if err := h.cdcEngine.StopPipeline(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	p, _ := h.cdcEngine.GetPipeline(id)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "pipeline stopped", "pipeline": p})
}

// GetCDCSourceTypes GET /api/v1/cdc/sources
func (h *CDCETLHandler) GetCDCSourceTypes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "success", "sources": h.cdcEngine.GetSourceTypes()})
}

// GetCDCSinkTypes GET /api/v1/cdc/sinks
func (h *CDCETLHandler) GetCDCSinkTypes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "success", "sinks": h.cdcEngine.GetSinkTypes()})
}

// GetCDCObservability GET /api/v1/cdc/observability
func (h *CDCETLHandler) GetCDCObservability(c *gin.Context) {
	obs := h.cdcEngine.GetObservability()
	c.JSON(http.StatusOK, gin.H{"status": "success", "observability": obs})
}

// ================
// Combined Endpoints
// ================

// GetPlatformOverview GET /api/v1/data-platform/overview
func (h *CDCETLHandler) GetPlatformOverview(c *gin.Context) {
	etlObs := h.etlEngine.GetObservability()
	cdcObs := h.cdcEngine.GetObservability()
	etlPipelines := h.etlEngine.ListPipelines()
	cdcPipelines := h.cdcEngine.ListPipelines()

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"overview": gin.H{
			"etl": gin.H{
				"pipelines_total":      etlObs.PipelinesTotal,
				"runs_total":           etlObs.RunsTotal,
				"runs_success":         etlObs.RunsSuccess,
				"runs_failed":          etlObs.RunsFailed,
				"runs_running":         etlObs.RunsRunning,
				"total_rows_read":      etlObs.TotalRowsRead,
				"total_rows_written":   etlObs.TotalRowsWrite,
				"avg_duration_seconds": etlObs.AvgDuration,
				"pipelines":            etlPipelines,
			},
			"cdc": gin.H{
				"pipelines_total":   cdcObs.PipelinesTotal,
				"pipelines_active":  cdcObs.PipelinesActive,
				"pipelines_paused":  cdcObs.PipelinesPaused,
				"pipelines_failed":  cdcObs.PipelinesFailed,
				"total_events":      cdcObs.TotalEvents,
				"events_per_second": cdcObs.EventsPerSecond,
				"total_errors":      cdcObs.TotalErrors,
				"error_rate":        cdcObs.ErrorRate,
				"avg_lag_ms":        cdcObs.AvgLagMs,
				"pipelines":         cdcPipelines,
			},
			"connectors": gin.H{
				"etl_connectors": len(h.etlEngine.GetConnectors()),
				"cdc_sources":    len(h.cdcEngine.GetSourceTypes()),
				"cdc_sinks":      len(h.cdcEngine.GetSinkTypes()),
			},
		},
	})
}
