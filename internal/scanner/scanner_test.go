package scanner

import (
	"fmt"
	"testing"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// Severity Tests
// ═══════════════════════════════════════════════════════════════════════════

func TestSeverity_Weight(t *testing.T) {
	tests := []struct {
		severity Severity
		want     int
	}{
		{SeverityCritical, 5},
		{SeverityHigh, 4},
		{SeverityMedium, 3},
		{SeverityLow, 2},
		{SeverityInfo, 1},
		{Severity("unknown"), 0},
		{Severity(""), 0},
	}
	for _, tc := range tests {
		t.Run(string(tc.severity), func(t *testing.T) {
			if got := tc.severity.Weight(); got != tc.want {
				t.Errorf("Severity(%q).Weight() = %d, want %d", tc.severity, got, tc.want)
			}
		})
	}
}

func TestSeverity_WeightOrdering(t *testing.T) {
	// Verify strict descending weight: critical > high > medium > low > info.
	sevs := []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}
	for i := 0; i < len(sevs)-1; i++ {
		if sevs[i].Weight() <= sevs[i+1].Weight() {
			t.Errorf("Weight(%q)=%d should be > Weight(%q)=%d",
				sevs[i], sevs[i].Weight(), sevs[i+1], sevs[i+1].Weight())
		}
	}
}

