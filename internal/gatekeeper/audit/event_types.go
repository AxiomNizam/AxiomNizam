package audit

// Severity levels for audit events.
const (
	SeverityInfo    = "info"
	SeverityWarning = "warning"
	SeverityError   = "error"
	SeverityCritical = "critical"
)

// Event category constants for filtering and reporting.
const (
	CategoryEnrollment  = "enrollment"
	CategoryVerification = "verification"
	CategoryDevice      = "device"
	CategoryRisk        = "risk"
	CategoryBackupCode  = "backup_code"
	CategoryFactor      = "factor"
)

// Event action constants.
const (
	ActionCreated   = "created"
	ActionVerified  = "verified"
	ActionFailed    = "failed"
	ActionDisabled  = "disabled"
	ActionRevoked   = "revoked"
	ActionExpired   = "expired"
	ActionUsed      = "used"
	ActionTrusted   = "trusted"
	ActionDetected  = "detected"
)

// EventFilter provides structured filtering for audit event queries.
type EventFilter struct {
	UserID    string
	EventType string
	Severity  string
	Category  string
	Action    string
	Limit     int
	Offset    int
}

// DefaultEventFilter returns a filter with sensible defaults.
func DefaultEventFilter() *EventFilter {
	return &EventFilter{
		Limit: 100,
	}
}
