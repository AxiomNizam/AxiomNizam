package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// APIMetric represents metrics for a single API endpoint
type APIMetric struct {
	Endpoint        string        `json:"endpoint"`
	Method          string        `json:"method"`
	TotalCalls      int64         `json:"total_calls"`
	SuccessCalls    int64         `json:"success_calls"`
	ErrorCalls      int64         `json:"error_calls"`
	AverageDuration int64         `json:"average_duration_ms"`
	MaxDuration     int64         `json:"max_duration_ms"`
	MinDuration     int64         `json:"min_duration_ms"`
	LastCalled      string        `json:"last_called"`
	StatusCodes     map[int]int64 `json:"status_codes"`
}

// APIMetricsTracker tracks API usage and call counts
type APIMetricsTracker struct {
	redisClient  *redis.Client
	localMetrics map[string]*APIMetric
	mu           sync.RWMutex
}

// NewAPIMetricsTracker creates a new API metrics tracker
func NewAPIMetricsTracker(redisClient *redis.Client) *APIMetricsTracker {
	return &APIMetricsTracker{
		redisClient:  redisClient,
		localMetrics: make(map[string]*APIMetric),
	}
}

// RecordAPICall records a call to an API endpoint
func (t *APIMetricsTracker) RecordAPICall(method, endpoint string, statusCode int, duration time.Duration) {
	if t == nil {
		return
	}

	key := fmt.Sprintf("api_metric:%s:%s", method, endpoint)
	durationMs := duration.Milliseconds()

	if t.redisClient != nil {
		// Use a short timeout so unreachable Redis doesn't block for minutes.
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		// Increment total calls
		t.redisClient.Incr(ctx, fmt.Sprintf("%s:calls", key))

		// Track status code
		t.redisClient.Incr(ctx, fmt.Sprintf("%s:status:%d", key, statusCode))

		// Track duration
		t.redisClient.Incr(ctx, fmt.Sprintf("%s:total_duration", key))
		t.redisClient.Set(ctx, fmt.Sprintf("%s:duration", key), durationMs, 0)

		// Update max duration
		maxKey := fmt.Sprintf("%s:max_duration", key)
		currentMax := t.redisClient.Get(ctx, maxKey).Val()
		if currentMax == "" {
			t.redisClient.Set(ctx, maxKey, durationMs, 0)
		} else {
			var currentMaxMs int64
			fmt.Sscanf(currentMax, "%d", &currentMaxMs)
			if durationMs > currentMaxMs {
				t.redisClient.Set(ctx, maxKey, durationMs, 0)
			}
		}

		// Update min duration
		minKey := fmt.Sprintf("%s:min_duration", key)
		currentMin := t.redisClient.Get(ctx, minKey).Val()
		if currentMin == "" {
			t.redisClient.Set(ctx, minKey, durationMs, 0)
		} else {
			var currentMinMs int64
			fmt.Sscanf(currentMin, "%d", &currentMinMs)
			if durationMs < currentMinMs {
				t.redisClient.Set(ctx, minKey, durationMs, 0)
			}
		}

		// Update last called timestamp
		t.redisClient.Set(ctx, fmt.Sprintf("%s:last_called", key), time.Now().Format(time.RFC3339), 0)
	}

	// Track in local cache for quick access
	t.mu.Lock()
	if metric, exists := t.localMetrics[key]; exists {
		metric.TotalCalls++
		if statusCode >= 200 && statusCode < 300 {
			metric.SuccessCalls++
		} else if statusCode >= 400 {
			metric.ErrorCalls++
		}
		metric.LastCalled = time.Now().Format(time.RFC3339)
		if metric.StatusCodes == nil {
			metric.StatusCodes = make(map[int]int64)
		}
		metric.StatusCodes[statusCode]++
	} else {
		statusCodes := make(map[int]int64)
		statusCodes[statusCode] = 1
		errorCount := int64(0)
		successCount := int64(1)
		if statusCode >= 200 && statusCode < 300 {
			successCount = 1
		} else if statusCode >= 400 {
			errorCount = 1
			successCount = 0
		}
		t.localMetrics[key] = &APIMetric{
			Endpoint:     endpoint,
			Method:       method,
			TotalCalls:   1,
			SuccessCalls: successCount,
			ErrorCalls:   errorCount,
			LastCalled:   time.Now().Format(time.RFC3339),
			StatusCodes:  statusCodes,
		}
	}
	t.mu.Unlock()
}

