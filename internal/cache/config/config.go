package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds cache module configuration.
type Config struct {
	// Type is the cache backend type (redis, memory, etc.)
	Type string

	// Host is the cache server host (for Redis)
	Host string

	// Port is the cache server port (for Redis)
	Port int

	// Password is the cache server password (for Redis)
	Password string

	// DB is the cache database number (for Redis)
	DB int

	// DefaultTTL is the default time-to-live for cache entries
	DefaultTTL time.Duration

	// MaxSize is the maximum size of in-memory cache
	MaxSize int

	// PoolSize is the connection pool size (for Redis)
	PoolSize int

	// Address is the full Redis address (host:port) — alternative to Host+Port
	Address string

	// MaxRetries for Redis operations
	MaxRetries int

	// ReadTimeout for Redis operations
	ReadTimeout time.Duration

	// WriteTimeout for Redis operations
	WriteTimeout time.Duration

	// MaxIdle connections in pool
	MaxIdle int

	// MaxActive connections in pool
	MaxActive int

	// Backend is the cache backend type (alias for Type)
	Backend string
}

// DefaultConfig returns production-safe defaults.
func DefaultConfig() Config {
	return Config{
		Type:         getEnv("CACHE_BACKEND", "memory"),
		Backend:      getEnv("CACHE_BACKEND", "memory"),
		Host:         getEnv("CACHE_REDIS_HOST", "localhost"),
		Port:         envInt("CACHE_REDIS_PORT", 6379),
		Address:      getEnv("CACHE_REDIS_ADDRESS", "localhost:6379"),
		Password:     getEnv("CACHE_REDIS_PASSWORD", ""),
		DB:           envInt("CACHE_REDIS_DB", 0),
		DefaultTTL:   envDuration("CACHE_DEFAULT_TTL", 5*time.Minute),
		MaxSize:      envInt("CACHE_MAX_SIZE", 1000),
		PoolSize:     envInt("CACHE_POOL_SIZE", 10),
		MaxRetries:   envInt("CACHE_MAX_RETRIES", 3),
		ReadTimeout:  envDuration("CACHE_READ_TIMEOUT", 3*time.Second),
		WriteTimeout: envDuration("CACHE_WRITE_TIMEOUT", 3*time.Second),
		MaxIdle:      envInt("CACHE_MAX_IDLE", 10),
		MaxActive:    envInt("CACHE_MAX_ACTIVE", 100),
	}
}

// LoadFromEnv creates a Config from defaults and overrides from env vars.
func LoadFromEnv() Config {
	return DefaultConfig()
}

// Validate checks the configuration for invalid values.
func (c Config) Validate() error {
	switch c.Backend {
	case "redis", "memory":
		// ok
	default:
		return fmt.Errorf("cache: unknown backend %q (expected redis or memory)", c.Backend)
	}
	if c.Backend == "redis" && c.Address == "" && c.Host == "" {
		return fmt.Errorf("cache: redis address or host required when backend is redis")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
