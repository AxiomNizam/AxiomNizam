package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// LineageNodeModel GORM model for lineage nodes
type LineageNodeModel struct {
	ID         string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID   string             `gorm:"index;type:varchar(255);not null"`
	Name       string             `gorm:"index;type:varchar(255);not null"`
	NodeType   string             `gorm:"index;type:varchar(100)"`
	Schema     string             `gorm:"type:varchar(255)"`
	Database   string             `gorm:"type:varchar(255)"`
	Metadata   datatypes.JSONType `gorm:"type:jsonb"`
	Columns    datatypes.JSONType `gorm:"type:jsonb"`
	Properties datatypes.JSONType `gorm:"type:jsonb"`
	CreatedAt  time.Time          `gorm:"index;autoCreateTime;type:timestamp"`
	UpdatedAt  time.Time          `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt  gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Relations
	OutgoingEdges []*LineageEdgeModel `gorm:"foreignKey:SourceNodeID;references:ID"`
	IncomingEdges []*LineageEdgeModel `gorm:"foreignKey:TargetNodeID;references:ID"`
}

// TableName specifies table name
func (LineageNodeModel) TableName() string {
	return "lineage_nodes"
}

// LineageEdgeModel GORM model for lineage edges
type LineageEdgeModel struct {
	ID             string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID       string             `gorm:"index;type:varchar(255);not null"`
	SourceNodeID   string             `gorm:"index;type:varchar(255);not null"`
	TargetNodeID   string             `gorm:"index;type:varchar(255);not null"`
	EdgeType       string             `gorm:"type:varchar(100)"`
	Transformation datatypes.JSONType `gorm:"type:jsonb"`
	Columns        datatypes.JSONType `gorm:"type:jsonb"`
	CreatedAt      time.Time          `gorm:"autoCreateTime;type:timestamp"`
	DeletedAt      gorm.DeletedAt     `gorm:"index;type:timestamp"`

	// Foreign Keys
	SourceNode *LineageNodeModel `gorm:"foreignKey:SourceNodeID;references:ID"`
	TargetNode *LineageNodeModel `gorm:"foreignKey:TargetNodeID;references:ID"`
}

// TableName specifies table name
func (LineageEdgeModel) TableName() string {
	return "lineage_edges"
}

// LineageProcessModel GORM model for lineage processes
type LineageProcessModel struct {
	ID          string             `gorm:"primaryKey;type:varchar(255)"`
	TenantID    string             `gorm:"index;type:varchar(255);not null"`
	ProcessID   string             `gorm:"index;type:varchar(255)"`
	ProcessName string             `gorm:"type:varchar(255)"`
	ProcessType string             `gorm:"type:varchar(100)"`
	Input       datatypes.JSONType `gorm:"type:jsonb"`
	Output      datatypes.JSONType `gorm:"type:jsonb"`
	Config      datatypes.JSONType `gorm:"type:jsonb"`
	CreatedAt   time.Time          `gorm:"autoCreateTime;type:timestamp"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime;type:timestamp"`
	DeletedAt   gorm.DeletedAt     `gorm:"index;type:timestamp"`
}

// TableName specifies table name
func (LineageProcessModel) TableName() string {
	return "lineage_processes"
}
