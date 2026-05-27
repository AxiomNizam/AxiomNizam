package tenant

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	manager        TenantManager
	dualWriteStore TenantDualWriteStore
}

// NewTenantHandler creates handler
func NewTenantHandler(manager TenantManager) *TenantHandler {
	return &TenantHandler{manager: manager}
}

// CreateTenant handles POST /api/v1/tenants
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: err.Error()})
		return
	}

	// Phase 3: reconciler-authoritative path
	if h.isAuthoritative() {
		t := &Tenant{Name: req.Name, DisplayName: req.DisplayName, Owner: req.Owner, Tier: req.Tier}
		resource := h.buildTenantResource(t)
		if h.dualWriteStore != nil {
			if err := h.dualWriteStore.Create(c.Request.Context(), resource); err != nil {
				_ = h.dualWriteStore.Update(c.Request.Context(), resource)
			}
		}
		c.JSON(http.StatusAccepted, TenantCreatedResponse{Name: resource.Name, Status: "Pending", Message: "tenant resource created"})
		return
	}

	tenant := &Tenant{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Tier:           req.Tier,
		IsolationLevel: req.IsolationLevel,
		CreatedAt:      time.Now(),
	}

	created, err := h.manager.CreateTenant(c.Request.Context(), tenant, req.Owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	h.dualWriteTenant(created)
	c.JSON(http.StatusCreated, TenantToResponse(created))
}

// GetTenant handles GET /api/v1/tenants/:id
func (h *TenantHandler) GetTenant(c *gin.Context) {
	id := c.Param("id")
	tenant, err := h.manager.GetTenant(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Message: "tenant not found"})
		return
	}

	c.JSON(http.StatusOK, TenantToResponse(tenant))
}

// UpdateTenant handles PATCH /api/v1/tenants/:id
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}

	if err := c.BindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: err.Error()})
		return
	}

	tenant, err := h.manager.GetTenant(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Message: "tenant not found"})
		return
	}

	tenant.UpdatedAt = time.Now()
	if err := h.manager.UpdateTenant(c.Request.Context(), tenant); err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, TenantToResponse(tenant))
}

// ListTenants handles GET /api/v1/tenants
func (h *TenantHandler) ListTenants(c *gin.Context) {
	ownerID := c.Query("owner")
	tenants, err := h.manager.ListTenants(c.Request.Context(), ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, TenantListResponse{Tenants: tenants, Count: len(tenants)})
}

// DeleteTenant handles DELETE /api/v1/tenants/:id
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.DeleteTenant(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "tenant deleted"})
}

// AddMember handles POST /api/v1/tenants/:id/members
func (h *TenantHandler) AddMember(c *gin.Context) {
	tenantID := c.Param("id")
	var req AddMemberRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: err.Error()})
		return
	}

	member, err := h.manager.AddMember(c.Request.Context(), tenantID, req.UserID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// RemoveMember handles DELETE /api/v1/tenants/:id/members/:userId
func (h *TenantHandler) RemoveMember(c *gin.Context) {
	tenantID := c.Param("id")
	userID := c.Param("userId")

	if err := h.manager.RemoveMember(c.Request.Context(), tenantID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "member removed"})
}

// GetQuota handles GET /api/v1/tenants/:id/quota
func (h *TenantHandler) GetQuota(c *gin.Context) {
	id := c.Param("id")
	quota, err := h.manager.GetQuota(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, MessageResponse{Message: "quota not found"})
		return
	}

	c.JSON(http.StatusOK, quota)
}

// CheckQuota handles POST /api/v1/tenants/:id/quota/check
func (h *TenantHandler) CheckQuota(c *gin.Context) {
	id := c.Param("id")
	var req CheckQuotaRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{Message: err.Error()})
		return
	}

	if err := h.manager.CheckQuota(c.Request.Context(), id, req.Resource, req.Amount); err != nil {
		c.JSON(http.StatusForbidden, MessageResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, QuotaCheckResponse{Allowed: true})
}

// RegisterTenantRoutes registers all tenant routes
func RegisterTenantRoutes(router *gin.Engine, manager TenantManager) {
	handler := NewTenantHandler(manager)

	group := router.Group("/api/v1/tenants")
	{
		group.POST("", handler.CreateTenant)
		group.GET("", handler.ListTenants)
		group.GET("/:id", handler.GetTenant)
		group.PATCH("/:id", handler.UpdateTenant)
		group.DELETE("/:id", handler.DeleteTenant)
		group.POST("/:id/members", handler.AddMember)
		group.DELETE("/:id/members/:userId", handler.RemoveMember)
		group.GET("/:id/quota", handler.GetQuota)
		group.POST("/:id/quota/check", handler.CheckQuota)
	}
}
