package netintel

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ===================================================
// Network Intelligence Handler — Full REST API
// Parsers, logs, topology, heatmaps, predictions,
// anomalies, alerts, forecasts, trends, tracks
// ===================================================

type Handler struct {
	parser    *ParserEngine
	analytics *AnalyticsEngine
	topology  *TopologyEngine
}

func NewHandler() *Handler {
	parser := NewParserEngine()
	analytics := NewAnalyticsEngine(parser)
	topo := NewTopologyEngine()
	return &Handler{
		parser:    parser,
		analytics: analytics,
		topology:  topo,
	}
}

// ========================
// Summary / Observability
// ========================

// GetSummary GET /api/v1/netintel/summary
func (h *Handler) GetSummary(c *gin.Context) {
	summary := h.analytics.GetSummary()
	c.JSON(http.StatusOK, SummaryResponse{Status: "success", Summary: summary})
}

// GetObservability GET /api/v1/netintel/observability
func (h *Handler) GetObservability(c *gin.Context) {
	stats := h.parser.GetEntryStats()
	c.JSON(http.StatusOK, ObservabilityResponse{Status: "success", Observability: stats})
}

// ========================
// Log Types
// ========================

// GetLogTypes GET /api/v1/netintel/log-types
func (h *Handler) GetLogTypes(c *gin.Context) {
	types := h.parser.GetLogTypes()
	c.JSON(http.StatusOK, LogTypesResponse{Status: "success", LogTypes: types, Total: len(types)})
}

// ========================
// Parser CRUD
// ========================

// ListParsers GET /api/v1/netintel/parsers
func (h *Handler) ListParsers(c *gin.Context) {
	parsers := h.parser.ListParsers()
	c.JSON(http.StatusOK, ParserListResponse{Status: "success", Parsers: parsers, Total: len(parsers)})
}

// GetParser GET /api/v1/netintel/parsers/:id
func (h *Handler) GetParser(c *gin.Context) {
	id := c.Param("id")
	p, ok := h.parser.GetParser(id)
	if !ok {
		c.JSON(http.StatusNotFound, MessageResponse{Error: "parser not found"})
		return
	}
	c.JSON(http.StatusOK, ParserResponse{Status: "success", Parser: p})
}

// CreateParser POST /api/v1/netintel/parsers
func (h *Handler) CreateParser(c *gin.Context) {
	var p ParserConfig
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: err.Error()})
		return
	}
	if p.Name == "" {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: "name is required"})
		return
	}
	if err := h.parser.CreateParser(&p); err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ParserResponse{Status: "success", Parser: p})
}

// UpdateParser PUT /api/v1/netintel/parsers/:id
func (h *Handler) UpdateParser(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: err.Error()})
		return
	}
	if err := h.parser.UpdateParser(id, updates); err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: err.Error()})
		return
	}
	p, _ := h.parser.GetParser(id)
	c.JSON(http.StatusOK, ParserResponse{Status: "success", Parser: p})
}

// DeleteParser DELETE /api/v1/netintel/parsers/:id
func (h *Handler) DeleteParser(c *gin.Context) {
	id := c.Param("id")
	if err := h.parser.DeleteParser(id); err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, MessageResponse{Message: "parser deleted"})
}

// ========================
// Log Entries
// ========================

// ListEntries GET /api/v1/netintel/logs
func (h *Handler) ListEntries(c *gin.Context) {
	logType := c.Query("type")
	severity := c.Query("severity")
	limit := 200
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 5000 {
			limit = v
		}
	}
	entries := h.parser.ListEntries(logType, limit, severity)
	c.JSON(http.StatusOK, EntryListResponse{Status: "success", Entries: entries, Total: len(entries)})
}

// IngestLog POST /api/v1/netintel/logs
func (h *Handler) IngestLog(c *gin.Context) {
	var entry ParsedEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: err.Error()})
		return
	}
	if err := h.parser.IngestLog(entry); err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, MessageResponse{Message: "log ingested"})
}

// GetEntryStats GET /api/v1/netintel/logs/stats
func (h *Handler) GetEntryStats(c *gin.Context) {
	stats := h.parser.GetEntryStats()
	c.JSON(http.StatusOK, StatsResponse{Status: "success", Stats: stats})
}

// ========================
// Topology
// ========================

// GetTopology GET /api/v1/netintel/topology
func (h *Handler) GetTopology(c *gin.Context) {
	graph := h.topology.GetGraph()
	c.JSON(http.StatusOK, TopologyResponse{Status: "success", Topology: graph})
}

// GetTopologyNode GET /api/v1/netintel/topology/nodes/:id
func (h *Handler) GetTopologyNode(c *gin.Context) {
	id := c.Param("id")
	node, ok := h.topology.GetNode(id)
	if !ok {
		c.JSON(http.StatusNotFound, MessageResponse{Error: "node not found"})
		return
	}
	c.JSON(http.StatusOK, TopologyNodeResponse{Status: "success", Node: node})
}

