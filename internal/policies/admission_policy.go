package policies

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// AdmissionPolicy enforces OPA-like policy rules for resource admission
type AdmissionPolicy struct {
	mu              sync.RWMutex
	policies        map[string]*PolicyDefinition
	rules           map[string][]*AdmissionRule
	denialReasons   map[string]string
	auditLog        []*AdmissionAudit
	maxAuditEntries int
}

// PolicyDefinition defines a complete admission policy
type PolicyDefinition struct {
	Name        string
	Description string
	Version     string
	Enabled     bool
	CreatedAt   time.Time
	Rules       []*AdmissionRule
	Scope       string // * for all, or specific kind
}

// AdmissionRule defines a single admission rule
type AdmissionRule struct {
	ID          string
	Name        string
	Description string
	Effect      string // Deny, Warn, Mutate
	Conditions  []*PolicyCondition
	Action      RuleFn
	Priority    int
}

// PolicyCondition checks a condition on a resource
type PolicyCondition struct {
	Path     string // JSONPath into resource
	Operator string // equals, notEquals, matches, contains, empty, exists, notExists, gt, lt
	Value    interface{}
}

// RuleFn is the action function for a rule
type RuleFn func(ctx context.Context, resource map[string]interface{}) (bool, string)

// AdmissionAudit tracks admission decisions
type AdmissionAudit struct {
	ID        string
	Timestamp time.Time
	Kind      string
	Name      string
	Namespace string
	Operation string
	Decision  string // Allowed, Denied, Warned, Mutated
	Reason    string
	PolicyID  string
	RuleID    string
}

// NewAdmissionPolicy creates a new admission policy engine
func NewAdmissionPolicy(maxAuditEntries int) *AdmissionPolicy {
	return &AdmissionPolicy{
		policies:        make(map[string]*PolicyDefinition),
		rules:           make(map[string][]*AdmissionRule),
		denialReasons:   make(map[string]string),
		auditLog:        make([]*AdmissionAudit, 0, maxAuditEntries),
		maxAuditEntries: maxAuditEntries,
	}
}

// RegisterPolicy registers a policy
func (ap *AdmissionPolicy) RegisterPolicy(ctx context.Context, policy *PolicyDefinition) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if _, exists := ap.policies[policy.Name]; exists {
		return fmt.Errorf("policy %s already exists", policy.Name)
	}

	policy.CreatedAt = time.Now()
	ap.policies[policy.Name] = policy

	// Index rules by kind
	key := policy.Scope
	if key == "" {
		key = "*"
	}
	ap.rules[key] = append(ap.rules[key], policy.Rules...)

	return nil
}

// DisablePolicy disables a policy
func (ap *AdmissionPolicy) DisablePolicy(policyName string) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	policy, exists := ap.policies[policyName]
	if !exists {
		return fmt.Errorf("policy %s not found", policyName)
	}

	policy.Enabled = false
	return nil
}

// EnablePolicy enables a policy
func (ap *AdmissionPolicy) EnablePolicy(policyName string) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	policy, exists := ap.policies[policyName]
	if !exists {
		return fmt.Errorf("policy %s not found", policyName)
	}

	policy.Enabled = true
	return nil
}

