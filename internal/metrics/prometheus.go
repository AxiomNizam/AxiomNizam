package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Registry is the application-wide Prometheus registry.
	// All per-module collectors should register here.
	Registry = prometheus.NewRegistry()

	// Gatherer exposes the default gatherer for the /metrics handler.
	Gatherer prometheus.Gatherer = Registry
)

func init() {
	// Register Go runtime metrics (goroutines, memory, GC).
	Registry.MustRegister(prometheus.NewGoCollector())
	Registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}

// RegisterCollector registers a Prometheus collector with the application registry.
// Call this from per-module metrics/ packages to wire collectors into /metrics.
func RegisterCollector(c prometheus.Collector) {
	Registry.MustRegister(c)
}

// MetricsHandler returns a Gin handler that serves Prometheus metrics at /metrics.
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.HandlerFor(Gatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// RegisterMetricsEndpoint registers GET /metrics on the given router.
func RegisterMetricsEndpoint(r *gin.Engine) {
	r.GET("/metrics", MetricsHandler())
}
