package challenge

import "time"

// ChallengeSession tracks the state of an active MFA challenge.
type ChallengeSession struct {
	ChallengeID string
	UserID      string
	FactorID    string
	Phase       string
	Attempts    int
	MaxAttempts int
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

// IsExpired returns true if the challenge has expired.
func (s *ChallengeSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsPending returns true if the challenge is still awaiting verification.
func (s *ChallengeSession) IsPending() bool {
	return s.Phase == "Waiting"
}

// IsVerified returns true if the challenge has been verified.
func (s *ChallengeSession) IsVerified() bool {
	return s.Phase == "Verified"
}

// IsFailed returns true if the challenge has failed.
func (s *ChallengeSession) IsFailed() bool {
	return s.Phase == "Failed"
}

// IsTerminal returns true if the challenge is in a final state.
func (s *ChallengeSession) IsTerminal() bool {
	return s.Phase == "Verified" || s.Phase == "Failed" || s.Phase == "Expired" || s.Phase == "Rejected"
}

// RemainingAttempts returns how many verification attempts are left.
func (s *ChallengeSession) RemainingAttempts() int {
	remaining := s.MaxAttempts - s.Attempts
	if remaining < 0 {
		return 0
	}
	return remaining
}

// TimeRemaining returns the duration until the challenge expires.
func (s *ChallengeSession) TimeRemaining() time.Duration {
	remaining := time.Until(s.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CanVerify returns true if the challenge can still accept verification attempts.
func (s *ChallengeSession) CanVerify() bool {
	return s.IsPending() && !s.IsExpired() && s.RemainingAttempts() > 0
}
