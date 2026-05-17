package totp

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// RecoveryCodeGenerator creates one-time backup codes.
type RecoveryCodeGenerator struct{}

func NewRecoveryCodeGenerator() *RecoveryCodeGenerator {
	return &RecoveryCodeGenerator{}
}

func (r *RecoveryCodeGenerator) Generate(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := generateSecureCode()
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	return codes, nil
}

func (r *RecoveryCodeGenerator) IsValid(code string) bool {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 12 {
		return false
	}
	// Must be alphanumeric only
	for _, c := range code {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

func generateSecureCode() (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%02X%02X-%02X%02X-%02X%02X", bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5]), nil
}