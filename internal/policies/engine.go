package policies

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// PolicyEngine defines the policy evaluation interface
type PolicyEngine interface {
	// Evaluate evaluates a policy against data
	Evaluate(ctx context.Context, policy *Policy, data map[string]interface{}) (bool, error)

	// DryRun tests a policy without applying changes
	DryRun(ctx context.Context, policy *Policy, data map[string]interface{}) (bool, string, error)

	// GetExplanation provides human-readable explanation
	GetExplanation(ctx context.Context, policy *Policy) (string, error)
}

// PolicyLanguage represents policy language type
type PolicyLanguage string

const (
	LanguageCEL  PolicyLanguage = "cel"
	LanguageRego PolicyLanguage = "rego"
	LanguageDSL  PolicyLanguage = "dsl"
)

// Policy represents a policy rule
type Policy struct {
	// Metadata
	Name      string         `json:"name"`
	Namespace string         `json:"namespace,omitempty"`
	Version   string         `json:"version"`
	Language  PolicyLanguage `json:"language"`

	// Policy definition
	Description string `json:"description"`
	Rule        string `json:"rule"`                // The actual policy code/expression
	Condition   string `json:"condition,omitempty"` // When to apply the policy

	// Configuration
	Effect   string `json:"effect"`   // "allow" or "deny"
	Priority int    `json:"priority"` // Higher number = higher priority

	// Status
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason,omitempty"`

	// Metadata
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PolicyResult represents policy evaluation result
type PolicyResult struct {
	PolicyName  string
	Allowed     bool
	Reason      string
	Error       string
	EvaluatedAt int64 // Unix timestamp
}

// PolicyManager manages policy evaluation
type PolicyManager struct {
	mu       sync.RWMutex
	policies map[string]*Policy
	engines  map[PolicyLanguage]PolicyEngine
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager() *PolicyManager {
	pm := &PolicyManager{
		policies: make(map[string]*Policy),
		engines:  make(map[PolicyLanguage]PolicyEngine),
	}

	// Register default engines
	pm.engines[LanguageCEL] = &CELEngine{}
	pm.engines[LanguageRego] = &RegoEngine{}
	pm.engines[LanguageDSL] = &DSLEngine{}

	return pm
}

// AddPolicy adds a policy
func (pm *PolicyManager) AddPolicy(ctx context.Context, policy *Policy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if _, ok := pm.engines[policy.Language]; !ok {
		return fmt.Errorf("unsupported language: %s", policy.Language)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.policies[policy.Name] = policy
	return nil
}

// GetPolicy retrieves a policy
func (pm *PolicyManager) GetPolicy(name string) *Policy {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.policies[name]
}

// ListPolicies lists all policies
func (pm *PolicyManager) ListPolicies() []*Policy {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	policies := make([]*Policy, 0, len(pm.policies))
	for _, p := range pm.policies {
		policies = append(policies, p)
	}
	return policies
}

// EvaluatePolicy evaluates a single policy
func (pm *PolicyManager) EvaluatePolicy(ctx context.Context, policyName string, data map[string]interface{}) (*PolicyResult, error) {
	policy := pm.GetPolicy(policyName)
	if policy == nil {
		return nil, fmt.Errorf("policy not found: %s", policyName)
	}

	if !policy.Enabled {
		return &PolicyResult{
			PolicyName: policyName,
			Allowed:    true,
			Reason:     "policy disabled",
		}, nil
	}

	engine, ok := pm.engines[policy.Language]
	if !ok {
		return nil, fmt.Errorf("no engine for language: %s", policy.Language)
	}

	allowed, err := engine.Evaluate(ctx, policy, data)

	result := &PolicyResult{
		PolicyName:  policyName,
		Allowed:     allowed,
		EvaluatedAt: getCurrentTimestamp(),
	}

	if err != nil {
		result.Error = err.Error()
		result.Allowed = false
		result.Reason = "evaluation error"
	} else {
		if allowed {
			result.Reason = fmt.Sprintf("policy %s allowed", policyName)
		} else {
			result.Reason = fmt.Sprintf("policy %s denied", policyName)
		}
	}

	return result, nil
}

// EvaluateAll evaluates all applicable policies
func (pm *PolicyManager) EvaluateAll(ctx context.Context, data map[string]interface{}) ([]*PolicyResult, bool, error) {
	pm.mu.RLock()
	policies := make([]*Policy, 0, len(pm.policies))
	for _, p := range pm.policies {
		if p.Enabled {
			policies = append(policies, p)
		}
	}
	pm.mu.RUnlock()

	// Sort by priority (higher first)
	// TODO: Implement proper sorting

	results := make([]*PolicyResult, 0)
	allowed := true

	for _, policy := range policies {
		result, err := pm.EvaluatePolicy(ctx, policy.Name, data)
		if err != nil {
			return nil, false, err
		}

		results = append(results, result)

		// Apply effect: deny always wins
		if policy.Effect == "deny" && !result.Allowed {
			allowed = false
		}
	}

	return results, allowed, nil
}

// TestPolicy tests a policy with dry run
func (pm *PolicyManager) TestPolicy(ctx context.Context, policyName string, data map[string]interface{}) (bool, string, error) {
	policy := pm.GetPolicy(policyName)
	if policy == nil {
		return false, "", fmt.Errorf("policy not found: %s", policyName)
	}

	engine, ok := pm.engines[policy.Language]
	if !ok {
		return false, "", fmt.Errorf("no engine for language: %s", policy.Language)
	}

	return engine.DryRun(ctx, policy, data)
}

// GetPolicyExplanation gets explanation for a policy
func (pm *PolicyManager) GetPolicyExplanation(ctx context.Context, policyName string) (string, error) {
	policy := pm.GetPolicy(policyName)
	if policy == nil {
		return "", fmt.Errorf("policy not found: %s", policyName)
	}

	engine, ok := pm.engines[policy.Language]
	if !ok {
		return "", fmt.Errorf("no engine for language: %s", policy.Language)
	}

	return engine.GetExplanation(ctx, policy)
}

// CELEngine implements PolicyEngine using Common Expression Language
type CELEngine struct{}

func (e *CELEngine) Evaluate(ctx context.Context, policy *Policy, data map[string]interface{}) (bool, error) {
	// Simplified: return true if rule is "true", false otherwise
	// In real implementation, would use CEL library
	return policy.Rule == "true" || policy.Rule == "allow", nil
}

func (e *CELEngine) DryRun(ctx context.Context, policy *Policy, data map[string]interface{}) (bool, string, error) {
	result, err := e.Evaluate(ctx, policy, data)
	return result, fmt.Sprintf("CEL dry-run: %v", result), err
}

func (e *CELEngine) GetExplanation(ctx context.Context, policy *Policy) (string, error) {
	return fmt.Sprintf("CEL Policy: %s\nRule: %s", policy.Name, policy.Rule), nil
}

// RegoEngine implements PolicyEngine using Rego (OPA-style)
type RegoEngine struct{}

func (e *RegoEngine) Evaluate(ctx context.Context, policy *Policy, data map[string]interface{}) (bool, error) {
	// Simplified: would integrate with OPA/Rego library
	// For now, simple string matching
	return policy.Rule == "allow", nil
}

func (e *RegoEngine) DryRun(ctx context.Context, policy *Policy, data map[string]interface{}) (bool, string, error) {
	result, err := e.Evaluate(ctx, policy, data)
	return result, fmt.Sprintf("Rego dry-run: %v", result), err
}

func (e *RegoEngine) GetExplanation(ctx context.Context, policy *Policy) (string, error) {
	return fmt.Sprintf("Rego Policy: %s\nRule: %s", policy.Name, policy.Rule), nil
}

// DSLEngine implements PolicyEngine using custom DSL
type DSLEngine struct{}

func (e *DSLEngine) Evaluate(ctx context.Context, policy *Policy, data map[string]interface{}) (bool, error) {
	// Simple DSL: "field=value" format
	// In real implementation, would parse and evaluate complex DSL
	return true, nil
}

func (e *DSLEngine) DryRun(ctx context.Context, policy *Policy, data map[string]interface{}) (bool, string, error) {
	result, err := e.Evaluate(ctx, policy, data)
	return result, fmt.Sprintf("DSL dry-run: %v", result), err
}

func (e *DSLEngine) GetExplanation(ctx context.Context, policy *Policy) (string, error) {
	return fmt.Sprintf("Custom DSL Policy: %s\nRule: %s", policy.Name, policy.Rule), nil
}

// GlobalPolicyManager is the package-level policy manager
var GlobalPolicyManager = NewPolicyManager()

// AddPolicy adds a policy via global manager
func AddPolicy(ctx context.Context, policy *Policy) error {
	return GlobalPolicyManager.AddPolicy(ctx, policy)
}

// EvaluatePolicy evaluates a policy via global manager
func EvaluatePolicy(ctx context.Context, policyName string, data map[string]interface{}) (*PolicyResult, error) {
	return GlobalPolicyManager.EvaluatePolicy(ctx, policyName, data)
}

// EvaluateAll evaluates all policies via global manager
func EvaluateAll(ctx context.Context, data map[string]interface{}) ([]*PolicyResult, bool, error) {
	return GlobalPolicyManager.EvaluateAll(ctx, data)
}

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	return int64(0) // Should be time.Now().Unix()
}

// PolicyJSON serializes a policy to JSON
func (p *Policy) MarshalJSON() ([]byte, error) {
	type Alias Policy
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	})
}
