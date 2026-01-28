package controllers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// RBACEngine provides Kubernetes-style RBAC (Role-Based Access Control)
// Supports resource-level permissions with verb-based actions
type RBACEngine struct {
	mu             sync.RWMutex
	roles          map[string]*Role
	roleBindings   map[string][]*RoleBinding
	clusterRoles   map[string]*ClusterRole
	clusterRbacs   map[string][]*ClusterRoleBinding
	auditLog       []*RBACAuditLog
	maxAuditLog    int
	apiGroups      map[string]bool // Track allowed API groups
}

// Role defines permissions for a namespace scope
type Role struct {
	Name        string
	Namespace   string
	Rules       []*PolicyRule
	CreatedAt   time.Time
	Labels      map[string]string
	Annotations map[string]string
}

// RoleBinding binds a Role to subjects (users/groups/serviceaccounts)
type RoleBinding struct {
	Name      string
	Namespace string
	Role      string
	Subjects  []*Subject
	Labels    map[string]string
}

// ClusterRole defines cluster-wide permissions
type ClusterRole struct {
	Name        string
	Rules       []*PolicyRule
	CreatedAt   time.Time
	Labels      map[string]string
	Annotations map[string]string
}

// ClusterRoleBinding binds a ClusterRole to subjects cluster-wide
type ClusterRoleBinding struct {
	Name      string
	Role      string
	Subjects  []*Subject
	Labels    map[string]string
}

// PolicyRule defines what verbs can be performed on which resources
type PolicyRule struct {
	Verbs             []string          // create, read, update, delete, list, watch, patch
	APIGroups         []string          // e.g., "", "batch", "apps"
	Resources         []string          // e.g., "pods", "deployments", "jobs"
	ResourceNames     []string          // specific resource names (optional)
	NonResourceURLs   []string          // for non-resource endpoints
	Conditions        []*RuleCondition  // additional conditions
}

// RuleCondition adds fine-grained control to policy rules
type RuleCondition struct {
	Type  string      // "OwnershipRequired", "LabelSelector", "TimeWindow", "IPRestriction"
	Value interface{} // varies by condition type
}

// Subject represents a user, group, or service account
type Subject struct {
	Type  string // User, Group, ServiceAccount
	Name  string
	Namespace string // only for ServiceAccount
}

// RBACAuditLog tracks RBAC decisions for compliance
type RBACAuditLog struct {
	ID          string
	Timestamp   time.Time
	UserID      string
	Subject     *Subject
	Action      string // create, read, update, delete, list, watch, patch
	Resource    string
	APIGroup    string
	Namespace   string
	Allowed     bool
	Reason      string
	MatchedRule *PolicyRule
}

// RBACDecision represents the outcome of an RBAC check
type RBACDecision struct {
	Allowed   bool
	Reason    string
	MatchedRules []*PolicyRule
	DecisionTime time.Time
}

// NewRBACEngine creates a new RBAC engine
func NewRBACEngine() *RBACEngine {
	return &RBACEngine{
		roles:          make(map[string]*Role),
		roleBindings:   make(map[string][]*RoleBinding),
		clusterRoles:   make(map[string]*ClusterRole),
		clusterRbacs:   make(map[string][]*ClusterRoleBinding),
		auditLog:       make([]*RBACAuditLog, 0, 10000),
		maxAuditLog:    10000,
		apiGroups:      make(map[string]bool),
	}
}

// CreateRole creates a new namespaced role
func (re *RBACEngine) CreateRole(ctx context.Context, role *Role) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	if role.Name == "" || role.Namespace == "" {
		return fmt.Errorf("role name and namespace required")
	}

	key := fmt.Sprintf("%s/%s", role.Namespace, role.Name)
	if _, exists := re.roles[key]; exists {
		return fmt.Errorf("role already exists")
	}

	role.CreatedAt = time.Now()
	re.roles[key] = role
	return nil
}

// CreateClusterRole creates a cluster-wide role
func (re *RBACEngine) CreateClusterRole(ctx context.Context, role *ClusterRole) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	if role.Name == "" {
		return fmt.Errorf("cluster role name required")
	}

	if _, exists := re.clusterRoles[role.Name]; exists {
		return fmt.Errorf("cluster role already exists")
	}

	role.CreatedAt = time.Now()
	re.clusterRoles[role.Name] = role
	return nil
}