func (t *APIMetricsTracker) localMetricsSnapshot() []APIMetric {
	if t == nil {
		return []APIMetric{}
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	metrics := make([]APIMetric, 0, len(t.localMetrics))
	for _, m := range t.localMetrics {
		metrics = append(metrics, *m)
	}

	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].Endpoint == metrics[j].Endpoint {
			return metrics[i].Method < metrics[j].Method
		}
		return metrics[i].Endpoint < metrics[j].Endpoint
	})

	return metrics
}

// GetAllMetrics retrieves all API metrics
func (t *APIMetricsTracker) GetAllMetrics() ([]APIMetric, error) {
	if t == nil {
		return []APIMetric{}, nil
	}

	if t.redisClient == nil {
		return t.localMetricsSnapshot(), nil
	}

	// Get all API metric keys from Redis
	ctx := context.Background()
	keys, err := t.redisClient.Keys(ctx, "api_metric:*:calls").Result()
	if err != nil {
		return t.localMetricsSnapshot(), nil
	}
	if len(keys) == 0 {
		return t.localMetricsSnapshot(), nil
	}

	metrics := make([]APIMetric, 0, len(keys))
	metricMap := make(map[string]*APIMetric)

	// Process each key
	for _, key := range keys {
		// Extract method and endpoint from key
		// Format: api_metric:METHOD:endpoint:calls
		var method, endpoint string
		fmt.Sscanf(key, "api_metric:%s:%s:calls", &method, &endpoint)

		baseKey := fmt.Sprintf("api_metric:%s:%s", method, endpoint)
		metricKey := fmt.Sprintf("%s:%s", method, endpoint)

		if _, exists := metricMap[metricKey]; !exists {
			// Get total calls
			totalCallsStr := t.redisClient.Get(ctx, fmt.Sprintf("%s:calls", baseKey)).Val()
			var totalCalls int64
			if totalCallsStr != "" {
				fmt.Sscanf(totalCallsStr, "%d", &totalCalls)
			}

			// Get status codes
			statusKeys, _ := t.redisClient.Keys(ctx, fmt.Sprintf("%s:status:*", baseKey)).Result()
			statusCodes := make(map[int]int64)
			var successCalls, errorCalls int64

			for _, statusKey := range statusKeys {
				var statusCode int
				fmt.Sscanf(statusKey, fmt.Sprintf("%s:status:%%d", baseKey), &statusCode)
				countStr := t.redisClient.Get(ctx, statusKey).Val()
				var count int64
				if countStr != "" {
					fmt.Sscanf(countStr, "%d", &count)
				}
				statusCodes[statusCode] = count

				if statusCode >= 200 && statusCode < 300 {
					successCalls += count
				} else if statusCode >= 400 {
					errorCalls += count
				}
			}

			// Get durations
			maxDurationStr := t.redisClient.Get(ctx, fmt.Sprintf("%s:max_duration", baseKey)).Val()
			minDurationStr := t.redisClient.Get(ctx, fmt.Sprintf("%s:min_duration", baseKey)).Val()
			var maxDuration, minDuration int64

			if maxDurationStr != "" {
				fmt.Sscanf(maxDurationStr, "%d", &maxDuration)
			}
			if minDurationStr != "" {
				fmt.Sscanf(minDurationStr, "%d", &minDuration)
			}

			lastCalled := t.redisClient.Get(ctx, fmt.Sprintf("%s:last_called", baseKey)).Val()

			metric := &APIMetric{
				Endpoint:     endpoint,
				Method:       method,
				TotalCalls:   totalCalls,
				SuccessCalls: successCalls,
				ErrorCalls:   errorCalls,
				MaxDuration:  maxDuration,
				MinDuration:  minDuration,
				LastCalled:   lastCalled,
				StatusCodes:  statusCodes,
			}

			if totalCalls > 0 {
				metric.AverageDuration = (maxDuration + minDuration) / 2
			}

			metricMap[metricKey] = metric
		}
	}

	for _, m := range metricMap {
		metrics = append(metrics, *m)
	}

	// Sort by endpoint name
	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].Endpoint == metrics[j].Endpoint {
			return metrics[i].Method < metrics[j].Method
		}
		return metrics[i].Endpoint < metrics[j].Endpoint
	})

	return metrics, nil
}

