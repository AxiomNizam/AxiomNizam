package metrics

import "time"

// RecordScan records a completed scan with its verdict.
func RecordScan(verdict string, duration time.Duration, bytes int64, cacheHit bool) {
	ScansTotal.Inc()
	BytesScanned.Add(float64(bytes))
	ScanDuration.Observe(duration.Seconds())

	switch verdict {
	case "clean":
		ScansClean.Inc()
	case "malware":
		ScansMalware.Inc()
	case "suspicious":
		ScansSuspicious.Inc()
	case "error":
		ScansError.Inc()
	}

	if cacheHit {
		CacheHits.Inc()
	} else {
		CacheMisses.Inc()
	}
}

// RecordThreat records a detected threat.
func RecordThreat() { ThreatsDetected.Inc() }

// RecordLayerError records an error from a specific scan layer.
func RecordLayerError(layer string) { LayerErrors.WithLabelValues(layer).Inc() }

// RecordLayerDuration records the duration of a specific scan layer.
func RecordLayerDuration(layer string, d time.Duration) {
	LayerDuration.WithLabelValues(layer).Observe(d.Seconds())
}

// SetEngineRunning sets the engine running state.
func SetEngineRunning(running bool) {
	if running {
		EngineRunning.Set(1)
	} else {
		EngineRunning.Set(0)
	}
}

// SetLoadedLayers sets the number of loaded scan layers.
func SetLoadedLayers(count float64) { LoadedLayers.Set(count) }

// SetCacheSize sets the current cache size.
func SetCacheSize(size float64) { CacheSize.Set(size) }
