package policyengine

import (
	"fmt"
	"sync"
	"time"
)

// PolicyEngine is the core policy evaluation engine
type PolicyEngine struct {
	mu           sync.RWMutex
	policies     map[string]Policy
	evaluators   map[string]PolicyEvaluator
	cache        map[string]EvaluationResult
	cacheTTL     time.Duration
	auditLog     []PolicyAuditLog
	maxAuditLogs int
}

// Policy defines the base policy interface
type Policy interface {
	GetID() string
	GetName() string
	GetType() string
	GetVersion() string
	GetEnabled() bool
	GetRules() []PolicyRule
	Validate() error
}

// PolicyRule defines a single rule within a policy
type PolicyRule struct {
	ID          string
	Name        string
	Priority    int
	Condition   string
	Action      string
	Effect      string // "allow" or "deny"
	Description string
}

// PolicyEvaluator evaluates policies
type PolicyEvaluator interface {
	Evaluate(context PolicyContext, policy Policy) (bool, error)
}

// PolicyContext holds context for policy evaluation
type PolicyContext struct {
	RequestID     string
	UserID        string
	Action        string
	Resource      string
	ResourceType  string
	Namespace     string
	Attributes    map[string]interface{}
	Timestamp     time.Time
	SourceIP      string
	RequestMethod string
}

// EvaluationResult holds the result of policy evaluation
type EvaluationResult struct {
	PolicyID   string
	Allowed    bool
	Reason     string
	Timestamp  time.Time
	DurationMs int64
}

// PolicyAuditLog holds audit log information
type PolicyAuditLog struct {
	Timestamp    time.Time
	PolicyID     string
	Action       string
	Context      PolicyContext
	Result       EvaluationResult
	ErrorMessage string
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		policies:     make(map[string]Policy),
		evaluators:   make(map[string]PolicyEvaluator),
		cache:        make(map[string]EvaluationResult),
		cacheTTL:     5 * time.Minute,
		auditLog:     make([]PolicyAuditLog, 0),
		maxAuditLogs: 10000,
	}
}

// RegisterPolicy registers a policy
func (pe *PolicyEngine) RegisterPolicy(policy Policy) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if err := policy.Validate(); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	if _, exists := pe.policies[policy.GetID()]; exists {
		return fmt.Errorf("policy already registered: %s", policy.GetID())
	}

	pe.policies[policy.GetID()] = policy
	return nil
}

// UnregisterPolicy unregisters a policy
func (pe *PolicyEngine) UnregisterPolicy(policyID string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if _, exists := pe.policies[policyID]; !exists {
		return fmt.Errorf("policy not found: %s", policyID)
	}

	delete(pe.policies, policyID)
	return nil
}

// RegisterEvaluator registers a policy evaluator
func (pe *PolicyEngine) RegisterEvaluator(policyType string, evaluator PolicyEvaluator) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.evaluators[policyType] = evaluator
}

// EvaluatePolicy evaluates a single policy
func (pe *PolicyEngine) EvaluatePolicy(policyID string, context PolicyContext) (EvaluationResult, error) {
	pe.mu.RLock()
	policy, exists := pe.policies[policyID]
	pe.mu.RUnlock()

	if !exists {
		return EvaluationResult{}, fmt.Errorf("policy not found: %s", policyID)
	}

	if !policy.GetEnabled() {
		return EvaluationResult{
			PolicyID:  policyID,
			Allowed:   true,
			Reason:    "policy disabled",
			Timestamp: time.Now(),
		}, nil
	}

	pe.mu.RLock()
	evaluator, exists := pe.evaluators[policy.GetType()]
	pe.mu.RUnlock()

	if !exists {
		return EvaluationResult{}, fmt.Errorf("no evaluator for policy type: %s", policy.GetType())
	}

	startTime := time.Now()
	allowed, err := evaluator.Evaluate(context, policy)
	duration := time.Since(startTime)

	result := EvaluationResult{
		PolicyID:   policyID,
		Allowed:    allowed,
		Timestamp:  startTime,
		DurationMs: duration.Milliseconds(),
	}

	if err != nil {
		result.Reason = fmt.Sprintf("evaluation error: %v", err)
	} else {
		if allowed {
			result.Reason = "policy allowed"
		} else {
			result.Reason = "policy denied"
		}
	}

	pe.logAudit(PolicyAuditLog{
		Timestamp:    time.Now(),
		PolicyID:     policyID,
		Action:       "evaluate",
		Context:      context,
		Result:       result,
		ErrorMessage: fmt.Sprintf("%v", err),
	})

	return result, err
}

// EvaluateAll evaluates all applicable policies
func (pe *PolicyEngine) EvaluateAll(context PolicyContext) ([]EvaluationResult, error) {
	pe.mu.RLock()
	policies := make([]Policy, 0, len(pe.policies))
	for _, p := range pe.policies {
		policies = append(policies, p)
	}
	pe.mu.RUnlock()

	results := make([]EvaluationResult, 0)
	var finalErr error

	for _, policy := range policies {
		result, err := pe.EvaluatePolicy(policy.GetID(), context)
		results = append(results, result)
		if err != nil {
			finalErr = err
		}
		if !result.Allowed {
			return results, finalErr
		}
	}

	return results, finalErr
}