// CreateRoleBinding binds a role to subjects
func (re *RBACEngine) CreateRoleBinding(ctx context.Context, binding *RoleBinding) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	if binding.Name == "" || binding.Namespace == "" || binding.Role == "" {
		return fmt.Errorf("binding name, namespace, and role required")
	}

	key := fmt.Sprintf("%s/%s", binding.Namespace, binding.Name)
	re.roleBindings[key] = append(re.roleBindings[key], binding)
	return nil
}

// CreateClusterRoleBinding binds a cluster role to subjects
func (re *RBACEngine) CreateClusterRoleBinding(ctx context.Context, binding *ClusterRoleBinding) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	if binding.Name == "" || binding.Role == "" {
		return fmt.Errorf("binding name and role required")
	}

	re.clusterRbacs[binding.Name] = append(re.clusterRbacs[binding.Name], binding)
	return nil
}

// CanPerform checks if a subject can perform an action on a resource
func (re *RBACEngine) CanPerform(ctx context.Context, userID string, resourceKind string, verb string, namespace string) (bool, string) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	subject := &Subject{Type: "User", Name: userID}
	decision := &RBACDecision{
		Allowed:   false,
		Reason:    "no matching rules",
		MatchedRules: make([]*PolicyRule, 0),
		DecisionTime: time.Now(),
	}

	// Check cluster-wide roles first
	allowed, rules := re.checkClusterRoles(subject, verb, resourceKind)
	if allowed {
		decision.Allowed = true
		decision.MatchedRules = append(decision.MatchedRules, rules...)
	}

	// Check namespaced roles
	if !allowed && namespace != "" {
		allowed, rules := re.checkNamespacedRoles(subject, verb, resourceKind, namespace)
		if allowed {
			decision.Allowed = true
			decision.MatchedRules = append(decision.MatchedRules, rules...)
		}
	}

	if !decision.Allowed {
		decision.Reason = fmt.Sprintf("no permissions for %s on %s", verb, resourceKind)
	} else {
		decision.Reason = "allowed by matching rules"
	}

	// Record audit log
	re.recordAuditLog(&RBACAuditLog{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		UserID:    userID,
		Subject:   subject,
		Action:    verb,
		Resource:  resourceKind,
		Namespace: namespace,
		Allowed:   decision.Allowed,
		Reason:    decision.Reason,
	})

	return decision.Allowed, decision.Reason
}

// checkClusterRoles checks if subject has permission via cluster roles
func (re *RBACEngine) checkClusterRoles(subject *Subject, verb string, resourceKind string) (bool, []*PolicyRule) {
	matchedRules := make([]*PolicyRule, 0)

	// Find bindings for this subject
	for _, bindings := range re.clusterRbacs {
		for _, binding := range bindings {
			if re.subjectMatches(subject, binding.Subjects) {
				// Get the cluster role
				if role, exists := re.clusterRoles[binding.Role]; exists {
					// Check if role has the required permission
					if rules := re.checkRuleMatch(role.Rules, verb, resourceKind); len(rules) > 0 {
						matchedRules = append(matchedRules, rules...)
						return true, matchedRules
					}
				}
			}
		}
	}

	return len(matchedRules) > 0, matchedRules
}

// checkNamespacedRoles checks if subject has permission via namespaced roles
func (re *RBACEngine) checkNamespacedRoles(subject *Subject, verb string, resourceKind string, namespace string) (bool, []*PolicyRule) {
	matchedRules := make([]*PolicyRule, 0)

	// Find bindings in the namespace
	for key, bindings := range re.roleBindings {
		parts := strings.Split(key, "/")
		if len(parts) == 2 && parts[0] != namespace {
			continue // Different namespace
		}

		for _, binding := range bindings {
			if re.subjectMatches(subject, binding.Subjects) {
				// Get the role
				roleKey := fmt.Sprintf("%s/%s", binding.Namespace, binding.Role)
				if role, exists := re.roles[roleKey]; exists {
					// Check if role has the required permission
					if rules := re.checkRuleMatch(role.Rules, verb, resourceKind); len(rules) > 0 {
						matchedRules = append(matchedRules, rules...)
						return true, matchedRules
					}
				}
			}
		}
	}

	return len(matchedRules) > 0, matchedRules
}

