package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"example.com/axiomnizam/internal/logging"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// GenericResource is the wire-format DTO served by the /api/v1/resources
// endpoints. It is intentionally JSON-stable (clients rely on this shape via
// the Postman collections) and acts as a thin projection of the canonical
// resources.BaseResource used elsewhere in the control plane.
//
// Invariants (P0.3):
//   - This type DOES NOT drive reconciliation. Status transitions must be
//     performed by a real controller reading from the work queue, not by a
//     fake `go func(){ sleep; Ready }` simulation inside the HTTP handler.
//   - GenericResource satisfies reconciler.Resource via GetKey / GetGeneration
//     / GetObservedGeneration so it can be enqueued for reconciliation.
type GenericResource struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   ResourceMetadata       `json:"metadata"`
	Spec       map[string]interface{} `json:"spec,omitempty"`
	Status     ResourceStatus         `json:"status,omitempty"`
}

// GetKey implements reconciler.Resource. Returns "namespace/name".
func (g *GenericResource) GetKey() string {
	if g == nil {
		return ""
	}
	if g.Metadata.Namespace == "" {
		return g.Metadata.Name
	}
	return g.Metadata.Namespace + "/" + g.Metadata.Name
}

// GetGeneration implements reconciler.Resource.
func (g *GenericResource) GetGeneration() int64 {
	if g == nil {
		return 0
	}
	return g.Metadata.Generation
}

// GetObservedGeneration implements reconciler.Resource.
func (g *GenericResource) GetObservedGeneration() int64 {
	if g == nil {
		return 0
	}
	return g.Status.ObservedGeneration
}

// ResourceMetadata holds resource metadata. The field set and JSON tags are
// kept backward-compatible with pre-P0 clients.
type ResourceMetadata struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	UID               string            `json:"uid,omitempty"`
	Generation        int64             `json:"generation,omitempty"`
	CreationTimestamp string            `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
}

// ResourceStatus holds resource status. ObservedGeneration was added in P0.1
// so the HTTP-surface DTO participates in generation-based reconciliation.
type ResourceStatus struct {
	Phase              string                   `json:"phase,omitempty"`
	Conditions         []map[string]interface{} `json:"conditions,omitempty"`
	Message            string                   `json:"message,omitempty"`
	ObservedGeneration int64                    `json:"observedGeneration,omitempty"`
}

// ResourceHandler manages generic Kubernetes-style resources
type ResourceHandler struct {
	mu        sync.RWMutex
	resources map[string]map[string]map[string]*GenericResource // kind -> namespace -> name -> resource
	etcd      *clientv3.Client
	stateKey  string
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler(etcd *clientv3.Client) *ResourceHandler {
	h := &ResourceHandler{
		resources: make(map[string]map[string]map[string]*GenericResource),
		etcd:      etcd,
		stateKey:  "axiomnizam:resources:state",
	}
	h.loadState()
	return h
}

func (h *ResourceHandler) loadState() {
	if h.etcd == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.etcd.Get(ctx, h.stateKey)
	if err != nil {
		logging.Z().Warn("resources: failed to load persisted state", zap.Error(err))
		return
	}
	if len(resp.Kvs) == 0 {
		return
	}

	var resourcesState map[string]map[string]map[string]*GenericResource
	if err := json.Unmarshal(resp.Kvs[0].Value, &resourcesState); err != nil {
		logging.Z().Warn("resources: failed to decode persisted state", zap.Error(err))
		return
	}
	if resourcesState == nil {
		resourcesState = make(map[string]map[string]map[string]*GenericResource)
	}
	h.resources = resourcesState
}

func (h *ResourceHandler) persistStateLocked() {
	if h.etcd == nil {
		return
	}

	payload, err := json.Marshal(h.resources)
	if err != nil {
		logging.Z().Warn("resources: failed to encode state", zap.Error(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.etcd.Put(ctx, h.stateKey, string(payload)); err != nil {
		logging.Z().Warn("resources: failed to persist state", zap.Error(err))
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
		h.persistStateLocked()
		c.JSON(http.StatusOK, &resource)
	} else {
		// Create
		resource.Metadata.UID = uuid.New().String()
		resource.Metadata.CreationTimestamp = now
		resource.Metadata.Generation = 1
		resource.Metadata.Namespace = ns
		resource.Status.Phase = "Pending"
		h.resources[kind][ns][name] = &resource
		h.persistStateLocked()
		c.JSON(http.StatusCreated, &resource)
	}

	// Real status transitions (Pending -> Reconciling -> Ready) are performed
	// by the controller that owns this Kind, not by the HTTP handler. See
	// P0.3 in ARCHITECTURE.md.
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
	h.persistStateLocked()

	c.JSON(http.StatusOK, existing)

	// Status transitions are owned by the controller for this Kind (P0.3).
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
	h.persistStateLocked()
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

// FindResourceByKindAndName finds the first matching resource across namespaces.
// It returns a defensive copy to avoid callers mutating internal handler state.
func (h *ResourceHandler) FindResourceByKindAndName(kind, name string) (*GenericResource, bool) {
	normalizedKind := normalizeKind(kind)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for storedKind, byNamespace := range h.resources {
		for _, byName := range byNamespace {
			for _, res := range byName {
				if !strings.EqualFold(res.Metadata.Name, name) {
					continue
				}

				if normalizedKind != "" {
					if !(strings.EqualFold(storedKind, normalizedKind) ||
						strings.EqualFold(res.Kind, normalizedKind) ||
						strings.EqualFold(res.Kind, kind)) {
						continue
					}
				}

				return cloneGenericResource(res), true
			}
		}
	}

	return nil, false
}

func cloneGenericResource(in *GenericResource) *GenericResource {
	if in == nil {
		return nil
	}

	out := *in
	out.Metadata.Labels = cloneStringMap(in.Metadata.Labels)
	out.Metadata.Annotations = cloneStringMap(in.Metadata.Annotations)
	out.Spec = cloneAnyMap(in.Spec)

	if len(in.Status.Conditions) > 0 {
		out.Status.Conditions = make([]map[string]interface{}, len(in.Status.Conditions))
		for i := range in.Status.Conditions {
			out.Status.Conditions[i] = cloneAnyMap(in.Status.Conditions[i])
		}
	}

	return &out
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}

	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneAnyMap(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return nil
	}

	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
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
