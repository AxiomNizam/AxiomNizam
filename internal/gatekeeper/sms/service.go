package sms

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// pendingCode tracks an OTP waiting for verification.
type pendingCode struct {
	code      string
	expiresAt time.Time
}

// Service handles SMS-based OTP delivery.
type Service struct {
	provider Provider
	codeTTL  time.Duration

	mu     sync.RWMutex
	codes  map[string]*pendingCode // key = phoneNumber
}

// NewService creates a new SMS service.
func NewService(provider Provider) *Service {
	return &Service{
		provider: provider,
		codeTTL:  5 * time.Minute,
		codes:    make(map[string]*pendingCode),
	}
}

// SendOTP generates and sends an OTP code to the given phone number.
// The code is stored in memory for later verification.
func (s *Service) SendOTP(ctx context.Context, phoneNumber string) (string, error) {
	code := generateCode(6)
	msg := fmt.Sprintf("Your AxiomNizam verification code is: %s (expires in 5 minutes)", code)
	if err := s.provider.Send(phoneNumber, msg); err != nil {
		return "", fmt.Errorf("sms send failed: %w", err)
	}

	s.mu.Lock()
	s.codes[phoneNumber] = &pendingCode{
		code:      code,
		expiresAt: time.Now().Add(s.codeTTL),
	}
	s.mu.Unlock()

	return code, nil
}

// VerifyOTP validates an OTP code against the stored code for the phone number.
// Returns true if the code matches and has not expired.
func (s *Service) VerifyOTP(ctx context.Context, phoneNumber, code string) (bool, error) {
	s.mu.Lock()
	pending, exists := s.codes[phoneNumber]
	if exists {
		delete(s.codes, phoneNumber) // single use
	}
	s.mu.Unlock()

	if !exists {
		return false, nil
	}

	if time.Now().After(pending.expiresAt) {
		return false, nil
	}

	return pending.code == code, nil
}

func generateCode(length int) string {
	digits := make([]byte, length)
	for i := range digits {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		digits[i] = byte('0' + n.Int64())
	}
	return string(digits)
}
