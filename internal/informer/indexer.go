// Package informer implements the shared-informer pattern from
// k8s.io/client-go/tools/cache.  An informer couples:
//
//   - A ListerWatcher that produces the initial resource snapshot and
//     a stream of incremental changes ("watch events");
//   - An in-memory Store with optional secondary indexes;
//   - A fan-out that delivers Add/Update/Delete events to any number
//     of ResourceEventHandler consumers.
//
// Consumers never talk to the backing store directly — they register
// handlers and query the indexer.  This pattern is the foundation of
// controller-runtime: a reconciler is just an event handler that
// enqueues the resource's key onto a workqueue.
//
// # Implementation scope
//
// This is not a drop-in replacement for client-go's informer.  It is
// deliberately small:
//
//   - No delta-FIFO:  events are delivered synchronously, in order,
//     from a single event-pump goroutine.
//   - No resync ticker: callers that need periodic resync should drive
//     it externally by calling Enqueue on known keys.
//   - No resource-version gap detection: the ListerWatcher contract
//     promises in-order events; reconnection logic is the caller's
//     responsibility.
//
// The API surface is intentionally close to client-go's so that code
// authored against this package can be migrated later if AxiomNizam
// adopts a richer watch infrastructure.
package informer

import (
	"context"
	"fmt"
	"sync"
)

// EventType enumerates the kinds of change an informer delivers.
type EventType string

const (
	// EventAdd is emitted for the initial list snapshot and for
	// subsequent create events.
	EventAdd EventType = "Add"
	// EventUpdate is emitted when an existing object's content changes.
	EventUpdate EventType = "Update"
	// EventDelete is emitted when an object is removed from the source.
	EventDelete EventType = "Delete"
)

// Event is the payload delivered to handlers.  OldObj is populated
// only for EventUpdate and EventDelete.  Objects are passed by
// reference — handlers MUST NOT mutate them in place.
type Event struct {
	Type   EventType
	Obj    map[string]interface{}
	OldObj map[string]interface{}
}

// ResourceEventHandler is the consumer-side interface.  Handlers are
// invoked serially from the informer's event-pump goroutine; a slow
// handler backs up delivery for all handlers — use a local queue if
// the work is expensive.
type ResourceEventHandler interface {
	OnAdd(obj map[string]interface{})
	OnUpdate(oldObj, newObj map[string]interface{})
	OnDelete(obj map[string]interface{})
}

// HandlerFuncs is the bag-of-closures adapter.  Any nil field is
// treated as a no-op.
type HandlerFuncs struct {
	AddFunc    func(obj map[string]interface{})
	UpdateFunc func(oldObj, newObj map[string]interface{})
	DeleteFunc func(obj map[string]interface{})
}

// OnAdd implements ResourceEventHandler.
func (h HandlerFuncs) OnAdd(obj map[string]interface{}) {
	if h.AddFunc != nil {
		h.AddFunc(obj)
	}
}

// OnUpdate implements ResourceEventHandler.
func (h HandlerFuncs) OnUpdate(oldObj, newObj map[string]interface{}) {
	if h.UpdateFunc != nil {
		h.UpdateFunc(oldObj, newObj)
	}
}

// OnDelete implements ResourceEventHandler.
func (h HandlerFuncs) OnDelete(obj map[string]interface{}) {
	if h.DeleteFunc != nil {
		h.DeleteFunc(obj)
	}
}

// ListerWatcher is the source of truth.  Implementations bridge the
// concrete backing store (etcd, SQL, in-memory) to the informer.
type ListerWatcher interface {
	// List returns the full current snapshot.  The informer calls
	// List exactly once at startup; subsequent state is observed via
	// the Watch channel.
	List(ctx context.Context) ([]map[string]interface{}, error)

	// Watch returns a channel that emits Events until ctx is done or
	// the source disconnects (in which case the channel is closed).
	// The informer does not attempt reconnection — that is the
	// ListerWatcher's responsibility.
	Watch(ctx context.Context) (<-chan Event, error)
}

// KeyFunc extracts a unique identifier from an object.  Typical
// implementations return "namespace/name" or the object's UID.
type KeyFunc func(obj map[string]interface{}) (string, error)

// IndexFunc computes secondary-index entries for an object.  Returning
// multiple values from one object is allowed (e.g. indexing a Pod on
// each of its labels).
type IndexFunc func(obj map[string]interface{}) ([]string, error)

