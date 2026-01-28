package events

// EventExtensions provides extended event functionality
// ResourceEvent, ObjectReference, EventRecorder, SimpleEventRecorder are defined in resource_events.go and recorder.go

// GlobalEventRecorder is the package-level event recorder
var GlobalEventRecorder EventRecorder

// CommonEventReasons defines common event reasons
var CommonEventReasons = struct {
	PolicyApplied     string
	PolicyDenied      string
	QuotaExceeded     string
	ValidationFailed  string
	ReconcileFailed   string
	ResourceCreated   string
	ResourceUpdated   string
	ResourceDeleted   string
	WorkflowStarted   string
	WorkflowCompleted string
	WorkflowFailed    string
}{
	PolicyApplied:     "PolicyApplied",
	PolicyDenied:      "PolicyDenied",
	QuotaExceeded:     "QuotaExceeded",
	ValidationFailed:  "ValidationFailed",
	ReconcileFailed:   "ReconcileFailed",
	ResourceCreated:   "ResourceCreated",
	ResourceUpdated:   "ResourceUpdated",
	ResourceDeleted:   "ResourceDeleted",
	WorkflowStarted:   "WorkflowStarted",
	WorkflowCompleted: "WorkflowCompleted",
	WorkflowFailed:    "WorkflowFailed",
}

func init() {
	// Initialize global recorder from primary recorder
	// This will be set by the main application
}
