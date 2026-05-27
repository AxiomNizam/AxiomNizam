package etl

import (
	"errors"
	"fmt"
)

var (
	// ErrPipelineNotFound is returned when an ETL pipeline doesn't exist.
	ErrPipelineNotFound = errors.New("etl: pipeline not found")

	// ErrPipelineAlreadyRunning is returned when starting an already-running pipeline.
	ErrPipelineAlreadyRunning = errors.New("etl: pipeline already running")

	// ErrPipelineNotRunning is returned when pausing/stopping a non-running pipeline.
	ErrPipelineNotRunning = errors.New("etl: pipeline not running")

	// ErrInvalidStep is returned when a pipeline step configuration is invalid.
	ErrInvalidStep = errors.New("etl: invalid step configuration")

	// ErrStepFailed is returned when a pipeline step execution fails.
	ErrStepFailed = errors.New("etl: step execution failed")

	// ErrConnectorNotFound is returned when a connector type doesn't exist.
	ErrConnectorNotFound = errors.New("etl: connector not found")

	// ErrRunNotFound is returned when a pipeline run doesn't exist.
	ErrRunNotFound = errors.New("etl: run not found")

	// ErrTransformFailed is returned when a data transformation fails.
	ErrTransformFailed = errors.New("etl: transform failed")
)

// PipelineError wraps an ETL pipeline error with context.
type PipelineError struct {
	Op         string // operation (e.g., "CreatePipeline", "RunPipeline")
	PipelineID string
	Err        error
}

func (e *PipelineError) Error() string {
	if e.PipelineID != "" {
		return fmt.Sprintf("etl: %s pipeline %q: %v", e.Op, e.PipelineID, e.Err)
	}
	return fmt.Sprintf("etl: %s: %v", e.Op, e.Err)
}

func (e *PipelineError) Unwrap() error {
	return e.Err
}