// AdmitResource checks if a resource can be admitted
func (ap *AdmissionPolicy) AdmitResource(ctx context.Context, kind, name, namespace, operation string, resource map[string]interface{}) (*AdmissionDecision, error) {
	ap.mu.RLock()

	// Get applicable rules
	wildCardRules := ap.rules["*"]
	kindRules := ap.rules[kind]

	var applicableRules []*AdmissionRule
	applicableRules = append(applicableRules, wildCardRules...)
	applicableRules = append(applicableRules, kindRules...)

	// Sort by priority
	sortRulesByPriority(applicableRules)

	ap.mu.RUnlock()

	decision := &AdmissionDecision{
		Allowed:      true,
		Warnings:     []string{},
		Mutations:    []Mutation{},
		MatchedRules: []*AdmissionRule{},
	}

	// Evaluate rules
	for _, rule := range applicableRules {
		policy := ap.getPolicyByRule(rule)
		if policy == nil || !policy.Enabled {
			continue
		}

		// Check if all conditions match
		if !ap.evaluateConditions(resource, rule.Conditions) {
			continue
		}

		// Execute rule action
		ok, reason := rule.Action(ctx, resource)

		if !ok {
			decision.MatchedRules = append(decision.MatchedRules, rule)

			switch rule.Effect {
			case "Deny":
				decision.Allowed = false
				decision.Reason = reason
				decision.DeniedBy = rule.ID

				// Record audit
				ap.recordAudit(&AdmissionAudit{
					ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
					Timestamp: time.Now(),
					Kind:      kind,
					Name:      name,
					Namespace: namespace,
					Operation: operation,
					Decision:  "Denied",
					Reason:    reason,
					PolicyID:  policy.Name,
					RuleID:    rule.ID,
				})

				return decision, nil

			case "Warn":
				decision.Warnings = append(decision.Warnings, fmt.Sprintf("[%s] %s", rule.Name, reason))

			case "Mutate":
				// Execute mutation action if provided
				decision.Mutations = append(decision.Mutations, Mutation{
					RuleID: rule.ID,
					Action: reason,
				})
			}
		}
	}

	// Record audit
	auditDecision := "Allowed"
	if len(decision.Warnings) > 0 {
		auditDecision = "Warned"
	}
	if len(decision.Mutations) > 0 {
		auditDecision = "Mutated"
	}

	ap.recordAudit(&AdmissionAudit{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
		Operation: operation,
		Decision:  auditDecision,
		Reason:    "Passed admission policies",
	})

	return decision, nil
}

// evaluateConditions checks if all conditions match
func (ap *AdmissionPolicy) evaluateConditions(resource map[string]interface{}, conditions []*PolicyCondition) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, cond := range conditions {
		value := getValueAtPath(resource, cond.Path)
		if !evaluateCondition(value, cond.Operator, cond.Value) {
			return false
		}
	}

	return true
}

// getPolicyByRule finds the policy that owns a rule
func (ap *AdmissionPolicy) getPolicyByRule(rule *AdmissionRule) *PolicyDefinition {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	for _, policy := range ap.policies {
		for _, r := range policy.Rules {
			if r.ID == rule.ID {
				return policy
			}
		}
	}

	return nil
}

// recordAudit records an admission audit entry
func (ap *AdmissionPolicy) recordAudit(audit *AdmissionAudit) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	ap.auditLog = append(ap.auditLog, audit)

	// Keep only maxAuditEntries
	if len(ap.auditLog) > ap.maxAuditEntries {
		ap.auditLog = ap.auditLog[len(ap.auditLog)-ap.maxAuditEntries:]
	}
}

// GetAuditLog returns audit log entries
func (ap *AdmissionPolicy) GetAuditLog(ctx context.Context, filters map[string]interface{}) []*AdmissionAudit {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	var results []*AdmissionAudit
	for _, entry := range ap.auditLog {
		if matchesAuditFilters(entry, filters) {
			results = append(results, entry)
		}
	}

	return results
}

// AdmissionDecision represents the outcome of admission evaluation
type AdmissionDecision struct {
	Allowed      bool
	Reason       string
	DeniedBy     string
	Warnings     []string
	Mutations    []Mutation
	MatchedRules []*AdmissionRule
}

// Mutation represents a resource mutation
type Mutation struct {
	RuleID string
	Action string
}

// CommonPolicies provides standard admission policies
type CommonPolicies struct{}

// NewCommonPolicies creates common policies
func NewCommonPolicies() *CommonPolicies {
	return &CommonPolicies{}
}

