package cache

import (
	"example.com/axiomnizam/internal/logging"
	"fmt"
	"time"
)

// Manager handles cache initialization and management
type Manager struct {
	cache   Cache
	config  *CacheConfig
	backend string
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config *CacheConfig) (*Manager, error) {
	if config == nil {
		config = &CacheConfig{
			Type:       "memory",
			DefaultTTL: 15 * time.Minute,
			MaxSize:    1000,
		}
	}


	manager := &Manager{
		config: config,	}

	// Initialize cache backend
	if err := manager.initializeCache(); err != nil {
		return nil, err
	}

	return manager, nil
}

// initializeCache initializes the cache backend based on configuration
func (m *Manager) initializeCache() error {
	switch m.config.Type {
	case "redis":
		cache, err := NewRedisCache(m.config)
		if err != nil {
			return fmt.Errorf("failed to initialize Redis cache: %w", err)
		}
		m.cache = cache
		m.backend = "Redis"
		logging.Z().Info(fmt.Sprintf("Redis cache initialized: %s:%d", m.config.Host, m.config.Port))

	case "memory":
		m.cache = NewMemoryCache(m.config.MaxSize)
		m.backend = "Memory"
		logging.Z().Info(fmt.Sprintf("Memory cache initialized (max size: %d)", m.config.MaxSize))

	default:
		return fmt.Errorf("unknown cache type: %s", m.config.Type)
	}

	return nil
}

// GetCache returns the underlying cache instance
func (m *Manager) GetCache() Cache {
	return m.cache
}

// GetBackend returns the cache backend type
func (m *Manager) GetBackend() string {
	return m.backend
}

// Health checks the health of the cache backend
func (m *Manager) Health() error {
	if m.cache == nil {
		return fmt.Errorf("cache not initialized")
	}

	return m.cache.Health(nil)
}

// Close closes the cache connection
func (m *Manager) Close() error {
	if m.cache != nil {
		return m.cache.Close()
	}
	return nil
}

// SwitchBackend switches to a different cache backend
func (m *Manager) SwitchBackend(cacheType string) error {
	oldBackend := m.backend

	m.config.Type = cacheType
	if err := m.initializeCache(); err != nil {
		logging.Z().Info(fmt.Sprintf("Error switching to %s cache, keeping %s", cacheType, oldBackend))
		return err
	}

	logging.Z().Info(fmt.Sprintf("Switched from %s to %s cache", oldBackend, m.backend))
	return nil
}

// GetConfig returns the current cache configuration
func (m *Manager) GetConfig() *CacheConfig {
	return m.config
}

// DefaultCacheConfig returns a sensible default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Type:       "memory",
		DefaultTTL: 15 * time.Minute,
		MaxSize:    1000,
	}
}

// RedisCacheConfig returns a Redis cache configuration with sensible defaults
func RedisCacheConfig(host string, port int, password string) *CacheConfig {
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 6379
	}

	return &CacheConfig{
		Type:       "redis",
		Host:       host,
		Port:       port,
		Password:   password,
		DB:         0,
		DefaultTTL: 15 * time.Minute,
		PoolSize:   10,
	}
}

// MemoryCacheConfig returns a memory cache configuration
func MemoryCacheConfig(maxSize int) *CacheConfig {
	if maxSize <= 0 {
		maxSize = 1000
	}

	return &CacheConfig{
		Type:       "memory",
		DefaultTTL: 15 * time.Minute,
		MaxSize:    maxSize,
	}
}

// BuildCacheConfig creates a cache config from environment variables or defaults
// This is a helper function that can be expanded based on your config system
func BuildCacheConfig(cacheType string) *CacheConfig {
	switch cacheType {
	case "redis":
		return RedisCacheConfig("localhost", 6379, "")
	case "memory":
		return MemoryCacheConfig(1000)
	default:
		return DefaultCacheConfig()
	}
}
