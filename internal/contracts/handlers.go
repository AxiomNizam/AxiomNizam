package contracts

// =====================================================
// WS-2.2 — Data Contracts REST API Handlers
//
// Provides CRUD operations for data contracts, plus validation
// and diff endpoints.
// =====================================================

import (
	"net/http"
	"time"

	"example.com/axiomnizam/internal/platform/store"
	"github.com/gin-gonic/gin"
)

// ContractHandlers provides REST API handlers for data contracts.
type ContractHandlers struct {
	store store.ResourceStore[*DataContractResource]
}

// NewContractHandlers creates new handlers.
func NewContractHandlers(s store.ResourceStore[*DataContractResource]) *ContractHandlers {
	return &ContractHandlers{store: s}
}

// RegisterRoutes registers contract API routes.
func (h *ContractHandlers) RegisterRoutes(rg *gin.RouterGroup) {
	contracts := rg.Group("/contracts")
	{
		contracts.GET("", h.ListContracts)
		contracts.GET("/:name", h.GetContract)
		contracts.POST("", h.CreateContract)
		contracts.PUT("/:name", h.UpdateContract)
		contracts.DELETE("/:name", h.DeleteContract)
		contracts.GET("/:name/violations", h.GetViolations)
		contracts.POST("/:name/validate", h.ValidateContract)
	}
}

// ListContracts returns all data contracts, optionally filtered.
func (h *ContractHandlers) ListContracts(c *gin.Context) {
	contracts, err := h.store.List(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Apply filters.
	producerFilter := c.Query("producer")
	assetFilter := c.Query("asset")
	statusFilter := c.Query("status")

	var filtered []*DataContractResource
	for _, contract := range contracts {
		if producerFilter != "" && contract.Spec.Producer != producerFilter {
			continue
		}
		if assetFilter != "" && contract.Spec.AssetRef != assetFilter {
			continue
		}
		if statusFilter == "compliant" && !contract.Status.Compliant {
			continue
		}
		if statusFilter == "violated" && contract.Status.Compliant {
			continue
		}
		filtered = append(filtered, contract)
	}

	c.JSON(http.StatusOK, gin.H{
		"contracts": filtered,
		"count":     len(filtered),
	})
}

// GetContract returns a single contract by name.
func (h *ContractHandlers) GetContract(c *gin.Context) {
	name := c.Param("name")
	contract, err := h.store.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found", "name": name})
		return
	}
	c.JSON(http.StatusOK, contract)
}

// CreateContract creates a new data contract.
func (h *ContractHandlers) CreateContract(c *gin.Context) {
	var contract DataContractResource
	if err := c.ShouldBindJSON(&contract); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults.
	contract.Kind = DataContractKind
	contract.APIVersion = DataContractAPIVersion
	if contract.Spec.Compatibility == "" {
		contract.Spec.Compatibility = CompatBackward
	}
	if !contract.Spec.Enabled {
		contract.Spec.Enabled = true
	}
	now := time.Now()
	contract.CreatedAt = now
	contract.Generation = 1
	contract.Status.Phase = "Pending"

	if err := h.store.Create(c.Request.Context(), &contract); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, contract)
}

// UpdateContract updates an existing data contract.
func (h *ContractHandlers) UpdateContract(c *gin.Context) {
	name := c.Param("name")
	existing, err := h.store.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found", "name": name})
		return
	}

	var updated DataContractResource
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

// DeleteContract deletes a data contract.
func (h *ContractHandlers) DeleteContract(c *gin.Context) {
	name := c.Param("name")
	if err := h.store.Delete(c.Request.Context(), name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found", "name": name})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": name})
}

// GetViolations returns current violations for a contract.
func (h *ContractHandlers) GetViolations(c *gin.Context) {
	name := c.Param("name")
	contract, err := h.store.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found", "name": name})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contract":   name,
		"compliant":  contract.Status.Compliant,
		"violations": contract.Status.Violations,
		"validatedAt": contract.Status.LastValidatedAt,
	})
}

// ValidateContract triggers an immediate validation of the contract.
func (h *ContractHandlers) ValidateContract(c *gin.Context) {
	name := c.Param("name")
	contract, err := h.store.Get(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found", "name": name})
		return
	}

	// Bump generation to trigger reconciler re-evaluation.
	contract.Generation++
	if err := h.store.Update(c.Request.Context(), contract); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":  "validation triggered",
		"contract": name,
	})
}
