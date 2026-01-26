package tenant

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TenantHandler handles tenant endpoints
type TenantHandler struct {
	manager TenantManager
}

// NewTenantHandler creates handler
func NewTenantHandler(manager TenantManager) *TenantHandler {
	return &TenantHandler{manager: manager}
}

// CreateTenant handles POST /api/v1/tenants
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Email     string `json:"email" binding:"required"`
		Owner     string `json:"owner" binding:"required"`
		Tier      TenantTier `json:"tier"`
		Isolation TenantIsolation `json:"isolation"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant := &Tenant{
		Name:       req.Name,
		Email:      req.Email,
		Status:     "Active",
		Tier:       req.Tier,
		Isolation:  req.Isolation,
		CreatedAt:  time.Now(),
	}

	created, err := h.manager.CreateTenant(tenant, req.Owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GetTenant handles GET /api/v1/tenants/:id
func (h *TenantHandler) GetTenant(c *gin.Context) {
	id := c.Param("id")
	tenant, err := h.manager.GetTenant(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// UpdateTenant handles PATCH /api/v1/tenants/:id
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}

	if err := c.BindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant, err := h.manager.GetTenant(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	tenant.UpdatedAt = time.Now()
	updated, err := h.manager.UpdateTenant(tenant)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// ListTenants handles GET /api/v1/tenants
func (h *TenantHandler) ListTenants(c *gin.Context) {
	tenants, err := h.manager.ListTenants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tenants": tenants, "count": len(tenants)})
}

// DeleteTenant handles DELETE /api/v1/tenants/:id
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	id := c.Param("id")
	if err := h.manager.DeleteTenant(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tenant deleted"})
}

// AddMember handles POST /api/v1/tenants/:id/members
func (h *TenantHandler) AddMember(c *gin.Context) {
	tenantID := c.Param("id")
	var req struct {
		UserID string      `json:"userId" binding:"required"`
		Role   MemberRole  `json:"role" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member := &TenantMember{
		TenantID: tenantID,
		UserID:   req.UserID,
		Role:     req.Role,
		Status:   "Active",
		JoinedAt: time.Now(),
	}

	if err := h.manager.AddMember(member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// RemoveMember handles DELETE /api/v1/tenants/:id/members/:userId
func (h *TenantHandler) RemoveMember(c *gin.Context) {
	tenantID := c.Param("id")
	userID := c.Param("userId")

	if err := h.manager.RemoveMember(tenantID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

// GetQuota handles GET /api/v1/tenants/:id/quota
func (h *TenantHandler) GetQuota(c *gin.Context) {
	id := c.Param("id")
	quota, err := h.manager.GetQuota(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quota not found"})
		return
	}

	c.JSON(http.StatusOK, quota)
}

// CheckQuota handles POST /api/v1/tenants/:id/quota/check
func (h *TenantHandler) CheckQuota(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Resource string `json:"resource" binding:"required"`
		Amount   int64  `json:"amount" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ok, err := h.manager.CheckQuota(id, req.Resource, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "quota exceeded"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"allowed": true})
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