// CreatePIIProtectionPolicy creates a policy that blocks PII exposure
func (cp *CommonPolicies) CreatePIIProtectionPolicy() *PolicyDefinition {
	return &PolicyDefinition{
		Name:        "pii-protection",
		Description: "Prevents exposure of Personally Identifiable Information",
		Version:     "1.0",
		Enabled:     true,
		Scope:       "*",
		Rules: []*AdmissionRule{
			{
				ID:          "no-pii-in-logs",
				Name:        "No PII in Logs",
				Description: "Logs must not contain PII (SSN, credit cards, etc.)",
				Effect:      "Deny",
				Priority:    10,
				Conditions: []*PolicyCondition{
					{
						Path:     "spec.config",
						Operator: "matches",
						Value:    "(?:\\d{3}-\\d{2}-\\d{4}|\\d{16})", // SSN or credit card patterns
					},
				},
				Action: func(ctx context.Context, resource map[string]interface{}) (bool, string) {
					return false, "Resource contains PII patterns (SSN, credit card numbers)"
				},
			},
			{
				ID:          "no-pii-in-emails",
				Name:        "No PII in Emails",
				Description: "Email fields must not contain personal information",
				Effect:      "Warn",
				Priority:    5,
				Conditions: []*PolicyCondition{
					{
						Path:     "spec.notifications.email",
						Operator: "exists",
					},
				},
				Action: func(ctx context.Context, resource map[string]interface{}) (bool, string) {
					return true, "Review email configuration for PII"
				},
			},
		},
	}
}

// CreateEncryptionPolicy creates a policy that enforces encryption
func (cp *CommonPolicies) CreateEncryptionPolicy() *PolicyDefinition {
	return &PolicyDefinition{
		Name:        "encryption-required",
		Description: "Enforces encryption at rest and in transit",
		Version:     "1.0",
		Enabled:     true,
		Scope:       "*",
		Rules: []*AdmissionRule{
			{
				ID:          "encryption-at-rest",
				Name:        "Encryption at Rest",
				Description: "All data must be encrypted at rest",
				Effect:      "Deny",
				Priority:    10,
				Conditions: []*PolicyCondition{
					{
						Path:     "spec.encryption.atRest.enabled",
						Operator: "equals",
						Value:    false,
					},
				},
				Action: func(ctx context.Context, resource map[string]interface{}) (bool, string) {
					return false, "Encryption at rest must be enabled"
				},
			},
			{
				ID:          "encryption-in-transit",
				Name:        "Encryption in Transit",
				Description: "All data in transit must use TLS",
				Effect:      "Deny",
				Priority:    10,
				Conditions: []*PolicyCondition{
					{
						Path:     "spec.tls.enabled",
						Operator: "equals",
						Value:    false,
					},
				},
				Action: func(ctx context.Context, resource map[string]interface{}) (bool, string) {
					return false, "TLS must be enabled for data in transit"
				},
			},
		},
	}
}

// CreateOwnershipPolicy creates a policy that enforces ownership
func (cp *CommonPolicies) CreateOwnershipPolicy() *PolicyDefinition {
	return &PolicyDefinition{
		Name:        "ownership-required",
		Description: "All resources must have defined owner and team",
		Version:     "1.0",
		Enabled:     true,
		Scope:       "*",
		Rules: []*AdmissionRule{
			{
				ID:          "owner-label",
				Name:        "Owner Label",
				Description: "Resources must have owner label",
				Effect:      "Deny",
				Priority:    10,
				Conditions: []*PolicyCondition{
					{
						Path:     "metadata.labels.owner",
						Operator: "notExists",
					},
				},
				Action: func(ctx context.Context, resource map[string]interface{}) (bool, string) {
					return false, "Resources must have 'owner' label"
				},
			},
			{
				ID:          "team-label",
				Name:        "Team Label",
				Description: "Resources must have team label",
				Effect:      "Deny",
				Priority:    10,
				Conditions: []*PolicyCondition{
					{
						Path:     "metadata.labels.team",
						Operator: "notExists",
					},
				},
				Action: func(ctx context.Context, resource map[string]interface{}) (bool, string) {
					return false, "Resources must have 'team' label"
				},
			},
		},
	}
}

