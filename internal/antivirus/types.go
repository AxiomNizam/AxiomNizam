package antivirus

import (
	"fmt"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Scan Verdict
// ─────────────────────────────────────────────────────────────────────────────

// ScanVerdict represents the final outcome of a file scan.
// It is deliberately a typed string so that callers can use compile-time
// constant checks rather than comparing against arbitrary strings.
type ScanVerdict string

const (
	// VerdictClean indicates no threats were detected across all scan layers.
	VerdictClean ScanVerdict = "clean"

	// VerdictMalware indicates that one or more scan layers positively
	// identified a known or strongly-suspected malware signature.
	VerdictMalware ScanVerdict = "malware"

	// VerdictSuspicious indicates heuristic or entropy-based indicators
	// suggest the file may be malicious, but no definitive signature match
	// was found. Requires human review or additional analysis.
	VerdictSuspicious ScanVerdict = "suspicious"

	// VerdictError indicates the scan could not be completed due to an
	// internal error (e.g. corrupted file, resource exhaustion).
	VerdictError ScanVerdict = "error"
)

// IsTerminal returns true if the verdict represents a final, actionable state
// (i.e. not an error that should be retried).
func (v ScanVerdict) IsTerminal() bool {
	return v == VerdictClean || v == VerdictMalware || v == VerdictSuspicious
}

// IsThreat returns true if the verdict indicates the file should be
// quarantined or blocked.
func (v ScanVerdict) IsThreat() bool {
	return v == VerdictMalware || v == VerdictSuspicious
}

// String implements fmt.Stringer.
func (v ScanVerdict) String() string { return string(v) }

// ─────────────────────────────────────────────────────────────────────────────
// Threat Category
// ─────────────────────────────────────────────────────────────────────────────

// ThreatCategory classifies the general family of a detected threat.
type ThreatCategory string

const (
	CategoryTrojan      ThreatCategory = "trojan"
	CategoryWorm        ThreatCategory = "worm"
	CategoryRansomware  ThreatCategory = "ransomware"
	CategoryExploit     ThreatCategory = "exploit"
	CategoryBackdoor    ThreatCategory = "backdoor"
	CategoryCryptominer ThreatCategory = "cryptominer"
	CategoryWebshell    ThreatCategory = "webshell"
	CategoryDropper     ThreatCategory = "dropper"
	CategoryRootkit     ThreatCategory = "rootkit"
	CategoryAdware      ThreatCategory = "adware"
	CategorySpyware     ThreatCategory = "spyware"
	CategoryPacker      ThreatCategory = "packer"
	CategoryGeneric     ThreatCategory = "generic"
)

// String implements fmt.Stringer.
func (c ThreatCategory) String() string { return string(c) }

// ─────────────────────────────────────────────────────────────────────────────
// Threat Severity
// ─────────────────────────────────────────────────────────────────────────────

// ThreatSeverity indicates how dangerous a detected threat is.
type ThreatSeverity string

const (
	SeverityCritical ThreatSeverity = "critical"
	SeverityHigh     ThreatSeverity = "high"
	SeverityMedium   ThreatSeverity = "medium"
	SeverityLow      ThreatSeverity = "low"
)

// Weight returns a numeric weight for severity comparison and aggregation.
// Higher values indicate more severe threats.
func (s ThreatSeverity) Weight() int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

// String implements fmt.Stringer.
func (s ThreatSeverity) String() string { return string(s) }

// ─────────────────────────────────────────────────────────────────────────────
// Detection Layer
// ─────────────────────────────────────────────────────────────────────────────

// DetectionLayer identifies which scan layer produced a finding.
type DetectionLayer string

const (
	LayerHashDB    DetectionLayer = "hashdb"
	LayerPattern   DetectionLayer = "pattern"
	LayerHeuristic DetectionLayer = "heuristic"
	LayerYARA      DetectionLayer = "yara"
	LayerEntropy   DetectionLayer = "entropy"
)

// String implements fmt.Stringer.
func (l DetectionLayer) String() string { return string(l) }

// ─────────────────────────────────────────────────────────────────────────────
// Threat Info
// ─────────────────────────────────────────────────────────────────────────────

// ThreatInfo describes a single threat detection reported by one of the scan
// layers. Multiple ThreatInfo values may be produced for a single file (e.g.
// the file matches a known hash AND triggers a heuristic rule).
type ThreatInfo struct {
	// Name is the threat identifier, typically in the format
	// "Category.Platform.Family.Variant" (e.g. "Trojan.Win32.Emotet.A").
	Name string `json:"name"`

	// Category classifies the threat family (trojan, ransomware, etc.).
	Category ThreatCategory `json:"category"`

	// Severity indicates the danger level of the threat.
	Severity ThreatSeverity `json:"severity"`

	// Layer identifies which scan layer produced this detection.
	Layer DetectionLayer `json:"layer"`

	// Description is a human-readable explanation of what was detected.
	Description string `json:"description"`

	// Signature is the specific signature identifier that matched.
	// May be empty for heuristic or entropy-based detections.
	Signature string `json:"signature,omitempty"`

	// Confidence is a value between 0.0 and 1.0 indicating how certain
	// the detection is. Hash-DB lookups are always 1.0; heuristic
	// detections may be lower.
	Confidence float64 `json:"confidence"`

	// Offset is the byte offset within the file where the match occurred.
	// Set to -1 if not applicable (e.g. hash-based detection).
	Offset int64 `json:"offset,omitempty"`

	// Metadata carries layer-specific additional information (e.g. YARA
	// rule tags, entropy values, PE section names).
	Metadata map[string]string `json:"metadata,omitempty"`
}

// String returns a compact human-readable representation.
func (t ThreatInfo) String() string {
	return fmt.Sprintf("[%s] %s (%s, confidence=%.1f%%)", t.Layer, t.Name, t.Severity, t.Confidence*100)
}

// ─────────────────────────────────────────────────────────────────────────────
// Scan Result
// ─────────────────────────────────────────────────────────────────────────────

// ScanResult is the aggregated outcome of running all enabled scan layers
// against a single file. It is the primary value returned by Engine.Scan().
type ScanResult struct {
	// Verdict is the overall scan determination.
	Verdict ScanVerdict `json:"verdict"`

	// Threats lists every individual detection across all layers. Empty if
	// the verdict is VerdictClean.
	Threats []ThreatInfo `json:"threats"`

	// SHA256 is the hex-encoded SHA-256 digest of the scanned file.
	SHA256 string `json:"sha256"`

	// FileSize is the size of the scanned file in bytes.
	FileSize int64 `json:"fileSize"`

	// FileType is the detected MIME type of the file (via magic bytes).
	FileType string `json:"fileType"`

	// ScannedAt is the UTC timestamp when the scan was initiated.
	ScannedAt time.Time `json:"scannedAt"`

	// DurationMs is the wall-clock scan duration in milliseconds.
	DurationMs int64 `json:"durationMs"`

	// LayersRun lists the detection layers that were executed during the
	// scan, in execution order.
	LayersRun []string `json:"layersRun"`

	// CacheHit indicates whether the result was served from the scan
	// cache rather than performing a fresh scan.
	CacheHit bool `json:"cacheHit"`

	// EngineVersion is the version string of the antivirus engine.
	EngineVersion string `json:"engineVersion"`

	// SigDBVersion is the version string of the loaded signature database.
	SigDBVersion string `json:"sigDbVersion"`
}

// HighestSeverity returns the highest severity found among all threats, or
// an empty string if no threats exist.
func (r *ScanResult) HighestSeverity() ThreatSeverity {
	var highest ThreatSeverity
	for _, t := range r.Threats {
		if t.Severity.Weight() > highest.Weight() {
			highest = t.Severity
		}
	}
	return highest
}

// ThreatNames returns a deduplicated, sorted list of threat names.
func (r *ScanResult) ThreatNames() []string {
	seen := make(map[string]struct{}, len(r.Threats))
	names := make([]string, 0, len(r.Threats))
	for _, t := range r.Threats {
		if _, exists := seen[t.Name]; !exists {
			seen[t.Name] = struct{}{}
			names = append(names, t.Name)
		}
	}
	return names
}

// Summary returns a one-line human-readable summary of the scan result.
func (r *ScanResult) Summary() string {
	if r.Verdict == VerdictClean {
		return fmt.Sprintf("clean — %d layers checked in %dms", len(r.LayersRun), r.DurationMs)
	}
	return fmt.Sprintf("%s — %d threat(s) found [%s] in %dms",
		r.Verdict, len(r.Threats), strings.Join(r.ThreatNames(), ", "), r.DurationMs)
}

// ─────────────────────────────────────────────────────────────────────────────
// Scan Layer Interface
// ─────────────────────────────────────────────────────────────────────────────

// ScanLayer is the interface that every detection layer (hashdb, pattern,
// heuristic, yara, entropy) must implement. It mirrors the existing
// internal/scanner.Scanner interface pattern but is purpose-built for
// antivirus detection.
type ScanLayer interface {
	// Name returns a unique, stable identifier for this layer (e.g.
	// "hashdb", "pattern", "heuristic").
	Name() string

	// Scan examines the provided file data and returns any detected
	// threats. Implementations must be safe for concurrent use.
	// A nil/empty return with no error means no threats were found by
	// this layer.
	Scan(file *ScanTarget) ([]ThreatInfo, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Scan Target
// ─────────────────────────────────────────────────────────────────────────────

// ScanTarget carries all the information needed by scan layers to analyse
// a file. It is intentionally separate from the scanner.FileInfo type to
// decouple the antivirus engine from the existing file-scanner module.
type ScanTarget struct {
	// Filename is the original name of the uploaded file.
	Filename string

	// SHA256 is the hex-encoded SHA-256 digest, pre-computed by the
	// engine before layers are invoked (avoids duplicate hashing).
	SHA256 string

	// Size is the file size in bytes.
	Size int64

	// MIMEType is the detected MIME type from magic bytes.
	MIMEType string

	// Content is the raw file content. For files larger than
	// Config.MaxFileSize this may be truncated to the configured limit,
	// and FullContentAvailable will be false.
	Content []byte

	// FullContentAvailable indicates whether Content contains the
	// entirety of the file. When false, layers that require full-file
	// analysis (e.g. hash-DB) should skip or use the SHA256 field.
	FullContentAvailable bool
}

// ─────────────────────────────────────────────────────────────────────────────
// Engine Stats
// ─────────────────────────────────────────────────────────────────────────────

// EngineStats holds runtime statistics for the antivirus engine, exposed
// via the /antivirus/stats API endpoint.
type EngineStats struct {
	TotalScanned   int64   `json:"totalScanned"`
	ThreatsFound   int64   `json:"threatsFound"`
	CleanFiles     int64   `json:"cleanFiles"`
	ErrorCount     int64   `json:"errorCount"`
	CacheHits      int64   `json:"cacheHits"`
	CacheMisses    int64   `json:"cacheMisses"`
	CacheHitRate   float64 `json:"cacheHitRate"`
	AvgScanMs      float64 `json:"avgScanMs"`
	BytesScanned   int64   `json:"bytesScanned"`
	UptimeSeconds  int64   `json:"uptimeSeconds"`
	SigDBVersion   string  `json:"sigDbVersion"`
	EngineVersion  string  `json:"engineVersion"`
	LayersEnabled  []string `json:"layersEnabled"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Quarantine Action
// ─────────────────────────────────────────────────────────────────────────────

// QuarantineAction defines what happens when a threat is detected.
type QuarantineAction string

const (
	// QuarantineTag adds metadata tags to the object but does not move or
	// delete it. This is the safest default — allows manual review.
	QuarantineTag QuarantineAction = "tag"

	// QuarantineDelete permanently removes the infected object.
	QuarantineDelete QuarantineAction = "delete"

	// QuarantineMove moves the infected object to a quarantine bucket.
	QuarantineMove QuarantineAction = "move"
)

// String implements fmt.Stringer.
func (a QuarantineAction) String() string { return string(a) }

// ParseQuarantineAction converts a string to QuarantineAction, defaulting
// to QuarantineTag for unrecognised values.
func ParseQuarantineAction(s string) QuarantineAction {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "delete":
		return QuarantineDelete
	case "move":
		return QuarantineMove
	default:
		return QuarantineTag
	}
}
