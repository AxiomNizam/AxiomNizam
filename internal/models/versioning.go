package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ResourceVersionModel GORM model for resource versions
type ResourceVersionModel struct {
	ID                string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID          string             `gorm:"index;type:varchar(255);not null"`
	ResourceType      string             `gorm:"index;type:varchar(100);not null"`
	ResourceID        string             `gorm:"index;type:varchar(255);not null"`
	VersionNumber     int                `gorm:"type:int;not null"`
	Content           datatypes.JSONType `gorm:"type:jsonb"`
	ChangedBy         string             `gorm:"type:varchar(255)"`
	ChangeDescription string             `gorm:"type:text"`
	CreatedAt         time.Time          `gorm:"index;autoCreateTime;type:timestamp"`
	DeletedAt         gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Index for resource+version lookup
	Index string `gorm:"index:idx_resource_version;type:varchar(511)"`

	// Relations
	Snapshots []*VersionSnapshotModel `gorm:"foreignKey:VersionID;references:ID"`
}

// TableName specifies table name
func (ResourceVersionModel) TableName() string {
	return "resource_versions"
}

// VersionSnapshotModel GORM model for version snapshots
type VersionSnapshotModel struct {
	ID          string             `gorm:"primaryKey;type:varchar(255)"`
	VersionID   string             `gorm:"index;type:varchar(255);not null"`
	Name        string             `gorm:"index;type:varchar(255)"`
	Description string             `gorm:"type:text"`
	Content     datatypes.JSONType `gorm:"type:jsonb"`
	CreatedAt   time.Time          `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Foreign Key
	Version *ResourceVersionModel `gorm:"foreignKey:VersionID;references:ID"`
}

// TableName specifies table name
func (VersionSnapshotModel) TableName() string {
	return "version_snapshots"
}
