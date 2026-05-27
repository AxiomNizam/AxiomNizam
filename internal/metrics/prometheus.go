package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterCollector registers a Prometheus collector with the default registry.
// Call this from per-module metrics/ packages to wire collectors into /metrics.
// Note: modules using promauto already auto-register with the default registry,
// so this is only needed for manually-created collectors.
func RegisterCollector(c prometheus.Collector) {
	prometheus.MustRegister(c)
}

// MetricsHandler returns a Gin handler that serves Prometheus metrics at /metrics.
// All metrics registered with promauto or RegisterCollector are included.
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer,
		promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		}),
	)
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// RegisterMetricsEndpoint registers GET /metrics on the given router.
func RegisterMetricsEndpoint(r *gin.Engine) {
	r.GET("/metrics", MetricsHandler())
}
