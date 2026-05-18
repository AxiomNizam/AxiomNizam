package config

import (
	"errors"
	"os"
	"time"
)

// Config holds Gatekeeper 2FA module configuration.
type Config struct {
	// Security keys (loaded from environment variables)
	EncryptionKey []byte `yaml:"-"` // AES-256 key for encrypting secrets at rest
	HMACKey       []byte `yaml:"-"` // HMAC key for signing tokens

	// TOTP settings
	TOTP TOTPConfig `yaml:"totp"`

	// Challenge settings
	Challenge ChallengeConfig `yaml:"challenge"`

	// Policy settings
	Policy PolicyConfig `yaml:"policy"`

	// Risk scoring
	Risk RiskConfig `yaml:"risk"`

	// Trusted devices
	TrustedDevices TrustedDeviceConfig `yaml:"trusted_devices"`

	// Backup codes
	BackupCodes BackupCodeConfig `yaml:"backup_codes"`

	// Cache settings
	Cache CacheConfig `yaml:"cache"`

	// Database
	Database DatabaseConfig `yaml:"database"`
}

// TOTPConfig contains TOTP-specific settings.
type TOTPConfig struct {
	Issuer          string `yaml:"issuer"`            // Name shown in authenticator apps
	TimeStepSeconds int    `yaml:"time_step_seconds"` // RFC 6238 time step (default 30)
	Digits          int    `yaml:"digits"`            // OTP digits (default 6)
	BackupCodeCount int    `yaml:"backup_code_count"` // Number of backup codes to generate
	GracePeriodDays int    `yaml:"grace_period_days"` // Days before TOTP is mandatory
}

// ChallengeConfig contains challenge/verification settings.
type ChallengeConfig struct {
	TTLSeconds         int `yaml:"ttl_seconds"`          // Challenge expiration time
	MaxAttempts        int `yaml:"max_attempts"`         // Max failed attempts before lockout
	ClockSkewTolerance int `yaml:"clock_skew_tolerance"` // Allow ±N time steps for clock skew
}

// PolicyConfig contains MFA policy settings.
type PolicyConfig struct {
	Enforcement     string `yaml:"enforcement"`       // "optional", "required", "adaptive"
	MinimumFactors  int    `yaml:"minimum_factors"`   // Minimum factors to require
	GracePeriodDays int    `yaml:"grace_period_days"` // Days before MFA is mandatory
}

// RiskConfig contains risk scoring settings.
type RiskConfig struct {
	Enabled          bool `yaml:"enabled"`            // Enable adaptive risk scoring
	LowThreshold     int  `yaml:"low_threshold"`      // Risk score < threshold = low
	MediumThreshold  int  `yaml:"medium_threshold"`   // medium <= score < high
	HighThreshold    int  `yaml:"high_threshold"`     // high <= score < critical
	RequiresMFAScore int  `yaml:"requires_mfa_score"` // Require MFA above this score
	BlockThreshold   int  `yaml:"block_threshold"`    // Block above this score
}

// TrustedDeviceConfig contains trusted device settings.
type TrustedDeviceConfig struct {
	Enabled         bool          `yaml:"enabled"`          // Enable trusted device feature
	TTLDays         int           `yaml:"ttl_days"`         // Default device trust duration
	MaxPerUser      int           `yaml:"max_per_user"`     // Max devices per user
	CleanupInterval time.Duration `yaml:"cleanup_interval"` // Periodic cleanup interval
}

// BackupCodeConfig contains backup code settings.
type BackupCodeConfig struct {
	Count              int  `yaml:"count"`               // Number of backup codes to generate
	RequireAcknowledge bool `yaml:"require_acknowledge"` // Require user acknowledgment
}

// CacheConfig contains cache/session settings.
type CacheConfig struct {
	Backend string        `yaml:"backend"` // "memory", "redis"
	TTL     time.Duration `yaml:"ttl"`     // Default session TTL
}

// DatabaseConfig contains database settings.
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		TOTP: TOTPConfig{
			Issuer:          "AxiomNizam",
			TimeStepSeconds: 30,
			Digits:          6,
			BackupCodeCount: 10,
			GracePeriodDays: 7,
		},
		Challenge: ChallengeConfig{
			TTLSeconds:         300, // 5 minutes
			MaxAttempts:        3,
			ClockSkewTolerance: 1,
		},
		Policy: PolicyConfig{
			Enforcement:     "optional",
			MinimumFactors:  1,
			GracePeriodDays: 7,
		},
		Risk: RiskConfig{
			Enabled:          true,
			LowThreshold:     30,
			MediumThreshold:  60,
			HighThreshold:    80,
			RequiresMFAScore: 50,
			BlockThreshold:   90,
		},
		TrustedDevices: TrustedDeviceConfig{
			Enabled:         true,
			TTLDays:         90,
			MaxPerUser:      5,
			CleanupInterval: 24 * time.Hour,
		},
		BackupCodes: BackupCodeConfig{
			Count:              10,
			RequireAcknowledge: true,
		},
		Cache: CacheConfig{
			Backend: "memory",
			TTL:     15 * time.Minute,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "axiomnizam",
			Username: "postgres",
			Password: "postgres",
		},
	}
}

// LoadFromEnv loads security keys from environment variables.
// Call this after DefaultConfig() to populate sensitive fields.
func (c *Config) LoadFromEnv() {
	if key := os.Getenv("GATEKEEPER_ENCRYPTION_KEY"); key != "" {
		c.EncryptionKey = []byte(key)
	}
	if key := os.Getenv("GATEKEEPER_HMAC_KEY"); key != "" {
		c.HMACKey = []byte(key)
	}
	if issuer := os.Getenv("GATEKEEPER_TOTP_ISSUER"); issuer != "" {
		c.TOTP.Issuer = issuer
	}
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if len(c.EncryptionKey) < 32 {
		return errors.New("GATEKEEPER_ENCRYPTION_KEY must be at least 32 bytes")
	}

	if len(c.HMACKey) < 32 {
		return errors.New("GATEKEEPER_HMAC_KEY must be at least 32 bytes")
	}

	if c.TOTP.Digits < 4 || c.TOTP.Digits > 8 {
		return errors.New("TOTP digits must be between 4 and 8")
	}

	if c.Challenge.TTLSeconds < 60 {
		return errors.New("challenge TTL must be at least 60 seconds")
	}

	if c.Challenge.MaxAttempts < 1 {
		return errors.New("max attempts must be at least 1")
	}

	if c.Risk.Enabled {
		if c.Risk.LowThreshold >= c.Risk.MediumThreshold ||
			c.Risk.MediumThreshold >= c.Risk.HighThreshold {
			return errors.New("risk thresholds must be in ascending order")
		}
	}

	if c.TrustedDevices.TTLDays < 1 {
		return errors.New("trusted device TTL must be at least 1 day")
	}

	if c.BackupCodes.Count < 5 {
		return errors.New("must generate at least 5 backup codes")
	}

	return nil
}
