package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RoleModel GORM model for roles
type RoleModel struct {
	ID          string         `gorm:"primaryKey;type:varchar(255)"`
	TenantID    string         `gorm:"index;type:varchar(255);not null"`
	Name        string         `gorm:"uniqueIndex:idx_tenant_role;type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	Level       int            `gorm:"type:int"` // Hierarchy level
	Metadata    datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt   time.Time      `gorm:"index;autoCreateTime;type:timestamp"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Relations
	Bindings    []*RoleBindingModel `gorm:"foreignKey:RoleID;references:ID"`
	Permissions []*PermissionModel  `gorm:"foreignKey:RoleID;references:ID"`
}

// TableName specifies table name
func (RoleModel) TableName() string {
	return "roles"
}

// RoleBindingModel GORM model for role bindings
type RoleBindingModel struct {
	ID          string         `gorm:"primaryKey;type:varchar(255)"`
	TenantID    string         `gorm:"index;type:varchar(255);not null"`
	RoleID      string         `gorm:"index;type:varchar(255);not null"`
	SubjectID   string         `gorm:"index;type:varchar(255);not null"`
	SubjectType string         `gorm:"type:varchar(50)"`
	ExpiresAt   *time.Time     `gorm:"type:timestamp"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Foreign Key
	Role *RoleModel `gorm:"foreignKey:RoleID;references:ID"`
}

// TableName specifies table name
func (RoleBindingModel) TableName() string {
	return "role_bindings"
}

// PermissionModel GORM model for permissions
type PermissionModel struct {
	ID         string         `gorm:"primaryKey;type:varchar(255)"`
	TenantID   string         `gorm:"index;type:varchar(255);not null"`
	RoleID     string         `gorm:"index;type:varchar(255);not null"`
	Resource   string         `gorm:"index;type:varchar(255);not null"`
	Action     string         `gorm:"index;type:varchar(100);not null"`
	Effect     string         `gorm:"type:varchar(10)"` // Allow/Deny
	Conditions datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt  gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Foreign Key
	Role *RoleModel `gorm:"foreignKey:RoleID;references:ID"`
}

// TableName specifies table name
func (PermissionModel) TableName() string {
	return "permissions"
}

// AccessRequestModel GORM model for access requests
type AccessRequestModel struct {
	ID              string         `gorm:"primaryKey;type:varchar(255)"`
	TenantID        string         `gorm:"index;type:varchar(255);not null"`
	SubjectID       string         `gorm:"index;type:varchar(255);not null"`
	RoleID          string         `gorm:"index;type:varchar(255);not null"`
	Justification   string         `gorm:"type:text"`
	Status          string         `gorm:"index;type:varchar(50)"`
	ExpiryDate      *time.Time     `gorm:"type:timestamp"`
	ApprovedAt      *time.Time     `gorm:"type:timestamp"`
	ApprovedBy      string         `gorm:"type:varchar(255)"`
	RejectedAt      *time.Time     `gorm:"type:timestamp"`
	RejectionReason string         `gorm:"type:text"`
	RejectedBy      string         `gorm:"type:varchar(255)"`
	Metadata        datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt       time.Time      `gorm:"index;autoCreateTime;type:timestamp"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt       gorm.DeletedAt `gorm:"index;type:timestamp"`

	// Foreign Keys
	Role *RoleModel `gorm:"foreignKey:RoleID;references:ID"`
}

// TableName specifies table name
func (AccessRequestModel) TableName() string {
	return "access_requests"
}
