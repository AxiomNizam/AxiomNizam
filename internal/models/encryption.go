package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// EncryptionKeyModel GORM model for encryption keys
type EncryptionKeyModel struct {
	ID          string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID    string             `gorm:"index;type:varchar(255);not null"`
	Name        string             `gorm:"type:varchar(255)"`
	Algorithm   string             `gorm:"type:varchar(100)"`
	KeyMaterial string             `gorm:"type:text;not null"` // Encrypted/encoded
	Version     int                `gorm:"type:int"`
	Status      string             `gorm:"index;type:varchar(50)"`
	Metadata    datatypes.JSONType `gorm:"type:jsonb"`
	CreatedAt   time.Time          `gorm:"index;autoCreateTime;type:timestamp"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime;type:timestamp"`
	LastRotated *time.Time         `gorm:"type:timestamp"`
	DeletedAt   gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Relations
	Rotations []*KeyRotationModel        `gorm:"foreignKey:KeyID;references:ID"`
	AuditLogs []*EncryptionAuditLogModel `gorm:"foreignKey:KeyID;references:ID"`
}

// TableName specifies table name
func (EncryptionKeyModel) TableName() string {
	return "encryption_keys"
}

// EncryptionPolicyModel GORM model for encryption policies
type EncryptionPolicyModel struct {
	ID          string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID    string             `gorm:"index;type:varchar(255);not null"`
	Name        string             `gorm:"type:varchar(255)"`
	Description string             `gorm:"type:text"`
	Fields      datatypes.JSONType `gorm:"type:jsonb"`
	KeyID       string             `gorm:"type:varchar(255)"`
	Algorithm   string             `gorm:"type:varchar(100)"`
	Status      string             `gorm:"type:varchar(50)"`
	CreatedAt   time.Time          `gorm:"autoCreateTime;type:timestamp"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt     `gorm:"index;type:timestamp"`
}

// TableName specifies table name
func (EncryptionPolicyModel) TableName() string {
	return "encryption_policies"
}

// KeyRotationModel GORM model for key rotations
type KeyRotationModel struct {
	ID        string         `gorm:"primaryKey;type:varchar(255)"`
	KeyID     string         `gorm:"index;type:varchar(255);not null"`
	OldKeyID  string         `gorm:"type:varchar(255)"`
	NewKeyID  string         `gorm:"type:varchar(255)"`
	RotatedAt time.Time      `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Foreign Key
	Key *EncryptionKeyModel `gorm:"foreignKey:KeyID;references:ID"`
}

// TableName specifies table name
func (KeyRotationModel) TableName() string {
	return "key_rotations"
}

// EncryptionAuditLogModel GORM model for encryption audit logs
type EncryptionAuditLogModel struct {
	ID        string             `gorm:"primaryKey;type:varchar(255)"`
	KeyID     string             `gorm:"index;type:varchar(255);not null"`
	Action    string             `gorm:"index;type:varchar(100)"`
	UserID    string             `gorm:"type:varchar(255)"`
	Details   datatypes.JSONType `gorm:"type:jsonb"`
	Timestamp time.Time          `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Foreign Key
	Key *EncryptionKeyModel `gorm:"foreignKey:KeyID;references:ID"`
}

// TableName specifies table name
func (EncryptionAuditLogModel) TableName() string {
	return "encryption_audit_logs"
}
