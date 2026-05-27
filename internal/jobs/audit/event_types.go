package audit

// Severity levels for jobs audit events.
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

// Event category constants.
const (
	CategoryJob       = "job"
	CategorySchedule  = "schedule"
	CategoryQueue     = "queue"
	CategoryWorker    = "worker"
	CategoryDLQ       = "dlq"
)

// Event action constants.
const (
	ActionCreated   = "created"
	ActionStarted   = "started"
	ActionCompleted = "completed"
	ActionFailed    = "failed"
	ActionCancelled = "cancelled"
	ActionRetried   = "retried"
	ActionExpired   = "expired"
	ActionEnqueued  = "enqueued"
	ActionDequeued  = "dequeued"
)

// EventFilter provides structured filtering for jobs audit event queries.
type EventFilter struct {
	JobType   string
	EventType string
	Severity  string
	Category  string
	Status    string
	Limit     int
}

// DefaultEventFilter returns a filter with sensible defaults.
func DefaultEventFilter() *EventFilter {
	return &EventFilter{
		Limit: 100,
	}
}
