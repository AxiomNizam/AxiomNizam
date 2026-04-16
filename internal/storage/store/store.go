package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"example.com/axiomnizam/internal/storage/models"
	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// BucketStore provides in-memory CRUD operations for BucketResource objects.
// Follows the IAM storage pattern with key-prefixed lookups.
type BucketStore struct {
	mu      sync.RWMutex
	buckets map[string]*models.BucketResource // key = tenantID/bucketName
	etcd    *clientv3.Client
}

const (
	bucketStoreEtcdTimeout = 3 * time.Second
	bucketStoreEtcdPrefix  = "storage:bucketstore/"
)

// NewBucketStore creates an empty bucket store.
func NewBucketStore() *BucketStore {
	return &BucketStore{
		buckets: make(map[string]*models.BucketResource),
	}
}

// ConfigurePersistence enables etcd-backed persistence for bucket resources.
// Existing buckets are loaded from etcd when configured.
func (s *BucketStore) ConfigurePersistence(etcd *clientv3.Client) {
	s.mu.Lock()
	s.etcd = etcd
	s.mu.Unlock()
	s.loadFromEtcd()
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
	s.persistBucketUnlocked(b)
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
	s.persistBucketUnlocked(b)
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
	s.persistBucketUnlocked(b)
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
	s.deleteBucketFromEtcdUnlocked(tenantID, name)
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

func (s *BucketStore) loadFromEtcd() {
	s.mu.RLock()
	etcd := s.etcd
	s.mu.RUnlock()
	if etcd == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), bucketStoreEtcdTimeout)
	defer cancel()
	resp, err := etcd.Get(ctx, bucketStoreEtcdPrefix, clientv3.WithPrefix())
	if err != nil {
		log.Printf("storage bucket store: etcd load failed: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, kv := range resp.Kvs {
		var b models.BucketResource
		if err := json.Unmarshal(kv.Value, &b); err != nil {
			continue
		}
		cb := b
		s.buckets[bucketKey(cb.Metadata.TenantID, cb.Metadata.Name)] = &cb
	}
}

func (s *BucketStore) persistBucketUnlocked(b *models.BucketResource) {
	if b == nil || s.etcd == nil {
		return
	}
	data, err := json.Marshal(b)
	if err != nil {
		log.Printf("storage bucket store: marshal failed: %v", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), bucketStoreEtcdTimeout)
	defer cancel()
	key := bucketStoreEtcdPrefix + bucketKey(b.Metadata.TenantID, b.Metadata.Name)
	if _, err := s.etcd.Put(ctx, key, string(data)); err != nil {
		log.Printf("storage bucket store: etcd put failed for key %s: %v", key, err)
	}
}

func (s *BucketStore) deleteBucketFromEtcdUnlocked(tenantID, name string) {
	if s.etcd == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), bucketStoreEtcdTimeout)
	defer cancel()
	key := bucketStoreEtcdPrefix + bucketKey(tenantID, name)
	if _, err := s.etcd.Delete(ctx, key); err != nil {
		log.Printf("storage bucket store: etcd delete failed for key %s: %v", key, err)
	}
}

// TenantBucketName returns the storage-level bucket name for a tenant.
func TenantBucketName(prefix, tenantID, name string) string {
	return strings.ToLower(fmt.Sprintf("%s%s-%s", prefix, tenantID, name))
}
