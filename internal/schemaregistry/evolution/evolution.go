package evolution

// =====================================================
// WS-3.2 — Schema Evolution Manager
//
// Manages schema evolution workflows: migration planning,
// dual-write coordination, and consumer compatibility tracking.
// Generates step-by-step migration plans for safe schema transitions.
// =====================================================

import (
	"fmt"
	"strings"
	"time"
)

// MigrationStep represents a single step in a schema migration plan.
type MigrationStep struct {
	Order       int    `json:"order"`
	Action      string `json:"action"`      // add_field, remove_field, rename_field, change_type, add_default
	Field       string `json:"field"`
	Description string `json:"description"`
	Breaking    bool   `json:"breaking"`
	Rollback    string `json:"rollback,omitempty"`
}

// MigrationPlan describes the full plan for migrating between two schema versions.
type MigrationPlan struct {
	Subject       string          `json:"subject"`
	FromVersion   int             `json:"fromVersion"`
	ToVersion     int             `json:"toVersion"`
	Steps         []MigrationStep `json:"steps"`
	IsBreaking    bool            `json:"isBreaking"`
	EstimatedRisk string          `json:"estimatedRisk"` // low, medium, high
	GeneratedAt   time.Time       `json:"generatedAt"`
}

// ConsumerStatus tracks a consumer's schema compatibility status.
type ConsumerStatus struct {
	ConsumerID      string    `json:"consumerId"`
	CurrentVersion  int       `json:"currentVersion"`
	TargetVersion   int       `json:"targetVersion"`
	Compatible      bool      `json:"compatible"`
	LastCheckedAt   time.Time `json:"lastCheckedAt"`
	MigrationStatus string   `json:"migrationStatus"` // pending, in_progress, completed, failed
}

// DualWriteConfig describes a dual-write setup during migration.
type DualWriteConfig struct {
	Subject        string `json:"subject"`
	PrimaryVersion int    `json:"primaryVersion"`
	ShadowVersion  int    `json:"shadowVersion"`
	Enabled        bool   `json:"enabled"`
	WriteMode      string `json:"writeMode"` // primary_only, dual_write, shadow_only
}

// EvolutionManager coordinates schema evolution across producers and consumers.
type EvolutionManager struct {
	consumers  map[string]*ConsumerStatus  // consumerId -> status
	dualWrites map[string]*DualWriteConfig // subject -> config
}

// NewEvolutionManager creates a new schema evolution manager.
func NewEvolutionManager() *EvolutionManager {
	return &EvolutionManager{
		consumers:  make(map[string]*ConsumerStatus),
		dualWrites: make(map[string]*DualWriteConfig),
	}
}

// PlanMigration generates a step-by-step migration plan between two schemas.
func (m *EvolutionManager) PlanMigration(subject string, fromVersion, toVersion int, oldFields, newFields []FieldInfo) *MigrationPlan {
	plan := &MigrationPlan{
		Subject:     subject,
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		GeneratedAt: time.Now(),
	}

	oldMap := fieldMap(oldFields)
	newMap := fieldMap(newFields)
	order := 0

	// Detect added fields (safe — add with defaults first).
	for name, newField := range newMap {
		if _, exists := oldMap[name]; !exists {
			order++
			step := MigrationStep{
				Order:       order,
				Action:      "add_field",
				Field:       name,
				Description: fmt.Sprintf("Add new field '%s' (type: %s)", name, newField.Type),
				Breaking:    newField.Required && !newField.HasDefault,
				Rollback:    fmt.Sprintf("Remove field '%s'", name),
			}
			if step.Breaking {
				plan.IsBreaking = true
			}
			plan.Steps = append(plan.Steps, step)
		}
	}

	// Detect type changes (potentially breaking).
	for name, oldField := range oldMap {
		if newField, exists := newMap[name]; exists {
			if !strings.EqualFold(oldField.Type, newField.Type) {
				order++
				plan.Steps = append(plan.Steps, MigrationStep{
					Order:       order,
					Action:      "change_type",
					Field:       name,
					Description: fmt.Sprintf("Change type of '%s': %s → %s", name, oldField.Type, newField.Type),
					Breaking:    true,
					Rollback:    fmt.Sprintf("Revert type of '%s' to %s", name, oldField.Type),
				})
				plan.IsBreaking = true
			}
		}
	}

	// Detect removed fields (breaking — do last).
	for name := range oldMap {
		if _, exists := newMap[name]; !exists {
			order++
			plan.Steps = append(plan.Steps, MigrationStep{
				Order:       order,
				Action:      "remove_field",
				Field:       name,
				Description: fmt.Sprintf("Remove field '%s'", name),
				Breaking:    true,
				Rollback:    fmt.Sprintf("Re-add field '%s'", name),
			})
			plan.IsBreaking = true
		}
	}

	// Estimate risk.
	if plan.IsBreaking {
		if len(plan.Steps) > 3 {
			plan.EstimatedRisk = "high"
		} else {
			plan.EstimatedRisk = "medium"
		}
	} else {
		plan.EstimatedRisk = "low"
	}

	return plan
}

