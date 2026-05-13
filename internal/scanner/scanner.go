package scanner

import (
	"fmt"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Scanner Interface
// ─────────────────────────────────────────────────────────────────────────────

// Scanner is the interface that all individual scanners in the SafeGate
// pipeline must implement. Each scanner inspects a file and returns zero
// or more findings.
//
// Implementations:
//   - MetadataScanner  — file size, empty files, null bytes, suspicious filenames
//   - MIMEScanner      — content-type validation, type spoofing, executable detection
//   - SVGScanner       — XSS vectors in SVG files (script, event handlers, JS URIs)
//   - MacroScanner     — VBA macros, auto-exec, shell commands in Office/PDF files
//   - ArchiveScanner   — zip bombs, path traversal, executable entries in archives
//   - NativeAVScanner  — malware detection via internal antivirus engine
type Scanner interface {
	// Name returns a unique, stable identifier for this scanner.
	// Used in Finding.Scanner and in orchestrator logs.
	Name() string

	// Scan inspects the file and returns any findings.
	// Implementations must be safe to call concurrently.
	// Return (nil, nil) to indicate "no findings, no error".
	Scan(file *FileInfo) ([]Finding, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Orchestrator
// ─────────────────────────────────────────────────────────────────────────────

// Orchestrator runs all registered scanners against a file and aggregates
// their findings into a single ScanResult.
//
// Current behavior:
//   - Scanners run sequentially in registration order.
//   - If a scanner returns an error, an info-level finding is recorded and
//     the next scanner proceeds — the pipeline never aborts.
//   - After all scanners run, the result is marked as unsafe if any finding
//     has severity ≥ medium (critical, high, or medium).
type Orchestrator struct {
	scanners []Scanner
	config   Config
}

// NewOrchestrator creates an orchestrator with the given scanners.
// Uses DefaultConfig(). For custom configuration, use NewOrchestratorWithConfig.
func NewOrchestrator(scanners ...Scanner) *Orchestrator {
	return &Orchestrator{
		scanners: scanners,
		config:   DefaultConfig(),
	}
}

// NewOrchestratorWithConfig creates an orchestrator with the given scanners
// and explicit configuration.
func NewOrchestratorWithConfig(cfg Config, scanners ...Scanner) *Orchestrator {
	return &Orchestrator{
		scanners: scanners,
		config:   cfg,
	}
}

// Config returns the orchestrator's active configuration.
func (o *Orchestrator) Config() Config {
	return o.config
}

// Scan runs all registered scanners and returns aggregated results.
func (o *Orchestrator) Scan(file *FileInfo) *ScanResult {
	start := time.Now()

	result := &ScanResult{
		Safe:      true,
		Filename:  file.Filename,
		FileSize:  file.Size,
		MIMEType:  file.MIMEType,
		SHA256:    file.SHA256,
		ScannedAt: start.UTC(),
		Findings:  make([]Finding, 0),
		Scanners:  make([]string, 0, len(o.scanners)),
	}

	for _, s := range o.scanners {
		result.Scanners = append(result.Scanners, s.Name())

		findings, err := s.Scan(file)
		if err != nil {
			result.Findings = append(result.Findings, Finding{
				Scanner:     s.Name(),
				Severity:    SeverityInfo,
				Description: "Scanner unavailable",
				Details:     fmt.Sprintf("Scanner %q could not complete: %v", s.Name(), err),
			})
			continue
		}

		result.Findings = append(result.Findings, findings...)
	}

	// Determine safety: any medium+ finding makes the result unsafe.
	for _, f := range result.Findings {
		if f.Severity.IsThreat() {
			result.Safe = false
			break
		}
	}

	result.DurationMs = time.Since(start).Milliseconds()
	return result
}

// ScannerNames returns the names of all registered scanners.
func (o *Orchestrator) ScannerNames() []string {
	names := make([]string, len(o.scanners))
	for i, s := range o.scanners {
		names[i] = s.Name()
	}
	return names
}

// ScannerCount returns the number of registered scanners.
func (o *Orchestrator) ScannerCount() int {
	return len(o.scanners)
}
