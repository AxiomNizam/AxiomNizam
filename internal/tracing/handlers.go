package tracing

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TracingHandler handles tracing endpoints
type TracingHandler struct {
	manager TracingManager
}

// NewTracingHandler creates handler
func NewTracingHandler(manager TracingManager) *TracingHandler {
	return &TracingHandler{manager: manager}
}

// GetTrace handles GET /api/v1/traces/:traceId
func (h *TracingHandler) GetTrace(c *gin.Context) {
	traceID := c.Param("traceId")
	trace, err := h.manager.GetTrace(traceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trace not found"})
		return
	}

	c.JSON(http.StatusOK, trace)
}

// SearchTraces handles GET /api/v1/traces/search
func (h *TracingHandler) SearchTraces(c *gin.Context) {
	var req TraceSearchRequest
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.manager.SearchTraces(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"traces": results, "count": len(results)})
}

// GetSpan handles GET /api/v1/spans/:spanId
func (h *TracingHandler) GetSpan(c *gin.Context) {
	spanID := c.Param("spanId")
	span, err := h.manager.GetSpan(spanID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "span not found"})
		return
	}

	c.JSON(http.StatusOK, span)
}

// GetServiceMap handles GET /api/v1/service-map
func (h *TracingHandler) GetServiceMap(c *gin.Context) {
	tenantID := c.Query("tenantId")
	serviceMap, err := h.manager.GetServiceMap(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, serviceMap)
}

// GetServiceMetrics handles GET /api/v1/services/:service/metrics
func (h *TracingHandler) GetServiceMetrics(c *gin.Context) {
	service := c.Param("service")
	metrics, err := h.manager.GetServiceMetrics(service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetOperationMetrics handles GET /api/v1/services/:service/operations/:operation/metrics
func (h *TracingHandler) GetOperationMetrics(c *gin.Context) {
	service := c.Param("service")
	operation := c.Param("operation")

	metrics, err := h.manager.GetOperationMetrics(service, operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// ListServices handles GET /api/v1/services
func (h *TracingHandler) ListServices(c *gin.Context) {
	tenantID := c.Query("tenantId")
	services, err := h.manager.ListServices(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"services": services, "count": len(services)})
}

// GetErrorAnalysis handles GET /api/v1/errors/analysis
func (h *TracingHandler) GetErrorAnalysis(c *gin.Context) {
	tenantID := c.Query("tenantId")
	service := c.Query("service")

	analysis, err := h.manager.GetErrorAnalysis(tenantID, service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// RegisterTracingRoutes registers all tracing routes
func RegisterTracingRoutes(router *gin.Engine, manager TracingManager) {
	handler := NewTracingHandler(manager)

	group := router.Group("/api/v1")
	{
		group.GET("/traces/:traceId", handler.GetTrace)
		group.GET("/traces/search", handler.SearchTraces)
		group.GET("/spans/:spanId", handler.GetSpan)
		group.GET("/service-map", handler.GetServiceMap)
		group.GET("/services", handler.ListServices)
		group.GET("/services/:service/metrics", handler.GetServiceMetrics)
		group.GET("/services/:service/operations/:operation/metrics", handler.GetOperationMetrics)
		group.GET("/errors/analysis", handler.GetErrorAnalysis)
	}
}

// TracingManager interface
type TracingManager interface {
	GetTrace(traceID string) (*Trace, error)
	SearchTraces(req *TraceSearchRequest) ([]*TraceSearchResult, error)
	GetSpan(spanID string) (*Span, error)
	GetServiceMap(tenantID string) (*ServiceMap, error)
	GetServiceMetrics(service string) (*TraceMetrics, error)
	GetOperationMetrics(service, operation string) (*SpanMetrics, error)
	ListServices(tenantID string) ([]*ServiceInfo, error)
	GetErrorAnalysis(tenantID, service string) (*ErrorAnalysis, error)
}
