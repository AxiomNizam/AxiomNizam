package featurestore

// =====================================================
// WS-7.1 — Online Feature Serving Backend
//
// Provides low-latency point lookups for ML inference.
// Supports Redis, PostgreSQL, and in-memory backends for
// serving materialized features by entity key.
// =====================================================

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FeatureVector represents a set of features for a single entity.
type FeatureVector struct {
	EntityKey    string                 `json:"entityKey"`
	Features     map[string]interface{} `json:"features"`
	Timestamp    time.Time              `json:"timestamp"`
	GroupName    string                 `json:"groupName"`
}

// OnlineStore provides low-latency feature retrieval for serving.
type OnlineStore interface {
	// Get retrieves features for a single entity key.
	Get(ctx context.Context, group, entityKey string) (*FeatureVector, error)

	// MultiGet retrieves features for multiple entity keys.
	MultiGet(ctx context.Context, group string, entityKeys []string) ([]*FeatureVector, error)

	// Put writes features for a single entity key.
	Put(ctx context.Context, vector *FeatureVector) error

	// PutBatch writes features for multiple entity keys.
	PutBatch(ctx context.Context, vectors []*FeatureVector) error

	// Delete removes features for an entity key.
	Delete(ctx context.Context, group, entityKey string) error

	// Stats returns store statistics.
	Stats() OnlineStoreStats
}

// OnlineStoreStats tracks online store performance metrics.
type OnlineStoreStats struct {
	Backend      string        `json:"backend"`
	TotalKeys    int64         `json:"totalKeys"`
	TotalGets    int64         `json:"totalGets"`
	TotalPuts    int64         `json:"totalPuts"`
	AvgLatency   time.Duration `json:"avgLatency"`
	CacheHits    int64         `json:"cacheHits"`
	CacheMisses  int64         `json:"cacheMisses"`
}

// MemoryOnlineStore is an in-memory online feature store for development and testing.
type MemoryOnlineStore struct {
	mu       sync.RWMutex
	data     map[string]*FeatureVector // "group/entityKey" -> vector
	stats    OnlineStoreStats
}

// NewMemoryOnlineStore creates a new in-memory online store.
func NewMemoryOnlineStore() *MemoryOnlineStore {
	return &MemoryOnlineStore{
		data:  make(map[string]*FeatureVector),
		stats: OnlineStoreStats{Backend: "memory"},
	}
}

func (s *MemoryOnlineStore) Get(ctx context.Context, group, entityKey string) (*FeatureVector, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.stats.TotalGets++

	key := fmt.Sprintf("%s/%s", group, entityKey)
	v, ok := s.data[key]
	if !ok {
		s.stats.CacheMisses++
		return nil, fmt.Errorf("entity key %q not found in group %q", entityKey, group)
	}
	s.stats.CacheHits++
	return v, nil
}

func (s *MemoryOnlineStore) MultiGet(ctx context.Context, group string, entityKeys []string) ([]*FeatureVector, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*FeatureVector
	for _, ek := range entityKeys {
		key := fmt.Sprintf("%s/%s", group, ek)
		if v, ok := s.data[key]; ok {
			results = append(results, v)
			s.stats.CacheHits++
		} else {
			s.stats.CacheMisses++
		}
		s.stats.TotalGets++
	}
	return results, nil
}

func (s *MemoryOnlineStore) Put(ctx context.Context, vector *FeatureVector) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s/%s", vector.GroupName, vector.EntityKey)
	s.data[key] = vector
	s.stats.TotalPuts++
	s.stats.TotalKeys = int64(len(s.data))
	return nil
}

func (s *MemoryOnlineStore) PutBatch(ctx context.Context, vectors []*FeatureVector) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range vectors {
		key := fmt.Sprintf("%s/%s", v.GroupName, v.EntityKey)
		s.data[key] = v
		s.stats.TotalPuts++
	}
	s.stats.TotalKeys = int64(len(s.data))
	return nil
}

func (s *MemoryOnlineStore) Delete(ctx context.Context, group, entityKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s/%s", group, entityKey)
	delete(s.data, key)
	s.stats.TotalKeys = int64(len(s.data))
	return nil
}

func (s *MemoryOnlineStore) Stats() OnlineStoreStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}
