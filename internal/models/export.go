package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ExportJobModel GORM model for export jobs
type ExportJobModel struct {
	ID          string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID    string             `gorm:"index;type:varchar(255);not null"`
	UserID      string             `gorm:"type:varchar(255)"`
	Format      string             `gorm:"index;type:varchar(50)"`
	Status      string             `gorm:"index;type:varchar(50)"`
	Progress    int                `gorm:"type:int"`
	Filters     datatypes.JSONType `gorm:"type:jsonb"`
	FilePath    string             `gorm:"type:varchar(1024)"`
	FileSize    int64              `gorm:"type:bigint"`
	StartedAt   *time.Time         `gorm:"type:timestamp"`
	CompletedAt *time.Time         `gorm:"type:timestamp"`
	CreatedAt   time.Time          `gorm:"index;autoCreateTime;type:timestamp"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Relations
	Results []*ExportResultModel `gorm:"foreignKey:ExportID;references:ID"`
}

// TableName specifies table name
func (ExportJobModel) TableName() string {
	return "export_jobs"
}

// ExportResultModel GORM model for export results
type ExportResultModel struct {
	ID        string             `gorm:"primaryKey;type:varchar(255)"`
	ExportID  string             `gorm:"index;type:varchar(255);not null"`
	ItemID    string             `gorm:"type:varchar(255)"`
	Status    string             `gorm:"type:varchar(50)"`
	Data      datatypes.JSONType `gorm:"type:jsonb"`
	Error     string             `gorm:"type:text"`
	CreatedAt time.Time          `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Foreign Key
	Export *ExportJobModel `gorm:"foreignKey:ExportID;references:ID"`
}

// TableName specifies table name
func (ExportResultModel) TableName() string {
	return "export_results"
}

// ExportTemplateModel GORM model for export templates
type ExportTemplateModel struct {
	ID          string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID    string             `gorm:"index;type:varchar(255);not null"`
	Name        string             `gorm:"type:varchar(255)"`
	Description string             `gorm:"type:text"`
	Format      string             `gorm:"type:varchar(50)"`
	Config      datatypes.JSONType `gorm:"type:jsonb"`
	Columns     datatypes.JSONType `gorm:"type:jsonb"`
	CreatedAt   time.Time          `gorm:"autoCreateTime;type:timestamp"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt     `gorm:"index;type:timestamp"`
}

// TableName specifies table name
func (ExportTemplateModel) TableName() string {
	return "export_templates"
}
