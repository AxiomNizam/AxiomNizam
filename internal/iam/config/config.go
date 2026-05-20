package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Default configuration constants.
const (
	DefaultAccessTokenTTL  = 15 * time.Minute
	DefaultRefreshTokenTTL = 7 * 24 * time.Hour // 7 days
	DefaultSysadminEmail   = "admin@axiomnizam.local"
	DefaultRealm           = "axiomnizam"
	DefaultRSAKeyBits      = 2048
	DefaultBcryptCost      = 12
)

// Config holds IAM module configuration.
type Config struct {
	// IssuerURL is the public URL of this IAM instance (e.g. https://api.example.com).
	IssuerURL string

	// SysadminEmail is the bootstrap sysadmin account email.
	SysadminEmail string
	// SysadminPassword is the bootstrap sysadmin password.
	SysadminPassword string

	// AccessTokenTTL overrides the default 15-minute access token lifetime.
	AccessTokenTTL time.Duration
	// RefreshTokenTTL overrides the default 7-day refresh token lifetime.
	RefreshTokenTTL time.Duration

	// RSA key material — inline PEM or file path.
	RSAPrivateKey      string // IAM_RSA_PRIVATE_KEY (inline PEM)
	RSAPrivateKeyFile  string // IAM_RSA_PRIVATE_KEY_FILE (path to PEM file)
	RSAKeyBits         int    // key size for auto-generation (default 2048)

	// Realm name used for bootstrap and default realm creation.
	Realm string // IAM_REALM

	// Timeouts for etcd/KV operations.
	EtcdTimeout time.Duration

	// Crypto parameters.
	BcryptCost int

	// Client defaults.
	DefaultClientRateLimitMaxCalls int64
	DefaultClientTokenValidityMin  int
}

// DefaultConfig returns configuration populated from environment variables
// with sensible defaults.
func DefaultConfig() Config {
	return Config{
		IssuerURL:        strings.TrimSpace(os.Getenv("IAM_ISSUER_URL")),
		SysadminEmail:    envStr("IAM_SYSADMIN_EMAIL", DefaultSysadminEmail),
		SysadminPassword: os.Getenv("IAM_SYSADMIN_PASSWORD"),
		AccessTokenTTL:   envDuration("IAM_ACCESS_TOKEN_TTL", DefaultAccessTokenTTL),
		RefreshTokenTTL:  envDuration("IAM_REFRESH_TOKEN_TTL", DefaultRefreshTokenTTL),

		RSAPrivateKey:     os.Getenv("IAM_RSA_PRIVATE_KEY"),
		RSAPrivateKeyFile: os.Getenv("IAM_RSA_PRIVATE_KEY_FILE"),
		RSAKeyBits:        DefaultRSAKeyBits,

		Realm: envStr("IAM_REALM", DefaultRealm),

		EtcdTimeout: 3 * time.Second,

		BcryptCost: DefaultBcryptCost,

		DefaultClientRateLimitMaxCalls: 500,
		DefaultClientTokenValidityMin:  15,
	}
}

// LoadFromEnv creates a Config from defaults and overrides from env vars.
func LoadFromEnv() Config {
	return DefaultConfig()
}

// Validate checks the configuration for invalid values.
func (c Config) Validate() error {
	if c.AccessTokenTTL <= 0 {
		return fmt.Errorf("iam: access token TTL must be positive")
	}
	if c.RefreshTokenTTL <= 0 {
		return fmt.Errorf("iam: refresh token TTL must be positive")
	}
	if c.BcryptCost < 4 || c.BcryptCost > 31 {
		return fmt.Errorf("iam: bcrypt cost must be between 4 and 31")
	}
	return nil
}

// --- env helpers ---

func envStr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func envInt(key string, fallback int) int {
	if s := strings.TrimSpace(os.Getenv(key)); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return fallback
}
