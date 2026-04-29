package schemaregistry

// =====================================================
// WS-3.1 — Schema Registry In-Memory Cache
//
// Provides fast schema lookups by ID and by subject+version
// without hitting etcd on every request. The cache is populated
// on startup and updated by the reconciler on schema registration.
// =====================================================

import (
	"sync"
	"time"
)

// SchemaCache provides fast in-memory schema lookups.
type SchemaCache struct {
	mu sync.RWMutex

	// byID maps schema ID to schema resource.
	byID map[int64]*CachedSchema

	// bySubjectVersion maps "subject:version" to schema resource.
	bySubjectVersion map[string]*CachedSchema

	// latestBySubject maps subject to the latest schema.
	latestBySubject map[string]*CachedSchema

	// stats tracks cache performance.
	stats CacheStats
}

// CachedSchema wraps a schema with cache metadata.
type CachedSchema struct {
	Schema    *SchemaResource
	CachedAt  time.Time
	HitCount  int64
}

// CacheStats tracks cache performance metrics.
type CacheStats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Entries    int   `json:"entries"`
	Evictions  int64 `json:"evictions"`
}

// NewSchemaCache creates a new empty schema cache.
func NewSchemaCache() *SchemaCache {
	return &SchemaCache{
		byID:             make(map[int64]*CachedSchema),
		bySubjectVersion: make(map[string]*CachedSchema),
		latestBySubject:  make(map[string]*CachedSchema),
	}
}

// Put adds or updates a schema in the cache.
func (c *SchemaCache) Put(schema *SchemaResource) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cached := &CachedSchema{
		Schema:   schema,
		CachedAt: time.Now(),
	}

	// Index by ID.
	if schema.Status.SchemaID > 0 {
		c.byID[schema.Status.SchemaID] = cached
	}

	// Index by subject:version.
	if schema.Spec.Subject != "" && schema.Status.Version > 0 {
		key := subjectVersionKey(schema.Spec.Subject, schema.Status.Version)
		c.bySubjectVersion[key] = cached
	}

	// Update latest if applicable.
	if schema.Status.IsLatest && schema.Spec.Subject != "" {
		c.latestBySubject[schema.Spec.Subject] = cached
	}

	c.stats.Entries = len(c.byID)
}

// GetByID retrieves a schema by its global ID.
func (c *SchemaCache) GetByID(id int64) (*SchemaResource, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.byID[id]
	if !ok {
		c.mu.RUnlock()
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		c.mu.RLock()
		return nil, false
	}

	cached.HitCount++
	c.mu.RUnlock()
	c.mu.Lock()
	c.stats.Hits++
	c.mu.Unlock()
	c.mu.RLock()
	return cached.Schema, true
}

// GetBySubjectVersion retrieves a schema by subject and version number.
func (c *SchemaCache) GetBySubjectVersion(subject string, version int) (*SchemaResource, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := subjectVersionKey(subject, version)
	cached, ok := c.bySubjectVersion[key]
	if !ok {
		return nil, false
	}

	cached.HitCount++
	return cached.Schema, true
}

// GetLatest retrieves the latest schema for a subject.
func (c *SchemaCache) GetLatest(subject string) (*SchemaResource, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.latestBySubject[subject]
	if !ok {
		return nil, false
	}

	cached.HitCount++
	return cached.Schema, true
}

// Remove removes a schema from the cache.
func (c *SchemaCache) Remove(schema *SchemaResource) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if schema.Status.SchemaID > 0 {
		delete(c.byID, schema.Status.SchemaID)
	}

	if schema.Spec.Subject != "" && schema.Status.Version > 0 {
		key := subjectVersionKey(schema.Spec.Subject, schema.Status.Version)
		delete(c.bySubjectVersion, key)
	}

	// Only remove from latest if this was the latest.
	if schema.Status.IsLatest && schema.Spec.Subject != "" {
		if latest, ok := c.latestBySubject[schema.Spec.Subject]; ok {
			if latest.Schema.Status.SchemaID == schema.Status.SchemaID {
				delete(c.latestBySubject, schema.Spec.Subject)
			}
		}
	}

	c.stats.Entries = len(c.byID)
	c.stats.Evictions++
}

// Clear removes all entries from the cache.
func (c *SchemaCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.byID = make(map[int64]*CachedSchema)
	c.bySubjectVersion = make(map[string]*CachedSchema)
	c.latestBySubject = make(map[string]*CachedSchema)
	c.stats.Entries = 0
}

// Stats returns current cache statistics.
func (c *SchemaCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// Size returns the number of cached schemas.
func (c *SchemaCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.byID)
}

// Subjects returns all cached subject names.
func (c *SchemaCache) Subjects() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	subjects := make([]string, 0, len(c.latestBySubject))
	for subject := range c.latestBySubject {
		subjects = append(subjects, subject)
	}
	return subjects
}

// subjectVersionKey creates a cache key from subject and version.
func subjectVersionKey(subject string, version int) string {
	return subject + ":" + itoa(version)
}

// itoa converts int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
