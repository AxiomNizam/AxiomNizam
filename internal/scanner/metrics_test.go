package scanner

import (
	"context"
	"testing"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Metrics tests
// ─────────────────────────────────────────────────────────────────────────────

func TestMetrics_NewMetrics(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}
	if m.totalScans != 0 {
		t.Errorf("expected totalScans=0, got %d", m.totalScans)
	}
	if m.minDurationMs != -1 {
		t.Errorf("expected minDurationMs=-1, got %d", m.minDurationMs)
	}
	if m.startedAt.IsZero() {
		t.Error("expected startedAt to be set")
	}
}

func TestMetrics_Record_SafeScan(t *testing.T) {
	m := NewMetrics()
	m.Record(&ScanResult{
		Safe:       true,
		DurationMs: 50,
		ScannedAt:  time.Now().UTC(),
		Findings:   []Finding{},
		Timings: []ScannerTiming{
			{Scanner: "test_scanner", DurationMs: 50, FindingCount: 0},
		},
	})

	if m.totalScans != 1 {
		t.Errorf("expected totalScans=1, got %d", m.totalScans)
	}
	if m.totalSafe != 1 {
		t.Errorf("expected totalSafe=1, got %d", m.totalSafe)
	}
	if m.totalUnsafe != 0 {
		t.Errorf("expected totalUnsafe=0, got %d", m.totalUnsafe)
	}
}

func TestMetrics_Record_UnsafeScan(t *testing.T) {
	m := NewMetrics()
	m.Record(&ScanResult{
		Safe:       false,
		DurationMs: 100,
		ScannedAt:  time.Now().UTC(),
		Findings: []Finding{
			{Scanner: "test", Severity: SeverityHigh, Description: "threat"},
			{Scanner: "test", Severity: SeverityCritical, Description: "critical threat"},
		},
		Timings: []ScannerTiming{
			{Scanner: "test_scanner", DurationMs: 100, FindingCount: 2},
		},
	})

	if m.totalUnsafe != 1 {
		t.Errorf("expected totalUnsafe=1, got %d", m.totalUnsafe)
	}
	if m.totalFindings != 2 {
		t.Errorf("expected totalFindings=2, got %d", m.totalFindings)
	}
	if m.bySeverity[SeverityHigh] != 1 {
		t.Errorf("expected 1 high finding, got %d", m.bySeverity[SeverityHigh])
	}
	if m.bySeverity[SeverityCritical] != 1 {
		t.Errorf("expected 1 critical finding, got %d", m.bySeverity[SeverityCritical])
	}
}

func TestMetrics_Record_DurationTracking(t *testing.T) {
	m := NewMetrics()
	now := time.Now().UTC()

	m.Record(&ScanResult{Safe: true, DurationMs: 100, ScannedAt: now, Timings: []ScannerTiming{{Scanner: "a", DurationMs: 100}}})
	m.Record(&ScanResult{Safe: true, DurationMs: 50, ScannedAt: now, Timings: []ScannerTiming{{Scanner: "a", DurationMs: 50}}})
	m.Record(&ScanResult{Safe: true, DurationMs: 200, ScannedAt: now, Timings: []ScannerTiming{{Scanner: "a", DurationMs: 200}}})

	if m.totalDurationMs != 350 {
		t.Errorf("expected totalDurationMs=350, got %d", m.totalDurationMs)
	}
	if m.maxDurationMs != 200 {
		t.Errorf("expected maxDurationMs=200, got %d", m.maxDurationMs)
	}
	if m.minDurationMs != 50 {
		t.Errorf("expected minDurationMs=50, got %d", m.minDurationMs)
	}
}

