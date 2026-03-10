package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GenericResource represents a Kubernetes-style resource stored in the API server
type GenericResource struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   ResourceMetadata       `json:"metadata"`
	Spec       map[string]interface{} `json:"spec,omitempty"`
	Status     ResourceStatus         `json:"status,omitempty"`
}

// ResourceMetadata holds resource metadata
type ResourceMetadata struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	UID               string            `json:"uid,omitempty"`
	Generation        int64             `json:"generation,omitempty"`
	CreationTimestamp string            `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
}

// ResourceStatus holds resource status
type ResourceStatus struct {
	Phase      string                   `json:"phase,omitempty"`
	Conditions []map[string]interface{} `json:"conditions,omitempty"`
	Message    string                   `json:"message,omitempty"`
}

// ResourceHandler manages generic Kubernetes-style resources
type ResourceHandler struct {
	mu        sync.RWMutex
	resources map[string]map[string]map[string]*GenericResource // kind -> namespace -> name -> resource
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler() *ResourceHandler {
	return &ResourceHandler{
		resources: make(map[string]map[string]map[string]*GenericResource),
	}
}

func (h *ResourceHandler) ensureKind(kind string) {
	if h.resources[kind] == nil {
		h.resources[kind] = make(map[string]map[string]*GenericResource)
	}
}

func (h *ResourceHandler) ensureNamespace(kind, ns string) {
	h.ensureKind(kind)
	if h.resources[kind][ns] == nil {
		h.resources[kind][ns] = make(map[string]*GenericResource)
	}
}

// CreateOrUpdate creates or updates a resource (POST)
func (h *ResourceHandler) CreateOrUpdate(c *gin.Context) {
	kind := normalizeKind(c.Param("kind"))
	ns := c.Param("namespace")
	if ns == "" {
		ns = "default"
	}

	var resource GenericResource
	if err := c.ShouldBindJSON(&resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.ensureNamespace(kind, ns)

	name := resource.Metadata.Name
	if name == "" {
		// Try to extract from metadata
		c.JSON(http.StatusBadRequest, gin.H{"error": "metadata.name is required"})
		return
	}

	existing := h.resources[kind][ns][name]
	now := time.Now().UTC().Format(time.RFC3339)

	if existing != nil {
		// Update
		resource.Metadata.UID = existing.Metadata.UID
		resource.Metadata.CreationTimestamp = existing.Metadata.CreationTimestamp
		resource.Metadata.Generation = existing.Metadata.Generation + 1
		resource.Metadata.Namespace = ns
		resource.Status.Phase = "Reconciling"
		h.resources[kind][ns][name] = &resource
		c.JSON(http.StatusOK, &resource)
	} else {
		// Create
		resource.Metadata.UID = uuid.New().String()
		resource.Metadata.CreationTimestamp = now
		resource.Metadata.Generation = 1
		resource.Metadata.Namespace = ns
		resource.Status.Phase = "Pending"
		h.resources[kind][ns][name] = &resource
		c.JSON(http.StatusCreated, &resource)
	}

	// Simulate async reconciliation
	go func() {
		time.Sleep(500 * time.Millisecond)
		h.mu.Lock()
		defer h.mu.Unlock()
		if r, ok := h.resources[kind][ns][name]; ok {
			r.Status.Phase = "Ready"
			r.Status.Message = "Resource reconciled successfully"
		}
	}()
}

// Get retrieves a resource by namespace/kind/name
func (h *ResourceHandler) Get(c *gin.Context) {
	kind := normalizeKind(c.Param("kind"))
	ns := c.Param("namespace")
	name := c.Param("name")

	if ns == "" {
		ns = "default"
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.resources[kind] == nil || h.resources[kind][ns] == nil || h.resources[kind][ns][name] == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("resource not found: %s/%s/%s", kind, ns, name)})
		return
	}

	c.JSON(http.StatusOK, h.resources[kind][ns][name])
}

// List returns all resources of a kind in a namespace
func (h *ResourceHandler) List(c *gin.Context) {
	kind := normalizeKind(c.Param("kind"))
	ns := c.Param("namespace")

	if ns == "" {
		ns = "default"
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.resources[kind] == nil || h.resources[kind][ns] == nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	items := make([]*GenericResource, 0, len(h.resources[kind][ns]))
	for _, r := range h.resources[kind][ns] {
		items = append(items, r)
	}

	c.JSON(http.StatusOK, items)
}

// Update updates a resource (PUT)
func (h *ResourceHandler) Update(c *gin.Context) {
	kind := normalizeKind(c.Param("kind"))
	ns := c.Param("namespace")
	name := c.Param("name")

	if ns == "" {
		ns = "default"
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.resources[kind] == nil || h.resources[kind][ns] == nil || h.resources[kind][ns][name] == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("resource not found: %s/%s/%s", kind, ns, name)})
		return
	}

	existing := h.resources[kind][ns][name]

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Merge updates into spec
	if existing.Spec == nil {
		existing.Spec = make(map[string]interface{})
	}
	for k, v := range updates {
		existing.Spec[k] = v
	}
	existing.Metadata.Generation++
	existing.Status.Phase = "Reconciling"

	c.JSON(http.StatusOK, existing)

	// Simulate reconciliation
	go func() {
		time.Sleep(500 * time.Millisecond)
		h.mu.Lock()
		defer h.mu.Unlock()
		if r, ok := h.resources[kind][ns][name]; ok {
			r.Status.Phase = "Ready"
		}
	}()
}

// Delete removes a resource
func (h *ResourceHandler) Delete(c *gin.Context) {
	kind := normalizeKind(c.Param("kind"))
	ns := c.Param("namespace")
	name := c.Param("name")

	if ns == "" {
		ns = "default"
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.resources[kind] == nil || h.resources[kind][ns] == nil || h.resources[kind][ns][name] == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("resource not found: %s/%s/%s", kind, ns, name)})
		return
	}

	delete(h.resources[kind][ns], name)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s '%s' deleted", kind, name)})
}

// GetStatus returns the status subresource
func (h *ResourceHandler) GetStatus(c *gin.Context) {
	kind := normalizeKind(c.Param("kind"))
	ns := c.Param("namespace")
	name := c.Param("name")

	if ns == "" {
		ns = "default"
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.resources[kind] == nil || h.resources[kind][ns] == nil || h.resources[kind][ns][name] == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
		return
	}

	c.JSON(http.StatusOK, h.resources[kind][ns][name].Status)
}

// ListAll returns all resources of a kind across all namespaces (for non-namespaced queries)
func (h *ResourceHandler) ListAll(c *gin.Context) {
	kind := normalizeKind(c.Param("kind"))

	h.mu.RLock()
	defer h.mu.RUnlock()

	items := make([]*GenericResource, 0)

	if h.resources[kind] != nil {
		for _, nsResources := range h.resources[kind] {
			for _, r := range nsResources {
				items = append(items, r)
			}
		}
	}

	c.JSON(http.StatusOK, items)
}

// Events returns events for a resource
func (h *ResourceHandler) Events(c *gin.Context) {
	kind := normalizeKind(c.Param("kind"))
	ns := c.Param("namespace")
	name := c.Param("name")

	if ns == "" {
		ns = "default"
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.resources[kind] == nil || h.resources[kind][ns] == nil || h.resources[kind][ns][name] == nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	r := h.resources[kind][ns][name]
	events := []map[string]interface{}{
		{
			"type":    "Normal",
			"reason":  "Created",
			"message": fmt.Sprintf("%s '%s' was created", r.Kind, name),
			"age":     r.Metadata.CreationTimestamp,
		},
		{
			"type":    "Normal",
			"reason":  "Reconciled",
			"message": fmt.Sprintf("%s '%s' reconciled to generation %d", r.Kind, name, r.Metadata.Generation),
			"age":     time.Now().UTC().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, events)
}

// normalizeKind normalizes resource kind names for consistent storage
func normalizeKind(kind string) string {
	kind = strings.ToLower(kind)
	// Map plural to singular for consistency
	switch kind {
	case "apis", "api":
		return "apis"
	case "policies", "policy":
		return "policies"
	case "workflows", "workflow":
		return "workflows"
	case "datasources", "datasource":
		return "datasources"
	case "jobs", "job":
		return "jobs"
	default:
		return kind
	}
}
