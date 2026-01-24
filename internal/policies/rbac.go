package policies

import (
	"fmt"
)

// Role represents a user role in the system
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleUser    Role = "user"
	RoleGuest   Role = "guest"
)

// Permission represents a specific action that can be performed
type Permission string

// User permissions
const (
	PermissionCreateUser  Permission = "users:create"
	PermissionReadUser    Permission = "users:read"
	PermissionUpdateUser  Permission = "users:update"
	PermissionDeleteUser  Permission = "users:delete"
	PermissionListUsers   Permission = "users:list"
	PermissionManageRoles Permission = "users:manage_roles"
)

// Policy defines what a role can do
type Policy struct {
	Role        Role
	Permissions map[Permission]bool
}

// RBACManager manages role-based access control
type RBACManager struct {
	policies map[Role]*Policy
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager() *RBACManager {
	manager := &RBACManager{
		policies: make(map[Role]*Policy),
	}

	// Define policies for each role
	manager.defineAdminPolicy()
	manager.defineManagerPolicy()
	manager.defineUserPolicy()
	manager.defineGuestPolicy()

	return manager
}

// defineAdminPolicy sets up admin permissions (full access)
func (rm *RBACManager) defineAdminPolicy() {
	rm.policies[RoleAdmin] = &Policy{
		Role: RoleAdmin,
		Permissions: map[Permission]bool{
			PermissionCreateUser:  true,
			PermissionReadUser:    true,
			PermissionUpdateUser:  true,
			PermissionDeleteUser:  true,
			PermissionListUsers:   true,
			PermissionManageRoles: true,
		},
	}
}

// defineManagerPolicy sets up manager permissions
func (rm *RBACManager) defineManagerPolicy() {
	rm.policies[RoleManager] = &Policy{
		Role: RoleManager,
		Permissions: map[Permission]bool{
			PermissionCreateUser:  true,
			PermissionReadUser:    true,
			PermissionUpdateUser:  true,
			PermissionListUsers:   true,
			PermissionDeleteUser:  false, // Cannot delete users
			PermissionManageRoles: false, // Cannot manage roles
		},
	}
}

// defineUserPolicy sets up user permissions (basic access)
func (rm *RBACManager) defineUserPolicy() {
	rm.policies[RoleUser] = &Policy{
		Role: RoleUser,
		Permissions: map[Permission]bool{
			PermissionCreateUser:  false,
			PermissionReadUser:    true,
			PermissionUpdateUser:  true, // Can update self
			PermissionDeleteUser:  false,
			PermissionListUsers:   false, // Cannot list all users
			PermissionManageRoles: false,
		},
	}
}

// defineGuestPolicy sets up guest permissions (read-only)
func (rm *RBACManager) defineGuestPolicy() {
	rm.policies[RoleGuest] = &Policy{
		Role: RoleGuest,
		Permissions: map[Permission]bool{
			PermissionCreateUser:  false,
			PermissionReadUser:    false, // Limited access
			PermissionUpdateUser:  false,
			PermissionDeleteUser:  false,
			PermissionListUsers:   false,
			PermissionManageRoles: false,
		},
	}
}

// HasPermission checks if a role has a specific permission
func (rm *RBACManager) HasPermission(role Role, permission Permission) bool {
	policy, exists := rm.policies[role]
	if !exists {
		return false
	}

	hasPermission, exists := policy.Permissions[permission]
	return hasPermission && exists
}

// CanUserCreateUser checks if a user with given role can create users
func (rm *RBACManager) CanUserCreateUser(role Role) bool {
	return rm.HasPermission(role, PermissionCreateUser)
}

// CanUserReadUser checks if a user can read another user
func (rm *RBACManager) CanUserReadUser(role Role) bool {
	return rm.HasPermission(role, PermissionReadUser)
}

// CanUserUpdateUser checks if a user can update another user
func (rm *RBACManager) CanUserUpdateUser(role Role) bool {
	return rm.HasPermission(role, PermissionUpdateUser)
}

// CanUserDeleteUser checks if a user can delete users
func (rm *RBACManager) CanUserDeleteUser(role Role) bool {
	return rm.HasPermission(role, PermissionDeleteUser)
}

// CanUserListUsers checks if a user can list all users
func (rm *RBACManager) CanUserListUsers(role Role) bool {
	return rm.HasPermission(role, PermissionListUsers)
}

// CanUserManageRoles checks if a user can manage roles
func (rm *RBACManager) CanUserManageRoles(role Role) bool {
	return rm.HasPermission(role, PermissionManageRoles)
}

// GetRolePermissions returns all permissions for a role
func (rm *RBACManager) GetRolePermissions(role Role) ([]Permission, error) {
	policy, exists := rm.policies[role]
	if !exists {
		return nil, fmt.Errorf("role %s not found", role)
	}

	var permissions []Permission
	for perm, allowed := range policy.Permissions {
		if allowed {
			permissions = append(permissions, perm)
		}
	}

	return permissions, nil
}

// ValidatePermission is a helper to check and log permission checks
func (rm *RBACManager) ValidatePermission(role Role, permission Permission) (bool, error) {
	if !rm.HasPermission(role, permission) {
		return false, fmt.Errorf("role '%s' does not have permission '%s'", role, permission)
	}
	return true, nil
}

// AccessControl wraps authorization logic
type AccessControl struct {
	rbacManager *RBACManager
}

// NewAccessControl creates a new access control instance
func NewAccessControl() *AccessControl {
	return &AccessControl{
		rbacManager: NewRBACManager(),
	}
}

// CheckPermission checks if action is allowed
func (ac *AccessControl) CheckPermission(userRole Role, requiredPermission Permission) error {
	if !ac.rbacManager.HasPermission(userRole, requiredPermission) {
		return fmt.Errorf("access denied: %s permission required", requiredPermission)
	}
	return nil
}

// CheckMultiplePermissions checks if user has any of the required permissions
func (ac *AccessControl) CheckMultiplePermissions(userRole Role, permissions ...Permission) error {
	for _, permission := range permissions {
		if ac.rbacManager.HasPermission(userRole, permission) {
			return nil // User has at least one permission
		}
	}
	return fmt.Errorf("access denied: required permissions not found")
}

// Example usage in services
/*

// In UserService
func (s *userService) CreateUser(ctx context.Context, requesterRole Role, user *models.User) (*models.User, error) {
    // Check permission
    if err := s.accessControl.CheckPermission(requesterRole, PermissionCreateUser); err != nil {
        s.LogError("CreateUser permission denied", err)
        return nil, services.ErrForbidden
    }

    // ... rest of creation logic
}

// In AuthService
func (s *authService) UpdateUserRole(ctx context.Context, requesterRole Role, userID string, newRole Role) error {
    // Check if requester can manage roles
    if err := s.accessControl.CheckPermission(requesterRole, PermissionManageRoles); err != nil {
        s.LogError("UpdateUserRole permission denied", err)
        return services.ErrForbidden
    }

    // ... rest of update logic
}

*/
