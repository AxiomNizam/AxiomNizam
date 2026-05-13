package scanner

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Metrics — Thread-safe accumulation of scan statistics
// ─────────────────────────────────────────────────────────────────────────────

// Metrics collects and exposes operational statistics for the scanner
// pipeline. It is designed to be embedded in the Orchestrator and updated
// after every scan. All methods are safe for concurrent use.
//
// Use Snapshot() to obtain a point-in-time copy of all metrics for
// serialization or health endpoint responses.
type Metrics struct {
	mu sync.RWMutex

	// ── Global scan counters ─────────────────────────────────────────────
	totalScans   int64 // Total number of scans executed.
	totalSafe    int64 // Scans where result.Safe == true.
	totalUnsafe  int64 // Scans where result.Safe == false.

	// ── Finding counters ─────────────────────────────────────────────────
	totalFindings int64             // Total findings across all scans.
	bySeverity    map[Severity]int64 // Finding count per severity level.

	// ── Per-scanner counters ─────────────────────────────────────────────
	scannerScans    map[string]int64 // Number of times each scanner ran.
	scannerFindings map[string]int64 // Total findings per scanner.
	scannerErrors   map[string]int64 // Total errors per scanner.
	scannerTimeouts map[string]int64 // Total timeouts per scanner.
	scannerTotalMs  map[string]int64 // Cumulative execution time per scanner.

	// ── Timing ───────────────────────────────────────────────────────────
	totalDurationMs int64     // Cumulative scan duration across all scans.
	maxDurationMs   int64     // Longest single scan duration.
	minDurationMs   int64     // Shortest single scan duration (-1 = unset).
	lastScanAt      time.Time // Timestamp of the last completed scan.
	startedAt       time.Time // When the metrics collector was created.
}

// NewMetrics creates an initialized Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		bySeverity:      make(map[Severity]int64),
		scannerScans:    make(map[string]int64),
		scannerFindings: make(map[string]int64),
		scannerErrors:   make(map[string]int64),
		scannerTimeouts: make(map[string]int64),
		scannerTotalMs:  make(map[string]int64),
		minDurationMs:   -1,
		startedAt:       time.Now().UTC(),
	}
}

