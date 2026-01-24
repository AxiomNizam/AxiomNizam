package output

import (
	"fmt"
	"os"
)

type ErrorCode string

const (
	// Client errors
	ErrNotFound     ErrorCode = "NOT_FOUND"
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrForbidden    ErrorCode = "FORBIDDEN"
	ErrInvalidInput ErrorCode = "INVALID_INPUT"
	ErrConflict     ErrorCode = "CONFLICT"
	ErrInvalidYAML  ErrorCode = "INVALID_YAML"

	// Server errors
	ErrServerError ErrorCode = "SERVER_ERROR"
	ErrTimeout     ErrorCode = "TIMEOUT"
	ErrUnavailable ErrorCode = "UNAVAILABLE"
	ErrInternal    ErrorCode = "INTERNAL_ERROR"
)

var ErrorMessages = map[ErrorCode]string{
	ErrNotFound:     "Resource not found",
	ErrUnauthorized: "Authentication failed",
	ErrForbidden:    "Permission denied",
	ErrInvalidInput: "Invalid input provided",
	ErrConflict:     "Resource conflict or already exists",
	ErrInvalidYAML:  "Invalid YAML format",
	ErrServerError:  "Server error occurred",
	ErrTimeout:      "Request timed out",
	ErrUnavailable:  "Service unavailable",
	ErrInternal:     "Internal server error",
}

var ErrorSuggestions = map[ErrorCode]string{
	ErrNotFound:     "Check the resource name and namespace with 'list' command",
	ErrUnauthorized: "Run 'axiomnizamctl login' to authenticate",
	ErrForbidden:    "Check your RBAC policies and permissions",
	ErrInvalidInput: "Verify your input parameters and flags",
	ErrConflict:     "Resource may already exist, try 'describe' to check",
	ErrInvalidYAML:  "Check YAML syntax and indentation",
	ErrServerError:  "Check server logs and status",
	ErrTimeout:      "Server may be overloaded, try again",
	ErrUnavailable:  "Server is unavailable, check connectivity",
	ErrInternal:     "Contact support with error details",
}

// PrintError prints a formatted error message
func PrintError(code ErrorCode, details string) {
	fmt.Fprintf(os.Stderr, "❌ [%s] %s\n", code, ErrorMessages[code])
	if details != "" {
		fmt.Fprintf(os.Stderr, "   Details: %s\n", details)
	}
	if suggestion, ok := ErrorSuggestions[code]; ok {
		fmt.Fprintf(os.Stderr, "   💡 %s\n", suggestion)
	}
}

// PrintSuccess prints a success message
func PrintSuccess(msg string) {
	fmt.Printf("✅ %s\n", msg)
}

// PrintWarning prints a warning message
func PrintWarning(msg string) {
	fmt.Printf("⚠️  %s\n", msg)
}

// PrintInfo prints an info message
func PrintInfo(msg string) {
	fmt.Printf("ℹ️  %s\n", msg)
}
