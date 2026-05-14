package handlers

// Store-backed DataSource handler (Phase 3/4 reference implementation).
//
// This file demonstrates the thin-handler pattern prescribed by the
// Phase 3/4 checklist:
//
//   handler → store.ResourceStore[T] → informer → workqueue → reconciler
//
// The legacy `DataSourceHandler` in `datasource_handler.go` holds raw
// `*clientv3.Client`, an in-memory map, and a `sync.RWMutex`.  The new
// `DataSourceV1Handler` here owns NONE of that — it is a pure HTTP
// shim over `ResourceStore[*DataSourceV1Resource]`.  All persistence,
// watch fan-out, and concurrency concerns are delegated to the store;
// all runtime behaviour is delegated to the reconciler.
//
// Routes registered by this handler:
//
//   POST   /api/v2/datasources            -> Create
//   GET    /api/v2/datasources            -> List
//   GET    /api/v2/datasources/:name      -> Get
//   PUT    /api/v2/datasources/:name      -> Update
//   DELETE /api/v2/datasources/:name      -> Delete
//
// Once the legacy handler has been fully migrated away, the v2 routes
// become the sole surface and the old file can be deleted.

import (
	"errors"
	"net/http"
	"time"

	datasourceres "example.com/axiomnizam/internal/datasource"
	"example.com/axiomnizam/internal/platform/store"
	"example.com/axiomnizam/internal/resources"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DataSourceV1Handler is a thin Gin handler over a
// ResourceStore[*DataSourceV1Resource].  Construct it with any store
// implementation (in-memory for tests, EtcdStore[T] in production).
type DataSourceV1Handler struct {
	store store.ResourceStore[*datasourceres.DataSourceV1Resource]
}

// NewDataSourceV1Handler builds the handler.  `store` must be non-nil.
func NewDataSourceV1Handler(s store.ResourceStore[*datasourceres.DataSourceV1Resource]) *DataSourceV1Handler {
	return &DataSourceV1Handler{store: s}
}

// RegisterRoutes attaches the v2 datasource routes onto the given
// router group.
func (h *DataSourceV1Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/datasources", h.Create)
	rg.GET("/datasources", h.List)
	rg.GET("/datasources/:name", h.Get)
	rg.PUT("/datasources/:name", h.Update)
	rg.DELETE("/datasources/:name", h.Delete)
}

// Create stores a new datasource resource.  The handler performs no
// persistence itself — Create calls straight through to the store,
// which emits a watch event that the informer / workqueue /
// reconciler pipeline consumes.
func (h *DataSourceV1Handler) Create(c *gin.Context) {
	var in datasourceres.DataSourceV1Resource
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metadata.name is required"})
		return
	}

	// Server-owned metadata.
	in.TypeMeta = resources.TypeMeta{
		Kind:       datasourceres.DataSourceKind,
		APIVersion: datasourceres.DataSourceAPIVersion,
	}
	if in.UID == "" {
		in.UID = uuid.New().String()
	}
	now := time.Now()
	in.CreatedAt = now
	in.UpdatedAt = now
	in.Generation = 1
	in.Status.Phase = "Pending"
	in.Status.LastTransitionTime = now

	if err := h.store.Create(c.Request.Context(), &in); err != nil {
		if errors.Is(err, store.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, &in)
}

// List returns every datasource resource.
func (h *DataSourceV1Handler) List(c *gin.Context) {
	items, err := h.store.List(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Get fetches a datasource by name.
func (h *DataSourceV1Handler) Get(c *gin.Context) {
	ds, err := h.store.Get(c.Request.Context(), c.Param("name"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ds)
}

// Update replaces the spec of an existing datasource.  Status is
// controller-owned and is preserved across updates; the handler only
// bumps Generation so the reconciler notices.
func (h *DataSourceV1Handler) Update(c *gin.Context) {
	name := c.Param("name")
	existing, err := h.store.Get(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var patch datasourceres.DataSourceV1Resource
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing.Spec = patch.Spec
	existing.UpdatedAt = time.Now()
	existing.Generation++

	if err := h.store.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// Delete removes a datasource resource.
func (h *DataSourceV1Handler) Delete(c *gin.Context) {
	name := c.Param("name")
	if err := h.store.Delete(c.Request.Context(), name); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "datasource '" + name + "' deleted"})
}
