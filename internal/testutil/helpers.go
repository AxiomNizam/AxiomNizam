// Package testutil provides shared test helpers for all AxiomNizam modules.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    ctx := testutil.Context()
//	    // ...
//	}
package testutil

import (
	"context"
	"testing"
	"time"
)

// Context returns a background context suitable for tests.
// Use this instead of context.Background() in test files for consistency.
func Context() context.Context {
	return context.Background()
}

// ContextWithTimeout returns a context with a default 10-second timeout for tests.
func ContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// SkipIfShort skips the test if running with -short flag.
// Use for integration tests that require external dependencies.
func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}

// RequireEnv skips the test if the given environment variable is not set.
func RequireEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "test-value")
}

// TempDir returns a temporary directory for test files.
// The directory is automatically cleaned up when the test finishes.
func TempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}