func TestMetrics_Record_PerScanner(t *testing.T) {
	m := NewMetrics()
	now := time.Now().UTC()

	m.Record(&ScanResult{
		Safe: true, DurationMs: 100, ScannedAt: now,
		Timings: []ScannerTiming{
			{Scanner: "scanner_a", DurationMs: 30, FindingCount: 1},
			{Scanner: "scanner_b", DurationMs: 70, FindingCount: 0, Error: true},
		},
	})

	if m.scannerScans["scanner_a"] != 1 {
		t.Errorf("expected scanner_a runs=1, got %d", m.scannerScans["scanner_a"])
	}
	if m.scannerFindings["scanner_a"] != 1 {
		t.Errorf("expected scanner_a findings=1, got %d", m.scannerFindings["scanner_a"])
	}
	if m.scannerErrors["scanner_b"] != 1 {
		t.Errorf("expected scanner_b errors=1, got %d", m.scannerErrors["scanner_b"])
	}
	if m.scannerTotalMs["scanner_a"] != 30 {
		t.Errorf("expected scanner_a totalMs=30, got %d", m.scannerTotalMs["scanner_a"])
	}
}

func TestMetrics_Record_Timeouts(t *testing.T) {
	m := NewMetrics()
	m.Record(&ScanResult{
		Safe: true, DurationMs: 100, ScannedAt: time.Now().UTC(),
		Timings: []ScannerTiming{
			{Scanner: "slow", DurationMs: 100, Error: true, TimedOut: true},
		},
	})

	if m.scannerTimeouts["slow"] != 1 {
		t.Errorf("expected timeouts=1, got %d", m.scannerTimeouts["slow"])
	}
}

func TestMetrics_Snapshot(t *testing.T) {
	m := NewMetrics()
	now := time.Now().UTC()

	// Record 2 scans: 1 safe, 1 unsafe
	m.Record(&ScanResult{
		Safe: true, DurationMs: 50, ScannedAt: now,
		Findings: []Finding{},
		Timings:  []ScannerTiming{{Scanner: "a", DurationMs: 50}},
	})
	m.Record(&ScanResult{
		Safe: false, DurationMs: 150, ScannedAt: now,
		Findings: []Finding{
			{Severity: SeverityHigh, Description: "test"},
			{Severity: SeverityMedium, Description: "test2"},
		},
		Timings: []ScannerTiming{{Scanner: "a", DurationMs: 150, FindingCount: 2}},
	})

	snap := m.Snapshot()

	if snap.TotalScans != 2 {
		t.Errorf("expected TotalScans=2, got %d", snap.TotalScans)
	}
	if snap.TotalSafe != 1 {
		t.Errorf("expected TotalSafe=1, got %d", snap.TotalSafe)
	}
	if snap.TotalUnsafe != 1 {
		t.Errorf("expected TotalUnsafe=1, got %d", snap.TotalUnsafe)
	}
	if snap.SafetyRate != "50.0%" {
		t.Errorf("expected SafetyRate=50.0%%, got %s", snap.SafetyRate)
	}
	if snap.TotalFindings != 2 {
		t.Errorf("expected TotalFindings=2, got %d", snap.TotalFindings)
	}
	if snap.AvgDurationMs != 100 {
		t.Errorf("expected AvgDurationMs=100, got %d", snap.AvgDurationMs)
	}
	if snap.MaxDurationMs != 150 {
		t.Errorf("expected MaxDurationMs=150, got %d", snap.MaxDurationMs)
	}
	if snap.MinDurationMs != 50 {
		t.Errorf("expected MinDurationMs=50, got %d", snap.MinDurationMs)
	}
	if snap.FindingsBySeverity[string(SeverityHigh)] != 1 {
		t.Errorf("expected 1 high finding, got %d", snap.FindingsBySeverity[string(SeverityHigh)])
	}
	if snap.UptimeSeconds < 0 {
		t.Errorf("expected non-negative uptime, got %d", snap.UptimeSeconds)
	}
	if len(snap.Scanners) != 1 {
		t.Fatalf("expected 1 scanner, got %d", len(snap.Scanners))
	}
	if snap.Scanners[0].Name != "a" {
		t.Errorf("expected scanner name 'a', got %s", snap.Scanners[0].Name)
	}
	if snap.Scanners[0].TotalRuns != 2 {
		t.Errorf("expected 2 runs, got %d", snap.Scanners[0].TotalRuns)
	}
}

