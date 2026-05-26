package admission

import (
	"fmt"
	"log"
	"time"
)

// AdmissionPolicy defines rules for admitting requests
type AdmissionPolicy struct {
	ID                string
	Name              string
	Type              string
	Version           string
	Enabled           bool
	Operations        []string // "create", "update", "delete"
	ResourceTypes     []string
	Rules             []AdmissionRule
	MutatingRules     []MutatingRule
	ValidatingRules   []ValidatingRule
	FailurePolicy     string // "Fail", "Ignore"
	TimeoutSeconds    int32
	SideEffects       string // "None", "Some", "NoneOnDryRun", "Unknown"
	NamespaceSelector map[string]string
	Description       string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// AdmissionRule is a base admission rule
type AdmissionRule struct {
	Name        string
	Description string
	Rule        string
	Enabled     bool
	Effect      string // "allow" or "deny"
	Message     string
}

// MutatingRule modifies incoming requests
type MutatingRule struct {
	ID          string
	Name        string
	Description string
	Patch       []PatchOperation
	Enabled     bool
	Operations  []string
	Resources   []string
	Priority    int
}

// PatchOperation represents a JSON patch operation
type PatchOperation struct {
	Op    string      `json:"op"` // "add", "remove", "replace", "move", "copy", "test"
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
	From  string      `json:"from,omitempty"`
}

// ValidatingRule validates incoming requests
type ValidatingRule struct {
	ID                   string
	Name                 string
	Description          string
	ValidationExpression string
	Message              string
	Enabled              bool
	FailurePolicy        string
}

// AdmissionRequest represents a request to be admitted
type AdmissionRequest struct {
	UID       string
	Kind      string
	Namespace string
	Name      string
	Operation string // "CREATE", "UPDATE", "DELETE", "CONNECT"
	UserInfo  UserInfo
	Object    map[string]interface{}
	OldObject map[string]interface{}
	Options   map[string]interface{}
	DryRun    bool
}

// UserInfo contains user information
type UserInfo struct {
	Username string
	UID      string
	Groups   []string
	Extra    map[string]interface{}
}

// AdmissionResponse represents the response to an admission request
type AdmissionResponse struct {
	UID              string
	Allowed          bool
	Status           *Status
	Patch            []byte
	PatchType        string
	AuditAnnotations map[string]string
	Warnings         []string
}

// Status holds admission status information
type Status struct {
	Code    int32
	Message string
	Reason  string
	Details *StatusDetails
}

// StatusDetails holds additional status details
type StatusDetails struct {
	Name      string
	Group     string
	Kind      string
	UID       string
	Retryable bool
}

// GetID returns policy ID
func (ap *AdmissionPolicy) GetID() string {
	return ap.ID
}

// GetName returns policy name
func (ap *AdmissionPolicy) GetName() string {
	return ap.Name
}

// GetType returns policy type
func (ap *AdmissionPolicy) GetType() string {
	return ap.Type
}

// GetVersion returns version
func (ap *AdmissionPolicy) GetVersion() string {
	return ap.Version
}

// GetEnabled returns if enabled
func (ap *AdmissionPolicy) GetEnabled() bool {
	return ap.Enabled
}

// Validate validates the policy
func (ap *AdmissionPolicy) Validate() error {
	if ap.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}
	if ap.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}
	if len(ap.Operations) == 0 && len(ap.MutatingRules) == 0 && len(ap.ValidatingRules) == 0 {
		return fmt.Errorf("at least one operation or rule must be specified")
	}
	return nil
}

// AdmissionController handles admission control
type AdmissionController struct {
	policies           []*AdmissionPolicy
	mutatingWebhooks   []AdmissionWebhook
	validatingWebhooks []AdmissionWebhook
}

// AdmissionWebhook represents a webhook for admission control
type AdmissionWebhook struct {
	ID      string
	Name    string
	URL     string
	Timeout int32
	Rules   []WebhookRule
	Enabled bool
}

// WebhookRule defines when a webhook should be called
type WebhookRule struct {
	Operations  []string
	APIGroups   []string
	APIVersions []string
	Resources   []string
	Scope       string // "*", "Namespaced", "Cluster"
}

// NewAdmissionController creates a new admission controller
func NewAdmissionController() *AdmissionController {
	return &AdmissionController{
		policies:           make([]*AdmissionPolicy, 0),
		mutatingWebhooks:   make([]AdmissionWebhook, 0),
		validatingWebhooks: make([]AdmissionWebhook, 0),
	}
}

// AddPolicy adds an admission policy
func (ac *AdmissionController) AddPolicy(policy *AdmissionPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	ac.policies = append(ac.policies, policy)
	return nil
}

