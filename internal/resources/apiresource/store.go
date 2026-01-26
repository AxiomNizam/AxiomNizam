package apiresource

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Store manages APIResource persistence (in-memory for now)
type Store struct {
	mu        sync.RWMutex
	resources map[string]*APIResource // key = namespace/name
	versions  map[string]int64        // track versions
}

// NewStore creates a new storage backend
func NewStore() *Store {
	return &Store{
		resources: make(map[string]*APIResource),
		versions:  make(map[string]int64),
	}
}

// Create stores a new APIResource
func (s *Store) Create(ctx context.Context, api *APIResource) (*APIResource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := api.GetKey()
	if _, exists := s.resources[key]; exists {
		return nil, fmt.Errorf("resource already exists: %s", key)
	}

	// Set initial version
	s.versions[key] = 1
	api.Metadata.Generation = 1

	// Mark as pending
	api.SetPhase("Pending")
	api.SetMessage("Resource stored, waiting for reconciliation")

	// Deep copy to store
	stored := &APIResource{
		Metadata: api.Metadata,
		Spec:     api.Spec,
		Status:   api.Status,
	}

	s.resources[key] = stored
	return stored, nil
}

// Get retrieves an APIResource
func (s *Store) Get(ctx context.Context, namespace, name string) (*APIResource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := namespace + "/" + name
	api, exists := s.resources[key]
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", key)
	}

	return api, nil
}

// List retrieves all APIResources in a namespace
func (s *Store) List(ctx context.Context, namespace string) ([]*APIResource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*APIResource, 0)
	for key, api := range s.resources {
		if api.Metadata.Namespace == namespace {
			result = append(result, api)
		}
	}

	return result, nil
}

// Update updates an APIResource status
func (s *Store) Update(ctx context.Context, api *APIResource) (*APIResource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := api.GetKey()
	_, exists := s.resources[key]
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", key)
	}

	// Increment version on updates
	s.versions[key]++
	api.Metadata.Generation = s.versions[key]
	api.Metadata.UpdatedAt = time.Now()

	s.resources[key] = api
	return api, nil
}

// UpdateStatus updates only the status section
func (s *Store) UpdateStatus(ctx context.Context, namespace, name string, status StatusSection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := namespace + "/" + name
	api, exists := s.resources[key]
	if !exists {
		return fmt.Errorf("resource not found: %s", key)
	}

	api.Status = status
	api.Metadata.UpdatedAt = time.Now()
	return nil
}

// Delete removes an APIResource
func (s *Store) Delete(ctx context.Context, namespace, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := namespace + "/" + name
	if _, exists := s.resources[key]; !exists {
		return fmt.Errorf("resource not found: %s", key)
	}

	delete(s.resources, key)
	delete(s.versions, key)
	return nil
}

// GetAll returns all stored APIResources
func (s *Store) GetAll(ctx context.Context) []*APIResource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*APIResource, 0, len(s.resources))
	for _, api := range s.resources {
		result = append(result, api)
	}

	return result
}

// Exists checks if resource exists
func (s *Store) Exists(ctx context.Context, namespace, name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := namespace + "/" + name
	_, exists := s.resources[key]
	return exists
}
