package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// JobModel GORM model for jobs
type JobModel struct {
	ID          string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID    string             `gorm:"index;type:varchar(255);not null"`
	UserID      string             `gorm:"type:varchar(255)"`
	JobType     string             `gorm:"index;type:varchar(100)"`
	Status      string             `gorm:"index;type:varchar(50)"`
	Parameters  datatypes.JSONType `gorm:"type:jsonb"`
	Result      datatypes.JSONType `gorm:"type:jsonb"`
	ErrorMsg    string             `gorm:"type:text"`
	Progress    int                `gorm:"type:int"`
	StartedAt   *time.Time         `gorm:"type:timestamp"`
	CompletedAt *time.Time         `gorm:"type:timestamp"`
	CreatedAt   time.Time          `gorm:"index;autoCreateTime;type:timestamp"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Relations
	Logs []*JobLogModel `gorm:"foreignKey:JobID;references:ID"`
}

// TableName specifies table name
func (JobModel) TableName() string {
	return "jobs"
}

// JobLogModel GORM model for job logs
type JobLogModel struct {
	ID        string         `gorm:"primaryKey;type:varchar(255)"`
	JobID     string         `gorm:"index;type:varchar(255);not null"`
	Level     string         `gorm:"type:varchar(20)"`
	Message   string         `gorm:"type:text"`
	CreatedAt time.Time      `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Foreign Key
	Job *JobModel `gorm:"foreignKey:JobID;references:ID"`
}

// TableName specifies table name
func (JobLogModel) TableName() string {
	return "job_logs"
}
