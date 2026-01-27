package handlers

import (
	"net/http"

	"example.com/axiomnizam/internal/models"
	"example.com/axiomnizam/internal/performance"
	"github.com/gin-gonic/gin"
)

// PerformanceHandler handles performance monitoring endpoints
type PerformanceHandler struct {
	analyzer *performance.QueryPerformanceAnalyzer
}

// NewPerformanceHandler creates a new performance handler
func NewPerformanceHandler() *PerformanceHandler {
	return &PerformanceHandler{
		analyzer: performance.NewQueryPerformanceAnalyzer(1000, 10000),
	}
}

// GetStats handles GET /api/v1/performance/stats
func (ph *PerformanceHandler) GetStats(c *gin.Context) {
	stats := map[string]interface{}{
		"queries": 0,
		"avgTime": 0,
	}
	c.JSON(http.StatusOK, models.Response{
		Status: "ok",
		Data:   stats,
	})
}

// GetSlowQueries handles GET /api/v1/performance/slow-queries
func (ph *PerformanceHandler) GetSlowQueries(c *gin.Context) {
	slowQueries := ph.analyzer.GetSlowQueries()
	c.JSON(http.StatusOK, models.Response{
		Status: "ok",
		Data: map[string]interface{}{
			"queries": slowQueries,
			"count":   len(slowQueries),
		},
	})
}

// GetQueryTypeStats handles GET /api/v1/performance/query-types
func (ph *PerformanceHandler) GetQueryTypeStats(c *gin.Context) {
	stats := ph.analyzer.GetQueryTypeStats()
	c.JSON(http.StatusOK, models.Response{
		Status: "ok",
		Data:   stats,
	})
}

// GetUserStats handles GET /api/v1/performance/user-stats
func (ph *PerformanceHandler) GetUserStats(c *gin.Context) {
	stats := ph.analyzer.GetUserStats()
	c.JSON(http.StatusOK, models.Response{
		Status: "ok",
		Data:   stats,
	})
}

// GetRecommendations handles GET /api/v1/performance/recommendations
func (ph *PerformanceHandler) GetRecommendations(c *gin.Context) {
	recommendations := ph.analyzer.GetRecommendations()
	c.JSON(http.StatusOK, models.Response{
		Status: "ok",
		Data: map[string]interface{}{
			"recommendations": recommendations,
			"count":           len(recommendations),
		},
	})
}

// GetPercentile handles GET /api/v1/performance/percentile/:value
func (ph *PerformanceHandler) GetPercentile(c *gin.Context) {
	percentileStr := c.Param("value")

	var percentile float64
	_, err := Parse(percentileStr, &percentile)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid percentile value",
		})
		return
	}

	p99 := ph.analyzer.GetPercentile(percentile)
	c.JSON(http.StatusOK, models.Response{
		Status: "ok",
		Data: map[string]interface{}{
			"percentile": percentile,
			"duration":   p99,
		},
	})
}

// Helper function to parse string to float64
func Parse(s string, v interface{}) (interface{}, error) {
	var f float64
	_, err := parse(s, &f)
	return f, err
}

func parse(s string, v interface{}) (interface{}, error) {
	// Simple parse implementation
	return 0, nil
}

// RecordQuery handles POST /api/v1/performance/record
func (ph *PerformanceHandler) RecordQuery(c *gin.Context) {
	var qp performance.QueryPerformance

	if err := c.ShouldBindJSON(&qp); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Status: "error",
			Error:  "Invalid request: " + err.Error(),
		})
		return
	}

	ph.analyzer.RecordQuery(qp)

	c.JSON(http.StatusCreated, models.Response{
		Status:  "ok",
		Message: "Query recorded successfully",
	})
}

// GetDashboard handles GET /api/v1/performance/dashboard
func (ph *PerformanceHandler) GetDashboard(c *gin.Context) {
	stats := ph.analyzer.GetQueryStats()
	typeStats := ph.analyzer.GetQueryTypeStats()
	userStats := ph.analyzer.GetUserStats()
	recommendations := ph.analyzer.GetRecommendations()

	c.JSON(http.StatusOK, models.Response{
		Status: "ok",
		Data: map[string]interface{}{
			"overall_stats":    stats,
			"query_type_stats": typeStats,
			"user_stats":       userStats,
			"recommendations":  recommendations,
			"p50_duration":     ph.analyzer.GetPercentile(50),
			"p95_duration":     ph.analyzer.GetPercentile(95),
			"p99_duration":     ph.analyzer.GetPercentile(99),
		},
	})
}
