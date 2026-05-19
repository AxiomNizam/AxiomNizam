package config

import (
	"errors"
	"fmt"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("config validation: %s: %s", e.Field, e.Message)
}

// ValidateEncryptionKey checks if the encryption key meets minimum requirements.
func ValidateEncryptionKey(key []byte) error {
	if len(key) < MinEncryptionKeyBytes {
		return &ValidationError{
			Field:   "EncryptionKey",
			Message: fmt.Sprintf("must be at least %d bytes, got %d", MinEncryptionKeyBytes, len(key)),
		}
	}
	return nil
}

// ValidateHMACKey checks if the HMAC key meets minimum requirements.
func ValidateHMACKey(key []byte) error {
	if len(key) < MinHMACKeyBytes {
		return &ValidationError{
			Field:   "HMACKey",
			Message: fmt.Sprintf("must be at least %d bytes, got %d", MinHMACKeyBytes, len(key)),
		}
	}
	return nil
}

// ValidateTOTPConfig checks TOTP configuration values.
func ValidateTOTPConfig(cfg TOTPConfig) error {
	if cfg.Digits < 4 || cfg.Digits > 8 {
		return errors.New("TOTP digits must be between 4 and 8")
	}
	if cfg.TimeStepSeconds < 15 {
		return errors.New("TOTP time step must be at least 15 seconds")
	}
	if cfg.BackupCodeCount < 5 {
		return errors.New("must generate at least 5 backup codes")
	}
	return nil
}

// ValidateRiskThresholds checks that risk thresholds are in ascending order.
func ValidateRiskThresholds(cfg RiskConfig) error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.LowThreshold >= cfg.MediumThreshold {
		return errors.New("low threshold must be less than medium threshold")
	}
	if cfg.MediumThreshold >= cfg.HighThreshold {
		return errors.New("medium threshold must be less than high threshold")
	}
	return nil
}
