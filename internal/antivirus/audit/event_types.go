package audit

// Severity levels for antivirus audit events.
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

// Event category constants.
const (
	CategoryScan     = "scan"
	CategoryThreat   = "threat"
	CategoryEngine   = "engine"
	CategorySignature = "signature"
	CategoryCache    = "cache"
)

// Event action constants.
const (
	ActionScanned    = "scanned"
	ActionDetected   = "detected"
	ActionClean      = "clean"
	ActionError      = "error"
	ActionStarted    = "started"
	ActionStopped    = "stopped"
	ActionReloaded   = "reloaded"
	ActionEvicted    = "evicted"
)

// EventFilter provides structured filtering for antivirus audit event queries.
type EventFilter struct {
	EventType string
	Severity  string
	Category  string
	Filename  string
	Limit     int
}

// DefaultEventFilter returns a filter with sensible defaults.
func DefaultEventFilter() *EventFilter {
	return &EventFilter{
		Limit: 100,
	}
}
