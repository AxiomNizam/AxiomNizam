package scanner

import (
	"context"
	"fmt"
	"time"

	"example.com/axiomnizam/internal/antivirus"
)

// NativeAVScanner wraps the internal antivirus.Engine as a scanner.Scanner
// implementation. This is the drop-in replacement for ClamAVScanner in
// the SafeGate scanner orchestrator pipeline.
type NativeAVScanner struct {
	engine *antivirus.Engine
}

// NewNativeAVScanner creates a scanner.Scanner backed by the internal
// antivirus engine. If engine is nil, all scans return clean (no-op).
func NewNativeAVScanner(engine *antivirus.Engine) *NativeAVScanner {
	return &NativeAVScanner{engine: engine}
}

func (s *NativeAVScanner) Name() string { return "native_antivirus" }

func (s *NativeAVScanner) Scan(ctx context.Context, file *FileInfo) ([]Finding, error) {
	if s.engine == nil || !s.engine.IsRunning() {
		return nil, nil // engine not available — skip
	}

	// Safety-net timeout when caller provides no deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
	}

	result, err := s.engine.Scan(ctx, file.Content, file.Filename)
	if err != nil {
		return nil, fmt.Errorf("native antivirus scan failed: %w", err)
	}

	var findings []Finding
	if result.Verdict.IsThreat() {
		for _, t := range result.Threats {
			sev := SeverityMedium
			switch t.Severity {
			case antivirus.SeverityCritical:
				sev = SeverityCritical
			case antivirus.SeverityHigh:
				sev = SeverityHigh
			case antivirus.SeverityMedium:
				sev = SeverityMedium
			case antivirus.SeverityLow:
				sev = SeverityLow
			}

			findings = append(findings, Finding{
				Scanner:     s.Name(),
				Severity:    sev,
				Description: fmt.Sprintf("Malware detected: %s", t.Name),
				Details:     fmt.Sprintf("[%s] %s — category=%s, confidence=%.2f, sha256=%s", t.Severity, t.Description, t.Category, t.Confidence, result.SHA256),
			})
		}
	}

	return findings, nil
}
