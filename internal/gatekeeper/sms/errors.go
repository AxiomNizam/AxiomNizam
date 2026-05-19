package sms

import "errors"

var (
	// ErrInvalidPhoneNumber indicates an invalid phone number format.
	ErrInvalidPhoneNumber = errors.New("invalid phone number format")
	// ErrSendFailed indicates the SMS send operation failed.
	ErrSendFailed = errors.New("failed to send SMS")
	// ErrCodeExpired indicates the OTP code has expired.
	ErrCodeExpired = errors.New("verification code expired")
	// ErrCodeMismatch indicates the OTP code does not match.
	ErrCodeMismatch = errors.New("verification code does not match")
)
