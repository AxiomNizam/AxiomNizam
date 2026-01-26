package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// StreamModel GORM model for streams
type StreamModel struct {
	ID          string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID    string             `gorm:"index;type:varchar(255);not null"`
	Name        string             `gorm:"type:varchar(255)"`
	Description string             `gorm:"type:text"`
	Config      datatypes.JSONType `gorm:"type:jsonb"`
	Status      string             `gorm:"index;type:varchar(50)"`
	CreatedAt   time.Time          `gorm:"index;autoCreateTime;type:timestamp"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Relations
	Subscriptions []*StreamSubscriptionModel `gorm:"foreignKey:StreamID;references:ID"`
}

// TableName specifies table name
func (StreamModel) TableName() string {
	return "streams"
}

// StreamSubscriptionModel GORM model for stream subscriptions
type StreamSubscriptionModel struct {
	ID        string         `gorm:"primaryKey;type:varchar(255)"`
	StreamID  string         `gorm:"index;type:varchar(255);not null"`
	UserID    string         `gorm:"index;type:varchar(255)"`
	Active    bool           `gorm:"type:boolean;default:true"`
	CreatedAt time.Time      `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Foreign Key
	Stream *StreamModel `gorm:"foreignKey:StreamID;references:ID"`
}

// TableName specifies table name
func (StreamSubscriptionModel) TableName() string {
	return "stream_subscriptions"
}
