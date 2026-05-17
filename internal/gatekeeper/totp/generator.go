package totp

import (
	"crypto/rand"
)

// SecretGenerator generates random TOTP secrets.
type SecretGenerator interface {
	Generate() ([]byte, error)
}

// GeneratorImpl generates random secrets.
type GeneratorImpl struct{}

// NewSecretGenerator creates a new secret generator.
func NewSecretGenerator() SecretGenerator {
	return &GeneratorImpl{}
}

// Generate generates a 32-byte random secret.
func (g *GeneratorImpl) Generate() ([]byte, error) {
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	return secret, err
}
