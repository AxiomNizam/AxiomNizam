package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository is the base interface for all repositories
// It provides common database operations that can be used by all repositories
type Repository interface {
	// Create inserts a new record
	Create(ctx context.Context, value interface{}) error

	// GetByID retrieves a record by its ID
	GetByID(ctx context.Context, id interface{}) (interface{}, error)

	// Update updates an existing record
	Update(ctx context.Context, value interface{}) error

	// Delete removes a record
	Delete(ctx context.Context, id interface{}) error

	// FindAll retrieves all records
	FindAll(ctx context.Context) (interface{}, error)

	// Exists checks if a record exists by ID
	Exists(ctx context.Context, id interface{}) (bool, error)
}

// BaseRepository provides common repository operations
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// Create inserts a new record
func (r *BaseRepository) Create(ctx context.Context, value interface{}) error {
	if r.db == nil {
		return errors.New("database connection not initialized")
	}

	result := r.db.WithContext(ctx).Create(value)
	return result.Error
}

// Update updates an existing record
func (r *BaseRepository) Update(ctx context.Context, value interface{}) error {
	if r.db == nil {
		return errors.New("database connection not initialized")
	}

	result := r.db.WithContext(ctx).Save(value)
	return result.Error
}

// Delete removes a record by ID
func (r *BaseRepository) Delete(ctx context.Context, id interface{}) error {
	if r.db == nil {
		return errors.New("database connection not initialized")
	}

	result := r.db.WithContext(ctx).Delete(id)
	return result.Error
}

// GetDB returns the underlying GORM database connection
// This is useful for repositories that need more complex queries
func (r *BaseRepository) GetDB() *gorm.DB {
	return r.db
}

// Common query errors
var (
	// ErrNotFound is returned when a record is not found
	ErrNotFound = gorm.ErrRecordNotFound

	// ErrInvalidInput is returned when input is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrDuplicateEntry is returned when trying to create a duplicate record
	ErrDuplicateEntry = errors.New("duplicate entry")

	// ErrUnauthorized is returned when user is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when user is forbidden
	ErrForbidden = errors.New("forbidden")
)
