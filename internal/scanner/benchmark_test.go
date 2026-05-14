package scanner

import (
	"context"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// Benchmarks — Performance regression detection
// ─────────────────────────────────────────────────────────────────────────────

// BenchmarkOrchestrator_Scan_NoScanners measures baseline overhead of the
// orchestrator with no registered scanners.
func BenchmarkOrchestrator_Scan_NoScanners(b *testing.B) {
	orch := NewOrchestrator()
	file := &FileInfo{Filename: "test.txt", Size: 100, Content: make([]byte, 100)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orch.Scan(file)
	}
}

// BenchmarkOrchestrator_Scan_SingleScanner measures single-scanner overhead.
func BenchmarkOrchestrator_Scan_SingleScanner(b *testing.B) {
	orch := NewOrchestrator(&mockScanner{name: "bench_scanner"})
	file := &FileInfo{Filename: "test.txt", Size: 100, Content: make([]byte, 100)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orch.Scan(file)
	}
}

// BenchmarkOrchestrator_Scan_MultipleScanners measures overhead with 6 scanners
// (matching production configuration).
func BenchmarkOrchestrator_Scan_MultipleScanners(b *testing.B) {
	orch := NewOrchestrator(
		&mockScanner{name: "s1"},
		&mockScanner{name: "s2"},
		&mockScanner{name: "s3"},
		&mockScanner{name: "s4"},
		&mockScanner{name: "s5"},
		&mockScanner{name: "s6"},
	)
	file := &FileInfo{Filename: "test.txt", Size: 100, Content: make([]byte, 100)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orch.Scan(file)
	}
}

// BenchmarkOrchestrator_Scan_Parallel_6Scanners benchmarks parallel mode.
func BenchmarkOrchestrator_Scan_Parallel_6Scanners(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Parallel = true
	orch := NewOrchestratorWithConfig(cfg,
		&mockScanner{name: "s1"},
		&mockScanner{name: "s2"},
		&mockScanner{name: "s3"},
		&mockScanner{name: "s4"},
		&mockScanner{name: "s5"},
		&mockScanner{name: "s6"},
	)
	file := &FileInfo{Filename: "test.txt", Size: 100, Content: make([]byte, 100)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orch.Scan(file)
	}
}

// BenchmarkOrchestrator_Scan_Sequential_6Scanners benchmarks sequential mode.
func BenchmarkOrchestrator_Scan_Sequential_6Scanners(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Parallel = false
	orch := NewOrchestratorWithConfig(cfg,
		&mockScanner{name: "s1"},
		&mockScanner{name: "s2"},
		&mockScanner{name: "s3"},
		&mockScanner{name: "s4"},
		&mockScanner{name: "s5"},
		&mockScanner{name: "s6"},
	)
	file := &FileInfo{Filename: "test.txt", Size: 100, Content: make([]byte, 100)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orch.Scan(file)
	}
}

// BenchmarkOrchestrator_Scan_WithFindings measures overhead when scanners produce findings.
func BenchmarkOrchestrator_Scan_WithFindings(b *testing.B) {
	orch := NewOrchestrator(
		&mockScanner{name: "finder", findings: []Finding{
			{Scanner: "finder", Severity: SeverityHigh, Description: "test finding 1"},
			{Scanner: "finder", Severity: SeverityMedium, Description: "test finding 2"},
			{Scanner: "finder", Severity: SeverityLow, Description: "test finding 3"},
		}},
	)
	file := &FileInfo{Filename: "test.exe", Size: 100, Content: make([]byte, 100)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orch.Scan(file)
	}
}

// BenchmarkOrchestrator_ScanWithContext measures context-aware scan overhead.
func BenchmarkOrchestrator_ScanWithContext(b *testing.B) {
	orch := NewOrchestrator(&mockScanner{name: "ctx_bench"})
	file := &FileInfo{Filename: "test.txt", Size: 100, Content: make([]byte, 100)}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orch.ScanWithContext(ctx, file)
	}
}

// BenchmarkOrchestrator_LargeFile measures scan overhead with a 1MB file.
func BenchmarkOrchestrator_LargeFile(b *testing.B) {
	orch := NewOrchestrator(&mockScanner{name: "large_bench"})
	file := &FileInfo{
		Filename: "large.bin",
		Size:     1024 * 1024,
		Content:  make([]byte, 1024*1024),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orch.Scan(file)
	}
}

// BenchmarkMetrics_Record measures metrics recording overhead per scan.
func BenchmarkMetrics_Record(b *testing.B) {
	m := NewMetrics()
	result := &ScanResult{
		Safe: true, DurationMs: 50,
		Findings: []Finding{
			{Severity: SeverityHigh}, {Severity: SeverityLow},
		},
		Timings: []ScannerTiming{
			{Scanner: "s1", DurationMs: 20, FindingCount: 1},
			{Scanner: "s2", DurationMs: 30, FindingCount: 1},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Record(result)
	}
}

// BenchmarkMetrics_Snapshot measures snapshot generation overhead.
func BenchmarkMetrics_Snapshot(b *testing.B) {
	m := NewMetrics()
	// Seed with data
	for i := 0; i < 100; i++ {
		m.Record(&ScanResult{
			Safe: i%10 != 0, DurationMs: int64(i * 10),
			Findings: []Finding{{Severity: SeverityHigh}},
			Timings:  []ScannerTiming{{Scanner: "s1", DurationMs: int64(i)}},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Snapshot()
	}
}

// BenchmarkSeverity_Weight measures severity weight lookup performance.
func BenchmarkSeverity_Weight(b *testing.B) {
	sevs := []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sevs[i%len(sevs)].Weight()
	}
}

// BenchmarkScanResult_Summary measures summary generation overhead.
func BenchmarkScanResult_Summary(b *testing.B) {
	result := &ScanResult{
		Safe: false, Filename: "test.exe", FileSize: 1024,
		Findings: []Finding{
			{Severity: SeverityHigh}, {Severity: SeverityMedium}, {Severity: SeverityLow},
		},
		Scanners: []string{"s1", "s2", "s3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = result.Summary()
	}
}