// Record updates all metrics counters from a completed ScanResult.
// This is called automatically by the orchestrator after every scan.
func (m *Metrics) Record(result *ScanResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Global counters
	m.totalScans++
	if result.Safe {
		m.totalSafe++
	} else {
		m.totalUnsafe++
	}

	// Duration tracking
	m.totalDurationMs += result.DurationMs
	if result.DurationMs > m.maxDurationMs {
		m.maxDurationMs = result.DurationMs
	}
	if m.minDurationMs < 0 || result.DurationMs < m.minDurationMs {
		m.minDurationMs = result.DurationMs
	}
	m.lastScanAt = result.ScannedAt

	// Finding counters by severity
	m.totalFindings += int64(len(result.Findings))
	for _, f := range result.Findings {
		m.bySeverity[f.Severity]++
	}

	// Per-scanner timing breakdown
	for _, t := range result.Timings {
		m.scannerScans[t.Scanner]++
		m.scannerFindings[t.Scanner] += int64(t.FindingCount)
		m.scannerTotalMs[t.Scanner] += t.DurationMs
		if t.Error {
			m.scannerErrors[t.Scanner]++
		}
		if t.TimedOut {
			m.scannerTimeouts[t.Scanner]++
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// MetricsSnapshot — Serializable point-in-time copy
// ─────────────────────────────────────────────────────────────────────────────

// MetricsSnapshot is a serializable, immutable view of the metrics at a
// point in time. Safe to marshal to JSON for health/stats endpoints.
type MetricsSnapshot struct {
	// ── Global ───────────────────────────────────────────────────────────
	TotalScans   int64  `json:"total_scans"`
	TotalSafe    int64  `json:"total_safe"`
	TotalUnsafe  int64  `json:"total_unsafe"`
	SafetyRate   string `json:"safety_rate"` // e.g. "98.5%"

	// ── Findings ─────────────────────────────────────────────────────────
	TotalFindings       int64            `json:"total_findings"`
	FindingsBySeverity  map[string]int64 `json:"findings_by_severity"`

	// ── Timing ───────────────────────────────────────────────────────────
	TotalDurationMs     int64     `json:"total_duration_ms"`
	AvgDurationMs       int64     `json:"avg_duration_ms"`
	MaxDurationMs       int64     `json:"max_duration_ms"`
	MinDurationMs       int64     `json:"min_duration_ms"`
	LastScanAt          time.Time `json:"last_scan_at,omitempty"`
	UptimeSeconds       int64     `json:"uptime_seconds"`

	// ── Per-scanner ──────────────────────────────────────────────────────
	Scanners            []ScannerMetrics `json:"scanners"`
}

// ScannerMetrics holds per-scanner statistics.
type ScannerMetrics struct {
	Name         string `json:"name"`
	TotalRuns    int64  `json:"total_runs"`
	TotalFindings int64 `json:"total_findings"`
	TotalErrors  int64  `json:"total_errors"`
	TotalTimeouts int64 `json:"total_timeouts"`
	TotalMs      int64  `json:"total_ms"`
	AvgMs        int64  `json:"avg_ms"`
}

// Snapshot returns a thread-safe, serializable copy of the current metrics.
func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snap := MetricsSnapshot{
		TotalScans:      m.totalScans,
		TotalSafe:       m.totalSafe,
		TotalUnsafe:     m.totalUnsafe,
		TotalFindings:   m.totalFindings,
		TotalDurationMs: m.totalDurationMs,
		MaxDurationMs:   m.maxDurationMs,
		MinDurationMs:   m.minDurationMs,
		LastScanAt:      m.lastScanAt,
		UptimeSeconds:   int64(time.Since(m.startedAt).Seconds()),
	}

	// Safety rate calculation
	if m.totalScans > 0 {
		rate := float64(m.totalSafe) / float64(m.totalScans) * 100
		snap.SafetyRate = fmt.Sprintf("%.1f%%", rate)
		snap.AvgDurationMs = m.totalDurationMs / m.totalScans
	} else {
		snap.SafetyRate = "N/A"
	}

	if snap.MinDurationMs < 0 {
		snap.MinDurationMs = 0
	}

	// Severity distribution (use string keys for JSON)
	snap.FindingsBySeverity = make(map[string]int64, len(m.bySeverity))
	for sev, count := range m.bySeverity {
		snap.FindingsBySeverity[string(sev)] = count
	}

	// Per-scanner breakdown
	snap.Scanners = make([]ScannerMetrics, 0, len(m.scannerScans))
	for name, runs := range m.scannerScans {
		sm := ScannerMetrics{
			Name:          name,
			TotalRuns:     runs,
			TotalFindings: m.scannerFindings[name],
			TotalErrors:   m.scannerErrors[name],
			TotalTimeouts: m.scannerTimeouts[name],
			TotalMs:       m.scannerTotalMs[name],
		}
		if runs > 0 {
			sm.AvgMs = m.scannerTotalMs[name] / runs
		}
		snap.Scanners = append(snap.Scanners, sm)
	}

	return snap
}

// Reset clears all accumulated metrics. Useful for testing or periodic resets.
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalScans = 0
	m.totalSafe = 0
	m.totalUnsafe = 0
	m.totalFindings = 0
	m.totalDurationMs = 0
	m.maxDurationMs = 0
	m.minDurationMs = -1
	m.lastScanAt = time.Time{}

	m.bySeverity = make(map[Severity]int64)
	m.scannerScans = make(map[string]int64)
	m.scannerFindings = make(map[string]int64)
	m.scannerErrors = make(map[string]int64)
	m.scannerTimeouts = make(map[string]int64)
	m.scannerTotalMs = make(map[string]int64)
}

// TotalScans returns the total number of scans executed (atomic read).
func (m *Metrics) TotalScans() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalScans
}

// ─────────────────────────────────────────────────────────────────────────────
// Health — Pipeline health status
// ─────────────────────────────────────────────────────────────────────────────

// HealthStatus represents the operational state of the scanner pipeline.
type HealthStatus struct {
	Status        string          `json:"status"`         // "healthy", "degraded", or "unavailable"
	ScannerCount  int             `json:"scanner_count"`  // Number of registered scanners.
	Scanners      []string        `json:"scanners"`       // Names of registered scanners.
	TotalScans    int64           `json:"total_scans"`    // Scans since startup.
	UptimeSeconds int64           `json:"uptime_seconds"` // Time since metrics collection started.
	LastScanAt    time.Time       `json:"last_scan_at,omitempty"` // Last scan timestamp.
	ErrorRate     string          `json:"error_rate"`     // e.g. "0.5%"
	Metrics       *MetricsSnapshot `json:"metrics,omitempty"` // Full metrics if requested.
}

// ─────────────────────────────────────────────────────────────────────────────
// Atomic counter helper (used for lightweight counters)
// ─────────────────────────────────────────────────────────────────────────────

// AtomicCounter is a simple atomic int64 counter for high-frequency operations.
type AtomicCounter struct {
	val int64
}

// Inc increments the counter by 1 and returns the new value.
func (c *AtomicCounter) Inc() int64 {
	return atomic.AddInt64(&c.val, 1)
}

// Add increments the counter by delta and returns the new value.
func (c *AtomicCounter) Add(delta int64) int64 {
	return atomic.AddInt64(&c.val, delta)
}

// Load returns the current counter value.
func (c *AtomicCounter) Load() int64 {
	return atomic.LoadInt64(&c.val)
}

// Reset sets the counter to zero.
func (c *AtomicCounter) Reset() {
	atomic.StoreInt64(&c.val, 0)
}
