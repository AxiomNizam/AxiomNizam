package cdc

import (
	"errors"
	"fmt"
)

var (
	// ErrPipelineNotFound is returned when a CDC pipeline doesn't exist.
	ErrPipelineNotFound = errors.New("cdc: pipeline not found")

	// ErrPipelineAlreadyRunning is returned when starting an already-running pipeline.
	ErrPipelineAlreadyRunning = errors.New("cdc: pipeline already running")

	// ErrPipelineNotRunning is returned when pausing/stopping a non-running pipeline.
	ErrPipelineNotRunning = errors.New("cdc: pipeline not running")

	// ErrInvalidSource is returned when a CDC source configuration is invalid.
	ErrInvalidSource = errors.New("cdc: invalid source configuration")

	// ErrInvalidSink is returned when a CDC sink configuration is invalid.
	ErrInvalidSink = errors.New("cdc: invalid sink configuration")

	// ErrConnectorNotFound is returned when a connector type doesn't exist.
	ErrConnectorNotFound = errors.New("cdc: connector not found")

	// ErrStreamNotFound is returned when a CDC stream doesn't exist.
	ErrStreamNotFound = errors.New("cdc: stream not found")

	// ErrCaptureFailed is returned when a change capture fails.
	ErrCaptureFailed = errors.New("cdc: capture failed")
)

// PipelineError wraps a CDC pipeline error with context.
type PipelineError struct {
	Op         string // operation (e.g., "CreatePipeline", "StartPipeline")
	PipelineID string
	Err        error
}

func (e *PipelineError) Error() string {
	if e.PipelineID != "" {
		return fmt.Sprintf("cdc: %s pipeline %q: %v", e.Op, e.PipelineID, e.Err)
	}
	return fmt.Sprintf("cdc: %s: %v", e.Op, e.Err)
}

func (e *PipelineError) Unwrap() error {
	return e.Err
}
