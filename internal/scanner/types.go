package scanner

import (
	"fmt"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Severity
// ─────────────────────────────────────────────────────────────────────────────

// Severity represents the impact level of a scan finding.
// Levels follow industry convention (CVSS-aligned):
//
//	critical — Immediate threat: malware, executable injection, zip bomb
//	high     — Dangerous condition: type spoofing, oversized file, VBA macros
//	medium   — Suspicious indicator: null bytes, encrypted streams, excessive extensions
//	low      — Informational concern: empty file, URI actions
//	info     — Non-security note: scanner unavailable, metadata observation
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Weight returns a numeric weight for severity comparison and sorting.
// Higher values indicate greater severity.
func (s Severity) Weight() int {
	switch s {
	case SeverityCritical:
		return 5
	case SeverityHigh:
		return 4
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

// IsThreat returns true if the severity is medium or above.
func (s Severity) IsThreat() bool {
	return s == SeverityCritical || s == SeverityHigh || s == SeverityMedium
}

// ─────────────────────────────────────────────────────────────────────────────
// Finding
// ─────────────────────────────────────────────────────────────────────────────

// Finding represents a single security issue discovered by a scanner.
// Each scanner may produce zero or more findings per file.
type Finding struct {
	Scanner     string   `json:"scanner"`            // Name of the scanner that produced this finding.
	Severity    Severity `json:"severity"`           // Impact level (critical → info).
	Description string   `json:"description"`        // Human-readable title of the finding.
	Details     string   `json:"details,omitempty"`  // Extended technical detail (optional).
	Offset      int64    `json:"offset,omitempty"`   // Byte offset in file where issue was found (0 = N/A).
	Category    string   `json:"category,omitempty"` // Classification category (e.g. "xss", "macro", "bomb").
}

// ─────────────────────────────────────────────────────────────────────────────
// FileInfo
// ─────────────────────────────────────────────────────────────────────────────

// FileInfo holds metadata and content of a file to be scanned.
// All fields except Content are metadata derived from the upload context.
type FileInfo struct {
	Filename  string // Original filename (e.g. "report.pdf.exe").
	Extension string // Lowercase extension including dot (e.g. ".pdf").
	MIMEType  string // MIME type claimed by the uploader's Content-Type header.
	Size      int64  // Size of Content in bytes.
	SHA256    string // Hex-encoded SHA-256 hash of Content.
	Content   []byte // Raw file bytes. Nil-safe: scanners must handle nil gracefully.
	TempPath  string // Optional: path to temp file on disk (used for large files).
}

// ─────────────────────────────────────────────────────────────────────────────
// ScanResult
// ─────────────────────────────────────────────────────────────────────────────

// ScanResult is the aggregated result of running all registered scanners
// against a single file. Safe is true only when no finding has severity
// ≥ medium.
type ScanResult struct {
	Safe       bool      `json:"safe"`         // True if no medium/high/critical findings exist.
	Filename   string    `json:"filename"`     // Name of the scanned file.
	FileSize   int64     `json:"file_size"`    // Size of the scanned file in bytes.
	MIMEType   string    `json:"mime_type"`    // Detected or claimed MIME type.
	SHA256     string    `json:"sha256"`       // Hex-encoded SHA-256 hash.
	ScannedAt  time.Time `json:"scanned_at"`   // UTC timestamp when the scan started.
	DurationMs int64     `json:"duration_ms"`  // Total scan time in milliseconds.
	Findings   []Finding `json:"findings"`     // All findings from all scanners.
	Scanners   []string  `json:"scanners_run"` // Names of scanners that were executed.
	Timings    []ScannerTiming `json:"timings,omitempty"` // Per-scanner execution timing.
}

// ScannerTiming records the execution time and outcome of a single scanner.
type ScannerTiming struct {
	Scanner      string `json:"scanner"`               // Scanner name.
	DurationMs   int64  `json:"duration_ms"`           // Execution time in milliseconds.
	FindingCount int    `json:"finding_count"`         // Number of findings produced.
	Error        bool   `json:"error,omitempty"`       // True if the scanner returned an error.
	TimedOut     bool   `json:"timed_out,omitempty"`   // True if cancelled due to timeout.
}

// Summary returns a one-line human-readable summary of the scan result.
func (r *ScanResult) Summary() string {
	if r.Safe {
		return fmt.Sprintf("CLEAN %s (%d bytes, %d scanners, %dms)",
			r.Filename, r.FileSize, len(r.Scanners), r.DurationMs)
	}
	threats := 0
	var maxSev Severity
	for _, f := range r.Findings {
		if f.Severity.IsThreat() {
			threats++
			if maxSev == "" || f.Severity.Weight() > maxSev.Weight() {
				maxSev = f.Severity
			}
		}
	}
	return fmt.Sprintf("THREAT %s (%d findings, max_severity=%s, %d bytes, %dms)",
		r.Filename, threats, maxSev, r.FileSize, r.DurationMs)
}

// FindingsBySeverity returns findings filtered to the given severity level.
func (r *ScanResult) FindingsBySeverity(sev Severity) []Finding {
	var result []Finding
	for _, f := range r.Findings {
		if f.Severity == sev {
			result = append(result, f)
		}
	}
	return result
}

// ThreatFindings returns only findings with severity ≥ medium.
func (r *ScanResult) ThreatFindings() []Finding {
	var result []Finding
	for _, f := range r.Findings {
		if f.Severity.IsThreat() {
			result = append(result, f)
		}
	}
	return result
}

// ScannerRan returns true if a scanner with the given name was executed.
func (r *ScanResult) ScannerRan(name string) bool {
	for _, s := range r.Scanners {
		if strings.EqualFold(s, name) {
			return true
		}
	}
	return false
}
