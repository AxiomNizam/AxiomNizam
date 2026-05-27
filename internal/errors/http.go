package errors

import (
	stderrors "errors"
	"net/http"
)

// HTTPStatusFromError maps a typed error to an HTTP status code.
// Returns 500 for unrecognized errors.
func HTTPStatusFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if stderrors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}
	if stderrors.Is(err, ErrAlreadyExists) || stderrors.Is(err, ErrConflict) {
		return http.StatusConflict
	}
	if stderrors.Is(err, ErrUnauthorized) {
		return http.StatusUnauthorized
	}
	if stderrors.Is(err, ErrForbidden) {
		return http.StatusForbidden
	}
	if stderrors.Is(err, ErrInvalidInput) || stderrors.Is(err, ErrPreconditionFailed) {
		return http.StatusBadRequest
	}
	if stderrors.Is(err, ErrTimeout) {
		return http.StatusGatewayTimeout
	}
	if stderrors.Is(err, ErrUnavailable) {
		return http.StatusServiceUnavailable
	}
	if stderrors.Is(err, ErrNotImplemented) {
		return http.StatusNotImplemented
	}
	if stderrors.Is(err, ErrRateLimited) {
		return http.StatusTooManyRequests
	}
	return http.StatusInternalServerError
}

// ErrorResponse is the standard JSON error response body.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// CodeFromError returns a string error code for the given error.
func CodeFromError(err error) string {
	if err == nil {
		return ""
	}
	if stderrors.Is(err, ErrNotFound) {
		return "NOT_FOUND"
	}
	if stderrors.Is(err, ErrAlreadyExists) {
		return "ALREADY_EXISTS"
	}
	if stderrors.Is(err, ErrConflict) {
		return "CONFLICT"
	}
	if stderrors.Is(err, ErrUnauthorized) {
		return "UNAUTHORIZED"
	}
	if stderrors.Is(err, ErrForbidden) {
		return "FORBIDDEN"
	}
	if stderrors.Is(err, ErrInvalidInput) {
		return "INVALID_INPUT"
	}
	if stderrors.Is(err, ErrTimeout) {
		return "TIMEOUT"
	}
	if stderrors.Is(err, ErrUnavailable) {
		return "UNAVAILABLE"
	}
	if stderrors.Is(err, ErrNotImplemented) {
		return "NOT_IMPLEMENTED"
	}
	if stderrors.Is(err, ErrRateLimited) {
		return "RATE_LIMITED"
	}
	return "INTERNAL_ERROR"
}
