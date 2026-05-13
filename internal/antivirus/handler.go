package antivirus

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────────────────────────────────
// API Handler
//
// Provides HTTP endpoints for antivirus engine management. These are
// designed to be mounted into the storage admin router group.
//
// Endpoints:
//
//	GET  /antivirus/status   — engine status + signature DB version + layer info
//	GET  /antivirus/stats    — scan statistics (totals, cache hit rate, avg scan time)
//	POST /antivirus/scan     — manual scan of raw content (multipart upload)
//	GET  /antivirus/threats  — list recent threat detections from audit log
//	POST /antivirus/update   — trigger immediate signature database update
//	PUT  /antivirus/config   — view current engine configuration
// ─────────────────────────────────────────────────────────────────────────────

// ThreatRecord is a compact threat entry returned by the /threats endpoint.
type ThreatRecord struct {
	Filename  string      `json:"filename"`
	SHA256    string      `json:"sha256"`
	Verdict   ScanVerdict `json:"verdict"`
	Threats   []string    `json:"threats"`
	Severity  string      `json:"severity"`
	ScannedAt time.Time   `json:"scannedAt"`
	DurationMs int64     `json:"durationMs"`
}

// APIHandler exposes antivirus management endpoints. It references the
// Engine directly and is embedded into the storage admin handler.
type APIHandler struct {
	engine *Engine
}

// NewAPIHandler creates a new antivirus API handler.
func NewAPIHandler(engine *Engine) *APIHandler {
	return &APIHandler{engine: engine}
}

