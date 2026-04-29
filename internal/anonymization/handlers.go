package anonymization

import (
	"net/http"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/platform/validate"
	"github.com/gin-gonic/gin"
)

type AnonymizationHandlers struct {
	store store.ResourceStore[*AnonymizationPolicyResource]
}

func NewAnonymizationHandlers(s store.ResourceStore[*AnonymizationPolicyResource]) *AnonymizationHandlers {
	return &AnonymizationHandlers{store: s}
}

func (h *AnonymizationHandlers) RegisterRoutes(rg *gin.RouterGroup) {
	anon := rg.Group("/anonymization")
	{
		anon.GET("/policies", h.ListPolicies)
		anon.GET("/policies/:name", h.GetPolicy)
		anon.POST("/policies", h.CreatePolicy)
		anon.PUT("/policies/:name", h.UpdatePolicy)
		anon.DELETE("/policies/:name", h.DeletePolicy)
		anon.POST("/policies/:name/run", h.TriggerRun)
	}
}

func (h *AnonymizationHandlers) ListPolicies(c *gin.Context) {
	policies, err := h.store.List(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"policies": policies, "count": len(policies)})
}

func (h *AnonymizationHandlers) GetPolicy(c *gin.Context) {
	name := validate.PathParam(c, "name")
	if name == "" {
		return
	}
	policy, err := h.store.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found", "name": name})
		return
	}
	c.JSON(http.StatusOK, policy)
}

func (h *AnonymizationHandlers) CreatePolicy(c *gin.Context) {
	var policy AnonymizationPolicyResource
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	policy.Kind = AnonymizationPolicyKind
	policy.APIVersion = AnonymizationPolicyAPIVersion
	now := time.Now()
	policy.CreatedAt = now
	policy.Generation = 1
	policy.Status.Phase = "Pending"
	if err := h.store.Create(c.Request.Context(), &policy); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, policy)
}

func (h *AnonymizationHandlers) UpdatePolicy(c *gin.Context) {
	name := validate.PathParam(c, "name")
	if name == "" {
		return
	}
	existing, err := h.store.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found", "name": name})
		return
	}
	var updated AnonymizationPolicyResource
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated.ObjectMeta = existing.ObjectMeta
	updated.Generation = existing.Generation + 1
	updated.Status = existing.Status
	if err := h.store.Update(c.Request.Context(), &updated); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *AnonymizationHandlers) DeletePolicy(c *gin.Context) {
	name := validate.PathParam(c, "name")
	if name == "" {
		return
	}
	if err := h.store.Delete(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found", "name": name})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": name})
}

func (h *AnonymizationHandlers) TriggerRun(c *gin.Context) {
	name := validate.PathParam(c, "name")
	if name == "" {
		return
	}
	policy, err := h.store.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found", "name": name})
		return
	}
	policy.Generation++
	_ = h.store.Update(c.Request.Context(), policy)
	c.JSON(http.StatusAccepted, gin.H{"message": "anonymization run triggered", "policy": name})
}
