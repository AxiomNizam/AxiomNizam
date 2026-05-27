package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ─────────────────────────────────────────────────────────────────────────────
// Counter metrics
// ─────────────────────────────────────────────────────────────────────────────

var (
	ScansTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_antivirus",
		Name:      "scans_total",
		Help:      "Total number of file scans performed",
	})

	ScansClean = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_antivirus",
		Name:      "scans_clean_total",
		Help:      "Total number of scans with clean verdict",
	})

	ScansMalware = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_antivirus",
		Name:      "scans_malware_total",
		Help:      "Total number of scans with malware verdict",
	})

	ScansSuspicious = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_antivirus",
		Name:      "scans_suspicious_total",
		Help:      "Total number of scans with suspicious verdict",
	})

	ScansError = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_antivirus",
		Name:      "scans_error_total",
		Help:      "Total number of scans that ended in error",
	})

	ThreatsDetected = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_antivirus",
		Name:      "threats_detected_total",
		Help:      "Total number of threats detected across all scans",
	})

	CacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_antivirus",
		Name:      "cache_hits_total",
		Help:      "Total number of scan cache hits",
	})

	CacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_antivirus",
		Name:      "cache_misses_total",
		Help:      "Total number of scan cache misses",
	})

	BytesScanned = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "axiom_antivirus",
		Name:      "bytes_scanned_total",
		Help:      "Total bytes scanned",
	})

	LayerErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "axiom_antivirus",
		Name:      "layer_errors_total",
		Help:      "Total errors per scan layer",
	}, []string{"layer"})
)

// ─────────────────────────────────────────────────────────────────────────────
// Gauge metrics
// ─────────────────────────────────────────────────────────────────────────────

var (
	EngineRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_antivirus",
		Name:      "engine_running",
		Help:      "Whether the antivirus engine is running (1) or stopped (0)",
	})

	LoadedLayers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_antivirus",
		Name:      "loaded_layers",
		Help:      "Number of registered scan layers",
	})

	CacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_antivirus",
		Name:      "cache_size",
		Help:      "Current number of entries in the scan cache",
	})

	SignatureDBVersion = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_antivirus",
		Name:      "signature_db_version",
		Help:      "Current signature database version (as timestamp)",
	})
)

// ─────────────────────────────────────────────────────────────────────────────
// Histogram metrics
// ─────────────────────────────────────────────────────────────────────────────

var (
	ScanDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "axiom_antivirus",
		Name:      "scan_duration_seconds",
		Help:      "Time taken to complete a file scan",
		Buckets:   prometheus.DefBuckets,
	})

	LayerDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "axiom_antivirus",
		Name:      "layer_duration_seconds",
		Help:      "Time taken per scan layer",
		Buckets:   prometheus.DefBuckets,
	}, []string{"layer"})
)
