package policies

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"
)

// PolicyEngine evaluates policies against resources
type PolicyEngine struct {
	mu              sync.RWMutex
	policies        map[string]*Policy
	rules           map[string]*PolicyRule
	evaluationCache map[string]*EvaluationResult
	cacheTTL        time.Duration
}

// Policy defines a policy
type Policy struct {
	Name        string
	Namespace   string
	Version     string
	Kind        string // Target resource kind
	Description string
	Enabled     bool
	Rules       []*PolicyRule
	Actions     []PolicyAction
	Exceptions  []*PolicyException
	Priority    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PolicyRule defines a rule in a policy
type PolicyRule struct {
	Name        string
	Description string
	Effect      string // Allow, Deny
	Subjects    []Subject
	Actions     []string
	Resources   []string
	Conditions  []Condition
}

// Subject represents a policy subject
type Subject struct {
	Type  string // user, group, serviceAccount
	Name  string
	Match string // exact, prefix, pattern
}

// Condition represents a policy condition
type Condition struct {
	Field    string
	Operator string // equals, notEquals, in, notIn, matches, gt, lt, exists
	Value    interface{}
}

// PolicyAction defines an action to take
type PolicyAction struct {
	Type   string // log, alert, block, mutate, transform
	Config map[string]interface{}
}

// PolicyException defines an exception to a policy
type PolicyException struct {
	Name        string
	Description string
	Resources   []string
	Users       []string
	ExpiresAt   *time.Time
}

// EvaluationResult represents a policy evaluation result
type EvaluationResult struct {
	PolicyName   string
	Allowed      bool
	Reason       string
	MatchedRules []*PolicyRule
	Actions      []PolicyAction
	Timestamp    time.Time
	CachedUntil  time.Time
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine(cacheTTL time.Duration) *PolicyEngine {
	return &PolicyEngine{
		policies:        make(map[string]*Policy),
		rules:           make(map[string]*PolicyRule),
		evaluationCache: make(map[string]*EvaluationResult),
		cacheTTL:        cacheTTL,
	}
}

// CreatePolicy creates a new policy
func (pe *PolicyEngine) CreatePolicy(ctx context.Context, policy *Policy) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if _, exists := pe.policies[policy.Name]; exists {
		return fmt.Errorf("policy %s already exists", policy.Name)
	}

	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	pe.policies[policy.Name] = policy

	// Register rules
	for _, rule := range policy.Rules {
		ruleKey := policy.Name + "/" + rule.Name
		pe.rules[ruleKey] = rule
	}

	return nil
}

// GetPolicy gets a policy
func (pe *PolicyEngine) GetPolicy(name string) *Policy {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	return pe.policies[name]
}

// ListPolicies lists policies
func (pe *PolicyEngine) ListPolicies(ctx context.Context) []*Policy {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	policies := make([]*Policy, 0, len(pe.policies))
	for _, p := range pe.policies {
		policies = append(policies, p)
	}

	return policies
}

// UpdatePolicy updates a policy
func (pe *PolicyEngine) UpdatePolicy(ctx context.Context, policy *Policy) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	existing, exists := pe.policies[policy.Name]
	if !exists {
		return fmt.Errorf("policy %s not found", policy.Name)
	}

	policy.CreatedAt = existing.CreatedAt
	policy.UpdatedAt = time.Now()

	// Clear old rules
	for _, rule := range existing.Rules {
		ruleKey := policy.Name + "/" + rule.Name
		delete(pe.rules, ruleKey)
	}

	// Register new rules
	for _, rule := range policy.Rules {
		ruleKey := policy.Name + "/" + rule.Name
		pe.rules[ruleKey] = rule
	}

	pe.policies[policy.Name] = policy

	// Invalidate cache
	for key := range pe.evaluationCache {
		if key[:len(policy.Name)] == policy.Name {
			delete(pe.evaluationCache, key)
		}
	}

	return nil
}

// DeletePolicy deletes a policy
func (pe *PolicyEngine) DeletePolicy(ctx context.Context, name string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	policy, exists := pe.policies[name]
	if !exists {
		return fmt.Errorf("policy %s not found", name)
	}

	// Clear rules
	for _, rule := range policy.Rules {
		ruleKey := name + "/" + rule.Name
		delete(pe.rules, ruleKey)
	}

	delete(pe.policies, name)

	// Invalidate cache
	for key := range pe.evaluationCache {
		if key[:len(name)] == name {
			delete(pe.evaluationCache, key)
		}
	}

	return nil
}