// CreateNetworkPolicyDefaults creates a policy enforcing network security
func (cp *CommonPolicies) CreateNetworkPolicyDefaults() *PolicyDefinition {
	return &PolicyDefinition{
		Name:        "network-security",
		Description: "Enforces network security defaults",
		Version:     "1.0",
		Enabled:     true,
		Scope:       "*",
		Rules: []*AdmissionRule{
			{
				ID:          "no-public-endpoints",
				Name:        "No Public Endpoints",
				Description: "Production resources must not be publicly exposed",
				Effect:      "Deny",
				Priority:    10,
				Conditions: []*PolicyCondition{
					{
						Path:     "metadata.labels.environment",
						Operator: "equals",
						Value:    "production",
					},
					{
						Path:     "spec.public",
						Operator: "equals",
						Value:    true,
					},
				},
				Action: func(ctx context.Context, resource map[string]interface{}) (bool, string) {
					return false, "Production resources cannot be publicly exposed"
				},
			},
		},
	}
}

// CreateResourceLimitPolicy creates a policy enforcing resource limits
func (cp *CommonPolicies) CreateResourceLimitPolicy() *PolicyDefinition {
	return &PolicyDefinition{
		Name:        "resource-limits",
		Description: "Enforces resource limit constraints",
		Version:     "1.0",
		Enabled:     true,
		Scope:       "*",
		Rules: []*AdmissionRule{
			{
				ID:          "max-cpu-limit",
				Name:        "Max CPU Limit",
				Description: "Single resource CPU cannot exceed 16 cores",
				Effect:      "Deny",
				Priority:    5,
				Conditions: []*PolicyCondition{
					{
						Path:     "spec.resources.limits.cpu",
						Operator: "gt",
						Value:    16,
					},
				},
				Action: func(ctx context.Context, resource map[string]interface{}) (bool, string) {
					return false, "CPU limit cannot exceed 16 cores"
				},
			},
			{
				ID:          "max-memory-limit",
				Name:        "Max Memory Limit",
				Description: "Single resource memory cannot exceed 1TB",
				Effect:      "Deny",
				Priority:    5,
				Conditions: []*PolicyCondition{
					{
						Path:     "spec.resources.limits.memory",
						Operator: "gt",
						Value:    1099511627776, // 1TB in bytes
					},
				},
				Action: func(ctx context.Context, resource map[string]interface{}) (bool, string) {
					return false, "Memory limit cannot exceed 1TB"
				},
			},
		},
	}
}

// Helper functions

// getValueAtPath retrieves a value at JSONPath
func getValueAtPath(obj map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(obj)

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

// evaluateCondition evaluates a single condition
func evaluateCondition(value interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "equals":
		return value == expected
	case "notEquals":
		return value != expected
	case "empty":
		return value == nil || value == ""
	case "exists":
		return value != nil
	case "notExists":
		return value == nil
	case "matches":
		if str, ok := value.(string); ok {
			if pattern, ok := expected.(string); ok {
				if re, err := regexp.Compile(pattern); err == nil {
					return re.MatchString(str)
				}
			}
		}
		return false
	case "contains":
		if str, ok := value.(string); ok {
			if substr, ok := expected.(string); ok {
				return strings.Contains(str, substr)
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

// sortRulesByPriority sorts rules by priority (highest first)
func sortRulesByPriority(rules []*AdmissionRule) {
	for i := 0; i < len(rules); i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[j].Priority > rules[i].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
}

// matchesAuditFilters checks if audit entry matches filters
func matchesAuditFilters(entry *AdmissionAudit, filters map[string]interface{}) bool {
	for key, expectedValue := range filters {
		var objValue interface{}

		switch key {
		case "decision":
			objValue = entry.Decision
		case "kind":
			objValue = entry.Kind
		case "policy_id":
			objValue = entry.PolicyID
		case "operation":
			objValue = entry.Operation
		}

		if objValue != expectedValue {
			return false
		}
	}

	return true
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
