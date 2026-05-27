package apibuilder

import (
	"errors"
	"fmt"
)

var (
	// ErrAPINotFound is returned when a custom API doesn't exist.
	ErrAPINotFound = errors.New("apibuilder: API not found")

	// ErrDuplicatePath is returned when an API path already exists.
	ErrDuplicatePath = errors.New("apibuilder: API path already exists")

	// ErrInvalidSQL is returned when SQL template validation fails.
	ErrInvalidSQL = errors.New("apibuilder: invalid SQL template")

	// ErrUploadNotFound is returned when a CSV upload doesn't exist.
	ErrUploadNotFound = errors.New("apibuilder: upload not found")

	// ErrDashboardNotFound is returned when a dashboard doesn't exist.
	ErrDashboardNotFound = errors.New("apibuilder: dashboard not found")

	// ErrConversionFailed is returned when a dashboard↔GIS conversion fails.
	ErrConversionFailed = errors.New("apibuilder: conversion failed")

	// ErrScanFailed is returned when a file scan fails.
	ErrScanFailed = errors.New("apibuilder: scan failed")

	// ErrFileTooLarge is returned when an uploaded file exceeds the size limit.
	ErrFileTooLarge = errors.New("apibuilder: file too large")

	// ErrUnsupportedFileType is returned for unsupported file types.
	ErrUnsupportedFileType = errors.New("apibuilder: unsupported file type")
)

// APIError wraps an API builder error with context.
type APIError struct {
	Op   string // operation (e.g., "CreateAPI", "ScanFile")
	Name string // resource name
	Err  error
}

func (e *APIError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("apibuilder: %s %q: %v", e.Op, e.Name, e.Err)
	}
	return fmt.Sprintf("apibuilder: %s: %v", e.Op, e.Err)
}

func (e *APIError) Unwrap() error {
	return e.Err
}
