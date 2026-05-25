package rbac

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RBACHandler struct {
	manager            RBACManager
	roleDualWriteStore RBACRoleDualWriteStore
}

// NewRBACHandler creates handler
func NewRBACHandler(manager RBACManager) *RBACHandler {
	return &RBACHandler{manager: manager}
}

// CreateRole handles POST /api/v1/roles
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req struct {
		TenantID    string       `json:"tenantId" binding:"required"`
		Name        string       `json:"name" binding:"required"`
		Description string       `json:"description"`
		Permissions []Permission `json:"permissions"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: err.Error()})
		return
	}

	// Phase 3: reconciler-authoritative path
	if h.isAuthoritative() {
		role := &Role{TenantID: req.TenantID, Name: req.Name, Description: req.Description, Type: "CUSTOM", Permissions: req.Permissions, IsActive: true}
		resource := h.buildRoleResource(role)
		resource.Status.Phase = "Pending"
		if h.roleDualWriteStore != nil {
			if err := h.roleDualWriteStore.Create(c.Request.Context(), resource); err != nil {
				_ = h.roleDualWriteStore.Update(c.Request.Context(), resource)
			}
		}
		c.JSON(http.StatusAccepted, RoleCreatedResponse{Name: resource.Name, Status: "Pending", Message: "role resource created"})
		return
	}

	role := &Role{
		TenantID:    req.TenantID,
		Name:        req.Name,
		Description: req.Description,
		Type:        "CUSTOM",
		Permissions: req.Permissions,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	created, err := h.manager.CreateRole(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	h.dualWriteRole(created)
	c.JSON(http.StatusCreated, created)
}

// GetRole handles GET /api/v1/roles/:id
func (h *RBACHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	role, err := h.manager.GetRole(id)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: "role not found"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// ListRoles handles GET /api/v1/roles
func (h *RBACHandler) ListRoles(c *gin.Context) {
	tenantID := c.Query("tenantId")
	roles, err := h.manager.ListRoles(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, RoleListResponse{Roles: roles, Count: len(roles)})
}

// UpdateRole handles PATCH /api/v1/roles/:id
func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: err.Error()})
		return
	}

	role, err := h.manager.GetRole(id)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Error: "role not found"})
		return
	}

	role.UpdatedAt = time.Now()
	updated, err := h.manager.UpdateRole(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteRole handles DELETE /api/v1/roles/:id
func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.DeleteRole(id); err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "role deleted"})
}

// BindRole handles POST /api/v1/role-bindings
func (h *RBACHandler) BindRole(c *gin.Context) {
	var req struct {
		TenantID      string        `json:"tenantId" binding:"required"`
		RoleID        string        `json:"roleId" binding:"required"`
		PrincipalType PrincipalType `json:"principalType" binding:"required"`
		PrincipalID   string        `json:"principalId" binding:"required"`
		Scope         string        `json:"scope"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: err.Error()})
		return
	}

	binding := &RoleBinding{
		TenantID:      req.TenantID,
		RoleID:        req.RoleID,
		PrincipalType: req.PrincipalType,
		PrincipalID:   req.PrincipalID,
		Scope:         req.Scope,
		Effective:     true,
		CreatedAt:     time.Now(),
	}

	created, err := h.manager.CreateRoleBinding(binding)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// ListBindings handles GET /api/v1/role-bindings
func (h *RBACHandler) ListBindings(c *gin.Context) {
	tenantID := c.Query("tenantId")
	principalID := c.Query("principalId")

	bindings, err := h.manager.ListRoleBindings(tenantID, principalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, BindingListResponse{Bindings: bindings, Count: len(bindings)})
}

// DeleteBinding handles DELETE /api/v1/role-bindings/:id
func (h *RBACHandler) DeleteBinding(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.DeleteRoleBinding(id); err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "binding deleted"})
}

