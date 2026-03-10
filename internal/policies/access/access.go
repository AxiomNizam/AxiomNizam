package access

import (
	"fmt"
	"time"
)

// AccessPolicy defines access control rules
type AccessPolicy struct {
	ID          string
	Name        string
	Type        string
	Version     string
	Enabled     bool
	Principal   Principal
	Resources   []Resource
	Actions     []string
	Effect      string // "allow" or "deny"
	Conditions  []AccessCondition
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Principal represents who is requesting access
type Principal struct {
	Type       string // "user", "group", "service", "role"
	ID         string
	Name       string
	Tags       map[string]string
	Attributes map[string]interface{}
}

// Resource represents what is being accessed
type Resource struct {
	Type      string // "database", "table", "api", "file"
	ID        string
	Name      string
	Namespace string
	Tags      map[string]string
	Sensitive bool
}

// AccessCondition defines conditions for access
type AccessCondition struct {
	Type     string // "time", "ip", "mfa", "location"
	Operator string // "equals", "notequals", "in", "notin", "matches"
	Value    string
	Values   []string
}

// GetID returns the policy ID
func (ap *AccessPolicy) GetID() string {
	return ap.ID
}

// GetName returns the policy name
func (ap *AccessPolicy) GetName() string {
	return ap.Name
}

// GetType returns the policy type
func (ap *AccessPolicy) GetType() string {
	return ap.Type
}

// GetVersion returns the version
func (ap *AccessPolicy) GetVersion() string {
	return ap.Version
}

// GetEnabled returns if enabled
func (ap *AccessPolicy) GetEnabled() bool {
	return ap.Enabled
}

// Validate validates the policy
func (ap *AccessPolicy) Validate() error {
	if ap.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}
	if ap.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}
	if ap.Principal.ID == "" {
		return fmt.Errorf("principal ID cannot be empty")
	}
	if len(ap.Resources) == 0 {
		return fmt.Errorf("at least one resource must be specified")
	}
	if ap.Effect != "allow" && ap.Effect != "deny" {
		return fmt.Errorf("effect must be 'allow' or 'deny'")
	}
	return nil
}

// CanAccess checks if a principal can perform an action on a resource
func (ap *AccessPolicy) CanAccess(principalID, principalType, action string, resourceID, resourceType string) bool {
	// Check principal match
	if ap.Principal.ID != principalID || ap.Principal.Type != principalType {
		return false
	}

	// Check action
	actionAllowed := false
	for _, a := range ap.Actions {
		if a == "*" || a == action {
			actionAllowed = true
			break
		}
	}
	if !actionAllowed {
		return false
	}

	// Check resource
	resourceAllowed := false
	for _, r := range ap.Resources {
		if (r.Type == "*" || r.Type == resourceType) &&
			(r.ID == "*" || r.ID == resourceID) {
			resourceAllowed = true
			break
		}
	}

	return resourceAllowed
}

// AccessPolicyBuilder builds access policies fluently
type AccessPolicyBuilder struct {
	policy *AccessPolicy
}

