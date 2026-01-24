package lifecycle

import (
	"fmt"
	"time"
)

// LifecyclePolicy defines resource lifecycle policies
type LifecyclePolicy struct {
	ID            string
	Name          string
	Type          string
	Version       string
	Enabled       bool
	Rules         []LifecycleRule
	DeletePolicy  DeletePolicy
	TransitionPolicy TransitionPolicy
	Description   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// LifecycleRule defines a lifecycle rule
type LifecycleRule struct {
	ID            string
	Filter        LifecycleFilter
	Transitions   []Transition
	Expiration    *Expiration
	NoncurrentVersionTransitions []Transition
	NoncurrentVersionExpiration  *Expiration
	AbortIncompleteMultipart *AbortIncompleteMultipart
}

// LifecycleFilter filters objects for lifecycle rules
type LifecycleFilter struct {
	Prefix       string
	Tags         map[string]string
	ObjectSize   *SizeRange
	CreatedTime  *TimeRange
}

// SizeRange defines a size range filter
type SizeRange struct {
	Min int64
	Max int64
}

// TimeRange defines a time range filter
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Transition defines a transition action
type Transition struct {
	Days         int
	Date         time.Time
	StorageClass string // "STANDARD", "STANDARD_IA", "GLACIER", "DEEP_ARCHIVE"
	Status       string // "pending", "completed", "failed"
}

// Expiration defines an expiration action
type Expiration struct {
	Days int
	Date time.Time
}

// AbortIncompleteMultipart defines cleanup of incomplete uploads
type AbortIncompleteMultipart struct {
	DaysAfterInitiation int
}

// DeletePolicy defines deletion behavior
type DeletePolicy struct {
	RetentionPeriod time.Duration
	RequireApproval bool
	NotifyBefore    time.Duration
	CanRecover      bool
}

// TransitionPolicy defines storage class transition behavior
type TransitionPolicy struct {
	AutoTransition bool
	MinimumDays    int
	AllowedClasses []string
}

// GetID returns policy ID
func (lp *LifecyclePolicy) GetID() string {
	return lp.ID
}

// GetName returns policy name
func (lp *LifecyclePolicy) GetName() string {
	return lp.Name
}

// GetType returns policy type
func (lp *LifecyclePolicy) GetType() string {
	return lp.Type
}

// GetVersion returns version
func (lp *LifecyclePolicy) GetVersion() string {
	return lp.Version
}

// GetEnabled returns if enabled
func (lp *LifecyclePolicy) GetEnabled() bool {
	return lp.Enabled
}

// Validate validates the policy
func (lp *LifecyclePolicy) Validate() error {
	if lp.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}
	if lp.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}
	if len(lp.Rules) == 0 {
		return fmt.Errorf("at least one lifecycle rule must be defined")
	}
	return nil
}

// LifecycleManager manages resource lifecycles
type LifecycleManager struct {
	policies     []*LifecyclePolicy
	transitions  map[string]TransitionState
	expirations  map[string]ExpirationState
}

// TransitionState tracks transition progress
type TransitionState struct {
	ResourceID     string
	SourceClass    string
	TargetClass    string
	StartTime      time.Time
	CompletionTime time.Time
	Status         string // "pending", "in-progress", "completed", "failed"
	ErrorMessage   string
}

// ExpirationState tracks expiration progress
type ExpirationState struct {
	ResourceID     string
	ExpirationDate time.Time
	ReminderSent   bool
	Status         string // "pending", "scheduled", "deleted"
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{
		policies:    make([]*LifecyclePolicy, 0),
		transitions: make(map[string]TransitionState),
		expirations: make(map[string]ExpirationState),
	}
}

// RegisterPolicy registers a lifecycle policy
func (lm *LifecycleManager) RegisterPolicy(policy *LifecyclePolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	lm.policies = append(lm.policies, policy)
	return nil
}

