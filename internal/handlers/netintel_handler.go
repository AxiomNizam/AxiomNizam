package handlers

import (
	"net/http"
	"strconv"

	"example.com/axiomnizam/internal/netintel"
	"github.com/gin-gonic/gin"
)

// ===================================================
// Network Intelligence Handler — Full REST API
// Parsers, logs, topology, heatmaps, predictions,
// anomalies, alerts, forecasts, trends, tracks
// ===================================================

type NetIntelHandler struct {
	parser    *netintel.ParserEngine
	analytics *netintel.AnalyticsEngine
	topology  *netintel.TopologyEngine
}

func NewNetIntelHandler() *NetIntelHandler {
	parser := netintel.NewParserEngine()
	analytics := netintel.NewAnalyticsEngine(parser)
	topo := netintel.NewTopologyEngine()
	return &NetIntelHandler{
		parser:    parser,
		analytics: analytics,
		topology:  topo,
	}
}

// ========================
// Summary / Observability
// ========================

// GetSummary GET /api/v1/netintel/summary
func (h *NetIntelHandler) GetSummary(c *gin.Context) {
	summary := h.analytics.GetSummary()
	c.JSON(http.StatusOK, gin.H{"status": "success", "summary": summary})
}

// GetObservability GET /api/v1/netintel/observability
func (h *NetIntelHandler) GetObservability(c *gin.Context) {
	stats := h.parser.GetEntryStats()
	c.JSON(http.StatusOK, gin.H{"status": "success", "observability": stats})
}

// ========================
// Log Types
// ========================

// GetLogTypes GET /api/v1/netintel/log-types
func (h *NetIntelHandler) GetLogTypes(c *gin.Context) {
	types := h.parser.GetLogTypes()
	c.JSON(http.StatusOK, gin.H{"status": "success", "log_types": types, "total": len(types)})
}

// ========================
// Parser CRUD
// ========================

// ListParsers GET /api/v1/netintel/parsers
func (h *NetIntelHandler) ListParsers(c *gin.Context) {
	parsers := h.parser.ListParsers()
	c.JSON(http.StatusOK, gin.H{"status": "success", "parsers": parsers, "total": len(parsers)})
}

// GetParser GET /api/v1/netintel/parsers/:id
func (h *NetIntelHandler) GetParser(c *gin.Context) {
	id := c.Param("id")
	p, ok := h.parser.GetParser(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "parser not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "parser": p})
}

// CreateParser POST /api/v1/netintel/parsers
func (h *NetIntelHandler) CreateParser(c *gin.Context) {
	var p netintel.ParserConfig
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if p.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "name is required"})
		return
	}
	if err := h.parser.CreateParser(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "parser": p})
}

// UpdateParser PUT /api/v1/netintel/parsers/:id
func (h *NetIntelHandler) UpdateParser(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if err := h.parser.UpdateParser(id, updates); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	p, _ := h.parser.GetParser(id)
	c.JSON(http.StatusOK, gin.H{"status": "success", "parser": p})
}

// DeleteParser DELETE /api/v1/netintel/parsers/:id
func (h *NetIntelHandler) DeleteParser(c *gin.Context) {
	id := c.Param("id")
	if err := h.parser.DeleteParser(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "parser deleted"})
}

// ========================
// Log Entries
// ========================

// ListEntries GET /api/v1/netintel/logs
func (h *NetIntelHandler) ListEntries(c *gin.Context) {
	logType := c.Query("type")
	severity := c.Query("severity")
	limit := 200
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 5000 {
			limit = v
		}
	}
	entries := h.parser.ListEntries(logType, limit, severity)
	c.JSON(http.StatusOK, gin.H{"status": "success", "entries": entries, "total": len(entries)})
}

// IngestLog POST /api/v1/netintel/logs
func (h *NetIntelHandler) IngestLog(c *gin.Context) {
	var entry netintel.ParsedEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if err := h.parser.IngestLog(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "log ingested"})
}

// GetEntryStats GET /api/v1/netintel/logs/stats
func (h *NetIntelHandler) GetEntryStats(c *gin.Context) {
	stats := h.parser.GetEntryStats()
	c.JSON(http.StatusOK, gin.H{"status": "success", "stats": stats})
}

// ========================
// Topology
// ========================

// GetTopology GET /api/v1/netintel/topology
func (h *NetIntelHandler) GetTopology(c *gin.Context) {
	graph := h.topology.GetGraph()
	c.JSON(http.StatusOK, gin.H{"status": "success", "topology": graph})
}

