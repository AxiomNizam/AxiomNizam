package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds job scheduler configuration.
type Config struct {
	// Max size of queue
	MaxQueueSize int

	// Max retries for failed jobs
	MaxRetries int

	// Default timeout for jobs
	DefaultTimeout time.Duration

	// Default job priority (1=low, 5=normal, 10=high, 20=critical)
	DefaultPriority int

	// Number of worker goroutines
	NumWorkers int

	// Enable persistence of job results
	PersistResults bool

	// Log level for job execution
	LogLevel string

	// Dead letter queue settings
	DLQMaxSize   int           // max entries in DLQ (default: 10000)
	DLQRetention time.Duration // how long to keep DLQ entries (default: 720h = 30 days)

	// Channel buffer sizes
	EmailQueueSize  int // email queue channel buffer (default: 100)
	ResultQueueSize int // result channel buffer (default: 100)

	// Health check threshold
	HealthFailureRate float64 // failure rate threshold for degraded status (default: 0.5)
}

// DefaultConfig returns production-safe defaults.
func DefaultConfig() Config {
	return Config{
		MaxQueueSize:    10000,
		MaxRetries:      3,
		DefaultTimeout:  30 * time.Minute,
		DefaultPriority: 5, // normal
		NumWorkers:      10,
		PersistResults:  true,
		LogLevel:        "info",

		DLQMaxSize:   10000,
		DLQRetention: 30 * 24 * time.Hour,

		EmailQueueSize:  100,
		ResultQueueSize: 100,

		HealthFailureRate: 0.5,
	}
}

// LoadFromEnv creates a Config from defaults and overrides from env vars.
func LoadFromEnv() Config {
	cfg := DefaultConfig()
	if v := envInt("JOB_MAX_QUEUE_SIZE"); v > 0 {
		cfg.MaxQueueSize = v
	}
	if v := envInt("JOB_MAX_RETRIES"); v >= 0 {
		cfg.MaxRetries = v
	}
	if v := envDuration("JOB_DEFAULT_TIMEOUT"); v > 0 {
		cfg.DefaultTimeout = v
	}
	if v := envInt("JOB_DEFAULT_PRIORITY"); v > 0 {
		cfg.DefaultPriority = v
	}
	if v := envInt("JOB_NUM_WORKERS"); v > 0 {
		cfg.NumWorkers = v
	}
	if v := envBool("JOB_PERSIST_RESULTS"); v != nil {
		cfg.PersistResults = *v
	}
	if v := os.Getenv("JOB_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	return cfg
}

// Validate checks the configuration for invalid values.
func (c Config) Validate() error {
	if c.MaxQueueSize < 1 {
		return fmt.Errorf("jobs: max queue size must be >= 1, got %d", c.MaxQueueSize)
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("jobs: max retries must be >= 0, got %d", c.MaxRetries)
	}
	if c.NumWorkers < 1 {
		return fmt.Errorf("jobs: num workers must be >= 1, got %d", c.NumWorkers)
	}
	if c.DefaultPriority < 1 || c.DefaultPriority > 20 {
		return fmt.Errorf("jobs: default priority must be 1-20, got %d", c.DefaultPriority)
	}
	return nil
}

func envInt(key string) int {
	v := os.Getenv(key)
	if v == "" {
		return -1
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return -1
	}
	return n
}

func envBool(key string) *bool {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}
	return &b
}

func envDuration(key string) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0
	}
	return d
}
