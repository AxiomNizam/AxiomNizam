package config

import "time"

// Default configuration constants.
const (
	DefaultTOTPIssuer          = "AxiomNizam"
	DefaultTOTPTimeStep        = 30
	DefaultTOTPDigits          = 6
	DefaultBackupCodeCount     = 10
	DefaultGracePeriodDays     = 7

	DefaultChallengeTTL        = 5 * time.Minute
	DefaultMaxAttempts         = 3
	DefaultClockSkewTolerance  = 1

	DefaultDeviceTTLDays       = 90
	DefaultMaxDevicesPerUser   = 5
	DefaultDeviceCleanupInterval = 24 * time.Hour

	DefaultCacheBackend        = "memory"
	DefaultCacheTTL            = 15 * time.Minute

	DefaultDBHost              = "localhost"
	DefaultDBPort              = 5432
	DefaultDBName              = "axiomnizam"

	DefaultRiskLowThreshold    = 30
	DefaultRiskMediumThreshold = 60
	DefaultRiskHighThreshold   = 80
	DefaultRiskMFAScore        = 50
	DefaultRiskBlockThreshold  = 90

	// Minimum key lengths for security
	MinEncryptionKeyBytes      = 32
	MinHMACKeyBytes            = 32
)
