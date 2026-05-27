package metrics

// Metric label constants for consistent Prometheus label usage.
const (
	LabelBackend  = "backend"   // rabbitmq, kafka
	LabelStepType = "step_type" // publish, consume, transform, route, ack
	LabelStatus   = "status"    // success, failure, timeout, retry
)

// Common label value sets.
var (
	Backends  = []string{"rabbitmq", "kafka"}
	StepTypes = []string{"publish", "consume", "transform", "route", "ack"}
	Statuses  = []string{"success", "failure", "timeout", "retry"}
)