// Indexer is the cached, indexed store the informer maintains.
type Indexer struct {
	mu      sync.RWMutex
	keyFunc KeyFunc
	items   map[string]map[string]interface{}
	indexes map[string]IndexFunc           // index name → extractor
	buckets map[string]map[string][]string // index name → value → keys
}

// NewIndexer constructs an empty indexer.  indexes is a map of name →
// extractor; additional indexes can be added via AddIndexer (not
// implemented because late-added indexes need a snapshot reindex pass
// that the current workload does not require).
func NewIndexer(keyFunc KeyFunc, indexes map[string]IndexFunc) *Indexer {
	if indexes == nil {
		indexes = map[string]IndexFunc{}
	}
	buckets := make(map[string]map[string][]string, len(indexes))
	for name := range indexes {
		buckets[name] = map[string][]string{}
	}
	return &Indexer{
		keyFunc: keyFunc,
		items:   map[string]map[string]interface{}{},
		indexes: indexes,
		buckets: buckets,
	}
}

// Add inserts or replaces obj, updating all registered indexes.
func (i *Indexer) Add(obj map[string]interface{}) error {
	key, err := i.keyFunc(obj)
	if err != nil {
		return err
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	if old, ok := i.items[key]; ok {
		i.unindex(key, old)
	}
	i.items[key] = obj
	return i.index(key, obj)
}

// Delete removes obj by key, updating indexes.  Unknown keys are a no-op.
func (i *Indexer) Delete(obj map[string]interface{}) error {
	key, err := i.keyFunc(obj)
	if err != nil {
		return err
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	if old, ok := i.items[key]; ok {
		i.unindex(key, old)
		delete(i.items, key)
	}
	return nil
}

// Get returns the cached object for key, or (nil, false).
func (i *Indexer) Get(key string) (map[string]interface{}, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	v, ok := i.items[key]
	return v, ok
}

// List returns every cached object.  The returned slice aliases
// internal storage — callers that intend to mutate objects must copy.
func (i *Indexer) List() []map[string]interface{} {
	i.mu.RLock()
	defer i.mu.RUnlock()
	out := make([]map[string]interface{}, 0, len(i.items))
	for _, v := range i.items {
		out = append(out, v)
	}
	return out
}

// ByIndex returns the cached objects whose index `indexName` contains
// `value`.  An unknown indexName yields a typed error so callers can
// distinguish "no matches" from "misspelt index".
func (i *Indexer) ByIndex(indexName, value string) ([]map[string]interface{}, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	bucket, ok := i.buckets[indexName]
	if !ok {
		return nil, fmt.Errorf("unknown index %q", indexName)
	}
	keys := bucket[value]
	out := make([]map[string]interface{}, 0, len(keys))
	for _, k := range keys {
		if v, ok := i.items[k]; ok {
			out = append(out, v)
		}
	}
	return out, nil
}

// index inserts key into every registered index bucket.
func (i *Indexer) index(key string, obj map[string]interface{}) error {
	for name, fn := range i.indexes {
		values, err := fn(obj)
		if err != nil {
			return fmt.Errorf("indexer %q: %w", name, err)
		}
		bucket := i.buckets[name]
		for _, v := range values {
			bucket[v] = append(bucket[v], key)
		}
	}
	return nil
}

// unindex removes key from every bucket entry it was placed into.  It
// re-runs the index functions because they are pure and the cost of
// recomputing is lower than the bookkeeping required to remember them.
func (i *Indexer) unindex(key string, obj map[string]interface{}) {
	for name, fn := range i.indexes {
		values, err := fn(obj)
		if err != nil {
			continue
		}
		bucket := i.buckets[name]
		for _, v := range values {
			list := bucket[v]
			filtered := list[:0]
			for _, existing := range list {
				if existing != key {
					filtered = append(filtered, existing)
				}
			}
			if len(filtered) == 0 {
				delete(bucket, v)
			} else {
				bucket[v] = filtered
			}
		}
	}
}

// MetaNamespaceKeyFunc returns "namespace/name" or "name" when the
// object is cluster-scoped.  Matches the k8s client-go helper.
func MetaNamespaceKeyFunc(obj map[string]interface{}) (string, error) {
	meta, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("object has no metadata")
	}
	name, _ := meta["name"].(string)
	if name == "" {
		return "", fmt.Errorf("object has no metadata.name")
	}
	if ns, _ := meta["namespace"].(string); ns != "" {
		return ns + "/" + name, nil
	}
	return name, nil
}