// RegisterConsumer adds a consumer to track during migration.
func (m *EvolutionManager) RegisterConsumer(consumerID string, currentVersion int) {
	m.consumers[consumerID] = &ConsumerStatus{
		ConsumerID:      consumerID,
		CurrentVersion:  currentVersion,
		LastCheckedAt:   time.Now(),
		MigrationStatus: "pending",
	}
}

// UpdateConsumerVersion marks a consumer as migrated to a new version.
func (m *EvolutionManager) UpdateConsumerVersion(consumerID string, version int) bool {
	c, ok := m.consumers[consumerID]
	if !ok {
		return false
	}
	c.CurrentVersion = version
	c.LastCheckedAt = time.Now()
	if version >= c.TargetVersion {
		c.MigrationStatus = "completed"
		c.Compatible = true
	} else {
		c.MigrationStatus = "in_progress"
	}
	return true
}

// ListConsumers returns all tracked consumers for a given target version.
func (m *EvolutionManager) ListConsumers(targetVersion int) []*ConsumerStatus {
	var result []*ConsumerStatus
	for _, c := range m.consumers {
		c.TargetVersion = targetVersion
		c.Compatible = c.CurrentVersion >= targetVersion
		result = append(result, c)
	}
	return result
}

// AllConsumersMigrated checks if all tracked consumers are on the target version.
func (m *EvolutionManager) AllConsumersMigrated(targetVersion int) bool {
	for _, c := range m.consumers {
		if c.CurrentVersion < targetVersion {
			return false
		}
	}
	return len(m.consumers) > 0
}

// EnableDualWrite sets up dual-write for a subject during migration.
func (m *EvolutionManager) EnableDualWrite(subject string, primaryVersion, shadowVersion int) {
	m.dualWrites[subject] = &DualWriteConfig{
		Subject:        subject,
		PrimaryVersion: primaryVersion,
		ShadowVersion:  shadowVersion,
		Enabled:        true,
		WriteMode:      "dual_write",
	}
}

// DisableDualWrite turns off dual-write after migration completes.
func (m *EvolutionManager) DisableDualWrite(subject string) bool {
	dw, ok := m.dualWrites[subject]
	if !ok {
		return false
	}
	dw.Enabled = false
	dw.WriteMode = "primary_only"
	return true
}

// GetDualWriteConfig returns the current dual-write configuration for a subject.
func (m *EvolutionManager) GetDualWriteConfig(subject string) (*DualWriteConfig, bool) {
	dw, ok := m.dualWrites[subject]
	return dw, ok
}

// --- Types ---

// FieldInfo describes a single field for migration planning.
type FieldInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Required   bool   `json:"required"`
	HasDefault bool   `json:"hasDefault"`
}

func fieldMap(fields []FieldInfo) map[string]FieldInfo {
	m := make(map[string]FieldInfo)
	for _, f := range fields {
		m[f.Name] = f
	}
	return m
}
