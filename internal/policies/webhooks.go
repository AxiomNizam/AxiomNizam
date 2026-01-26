package policies

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AdmissionReview represents an admission request/response (Kubernetes-style webhooks)
type AdmissionReview struct {
	UID       string
	Kind      string
	Name      string
	Namespace string
	Operation string // CREATE, UPDATE, DELETE, CONNECT
	Object    map[string]interface{}
	OldObject map[string]interface{}
	UserInfo  UserInfo
	Timestamp time.Time
}

// AdmissionResponse represents the admission decision
type AdmissionResponse struct {
	Allowed   bool
	Reason    string
	Patch     []byte // JSON patch if modification needed
	PatchType string // JSONPatch
	Warnings  []string
}

// UserInfo contains information about the user making the request
type UserInfo struct {
	Username string
	UID      string
	Groups   []string
	Extras   map[string][]string
}

// AdmissionValidator validates resources before admission (validating webhook)
type AdmissionValidator struct {
	mu         sync.RWMutex
	validators map[string][]ValidatingHook
}

// ValidatingHook validates a resource
type ValidatingHook struct {
	Name      string
	Kind      string
	Operation string // CREATE, UPDATE, DELETE, or * for all
	Fn        func(ctx context.Context, review *AdmissionReview) error
}

// NewAdmissionValidator creates a new admission validator
func NewAdmissionValidator() *AdmissionValidator {
	return &AdmissionValidator{
		validators: make(map[string][]ValidatingHook),
	}
}

// RegisterValidator registers a validating webhook
func (av *AdmissionValidator) RegisterValidator(hook ValidatingHook) error {
	if hook.Name == "" || hook.Fn == nil {
		return fmt.Errorf("validator name and function required")
	}

	av.mu.Lock()
	defer av.mu.Unlock()

	key := fmt.Sprintf("%s/%s", hook.Kind, hook.Operation)
	av.validators[key] = append(av.validators[key], hook)

	return nil
}

// Validate runs all validators for a resource
func (av *AdmissionValidator) Validate(ctx context.Context, review *AdmissionReview) *AdmissionResponse {
	av.mu.RLock()
	key := fmt.Sprintf("%s/%s", review.Kind, review.Operation)
	validators := av.validators[key]
	wildcard := av.validators[fmt.Sprintf("%s/*", review.Kind)]
	av.mu.RUnlock()

	validators = append(validators, wildcard...)

	response := &AdmissionResponse{Allowed: true}

	for _, validator := range validators {
		if err := validator.Fn(ctx, review); err != nil {
			response.Allowed = false
			response.Reason = fmt.Sprintf("%s: %v", validator.Name, err)
			return response
		}
	}

	return response
}

// MutatingAdmission mutates resources before admission (mutating webhook)
type MutatingAdmission struct {
	mu       sync.RWMutex
	mutators map[string][]MutatingHook
}

// MutatingHook mutates a resource
type MutatingHook struct {
	Name      string
	Kind      string
	Operation string
	Fn        func(ctx context.Context, review *AdmissionReview) ([]byte, error)
}

// NewMutatingAdmission creates a new mutating admission
func NewMutatingAdmission() *MutatingAdmission {
	return &MutatingAdmission{
		mutators: make(map[string][]MutatingHook),
	}
}

// RegisterMutator registers a mutating webhook
func (ma *MutatingAdmission) RegisterMutator(hook MutatingHook) error {
	if hook.Name == "" || hook.Fn == nil {
		return fmt.Errorf("mutator name and function required")
	}

	ma.mu.Lock()
	defer ma.mu.Unlock()

	key := fmt.Sprintf("%s/%s", hook.Kind, hook.Operation)
	ma.mutators[key] = append(ma.mutators[key], hook)

	return nil
}

// Mutate runs all mutators for a resource
func (ma *MutatingAdmission) Mutate(ctx context.Context, review *AdmissionReview) *AdmissionResponse {
	ma.mu.RLock()
	key := fmt.Sprintf("%s/%s", review.Kind, review.Operation)
	mutators := ma.mutators[key]
	wildcard := ma.mutators[fmt.Sprintf("%s/*", review.Kind)]
	ma.mu.RUnlock()

	mutators = append(mutators, wildcard...)

	response := &AdmissionResponse{Allowed: true}

	for _, mutator := range mutators {
		patch, err := mutator.Fn(ctx, review)
		if err != nil {
			response.Allowed = false
			response.Reason = fmt.Sprintf("%s: %v", mutator.Name, err)
			return response
		}
		if patch != nil {
			response.Patch = patch
			response.PatchType = "JSONPatch"
		}
	}

	return response
}

