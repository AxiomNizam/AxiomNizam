package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType defines error classification
type ErrorType string

const (
	// User-facing errors
	ErrorTypeValidation   ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound     ErrorType = "NOT_FOUND"
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden    ErrorType = "FORBIDDEN"
	ErrorTypeConflict     ErrorType = "CONFLICT"
	ErrorTypeBadRequest   ErrorType = "BAD_REQUEST"

	// System/retryable errors
	ErrorTypeDatabase    ErrorType = "DATABASE_ERROR"
	ErrorTypeTimeout     ErrorType = "TIMEOUT"
	ErrorTypeInternal    ErrorType = "INTERNAL_SERVER_ERROR"
	ErrorTypeUnavailable ErrorType = "SERVICE_UNAVAILABLE"

	// External service errors
	ErrorTypeExternal ErrorType = "EXTERNAL_SERVICE_ERROR"
	ErrorTypeNetwork  ErrorType = "NETWORK_ERROR"
)

// AppError represents an application error with metadata
type AppError struct {
	Type       ErrorType
	Message    string
	StatusCode int
	Details    map[string]interface{}
	Cause      error // wrapped error
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithCause adds a wrapped error
func (e *AppError) WithCause(err error) *AppError {
	e.Cause = err
	return e
}

// WithDetails adds additional details
func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// IsRetryable returns true if error is retryable
func (e *AppError) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeDatabase, ErrorTypeTimeout, ErrorTypeUnavailable, ErrorTypeNetwork:
		return true
	default:
		return false
	}
}

// IsClientError returns true if error is client-caused
func (e *AppError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns true if error is server-caused
func (e *AppError) IsServerError() bool {
	return e.StatusCode >= 500
}

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Details:    make(map[string]interface{}),
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
		Details:    make(map[string]interface{}),
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Details:    make(map[string]interface{}),
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
		Details:    make(map[string]interface{}),
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
		Details:    make(map[string]interface{}),
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Details:    make(map[string]interface{}),
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeDatabase,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Details:    make(map[string]interface{}),
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeTimeout,
		Message:    message,
		StatusCode: http.StatusRequestTimeout,
		Details:    make(map[string]interface{}),
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Details:    make(map[string]interface{}),
	}
}

// NewUnavailableError creates a service unavailable error
func NewUnavailableError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeUnavailable,
		Message:    message,
		StatusCode: http.StatusServiceUnavailable,
		Details:    make(map[string]interface{}),
	}
}

// NewExternalError creates an external service error
func NewExternalError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeExternal,
		Message:    message,
		StatusCode: http.StatusBadGateway,
		Details:    make(map[string]interface{}),
	}
}

// NewNetworkError creates a network error
func NewNetworkError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeNetwork,
		Message:    message,
		StatusCode: http.StatusGatewayTimeout,
		Details:    make(map[string]interface{}),
	}
}

// As extracts AppError from wrapped error chain
func As(err error, target *AppError) bool {
	return errors.As(err, &target)
}

// Is checks if error is of a specific type
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// Wrap wraps an error with message and type
func Wrap(err error, errorType ErrorType, message string) *AppError {
	return &AppError{
		Type:       errorType,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Cause:      err,
		Details:    make(map[string]interface{}),
	}
}

// Wrapf wraps an error with formatted message
func Wrapf(err error, errorType ErrorType, format string, args ...interface{}) *AppError {
	return Wrap(err, errorType, fmt.Sprintf(format, args...))
}

// ClassifyError attempts to classify a generic error
func ClassifyError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	// Try to extract AppError from chain
	var appError *AppError
	if errors.As(err, &appError) {
		return appError
	}

	// Default to internal error
	return NewInternalError(err.Error())
}

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Type    string                 `json:"type"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ToResponse converts AppError to ErrorResponse for JSON marshaling
func (e *AppError) ToResponse() ErrorResponse {
	return ErrorResponse{
		Type:    string(e.Type),
		Message: e.Message,
		Code:    e.StatusCode,
		Details: e.Details,
	}
}

// FromResponse creates AppError from ErrorResponse
func FromResponse(resp ErrorResponse) *AppError {
	return &AppError{
		Type:       ErrorType(resp.Type),
		Message:    resp.Message,
		StatusCode: resp.Code,
		Details:    resp.Details,
	}
}

// ErrorChain represents a chain of errors for debugging
type ErrorChain struct {
	errors []error
}

// NewErrorChain creates a new error chain
func NewErrorChain() *ErrorChain {
	return &ErrorChain{
		errors: make([]error, 0),
	}
}

// Add appends an error to the chain
func (ec *ErrorChain) Add(err error) {
	if err != nil {
		ec.errors = append(ec.errors, err)
	}
}

// HasError returns true if chain has errors
func (ec *ErrorChain) HasError() bool {
	return len(ec.errors) > 0
}

// Error returns formatted error chain
func (ec *ErrorChain) Error() string {
	if len(ec.errors) == 0 {
		return ""
	}
	var msg string
	for i, err := range ec.errors {
		if i > 0 {
			msg += " -> "
		}
		msg += err.Error()
	}
	return msg
}

// All returns all errors in chain
func (ec *ErrorChain) All() []error {
	return ec.errors
}

// First returns the first error in chain
func (ec *ErrorChain) First() error {
	if len(ec.errors) > 0 {
		return ec.errors[0]
	}
	return nil
}

// Last returns the last error in chain
func (ec *ErrorChain) Last() error {
	if len(ec.errors) > 0 {
		return ec.errors[len(ec.errors)-1]
	}
	return nil
}
