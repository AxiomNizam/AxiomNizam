package totp

import "errors"

var (
	// ErrInvalidSecret indicates the TOTP secret is empty or malformed.
	ErrInvalidSecret = errors.New("invalid TOTP secret")

	// ErrInvalidCode indicates the OTP code format is invalid.
	ErrInvalidCode = errors.New("invalid OTP code format")

	// ErrCodeExpired indicates the OTP code has expired.
	ErrCodeExpired = errors.New("OTP code expired")

	// ErrCodeMismatch indicates the OTP code does not match.
	ErrCodeMismatch = errors.New("OTP code does not match")

	// ErrSecretGenerationFailed indicates secret generation failed.
	ErrSecretGenerationFailed = errors.New("failed to generate secret")

	// ErrQRCodeGenerationFailed indicates QR code generation failed.
	ErrQRCodeGenerationFailed = errors.New("failed to generate QR code")
)