// checkRuleMatch checks if any rule matches the verb and resource
func (re *RBACEngine) checkRuleMatch(rules []*PolicyRule, verb string, resourceKind string) []*PolicyRule {
	matched := make([]*PolicyRule, 0)

	for _, rule := range rules {
		if re.verbMatches(rule.Verbs, verb) && re.resourceMatches(rule.Resources, resourceKind) {
			matched = append(matched, rule)
		}
	}

	return matched
}

// verbMatches checks if verb is in the allowed verbs (supports wildcards)
func (re *RBACEngine) verbMatches(verbs []string, verb string) bool {
	for _, v := range verbs {
		if v == "*" || v == verb {
			return true
		}
	}
	return false
}

// resourceMatches checks if resource is in the allowed resources (supports wildcards)
func (re *RBACEngine) resourceMatches(resources []string, resource string) bool {
	for _, r := range resources {
		if r == "*" || r == resource {
			return true
		}
	}
	return false
}

// subjectMatches checks if subject matches any in the list
func (re *RBACEngine) subjectMatches(subject *Subject, subjects []*Subject) bool {
	for _, s := range subjects {
		if s.Type == subject.Type && s.Name == subject.Name {
			if s.Type == "ServiceAccount" {
				if s.Namespace == "" || s.Namespace == subject.Namespace {
					return true
				}
			} else {
				return true
			}
		}
	}
	return false
}

// recordAuditLog records RBAC decision
func (re *RBACEngine) recordAuditLog(audit *RBACAuditLog) {
	re.auditLog = append(re.auditLog, audit)
	if len(re.auditLog) > re.maxAuditLog {
		re.auditLog = re.auditLog[len(re.auditLog)-re.maxAuditLog:]
	}
}

// GetAuditLog returns RBAC audit log with filtering
func (re *RBACEngine) GetAuditLog(ctx context.Context, userID string, allowed *bool, limit int) []*RBACAuditLog {
	re.mu.RLock()
	defer re.mu.RUnlock()

	result := make([]*RBACAuditLog, 0)
	count := 0

	for i := len(re.auditLog) - 1; i >= 0 && count < limit; i-- {
		audit := re.auditLog[i]
		if (userID == "" || audit.UserID == userID) &&
			(allowed == nil || audit.Allowed == *allowed) {
			result = append(result, audit)
			count++
		}
	}

	return result
}

// ListRoles returns all roles in a namespace
func (re *RBACEngine) ListRoles(ctx context.Context, namespace string) []*Role {
	re.mu.RLock()
	defer re.mu.RUnlock()

	result := make([]*Role, 0)
	prefix := fmt.Sprintf("%s/", namespace)

	for key, role := range re.roles {
		if strings.HasPrefix(key, prefix) {
			result = append(result, role)
		}
	}

	return result
}

// ListClusterRoles returns all cluster roles
func (re *RBACEngine) ListClusterRoles(ctx context.Context) []*ClusterRole {
	re.mu.RLock()
	defer re.mu.RUnlock()

	result := make([]*ClusterRole, 0)
	for _, role := range re.clusterRoles {
		result = append(result, role)
	}

	return result
}

// GetRole retrieves a specific role
func (re *RBACEngine) GetRole(ctx context.Context, name string, namespace string) (*Role, error) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	key := fmt.Sprintf("%s/%s", namespace, name)
	if role, exists := re.roles[key]; exists {
		return role, nil
	}

	return nil, fmt.Errorf("role not found")
}

// GetClusterRole retrieves a specific cluster role
func (re *RBACEngine) GetClusterRole(ctx context.Context, name string) (*ClusterRole, error) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	if role, exists := re.clusterRoles[name]; exists {
		return role, nil
	}

	return nil, fmt.Errorf("cluster role not found")
}

// DeleteRole deletes a role
func (re *RBACEngine) DeleteRole(ctx context.Context, name string, namespace string) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	key := fmt.Sprintf("%s/%s", namespace, name)
	if _, exists := re.roles[key]; !exists {
		return fmt.Errorf("role not found")
	}

	delete(re.roles, key)

	// Clean up associated bindings
	for k, bindings := range re.roleBindings {
		filtered := make([]*RoleBinding, 0)
		for _, b := range bindings {
			if b.Role != name {
				filtered = append(filtered, b)
			}
		}
		if len(filtered) == 0 {
			delete(re.roleBindings, k)
		} else {
			re.roleBindings[k] = filtered
		}
	}

	return nil
}

