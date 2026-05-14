package federation

// =====================================================
// WS-5.1 — Federated Query Result Cache
//
// In-memory LRU cache for federated query results.
// Keys are query fingerprints; entries have TTL-based expiry.
// Reduces cross-source query load for repeated analytical queries.
// =====================================================

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// CacheEntry holds a cached query result with metadata.
type CacheEntry struct {
	Key          string          `json:"key"`
	Query        string          `json:"query"`
	Result       *MergedResult   `json:"result"`
	CreatedAt    time.Time       `json:"createdAt"`
	ExpiresAt    time.Time       `json:"expiresAt"`
	HitCount     int64           `json:"hitCount"`
	ByteEstimate int64           `json:"byteEstimate"`
}

// MergedResult represents the cached output of a federated query.
type MergedResult struct {
	Columns  []string                 `json:"columns"`
	Rows     []map[string]interface{} `json:"rows"`
	RowCount int64                    `json:"rowCount"`
	Sources  []string                 `json:"sources"`
}

// CacheStats exposes cache performance metrics.
type CacheStats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Entries    int   `json:"entries"`
	BytesUsed  int64 `json:"bytesUsed"`
	Evictions  int64 `json:"evictions"`
	HitRate    float64 `json:"hitRate"`
}

// ResultCache provides an in-memory LRU cache for federated query results.
type ResultCache struct {
	mu         sync.RWMutex
	entries    map[string]*CacheEntry
	order      []string // LRU order: oldest first
	maxEntries int
	defaultTTL time.Duration
	maxBytes   int64

	// Stats
	hits      int64
	misses    int64
	evictions int64
	bytesUsed int64
}

// NewResultCache creates a new federated result cache.
func NewResultCache(maxEntries int, defaultTTL time.Duration, maxBytes int64) *ResultCache {
	if maxEntries <= 0 {
		maxEntries = 1000
	}
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Minute
	}
	if maxBytes <= 0 {
		maxBytes = 256 * 1024 * 1024 // 256 MB
	}
	return &ResultCache{
		entries:    make(map[string]*CacheEntry),
		maxEntries: maxEntries,
		defaultTTL: defaultTTL,
		maxBytes:   maxBytes,
	}
}

// Get retrieves a cached result by query fingerprint.
func (c *ResultCache) Get(query string) (*MergedResult, bool) {
	key := c.fingerprint(query)

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		c.misses++
		return nil, false
	}

	// Check TTL expiry.
	if time.Now().After(entry.ExpiresAt) {
		c.evictEntry(key)
		c.misses++
		return nil, false
	}

	entry.HitCount++
	c.hits++

	// Move to end of LRU order (most recently used).
	c.touchLRU(key)

	return entry.Result, true
}

// Put stores a query result in the cache with the default TTL.
func (c *ResultCache) Put(query string, result *MergedResult) {
	c.PutWithTTL(query, result, c.defaultTTL)
}

// PutWithTTL stores a query result with a custom TTL.
func (c *ResultCache) PutWithTTL(query string, result *MergedResult, ttl time.Duration) {
	key := c.fingerprint(query)
	now := time.Now()

	byteEst := estimateResultSize(result)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict expired entries first.
	c.evictExpired(now)

	// Evict LRU entries if at capacity.
	for len(c.entries) >= c.maxEntries || (c.bytesUsed+byteEst > c.maxBytes && len(c.entries) > 0) {
		c.evictOldest()
	}

	// Store entry.
	c.entries[key] = &CacheEntry{
		Key:          key,
		Query:        query,
		Result:       result,
		CreatedAt:    now,
		ExpiresAt:    now.Add(ttl),
		ByteEstimate: byteEst,
	}
	c.order = append(c.order, key)
	c.bytesUsed += byteEst
}

// Invalidate removes a specific query from the cache.
func (c *ResultCache) Invalidate(query string) bool {
	key := c.fingerprint(query)

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.entries[key]; ok {
		c.evictEntry(key)
		return true
	}
	return false
}

// InvalidateBySource removes all cached results that used a specific data source.
func (c *ResultCache) InvalidateBySource(source string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for key, entry := range c.entries {
		for _, s := range entry.Result.Sources {
			if s == source {
				c.evictEntry(key)
				count++
				break
			}
		}
	}
	return count
}

// Clear removes all entries from the cache.
func (c *ResultCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.order = nil
	c.bytesUsed = 0
}

// Stats returns current cache performance statistics.
func (c *ResultCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100.0
	}

	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Entries:   len(c.entries),
		BytesUsed: c.bytesUsed,
		Evictions: c.evictions,
		HitRate:   hitRate,
	}
}

// --- Internal helpers ---

func (c *ResultCache) fingerprint(query string) string {
	h := sha256.Sum256([]byte(query))
	return hex.EncodeToString(h[:])[:16]
}

func (c *ResultCache) evictEntry(key string) {
	entry, ok := c.entries[key]
	if !ok {
		return
	}
	c.bytesUsed -= entry.ByteEstimate
	delete(c.entries, key)
	c.evictions++

	// Remove from LRU order.
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

func (c *ResultCache) evictOldest() {
	if len(c.order) == 0 {
		return
	}
	c.evictEntry(c.order[0])
}

func (c *ResultCache) evictExpired(now time.Time) {
	var expired []string
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			expired = append(expired, key)
		}
	}
	for _, key := range expired {
		c.evictEntry(key)
	}
}

func (c *ResultCache) touchLRU(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, key)
			return
		}
	}
}

func estimateResultSize(result *MergedResult) int64 {
	if result == nil {
		return 0
	}
	// Rough estimate: 100 bytes per row + 50 bytes per column header.
	return int64(len(result.Columns))*50 + result.RowCount*100
}
