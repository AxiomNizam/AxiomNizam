package sms

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Service handles SMS-based OTP delivery.
type Service struct {
	provider Provider
	codeTTL  time.Duration
}

// NewService creates a new SMS service.
func NewService(provider Provider) *Service {
	return &Service{
		provider: provider,
		codeTTL:  5 * time.Minute,
	}
}

// SendOTP generates and sends an OTP code to the given phone number.
func (s *Service) SendOTP(ctx context.Context, phoneNumber string) (string, error) {
	code := generateCode(6)
	msg := fmt.Sprintf("Your AxiomNizam verification code is: %s (expires in 5 minutes)", code)
	if err := s.provider.Send(phoneNumber, msg); err != nil {
		return "", fmt.Errorf("sms send failed: %w", err)
	}
	return code, nil
}

// VerifyOTP validates an OTP code. This is a placeholder - real implementation
// would check against the stored code with TTL.
func (s *Service) VerifyOTP(ctx context.Context, phoneNumber, code string) (bool, error) {
	// TODO: Implement actual code verification with storage
	return false, nil
}

func generateCode(length int) string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	digits := make([]byte, length)
	for i := range digits {
		digits[i] = byte('0' + rng.Intn(10))
	}
	return string(digits)
}