// Admit decides whether to admit a request
func (ac *AdmissionController) Admit(request AdmissionRequest) AdmissionResponse {
	response := AdmissionResponse{
		UID:              request.UID,
		Allowed:          true,
		AuditAnnotations: make(map[string]string),
		Warnings:         make([]string, 0),
	}

	// Apply validating rules
	for _, policy := range ac.policies {
		if !policy.Enabled {
			continue
		}

		// Check if policy applies to this request
		if !ac.policyAppliesToRequest(policy, request) {
			continue
		}

		// Apply validating rules
		for _, rule := range policy.ValidatingRules {
			if !rule.Enabled {
				continue
			}

			// In production, evaluate the CEL expression
			// For now, assume all rules pass
			if rule.Message != "" && rule.FailurePolicy == "Fail" {
				response.Allowed = false
				response.Status = &Status{
					Code:    400,
					Message: rule.Message,
					Reason:  "ValidationFailed",
				}
				return response
			}
		}

		// Apply mutating rules
		for _, rule := range policy.MutatingRules {
			if !rule.Enabled {
				continue
			}
			// Apply patches to object
			if patchErr := ac.applyPatches(request.Object, rule.Patch); patchErr != nil {
				log.Printf("admission: patch application failed (rule=%s): %v", rule.Name, patchErr)
			}
		}
	}

	return response
}

func (ac *AdmissionController) policyAppliesToRequest(policy *AdmissionPolicy, request AdmissionRequest) bool {
	// Check operation
	operationMatches := false
	for _, op := range policy.Operations {
		if op == "*" || op == request.Operation {
			operationMatches = true
			break
		}
	}

	// Check resource type
	resourceMatches := false
	for _, rt := range policy.ResourceTypes {
		if rt == "*" || rt == request.Kind {
			resourceMatches = true
			break
		}
	}

	// Check namespace selector
	namespaceMatches := len(policy.NamespaceSelector) == 0 // Empty selector matches all namespaces
	for key, value := range policy.NamespaceSelector {
		// In production, actually match namespace labels
		_ = key
		_ = value
		namespaceMatches = true
	}

	return operationMatches && resourceMatches && namespaceMatches
}

func (ac *AdmissionController) applyPatches(object map[string]interface{}, patches []PatchOperation) error {
	// Simplified patch application
	for _, patch := range patches {
		if patch.Op == "add" {
			// Apply add operation
		} else if patch.Op == "remove" {
			// Apply remove operation
		} else if patch.Op == "replace" {
			// Apply replace operation
		}
	}
	return nil
}

// ValidatingAdmissionPolicy is a simple validating policy
type ValidatingAdmissionPolicy struct {
	ID       string
	Name     string
	Rules    []ValidationRule
	Enabled  bool
	Severity string // "audit", "warn", "deny"
}

// ValidationRule defines a validation rule
type ValidationRule struct {
	Expression string // CEL expression
	Message    string
}

// Validate validates a request
func (vap *ValidatingAdmissionPolicy) Validate(object map[string]interface{}) (bool, string) {
	if !vap.Enabled {
		return true, ""
	}

	for _, rule := range vap.Rules {
		// In production, evaluate CEL expression
		// For now, return success
		_ = rule.Expression
		_ = object
	}

	return true, ""
}

// MutatingAdmissionPolicy is a simple mutating policy
type MutatingAdmissionPolicy struct {
	ID       string
	Name     string
	Rules    []MutatingRule
	Enabled  bool
	Priority int
}

// Mutate mutates a request
func (mp *MutatingAdmissionPolicy) Mutate(object map[string]interface{}) (map[string]interface{}, error) {
	if !mp.Enabled {
		return object, nil
	}

	for _, rule := range mp.Rules {
		if !rule.Enabled {
			continue
		}

		for _, patch := range rule.Patch {
			if patch.Op == "add" {
				// Add field
				setNestedField(object, patch.Path, patch.Value)
			}
		}
	}

	return object, nil
}

// setNestedField sets a value in a nested map using dot notation path
func setNestedField(m map[string]interface{}, path string, value interface{}) {
	keys := parseJsonPath(path)
	if len(keys) == 0 {
		return
	}

	current := m
	for i, key := range keys[:len(keys)-1] {
		if next, ok := current[key]; ok {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return
			}
		} else {
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}

		_ = i
	}

	current[keys[len(keys)-1]] = value
}

// parseJsonPath parses a JSON path like "/metadata/labels/app"
func parseJsonPath(path string) []string {
	// Simple implementation - split by /
	if path == "" {
		return []string{}
	}

	parts := []string{}
	current := ""

	for _, char := range path {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