// NewAccessPolicyBuilder creates a new access policy builder
func NewAccessPolicyBuilder(id, name string) *AccessPolicyBuilder {
	return &AccessPolicyBuilder{
		policy: &AccessPolicy{
			ID:         id,
			Name:       name,
			Type:       "access",
			Version:    "1.0",
			Enabled:    true,
			Effect:     "allow",
			Resources:  make([]Resource, 0),
			Actions:    make([]string, 0),
			Conditions: make([]AccessCondition, 0),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}
}

// WithPrincipal sets the principal
func (apb *AccessPolicyBuilder) WithPrincipal(principalType, principalID, name string) *AccessPolicyBuilder {
	apb.policy.Principal = Principal{
		Type:       principalType,
		ID:         principalID,
		Name:       name,
		Tags:       make(map[string]string),
		Attributes: make(map[string]interface{}),
	}
	return apb
}

// AddResource adds a resource
func (apb *AccessPolicyBuilder) AddResource(resourceType, resourceID, resourceName string) *AccessPolicyBuilder {
	apb.policy.Resources = append(apb.policy.Resources, Resource{
		Type:      resourceType,
		ID:        resourceID,
		Name:      resourceName,
		Tags:      make(map[string]string),
		Sensitive: false,
	})
	return apb
}

// AddAction adds an action
func (apb *AccessPolicyBuilder) AddAction(action string) *AccessPolicyBuilder {
	apb.policy.Actions = append(apb.policy.Actions, action)
	return apb
}

// WithEffect sets the effect
func (apb *AccessPolicyBuilder) WithEffect(effect string) *AccessPolicyBuilder {
	apb.policy.Effect = effect
	return apb
}

// AddCondition adds a condition
func (apb *AccessPolicyBuilder) AddCondition(condition AccessCondition) *AccessPolicyBuilder {
	apb.policy.Conditions = append(apb.policy.Conditions, condition)
	return apb
}

// WithDescription sets the description
func (apb *AccessPolicyBuilder) WithDescription(description string) *AccessPolicyBuilder {
	apb.policy.Description = description
	return apb
}

// Build builds the policy
func (apb *AccessPolicyBuilder) Build() (*AccessPolicy, error) {
	if err := apb.policy.Validate(); err != nil {
		return nil, err
	}
	return apb.policy, nil
}

// AttributeBasedAccessControl (ABAC) implementation
type ABACEngine struct {
	policies []*AccessPolicy
}

// NewABACEngine creates a new ABAC engine
func NewABACEngine() *ABACEngine {
	return &ABACEngine{
		policies: make([]*AccessPolicy, 0),
	}
}

// AddPolicy adds a policy to the engine
func (ae *ABACEngine) AddPolicy(policy *AccessPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	ae.policies = append(ae.policies, policy)
	return nil
}

// Evaluate evaluates access based on ABAC policies
func (ae *ABACEngine) Evaluate(principal Principal, resource Resource, action string) bool {
	for _, policy := range ae.policies {
		if !policy.Enabled {
			continue
		}

		// Match principal
		if !ae.matchPrincipal(policy.Principal, principal) {
			continue
		}

		// Match action
		if !ae.matchAction(policy.Actions, action) {
			continue
		}

		// Match resource
		if !ae.matchResource(policy.Resources, resource) {
			continue
		}

		// Evaluate conditions
		if !ae.evaluateConditions(policy.Conditions) {
			continue
		}

		// Return based on effect
		return policy.Effect == "allow"
	}

	return false
}

func (ae *ABACEngine) matchPrincipal(policyPrincipal Principal, requestPrincipal Principal) bool {
	if policyPrincipal.Type != requestPrincipal.Type {
		return false
	}
	if policyPrincipal.ID != "*" && policyPrincipal.ID != requestPrincipal.ID {
		return false
	}
	return true
}

func (ae *ABACEngine) matchAction(allowedActions []string, requestedAction string) bool {
	for _, action := range allowedActions {
		if action == "*" || action == requestedAction {
			return true
		}
	}
	return false
}

func (ae *ABACEngine) matchResource(allowedResources []Resource, requestedResource Resource) bool {
	for _, resource := range allowedResources {
		if (resource.Type == "*" || resource.Type == requestedResource.Type) &&
			(resource.ID == "*" || resource.ID == requestedResource.ID) {
			return true
		}
	}
	return false
}

func (ae *ABACEngine) evaluateConditions(conditions []AccessCondition) bool {
	for _, condition := range conditions {
		if !ae.evaluateCondition(condition) {
			return false
		}
	}
	return true
}

func (ae *ABACEngine) evaluateCondition(condition AccessCondition) bool {
	// Simplified condition evaluation
	switch condition.Type {
	case "time":
		// In production, check time range against condition.Value
		return true
	case "ip":
		// In production, check IP whitelist against condition.Values
		if len(condition.Values) == 0 {
			return true
		}
		return true
	case "mfa":
		// In production, verify MFA status from context
		if condition.Value == "required" {
			return true
		}
		return true
	default:
		return true
	}
}

// RoleBasedAccessControl (RBAC) implementation
type RBACEngine struct {
	roleAssignments map[string][]string      // userID -> roles
	rolePolicies    map[string]*AccessPolicy // role -> policy
}

// NewRBACEngine creates a new RBAC engine
func NewRBACEngine() *RBACEngine {
	return &RBACEngine{
		roleAssignments: make(map[string][]string),
		rolePolicies:    make(map[string]*AccessPolicy),
	}
}

// AssignRole assigns a role to a user
func (re *RBACEngine) AssignRole(userID, role string) {
	if roles, exists := re.roleAssignments[userID]; exists {
		// Check if role already assigned
		for _, r := range roles {
			if r == role {
				return
			}
		}
		re.roleAssignments[userID] = append(roles, role)
	} else {
		re.roleAssignments[userID] = []string{role}
	}
}

// DefineRolePolicy defines a policy for a role
func (re *RBACEngine) DefineRolePolicy(role string, policy *AccessPolicy) {
	re.rolePolicies[role] = policy
}

// CanAccess checks if a user can perform an action
func (re *RBACEngine) CanAccess(userID, action, resourceType, resourceID string) bool {
	roles, exists := re.roleAssignments[userID]
	if !exists {
		return false
	}

	for _, role := range roles {
		policy, exists := re.rolePolicies[role]
		if !exists {
			continue
		}

		if policy.CanAccess(userID, "user", action, resourceID, resourceType) {
			return true
		}
	}

	return false
}

// PolicyBasedAccessControl (PBAC) implementation - most flexible
type PBACEngine struct {
	policies []*AccessPolicy
}

// NewPBACEngine creates a new PBAC engine
func NewPBACEngine() *PBACEngine {
	return &PBACEngine{
		policies: make([]*AccessPolicy, 0),
	}
}

// AddPolicy adds a policy
func (pe *PBACEngine) AddPolicy(policy *AccessPolicy) {
	pe.policies = append(pe.policies, policy)
}

// Evaluate evaluates access
func (pe *PBACEngine) Evaluate(principalID, principalType, action string, resourceID, resourceType string) bool {
	var allowPolicies []*AccessPolicy
	var denyPolicies []*AccessPolicy

	for _, policy := range pe.policies {
		if !policy.Enabled {
			continue
		}

		if policy.Principal.ID == principalID && policy.Principal.Type == principalType {
			if ae := matchPolicy(policy, action, resourceID, resourceType); ae {
				if policy.Effect == "allow" {
					allowPolicies = append(allowPolicies, policy)
				} else {
					denyPolicies = append(denyPolicies, policy)
				}
			}
		}
	}

	// Explicit deny takes precedence
	if len(denyPolicies) > 0 {
		return false
	}

	return len(allowPolicies) > 0
}

func matchPolicy(policy *AccessPolicy, action, resourceID, resourceType string) bool {
	// Check action
	actionMatches := false
	for _, a := range policy.Actions {
		if a == "*" || a == action {
			actionMatches = true
			break
		}
	}
	if !actionMatches {
		return false
	}

	// Check resource
	for _, r := range policy.Resources {
		if (r.Type == "*" || r.Type == resourceType) &&
			(r.ID == "*" || r.ID == resourceID) {
			return true
		}
	}

	return false
}
