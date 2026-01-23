package auth

import (
	"fmt"
	"sync"
	"time"
)

// TokenUsage tracks API call count and validity for each token
type TokenUsage struct {
	Username       string
	CallCount      int64
	IssuedAt       time.Time
	LastUsedAt     time.Time
	MaxCallsPerDay int64 // 500 calls per token
}

// RateLimiter manages rate limiting and token validity
type RateLimiter struct {
	tokens        map[string]*TokenUsage
	mu            sync.RWMutex
	maxCalls      int64         // 500 calls per token
	tokenValidity time.Duration // 10 minutes
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxCalls int64, tokenValidityMinutes int) *RateLimiter {
	rl := &RateLimiter{
		tokens:        make(map[string]*TokenUsage),
		maxCalls:      maxCalls,                                          // 500
		tokenValidity: time.Duration(tokenValidityMinutes) * time.Minute, // 10 minutes
	}

	// Start cleanup goroutine to remove expired tokens
	go rl.cleanupExpiredTokens()

	return rl
}

// RegisterToken registers a new token for tracking
func (rl *RateLimiter) RegisterToken(token string, username string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.tokens[token] = &TokenUsage{
		Username:       username,
		CallCount:      0,
		IssuedAt:       time.Now(),
		LastUsedAt:     time.Now(),
		MaxCallsPerDay: rl.maxCalls,
	}
}

// CheckRateLimit checks if token is valid and has remaining calls
// Returns: (allowed bool, callsRemaining int64, expiresAt time.Time, err error)
func (rl *RateLimiter) CheckRateLimit(token string) (bool, int64, time.Time, error) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	usage, exists := rl.tokens[token]
	if !exists {
		return false, 0, time.Time{}, fmt.Errorf("token not tracked or invalid")
	}

	// Check if token has expired
	expiresAt := usage.IssuedAt.Add(rl.tokenValidity)
	if time.Now().After(expiresAt) {
		return false, 0, expiresAt, fmt.Errorf("token expired")
	}

	// Check if call limit exceeded
	if usage.CallCount >= usage.MaxCallsPerDay {
		return false, 0, expiresAt, fmt.Errorf("api call limit exceeded: %d/%d calls used", usage.CallCount, usage.MaxCallsPerDay)
	}

	// Token is valid, calculate remaining calls
	callsRemaining := usage.MaxCallsPerDay - usage.CallCount

	return true, callsRemaining, expiresAt, nil
}

// IncrementCallCount increments the call count for a token
func (rl *RateLimiter) IncrementCallCount(token string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	usage, exists := rl.tokens[token]
	if !exists {
		return fmt.Errorf("token not tracked")
	}

	usage.CallCount++
	usage.LastUsedAt = time.Now()

	return nil
}

// GetTokenStats returns stats for a token
func (rl *RateLimiter) GetTokenStats(token string) (map[string]interface{}, error) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	usage, exists := rl.tokens[token]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}

	expiresAt := usage.IssuedAt.Add(rl.tokenValidity)
	isExpired := time.Now().After(expiresAt)

	return map[string]interface{}{
		"username":        usage.Username,
		"calls_made":      usage.CallCount,
		"max_calls":       usage.MaxCallsPerDay,
		"calls_remaining": usage.MaxCallsPerDay - usage.CallCount,
		"issued_at":       usage.IssuedAt.Format(time.RFC3339),
		"expires_at":      expiresAt.Format(time.RFC3339),
		"is_expired":      isExpired,
		"time_remaining":  expiresAt.Sub(time.Now()).String(),
		"last_used":       usage.LastUsedAt.Format(time.RFC3339),
	}, nil
}

// RevokeToken removes a token from tracking (for logout)
func (rl *RateLimiter) RevokeToken(token string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.tokens, token)
}

// cleanupExpiredTokens periodically removes expired tokens
func (rl *RateLimiter) cleanupExpiredTokens() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()

		now := time.Now()
		for token, usage := range rl.tokens {
			expiresAt := usage.IssuedAt.Add(rl.tokenValidity)
			if now.After(expiresAt) {
				delete(rl.tokens, token)
			}
		}

		rl.mu.Unlock()
	}
}

// GetActiveTokenCount returns number of active tokens
func (rl *RateLimiter) GetActiveTokenCount() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	count := 0
	now := time.Now()

	for _, usage := range rl.tokens {
		expiresAt := usage.IssuedAt.Add(rl.tokenValidity)
		if now.Before(expiresAt) {
			count++
		}
	}

	return count
}

// GetAllTokenStats returns stats for all active tokens
func (rl *RateLimiter) GetAllTokenStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := make(map[string]interface{})
	activeTokens := 0
	totalCalls := int64(0)
	now := time.Now()

	tokenList := make([]map[string]interface{}, 0)

	for _, usage := range rl.tokens {
		expiresAt := usage.IssuedAt.Add(rl.tokenValidity)
		if now.Before(expiresAt) {
			activeTokens++
			totalCalls += usage.CallCount

			tokenList = append(tokenList, map[string]interface{}{
				"username":        usage.Username,
				"calls_made":      usage.CallCount,
				"calls_remaining": usage.MaxCallsPerDay - usage.CallCount,
				"expires_at":      expiresAt.Format(time.RFC3339),
				"time_remaining":  expiresAt.Sub(now).String(),
			})
		}
	}

	stats["active_tokens"] = activeTokens
	stats["total_api_calls"] = totalCalls
	stats["tokens"] = tokenList

	return stats
}