// RegisterRoutes mounts antivirus routes under the given router group.
// Typically called as: handler.RegisterRoutes(adminGroup.Group("/antivirus"))
func (h *APIHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if h == nil || h.engine == nil {
		return
	}

	av := rg.Group("/antivirus")
	{
		av.GET("/status", h.Status)
		av.GET("/stats", h.ScanStats)
		av.POST("/scan", h.ManualScan)
		av.GET("/threats", h.ListThreats)
		av.GET("/config", h.GetConfig)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /antivirus/status
// ─────────────────────────────────────────────────────────────────────────────

// Status returns the current engine status, including whether it's running,
// the signature database version, enabled layers, and uptime.
func (h *APIHandler) Status(c *gin.Context) {
	stats := h.engine.Stats()

	avStatus := "disabled"
	if h.engine.IsRunning() {
		avStatus = "running"
	} else if h.engine.cfg.Enabled {
		avStatus = "stopped"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        avStatus,
		"engineVersion": EngineVersion,
		"sigDbVersion":  stats.SigDBVersion,
		"layersEnabled": stats.LayersEnabled,
		"layerCount":    len(stats.LayersEnabled),
		"uptimeSeconds": stats.UptimeSeconds,
		"scanCapacity": gin.H{
			"workers":     h.engine.cfg.Workers,
			"queueSize":   h.engine.cfg.QueueSize,
			"maxFileSize": h.engine.cfg.MaxFileSize,
		},
		"features": gin.H{
			"hashDB":    h.engine.cfg.HashDBEnabled,
			"pattern":   h.engine.cfg.PatternEnabled,
			"heuristic": h.engine.cfg.HeuristicEnabled,
			"entropy":   h.engine.cfg.EntropyEnabled,
			"yara":      h.engine.cfg.YARAEnabled,
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /antivirus/stats
// ─────────────────────────────────────────────────────────────────────────────

// ScanStats returns scan performance statistics including totals, cache
// performance, and average scan times.
func (h *APIHandler) ScanStats(c *gin.Context) {
	stats := h.engine.Stats()

	c.JSON(http.StatusOK, gin.H{
		"totalScanned":  stats.TotalScanned,
		"threatsFound":  stats.ThreatsFound,
		"cleanFiles":    stats.CleanFiles,
		"errorCount":    stats.ErrorCount,
		"bytesScanned":  stats.BytesScanned,
		"avgScanMs":     fmt.Sprintf("%.2f", stats.AvgScanMs),
		"cache": gin.H{
			"hits":    stats.CacheHits,
			"misses":  stats.CacheMisses,
			"hitRate": fmt.Sprintf("%.4f", stats.CacheHitRate),
		},
		"uptimeSeconds": stats.UptimeSeconds,
		"engineVersion": stats.EngineVersion,
		"sigDbVersion":  stats.SigDBVersion,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /antivirus/scan
// ─────────────────────────────────────────────────────────────────────────────

// ManualScan allows manual scanning of uploaded content. Accepts a multipart
// file upload with field name "file" and returns the full scan result.
//
// Request: multipart/form-data with "file" field
// Response: ScanResult JSON
func (h *APIHandler) ManualScan(c *gin.Context) {
	if !h.engine.IsRunning() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "antivirus engine is not running",
		})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file upload required (form field: 'file'): " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Read content with size limit.
	maxSize := h.engine.MaxFileSize()
	content, err := io.ReadAll(io.LimitReader(file, maxSize+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read uploaded file: " + err.Error(),
		})
		return
	}
	if int64(len(content)) > maxSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error":   "file exceeds maximum scan size",
			"maxSize": maxSize,
		})
		return
	}

	filename := header.Filename
	if filename == "" {
		filename = "manual-upload"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	result, err := h.engine.Scan(ctx, content, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "scan failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /antivirus/threats
// ─────────────────────────────────────────────────────────────────────────────

// ListThreats returns recent threat detections. It searches the engine's
// internal threat log. Supports ?limit=N query parameter (default: 50).
func (h *APIHandler) ListThreats(c *gin.Context) {
	threats := h.engine.RecentThreats()

	// Convert to compact format.
	records := make([]ThreatRecord, 0, len(threats))
	for _, r := range threats {
		names := make([]string, 0, len(r.Threats))
		for _, t := range r.Threats {
			names = append(names, t.Name)
		}

		severity := "unknown"
		if hs := r.HighestSeverity(); hs != "" {
			severity = string(hs)
		}

		records = append(records, ThreatRecord{
			Filename:   r.Filename,
			SHA256:     r.SHA256,
			Verdict:    r.Verdict,
			Threats:    names,
			Severity:   severity,
			ScannedAt:  r.ScannedAt,
			DurationMs: r.DurationMs,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"threats": records,
		"count":   len(records),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /antivirus/config
// ─────────────────────────────────────────────────────────────────────────────

// GetConfig returns the current engine configuration (read-only view).
func (h *APIHandler) GetConfig(c *gin.Context) {
	cfg := h.engine.cfg

	c.JSON(http.StatusOK, gin.H{
		"enabled":          cfg.Enabled,
		"workers":          cfg.Workers,
		"queueSize":        cfg.QueueSize,
		"maxFileSize":      cfg.MaxFileSize,
		"cacheSize":        cfg.CacheSize,
		"cacheTTL":         cfg.CacheTTL.String(),
		"updateURL":        redactURL(cfg.UpdateURL),
		"updateInterval":   cfg.UpdateInterval.String(),
		"sigDir":           cfg.SigDir,
		"quarantineAction": cfg.QuarantineAction,
		"layers": gin.H{
			"hashDB":    cfg.HashDBEnabled,
			"pattern":   cfg.PatternEnabled,
			"heuristic": cfg.HeuristicEnabled,
			"entropy":   cfg.EntropyEnabled,
			"yara":      cfg.YARAEnabled,
		},
	})
}

// redactURL masks sensitive parts of URLs for safe display.
func redactURL(url string) string {
	if url == "" {
		return "(not configured)"
	}
	// Mask anything after :// and before the first /path segment.
	if idx := strings.Index(url, "://"); idx >= 0 {
		rest := url[idx+3:]
		if slashIdx := strings.Index(rest, "/"); slashIdx > 0 {
			return url[:idx+3] + "***" + rest[slashIdx:]
		}
	}
	return url
}
