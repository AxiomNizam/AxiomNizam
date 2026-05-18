package email

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Service handles email-based OTP delivery.
type Service struct {
	provider Provider
	codeTTL  time.Duration
}

// NewService creates a new email service.
func NewService(provider Provider) *Service {
	return &Service{
		provider: provider,
		codeTTL:  5 * time.Minute,
	}
}

// SendOTP generates and sends an OTP code to the given email address.
func (s *Service) SendOTP(ctx context.Context, emailAddress string) (string, error) {
	code := generateCode(6)
	subject := "Your AxiomNizam Verification Code"
	body := fmt.Sprintf("Your verification code is: %s\n\nThis code expires in 5 minutes.", code)
	if err := s.provider.Send(emailAddress, subject, body); err != nil {
		return "", fmt.Errorf("email send failed: %w", err)
	}
	return code, nil
}

// VerifyOTP validates an OTP code. This is a placeholder - real implementation
// would check against the stored code with TTL.
func (s *Service) VerifyOTP(ctx context.Context, emailAddress, code string) (bool, error) {
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
