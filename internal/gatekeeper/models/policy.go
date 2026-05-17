package models

import (
	"time"

	"github.com/google/uuid"
)

// MFAPolicy defines when and how MFA is enforced.
// Modeled after K8s NetworkPolicy and RBAC: centralized enforcement rules.
type MFAPolicy struct {
	ID          uuid.UUID `db:"id"          json:"id"`
	TenantID    uuid.UUID `db:"tenant_id"   json:"tenant_id"`
	Name        string    `db:"name"        json:"name"`
	Description string    `db:"description" json:"description"`

	// Enforcement mode: Optional, Required, or Adaptive
	Enforcement PolicyEnforcement `db:"enforcement" json:"enforcement"`

	// Factors allowed by this policy
	AllowedFactors []FactorType `db:"allowed_factors" json:"allowed_factors"`

	// Minimum factors required (e.g., 2 for multi-factor)
	MinimumFactors int `db:"minimum_factors" json:"minimum_factors"`

	// Grace period: time after which MFA becomes mandatory (in days)
	GracePeriodDays int `db:"grace_period_days" json:"grace_period_days"`

	// Challenge TTL in seconds
	ChallengeTTLSeconds int `db:"challenge_ttl_seconds" json:"challenge_ttl_seconds"`

	// Max verification attempts before lockout
	MaxAttempts int `db:"max_attempts" json:"max_attempts"`

	// Trusted device settings
	TrustDeviceEnabled bool `db:"trust_device_enabled"    json:"trust_device_enabled"`
	TrustDeviceTTLDays int  `db:"trust_device_ttl_days"   json:"trust_device_ttl_days"`
	MaxTrustedDevices  int  `db:"max_trusted_devices"     json:"max_trusted_devices"`

	// Backup code settings
	BackupCodesCount    int  `db:"backup_codes_count" json:"backup_codes_count"`
	BackupCodesRequired bool `db:"backup_codes_required" json:"backup_codes_required"`

	// Risk-based settings
	RiskScoringEnabled bool     `db:"risk_scoring_enabled"    json:"risk_scoring_enabled"`
	RiskThreshold      int      `db:"risk_threshold"          json:"risk_threshold"` // 0-100
	RiskActions        []string `db:"risk_actions"            json:"risk_actions"`   // e.g., "require_mfa", "challenge", "block"

	// User/group targeting (K8s-style labels)
	Selector map[string]string `db:"selector" json:"selector"`

	// Resource constraints
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	ResourceVersion int64 `db:"resource_version" json:"resource_version"`
}

// PolicyRule applies to a specific scope (user, role, resource).
type PolicyRule struct {
	ID       uuid.UUID `db:"id"       json:"id"`
	PolicyID uuid.UUID `db:"policy_id" json:"policy_id"`

	// Scope: "user", "group", "role", "resource"
	Scope string `db:"scope" json:"scope"`

	// Target: user ID, group ID, role name, or resource path
	Target string `db:"target" json:"target"`

	// Priority determines ordering when multiple rules match
	Priority int `db:"priority" json:"priority"`

	// Override allows this rule to bypass parent policies
	Override bool `db:"override" json:"override"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// PolicyCondition adds fine-grained logic (time windows, IP ranges, etc).
type PolicyCondition struct {
	ID       uuid.UUID `db:"id"       json:"id"`
	PolicyID uuid.UUID `db:"policy_id" json:"policy_id"`

	// Type: "time_window", "ip_range", "geo", "device", "browser", etc.
	Type string `db:"type" json:"type"`

	// Key/Value pairs (e.g., start_hour=22, end_hour=6)
	Properties map[string]interface{} `db:"properties" json:"properties"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// IsApplicable returns true if the policy should be enforced given the user's context.
func (p *MFAPolicy) IsApplicable(userLabels map[string]string) bool {
	// K8s-style label matching: all selector labels must match
	if len(p.Selector) == 0 {
		return true // No selector = applies to everyone
	}

	for k, v := range p.Selector {
		if userLabels[k] != v {
			return false
		}
	}
	return true
}

// CanUseFactor returns true if the given factor type is allowed.
func (p *MFAPolicy) CanUseFactor(ft FactorType) bool {
	for _, allowed := range p.AllowedFactors {
		if allowed == ft {
			return true
		}
	}
	return false
}
