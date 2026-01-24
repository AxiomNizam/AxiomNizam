package apiserver

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"example.com/axiomnizam/internal/resources"
	"github.com/gin-gonic/gin"
)

// ResourceStore stores resources in memory with persistence
type ResourceStore struct {
	mu        sync.RWMutex
	resources map[string]map[string]resources.Resource // namespace -> name -> resource
	watchers  map[string][]ResourceWatcher             // namespace -> list of watchers
}

// ResourceWatcher watches for resource changes
type ResourceWatcher interface {
	OnAdd(resource resources.Resource)
	OnUpdate(oldResource, newResource resources.Resource)
	OnDelete(resource resources.Resource)
}

// NewResourceStore creates a new resource store
func NewResourceStore() *ResourceStore {
	return &ResourceStore{
		resources: make(map[string]map[string]resources.Resource),
		watchers:  make(map[string][]ResourceWatcher),
	}
}

// Create stores a new resource
func (rs *ResourceStore) Create(resource resources.Resource) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	meta := resource.GetObjectMeta()
	namespace := meta.Namespace
	if namespace == "" {
		namespace = "default"
	}

	if rs.resources[namespace] == nil {
		rs.resources[namespace] = make(map[string]resources.Resource)
	}

	if rs.resources[namespace][meta.Name] != nil {
		return fmt.Errorf("resource already exists: %s/%s", namespace, meta.Name)
	}

	rs.resources[namespace][meta.Name] = resource.DeepCopy()
	rs.notifyWatchers(namespace, WatchEventAdded, resource, nil)

	return nil
}

// Get retrieves a resource
func (rs *ResourceStore) Get(namespace, name string) (resources.Resource, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	if namespace == "" {
		namespace = "default"
	}

	if rs.resources[namespace] == nil {
		return nil, fmt.Errorf("resource not found: %s/%s", namespace, name)
	}

	resource := rs.resources[namespace][name]
	if resource == nil {
		return nil, fmt.Errorf("resource not found: %s/%s", namespace, name)
	}

	return resource.DeepCopy(), nil
}

// Update updates an existing resource
func (rs *ResourceStore) Update(resource resources.Resource) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	meta := resource.GetObjectMeta()
	namespace := meta.Namespace
	if namespace == "" {
		namespace = "default"
	}

	if rs.resources[namespace] == nil {
		return fmt.Errorf("resource not found: %s/%s", namespace, meta.Name)
	}

	oldResource := rs.resources[namespace][meta.Name]
	if oldResource == nil {
		return fmt.Errorf("resource not found: %s/%s", namespace, meta.Name)
	}

	rs.resources[namespace][meta.Name] = resource.DeepCopy()
	rs.notifyWatchers(namespace, WatchEventModified, resource, oldResource)

	return nil
}

// Delete removes a resource
func (rs *ResourceStore) Delete(namespace, name string) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if namespace == "" {
		namespace = "default"
	}

	if rs.resources[namespace] == nil {
		return fmt.Errorf("resource not found: %s/%s", namespace, name)
	}

	resource := rs.resources[namespace][name]
	if resource == nil {
		return fmt.Errorf("resource not found: %s/%s", namespace, name)
	}

	delete(rs.resources[namespace], name)
	rs.notifyWatchers(namespace, WatchEventDeleted, resource, nil)

	return nil
}

// List returns all resources in a namespace
func (rs *ResourceStore) List(namespace string, selector map[string]string) ([]resources.Resource, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	if namespace == "" {
		namespace = "default"
	}

	var result []resources.Resource

	if rs.resources[namespace] == nil {
		return result, nil
	}

	for _, resource := range rs.resources[namespace] {
		// Filter by labels if selector provided
		if len(selector) > 0 {
			if !resource.GetObjectMeta().MatchesLabels(selector) {
				continue
			}
		}
		result = append(result, resource.DeepCopy())
	}

	return result, nil
}

// Watch registers a watcher for resource changes
func (rs *ResourceStore) Watch(namespace string, watcher ResourceWatcher) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if namespace == "" {
		namespace = "default"
	}

	rs.watchers[namespace] = append(rs.watchers[namespace], watcher)
}

// notifyWatchers notifies all watchers of a change
func (rs *ResourceStore) notifyWatchers(namespace string, eventType WatchEventType, newResource, oldResource resources.Resource) {
	watchers := rs.watchers[namespace]

	for _, watcher := range watchers {
		switch eventType {
		case WatchEventAdded:
			go watcher.OnAdd(newResource)
		case WatchEventModified:
			go watcher.OnUpdate(oldResource, newResource)
		case WatchEventDeleted:
			go watcher.OnDelete(newResource)
		}
	}
}

