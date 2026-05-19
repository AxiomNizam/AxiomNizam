package email

import "errors"

var (
	// ErrInvalidEmail indicates an invalid email format.
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrSendFailed indicates the email send operation failed.
	ErrSendFailed = errors.New("failed to send email")
	// ErrCodeExpired indicates the OTP code has expired.
	ErrCodeExpired = errors.New("verification code expired")
	// ErrCodeMismatch indicates the OTP code does not match.
	ErrCodeMismatch = errors.New("verification code does not match")
)