func TestSeverity_IsThreat(t *testing.T) {
	tests := []struct {
		severity Severity
		want     bool
	}{
		{SeverityCritical, true},
		{SeverityHigh, true},
		{SeverityMedium, true},
		{SeverityLow, false},
		{SeverityInfo, false},
		{Severity(""), false},
	}
	for _, tc := range tests {
		t.Run(string(tc.severity), func(t *testing.T) {
			if got := tc.severity.IsThreat(); got != tc.want {
				t.Errorf("Severity(%q).IsThreat() = %v, want %v", tc.severity, got, tc.want)
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// ScanResult Tests
// ═══════════════════════════════════════════════════════════════════════════

func TestScanResult_Summary_Clean(t *testing.T) {
	r := &ScanResult{
		Safe:       true,
		Filename:   "clean.txt",
		FileSize:   1024,
		Scanners:   []string{"a", "b"},
		DurationMs: 42,
	}
	summary := r.Summary()
	if summary == "" {
		t.Fatal("Summary() returned empty string")
	}
	if got := summary; got != "CLEAN clean.txt (1024 bytes, 2 scanners, 42ms)" {
		t.Errorf("Summary() = %q, want specific format", got)
	}
}

func TestScanResult_Summary_Threat(t *testing.T) {
	r := &ScanResult{
		Safe:     false,
		Filename: "malware.exe",
		FileSize: 2048,
		Findings: []Finding{
			{Severity: SeverityInfo, Description: "info"},
			{Severity: SeverityCritical, Description: "malware"},
			{Severity: SeverityHigh, Description: "suspicious"},
		},
		DurationMs: 100,
	}
	summary := r.Summary()
	// Should include "THREAT" and count only medium+ findings.
	if summary == "" {
		t.Fatal("Summary() returned empty string")
	}
	// 2 threat findings: critical + high (info is not a threat).
	expected := "THREAT malware.exe (2 findings, max_severity=critical, 2048 bytes, 100ms)"
	if summary != expected {
		t.Errorf("Summary() = %q, want %q", summary, expected)
	}
}

func TestScanResult_FindingsBySeverity(t *testing.T) {
	r := &ScanResult{
		Findings: []Finding{
			{Severity: SeverityCritical, Description: "crit-1"},
			{Severity: SeverityHigh, Description: "high-1"},
			{Severity: SeverityCritical, Description: "crit-2"},
			{Severity: SeverityInfo, Description: "info-1"},
		},
	}

	crits := r.FindingsBySeverity(SeverityCritical)
	if len(crits) != 2 {
		t.Errorf("FindingsBySeverity(critical) returned %d, want 2", len(crits))
	}

	highs := r.FindingsBySeverity(SeverityHigh)
	if len(highs) != 1 {
		t.Errorf("FindingsBySeverity(high) returned %d, want 1", len(highs))
	}

	lows := r.FindingsBySeverity(SeverityLow)
	if len(lows) != 0 {
		t.Errorf("FindingsBySeverity(low) returned %d, want 0", len(lows))
	}
}

func TestScanResult_ThreatFindings(t *testing.T) {
	r := &ScanResult{
		Findings: []Finding{
			{Severity: SeverityCritical, Description: "a"},
			{Severity: SeverityHigh, Description: "b"},
			{Severity: SeverityMedium, Description: "c"},
			{Severity: SeverityLow, Description: "d"},
			{Severity: SeverityInfo, Description: "e"},
		},
	}

	threats := r.ThreatFindings()
	if len(threats) != 3 {
		t.Errorf("ThreatFindings() returned %d, want 3 (critical+high+medium)", len(threats))
	}
}

func TestScanResult_ScannerRan(t *testing.T) {
	r := &ScanResult{
		Scanners: []string{"metadata_scanner", "mime_type_validator", "svg_xss_scanner"},
	}

	if !r.ScannerRan("metadata_scanner") {
		t.Error("ScannerRan(metadata_scanner) = false, want true")
	}
	// Case-insensitive.
	if !r.ScannerRan("METADATA_SCANNER") {
		t.Error("ScannerRan(METADATA_SCANNER) = false, want true (case-insensitive)")
	}
	if r.ScannerRan("nonexistent") {
		t.Error("ScannerRan(nonexistent) = true, want false")
	}
}

func TestScanResult_FindingsBySeverity_Empty(t *testing.T) {
	r := &ScanResult{Findings: nil}
	if got := r.FindingsBySeverity(SeverityCritical); got != nil {
		t.Errorf("FindingsBySeverity on nil findings returned %v, want nil", got)
	}
}

func TestScanResult_ThreatFindings_NoThreats(t *testing.T) {
	r := &ScanResult{
		Findings: []Finding{
			{Severity: SeverityLow},
			{Severity: SeverityInfo},
		},
	}
	if got := r.ThreatFindings(); len(got) != 0 {
		t.Errorf("ThreatFindings() returned %d, want 0 for low/info only", len(got))
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// Config Tests
// ═══════════════════════════════════════════════════════════════════════════

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxFileSize != 100*1024*1024 {
		t.Errorf("MaxFileSize = %d, want %d", cfg.MaxFileSize, 100*1024*1024)
	}
	if cfg.Timeout != 2*time.Minute {
		t.Errorf("Timeout = %v, want 2m", cfg.Timeout)
	}
	if !cfg.Parallel {
		t.Error("Parallel = false, want true")
	}
	if cfg.NullByteSampleSize != 8192 {
		t.Errorf("NullByteSampleSize = %d, want 8192", cfg.NullByteSampleSize)
	}
	if cfg.MaxFilenameLength != 255 {
		t.Errorf("MaxFilenameLength = %d, want 255", cfg.MaxFilenameLength)
	}
	if cfg.ArchiveMaxDepth != 5 {
		t.Errorf("ArchiveMaxDepth = %d, want 5", cfg.ArchiveMaxDepth)
	}
	if cfg.ArchiveMaxDecompressedSize != 1024*1024*1024 {
		t.Errorf("ArchiveMaxDecompressedSize = %d, want %d", cfg.ArchiveMaxDecompressedSize, 1024*1024*1024)
	}
	if cfg.ArchiveMaxFiles != 10000 {
		t.Errorf("ArchiveMaxFiles = %d, want 10000", cfg.ArchiveMaxFiles)
	}
	if cfg.ArchiveCompressionRatioLimit != 100.0 {
		t.Errorf("ArchiveCompressionRatioLimit = %.1f, want 100.0", cfg.ArchiveCompressionRatioLimit)
	}
	if len(cfg.AllowedMIMETypes) == 0 {
		t.Error("AllowedMIMETypes is empty, want built-in defaults")
	}
}

func TestDefaultConfig_AllowedMIMETypes(t *testing.T) {
	cfg := DefaultConfig()

	// Verify essential MIME types are present.
	required := []string{
		"text/plain", "text/csv", "application/json", "application/pdf",
		"application/zip", "image/png", "image/jpeg",
	}
	typeSet := make(map[string]bool, len(cfg.AllowedMIMETypes))
	for _, m := range cfg.AllowedMIMETypes {
		typeSet[m] = true
	}
	for _, r := range required {
		if !typeSet[r] {
			t.Errorf("AllowedMIMETypes missing required type %q", r)
		}
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set env vars for testing.
	t.Setenv("SCANNER_TIMEOUT", "30s")
	t.Setenv("SCANNER_PARALLEL", "false")
	t.Setenv("SCANNER_MAX_FILE_SIZE", "52428800")
	t.Setenv("SCANNER_NULL_BYTE_SAMPLE_SIZE", "4096")
	t.Setenv("SCANNER_MAX_FILENAME_LENGTH", "128")
	t.Setenv("SCANNER_ARCHIVE_MAX_DEPTH", "3")
	t.Setenv("SCANNER_ARCHIVE_MAX_DECOMPRESS", "536870912")
	t.Setenv("SCANNER_ARCHIVE_MAX_FILES", "5000")
	t.Setenv("SCANNER_ARCHIVE_RATIO_LIMIT", "50.0")
	t.Setenv("SCANNER_ALLOWED_MIME_TYPES", "text/plain,application/json,image/png")

	cfg := LoadConfigFromEnv()

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", cfg.Timeout)
	}
	if cfg.Parallel {
		t.Error("Parallel = true, want false")
	}
	if cfg.MaxFileSize != 52428800 {
		t.Errorf("MaxFileSize = %d, want 52428800", cfg.MaxFileSize)
	}
	if cfg.NullByteSampleSize != 4096 {
		t.Errorf("NullByteSampleSize = %d, want 4096", cfg.NullByteSampleSize)
	}
	if cfg.MaxFilenameLength != 128 {
		t.Errorf("MaxFilenameLength = %d, want 128", cfg.MaxFilenameLength)
	}
	if cfg.ArchiveMaxDepth != 3 {
		t.Errorf("ArchiveMaxDepth = %d, want 3", cfg.ArchiveMaxDepth)
	}
	if cfg.ArchiveMaxDecompressedSize != 536870912 {
		t.Errorf("ArchiveMaxDecompressedSize = %d, want 536870912", cfg.ArchiveMaxDecompressedSize)
	}
	if cfg.ArchiveMaxFiles != 5000 {
		t.Errorf("ArchiveMaxFiles = %d, want 5000", cfg.ArchiveMaxFiles)
	}
	if cfg.ArchiveCompressionRatioLimit != 50.0 {
		t.Errorf("ArchiveCompressionRatioLimit = %.1f, want 50.0", cfg.ArchiveCompressionRatioLimit)
	}
	if len(cfg.AllowedMIMETypes) != 3 {
		t.Errorf("AllowedMIMETypes has %d entries, want 3", len(cfg.AllowedMIMETypes))
	}
}

func TestLoadConfigFromEnv_Defaults(t *testing.T) {
	// No env vars set — should return DefaultConfig values.
	cfg := LoadConfigFromEnv()
	def := DefaultConfig()

	if cfg.MaxFileSize != def.MaxFileSize {
		t.Errorf("MaxFileSize = %d, want default %d", cfg.MaxFileSize, def.MaxFileSize)
	}
	if cfg.Timeout != def.Timeout {
		t.Errorf("Timeout = %v, want default %v", cfg.Timeout, def.Timeout)
	}
	if cfg.Parallel != def.Parallel {
		t.Errorf("Parallel = %v, want default %v", cfg.Parallel, def.Parallel)
	}
}

func TestLoadConfigFromEnv_InvalidValues(t *testing.T) {
	// Invalid env var values should be silently ignored (defaults used).
	t.Setenv("SCANNER_TIMEOUT", "not-a-duration")
	t.Setenv("SCANNER_PARALLEL", "maybe")
	t.Setenv("SCANNER_MAX_FILE_SIZE", "abc")
	t.Setenv("SCANNER_ARCHIVE_RATIO_LIMIT", "xyz")

	cfg := LoadConfigFromEnv()
	def := DefaultConfig()

	if cfg.Timeout != def.Timeout {
		t.Errorf("Timeout = %v, want default %v (invalid env ignored)", cfg.Timeout, def.Timeout)
	}
	if cfg.Parallel != def.Parallel {
		t.Errorf("Parallel = %v, want default %v (invalid env ignored)", cfg.Parallel, def.Parallel)
	}
	if cfg.MaxFileSize != def.MaxFileSize {
		t.Errorf("MaxFileSize = %d, want default %d (invalid env ignored)", cfg.MaxFileSize, def.MaxFileSize)
	}
	if cfg.ArchiveCompressionRatioLimit != def.ArchiveCompressionRatioLimit {
		t.Errorf("ArchiveCompressionRatioLimit = %.1f, want default %.1f (invalid env ignored)",
			cfg.ArchiveCompressionRatioLimit, def.ArchiveCompressionRatioLimit)
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// Orchestrator Tests
// ═══════════════════════════════════════════════════════════════════════════

// mockScanner is a test-only scanner implementation.
type mockScanner struct {
	name     string
	findings []Finding
	err      error
}

func (m *mockScanner) Name() string                        { return m.name }
func (m *mockScanner) Scan(_ *FileInfo) ([]Finding, error) { return m.findings, m.err }

func TestOrchestrator_Scan_Clean(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{name: "a"},
		&mockScanner{name: "b"},
	)

	result := orch.Scan(&FileInfo{Filename: "test.txt", Size: 100})

	if !result.Safe {
		t.Error("Expected safe=true for clean scan")
	}
	if len(result.Findings) != 0 {
		t.Errorf("Expected 0 findings, got %d", len(result.Findings))
	}
	if len(result.Scanners) != 2 {
		t.Errorf("Expected 2 scanners, got %d", len(result.Scanners))
	}
	if result.Filename != "test.txt" {
		t.Errorf("Filename = %q, want test.txt", result.Filename)
	}
	if result.DurationMs < 0 {
		t.Errorf("DurationMs = %d, should be >= 0", result.DurationMs)
	}
}

func TestOrchestrator_Scan_WithFindings(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{name: "clean"},
		&mockScanner{
			name: "threatfinder",
			findings: []Finding{
				{Scanner: "threatfinder", Severity: SeverityCritical, Description: "malware"},
			},
		},
	)

	result := orch.Scan(&FileInfo{Filename: "bad.exe"})

	if result.Safe {
		t.Error("Expected safe=false when critical finding exists")
	}
	if len(result.Findings) != 1 {
		t.Errorf("Expected 1 finding, got %d", len(result.Findings))
	}
	if result.Findings[0].Severity != SeverityCritical {
		t.Errorf("Finding severity = %q, want critical", result.Findings[0].Severity)
	}
}

func TestOrchestrator_Scan_ErrorHandling(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{name: "failing", err: fmt.Errorf("connection refused")},
		&mockScanner{name: "working"},
	)

	result := orch.Scan(&FileInfo{Filename: "test.txt"})

	// Should still be safe — scanner errors produce info-level findings.
	if !result.Safe {
		t.Error("Expected safe=true when only info-level error findings exist")
	}
	if len(result.Findings) != 1 {
		t.Fatalf("Expected 1 error finding, got %d", len(result.Findings))
	}
	if result.Findings[0].Severity != SeverityInfo {
		t.Errorf("Error finding severity = %q, want info", result.Findings[0].Severity)
	}
	if result.Findings[0].Scanner != "failing" {
		t.Errorf("Error finding scanner = %q, want failing", result.Findings[0].Scanner)
	}
	// Both scanners should be in the scanners list.
	if len(result.Scanners) != 2 {
		t.Errorf("Expected 2 scanners, got %d", len(result.Scanners))
	}
}

func TestOrchestrator_Scan_InfoOnlyIsSafe(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{
			name: "info-only",
			findings: []Finding{
				{Scanner: "info-only", Severity: SeverityInfo, Description: "note"},
			},
		},
	)

	result := orch.Scan(&FileInfo{Filename: "test.txt"})
	if !result.Safe {
		t.Error("Info-only findings should not make result unsafe")
	}
}

func TestOrchestrator_Scan_LowOnlyIsSafe(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{
			name: "low-only",
			findings: []Finding{
				{Scanner: "low-only", Severity: SeverityLow, Description: "minor issue"},
			},
		},
	)

	result := orch.Scan(&FileInfo{Filename: "test.txt"})
	if !result.Safe {
		t.Error("Low-only findings should not make result unsafe")
	}
}

func TestOrchestrator_Scan_MediumIsUnsafe(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{
			name: "med",
			findings: []Finding{
				{Scanner: "med", Severity: SeverityMedium, Description: "suspicious"},
			},
		},
	)

	result := orch.Scan(&FileInfo{Filename: "test.txt"})
	if result.Safe {
		t.Error("Medium findings should make result unsafe")
	}
}

func TestOrchestrator_ScannerNames(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{name: "alpha"},
		&mockScanner{name: "beta"},
		&mockScanner{name: "gamma"},
	)

	names := orch.ScannerNames()
	if len(names) != 3 {
		t.Fatalf("Expected 3 names, got %d", len(names))
	}
	if names[0] != "alpha" || names[1] != "beta" || names[2] != "gamma" {
		t.Errorf("Names = %v, want [alpha beta gamma]", names)
	}
}

func TestOrchestrator_ScannerCount(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{name: "a"},
		&mockScanner{name: "b"},
	)
	if orch.ScannerCount() != 2 {
		t.Errorf("ScannerCount() = %d, want 2", orch.ScannerCount())
	}
}

