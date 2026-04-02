package authz

import (
	"errors"
	"strings"
	"sync"
	"time"
)

// Role represents an IAM role.
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	System      bool         `json:"system"` // system roles cannot be deleted
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Permission represents a single permission grant.
type Permission struct {
	Resource string `json:"resource"` // e.g. "users", "clients", "roles", "*"
	Action   string `json:"action"`   // "create", "read", "update", "delete", "execute", "*"
}

// RoleBinding ties a user to a role.
type RoleBinding struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateRoleRequest is the admin API payload.
type CreateRoleRequest struct {
	Name        string       `json:"name" binding:"required"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions" binding:"required"`
}

// AssignRoleRequest binds a user to a role.
type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}

// RoleRepository persists roles.
type RoleRepository interface {
	CreateRole(role *Role) error
	GetRole(id string) (*Role, error)
	GetRoleByName(name string) (*Role, error)
	ListRoles() ([]*Role, error)
	UpdateRole(role *Role) error
	DeleteRole(id string) error
}

// RoleBindingRepository persists user-role mappings.
type RoleBindingRepository interface {
	CreateBinding(b *RoleBinding) error
	DeleteBinding(id string) error
	ListBindingsForUser(userID string) ([]*RoleBinding, error)
	ListAllBindings() ([]*RoleBinding, error)
}

// Authorizer evaluates permissions.
type Authorizer struct {
	mu       sync.RWMutex
	roles    RoleRepository
	bindings RoleBindingRepository
}

// NewAuthorizer creates an authorizer.
func NewAuthorizer(roles RoleRepository, bindings RoleBindingRepository) *Authorizer {
	return &Authorizer{roles: roles, bindings: bindings}
}

// GetUserRoles returns all roles assigned to a user.
func (a *Authorizer) GetUserRoles(userID string) ([]*Role, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	bindings, err := a.bindings.ListBindingsForUser(userID)
	if err != nil {
		return nil, err
	}

	roles := make([]*Role, 0, len(bindings))
	for _, b := range bindings {
		role, err := a.roles.GetRole(b.RoleID)
		if err != nil {
			continue
		}
		if role != nil {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

// GetUserRoleNames returns a flat list of role names for a user.
func (a *Authorizer) GetUserRoleNames(userID string) ([]string, error) {
	roles, err := a.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(roles))
	for _, r := range roles {
		names = append(names, r.Name)
	}
	return names, nil
}

// CheckPermission evaluates whether a user has a specific permission.
func (a *Authorizer) CheckPermission(userID, resource, action string) (bool, error) {
	roles, err := a.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	resource = strings.ToLower(strings.TrimSpace(resource))
	action = strings.ToLower(strings.TrimSpace(action))

	for _, role := range roles {
		// Sysadmin role has implicit all-access
		if strings.ToLower(role.Name) == "sysadmin" {
			return true, nil
		}
		for _, perm := range role.Permissions {
			rMatch := perm.Resource == "*" || strings.ToLower(perm.Resource) == resource
			aMatch := perm.Action == "*" || strings.ToLower(perm.Action) == action
			if rMatch && aMatch {
				return true, nil
			}
		}
	}
	return false, nil
}

// AssignRole binds a role to a user.
func (a *Authorizer) AssignRole(userID, roleID string) (*RoleBinding, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	role, err := a.roles.GetRole(roleID)
	if err != nil || role == nil {
		return nil, errors.New("role not found")
	}

	// Prevent duplicate bindings
	existing, _ := a.bindings.ListBindingsForUser(userID)
	for _, b := range existing {
		if b.RoleID == roleID {
			return b, nil // already assigned
		}
	}

	b := &RoleBinding{
		ID:        generateBindingID(userID, roleID),
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: time.Now().UTC(),
	}
	if err := a.bindings.CreateBinding(b); err != nil {
		return nil, err
	}
	return b, nil
}

// RevokeRole removes a role binding.
func (a *Authorizer) RevokeRole(bindingID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.bindings.DeleteBinding(bindingID)
}

func generateBindingID(userID, roleID string) string {
	return "rb-" + userID[:8] + "-" + roleID[:8] + "-" + time.Now().UTC().Format("20060102150405")
}

// DefaultSystemRoles returns the built-in IAM roles.
func DefaultSystemRoles() []*Role {
	now := time.Now().UTC()
	return []*Role{
		{
			ID:          "role-sysadmin",
			Name:        "sysadmin",
			Description: "Master realm administrator with full access",
			System:      true,
			CreatedAt:   now,
			UpdatedAt:   now,
			Permissions: []Permission{
				{Resource: "*", Action: "*"},
			},
		},
		{
			ID:          "role-admin",
			Name:        "admin",
			Description: "Application administrator",
			System:      true,
			CreatedAt:   now,
			UpdatedAt:   now,
			Permissions: []Permission{
				{Resource: "users", Action: "read"},
				{Resource: "users", Action: "create"},
				{Resource: "users", Action: "update"},
				{Resource: "clients", Action: "read"},
				{Resource: "clients", Action: "create"},
				{Resource: "roles", Action: "read"},
				{Resource: "audit", Action: "read"},
			},
		},
		{
			ID:          "role-user",
			Name:        "user",
			Description: "Standard authenticated user",
			System:      true,
			CreatedAt:   now,
			UpdatedAt:   now,
			Permissions: []Permission{
				{Resource: "profile", Action: "read"},
				{Resource: "profile", Action: "update"},
			},
		},
	}
}
