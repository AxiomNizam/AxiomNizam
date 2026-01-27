package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RLSPolicy defines row-level security policy
type RLSPolicy struct {
	ID           string
	TableName    string
	PolicyName   string
	Description  string
	UserID       string
	RoleID       string
	Attributes   map[string]string
	Predicate    string // SQL WHERE clause predicate
	PolicyType   string // SELECT, INSERT, UPDATE, DELETE
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserContext contains user information for policy evaluation
type UserContext struct {
	UserID     string
	RoleID     string
	Attributes map[string]string
	TenantID   string
	Groups     []string
}

// RowLevelSecurityManager manages RLS policies and enforcement
type RowLevelSecurityManager struct {
	mu              sync.RWMutex
	policies        map[string][]*RLSPolicy // table -> policies
	userContexts    map[string]*UserContext  // user_id -> context
	policyCache     map[string]*PolicyCache
	attributeRules  map[string]func(*UserContext, map[string]interface{}) bool
	deniedRows      map[string]bool // cache of denied rows
	maxCacheSize    int
	policyCacheTTL  time.Duration
	auditLog        []*RLSAuditLog
	maxAuditLogSize int
}

// PolicyCache caches policy evaluation results
type PolicyCache struct {
	RuleID    string
	UserID    string
	Timestamp time.Time
	Result    bool
}

// RLSAuditLog tracks RLS decisions
type RLSAuditLog struct {
	ID           string
	Timestamp    time.Time
	UserID       string
	TableName    string
	Operation    string
	RowID        interface{}
	Result       string // allowed, denied
	PolicyID     string
	Reason       string
}

// NewRowLevelSecurityManager creates a new RLS manager
func NewRowLevelSecurityManager() *RowLevelSecurityManager {
	return &RowLevelSecurityManager{
		policies:        make(map[string][]*RLSPolicy),
		userContexts:    make(map[string]*UserContext),
		policyCache:     make(map[string]*PolicyCache),
		attributeRules:  make(map[string]func(*UserContext, map[string]interface{}) bool),
		deniedRows:      make(map[string]bool),
		maxCacheSize:    100000,
		policyCacheTTL:  5 * time.Minute,
		auditLog:        make([]*RLSAuditLog, 0),
		maxAuditLogSize: 10000,
	}
}

// AddPolicy adds an RLS policy
func (rlsm *RowLevelSecurityManager) AddPolicy(policy *RLSPolicy) error {
	rlsm.mu.Lock()
	defer rlsm.mu.Unlock()

	if policy.ID == "" {
		policy.ID = fmt.Sprintf("pol-%d", time.Now().UnixNano())
	}
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	if _, exists := rlsm.policies[policy.TableName]; !exists {
		rlsm.policies[policy.TableName] = make([]*RLSPolicy, 0)
	}

	rlsm.policies[policy.TableName] = append(rlsm.policies[policy.TableName], policy)
	return nil
}

// RemovePolicy removes an RLS policy
func (rlsm *RowLevelSecurityManager) RemovePolicy(tableID, policyID string) error {
	rlsm.mu.Lock()
	defer rlsm.mu.Unlock()

	if policies, exists := rlsm.policies[tableID]; exists {
		for i, p := range policies {
			if p.ID == policyID {
				rlsm.policies[tableID] = append(policies[:i], policies[i+1:]...)
				return nil
			}
		}
	}

	return fmt.Errorf("policy not found")
}

// RegisterUserContext registers user context for authorization
func (rlsm *RowLevelSecurityManager) RegisterUserContext(ctx *UserContext) error {
	rlsm.mu.Lock()
	defer rlsm.mu.Unlock()

	if ctx.UserID == "" {
		return fmt.Errorf("user ID required")
	}

	rlsm.userContexts[ctx.UserID] = ctx
	return nil
}

// CanSelectRow checks if user can select a row
func (rlsm *RowLevelSecurityManager) CanSelectRow(ctx context.Context, userID, tableID string, row map[string]interface{}) (bool, string, error) {
	return rlsm.checkRowAccess(userID, tableID, row, "SELECT")
}

// CanUpdateRow checks if user can update a row
func (rlsm *RowLevelSecurityManager) CanUpdateRow(ctx context.Context, userID, tableID string, row map[string]interface{}) (bool, string, error) {
	return rlsm.checkRowAccess(userID, tableID, row, "UPDATE")
}

// CanDeleteRow checks if user can delete a row
func (rlsm *RowLevelSecurityManager) CanDeleteRow(ctx context.Context, userID, tableID string, row map[string]interface{}) (bool, string, error) {
	return rlsm.checkRowAccess(userID, tableID, row, "DELETE")
}

// CanInsertRow checks if user can insert a row
func (rlsm *RowLevelSecurityManager) CanInsertRow(ctx context.Context, userID, tableID string, row map[string]interface{}) (bool, string, error) {
	return rlsm.checkRowAccess(userID, tableID, row, "INSERT")
}

// checkRowAccess checks if user can access a row
func (rlsm *RowLevelSecurityManager) checkRowAccess(userID, tableID string, row map[string]interface{}, operation string) (bool, string, error) {
	rlsm.mu.RLock()

	userCtx, exists := rlsm.userContexts[userID]
	if !exists {
		rlsm.mu.RUnlock()
		rlsm.auditAccess(userID, tableID, "", operation, "denied", "User context not found")
		return false, "User context not found", nil
	}

	policies, exists := rlsm.policies[tableID]
	if !exists {
		rlsm.mu.RUnlock()
		return true, "No policies", nil
	}

	rlsm.mu.RUnlock()

	// Evaluate all applicable policies
	for _, policy := range policies {
		if !policy.IsActive {
			continue
		}

		if policy.PolicyType != operation && policy.PolicyType != "*" {
			continue
		}

		// Check if policy applies to user
		if !rlsm.policyAppliesToUser(userCtx, policy) {
			continue
		}

		// Evaluate predicate
		if !rlsm.evaluatePredicate(userCtx, row, policy) {
			policyID := policy.ID
			rlsm.auditAccess(userID, tableID, "", operation, "denied", fmt.Sprintf("Policy %s denied access", policyID))
			return false, fmt.Sprintf("Policy %s denied access", policyID), nil
		}
	}

	rlsm.auditAccess(userID, tableID, "", operation, "allowed", "")
	return true, "Access allowed", nil
}

// policyAppliesToUser checks if policy applies to user
func (rlsm *RowLevelSecurityManager) policyAppliesToUser(userCtx *UserContext, policy *RLSPolicy) bool {
	// Check user ID
	if policy.UserID != "" && policy.UserID != userCtx.UserID {
		return false
	}

	// Check role ID
	if policy.RoleID != "" && policy.RoleID != userCtx.RoleID {
		return false
	}

	// Check attributes
	for k, v := range policy.Attributes {
		if userVal, exists := userCtx.Attributes[k]; !exists || userVal != v {
			return false
		}
	}

	return true
}

// evaluatePredicate evaluates if row matches predicate
func (rlsm *RowLevelSecurityManager) evaluatePredicate(userCtx *UserContext, row map[string]interface{}, policy *RLSPolicy) bool {
	// Simple predicate evaluation
	// In production, this would use a full SQL parser

	// Check if user owns the row
	if ownedByField, ok := row["owned_by"]; ok {
		if ownedByField == userCtx.UserID {
			return true
		}
	}

	// Check tenant match
	if tenantField, ok := row["tenant_id"]; ok {
		if tenantField == userCtx.TenantID {
			return true
		}
	}

	// Custom attribute matching
	if customCheck, exists := rlsm.attributeRules[policy.ID]; exists {
		return customCheck(userCtx, row)
	}

	// Default allow if no specific denial
	return true
}

// FilterRows filters rows based on RLS policies
func (rlsm *RowLevelSecurityManager) FilterRows(ctx context.Context, userID, tableID string, rows []map[string]interface{}) ([]map[string]interface{}, error) {
	rlsm.mu.RLock()
	userCtx, exists := rlsm.userContexts[userID]
	rlsm.mu.RUnlock()

	if !exists {
		return make([]map[string]interface{}, 0), fmt.Errorf("user context not found")
	}

	filtered := make([]map[string]interface{}, 0)

	for _, row := range rows {
		allowed, _, err := rlsm.checkRowAccess(userID, tableID, row, "SELECT")
		if err != nil {
			continue
		}

		if allowed {
			filtered = append(filtered, row)
		}
	}

	return filtered, nil
}

// RegisterAttributeRule registers custom attribute-based rule
func (rlsm *RowLevelSecurityManager) RegisterAttributeRule(policyID string, rule func(*UserContext, map[string]interface{}) bool) {
	rlsm.mu.Lock()
	defer rlsm.mu.Unlock()

	rlsm.attributeRules[policyID] = rule
}

// GetApplicablePolicies returns policies applicable to user
func (rlsm *RowLevelSecurityManager) GetApplicablePolicies(userID, tableID string) []*RLSPolicy {
	rlsm.mu.RLock()
	defer rlsm.mu.RUnlock()

	userCtx, exists := rlsm.userContexts[userID]
	if !exists {
		return make([]*RLSPolicy, 0)
	}

	applicable := make([]*RLSPolicy, 0)

	if policies, exists := rlsm.policies[tableID]; exists {
		for _, policy := range policies {
			if rlsm.policyAppliesToUser(userCtx, policy) {
				applicable = append(applicable, policy)
			}
		}
	}

	return applicable
}

// GetAuditLog returns RLS audit log entries
func (rlsm *RowLevelSecurityManager) GetAuditLog(limit int) []*RLSAuditLog {
	rlsm.mu.RLock()
	defer rlsm.mu.RUnlock()

	if limit > len(rlsm.auditLog) {
		limit = len(rlsm.auditLog)
	}
	if limit == 0 {
		return make([]*RLSAuditLog, 0)
	}

	return rlsm.auditLog[len(rlsm.auditLog)-limit:]
}

// auditAccess logs access decision
func (rlsm *RowLevelSecurityManager) auditAccess(userID, tableID, rowID, operation, result, reason string) {
	rlsm.mu.Lock()
	defer rlsm.mu.Unlock()

	log := &RLSAuditLog{
		ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		UserID:    userID,
		TableName: tableID,
		Operation: operation,
		RowID:     rowID,
		Result:    result,
		Reason:    reason,
	}

	rlsm.auditLog = append(rlsm.auditLog, log)

	if len(rlsm.auditLog) > rlsm.maxAuditLogSize {
		rlsm.auditLog = rlsm.auditLog[1:]
	}
}

// GetSecurityStats returns RLS statistics
func (rlsm *RowLevelSecurityManager) GetSecurityStats() map[string]interface{} {
	rlsm.mu.RLock()
	defer rlsm.mu.RUnlock()

	deniedCount := 0
	allowedCount := 0

	for _, log := range rlsm.auditLog {
		if log.Result == "denied" {
			deniedCount++
		} else {
			allowedCount++
		}
	}

	return map[string]interface{}{
		"total_policies":       len(rlsm.policies),
		"active_users":         len(rlsm.userContexts),
		"audit_log_entries":    len(rlsm.auditLog),
		"cache_entries":        len(rlsm.policyCache),
		"access_allowed":       allowedCount,
		"access_denied":        deniedCount,
		"denial_rate":          float64(deniedCount) / float64(deniedCount+allowedCount),
	}
}