// CheckPermission handles POST /api/v1/permissions/check
func (h *RBACHandler) CheckPermission(c *gin.Context) {
	var req PermissionCheck
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: err.Error()})
		return
	}

	result, err := h.manager.CheckPermission(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListPermissions handles GET /api/v1/permissions
func (h *RBACHandler) ListPermissions(c *gin.Context) {
	tenantID := c.Query("tenantId")
	resource := c.Query("resource")

	permissions, err := h.manager.ListPermissions(tenantID, resource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, PermissionListResponse{Permissions: permissions, Count: len(permissions)})
}

// CreateAccessRequest handles POST /api/v1/access-requests
func (h *RBACHandler) CreateAccessRequest(c *gin.Context) {
	var req struct {
		TenantID      string `json:"tenantId" binding:"required"`
		PrincipalID   string `json:"principalId" binding:"required"`
		ResourceType  string `json:"resourceType" binding:"required"`
		ResourceID    string `json:"resourceId"`
		Action        string `json:"action" binding:"required"`
		Duration      int    `json:"duration"`
		Justification string `json:"justification"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: err.Error()})
		return
	}

	requestedAt := time.Now()
	var expiresAt time.Time
	if req.Duration > 0 {
		expiresAt = requestedAt.Add(time.Duration(req.Duration) * time.Second)
	}

	accessReq := &AccessRequest{
		TenantID:      req.TenantID,
		PrincipalType: PrincipalTypeUser,
		PrincipalID:   req.PrincipalID,
		ResourceType:  req.ResourceType,
		ResourceID:    req.ResourceID,
		Action:        req.Action,
		Duration:      req.Duration,
		Justification: req.Justification,
		Status:        RequestStatusPending,
		RequestedAt:   requestedAt,
		ExpiresAt:     expiresAt,
	}

	created, err := h.manager.CreateAccessRequest(accessReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// ListAccessRequests handles GET /api/v1/access-requests
func (h *RBACHandler) ListAccessRequests(c *gin.Context) {
	tenantID := c.Query("tenantId")
	principalID := c.Query("principalId")
	status := c.Query("status")

	requests, err := h.manager.ListAccessRequests(tenantID, principalID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, AccessRequestListResponse{AccessRequests: requests, Count: len(requests)})
}

// ApproveAccessRequest handles POST /api/v1/access-requests/:id/approve
func (h *RBACHandler) ApproveAccessRequest(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ApprovedBy string `json:"approvedBy" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: err.Error()})
		return
	}

	approved, err := h.manager.ApproveAccessRequest(id, req.ApprovedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, approved)
}

// RejectAccessRequest handles POST /api/v1/access-requests/:id/reject
func (h *RBACHandler) RejectAccessRequest(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		RejectedBy string `json:"rejectedBy" binding:"required"`
		Reason     string `json:"reason"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Error: err.Error()})
		return
	}

	rejected, err := h.manager.RejectAccessRequest(id, req.RejectedBy, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, rejected)
}

// RegisterRBACRoutes registers all RBAC routes
func RegisterRBACRoutes(router *gin.Engine, manager RBACManager) {
	handler := NewRBACHandler(manager)

	group := router.Group("/api/v1")
	{
		// Roles
		group.POST("/roles", handler.CreateRole)
		group.GET("/roles", handler.ListRoles)
		group.GET("/roles/:id", handler.GetRole)
		group.PATCH("/roles/:id", handler.UpdateRole)
		group.DELETE("/roles/:id", handler.DeleteRole)

		// Role Bindings
		group.POST("/role-bindings", handler.BindRole)
		group.GET("/role-bindings", handler.ListBindings)
		group.DELETE("/role-bindings/:id", handler.DeleteBinding)

		// Permissions
		group.GET("/permissions", handler.ListPermissions)
		group.POST("/permissions/check", handler.CheckPermission)

		// Access Requests
		group.POST("/access-requests", handler.CreateAccessRequest)
		group.GET("/access-requests", handler.ListAccessRequests)
		group.POST("/access-requests/:id/approve", handler.ApproveAccessRequest)
		group.POST("/access-requests/:id/reject", handler.RejectAccessRequest)
	}
}

// RBACManager interface
type RBACManager interface {
	CreateRole(role *Role) (*Role, error)
	GetRole(id string) (*Role, error)
	ListRoles(tenantID string) ([]*Role, error)
	UpdateRole(role *Role) (*Role, error)
	DeleteRole(id string) error
	CreateRoleBinding(binding *RoleBinding) (*RoleBinding, error)
	ListRoleBindings(tenantID, principalID string) ([]*RoleBinding, error)
	DeleteRoleBinding(id string) error
	CheckPermission(req *PermissionCheck) (*PermissionCheckResult, error)
	ListPermissions(tenantID, resource string) ([]*Permission, error)
	CreateAccessRequest(req *AccessRequest) (*AccessRequest, error)
	ListAccessRequests(tenantID, principalID, status string) ([]*AccessRequest, error)
	ApproveAccessRequest(id, approvedBy string) (*AccessRequest, error)
	RejectAccessRequest(id, rejectedBy, reason string) (*AccessRequest, error)
}
