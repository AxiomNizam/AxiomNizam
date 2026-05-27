package metrics

import "time"

// RecordMessageSent records a message sent to a backend.
func RecordMessageSent() { MessagesSent.Inc() }

// RecordMessageReceived records a message received from a backend.
func RecordMessageReceived() { MessagesReceived.Inc() }

// RecordMessageAcked records a message acknowledgement.
func RecordMessageAcked() { MessagesAcked.Inc() }

// RecordMessageFailed records a message processing failure.
func RecordMessageFailed() { MessagesFailed.Inc() }

// RecordMessageDLQ records a message sent to the dead-letter queue.
func RecordMessageDLQ() { MessagesDLQ.Inc() }

// RecordBackendConnection records a backend connection attempt.
func RecordBackendConnection() { BackendConnections.Inc() }

// RecordBackendError records a backend error.
func RecordBackendError(backend string) { BackendErrors.WithLabelValues(backend).Inc() }

// RecordWorkflowStarted records a workflow start.
func RecordWorkflowStarted() { WorkflowsStarted.Inc() }

// RecordWorkflowCompleted records a workflow completion.
func RecordWorkflowCompleted() { WorkflowsCompleted.Inc() }

// RecordWorkflowFailed records a workflow failure.
func RecordWorkflowFailed() { WorkflowsFailed.Inc() }

// RecordStepExecuted records a workflow step execution.
func RecordStepExecuted() { StepsExecuted.Inc() }

// RecordStepFailed records a workflow step failure.
func RecordStepFailed() { StepsFailed.Inc() }

// RecordMessageLatency records end-to-end message latency.
func RecordMessageLatency(d time.Duration) { MessageLatency.Observe(d.Seconds()) }

// RecordWorkflowDuration records total workflow execution time.
func RecordWorkflowDuration(d time.Duration) { WorkflowDuration.Observe(d.Seconds()) }

// RecordStepDuration records the duration of a specific step type.
func RecordStepDuration(stepType string, d time.Duration) {
	StepDuration.WithLabelValues(stepType).Observe(d.Seconds())
}

// SetActiveProducers sets the current active producer count.
func SetActiveProducers(count float64) { ActiveProducers.Set(count) }

// SetActiveConsumers sets the current active consumer count.
func SetActiveConsumers(count float64) { ActiveConsumers.Set(count) }

// SetDLQSize sets the current dead-letter queue size.
func SetDLQSize(size float64) { DLQSize.Set(size) }

// SetActiveMessages sets the current in-flight message count.
func SetActiveMessages(count float64) { ActiveMessages.Set(count) }

// SetActiveWorkflows sets the current running workflow count.
func SetActiveWorkflows(count float64) { ActiveWorkflows.Set(count) }
