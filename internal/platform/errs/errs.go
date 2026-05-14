// Package errs is the canonical home for cross-cutting sentinel errors
// and error-wrapping helpers.
//
// P3.2 — individual packages (apiserver, platform/store, rbac, jobs)
// each defined their own `ErrNotFound` / `ErrAlreadyExists` /
// `ErrConflict` constants.  Callers have to know which flavour to
// `errors.Is` against, which is brittle.
//
// New code should compare against the sentinels here and wrap with
// `fmt.Errorf("...: %w", errs.ErrNotFound)` (or the Wrap helpers
// below).  Legacy packages can alias their local constants to these,
// which keeps existing `errors.Is` call sites working.
package errs

import (
	"errors"
	"fmt"
)

// Canonical sentinels.  These are intentionally small and orthogonal.
var (
	// ErrNotFound is returned when a lookup by key yields nothing.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists is returned when a Create would overwrite a
	// resource that is already persisted.
	ErrAlreadyExists = errors.New("already exists")

	// ErrConflict is returned when an optimistic-concurrency or
	// resource-version check fails.
	ErrConflict = errors.New("conflict")

	// ErrInvalidInput covers validation failures: missing fields,
	// malformed identifiers, out-of-range values, etc.
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned for authentication failures.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned for authorisation failures (authn ok,
	// authz denied).
	ErrForbidden = errors.New("forbidden")

	// ErrUnavailable is returned when a downstream dependency is
	// temporarily unreachable and the caller should retry.
	ErrUnavailable = errors.New("unavailable")

	// ErrInternal is returned for unexpected conditions the caller
	// cannot recover from.
	ErrInternal = errors.New("internal error")
)

// NotFoundf wraps ErrNotFound with a formatted message.
func NotFoundf(format string, a ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, a...), ErrNotFound)
}

// AlreadyExistsf wraps ErrAlreadyExists with a formatted message.
func AlreadyExistsf(format string, a ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, a...), ErrAlreadyExists)
}

// Conflictf wraps ErrConflict with a formatted message.
func Conflictf(format string, a ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, a...), ErrConflict)
}

// InvalidInputf wraps ErrInvalidInput with a formatted message.
func InvalidInputf(format string, a ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, a...), ErrInvalidInput)
}

// Is is a convenience alias for errors.Is so call sites can import a
// single symbol from this package.
func Is(err, target error) bool { return errors.Is(err, target) }

// As is a convenience alias for errors.As.
func As(err error, target any) bool { return errors.As(err, target) }
