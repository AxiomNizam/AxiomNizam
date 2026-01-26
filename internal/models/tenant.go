package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TenantModel GORM model for tenants
type TenantModel struct {
	ID          string             `gorm:"primaryKey;type:varchar(255)"`
	Name        string             `gorm:"index;type:varchar(255);not null"`
	Owner       string             `gorm:"type:varchar(255)"`
	Description string             `gorm:"type:text"`
	Metadata    datatypes.JSONType `gorm:"type:jsonb"`
	Status      string             `gorm:"index;type:varchar(50)"`
	CreatedAt   time.Time          `gorm:"index;autoCreateTime;type:timestamp"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Relations
	Members []*TenantMemberModel `gorm:"foreignKey:TenantID;references:ID"`
	Quotas  []*TenantQuotaModel  `gorm:"foreignKey:TenantID;references:ID"`
}

// TableName specifies table name
func (TenantModel) TableName() string {
	return "tenants"
}

// TenantMemberModel GORM model for tenant members
type TenantMemberModel struct {
	ID        string         `gorm:"primaryKey;type:varchar(255)"`
	TenantID  string         `gorm:"index;type:varchar(255);not null"`
	UserID    string         `gorm:"index;type:varchar(255);not null"`
	Role      string         `gorm:"type:varchar(50)"`
	Status    string         `gorm:"type:varchar(50)"`
	AddedAt   time.Time      `gorm:"autoCreateTime;type:timestamp"`
	RemovedAt *time.Time     `gorm:"type:timestamp"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Foreign Key
	Tenant *TenantModel `gorm:"foreignKey:TenantID;references:ID"`
}

// TableName specifies table name
func (TenantMemberModel) TableName() string {
	return "tenant_members"
}

// TenantQuotaModel GORM model for tenant quotas
type TenantQuotaModel struct {
	ID        string         `gorm:"primaryKey;type:varchar(255)"`
	TenantID  string         `gorm:"index;type:varchar(255);not null;uniqueIndex:idx_tenant_resource"`
	Resource  string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_tenant_resource"`
	Limit     int64          `gorm:"type:bigint"`
	Used      int64          `gorm:"type:bigint"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Foreign Key
	Tenant *TenantModel `gorm:"foreignKey:TenantID;references:ID"`
}

// TableName specifies table name
func (TenantQuotaModel) TableName() string {
	return "tenant_quotas"
}
