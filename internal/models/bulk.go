package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// BulkOperationModel GORM model for bulk operations
type BulkOperationModel struct {
	ID            string         `gorm:"primaryKey;type:varchar(255)"`
	TenantID      string         `gorm:"index;type:varchar(255);not null"`
	UserID        string         `gorm:"type:varchar(255)"`
	OperationType string         `gorm:"index;type:varchar(100)"`
	Status        string         `gorm:"index;type:varchar(50)"`
	TotalItems    int            `gorm:"type:int"`
	SuccessCount  int            `gorm:"type:int"`
	FailureCount  int            `gorm:"type:int"`
	SkippedCount  int            `gorm:"type:int"`
	Parameters    datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt     time.Time      `gorm:"index;autoCreateTime;type:timestamp"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt     gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Relations
	Results []*BulkResultModel `gorm:"foreignKey:OperationID;references:ID"`
}

// TableName specifies table name
func (BulkOperationModel) TableName() string {
	return "bulk_operations"
}

// BulkResultModel GORM model for bulk operation results
type BulkResultModel struct {
	ID          string         `gorm:"primaryKey;type:varchar(255)"`
	OperationID string         `gorm:"index;type:varchar(255);not null"`
	ItemIndex   int            `gorm:"type:int"`
	Status      string         `gorm:"type:varchar(50)"`
	Result      datatypes.JSON `gorm:"type:jsonb"`
	Error       string         `gorm:"type:text"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Foreign Key
	Operation *BulkOperationModel `gorm:"foreignKey:OperationID;references:ID"`
}

// TableName specifies table name
func (BulkResultModel) TableName() string {
	return "bulk_results"
}
