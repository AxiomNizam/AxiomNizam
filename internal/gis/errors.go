package gis

import (
	"errors"
	"fmt"
)

var (
	// ErrLayerNotFound is returned when a GIS layer doesn't exist.
	ErrLayerNotFound = errors.New("gis: layer not found")

	// ErrRegionNotFound is returned when a GIS region doesn't exist.
	ErrRegionNotFound = errors.New("gis: region not found")

	// ErrMarkerNotFound is returned when a GIS marker doesn't exist.
	ErrMarkerNotFound = errors.New("gis: marker not found")

	// ErrDatasetNotFound is returned when a GIS dataset doesn't exist.
	ErrDatasetNotFound = errors.New("gis: dataset not found")

	// ErrDashboardNotFound is returned when a specialized dashboard doesn't exist.
	ErrDashboardNotFound = errors.New("gis: dashboard not found")

	// ErrInvalidGeoJSON is returned when GeoJSON validation fails.
	ErrInvalidGeoJSON = errors.New("gis: invalid GeoJSON")

	// ErrInvalidCoordinates is returned when lat/lng coordinates are out of range.
	ErrInvalidCoordinates = errors.New("gis: invalid coordinates")

	// ErrDuplicateEntity is returned when a GIS entity with the same ID already exists.
	ErrDuplicateEntity = errors.New("gis: duplicate entity")
)

// EntityError wraps a GIS entity error with context.
type EntityError struct {
	Op       string // operation (e.g., "CreateLayer", "UpdateMarker")
	EntityID string
	Err      error
}

func (e *EntityError) Error() string {
	if e.EntityID != "" {
		return fmt.Sprintf("gis: %s %q: %v", e.Op, e.EntityID, e.Err)
	}
	return fmt.Sprintf("gis: %s: %v", e.Op, e.Err)
}

func (e *EntityError) Unwrap() error {
	return e.Err
}
