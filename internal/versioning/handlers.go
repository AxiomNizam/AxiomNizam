package versioning

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// VersionHandler handles versioning endpoints
type VersionHandler struct {
	manager VersionManager
}

// NewVersionHandler creates handler
func NewVersionHandler(manager VersionManager) *VersionHandler {
	return &VersionHandler{manager: manager}
}

// GetVersion handles GET /api/v1/versions/:resourceType/:resourceId/:version
func (h *VersionHandler) GetVersion(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")
	versionStr := c.Param("version")

	version, _ := strconv.ParseInt(versionStr, 10, 64)
	rv, err := h.manager.GetVersion(resourceType, resourceID, version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	c.JSON(http.StatusOK, rv)
}

// ListVersions handles GET /api/v1/versions/:resourceType/:resourceId
func (h *VersionHandler) ListVersions(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")

	versions, err := h.manager.ListVersions(resourceType, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions, "count": len(versions)})
}

// GetHistory handles GET /api/v1/history/:resourceType/:resourceId
func (h *VersionHandler) GetHistory(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")

	history, err := h.manager.GetHistory(resourceType, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetDiff handles GET /api/v1/diff/:resourceType/:resourceId
func (h *VersionHandler) GetDiff(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	from, _ := strconv.ParseInt(fromStr, 10, 64)
	to, _ := strconv.ParseInt(toStr, 10, 64)

	diff, err := h.manager.GetDiff(resourceType, resourceID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, diff)
}

// CreateSnapshot handles POST /api/v1/snapshots/:resourceType/:resourceId
func (h *VersionHandler) CreateSnapshot(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	snapshot := &Snapshot{
		Name:         req.Name,
		Description:  req.Description,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Tags:         req.Tags,
		CreatedAt:    time.Now(),
	}

	created, err := h.manager.CreateSnapshot(snapshot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// Rollback handles POST /api/v1/versions/:resourceType/:resourceId/rollback
func (h *VersionHandler) Rollback(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")

	var req struct {
		TargetVersion int64  `json:"targetVersion" binding:"required"`
		Reason        string `json:"reason"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.manager.Rollback(resourceType, resourceID, req.TargetVersion, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RegisterVersioningRoutes registers all versioning routes
func RegisterVersioningRoutes(router *gin.Engine, manager VersionManager) {
	handler := NewVersionHandler(manager)

	group := router.Group("/api/v1")
	{
		group.GET("/versions/:resourceType/:resourceId/:version", handler.GetVersion)
		group.GET("/versions/:resourceType/:resourceId", handler.ListVersions)
		group.GET("/history/:resourceType/:resourceId", handler.GetHistory)
		group.GET("/diff/:resourceType/:resourceId", handler.GetDiff)
		group.POST("/snapshots/:resourceType/:resourceId", handler.CreateSnapshot)
		group.POST("/versions/:resourceType/:resourceId/rollback", handler.Rollback)
	}
}

// VersionManager interface
type VersionManager interface {
	GetVersion(resourceType, resourceID string, version int64) (*ResourceVersion, error)
	ListVersions(resourceType, resourceID string) ([]*ResourceVersion, error)
	GetHistory(resourceType, resourceID string) (*VersionHistory, error)
	GetDiff(resourceType, resourceID string, from, to int64) (*VersionDiff, error)
	CreateSnapshot(snapshot *Snapshot) (*Snapshot, error)
	Rollback(resourceType, resourceID string, targetVersion int64, reason string) (*RollbackResult, error)
}
