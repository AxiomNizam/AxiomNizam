// Package archivescan provides the ArchiveScanner for the SafeGate pipeline.
// It detects zip bombs (via compression ratio and decompressed size analysis),
// path traversal attacks, and executable files hidden inside archives.
package archivescan

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"strings"

	"example.com/axiomnizam/internal/scanner"
)

// Scanner detects zip bombs, path traversal, and malicious archives.
type Scanner struct {
	maxDepth        int
	maxDecompressed int64
}

// NewScanner creates an ArchiveScanner with the given depth and size limits.
func NewScanner(maxDepth int, maxDecompressed int64) *Scanner {
	return &Scanner{maxDepth: maxDepth, maxDecompressed: maxDecompressed}
}

func (s *Scanner) Name() string { return "archive_bomb_scanner" }

func (s *Scanner) Scan(_ context.Context, file *scanner.FileInfo) ([]scanner.Finding, error) {
	if !isArchive(file) {
		return nil, nil
	}

	var findings []scanner.Finding

	if isZipArchive(file) {
		f, err := s.analyzeZip(file.Content, 0)
		if err != nil {
			findings = append(findings, scanner.Finding{
				Scanner: s.Name(), Severity: scanner.SeverityMedium,
				Description: "Failed to analyze archive", Details: err.Error(),
			})
		}
		findings = append(findings, f...)
	}

	return findings, nil
}

func (s *Scanner) analyzeZip(data []byte, depth int) ([]scanner.Finding, error) {
	var findings []scanner.Finding

	if depth > s.maxDepth {
		findings = append(findings, scanner.Finding{
			Scanner: s.Name(), Severity: scanner.SeverityHigh,
			Description: "Archive exceeds maximum nesting depth",
			Details:     fmt.Sprintf("Depth %d exceeds limit %d — possible zip bomb", depth, s.maxDepth),
		})
		return findings, nil
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to read zip: %w", err)
	}

	var totalUncompressed uint64
	var fileCount int

	for _, f := range reader.File {
		fileCount++
		totalUncompressed += f.UncompressedSize64

		if f.CompressedSize64 > 0 {
			ratio := float64(f.UncompressedSize64) / float64(f.CompressedSize64)
			if ratio > 100 {
				findings = append(findings, scanner.Finding{
					Scanner: s.Name(), Severity: scanner.SeverityCritical,
					Description: "Possible zip bomb detected",
					Details:     fmt.Sprintf("File %q has %.0f:1 compression ratio", f.Name, ratio),
				})
			}
		}

		if int64(totalUncompressed) > s.maxDecompressed {
			findings = append(findings, scanner.Finding{
				Scanner: s.Name(), Severity: scanner.SeverityCritical,
				Description: "Archive decompressed size exceeds limit",
				Details:     fmt.Sprintf("Total %d bytes exceeds %d byte limit", totalUncompressed, s.maxDecompressed),
			})
			return findings, nil
		}

		if strings.Contains(f.Name, "..") {
			findings = append(findings, scanner.Finding{
				Scanner: s.Name(), Severity: scanner.SeverityCritical,
				Description: "Archive contains path traversal",
				Details:     fmt.Sprintf("Entry %q contains '..' — directory traversal attack", f.Name),
			})
		}

		if isExecutableExtension(strings.ToLower(f.Name)) {
			findings = append(findings, scanner.Finding{
				Scanner: s.Name(), Severity: scanner.SeverityHigh,
				Description: "Archive contains executable file",
				Details:     fmt.Sprintf("Found executable: %q", f.Name),
			})
		}
	}

	if fileCount > 10000 {
		findings = append(findings, scanner.Finding{
			Scanner: s.Name(), Severity: scanner.SeverityHigh,
			Description: "Archive contains excessive files",
			Details:     fmt.Sprintf("%d files — possible resource exhaustion", fileCount),
		})
	}

	return findings, nil
}

func isArchive(file *scanner.FileInfo) bool {
	ext := strings.ToLower(file.Extension)
	switch ext {
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz",
		".docx", ".xlsx", ".pptx", ".docm", ".xlsm", ".pptm", ".jar":
		return true
	}
	mime := strings.ToLower(file.MIMEType)
	return strings.Contains(mime, "zip") || strings.Contains(mime, "rar") ||
		strings.Contains(mime, "7z") || strings.Contains(mime, "compressed")
}

func isZipArchive(file *scanner.FileInfo) bool {
	if len(file.Content) >= 4 &&
		file.Content[0] == 0x50 && file.Content[1] == 0x4B &&
		file.Content[2] == 0x03 && file.Content[3] == 0x04 {
		return true
	}
	ext := strings.ToLower(file.Extension)
	switch ext {
	case ".zip", ".docx", ".xlsx", ".pptx", ".docm", ".xlsm", ".pptm", ".jar":
		return true
	}
	return false
}

func isExecutableExtension(name string) bool {
	exts := []string{".exe", ".bat", ".cmd", ".com", ".msi", ".scr", ".pif",
		".sh", ".bash", ".ps1", ".vbs", ".vbe", ".js", ".wsh", ".wsf"}
	for _, ext := range exts {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