// DeleteClusterRole deletes a cluster role
func (re *RBACEngine) DeleteClusterRole(ctx context.Context, name string) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	if _, exists := re.clusterRoles[name]; !exists {
		return fmt.Errorf("cluster role not found")
	}

	delete(re.clusterRoles, name)

	// Clean up associated bindings
	for k, bindings := range re.clusterRbacs {
		filtered := make([]*ClusterRoleBinding, 0)
		for _, b := range bindings {
			if b.Role != name {
				filtered = append(filtered, b)
			}
		}
		if len(filtered) == 0 {
			delete(re.clusterRbacs, k)
		} else {
			re.clusterRbacs[k] = filtered
		}
	}

	return nil
}

// GetRBACStats returns RBAC statistics
func (re *RBACEngine) GetRBACStats(ctx context.Context) map[string]interface{} {
	re.mu.RLock()
	defer re.mu.RUnlock()

	return map[string]interface{}{
		"roles":                len(re.roles),
		"cluster_roles":        len(re.clusterRoles),
		"role_bindings":        len(re.roleBindings),
		"cluster_role_bindings": len(re.clusterRbacs),
		"audit_log_entries":    len(re.auditLog),
	}
}

// ResourceQuotaManager manages resource quotas per namespace
type ResourceQuotaManager struct {
	mu     sync.RWMutex
	quotas map[string]*Quota
}

// Quota defines resource limits for a namespace
type Quota struct {
	Namespace string
	Resources map[string]*ResourceLimit
	Used      map[string]int64
	CreatedAt time.Time
}

// ResourceLimit defines limits for a specific resource
type ResourceLimit struct {
	Kind        string
	MaxCount    int64
	MaxCPU      string
	MaxMemory   string
	Description string
}

// NewResourceQuotaManager creates a new quota manager
func NewResourceQuotaManager() *ResourceQuotaManager {
	return &ResourceQuotaManager{
		quotas: make(map[string]*Quota),
	}
}

// CreateQuota creates a new quota
func (rq *ResourceQuotaManager) CreateQuota(ctx context.Context, quota *Quota) error {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	if quota.Namespace == "" {
		return fmt.Errorf("namespace required")
	}

	if _, exists := rq.quotas[quota.Namespace]; exists {
		return fmt.Errorf("quota already exists")
	}

	quota.CreatedAt = time.Now()
	if quota.Used == nil {
		quota.Used = make(map[string]int64)
	}

	rq.quotas[quota.Namespace] = quota
	return nil
}

// CanAllocate checks if resource can be allocated
func (rq *ResourceQuotaManager) CanAllocate(ctx context.Context, namespace string, kind string, resource map[string]interface{}) (bool, string) {
	rq.mu.RLock()
	defer rq.mu.RUnlock()

	quota, exists := rq.quotas[namespace]
	if !exists {
		return true, "" // No quota defined
	}

	limit, hasLimit := quota.Resources[kind]
	if !hasLimit {
		return true, "" // No limit for this resource kind
	}

	used := quota.Used[kind]
	if used >= limit.MaxCount {
		return false, fmt.Sprintf("quota exceeded for %s: %d/%d", kind, used, limit.MaxCount)
	}

	return true, ""
}

// RecordUsage records resource usage
func (rq *ResourceQuotaManager) RecordUsage(ctx context.Context, namespace string, kind string, count int64) error {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	quota, exists := rq.quotas[namespace]
	if !exists {
		return fmt.Errorf("quota not found")
	}

	quota.Used[kind] += count
	return nil
}

// GetQuota retrieves quota for a namespace
func (rq *ResourceQuotaManager) GetQuota(ctx context.Context, namespace string) (*Quota, error) {
	rq.mu.RLock()
	defer rq.mu.RUnlock()

	if quota, exists := rq.quotas[namespace]; exists {
		return quota, nil
	}

	return nil, fmt.Errorf("quota not found")
}
