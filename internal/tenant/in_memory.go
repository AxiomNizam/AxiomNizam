package tenant

import (
	"fmt"
	"sync"
	"time"
)

// InMemoryTenantManager in-memory implementation
type InMemoryTenantManager struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant
	members map[string][]*TenantMember
}

// NewInMemoryTenantManager creates manager
func NewInMemoryTenantManager() *InMemoryTenantManager {
	return &InMemoryTenantManager{
		tenants: make(map[string]*Tenant),
		members: make(map[string][]*TenantMember),
	}
}

// CreateTenant creates new tenant
func (m *InMemoryTenantManager) CreateTenant(tenant *Tenant, owner string) (*Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tenant.ID == "" {
		tenant.ID = fmt.Sprintf("tenant-%d", time.Now().UnixNano())
	}

	tenant.Status = "Active"
	tenant.CreatedAt = time.Now()
	m.tenants[tenant.ID] = tenant

	// Add owner as member
	member := &TenantMember{
		ID:       fmt.Sprintf("member-%d", time.Now().UnixNano()),
		TenantID: tenant.ID,
		UserID:   owner,
		Role:     "Owner",
		Status:   "Active",
		JoinedAt: time.Now(),
	}
	m.members[tenant.ID] = append(m.members[tenant.ID], member)

	return tenant, nil
}

// GetTenant retrieves tenant
func (m *InMemoryTenantManager) GetTenant(id string) (*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, exists := m.tenants[id]
	if !exists {
		return nil, fmt.Errorf("tenant not found")
	}
	return tenant, nil
}

// UpdateTenant updates tenant
func (m *InMemoryTenantManager) UpdateTenant(tenant *Tenant) (*Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tenants[tenant.ID]; !exists {
		return nil, fmt.Errorf("tenant not found")
	}

	tenant.UpdatedAt = time.Now()
	m.tenants[tenant.ID] = tenant
	return tenant, nil
}

// DeleteTenant deletes tenant
func (m *InMemoryTenantManager) DeleteTenant(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tenant, exists := m.tenants[id]; exists {
		tenant.Status = "Archived"
		tenant.UpdatedAt = time.Now()
	}
	return nil
}

// ListTenants lists all tenants
func (m *InMemoryTenantManager) ListTenants() ([]*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Tenant
	for _, tenant := range m.tenants {
		if tenant.Status != "Archived" {
			result = append(result, tenant)
		}
	}
	return result, nil
}

// AddMember adds member to tenant
func (m *InMemoryTenantManager) AddMember(member *TenantMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if member.ID == "" {
		member.ID = fmt.Sprintf("member-%d", time.Now().UnixNano())
	}

	member.JoinedAt = time.Now()
	m.members[member.TenantID] = append(m.members[member.TenantID], member)
	return nil
}

// RemoveMember removes member from tenant
func (m *InMemoryTenantManager) RemoveMember(tenantID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	members := m.members[tenantID]
	for i, member := range members {
		if member.UserID == userID {
			members = append(members[:i], members[i+1:]...)
			m.members[tenantID] = members
			return nil
		}
	}
	return fmt.Errorf("member not found")
}

// GetQuota retrieves quota
func (m *InMemoryTenantManager) GetQuota(tenantID string) (*TenantQuota, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found")
	}
	return tenant.Quota, nil
}

// CheckQuota checks if resource limit exceeded
func (m *InMemoryTenantManager) CheckQuota(tenantID, resource string, amount int64) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return false, fmt.Errorf("tenant not found")
	}

	if tenant.Quota == nil {
		return true, nil // No quota limit
	}

	// Check specific resource limits
	switch resource {
	case "users":
		return tenant.Quota.UsersLimit == 0 || tenant.Quota.UsersUsed < tenant.Quota.UsersLimit, nil
	case "resources":
		return tenant.Quota.ResourcesLimit == 0 || tenant.Quota.ResourcesUsed < tenant.Quota.ResourcesLimit, nil
	}

	return true, nil
}
