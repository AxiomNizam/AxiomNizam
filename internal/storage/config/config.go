package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds storage module configuration.
type Config struct {
	// Core settings
	DataDir       string `json:"dataDir"`       // filesystem root for object data
	BucketPrefix  string `json:"bucketPrefix"`  // e.g., "axiom-"
	PresignSecret string `json:"presignSecret"` // HMAC key for presign tokens

	// Rate limits
	PresignRateLimitPerMinute  int `json:"presignRateLimitPerMinute"`  // presigned URL rate limit
	ObjectRateLimitPerMinute   int `json:"objectRateLimitPerMinute"`   // legacy combined rate limit
	ObjectReadRateLimit        int `json:"objectReadRateLimit"`        // per-minute read ops
	ObjectWriteRateLimit       int `json:"objectWriteRateLimit"`       // per-minute write ops

	// Controller settings
	ControllerResyncInterval time.Duration `json:"controllerResyncInterval"` // reconciliation interval
	ControllerDebug          bool          `json:"controllerDebug"`          // debug logging

	// Timeouts
	EtcdTimeout    time.Duration `json:"etcdTimeout"`    // etcd operation timeout
	AuditTimeout   time.Duration `json:"auditTimeout"`   // audit KV operation timeout
	MetricsTimeout time.Duration `json:"metricsTimeout"` // metrics KV operation timeout

	// Capacity limits
	MaxAuditEvents       int `json:"maxAuditEvents"`       // max events in memory
	MaxPersistentEvents  int `json:"maxPersistentEvents"`  // max events persisted to Raft KV
	DefaultQueryLimit    int `json:"defaultQueryLimit"`    // default event listing limit
}

// DefaultConfig returns configuration populated from environment variables
// with sensible defaults for a local native storage backend.
func DefaultConfig() Config {
	cfg := Config{
		DataDir:       envStr("STORAGE_DATA_DIR", "/data/storage"),
		BucketPrefix:  envStr("STORAGE_BUCKET_PREFIX", "axiom-"),
		PresignSecret: envStr("STORAGE_PRESIGN_SECRET", "axiom-native-storage-default-key"),

		PresignRateLimitPerMinute: envInt("STORAGE_PRESIGN_RATE_LIMIT_PER_MINUTE", 0),
		ObjectRateLimitPerMinute:  envInt("STORAGE_OBJECT_RATE_LIMIT_PER_MINUTE", 240),
		ObjectReadRateLimit:       envInt("STORAGE_OBJECT_READ_RATE_LIMIT_PER_MINUTE", 0),
		ObjectWriteRateLimit:      envInt("STORAGE_OBJECT_WRITE_RATE_LIMIT_PER_MINUTE", 0),

		ControllerResyncInterval: envDuration("STORAGE_CONTROLLER_RESYNC_INTERVAL", 7*time.Minute),
		ControllerDebug:          envBool("STORAGE_CONTROLLER_DEBUG", false),

		EtcdTimeout:    3 * time.Second,
		AuditTimeout:   5 * time.Second,
		MetricsTimeout: 3 * time.Second,

		MaxAuditEvents:      10000,
		MaxPersistentEvents: 1000,
		DefaultQueryLimit:   100,
	}

	// Read/Write rate limits fall back to the legacy combined limit.
	if cfg.ObjectReadRateLimit <= 0 {
		cfg.ObjectReadRateLimit = cfg.ObjectRateLimitPerMinute
	}
	if cfg.ObjectWriteRateLimit <= 0 {
		cfg.ObjectWriteRateLimit = cfg.ObjectRateLimitPerMinute
	}

	// Clamp controller resync interval to [5m, 10m].
	if cfg.ControllerResyncInterval < 5*time.Minute {
		cfg.ControllerResyncInterval = 5 * time.Minute
	}
	if cfg.ControllerResyncInterval > 10*time.Minute {
		cfg.ControllerResyncInterval = 10 * time.Minute
	}

	return cfg
}

// LoadFromEnv is an alias for DefaultConfig (env vars are read in DefaultConfig).
func LoadFromEnv() Config {
	return DefaultConfig()
}

// Validate checks the configuration for invalid values.
func (c Config) Validate() error {
	if c.DataDir == "" {
		return fmt.Errorf("storage: data dir must not be empty")
	}
	if c.MaxAuditEvents <= 0 {
		return fmt.Errorf("storage: max audit events must be positive")
	}
	return nil
}

// --- env helpers ---

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if s := strings.TrimSpace(os.Getenv(key)); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if s := strings.TrimSpace(os.Getenv(key)); s != "" {
		if v, err := time.ParseDuration(s); err == nil {
			return v
		}
	}
	return fallback
}