func TestOrchestrator_NoScanners(t *testing.T) {
	orch := NewOrchestrator()
	result := orch.Scan(&FileInfo{Filename: "test.txt"})

	if !result.Safe {
		t.Error("Empty orchestrator should return safe=true")
	}
	if len(result.Scanners) != 0 {
		t.Errorf("Expected 0 scanners, got %d", len(result.Scanners))
	}
	if len(result.Findings) != 0 {
		t.Errorf("Expected 0 findings, got %d", len(result.Findings))
	}
}

func TestOrchestrator_WithConfig(t *testing.T) {
	cfg := Config{
		Timeout:     10 * time.Second,
		MaxFileSize: 50 * 1024 * 1024,
	}
	orch := NewOrchestratorWithConfig(cfg, &mockScanner{name: "a"})
	if orch.Config().MaxFileSize != 50*1024*1024 {
		t.Errorf("Config().MaxFileSize = %d, want %d", orch.Config().MaxFileSize, 50*1024*1024)
	}
	if orch.Config().Timeout != 10*time.Second {
		t.Errorf("Config().Timeout = %v, want 10s", orch.Config().Timeout)
	}
}

func TestOrchestrator_MultipleFindings(t *testing.T) {
	orch := NewOrchestrator(
		&mockScanner{
			name: "s1",
			findings: []Finding{
				{Scanner: "s1", Severity: SeverityLow, Description: "a"},
				{Scanner: "s1", Severity: SeverityHigh, Description: "b"},
			},
		},
		&mockScanner{
			name: "s2",
			findings: []Finding{
				{Scanner: "s2", Severity: SeverityInfo, Description: "c"},
			},
		},
	)

	result := orch.Scan(&FileInfo{Filename: "multi.bin"})

	if result.Safe {
		t.Error("Expected unsafe when high-severity finding exists")
	}
	if len(result.Findings) != 3 {
		t.Errorf("Expected 3 findings, got %d", len(result.Findings))
	}
	if len(result.Scanners) != 2 {
		t.Errorf("Expected 2 scanners, got %d", len(result.Scanners))
	}
}

func TestOrchestrator_Scan_FileMetadata(t *testing.T) {
	orch := NewOrchestrator(&mockScanner{name: "a"})

	file := &FileInfo{
		Filename: "data.csv",
		Size:     4096,
		MIMEType: "text/csv",
		SHA256:   "abc123",
	}
	result := orch.Scan(file)

	if result.Filename != "data.csv" {
		t.Errorf("Filename = %q, want data.csv", result.Filename)
	}
	if result.FileSize != 4096 {
		t.Errorf("FileSize = %d, want 4096", result.FileSize)
	}
	if result.MIMEType != "text/csv" {
		t.Errorf("MIMEType = %q, want text/csv", result.MIMEType)
	}
	if result.SHA256 != "abc123" {
		t.Errorf("SHA256 = %q, want abc123", result.SHA256)
	}
	if result.ScannedAt.IsZero() {
		t.Error("ScannedAt should not be zero")
	}
}
