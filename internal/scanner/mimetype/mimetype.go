// Package mimetype provides the MIMEScanner for the SafeGate pipeline.
// It detects file type spoofing by comparing actual MIME type (from content
// magic bytes) against the claimed extension, and checks for executable
// signatures hidden in non-executable files.
package mimetype

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"example.com/axiomnizam/internal/scanner"
)

// Scanner detects file type spoofing by comparing the actual MIME type
// (detected from file content magic bytes) against the claimed extension.
type Scanner struct {
	allowedTypes map[string]bool
}

// NewScanner creates a MIMEScanner with the given list of allowed MIME types.
func NewScanner(allowedTypes []string) *Scanner {
	m := make(map[string]bool, len(allowedTypes))
	for _, t := range allowedTypes {
		m[t] = true
	}
	return &Scanner{allowedTypes: m}
}

func (s *Scanner) Name() string { return "mime_type_validator" }

func (s *Scanner) Scan(_ context.Context, file *scanner.FileInfo) ([]scanner.Finding, error) {
	var findings []scanner.Finding

	// Detect actual MIME type from content bytes
	detected := http.DetectContentType(file.Content)
	detected = strings.Split(detected, ";")[0]
	detected = strings.TrimSpace(detected)

	// Check if the detected type is in the allowed list
	if !s.allowedTypes[detected] && detected != "application/octet-stream" {
		findings = append(findings, scanner.Finding{
			Scanner:     s.Name(),
			Severity:    scanner.SeverityHigh,
			Description: "Disallowed file type detected",
			Details:     fmt.Sprintf("Detected MIME type %q is not in the allowed list", detected),
		})
	}

	// Check for type spoofing: extension says one thing, content says another
	if file.MIMEType != "" && detected != "application/octet-stream" {
		claimedBase := strings.Split(file.MIMEType, ";")[0]
		claimedBase = strings.TrimSpace(claimedBase)
		if !typesCompatible(claimedBase, detected) {
			findings = append(findings, scanner.Finding{
				Scanner:     s.Name(),
				Severity:    scanner.SeverityCritical,
				Description: "File type spoofing detected",
				Details: fmt.Sprintf(
					"File claims to be %q but content detected as %q — possible disguised payload",
					claimedBase, detected,
				),
			})
		}
	}

	// Check for executable content in non-executable files
	if isExecutableSignature(file.Content) && !strings.Contains(detected, "executable") {
		findings = append(findings, scanner.Finding{
			Scanner:     s.Name(),
			Severity:    scanner.SeverityCritical,
			Description: "Executable content detected in non-executable file",
			Details:     "File contains executable signatures (PE/ELF/Mach-O) but is not declared as executable",
		})
	}

	return findings, nil
}

// CompatMap defines which MIME types are compatible with each other.
// This is exported so callers can inspect or extend the map.
var CompatMap = map[string][]string{
	"application/msword": {"application/zip", "application/x-cfbf", "application/octet-stream"},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {"application/zip", "application/octet-stream"},
	"application/vnd.ms-excel": {"application/zip", "application/x-cfbf", "application/octet-stream"},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {"application/zip", "application/octet-stream"},
	"application/vnd.ms-powerpoint":                                             {"application/zip", "application/x-cfbf", "application/octet-stream"},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {"application/zip", "application/octet-stream"},
	"image/svg+xml":                {"text/xml", "application/xml", "text/plain", "text/html"},
	"text/csv":                     {"text/plain", "application/octet-stream"},
	"application/json":             {"text/plain"},
	"application/x-rar-compressed": {"application/octet-stream"},
	"application/x-7z-compressed":  {"application/octet-stream"},
}

func typesCompatible(claimed, detected string) bool {
	claimed = strings.ToLower(claimed)
	detected = strings.ToLower(detected)

	if claimed == detected {
		return true
	}

	if compatible, ok := CompatMap[claimed]; ok {
		for _, c := range compatible {
			if detected == c {
				return true
			}
		}
	}

	if detected == "application/octet-stream" {
		return true
	}

	return false
}

// isExecutableSignature checks for PE, ELF, and Mach-O magic bytes.
func isExecutableSignature(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	if data[0] == 'M' && data[1] == 'Z' {
		return true
	}
	if data[0] == 0x7F && data[1] == 'E' && data[2] == 'L' && data[3] == 'F' {
		return true
	}
	if (data[0] == 0xFE && data[1] == 0xED && data[2] == 0xFA && data[3] == 0xCE) ||
		(data[0] == 0xFE && data[1] == 0xED && data[2] == 0xFA && data[3] == 0xCF) ||
		(data[0] == 0xCE && data[1] == 0xFA && data[2] == 0xED && data[3] == 0xFE) ||
		(data[0] == 0xCF && data[1] == 0xFA && data[2] == 0xED && data[3] == 0xFE) {
		return true
	}
	return false
}
