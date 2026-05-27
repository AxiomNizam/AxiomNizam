package metrics

// Metric label constants for consistent Prometheus label usage.
const (
	LabelJobType   = "job_type"   // email, webhook, scan, export, etc.
	LabelJobStatus = "job_status" // pending, running, completed, failed, cancelled
	LabelOutcome   = "outcome"    // success, failure, timeout, retry
	LabelQueue     = "queue"      // default, priority, dlq
)

// Common label value sets.
var (
	JobTypes   = []string{"email", "webhook", "scan", "export", "import", "cleanup", "reconcile"}
	JobStatuses = []string{"pending", "running", "completed", "failed", "cancelled", "retrying"}
	Outcomes   = []string{"success", "failure", "timeout", "retry"}
	Queues     = []string{"default", "priority", "dlq"}
)