// GetTopologyNode GET /api/v1/netintel/topology/nodes/:id
func (h *NetIntelHandler) GetTopologyNode(c *gin.Context) {
	id := c.Param("id")
	node, ok := h.topology.GetNode(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "node not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "node": node})
}

// UpdateTopologyNode PUT /api/v1/netintel/topology/nodes/:id
func (h *NetIntelHandler) UpdateTopologyNode(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if err := h.topology.UpdateNodeStatus(id, body.Status); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "node updated"})
}

// ========================
// Heatmaps
// ========================

// GetHeatmap GET /api/v1/netintel/heatmap
func (h *NetIntelHandler) GetHeatmap(c *gin.Context) {
	category := c.DefaultQuery("category", "wifi_signal")
	heatmap := h.parser.GenerateHeatmap(category)
	c.JSON(http.StatusOK, gin.H{"status": "success", "heatmap": heatmap})
}

// ========================
// Trends
// ========================

// GetTrends GET /api/v1/netintel/trends
func (h *NetIntelHandler) GetTrends(c *gin.Context) {
	metric := c.DefaultQuery("metric", "traffic")
	hours := 24
	if h := c.Query("hours"); h != "" {
		if v, err := strconv.Atoi(h); err == nil && v > 0 && v <= 168 {
			hours = v
		}
	}
	points := h.parser.GetTrend(metric, hours)
	c.JSON(http.StatusOK, gin.H{"status": "success", "metric": metric, "hours": hours, "trend": points})
}

// ========================
// Predictions
// ========================

// GetPredictions GET /api/v1/netintel/predictions
func (h *NetIntelHandler) GetPredictions(c *gin.Context) {
	predictions := h.analytics.PredictMovement()
	c.JSON(http.StatusOK, gin.H{"status": "success", "predictions": predictions, "total": len(predictions)})
}

// ========================
// Movement Tracks
// ========================

// ListTracks GET /api/v1/netintel/tracks
func (h *NetIntelHandler) ListTracks(c *gin.Context) {
	tracks := h.parser.ListTracks()
	c.JSON(http.StatusOK, gin.H{"status": "success", "tracks": tracks, "total": len(tracks)})
}

// GetTrack GET /api/v1/netintel/tracks/:mac
func (h *NetIntelHandler) GetTrack(c *gin.Context) {
	mac := c.Param("mac")
	track, ok := h.parser.GetTrack(mac)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "track not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "track": track})
}

// ========================
// Anomalies
// ========================

// ListAnomalies GET /api/v1/netintel/anomalies
func (h *NetIntelHandler) ListAnomalies(c *gin.Context) {
	status := c.Query("status")
	severity := c.Query("severity")
	anomalies := h.analytics.ListAnomalies(status, severity)
	c.JSON(http.StatusOK, gin.H{"status": "success", "anomalies": anomalies, "total": len(anomalies)})
}

// AcknowledgeAnomaly POST /api/v1/netintel/anomalies/:id/acknowledge
func (h *NetIntelHandler) AcknowledgeAnomaly(c *gin.Context) {
	id := c.Param("id")
	if err := h.analytics.AcknowledgeAnomaly(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "anomaly acknowledged"})
}

// ResolveAnomaly POST /api/v1/netintel/anomalies/:id/resolve
func (h *NetIntelHandler) ResolveAnomaly(c *gin.Context) {
	id := c.Param("id")
	if err := h.analytics.ResolveAnomaly(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "anomaly resolved"})
}

// ========================
// Alerts
// ========================

// ListAlerts GET /api/v1/netintel/alerts
func (h *NetIntelHandler) ListAlerts(c *gin.Context) {
	status := c.Query("status")
	alerts := h.analytics.ListAlerts(status)
	c.JSON(http.StatusOK, gin.H{"status": "success", "alerts": alerts, "total": len(alerts)})
}

// AcknowledgeAlert POST /api/v1/netintel/alerts/:id/acknowledge
func (h *NetIntelHandler) AcknowledgeAlert(c *gin.Context) {
	id := c.Param("id")
	if err := h.analytics.AcknowledgeAlert(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "alert acknowledged"})
}

// ResolveAlert POST /api/v1/netintel/alerts/:id/resolve
func (h *NetIntelHandler) ResolveAlert(c *gin.Context) {
	id := c.Param("id")
	if err := h.analytics.ResolveAlert(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "alert resolved"})
}

// ========================
// Forecasts
// ========================

// ListForecasts GET /api/v1/netintel/forecasts
func (h *NetIntelHandler) ListForecasts(c *gin.Context) {
	forecasts := h.analytics.ListForecasts()
	c.JSON(http.StatusOK, gin.H{"status": "success", "forecasts": forecasts})
}

// GetForecast GET /api/v1/netintel/forecasts/:metric
func (h *NetIntelHandler) GetForecast(c *gin.Context) {
	metric := c.Param("metric")
	forecast, ok := h.analytics.GetForecast(metric)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "forecast not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "forecast": forecast})
}
