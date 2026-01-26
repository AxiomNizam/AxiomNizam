package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AuditLogModel GORM model for audit logs
type AuditLogModel struct {
	ID           string                 `gorm:"primaryKey;type:varchar(255)"`
	TenantID     string                 `gorm:"index;type:varchar(255)"`
	UserID       string                 `gorm:"index;type:varchar(255)"`
	ActionType   string                 `gorm:"index;type:varchar(100)"`
	ResourceType string                 `gorm:"index;type:varchar(100)"`
	ResourceID   string                 `gorm:"index;type:varchar(255)"`
	Details      datatypes.JSONType     `gorm:"type:jsonb"`
	Hash         string                 `gorm:"index;type:varchar(256)"`
	CreatedAt    time.Time              `gorm:"index;autoCreateTime;type:timestamp"`
	UpdatedAt    time.Time              `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt    gorm.DeletedAt         `gorm:"index;type:timestamp"`
}

// TableName specifies table name
func (AuditLogModel) TableName() string {
	return "audit_logs"
}

// Scan implements sql.Scanner
func (a *AuditLogModel) Scan(value interface{}) error {
	return nil
}

// Value implements driver.Valuer
func (a AuditLogModel) Value() (driver.Value, error) {
	return json.Marshal(a)
}
