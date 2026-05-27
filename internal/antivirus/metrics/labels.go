package metrics

// Metric label constants for consistent Prometheus label usage.
const (
	LabelVerdict = "verdict" // clean, malware, suspicious, error
	LabelLayer   = "layer"   // hashdb, pattern, heuristic, yara, entropy
	LabelThreat  = "threat"  // threat category name
)

// Common label value sets.
var (
	Verdicts = []string{"clean", "malware", "suspicious", "error"}
	Layers   = []string{"hashdb", "pattern", "heuristic", "yara", "entropy"}
)