func TestMetrics_Snapshot_NoScans(t *testing.T) {
	m := NewMetrics()
	snap := m.Snapshot()

	if snap.TotalScans != 0 {
		t.Errorf("expected TotalScans=0, got %d", snap.TotalScans)
	}
	if snap.SafetyRate != "N/A" {
		t.Errorf("expected SafetyRate=N/A, got %s", snap.SafetyRate)
	}
	if snap.MinDurationMs != 0 {
		t.Errorf("expected MinDurationMs=0 (normalized), got %d", snap.MinDurationMs)
	}
}

func TestMetrics_Reset(t *testing.T) {
	m := NewMetrics()
	m.Record(&ScanResult{
		Safe: true, DurationMs: 100, ScannedAt: time.Now().UTC(),
		Findings: []Finding{{Severity: SeverityLow}},
		Timings:  []ScannerTiming{{Scanner: "a", DurationMs: 100, FindingCount: 1}},
	})

	m.Reset()

	if m.totalScans != 0 {
		t.Errorf("expected totalScans=0 after reset, got %d", m.totalScans)
	}
	if m.totalFindings != 0 {
		t.Errorf("expected totalFindings=0 after reset, got %d", m.totalFindings)
	}
	if m.minDurationMs != -1 {
		t.Errorf("expected minDurationMs=-1 after reset, got %d", m.minDurationMs)
	}
	if len(m.scannerScans) != 0 {
		t.Errorf("expected empty scannerScans after reset, got %d", len(m.scannerScans))
	}
}

