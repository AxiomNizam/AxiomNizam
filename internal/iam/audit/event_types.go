package audit

// Severity levels for IAM audit events.
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

// Event category constants.
const (
	CategoryAuth       = "auth"
	CategoryToken      = "token"
	CategoryPermission = "permission"
	CategorySession    = "session"
	CategoryUser       = "user"
	CategoryRole       = "role"
	CategoryGroup      = "group"
	CategoryClient     = "client"
)

// Event action constants.
const (
	ActionCreated   = "created"
	ActionUpdated   = "updated"
	ActionDeleted   = "deleted"
	ActionLogin     = "login"
	ActionLogout    = "logout"
	ActionGranted   = "granted"
	ActionDenied    = "denied"
	ActionRevoked   = "revoked"
	ActionRefreshed = "refreshed"
	ActionAssigned  = "assigned"
	ActionRemoved   = "removed"
)

// EventFilter provides structured filtering for IAM audit event queries.
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