// ApplyPolicies applies applicable policies to a resource
func (lm *LifecycleManager) ApplyPolicies(resourceID string, resourceMeta map[string]interface{}) []LifecycleAction {
	var actions []LifecycleAction

	for _, policy := range lm.policies {
		if !policy.Enabled {
			continue
		}

		for _, rule := range policy.Rules {
			if !lm.matchesFilter(resourceMeta, rule.Filter) {
				continue
			}

			// Check transitions
			for _, transition := range rule.Transitions {
				if lm.shouldTransition(resourceMeta, transition) {
					actions = append(actions, LifecycleAction{
						Type:         "transition",
						ResourceID:   resourceID,
						TargetClass:  transition.StorageClass,
						ScheduledFor: time.Now().Add(time.Duration(transition.Days) * 24 * time.Hour),
					})
				}
			}

			// Check expiration
			if rule.Expiration != nil {
				if lm.shouldExpire(resourceMeta, rule.Expiration) {
					actions = append(actions, LifecycleAction{
						Type:         "expire",
						ResourceID:   resourceID,
						ScheduledFor: time.Now().Add(time.Duration(rule.Expiration.Days) * 24 * time.Hour),
					})
				}
			}
		}
	}

	return actions
}

func (lm *LifecycleManager) matchesFilter(resourceMeta map[string]interface{}, filter LifecycleFilter) bool {
	// Check prefix
	if filter.Prefix != "" {
		if name, ok := resourceMeta["name"].(string); !ok || !stringStartsWith(name, filter.Prefix) {
			return false
		}
	}

	// Check tags
	if len(filter.Tags) > 0 {
		tags, ok := resourceMeta["tags"].(map[string]string)
		if !ok {
			return false
		}
		for k, v := range filter.Tags {
			if tags[k] != v {
				return false
			}
		}
	}

	return true
}

func (lm *LifecycleManager) shouldTransition(resourceMeta map[string]interface{}, transition Transition) bool {
	createdAt, ok := resourceMeta["created_at"].(time.Time)
	if !ok {
		return false
	}

	daysOld := int(time.Since(createdAt).Hours() / 24)
	return daysOld >= transition.Days
}

func (lm *LifecycleManager) shouldExpire(resourceMeta map[string]interface{}, expiration *Expiration) bool {
	createdAt, ok := resourceMeta["created_at"].(time.Time)
	if !ok {
		return false
	}

	daysOld := int(time.Since(createdAt).Hours() / 24)
	return daysOld >= expiration.Days
}

// LifecycleAction represents an action to perform
type LifecycleAction struct {
	Type        string // "transition", "expire", "delete", "tag"
	ResourceID  string
	TargetClass string
	ScheduledFor time.Time
	Executed    bool
	ExecutedAt  time.Time
	Error       string
}

// ExecuteAction executes a lifecycle action
func (lm *LifecycleManager) ExecuteAction(action LifecycleAction) error {
	switch action.Type {
	case "transition":
		return lm.executeTransition(action)
	case "expire":
		return lm.executeExpiration(action)
	case "delete":
		return lm.executeDeletion(action)
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func (lm *LifecycleManager) executeTransition(action LifecycleAction) error {
	state := TransitionState{
		ResourceID:  action.ResourceID,
		TargetClass: action.TargetClass,
		StartTime:   time.Now(),
		Status:      "in-progress",
	}
	lm.transitions[action.ResourceID] = state
	return nil
}

func (lm *LifecycleManager) executeExpiration(action LifecycleAction) error {
	state := ExpirationState{
		ResourceID:     action.ResourceID,
		ExpirationDate: action.ScheduledFor,
		Status:         "scheduled",
	}
	lm.expirations[action.ResourceID] = state
	return nil
}

func (lm *LifecycleManager) executeDeletion(action LifecycleAction) error {
	// In production, actually delete the resource
	return nil
}

// ResourceVersionControl manages versioning of resources
type ResourceVersionControl struct {
	KeepVersions     int
	DeleteNoncurrent bool
	TimeToDelete     time.Duration
}

// DataLifecycleStage represents a stage in data lifecycle
type DataLifecycleStage struct {
	Name            string
	Duration        time.Duration
	Action          string // "retain", "transition", "delete", "archive"
	TargetLocation  string
	Retention       time.Duration
}

// DataLifecycleFlow defines the complete lifecycle flow
type DataLifecycleFlow struct {
	ID     string
	Name   string
	Stages []DataLifecycleStage
}

// ExecuteFlow executes a lifecycle flow
func (dlf *DataLifecycleFlow) ExecuteFlow(resourceID string) ([]LifecycleAction, error) {
	var actions []LifecycleAction

	for _, stage := range dlf.Stages {
		action := LifecycleAction{
			ResourceID:   resourceID,
			Type:         stage.Action,
			TargetClass:  stage.TargetLocation,
			ScheduledFor: time.Now().Add(stage.Duration),
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// Helper function
func stringStartsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
