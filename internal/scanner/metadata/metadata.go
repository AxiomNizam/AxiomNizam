// Package metadata provides the MetadataScanner for the SafeGate pipeline.
// It checks file size limits, empty files, null byte injection in text files,
// and suspicious filename patterns (excessive length, double extensions).
package metadata

import (
	"context"
	"fmt"

	"example.com/axiomnizam/internal/scanner"
)

// Scanner checks file size, empty files, null bytes, and suspicious filenames.
type Scanner struct {
	maxFileSize int64
}

// NewScanner creates a MetadataScanner with the given maximum file size.
func NewScanner(maxFileSize int64) *Scanner {
	return &Scanner{maxFileSize: maxFileSize}
}

func (s *Scanner) Name() string { return "metadata_scanner" }

func (s *Scanner) Scan(_ context.Context, file *scanner.FileInfo) ([]scanner.Finding, error) {
	var findings []scanner.Finding

	if file.Size > s.maxFileSize {
		findings = append(findings, scanner.Finding{
			Scanner: s.Name(), Severity: scanner.SeverityHigh,
			Description: "File exceeds maximum allowed size",
			Details:     fmt.Sprintf("Size %d bytes exceeds %d byte limit", file.Size, s.maxFileSize),
		})
	}

	if file.Size == 0 {
		findings = append(findings, scanner.Finding{
			Scanner: s.Name(), Severity: scanner.SeverityLow,
			Description: "Empty file uploaded", Details: "File has zero bytes",
		})
	}

	// Null bytes in text files may indicate binary injection
	if isTextFile(file) {
		nullCount := 0
		for _, b := range file.Content {
			if b == 0x00 {
				nullCount++
			}
		}
		if nullCount > 0 {
			findings = append(findings, scanner.Finding{
				Scanner: s.Name(), Severity: scanner.SeverityMedium,
				Description: "Text file contains null bytes",
				Details:     fmt.Sprintf("Found %d null bytes — may indicate binary injection", nullCount),
			})
		}
	}

	findings = append(findings, checkFilename(file.Filename)...)
	return findings, nil
}

func isTextFile(file *scanner.FileInfo) bool {
	switch file.Extension {
	case ".txt", ".csv", ".json", ".xml", ".svg", ".html", ".htm":
		return true
	}
	return false
}

func checkFilename(name string) []scanner.Finding {
	var findings []scanner.Finding

	if len(name) > 255 {
		findings = append(findings, scanner.Finding{
			Scanner: "metadata_scanner", Severity: scanner.SeverityMedium,
			Description: "Filename exceeds maximum length",
			Details:     fmt.Sprintf("Filename is %d characters (max 255)", len(name)),
		})
	}

	// Double extensions: e.g. "report.pdf.exe"
	dots := 0
	for _, c := range name {
		if c == '.' {
			dots++
		}
	}
	if dots > 2 {
		findings = append(findings, scanner.Finding{
			Scanner: "metadata_scanner", Severity: scanner.SeverityMedium,
			Description: "File has many extensions",
			Details:     fmt.Sprintf("Filename %q has %d dots — possible extension spoofing", name, dots),
		})
	}

	return findings
}