// EvaluatePolicy evaluates policies
func (pe *PolicyEngine) EvaluatePolicy(ctx context.Context, subject Subject, action string, resource map[string]interface{}, resourceKind string) (*EvaluationResult, error) {
	cacheKey := subject.Type + ":" + subject.Name + ":" + action + ":" + resourceKind

	// Check cache
	pe.mu.RLock()
	if cached, exists := pe.evaluationCache[cacheKey]; exists {
		if time.Now().Before(cached.CachedUntil) {
			pe.mu.RUnlock()
			return cached, nil
		}
	}
	pe.mu.RUnlock()

	// Evaluate
	result := &EvaluationResult{
		Allowed:      true,
		MatchedRules: make([]*PolicyRule, 0),
		Actions:      make([]PolicyAction, 0),
		Timestamp:    time.Now(),
		CachedUntil:  time.Now().Add(pe.cacheTTL),
	}

	pe.mu.RLock()
	defer pe.mu.RUnlock()

	// Get enabled policies
	var applicablePolicies []*Policy
	for _, policy := range pe.policies {
		if !policy.Enabled {
			continue
		}
		if policy.Kind != resourceKind && policy.Kind != "*" {
			continue
		}
		applicablePolicies = append(applicablePolicies, policy)
	}

	// Evaluate rules
	for _, policy := range applicablePolicies {
		for _, rule := range policy.Rules {
			// Check if rule applies to subject
			if !pe.matchSubject(subject, rule.Subjects) {
				continue
			}

			// Check if rule applies to action
			if !pe.matchAction(action, rule.Actions) {
				continue
			}

			// Check if rule applies to resource
			if !pe.matchResources(resource, rule.Resources) {
				continue
			}

			// Check conditions
			if !pe.evaluateConditions(resource, rule.Conditions) {
				continue
			}

			// Rule matches
			result.MatchedRules = append(result.MatchedRules, rule)

			// Check exception
			if pe.hasException(policy, resource, subject.Name) {
				continue
			}

			if rule.Effect == "Deny" {
				result.Allowed = false
				result.Reason = fmt.Sprintf("Denied by rule %s in policy %s", rule.Name, policy.Name)
				result.Actions = policy.Actions
				break
			} else {
				result.PolicyName = policy.Name
			}
		}

		if !result.Allowed {
			break
		}
	}

	// Cache result
	pe.evaluationCache[cacheKey] = result

	return result, nil
}

// matchSubject checks if subject matches
func (pe *PolicyEngine) matchSubject(subject Subject, policySubjects []Subject) bool {
	for _, ps := range policySubjects {
		if ps.Type != subject.Type {
			continue
		}

		switch ps.Match {
		case "exact":
			if ps.Name == subject.Name {
				return true
			}
		case "prefix":
			if len(subject.Name) >= len(ps.Name) && subject.Name[:len(ps.Name)] == ps.Name {
				return true
			}
		case "pattern":
			if re, err := regexp.Compile(ps.Name); err == nil && re.MatchString(subject.Name) {
				return true
			}
		default:
			if ps.Name == subject.Name {
				return true
			}
		}
	}

	return false
}

// matchAction checks if action matches
func (pe *PolicyEngine) matchAction(action string, policyActions []string) bool {
	for _, pa := range policyActions {
		if pa == "*" || pa == action {
			return true
		}
		if re, err := regexp.Compile(pa); err == nil && re.MatchString(action) {
			return true
		}
	}

	return false
}

// matchResources checks if resource matches
func (pe *PolicyEngine) matchResources(resource map[string]interface{}, policyResources []string) bool {
	if len(policyResources) == 0 {
		return true
	}

	resourceName, ok := resource["name"].(string)
	if !ok {
		return false
	}

	for _, pr := range policyResources {
		if pr == "*" || pr == resourceName {
			return true
		}
		if re, err := regexp.Compile(pr); err == nil && re.MatchString(resourceName) {
			return true
		}
	}

	return false
}

// evaluateConditions evaluates conditions
func (pe *PolicyEngine) evaluateConditions(resource map[string]interface{}, conditions []Condition) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, cond := range conditions {
		value, exists := resource[cond.Field]
		if !exists && cond.Operator != "notExists" {
			return false
		}

		if !pe.evaluateCondition(value, cond.Operator, cond.Value) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single condition
func (pe *PolicyEngine) evaluateCondition(value interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "equals":
		return value == expected
	case "notEquals":
		return value != expected
	case "exists":
		return value != nil
	case "notExists":
		return value == nil
	case "in":
		if arr, ok := expected.([]interface{}); ok {
			for _, v := range arr {
				if v == value {
					return true
				}
			}
		}
		return false
	case "notIn":
		if arr, ok := expected.([]interface{}); ok {
			for _, v := range arr {
				if v == value {
					return false
				}
			}
		}
		return true
	case "matches":
		if str, ok := value.(string); ok {
			if pattern, ok := expected.(string); ok {
				if re, err := regexp.Compile(pattern); err == nil {
					return re.MatchString(str)
				}
			}
		}
		return false
	case "gt":
		return toNum(value) > toNum(expected)
	case "lt":
		return toNum(value) < toNum(expected)
	case "gte":
		return toNum(value) >= toNum(expected)
	case "lte":
		return toNum(value) <= toNum(expected)
	default:
		return false
	}
}

// hasException checks if there's an exception
func (pe *PolicyEngine) hasException(policy *Policy, resource map[string]interface{}, user string) bool {
	for _, exc := range policy.Exceptions {
		// Check if expired
		if exc.ExpiresAt != nil && time.Now().After(*exc.ExpiresAt) {
			continue
		}

		// Check user
		userMatches := false
		for _, u := range exc.Users {
			if u == "*" || u == user {
				userMatches = true
				break
			}
		}

		if !userMatches {
			continue
		}

		// Check resource
		resourceName, ok := resource["name"].(string)
		if !ok {
			continue
		}

		for _, r := range exc.Resources {
			if r == "*" || r == resourceName {
				return true
			}
		}
	}

	return false
}

// InvalidateCache invalidates the evaluation cache
func (pe *PolicyEngine) InvalidateCache() {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	pe.evaluationCache = make(map[string]*EvaluationResult)
}

// GetPoliciesForKind returns policies for a resource kind
func (pe *PolicyEngine) GetPoliciesForKind(kind string) []*Policy {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	var policies []*Policy
	for _, policy := range pe.policies {
		if policy.Enabled && (policy.Kind == kind || policy.Kind == "*") {
			policies = append(policies, policy)
		}
	}

	return policies
}

func toNum(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}
