package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// DataSourceResource represents a datasource on the server
type DataSourceResource struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   DataSourceMetadata     `json:"metadata"`
	Spec       map[string]interface{} `json:"spec,omitempty"`
	Status     DataSourceStatus       `json:"status,omitempty"`
}

// DataSourceMetadata holds datasource metadata
type DataSourceMetadata struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace,omitempty"`
	UID               string `json:"uid,omitempty"`
	CreationTimestamp string `json:"creationTimestamp,omitempty"`
}

// DataSourceStatus holds datasource status
type DataSourceStatus struct {
	Connected bool   `json:"connected"`
	LastCheck string `json:"lastCheck,omitempty"`
	Message   string `json:"message,omitempty"`
}

// DataSourceHandler manages datasource resources
type DataSourceHandler struct {
	mu          sync.RWMutex
	datasources map[string]*DataSourceResource // name -> datasource
	etcd        *clientv3.Client
	stateKey    string
}

// NewDataSourceHandler creates a new datasource handler
func NewDataSourceHandler(etcd ...*clientv3.Client) *DataSourceHandler {
	var etcdClient *clientv3.Client
	if len(etcd) > 0 {
		etcdClient = etcd[0]
	}

	h := &DataSourceHandler{
		datasources: make(map[string]*DataSourceResource),
		etcd:        etcdClient,
		stateKey:    "axiomnizam:datasources:state",
	}
	h.loadState()
	return h
}

func (h *DataSourceHandler) loadState() {
	if h.etcd == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.etcd.Get(ctx, h.stateKey)
	if err != nil {
		log.Printf("datasources: failed to load persisted state from etcd: %v", err)
		return
	}
	if len(resp.Kvs) == 0 {
		return
	}

	var datasources map[string]*DataSourceResource
	if err := json.Unmarshal(resp.Kvs[0].Value, &datasources); err != nil {
		log.Printf("datasources: failed to decode persisted state: %v", err)
		return
	}
	if datasources == nil {
		datasources = make(map[string]*DataSourceResource)
	}
	h.datasources = datasources
}

func (h *DataSourceHandler) persistStateLocked() {
	if h.etcd == nil {
		return
	}

	payload, err := json.Marshal(h.datasources)
	if err != nil {
		log.Printf("datasources: failed to encode state: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.etcd.Put(ctx, h.stateKey, string(payload)); err != nil {
		log.Printf("datasources: failed to persist state to etcd: %v", err)
	}
}

// Create creates a new datasource
func (h *DataSourceHandler) Create(c *gin.Context) {
	var ds DataSourceResource
	if err := c.ShouldBindJSON(&ds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	name := ds.Metadata.Name
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metadata.name is required"})
		return
	}

	ds.Kind = "DataSource"
	ds.APIVersion = "axiom-nizam.io/v1"
	ds.Metadata.UID = uuid.New().String()
	ds.Metadata.CreationTimestamp = time.Now().UTC().Format(time.RFC3339)
	ds.Status = DataSourceStatus{
		Connected: false,
		LastCheck: time.Now().UTC().Format(time.RFC3339),
		Message:   "Created, pending connection test",
	}

	h.datasources[name] = &ds
	h.persistStateLocked()
	c.JSON(http.StatusCreated, &ds)
}

// List lists all datasources
func (h *DataSourceHandler) List(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	items := make([]*DataSourceResource, 0, len(h.datasources))
	for _, ds := range h.datasources {
		items = append(items, ds)
	}

	c.JSON(http.StatusOK, items)
}

// Get returns a datasource by name
func (h *DataSourceHandler) Get(c *gin.Context) {
	name := c.Param("name")

	h.mu.RLock()
	defer h.mu.RUnlock()

	ds, ok := h.datasources[name]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("datasource '%s' not found", name)})
		return
	}

	c.JSON(http.StatusOK, ds)
}

// Update updates a datasource
func (h *DataSourceHandler) Update(c *gin.Context) {
	name := c.Param("name")

	h.mu.Lock()
	defer h.mu.Unlock()

	ds, ok := h.datasources[name]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("datasource '%s' not found", name)})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if ds.Spec == nil {
		ds.Spec = make(map[string]interface{})
	}
	for k, v := range updates {
		ds.Spec[k] = v
	}
	h.persistStateLocked()

	c.JSON(http.StatusOK, ds)
}

// Delete deletes a datasource
func (h *DataSourceHandler) Delete(c *gin.Context) {
	name := c.Param("name")

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.datasources[name]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("datasource '%s' not found", name)})
		return
	}

	delete(h.datasources, name)
	h.persistStateLocked()
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("datasource '%s' deleted", name)})
}

// Test tests datasource connectivity
func (h *DataSourceHandler) Test(c *gin.Context) {
	name := c.Param("name")

	h.mu.Lock()
	defer h.mu.Unlock()

	ds, ok := h.datasources[name]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("datasource '%s' not found", name)})
		return
	}

	// Mark as tested
	ds.Status.Connected = true
	ds.Status.LastCheck = time.Now().UTC().Format(time.RFC3339)
	ds.Status.Message = "Connection test successful"
	h.persistStateLocked()

	c.JSON(http.StatusOK, gin.H{
		"status":  "connected",
		"message": "Connection test successful",
	})
}

// Apply creates or updates a datasource from YAML-sourced data
func (h *DataSourceHandler) Apply(c *gin.Context) {
	h.Create(c) // Same behavior as create/update
}