// UpdateTopologyNode PUT /api/v1/netintel/topology/nodes/:id
func (h *Handler) UpdateTopologyNode(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: err.Error()})
		return
	}
	if err := h.topology.UpdateNodeStatus(id, body.Status); err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, MessageResponse{Message: "node updated"})
}

// ========================
// Heatmaps
// ========================

// GetHeatmap GET /api/v1/netintel/heatmap
func (h *Handler) GetHeatmap(c *gin.Context) {
	category := c.DefaultQuery("category", "wifi_signal")
	heatmap := h.parser.GenerateHeatmap(category)
	c.JSON(http.StatusOK, HeatmapResponse{Status: "success", Heatmap: heatmap})
}

// ========================
// Trends
// ========================

// GetTrends GET /api/v1/netintel/trends
func (h *Handler) GetTrends(c *gin.Context) {
	metric := c.DefaultQuery("metric", "traffic")
	hours := 24
	if h := c.Query("hours"); h != "" {
		if v, err := strconv.Atoi(h); err == nil && v > 0 && v <= 168 {
			hours = v
		}
	}
	points := h.parser.GetTrend(metric, hours)
	c.JSON(http.StatusOK, TrendsResponse{Status: "success", Metric: metric, Hours: hours, Trend: points})
}

// ========================
// Predictions
// ========================

// GetPredictions GET /api/v1/netintel/predictions
func (h *Handler) GetPredictions(c *gin.Context) {
	predictions := h.analytics.PredictMovement()
	c.JSON(http.StatusOK, PredictionsResponse{Status: "success", Predictions: predictions, Total: len(predictions)})
}

// ========================
// Movement Tracks
// ========================

// ListTracks GET /api/v1/netintel/tracks
func (h *Handler) ListTracks(c *gin.Context) {
	tracks := h.parser.ListTracks()
	c.JSON(http.StatusOK, TrackListResponse{Status: "success", Tracks: tracks, Total: len(tracks)})
}

// GetTrack GET /api/v1/netintel/tracks/:mac
func (h *Handler) GetTrack(c *gin.Context) {
	mac := c.Param("mac")
	track, ok := h.parser.GetTrack(mac)
	if !ok {
		c.JSON(http.StatusNotFound, MessageResponse{Error: "track not found"})
		return
	}
	c.JSON(http.StatusOK, TrackResponse{Status: "success", Track: track})
}

// ========================
// Anomalies
// ========================

// ListAnomalies GET /api/v1/netintel/anomalies
func (h *Handler) ListAnomalies(c *gin.Context) {
	status := c.Query("status")
	severity := c.Query("severity")
	anomalies := h.analytics.ListAnomalies(status, severity)
	c.JSON(http.StatusOK, AnomalyListResponse{Status: "success", Anomalies: anomalies, Total: len(anomalies)})
}

// AcknowledgeAnomaly POST /api/v1/netintel/anomalies/:id/acknowledge
func (h *Handler) AcknowledgeAnomaly(c *gin.Context) {
	id := c.Param("id")
	if err := h.analytics.AcknowledgeAnomaly(id); err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, MessageResponse{Message: "anomaly acknowledged"})
}

// ResolveAnomaly POST /api/v1/netintel/anomalies/:id/resolve
func (h *Handler) ResolveAnomaly(c *gin.Context) {
	id := c.Param("id")
	if err := h.analytics.ResolveAnomaly(id); err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, MessageResponse{Message: "anomaly resolved"})
}

// ========================
// Alerts
// ========================

// ListAlerts GET /api/v1/netintel/alerts
func (h *Handler) ListAlerts(c *gin.Context) {
	status := c.Query("status")
	alerts := h.analytics.ListAlerts(status)
	c.JSON(http.StatusOK, AlertListResponse{Status: "success", Alerts: alerts, Total: len(alerts)})
}

// AcknowledgeAlert POST /api/v1/netintel/alerts/:id/acknowledge
func (h *Handler) AcknowledgeAlert(c *gin.Context) {
	id := c.Param("id")
	if err := h.analytics.AcknowledgeAlert(id); err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, MessageResponse{Message: "alert acknowledged"})
}

// ResolveAlert POST /api/v1/netintel/alerts/:id/resolve
func (h *Handler) ResolveAlert(c *gin.Context) {
	id := c.Param("id")
	if err := h.analytics.ResolveAlert(id); err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, MessageResponse{Message: "alert resolved"})
}

// ========================
// Forecasts
// ========================

// ListForecasts GET /api/v1/netintel/forecasts
func (h *Handler) ListForecasts(c *gin.Context) {
	forecasts := h.analytics.ListForecasts()
	c.JSON(http.StatusOK, ForecastListResponse{Status: "success", Forecasts: forecasts})
}

// GetForecast GET /api/v1/netintel/forecasts/:metric
func (h *Handler) GetForecast(c *gin.Context) {
	metric := c.Param("metric")
	forecast, ok := h.analytics.GetForecast(metric)
	if !ok {
		c.JSON(http.StatusNotFound, MessageResponse{Error: "forecast not found"})
		return
	}
	c.JSON(http.StatusOK, ForecastResponse{Status: "success", Forecast: forecast})
}