// GetMetricsByEndpoint retrieves metrics for a specific endpoint
func (t *APIMetricsTracker) GetMetricsByEndpoint(endpoint string) ([]APIMetric, error) {
	metrics, err := t.GetAllMetrics()
	if err != nil {
		return nil, err
	}

	filtered := make([]APIMetric, 0)
	for _, m := range metrics {
		if m.Endpoint == endpoint {
			filtered = append(filtered, m)
		}
	}

	return filtered, nil
}

// GetAPICountValue returns the total count of unique API endpoints
func (t *APIMetricsTracker) GetAPICountValue() (int, error) {
	metrics, err := t.GetAllMetrics()
	if err != nil {
		return 0, err
	}

	// Count unique endpoints
	endpointMap := make(map[string]bool)
	for _, m := range metrics {
		endpointMap[m.Endpoint] = true
	}

	return len(endpointMap), nil
}

// GetEndpointUsage returns all endpoints with their usage count
func (t *APIMetricsTracker) GetEndpointUsage() (map[string]int64, error) {
	metrics, err := t.GetAllMetrics()
	if err != nil {
		return nil, err
	}

	usage := make(map[string]int64)
	for _, m := range metrics {
		usage[m.Endpoint] += m.TotalCalls
	}

	return usage, nil
}

// Handler methods for API endpoints

// GetAllAPIMetrics returns all API metrics
func (t *APIMetricsTracker) GetAllAPIMetrics(c *gin.Context) {
	metrics, err := t.GetAllMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	apiCount, _ := t.GetAPICountValue()
	usage, _ := t.GetEndpointUsage()

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"total_unique_endpoints": apiCount,
			"total_calls": func() int64 {
				var total int64
				for _, m := range metrics {
					total += m.TotalCalls
				}
				return total
			}(),
			"endpoints":      metrics,
			"endpoint_usage": usage,
		},
	})
}

// GetAPICount returns the count of APIs and their usage
func (t *APIMetricsTracker) GetAPICount(c *gin.Context) {
	apiCount, err := t.GetAPICountValue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	usage, _ := t.GetEndpointUsage()
	var totalCalls int64
	for _, count := range usage {
		totalCalls += count
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"total_unique_endpoints": apiCount,
			"total_api_calls":        totalCalls,
			"endpoint_usage":         usage,
		},
	})
}

// GetAPIStats returns detailed API statistics
func (t *APIMetricsTracker) GetAPIStats(c *gin.Context) {
	endpoint := c.Query("endpoint")
	var metrics []APIMetric
	var err error

	if endpoint != "" {
		metrics, err = t.GetMetricsByEndpoint(endpoint)
	} else {
		metrics, err = t.GetAllMetrics()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	// Calculate aggregates
	var totalCalls, successCalls, errorCalls int64
	var totalDuration, count int64
	avgDuration := int64(0)

	for _, m := range metrics {
		totalCalls += m.TotalCalls
		successCalls += m.SuccessCalls
		errorCalls += m.ErrorCalls
		totalDuration += m.AverageDuration
		count++
	}

	if count > 0 {
		avgDuration = totalDuration / count
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"total_calls":         totalCalls,
			"success_calls":       successCalls,
			"error_calls":         errorCalls,
			"average_duration_ms": avgDuration,
			"metrics":             metrics,
		},
	})
}

// MetricsMiddleware middleware to track API calls
func MetricsMiddleware(tracker *APIMetricsTracker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tracker == nil {
			c.Next()
			return
		}

		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)
		method := c.Request.Method
		endpoint := c.Request.URL.Path
		statusCode := c.Writer.Status()

		// Record the metric asynchronously so that slow/unreachable
		// Redis connections do not block the HTTP response.
		go tracker.RecordAPICall(method, endpoint, statusCode, duration)
	}
}
