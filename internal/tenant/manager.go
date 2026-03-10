package tenant

import (
	"context"
	"fmt"
	"sync"
)

const (
	errTenantNotFound = "tenant not found: %s"
	errQuotaNotFound  = "quota not found: %s"
)

// TenantManager manages tenant lifecycle
type TenantManager interface {
	// CreateTenant creates new tenant
	CreateTenant(ctx context.Context, tenant *Tenant, ownerID string) (*Tenant, error)

	// GetTenant retrieves tenant by ID
	GetTenant(ctx context.Context, tenantID string) (*Tenant, error)

	// UpdateTenant updates tenant
	UpdateTenant(ctx context.Context, tenant *Tenant) error

	// DeleteTenant soft deletes tenant
	DeleteTenant(ctx context.Context, tenantID string) error

	// ListTenants lists all accessible tenants
	ListTenants(ctx context.Context, ownerID string) ([]*Tenant, error)

	// GetQuota gets tenant resource quota
	GetQuota(ctx context.Context, tenantID string) (*TenantQuota, error)

	// UpdateQuota updates quota limits
	UpdateQuota(ctx context.Context, tenantID string, quota *TenantQuota) error

	// CheckQuota checks if action exceeds quota
	CheckQuota(ctx context.Context, tenantID string, quotaType string, requested int64) error

	// AddMember adds user to tenant
	AddMember(ctx context.Context, tenantID, userID string, role MemberRole) (*TenantMember, error)

	// RemoveMember removes user from tenant
	RemoveMember(ctx context.Context, tenantID, userID string) error

	// ListMembers lists tenant members
	ListMembers(ctx context.Context, tenantID string) ([]*TenantMember, error)

	// UpdateMemberRole changes member role
	UpdateMemberRole(ctx context.Context, tenantID, userID string, role MemberRole) error

	// CanAccess checks if user can access tenant
	CanAccess(ctx context.Context, tenantID, userID string) (bool, error)

	// GetIsolationStrategy returns isolation strategy
	GetIsolationStrategy(ctx context.Context, tenantID string) (TenantIsolation, error)
}

// InMemoryTenantManager in-memory implementation for testing
type InMemoryTenantManager struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant
	members map[string][]*TenantMember // tenantID -> members
	quotas  map[string]*TenantQuota    // tenantID -> quota
}

// NewInMemoryTenantManager creates manager
func NewInMemoryTenantManager() *InMemoryTenantManager {
	return &InMemoryTenantManager{
		tenants: make(map[string]*Tenant),
		members: make(map[string][]*TenantMember),
		quotas:  make(map[string]*TenantQuota),
	}
}

// CreateTenant creates tenant
func (itm *InMemoryTenantManager) CreateTenant(ctx context.Context, tenant *Tenant, ownerID string) (*Tenant, error) {
	itm.mu.Lock()
	defer itm.mu.Unlock()

	if _, exists := itm.tenants[tenant.ID]; exists {
		return nil, fmt.Errorf("tenant already exists: %s", tenant.ID)
	}

	tenant.Owner = ownerID
	tenant.Status = TenantActive
	itm.tenants[tenant.ID] = tenant

	// Create default quota
	quota := &TenantQuota{
		TenantID:      tenant.ID,
		MaxUsers:      10,
		MaxResources:  100,
		MaxQueries:    10000,
		MaxStorage:    1073741824, // 1GB
		MaxAPIcalls:   100000,     // Per day
		MaxConcurrent: 10,
		QueryTimeout:  300, // 5 minutes
	}
	itm.quotas[tenant.ID] = quota

	// Add owner as member
	itm.members[tenant.ID] = []*TenantMember{
		{
			ID:       fmt.Sprintf("%s-owner", tenant.ID),
			TenantID: tenant.ID,
			UserID:   ownerID,
			Role:     RoleOwner,
			Status:   MemberActive,
		},
	}

	return tenant, nil
}

// GetTenant retrieves tenant
func (itm *InMemoryTenantManager) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	itm.mu.RLock()
	defer itm.mu.RUnlock()

	tenant, exists := itm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf(errTenantNotFound, tenantID)
	}
	return tenant, nil
}

// UpdateTenant updates tenant
func (itm *InMemoryTenantManager) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	itm.mu.Lock()
	defer itm.mu.Unlock()

	if _, exists := itm.tenants[tenant.ID]; !exists {
		return fmt.Errorf(errTenantNotFound, tenant.ID)
	}
	itm.tenants[tenant.ID] = tenant
	return nil
}

// DeleteTenant soft deletes
func (itm *InMemoryTenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	itm.mu.Lock()
	defer itm.mu.Unlock()

	tenant, exists := itm.tenants[tenantID]
	if !exists {
		return fmt.Errorf(errTenantNotFound, tenantID)
	}
	tenant.Status = TenantArchived
	return nil
}

// ListTenants lists tenants
func (itm *InMemoryTenantManager) ListTenants(ctx context.Context, ownerID string) ([]*Tenant, error) {
	itm.mu.RLock()
	defer itm.mu.RUnlock()

	result := make([]*Tenant, 0)
	for _, tenant := range itm.tenants {
		if tenant.Owner == ownerID && tenant.Status == TenantActive {
			result = append(result, tenant)
		}
	}
	return result, nil
}

