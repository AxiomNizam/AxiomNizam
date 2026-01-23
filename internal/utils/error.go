package utils

import (
	"fmt"
	"net/http"
)

// ErrorType defines the type of error
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound     ErrorType = "NOT_FOUND"
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden    ErrorType = "FORBIDDEN"
	ErrorTypeConflict     ErrorType = "CONFLICT"
	ErrorTypeInternal     ErrorType = "INTERNAL_SERVER_ERROR"
	ErrorTypeBadRequest   ErrorType = "BAD_REQUEST"
	ErrorTypeDatabase     ErrorType = "DATABASE_ERROR"
	ErrorTypeExternal     ErrorType = "EXTERNAL_SERVICE_ERROR"
	ErrorTypeTimeout      ErrorType = "TIMEOUT"
)

// CustomError represents a custom error with additional information
type CustomError struct {
	Type       ErrorType   `json:"type"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Details    interface{} `json:"details,omitempty"`
	Err        error       `json:"-"` // Original error for logging
}

// Error implements the error interface
func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// NewValidationError creates a validation error
func NewValidationError(message string, details interface{}) *CustomError {
	return &CustomError{
		Type:       ErrorTypeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Details:    details,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message string) *CustomError {
	return &CustomError{
		Type:       ErrorTypeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *CustomError {
	return &CustomError{
		Type:       ErrorTypeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *CustomError {
	return &CustomError{
		Type:       ErrorTypeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string, details interface{}) *CustomError {
	return &CustomError{
		Type:       ErrorTypeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
		Details:    details,
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string, err error) *CustomError {
	return &CustomError{
		Type:       ErrorTypeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *CustomError {
	return &CustomError{
		Type:       ErrorTypeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(message string, err error) *CustomError {
	return &CustomError{
		Type:       ErrorTypeDatabase,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewExternalServiceError creates an external service error
func NewExternalServiceError(message string, err error) *CustomError {
	return &CustomError{
		Type:       ErrorTypeExternal,
		Message:    message,
		StatusCode: http.StatusBadGateway,
		Err:        err,
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(message string) *CustomError {
	return &CustomError{
		Type:       ErrorTypeTimeout,
		Message:    message,
		StatusCode: http.StatusGatewayTimeout,
	}
}

// WithDetails adds details to an error
func (e *CustomError) WithDetails(details interface{}) *CustomError {
	e.Details = details
	return e
}

// WithOriginalError sets the original error
func (e *CustomError) WithOriginalError(err error) *CustomError {
	e.Err = err
	return e
}

// IsCustomError checks if an error is a CustomError
func IsCustomError(err error) bool {
	_, ok := err.(*CustomError)
	return ok
}

// AsCustomError converts an error to CustomError if possible
func AsCustomError(err error) *CustomError {
	if customErr, ok := err.(*CustomError); ok {
		return customErr
	}
	return NewInternalError("Unknown error", err)
}

// ErrorResponse represents the error response structure for API
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// ToErrorResponse converts CustomError to ErrorResponse for API response
func (e *CustomError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Error:   string(e.Type),
		Message: e.Message,
		Code:    string(e.Type),
		Details: e.Details,
	}
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationErrorDetail `json:"errors"`
}

type ValidationErrorDetail struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// NewValidationErrors creates a validation errors collection
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: []ValidationErrorDetail{},
	}
}

// AddError adds a validation error detail
func (ve *ValidationErrors) AddError(field, message string, value interface{}) {
	ve.Errors = append(ve.Errors, ValidationErrorDetail{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors checks if there are any errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// ToCustomError converts ValidationErrors to a CustomError
func (ve *ValidationErrors) ToCustomError() *CustomError {
	return NewValidationError("Validation failed", ve.Errors)
}
