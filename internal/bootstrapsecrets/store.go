package bootstrapsecrets

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const bootstrapTimeout = 5 * time.Second

// Record stores a shared bootstrap secret persisted in PostgreSQL.
type Record struct {
	SecretKey   string    `gorm:"column:secret_key;primaryKey;type:text"`
	SecretValue string    `gorm:"column:secret_value;type:text;not null"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Record) TableName() string {
	return "iam_bootstrap_secrets"
}

// Ensure returns an existing secret value for key, or atomically stores and returns a generated value.
func Ensure(pg *gorm.DB, key string, generate func() (string, error)) (string, error) {
	if pg == nil {
		return "", errors.New("postgres bootstrap store is unavailable")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("bootstrap key is required")
	}
	if generate == nil {
		return "", errors.New("bootstrap generator is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), bootstrapTimeout)
	defer cancel()

	db := pg.WithContext(ctx)
	if err := db.AutoMigrate(&Record{}); err != nil {
		return "", fmt.Errorf("bootstrap secret table migration: %w", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		return "", fmt.Errorf("bootstrap secret transaction start: %w", tx.Error)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	var existing string
	if err := tx.Raw("SELECT secret_value FROM iam_bootstrap_secrets WHERE secret_key = ?", key).Scan(&existing).Error; err != nil {
		return "", fmt.Errorf("bootstrap secret query: %w", err)
	}
	existing = strings.TrimSpace(existing)
	if existing != "" {
		if err := tx.Commit().Error; err != nil {
			return "", fmt.Errorf("bootstrap secret commit existing: %w", err)
		}
		tx = nil
		return existing, nil
	}

	candidate, err := generate()
	if err != nil {
		return "", err
	}
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return "", errors.New("generated bootstrap secret is empty")
	}

	if err := tx.Exec(
		"INSERT INTO iam_bootstrap_secrets (secret_key, secret_value, created_at, updated_at) VALUES (?, ?, NOW(), NOW()) ON CONFLICT (secret_key) DO NOTHING",
		key,
		candidate,
	).Error; err != nil {
		return "", fmt.Errorf("bootstrap secret upsert: %w", err)
	}

	var resolved string
	if err := tx.Raw("SELECT secret_value FROM iam_bootstrap_secrets WHERE secret_key = ?", key).Scan(&resolved).Error; err != nil {
		return "", fmt.Errorf("bootstrap secret resolve: %w", err)
	}
	resolved = strings.TrimSpace(resolved)
	if resolved == "" {
		return "", fmt.Errorf("bootstrap secret resolved empty for key %q", key)
	}

	if err := tx.Commit().Error; err != nil {
		return "", fmt.Errorf("bootstrap secret commit: %w", err)
	}
	tx = nil
	return resolved, nil
}