// logAudit logs a policy audit event
func (pe *PolicyEngine) logAudit(log PolicyAuditLog) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	pe.auditLog = append(pe.auditLog, log)

	// Maintain max size
	if len(pe.auditLog) > pe.maxAuditLogs {
		pe.auditLog = pe.auditLog[len(pe.auditLog)-pe.maxAuditLogs:]
	}
}

// GetAuditLogs returns audit logs
func (pe *PolicyEngine) GetAuditLogs(limit int) []PolicyAuditLog {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	if limit > len(pe.auditLog) {
		limit = len(pe.auditLog)
	}

	logs := make([]PolicyAuditLog, limit)
	copy(logs, pe.auditLog[len(pe.auditLog)-limit:])
	return logs
}

// GetPolicy returns a policy by ID
func (pe *PolicyEngine) GetPolicy(policyID string) (Policy, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	policy, exists := pe.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", policyID)
	}

	return policy, nil
}

// ListPolicies returns all policies of a specific type
func (pe *PolicyEngine) ListPolicies(policyType string) []Policy {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	var policies []Policy
	for _, p := range pe.policies {
		if policyType == "" || p.GetType() == policyType {
			policies = append(policies, p)
		}
	}

	return policies
}

// UpdatePolicy updates a policy
func (pe *PolicyEngine) UpdatePolicy(policyID string, policy Policy) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if _, exists := pe.policies[policyID]; !exists {
		return fmt.Errorf("policy not found: %s", policyID)
	}

	if err := policy.Validate(); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	pe.policies[policyID] = policy
	return nil
}

// EnablePolicy enables a policy
func (pe *PolicyEngine) EnablePolicy(policyID string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	policy, exists := pe.policies[policyID]
	if !exists {
		return fmt.Errorf("policy not found: %s", policyID)
	}

	// Create a wrapper to enable the policy
	if basePolicyCfg, ok := policy.(*BasePolicy); ok {
		basePolicyCfg.Enabled = true
	}

	return nil
}

// DisablePolicy disables a policy
func (pe *PolicyEngine) DisablePolicy(policyID string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	policy, exists := pe.policies[policyID]
	if !exists {
		return fmt.Errorf("policy not found: %s", policyID)
	}

	if basePolicyCfg, ok := policy.(*BasePolicy); ok {
		basePolicyCfg.Enabled = false
	}

	return nil
}

// BasePolicy provides a base implementation of Policy
type BasePolicy struct {
	ID          string
	Name        string
	Type        string
	Version     string
	Enabled     bool
	Rules       []PolicyRule
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GetID returns the policy ID
func (bp *BasePolicy) GetID() string {
	return bp.ID
}

// GetName returns the policy name
func (bp *BasePolicy) GetName() string {
	return bp.Name
}

// GetType returns the policy type
func (bp *BasePolicy) GetType() string {
	return bp.Type
}

// GetVersion returns the policy version
func (bp *BasePolicy) GetVersion() string {
	return bp.Version
}

// GetEnabled returns if the policy is enabled
func (bp *BasePolicy) GetEnabled() bool {
	return bp.Enabled
}

// GetRules returns the policy rules
func (bp *BasePolicy) GetRules() []PolicyRule {
	return bp.Rules
}

// Validate validates the policy
func (bp *BasePolicy) Validate() error {
	if bp.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}
	if bp.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}
	if bp.Type == "" {
		return fmt.Errorf("policy type cannot be empty")
	}
	if len(bp.Rules) == 0 {
		return fmt.Errorf("policy must have at least one rule")
	}
	return nil
}

// PolicyBuilder builds policies fluently
type PolicyBuilder struct {
	policy *BasePolicy
}

// NewPolicyBuilder creates a new policy builder
func NewPolicyBuilder(id, name, policyType string) *PolicyBuilder {
	return &PolicyBuilder{
		policy: &BasePolicy{
			ID:        id,
			Name:      name,
			Type:      policyType,
			Version:   "1.0",
			Enabled:   true,
			Rules:     make([]PolicyRule, 0),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// WithVersion sets the version
func (pb *PolicyBuilder) WithVersion(version string) *PolicyBuilder {
	pb.policy.Version = version
	return pb
}

// WithDescription sets the description
func (pb *PolicyBuilder) WithDescription(description string) *PolicyBuilder {
	pb.policy.Description = description
	return pb
}

// WithEnabled sets if enabled
func (pb *PolicyBuilder) WithEnabled(enabled bool) *PolicyBuilder {
	pb.policy.Enabled = enabled
	return pb
}

// AddRule adds a rule
func (pb *PolicyBuilder) AddRule(rule PolicyRule) *PolicyBuilder {
	pb.policy.Rules = append(pb.policy.Rules, rule)
	return pb
}

// Build builds the policy
func (pb *PolicyBuilder) Build() (*BasePolicy, error) {
	if err := pb.policy.Validate(); err != nil {
		return nil, err
	}
	return pb.policy, nil
}

// PolicyStatus holds the status of a policy
type PolicyStatus struct {
	PolicyID         string
	Enabled          bool
	TotalEvaluations int64
	SuccessfulEvals  int64
	FailedEvals      int64
	AvgEvalTime      time.Duration
	LastEvaluated    time.Time
	ErrorRate        float64
}
