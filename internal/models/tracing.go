package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TraceModel GORM model for traces
type TraceModel struct {
	ID        string         `gorm:"primaryKey;type:varchar(255)"`
	TenantID  string         `gorm:"index;type:varchar(255);not null"`
	TraceID   string         `gorm:"uniqueIndex;type:varchar(255);not null"`
	Service   string         `gorm:"index;type:varchar(255)"`
	Operation string         `gorm:"index;type:varchar(255)"`
	Status    string         `gorm:"index;type:varchar(50)"`
	Duration  int64          `gorm:"type:bigint"` // microseconds
	StartTime time.Time      `gorm:"index;autoCreateTime;type:timestamp"`
	EndTime   *time.Time     `gorm:"type:timestamp"`
	Tags      datatypes.JSON `gorm:"type:jsonb"`
	Metadata  datatypes.JSON `gorm:"type:jsonb"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:timestamp"`
}

// TableName specifies table name
func (TraceModel) TableName() string {
	return "traces"
}

// SpanModel GORM model for spans
type SpanModel struct {
	ID           string         `gorm:"primaryKey;type:varchar(255)"`
	TraceID      string         `gorm:"index;type:varchar(255);not null"`
	SpanID       string         `gorm:"uniqueIndex;type:varchar(255);not null"`
	ParentSpanID string         `gorm:"index;type:varchar(255)"`
	Service      string         `gorm:"index;type:varchar(255)"`
	Operation    string         `gorm:"index;type:varchar(255)"`
	Status       string         `gorm:"type:varchar(50)"`
	Duration     int64          `gorm:"type:bigint"` // microseconds
	StartTime    time.Time      `gorm:"autoCreateTime;type:timestamp"`
	EndTime      *time.Time     `gorm:"type:timestamp"`
	Tags         datatypes.JSON `gorm:"type:jsonb"`
	Logs         datatypes.JSON `gorm:"type:jsonb"`
	Error        string         `gorm:"type:text"`
	DeletedAt    gorm.DeletedAt `gorm:"index;type:timestamp"`
}

// TableName specifies table name
func (SpanModel) TableName() string {
	return "spans"
}

// ServiceMetricsModel GORM model for service metrics
type ServiceMetricsModel struct {
	ID           string         `gorm:"primaryKey;type:varchar(255)"`
	TenantID     string         `gorm:"index;type:varchar(255);not null"`
	Service      string         `gorm:"uniqueIndex:idx_tenant_service;type:varchar(255);not null"`
	RequestCount int            `gorm:"type:int"`
	ErrorCount   int            `gorm:"type:int"`
	P99Duration  int64          `gorm:"type:bigint"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt    gorm.DeletedAt `gorm:"index;type:timestamp"`
}

// TableName specifies table name
func (ServiceMetricsModel) TableName() string {
	return "service_metrics"
}
