package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WebhookModel GORM model for webhooks
type WebhookModel struct {
	ID          string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID    string             `gorm:"index;type:varchar(255);not null"`
	URL         string             `gorm:"type:varchar(1024)"`
	EventTypes  datatypes.JSONType `gorm:"type:jsonb"`
	Headers     datatypes.JSONType `gorm:"type:jsonb"`
	Active      bool               `gorm:"type:boolean;default:true"`
	Description string             `gorm:"type:text"`
	CreatedAt   time.Time          `gorm:"index;autoCreateTime;type:timestamp"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Relations
	DeliveryLogs []*WebhookDeliveryLogModel `gorm:"foreignKey:WebhookID;references:ID"`
}

// TableName specifies table name
func (WebhookModel) TableName() string {
	return "webhooks"
}

// WebhookDeliveryLogModel GORM model for webhook delivery logs
type WebhookDeliveryLogModel struct {
	ID           string         `gorm:"primaryKey;type:varchar(255)"`
	WebhookID    string         `gorm:"index;type:varchar(255);not null"`
	EventID      string         `gorm:"type:varchar(255)"`
	Status       string         `gorm:"index;type:varchar(50)"`
	StatusCode   int            `gorm:"type:int"`
	ResponseBody string         `gorm:"type:text"`
	Error        string         `gorm:"type:text"`
	Attempts     int            `gorm:"type:int"`
	NextRetry    *time.Time     `gorm:"type:timestamp"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt    gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Foreign Key
	Webhook *WebhookModel `gorm:"foreignKey:WebhookID;references:ID"`
}

// TableName specifies table name
func (WebhookDeliveryLogModel) TableName() string {
	return "webhook_delivery_logs"
}