// GetQuota retrieves quota
func (itm *InMemoryTenantManager) GetQuota(ctx context.Context, tenantID string) (*TenantQuota, error) {
	itm.mu.RLock()
	defer itm.mu.RUnlock()

	quota, exists := itm.quotas[tenantID]
	if !exists {
		return nil, fmt.Errorf(errQuotaNotFound, tenantID)
	}
	return quota, nil
}

// UpdateQuota updates quota
func (itm *InMemoryTenantManager) UpdateQuota(ctx context.Context, tenantID string, quota *TenantQuota) error {
	itm.mu.Lock()
	defer itm.mu.Unlock()

	if _, exists := itm.quotas[tenantID]; !exists {
		return fmt.Errorf(errQuotaNotFound, tenantID)
	}
	itm.quotas[tenantID] = quota
	return nil
}

// CheckQuota checks quota
func (itm *InMemoryTenantManager) CheckQuota(ctx context.Context, tenantID string, quotaType string, requested int64) error {
	itm.mu.RLock()
	defer itm.mu.RUnlock()

	quota, exists := itm.quotas[tenantID]
	if !exists {
		return fmt.Errorf(errQuotaNotFound, tenantID)
	}

	switch quotaType {
	case "users":
		if quota.UsedUsers+int(requested) > quota.MaxUsers {
			return fmt.Errorf("user quota exceeded: %d/%d", quota.UsedUsers, quota.MaxUsers)
		}
	case "resources":
		if quota.UsedResources+int(requested) > quota.MaxResources {
			return fmt.Errorf("resource quota exceeded: %d/%d", quota.UsedResources, quota.MaxResources)
		}
	case "queries":
		if quota.UsedQueries+requested > quota.MaxQueries {
			return fmt.Errorf("query quota exceeded: %d/%d", quota.UsedQueries, quota.MaxQueries)
		}
	case "storage":
		if quota.UsedStorage+requested > quota.MaxStorage {
			return fmt.Errorf("storage quota exceeded: %d/%d", quota.UsedStorage, quota.MaxStorage)
		}
	case "api":
		if quota.UsedAPICalls+requested > quota.MaxAPIcalls {
			return fmt.Errorf("API call quota exceeded: %d/%d", quota.UsedAPICalls, quota.MaxAPIcalls)
		}
	}

	return nil
}

// AddMember adds member
func (itm *InMemoryTenantManager) AddMember(ctx context.Context, tenantID, userID string, role MemberRole) (*TenantMember, error) {
	itm.mu.Lock()
	defer itm.mu.Unlock()

	if _, exists := itm.tenants[tenantID]; !exists {
		return nil, fmt.Errorf(errTenantNotFound, tenantID)
	}

	member := &TenantMember{
		ID:       fmt.Sprintf("%s-%s", tenantID, userID),
		TenantID: tenantID,
		UserID:   userID,
		Role:     role,
		Status:   MemberActive,
	}

	itm.members[tenantID] = append(itm.members[tenantID], member)
	return member, nil
}

// RemoveMember removes member
func (itm *InMemoryTenantManager) RemoveMember(ctx context.Context, tenantID, userID string) error {
	itm.mu.Lock()
	defer itm.mu.Unlock()

	members := itm.members[tenantID]
	filtered := make([]*TenantMember, 0)
	found := false

	for _, m := range members {
		if m.UserID != userID {
			filtered = append(filtered, m)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("member not found: %s", userID)
	}

	itm.members[tenantID] = filtered
	return nil
}

// ListMembers lists members
func (itm *InMemoryTenantManager) ListMembers(ctx context.Context, tenantID string) ([]*TenantMember, error) {
	itm.mu.RLock()
	defer itm.mu.RUnlock()

	members := make([]*TenantMember, len(itm.members[tenantID]))
	copy(members, itm.members[tenantID])
	return members, nil
}

// UpdateMemberRole updates role
func (itm *InMemoryTenantManager) UpdateMemberRole(ctx context.Context, tenantID, userID string, role MemberRole) error {
	itm.mu.Lock()
	defer itm.mu.Unlock()

	for _, m := range itm.members[tenantID] {
		if m.UserID == userID {
			m.Role = role
			return nil
		}
	}

	return fmt.Errorf("member not found: %s", userID)
}

// CanAccess checks access
func (itm *InMemoryTenantManager) CanAccess(ctx context.Context, tenantID, userID string) (bool, error) {
	itm.mu.RLock()
	defer itm.mu.RUnlock()

	for _, m := range itm.members[tenantID] {
		if m.UserID == userID && m.Status == MemberActive {
			return true, nil
		}
	}

	return false, nil
}

// GetIsolationStrategy gets strategy
func (itm *InMemoryTenantManager) GetIsolationStrategy(ctx context.Context, tenantID string) (TenantIsolation, error) {
	tenant, err := itm.GetTenant(ctx, tenantID)
	if err != nil {
		return "", err
	}
	return tenant.IsolationLevel, nil
}
