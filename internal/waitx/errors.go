package waitx

import (
	"errors"
	"fmt"
)

var (
	// ErrCheckerRequired is returned when a nil checker is passed to Wait.
	ErrCheckerRequired = errors.New("waitx: checker is required")

	// ErrCheckTimedOut is returned when a check times out.
	ErrCheckTimedOut = errors.New("waitx: check timed out")

	// ErrCheckAborted is returned when a check is aborted (e.g., context cancelled).
	ErrCheckAborted = errors.New("waitx: check aborted")

	// ErrInvalidConfig is returned when a check configuration is invalid.
	ErrInvalidConfig = errors.New("waitx: invalid configuration")

	// ErrUnsupportedCheckType is returned for unknown check types.
	ErrUnsupportedCheckType = errors.New("waitx: unsupported check type")

	// ErrTargetRequired is returned when a check target is empty.
	ErrTargetRequired = errors.New("waitx: target is required")
)

// CheckError wraps a check failure with checker name and context.
type CheckError struct {
	CheckerName string
	Attempt     int
	Err         error
}

func (e *CheckError) Error() string {
	if e.Attempt > 0 {
		return fmt.Sprintf("waitx: check %q failed (attempt %d): %v", e.CheckerName, e.Attempt, e.Err)
	}
	return fmt.Sprintf("waitx: check %q failed: %v", e.CheckerName, e.Err)
}

func (e *CheckError) Unwrap() error {
	return e.Err
}
