package apiresource

import (
	"time"
)

// Spec represents the desired state of a resource
type Spec struct {
	// Common fields
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Owner       string                 `json:"owner,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`

	// Validation rules
	ValidationRules []ValidationRule `json:"validation_rules,omitempty"`

	// Policy settings
	Policies []PolicyReference `json:"policies,omitempty"`

	// Rate limiting
	RateLimit *RateLimitSpec `json:"rate_limit,omitempty"`

	// Caching
	Cache *CacheSpec `json:"cache,omitempty"`

	// Scheduling
	Schedule *ScheduleSpec `json:"schedule,omitempty"`

	// Custom spec data
	Data map[string]interface{} `json:"data,omitempty"`
}

// ValidationRule represents a validation rule for a resource
type ValidationRule struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"` // regex, length, enum, custom
	Pattern   string   `json:"pattern,omitempty"`
	MinLength int      `json:"min_length,omitempty"`
	MaxLength int      `json:"max_length,omitempty"`
	Values    []string `json:"values,omitempty"`
	Message   string   `json:"message,omitempty"`
	Required  bool     `json:"required,omitempty"`
}

// PolicyReference references a policy to be applied
type PolicyReference struct {
	Name       string                 `json:"name"`
	Version    string                 `json:"version,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// RateLimitSpec defines rate limiting configuration
type RateLimitSpec struct {
	RequestsPerSecond int    `json:"requests_per_second"`
	Burst             int    `json:"burst,omitempty"`
	WindowSize        string `json:"window_size,omitempty"` // e.g., "1m", "1h"
	Key               string `json:"key,omitempty"`         // what to rate limit on (ip, user, api_key)
	Enabled           bool   `json:"enabled"`
}

// CacheSpec defines caching configuration
type CacheSpec struct {
	Enabled    bool   `json:"enabled"`
	TTL        string `json:"ttl,omitempty"`
	MaxSize    int    `json:"max_size,omitempty"`
	Strategy   string `json:"strategy,omitempty"` // lru, lfu, fifo
	KeyPattern string `json:"key_pattern,omitempty"`
}

// ScheduleSpec defines scheduling configuration
type ScheduleSpec struct {
	CronExpression string `json:"cron_expression"`
	Timezone       string `json:"timezone,omitempty"`
	Enabled        bool   `json:"enabled"`
}

// Status represents the actual state of a resource
type Status struct {
	// Phase of the resource lifecycle
	Phase string `json:"phase"` // Pending, Active, Failed, Terminating

	// Conditions describe the state
	Conditions []Condition `json:"conditions,omitempty"`

	// ObservedGeneration indicates which generation was last processed
	ObservedGeneration int64 `json:"observed_generation"`

	// LastUpdateTime when status was last updated
	LastUpdateTime time.Time `json:"last_update_time"`

	// LastTransitionTime when last condition changed
	LastTransitionTime time.Time `json:"last_transition_time"`

	// Message provides human-readable status
	Message string `json:"message,omitempty"`

	// Reason provides machine-readable status reason
	Reason string `json:"reason,omitempty"`

	// Statistics about the resource
	Statistics Statistics `json:"statistics,omitempty"`

	// Custom status data
	Data map[string]interface{} `json:"data,omitempty"`
}

// Condition describes an aspect of the resource's state
type Condition struct {
	// Type of condition
	Type string `json:"type"`

	// Status of the condition
	Status ConditionStatus `json:"status"` // True, False, Unknown

	// Reason is a short machine-readable explanation
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable explanation
	Message string `json:"message,omitempty"`

	// FirstObservedTime when the condition was first observed
	FirstObservedTime time.Time `json:"first_observed_time"`

	// LastObservedTime last time condition was probed
	LastObservedTime time.Time `json:"last_observed_time"`

	// LastTransitionTime when the condition last changed
	LastTransitionTime time.Time `json:"last_transition_time"`
}

// ConditionStatus represents the status value of a condition
type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// Standard condition types
const (
	ConditionTypeReady       = "Ready"
	ConditionTypeReconciling = "Reconciling"
	ConditionTypeSynced      = "Synced"
	ConditionTypeAvailable   = "Available"
	ConditionTypeFailed      = "Failed"
	ConditionTypeTerminating = "Terminating"
	ConditionTypeConfigured  = "Configured"
	ConditionTypeHealthy     = "Healthy"
)

// Statistics tracks statistics about the resource
type Statistics struct {
	CreatedCount    int64     `json:"created_count,omitempty"`
	UpdatedCount    int64     `json:"updated_count,omitempty"`
	DeletedCount    int64     `json:"deleted_count,omitempty"`
	ErrorCount      int64     `json:"error_count,omitempty"`
	LastError       string    `json:"last_error,omitempty"`
	ProcessedItems  int64     `json:"processed_items,omitempty"`
	FailedItems     int64     `json:"failed_items,omitempty"`
	AverageDuration string    `json:"average_duration,omitempty"`
	LastSuccessTime time.Time `json:"last_success_time,omitempty"`
	LastFailureTime time.Time `json:"last_failure_time,omitempty"`
}

// StatusHelper provides helper methods for status management
type StatusHelper struct {
	status *Status
}

// NewStatusHelper creates a new status helper
func NewStatusHelper() *StatusHelper {
	return &StatusHelper{
		status: &Status{
			Phase:              "Pending",
			ObservedGeneration: 0,
			LastUpdateTime:     time.Now(),
			LastTransitionTime: time.Now(),
		},
	}
}

// SetPhase sets the phase
func (sh *StatusHelper) SetPhase(phase string) *StatusHelper {
	sh.status.Phase = phase
	sh.status.LastUpdateTime = time.Now()
	return sh
}

// SetMessage sets the message
func (sh *StatusHelper) SetMessage(message string) *StatusHelper {
	sh.status.Message = message
	sh.status.LastUpdateTime = time.Now()
	return sh
}

// SetReason sets the reason
func (sh *StatusHelper) SetReason(reason string) *StatusHelper {
	sh.status.Reason = reason
	sh.status.LastUpdateTime = time.Now()
	return sh
}

// SetObservedGeneration sets the observed generation
func (sh *StatusHelper) SetObservedGeneration(generation int64) *StatusHelper {
	sh.status.ObservedGeneration = generation
	return sh
}

// AddCondition adds or updates a condition
func (sh *StatusHelper) AddCondition(condType string, status ConditionStatus, reason, message string) *StatusHelper {
	now := time.Now()

	// Check if condition already exists
	for i, cond := range sh.status.Conditions {
		if cond.Type == condType {
			if cond.Status != status {
				sh.status.Conditions[i].LastTransitionTime = now
			}
			sh.status.Conditions[i].Status = status
			sh.status.Conditions[i].Reason = reason
			sh.status.Conditions[i].Message = message
			sh.status.Conditions[i].LastObservedTime = now
			sh.status.LastTransitionTime = now
			return sh
		}
	}

	// Add new condition
	cond := Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		FirstObservedTime:  now,
		LastObservedTime:   now,
		LastTransitionTime: now,
	}
	sh.status.Conditions = append(sh.status.Conditions, cond)
	sh.status.LastTransitionTime = now

	return sh
}

// RemoveCondition removes a condition
func (sh *StatusHelper) RemoveCondition(condType string) *StatusHelper {
	filtered := make([]Condition, 0)
	for _, cond := range sh.status.Conditions {
		if cond.Type != condType {
			filtered = append(filtered, cond)
		}
	}
	sh.status.Conditions = filtered
	return sh
}

// GetCondition gets a condition
func (sh *StatusHelper) GetCondition(condType string) *Condition {
	for i, cond := range sh.status.Conditions {
		if cond.Type == condType {
			return &sh.status.Conditions[i]
		}
	}
	return nil
}

// IsReady returns true if the resource is ready
func (sh *StatusHelper) IsReady() bool {
	cond := sh.GetCondition(ConditionTypeReady)
	return cond != nil && cond.Status == ConditionTrue
}

// IsReconciling returns true if the resource is being reconciled
func (sh *StatusHelper) IsReconciling() bool {
	cond := sh.GetCondition(ConditionTypeReconciling)
	return cond != nil && cond.Status == ConditionTrue
}

// IsFailed returns true if the resource has failed
func (sh *StatusHelper) IsFailed() bool {
	cond := sh.GetCondition(ConditionTypeFailed)
	return cond != nil && cond.Status == ConditionTrue
}

// Build returns the built status
func (sh *StatusHelper) Build() *Status {
	sh.status.LastUpdateTime = time.Now()
	return sh.status
}

// SpecHelper provides helper methods for spec management
type SpecHelper struct {
	spec *Spec
}

// NewSpecHelper creates a new spec helper
func NewSpecHelper(specType string) *SpecHelper {
	return &SpecHelper{
		spec: &Spec{
			Type:   specType,
			Config: make(map[string]interface{}),
			Data:   make(map[string]interface{}),
		},
	}
}

// SetDescription sets the description
func (sh *SpecHelper) SetDescription(desc string) *SpecHelper {
	sh.spec.Description = desc
	return sh
}

// SetOwner sets the owner
func (sh *SpecHelper) SetOwner(owner string) *SpecHelper {
	sh.spec.Owner = owner
	return sh
}

// SetConfig sets a config value
func (sh *SpecHelper) SetConfig(key string, value interface{}) *SpecHelper {
	sh.spec.Config[key] = value
	return sh
}

// SetData sets a data value
func (sh *SpecHelper) SetData(key string, value interface{}) *SpecHelper {
	sh.spec.Data[key] = value
	return sh
}

// AddPolicy adds a policy
func (sh *SpecHelper) AddPolicy(name, version string, params map[string]interface{}) *SpecHelper {
	policy := PolicyReference{
		Name:       name,
		Version:    version,
		Parameters: params,
	}
	sh.spec.Policies = append(sh.spec.Policies, policy)
	return sh
}

// SetRateLimit sets rate limit spec
func (sh *SpecHelper) SetRateLimit(rps int, burst int, enabled bool) *SpecHelper {
	sh.spec.RateLimit = &RateLimitSpec{
		RequestsPerSecond: rps,
		Burst:             burst,
		Enabled:           enabled,
	}
	return sh
}

// SetCache sets cache spec
func (sh *SpecHelper) SetCache(enabled bool, ttl string, maxSize int) *SpecHelper {
	sh.spec.Cache = &CacheSpec{
		Enabled: enabled,
		TTL:     ttl,
		MaxSize: maxSize,
	}
	return sh
}

// Build returns the built spec
func (sh *SpecHelper) Build() *Spec {
	return sh.spec
}