// PolicyEngine implements CEL-style policy evaluation (Cloud Event Language)
type PolicyEngine struct {
	mu       sync.RWMutex
	policies map[string]*Policy
}

// Policy defines allow/deny rules
type Policy struct {
	Name        string
	Kind        string
	Effect      string // Allow or Deny
	Conditions  []PolicyCondition
	Actions     []string
	Priority    int
	Description string
}

// PolicyCondition represents a CEL expression condition
type PolicyCondition struct {
	Field    string
	Operator string // ==, !=, contains, in, matches
	Value    interface{}
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		policies: make(map[string]*Policy),
	}
}

// RegisterPolicy registers a policy
func (pe *PolicyEngine) RegisterPolicy(policy *Policy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name required")
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	pe.policies[policy.Name] = policy
	return nil
}

// EvaluatePolicy evaluates if a resource matches policy conditions
func (pe *PolicyEngine) EvaluatePolicy(policy *Policy, resource map[string]interface{}) bool {
	if len(policy.Conditions) == 0 {
		return true
	}

	// All conditions must match (AND)
	for _, condition := range policy.Conditions {
		if !pe.evaluateCondition(condition, resource) {
			return false
		}
	}
	return true
}

// evaluateCondition evaluates a single condition
func (pe *PolicyEngine) evaluateCondition(condition PolicyCondition, resource map[string]interface{}) bool {
	value, ok := resource[condition.Field]
	if !ok {
		return false
	}

	switch condition.Operator {
	case "==":
		return value == condition.Value
	case "!=":
		return value != condition.Value
	case "contains":
		str, ok := value.(string)
		if !ok {
			return false
		}
		pat, ok := condition.Value.(string)
		if !ok {
			return false
		}
		return contains(str, pat)
	case "in":
		list, ok := condition.Value.([]interface{})
		if !ok {
			return false
		}
		for _, item := range list {
			if item == value {
				return true
			}
		}
		return false
	case "matches": // regex
		// Simplified - would use regex in real implementation
		return value == condition.Value
	}
	return false
}

// AuditPolicy logs resource access (Kubernetes-style audit logs)
type AuditPolicy struct {
	Kind       string
	Verbs      []string // get, list, create, update, delete
	Omit       bool
	OmitStages []string // RequestReceived, ResponseStarted, ResponseComplete
}

// AuditLog represents an audit log entry
type AuditLog struct {
	Timestamp    time.Time
	Level        string // Metadata, RequestResponse, RequestReceivedEvent
	Verb         string
	Kind         string
	Name         string
	Namespace    string
	UserName     string
	UserGroups   []string
	Status       string
	StatusCode   int
	RequestSize  int
	ResponseSize int
}

// AuditLogger logs resource operations
type AuditLogger struct {
	mu     sync.RWMutex
	logs   []AuditLog
	maxLen int
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(maxLen int) *AuditLogger {
	return &AuditLogger{
		logs:   make([]AuditLog, 0),
		maxLen: maxLen,
	}
}

// LogOperation logs a resource operation
func (al *AuditLogger) LogOperation(log AuditLog) {
	al.mu.Lock()
	defer al.mu.Unlock()

	log.Timestamp = time.Now()
	al.logs = append(al.logs, log)

	// Limit log size
	if len(al.logs) > al.maxLen {
		al.logs = al.logs[len(al.logs)-al.maxLen:]
	}
}

// GetLogs returns audit logs
func (al *AuditLogger) GetLogs() []AuditLog {
	al.mu.RLock()
	defer al.mu.RUnlock()

	result := make([]AuditLog, len(al.logs))
	copy(result, al.logs)
	return result
}

// SearchLogs searches audit logs
func (al *AuditLogger) SearchLogs(filter func(AuditLog) bool) []AuditLog {
	al.mu.RLock()
	defer al.mu.RUnlock()

	result := make([]AuditLog, 0)
	for _, log := range al.logs {
		if filter(log) {
			result = append(result, log)
		}
	}
	return result
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
