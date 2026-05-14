// Package cache implements the client-side object stores that sit in
// front of every controller's workqueue.  Two variants are provided:
//
//   - ThreadSafeStore: a keyed map with pluggable indexers — the
//     foundation of every SharedIndexInformer.
//   - DeltaFIFO:       an ordered queue of object deltas (Added,
//     Updated, Deleted, Sync).  The reflector writes here; the
//     informer pops from here and fans out events to listeners.
//
// Kept in one package so tests and controllers that need both don't
// cross package boundaries for what is conceptually one subsystem.
package cache

import (
	"sync"
)

// KeyFunc derives the storage key for obj.  The convention is
// "namespace/name" for namespaced objects and just "name" for
// cluster-scoped — matching k8s.MetaNamespaceKeyFunc.
type KeyFunc func(obj interface{}) (string, error)

// IndexFunc extracts the secondary-key values from obj — e.g. all
// pods owned by a given ReplicaSet.  Returns a slice so a single
// object can appear under multiple index keys.
type IndexFunc func(obj interface{}) ([]string, error)

// Indexers is a named-indexer registry.  Each entry defines one
// axis along which the store supports fan-out queries.
type Indexers map[string]IndexFunc

// Indices is the precomputed map: indexName → indexKey → set-of-keys.
type Indices map[string]map[string]map[string]struct{}

// ThreadSafeStore is a goroutine-safe keyed map with indexers.  It
// is the non-indexed primary store plus an arbitrary number of
// secondary indexes kept in sync on every mutation.
type ThreadSafeStore struct {
	mu       sync.RWMutex
	items    map[string]interface{}
	indexers Indexers
	indices  Indices
}

// NewThreadSafeStore returns a store configured with the given
// indexers.  Passing nil indexers is valid — the store degrades to a
// plain keyed map.
func NewThreadSafeStore(indexers Indexers) *ThreadSafeStore {
	if indexers == nil {
		indexers = Indexers{}
	}
	return &ThreadSafeStore{
		items:    map[string]interface{}{},
		indexers: indexers,
		indices:  Indices{},
	}
}

// Add inserts or replaces obj under key.
func (s *ThreadSafeStore) Add(key string, obj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, existed := s.items[key]
	s.items[key] = obj
	if existed {
		s.updateIndices(old, obj, key)
		return
	}
	s.updateIndices(nil, obj, key)
}

// Update is an alias for Add — matches upstream semantics.
func (s *ThreadSafeStore) Update(key string, obj interface{}) { s.Add(key, obj) }

// Delete removes the entry and its index references.
func (s *ThreadSafeStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	obj, ok := s.items[key]
	if !ok {
		return
	}
	delete(s.items, key)
	s.updateIndices(obj, nil, key)
}

// Get returns the stored object and whether it was present.
func (s *ThreadSafeStore) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	obj, ok := s.items[key]
	return obj, ok
}

// List returns a snapshot of all values.
func (s *ThreadSafeStore) List() []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]interface{}, 0, len(s.items))
	for _, v := range s.items {
		out = append(out, v)
	}
	return out
}

// ListKeys returns a snapshot of all primary keys.
func (s *ThreadSafeStore) ListKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.items))
	for k := range s.items {
		out = append(out, k)
	}
	return out
}

// ByIndex returns every object whose indexFunc output contains indexKey.
func (s *ThreadSafeStore) ByIndex(indexName, indexKey string) []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.indices[indexName][indexKey]
	out := make([]interface{}, 0, len(bucket))
	for k := range bucket {
		if v, ok := s.items[k]; ok {
			out = append(out, v)
		}
	}
	return out
}

// Replace is the bulk-set used by reflector resyncs.  It swaps the
// entire item map and rebuilds every index.
func (s *ThreadSafeStore) Replace(items map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = items
	s.indices = Indices{}
	for k, v := range items {
		s.updateIndices(nil, v, k)
	}
}

// updateIndices reflects a single-entry change in every indexer.
// old=nil means insert; new=nil means delete; both non-nil means update.
// Callers must hold s.mu for writing.
func (s *ThreadSafeStore) updateIndices(oldObj, newObj interface{}, key string) {
	for name, fn := range s.indexers {
		if oldObj != nil {
			vals, err := fn(oldObj)
			if err == nil {
				for _, v := range vals {
					if bucket, ok := s.indices[name][v]; ok {
						delete(bucket, key)
						if len(bucket) == 0 {
							delete(s.indices[name], v)
						}
					}
				}
			}
		}
		if newObj != nil {
			vals, err := fn(newObj)
			if err != nil {
				continue
			}
			if s.indices[name] == nil {
				s.indices[name] = map[string]map[string]struct{}{}
			}
			for _, v := range vals {
				if s.indices[name][v] == nil {
					s.indices[name][v] = map[string]struct{}{}
				}
				s.indices[name][v][key] = struct{}{}
			}
		}
	}
}
