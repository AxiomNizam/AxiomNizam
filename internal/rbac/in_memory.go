package rbac

import (
	"fmt"
	"sync"
	"time"
)

// InMemoryRBACManager in-memory RBAC implementation
type InMemoryRBACManager struct {
	mu             sync.RWMutex
	roles          map[string]*Role
	bindings       map[string]*RoleBinding
	permissions    map[string]*Permission
	accessRequests map[string]*AccessRequest
}

// NewInMemoryRBACManager creates manager
func NewInMemoryRBACManager() *InMemoryRBACManager {
	return &InMemoryRBACManager{
		roles:          make(map[string]*Role),
		bindings:       make(map[string]*RoleBinding),
		permissions:    make(map[string]*Permission),
		accessRequests: make(map[string]*AccessRequest),
	}
}

// CreateRole creates role
func (m *InMemoryRBACManager) CreateRole(role *Role) (*Role, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if role.ID == "" {
		role.ID = fmt.Sprintf("role-%d", time.Now().UnixNano())
	}
	if role.CreatedAt.IsZero() {
		role.CreatedAt = time.Now()
	}

	m.roles[role.ID] = role
	return role, nil
}

// GetRole retrieves role
func (m *InMemoryRBACManager) GetRole(id string) (*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	role, exists := m.roles[id]
	if !exists {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

// ListRoles lists roles
func (m *InMemoryRBACManager) ListRoles(tenantID string) ([]*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Role
	for _, r := range m.roles {
		if tenantID != "" && r.TenantID != tenantID {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

// UpdateRole updates role
func (m *InMemoryRBACManager) UpdateRole(role *Role) (*Role, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.roles[role.ID]; !exists {
		return nil, fmt.Errorf("role not found")
	}

	role.UpdatedAt = time.Now()
	m.roles[role.ID] = role
	return role, nil
}

// DeleteRole deletes role
func (m *InMemoryRBACManager) DeleteRole(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.roles, id)
	return nil
}

// CreateRoleBinding creates role binding
func (m *InMemoryRBACManager) CreateRoleBinding(binding *RoleBinding) (*RoleBinding, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if binding.ID == "" {
		binding.ID = fmt.Sprintf("binding-%d", time.Now().UnixNano())
	}
	if binding.CreatedAt.IsZero() {
		binding.CreatedAt = time.Now()
	}

	m.bindings[binding.ID] = binding
	return binding, nil
}

// GetRoleBinding retrieves role binding
func (m *InMemoryRBACManager) GetRoleBinding(id string) (*RoleBinding, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	binding, exists := m.bindings[id]
	if !exists {
		return nil, fmt.Errorf("binding not found")
	}
	return binding, nil
}

// ListRoleBindings lists role bindings
func (m *InMemoryRBACManager) ListRoleBindings(roleID, subjectID string) ([]*RoleBinding, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*RoleBinding
	for _, b := range m.bindings {
		if roleID != "" && b.RoleID != roleID {
			continue
		}
		if subjectID != "" && b.SubjectID != subjectID {
			continue
		}
		result = append(result, b)
	}
	return result, nil
}

// DeleteRoleBinding deletes role binding
func (m *InMemoryRBACManager) DeleteRoleBinding(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.bindings, id)
	return nil
}

// CreatePermission creates permission
func (m *InMemoryRBACManager) CreatePermission(permission *Permission) (*Permission, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if permission.ID == "" {
		permission.ID = fmt.Sprintf("perm-%d", time.Now().UnixNano())
	}

	m.permissions[permission.ID] = permission
	return permission, nil
}

// ListPermissions lists permissions
func (m *InMemoryRBACManager) ListPermissions(roleID string) ([]*Permission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Permission
	for _, p := range m.permissions {
		if roleID != "" && p.RoleID != roleID {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

// CheckPermission checks if subject has permission
func (m *InMemoryRBACManager) CheckPermission(subjectID, resource, action string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get role bindings for subject
	for _, binding := range m.bindings {
		if binding.SubjectID != subjectID {
			continue
		}

		// Get role
		role, exists := m.roles[binding.RoleID]
		if !exists {
			continue
		}

		// Check if role has permission
		for _, perm := range m.permissions {
			if perm.RoleID == role.ID && perm.Resource == resource && perm.Action == action {
				return true, nil
			}
		}
	}

	return false, nil
}

// CreateAccessRequest creates access request
func (m *InMemoryRBACManager) CreateAccessRequest(request *AccessRequest) (*AccessRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if request.ID == "" {
		request.ID = fmt.Sprintf("request-%d", time.Now().UnixNano())
	}
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now()
	}

	request.Status = "pending"
	m.accessRequests[request.ID] = request
	return request, nil
}

// GetAccessRequest retrieves access request
func (m *InMemoryRBACManager) GetAccessRequest(id string) (*AccessRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	request, exists := m.accessRequests[id]
	if !exists {
		return nil, fmt.Errorf("request not found")
	}
	return request, nil
}

// ListAccessRequests lists access requests
func (m *InMemoryRBACManager) ListAccessRequests(subjectID, status string) ([]*AccessRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*AccessRequest
	for _, r := range m.accessRequests {
		if subjectID != "" && r.SubjectID != subjectID {
			continue
		}
		if status != "" && r.Status != status {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

// ApproveAccessRequest approves access request
func (m *InMemoryRBACManager) ApproveAccessRequest(requestID, approverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	request, exists := m.accessRequests[requestID]
	if !exists {
		return fmt.Errorf("request not found")
	}

	request.Status = "approved"
	request.ApprovedAt = time.Now()
	request.ApprovedBy = approverID

	// Create role binding
	binding := &RoleBinding{
		ID:        fmt.Sprintf("binding-%d", time.Now().UnixNano()),
		RoleID:    request.RoleID,
		SubjectID: request.SubjectID,
		CreatedAt: time.Now(),
	}
	m.bindings[binding.ID] = binding

	return nil
}

// RejectAccessRequest rejects access request
func (m *InMemoryRBACManager) RejectAccessRequest(requestID, approverID, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	request, exists := m.accessRequests[requestID]
	if !exists {
		return fmt.Errorf("request not found")
	}

	request.Status = "rejected"
	request.RejectionReason = reason
	request.RejectedBy = approverID

	return nil
}