func TestMetrics_TotalScans(t *testing.T) {
	m := NewMetrics()
	m.Record(&ScanResult{Safe: true, DurationMs: 10, ScannedAt: time.Now().UTC(), Timings: []ScannerTiming{}})
	m.Record(&ScanResult{Safe: true, DurationMs: 20, ScannedAt: time.Now().UTC(), Timings: []ScannerTiming{}})

	if m.TotalScans() != 2 {
		t.Errorf("expected TotalScans()=2, got %d", m.TotalScans())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Orchestrator metrics integration tests
// ─────────────────────────────────────────────────────────────────────────────

func TestOrchestrator_Metrics_AutoRecord(t *testing.T) {
	orch := NewOrchestrator(&mockScanner{name: "clean", findings: nil})
	file := &FileInfo{Filename: "test.txt", Size: 100, Content: []byte("test")}

	orch.Scan(file)
	orch.Scan(file)

	snap := orch.Metrics().Snapshot()
	if snap.TotalScans != 2 {
		t.Errorf("expected 2 scans recorded, got %d", snap.TotalScans)
	}
	if snap.TotalSafe != 2 {
		t.Errorf("expected 2 safe scans, got %d", snap.TotalSafe)
	}
}

func TestOrchestrator_Metrics_FindingsTracked(t *testing.T) {
	orch := NewOrchestrator(&mockScanner{
		name: "threat",
		findings: []Finding{
			{Scanner: "threat", Severity: SeverityHigh, Description: "bad"},
		},
	})
	file := &FileInfo{Filename: "test.exe", Size: 100, Content: []byte("MZ")}

	orch.Scan(file)

	snap := orch.Metrics().Snapshot()
	if snap.TotalFindings != 1 {
		t.Errorf("expected 1 finding tracked, got %d", snap.TotalFindings)
	}
	if snap.TotalUnsafe != 1 {
		t.Errorf("expected 1 unsafe scan, got %d", snap.TotalUnsafe)
	}
	if snap.FindingsBySeverity[string(SeverityHigh)] != 1 {
		t.Errorf("expected 1 high-severity finding, got %d", snap.FindingsBySeverity[string(SeverityHigh)])
	}
}

func TestOrchestrator_Metrics_PerScannerTimings(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{name: "fast", findings: nil},
		&mockScanner{name: "slow", findings: []Finding{{Severity: SeverityLow}}},
	)
	file := &FileInfo{Filename: "test.txt", Size: 10, Content: []byte("hello")}

	orch.Scan(file)

	snap := orch.Metrics().Snapshot()
	if len(snap.Scanners) != 2 {
		t.Fatalf("expected 2 scanner metrics, got %d", len(snap.Scanners))
	}
}

func TestOrchestrator_Metrics_WithContext(t *testing.T) {
	orch := NewOrchestrator(&mockScanner{name: "ctx_scanner", findings: nil})
	file := &FileInfo{Filename: "test.txt", Size: 10, Content: []byte("hello")}

	ctx := context.Background()
	orch.ScanWithContext(ctx, file)

	if orch.Metrics().TotalScans() != 1 {
		t.Errorf("expected 1 scan via ScanWithContext, got %d", orch.Metrics().TotalScans())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Health tests
// ─────────────────────────────────────────────────────────────────────────────

func TestOrchestrator_Health_Healthy(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{name: "scanner_a"},
		&mockScanner{name: "scanner_b"},
	)
	file := &FileInfo{Filename: "test.txt", Size: 10, Content: []byte("hello")}
	orch.Scan(file)

	h := orch.Health(false)
	if h.Status != "healthy" {
		t.Errorf("expected status=healthy, got %s", h.Status)
	}
	if h.ScannerCount != 2 {
		t.Errorf("expected 2 scanners, got %d", h.ScannerCount)
	}
	if h.TotalScans != 1 {
		t.Errorf("expected 1 total scan, got %d", h.TotalScans)
	}
	if h.Metrics != nil {
		t.Error("expected no metrics when includeMetrics=false")
	}
}

func TestOrchestrator_Health_WithMetrics(t *testing.T) {
	orch := NewOrchestrator(&mockScanner{name: "scanner_a"})
	file := &FileInfo{Filename: "test.txt", Size: 10, Content: []byte("hello")}
	orch.Scan(file)

	h := orch.Health(true)
	if h.Metrics == nil {
		t.Fatal("expected metrics when includeMetrics=true")
	}
	if h.Metrics.TotalScans != 1 {
		t.Errorf("expected 1 total scan in metrics, got %d", h.Metrics.TotalScans)
	}
}

func TestOrchestrator_Health_NoScanners(t *testing.T) {
	orch := NewOrchestrator()
	h := orch.Health(false)

	if h.Status != "unavailable" {
		t.Errorf("expected status=unavailable with no scanners, got %s", h.Status)
	}
}

func TestOrchestrator_Health_ErrorRate(t *testing.T) {
	orch := NewOrchestrator(&mockScanner{name: "clean"})
	file := &FileInfo{Filename: "test.txt", Size: 10, Content: []byte("hello")}
	orch.Scan(file)

	h := orch.Health(false)
	if h.ErrorRate != "0.0%" {
		t.Errorf("expected error rate 0.0%%, got %s", h.ErrorRate)
	}
}

func TestOrchestrator_Health_ScannerList(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{name: "alpha"},
		&mockScanner{name: "beta"},
		&mockScanner{name: "gamma"},
	)

	h := orch.Health(false)
	if len(h.Scanners) != 3 {
		t.Errorf("expected 3 scanner names, got %d", len(h.Scanners))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// AtomicCounter tests
// ─────────────────────────────────────────────────────────────────────────────

func TestAtomicCounter_Inc(t *testing.T) {
	c := &AtomicCounter{}
	c.Inc()
	c.Inc()
	if c.Load() != 2 {
		t.Errorf("expected 2, got %d", c.Load())
	}
}

func TestAtomicCounter_Add(t *testing.T) {
	c := &AtomicCounter{}
	c.Add(10)
	c.Add(5)
	if c.Load() != 15 {
		t.Errorf("expected 15, got %d", c.Load())
	}
}

func TestAtomicCounter_Reset(t *testing.T) {
	c := &AtomicCounter{}
	c.Inc()
	c.Inc()
	c.Reset()
	if c.Load() != 0 {
		t.Errorf("expected 0 after reset, got %d", c.Load())
	}
}
