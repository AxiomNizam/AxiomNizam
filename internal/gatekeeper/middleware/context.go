package middleware

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	// UserIDKey is the context key for the authenticated user ID.
	UserIDKey contextKey = "user_id"
	// FactorIDKey is the context key for the active factor ID.
	FactorIDKey contextKey = "factor_id"
	// ChallengeIDKey is the context key for the active challenge ID.
	ChallengeIDKey contextKey = "challenge_id"
)

// WithUserID adds a user ID to the context.
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID extracts the user ID from the context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

// WithFactorID adds a factor ID to the context.
func WithFactorID(ctx context.Context, factorID uuid.UUID) context.Context {
	return context.WithValue(ctx, FactorIDKey, factorID)
}

// GetFactorID extracts the factor ID from the context.
func GetFactorID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(FactorIDKey).(uuid.UUID)
	return id, ok
}

// WithChallengeID adds a challenge ID to the context.
func WithChallengeID(ctx context.Context, challengeID uuid.UUID) context.Context {
	return context.WithValue(ctx, ChallengeIDKey, challengeID)
}

// GetChallengeID extracts the challenge ID from the context.
func GetChallengeID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ChallengeIDKey).(uuid.UUID)
	return id, ok
}
