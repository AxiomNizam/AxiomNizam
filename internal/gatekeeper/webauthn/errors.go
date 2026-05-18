package webauthn

import "errors"

var (
	// ErrNotImplemented indicates WebAuthn is not yet implemented.
	ErrNotImplemented = errors.New("webauthn not implemented")
	// ErrInvalidCredential indicates an invalid credential.
	ErrInvalidCredential = errors.New("invalid credential")
	// ErrChallengeExpired indicates the challenge has expired.
	ErrChallengeExpired = errors.New("challenge expired")
	// ErrUserNotFound indicates the user was not found.
	ErrUserNotFound = errors.New("user not found")
)
