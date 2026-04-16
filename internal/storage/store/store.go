package store

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"example.com/axiomnizam/internal/storage/models"
	"github.com/google/uuid"
)

// BucketStore provides in-memory CRUD operations for BucketResource objects.
// Follows the IAM storage pattern with key-prefixed lookups.
type BucketStore struct {
	mu      sync.RWMutex
	buckets map[string]*models.BucketResource // key = tenantID/bucketName
}

// NewBucketStore creates an empty bucket store.
func NewBucketStore() *BucketStore {
	return &BucketStore{
		buckets: make(map[string]*models.BucketResource),
	}
}

func bucketKey(tenantID, name string) string {
	return tenantID + "/" + name
}

// Create adds a new bucket resource. Returns an error if it already exists.
func (s *BucketStore) Create(b *models.BucketResource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := bucketKey(b.Metadata.TenantID, b.Metadata.Name)
	if _, exists := s.buckets[key]; exists {
		return fmt.Errorf("bucket %q already exists for tenant %q", b.Metadata.Name, b.Metadata.TenantID)
	}

	now := time.Now().UTC()
	b.APIVersion = "storage.axiom.dev/v1"
	b.Kind = "Bucket"
	b.Metadata.UID = uuid.New().String()
	b.Metadata.CreatedAt = now
	b.Metadata.UpdatedAt = now
	b.Generation = 1
	b.Status.Phase = models.BucketPhasePending

	s.buckets[key] = b
	return nil
}

// Get retrieves a bucket resource by tenant and name.
func (s *BucketStore) Get(tenantID, name string) (*models.BucketResource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := bucketKey(tenantID, name)
	b, ok := s.buckets[key]
	if !ok {
		return nil, fmt.Errorf("bucket %q not found for tenant %q", name, tenantID)
	}
	return b, nil
}

// Update replaces a bucket resource. Increments generation.
func (s *BucketStore) Update(b *models.BucketResource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := bucketKey(b.Metadata.TenantID, b.Metadata.Name)
	if _, exists := s.buckets[key]; !exists {
		return fmt.Errorf("bucket %q not found for tenant %q", b.Metadata.Name, b.Metadata.TenantID)
	}

	b.Generation++
	b.Metadata.UpdatedAt = time.Now().UTC()
	s.buckets[key] = b
	return nil
}

// UpdateStatus updates only the status of a bucket resource.
func (s *BucketStore) UpdateStatus(tenantID, name string, status models.BucketStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := bucketKey(tenantID, name)
	b, ok := s.buckets[key]
	if !ok {
		return fmt.Errorf("bucket %q not found for tenant %q", name, tenantID)
	}

	b.Status = status
	b.Metadata.UpdatedAt = time.Now().UTC()
	return nil
}

// Delete removes a bucket resource.
func (s *BucketStore) Delete(tenantID, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := bucketKey(tenantID, name)
	if _, exists := s.buckets[key]; !exists {
		return fmt.Errorf("bucket %q not found for tenant %q", name, tenantID)
	}
	delete(s.buckets, key)
	return nil
}

// List returns all bucket resources, optionally filtered by tenant.
func (s *BucketStore) List(tenantID string) []*models.BucketResource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.BucketResource
	for _, b := range s.buckets {
		if tenantID == "" || b.Metadata.TenantID == tenantID {
			result = append(result, b)
		}
	}
	return result
}

// ListAll returns all bucket resources across all tenants.
func (s *BucketStore) ListAll() []*models.BucketResource {
	return s.List("")
}

// TenantBucketName returns the storage-level bucket name for a tenant.
func TenantBucketName(prefix, tenantID, name string) string {
	return strings.ToLower(fmt.Sprintf("%s%s-%s", prefix, tenantID, name))
}
