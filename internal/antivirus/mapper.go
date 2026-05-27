package antivirus

import "fmt"

// StatusToResponse builds a StatusResponse from engine state.
func StatusToResponse(status string, version string, stats EngineStats, cfg *Config) StatusResponse {
	return StatusResponse{
		Status:        status,
		EngineVersion: version,
		SigDBVersion:  stats.SigDBVersion,
		LayersEnabled: stats.LayersEnabled,
		LayerCount:    len(stats.LayersEnabled),
		UptimeSeconds: stats.UptimeSeconds,
		ScanCapacity: ScanCapacity{
			Workers:     cfg.Workers,
			QueueSize:   cfg.QueueSize,
			MaxFileSize: cfg.MaxFileSize,
		},
		Features: FeaturesConfig{
			HashDB:    cfg.HashDBEnabled,
			Pattern:   cfg.PatternEnabled,
			Heuristic: cfg.HeuristicEnabled,
			Entropy:   cfg.EntropyEnabled,
			YARA:      cfg.YARAEnabled,
		},
	}
}

// StatsToResponse builds a StatsResponse from engine stats.
func StatsToResponse(stats EngineStats) StatsResponse {
	return StatsResponse{
		TotalScanned:  stats.TotalScanned,
		ThreatsFound:  stats.ThreatsFound,
		CleanFiles:    stats.CleanFiles,
		ErrorCount:    stats.ErrorCount,
		BytesScanned:  stats.BytesScanned,
		AvgScanMs:     fmt.Sprintf("%.2f", stats.AvgScanMs),
		Cache: CacheStats{
			Hits:    stats.CacheHits,
			Misses:  stats.CacheMisses,
			HitRate: fmt.Sprintf("%.4f", stats.CacheHitRate),
		},
		UptimeSeconds: stats.UptimeSeconds,
		EngineVersion: stats.EngineVersion,
		SigDBVersion:  stats.SigDBVersion,
	}
}

// ConfigToResponse builds a ConfigResponse from engine config.
func ConfigToResponse(cfg *Config) ConfigResponse {
	return ConfigResponse{
		Enabled:          cfg.Enabled,
		Workers:          cfg.Workers,
		QueueSize:        cfg.QueueSize,
		MaxFileSize:      cfg.MaxFileSize,
		CacheSize:        cfg.CacheSize,
		CacheTTL:         cfg.CacheTTL.String(),
		UpdateURL:        redactURL(cfg.UpdateURL),
		UpdateInterval:   cfg.UpdateInterval.String(),
		SigDir:           cfg.SigDir,
		QuarantineAction: string(cfg.QuarantineAction),
		Layers: FeaturesConfig{
			HashDB:    cfg.HashDBEnabled,
			Pattern:   cfg.PatternEnabled,
			Heuristic: cfg.HeuristicEnabled,
			Entropy:   cfg.EntropyEnabled,
			YARA:      cfg.YARAEnabled,
		},
	}
}

// ThreatsToResponse converts engine threats to ThreatListResponse.
func ThreatsToResponse(threats []ScanResult) ThreatListResponse {
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
	return ThreatListResponse{Threats: records, Count: len(records)}
}
