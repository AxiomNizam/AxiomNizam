package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// EventModel GORM model for events
type EventModel struct {
	ID        string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID  string             `gorm:"index;type:varchar(255);not null"`
	Topic     string             `gorm:"index;type:varchar(255);not null"`
	EventType string             `gorm:"index;type:varchar(100)"`
	Payload   datatypes.JSONType `gorm:"type:jsonb"`
	Metadata  datatypes.JSONType `gorm:"type:jsonb"`
	CreatedAt time.Time          `gorm:"index;autoCreateTime;type:timestamp"`
	DeletedAt gorm.DeletedAt     `gorm:"index;type:timestamp"`
}

// TableName specifies table name
func (EventModel) TableName() string {
	return "events"
}

// TopicModel GORM model for topics
type TopicModel struct {
	ID          string         `gorm:"primaryKey;type:varchar(255)"`
	TenantID    string         `gorm:"index;type:varchar(255);not null"`
	Name        string         `gorm:"uniqueIndex:idx_tenant_topic;type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	EventCount  int            `gorm:"type:int"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt `gorm:"index;type:timestamp"`
}

// TableName specifies table name
func (TopicModel) TableName() string {
	return "topics"
}

// SubscriptionModel GORM model for subscriptions
type SubscriptionModel struct {
	ID              string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID        string             `gorm:"index;type:varchar(255);not null"`
	Topic           string             `gorm:"index;type:varchar(255);not null"`
	Endpoint        string             `gorm:"type:varchar(1024)"`
	Filter          datatypes.JSONType `gorm:"type:jsonb"`
	DeliveryPolicy  datatypes.JSONType `gorm:"type:jsonb"`
	Status          string             `gorm:"index;type:varchar(50)"`
	MessageCount    int                `gorm:"type:int"`
	LastMessageTime *time.Time         `gorm:"type:timestamp"`
	CreatedAt       time.Time          `gorm:"autoCreateTime;type:timestamp"`
	UpdatedAt       time.Time          `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt       gorm.DeletedAt     `gorm:"index;type:timestamp"`
}

// TableName specifies table name
func (SubscriptionModel) TableName() string {
	return "subscriptions"
}

// DeadLetterEventModel GORM model for DLQ events
type DeadLetterEventModel struct {
	ID            string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID      string             `gorm:"index;type:varchar(255);not null"`
	Topic         string             `gorm:"index;type:varchar(255)"`
	OriginalEvent datatypes.JSONType `gorm:"type:jsonb"`
	Reason        string             `gorm:"type:text"`
	Attempts      int                `gorm:"type:int"`
	CreatedAt     time.Time          `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt     gorm.DeletedAt     `gorm:"index;type:timestamp"`
}

// TableName specifies table name
func (DeadLetterEventModel) TableName() string {
	return "dead_letter_events"
}