// WatchEventType represents type of watch event
type WatchEventType string

const (
	WatchEventAdded    WatchEventType = "ADDED"
	WatchEventModified WatchEventType = "MODIFIED"
	WatchEventDeleted  WatchEventType = "DELETED"
)

// APIServer provides REST API for resources
type APIServer struct {
	store  *ResourceStore
	router *gin.Engine
}

// NewAPIServer creates a new API server
func NewAPIServer(store *ResourceStore) *APIServer {
	return &APIServer{
		store:  store,
		router: gin.New(),
	}
}

// RegisterRoutes registers all API routes
func (as *APIServer) RegisterRoutes() {
	api := as.router.Group("/api/v1")

	// Generic resource endpoints
	api.POST("/:namespace/:kind", as.CreateResource)
	api.GET("/:namespace/:kind/:name", as.GetResource)
	api.PUT("/:namespace/:kind/:name", as.UpdateResource)
	api.DELETE("/:namespace/:kind/:name", as.DeleteResource)
	api.GET("/:namespace/:kind", as.ListResources)

	// Status subresource
	api.GET("/:namespace/:kind/:name/status", as.GetResourceStatus)
	api.PUT("/:namespace/:kind/:name/status", as.UpdateResourceStatus)
}

// CreateResource creates a new resource
func (as *APIServer) CreateResource(c *gin.Context) {
	kind := c.Param("kind")
	namespace := c.Param("namespace")

	var resource resources.Resource

	// Unmarshal based on kind
	switch kind {
	case "workloads":
		var wr resources.WorkloadResource
		if err := c.ShouldBindJSON(&wr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		wr.ObjectMeta.Namespace = namespace
		resource = &wr

	case "pipelines":
		var pr resources.PipelineResource
		if err := c.ShouldBindJSON(&pr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		pr.ObjectMeta.Namespace = namespace
		resource = &pr

	case "schedules":
		var sr resources.ScheduleResource
		if err := c.ShouldBindJSON(&sr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		sr.ObjectMeta.Namespace = namespace
		resource = &sr

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown resource kind: %s", kind)})
		return
	}

	if err := as.store.Create(resource); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resource)
}

// GetResource gets a resource
func (as *APIServer) GetResource(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	resource, err := as.store.Get(namespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resource)
}

// UpdateResource updates a resource
func (as *APIServer) UpdateResource(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	kind := c.Param("kind")

	// Get existing resource (just to verify it exists)
	_, err := as.store.Get(namespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Parse new resource
	var resource resources.Resource
	switch kind {
	case "workloads":
		var wr resources.WorkloadResource
		if err := c.ShouldBindJSON(&wr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		wr.ObjectMeta.Namespace = namespace
		wr.ObjectMeta.Name = name
		wr.ObjectMeta.UpdatedAt = time.Now()
		wr.ObjectMeta.Generation++
		resource = &wr

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown resource kind: %s", kind)})
		return
	}

	if err := as.store.Update(resource); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resource)
}

// DeleteResource deletes a resource
func (as *APIServer) DeleteResource(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if err := as.store.Delete(namespace, name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ListResources lists resources
func (as *APIServer) ListResources(c *gin.Context) {
	namespace := c.Param("namespace")

	// Parse label selector
	selector := make(map[string]string)
	if labelSelector := c.Query("labelSelector"); labelSelector != "" {
		// Parse selector format: key1=value1,key2=value2
		for _, pair := range strings.Split(labelSelector, ",") {
			parts := strings.Split(pair, "=")
			if len(parts) == 2 {
				selector[parts[0]] = parts[1]
			}
		}
	}

	resources, err := as.store.List(namespace, selector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": resources})
}

// GetResourceStatus gets resource status
func (as *APIServer) GetResourceStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	resource, err := as.store.Get(namespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resource.GetStatus())
}

// UpdateResourceStatus updates resource status
func (as *APIServer) UpdateResourceStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	resource, err := as.store.Get(namespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var status resources.ObjectStatus
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status.LastTransitionTime = time.Now()
	resource.SetStatus(&status)

	if err := as.store.Update(resource); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resource.GetStatus())
}

// Run starts the API server
func (as *APIServer) Run(addr string) error {
	as.RegisterRoutes()
	return as.router.Run(addr)
}
