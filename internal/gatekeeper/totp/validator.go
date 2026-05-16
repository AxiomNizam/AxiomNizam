package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"time"
)

// Validator validates TOTP codes.
type Validator interface {
	Validate(secret []byte, code string, now time.Time, timeStep int) bool
}

// ValidatorImpl validates TOTP codes using RFC 6238.
type ValidatorImpl struct{}

// NewValidator creates a new validator.
func NewValidator() Validator {
	return &ValidatorImpl{}
}

// Validate checks if a code is valid for a secret at a given time.
func (v *ValidatorImpl) Validate(secret []byte, code string, now time.Time, timeStep int) bool {
	if code == "" || len(code) != 6 {
		return false
	}

	// Calculate the current time counter (T)
	T := now.Unix() / int64(timeStep)

	// Generate the expected TOTP code
	expectedCode := v.generateCode(secret, T)

	// Compare codes (timing-safe comparison preferred)
	return code == expectedCode
}

// generateCode generates a TOTP code for a given time step.
func (v *ValidatorImpl) generateCode(secret []byte, T int64) string {
	// Convert time counter to bytes
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(T))

	// Generate HMAC-SHA1
	h := hmac.New(sha1.New, secret)
	h.Write(msg)
	digest := h.Sum(nil)

	// Extract 31-bit offset
	offset := digest[len(digest)-1] & 0x0f
	p := binary.BigEndian.Uint32(digest[offset:offset+4]) & 0x7fffffff

	// Generate 6-digit code
	otp := p % 1000000
	return fmt.Sprintf("%06d", otp)
}
