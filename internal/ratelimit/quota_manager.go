package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// QuotaLimit defines quota settings
type QuotaLimit struct {
	RequestsPerSecond int
	RequestsPerMinute int
	RequestsPerHour   int
	BytesPerMinute    int64
	MaxConcurrent     int
}

// UserQuota tracks user quotas
type UserQuota struct {
	UserID             string
	RequestCount       int64
	BytesConsumed      int64
	LastResetTime      time.Time
	ConcurrentRequests int
	Endpoint           string
	DailyLimit         int64
	DailyUsed          int64
	LastDailyResetTime time.Time
}

// QuotaManager manages user quotas
type QuotaManager struct {
	quotas       map[string]*UserQuota
	endpoints    map[string]*QuotaLimit
	mu           sync.RWMutex
	defaultLimit QuotaLimit
}

// NewQuotaManager creates a new quota manager
func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quotas:    make(map[string]*UserQuota),
		endpoints: make(map[string]*QuotaLimit),
		defaultLimit: QuotaLimit{
			RequestsPerSecond: 100,
			RequestsPerMinute: 6000,
			RequestsPerHour:   360000,
			BytesPerMinute:    10 * 1024 * 1024, // 10MB
			MaxConcurrent:     100,
		},
	}
}

// SetEndpointLimit sets rate limit for specific endpoint
func (qm *QuotaManager) SetEndpointLimit(endpoint string, limit QuotaLimit) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.endpoints[endpoint] = &limit
}

// SetUserDailyQuota sets daily quota for user
func (qm *QuotaManager) SetUserDailyQuota(userID string, dailyLimit int64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quota, exists := qm.quotas[userID]
	if !exists {
		quota = &UserQuota{
			UserID:             userID,
			LastResetTime:      time.Now(),
			LastDailyResetTime: time.Now(),
		}
		qm.quotas[userID] = quota
	}

	quota.DailyLimit = dailyLimit
}

// CheckQuota checks if request should be allowed
func (qm *QuotaManager) CheckQuota(userID string, endpoint string, requestSize int64) (allowed bool, remaining int64, err error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quota, exists := qm.quotas[userID]
	if !exists {
		quota = &UserQuota{
			UserID:             userID,
			LastResetTime:      time.Now(),
			LastDailyResetTime: time.Now(),
			Endpoint:           endpoint,
		}
		qm.quotas[userID] = quota
	}

	// Reset daily quota if day has passed
	if time.Since(quota.LastDailyResetTime) > 24*time.Hour {
		quota.DailyUsed = 0
		quota.LastDailyResetTime = time.Now()
	}

	// Check daily limit
	if quota.DailyLimit > 0 && quota.DailyUsed+requestSize > quota.DailyLimit {
		return false, 0, fmt.Errorf("daily quota exceeded: %d/%d bytes", quota.DailyUsed, quota.DailyLimit)
	}

	// Check byte limit per minute
	limit := qm.getLimit(endpoint)
	if quota.BytesConsumed+requestSize > limit.BytesPerMinute {
		return false, limit.BytesPerMinute - quota.BytesConsumed, fmt.Errorf("byte rate limit exceeded")
	}

	// Check concurrent requests
	if quota.ConcurrentRequests >= int64(limit.MaxConcurrent) {
		return false, 0, fmt.Errorf("max concurrent requests exceeded: %d", limit.MaxConcurrent)
	}

	quota.BytesConsumed += requestSize
	quota.DailyUsed += requestSize
	quota.ConcurrentRequests++
	quota.RequestCount++

	return true, limit.BytesPerMinute - quota.BytesConsumed, nil
}

// ReleaseQuota releases concurrent request count
func (qm *QuotaManager) ReleaseQuota(userID string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if quota, exists := qm.quotas[userID]; exists && quota.ConcurrentRequests > 0 {
		quota.ConcurrentRequests--
	}
}

// GetQuotaStatus returns quota status for user
func (qm *QuotaManager) GetQuotaStatus(userID string) map[string]interface{} {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quota, exists := qm.quotas[userID]
	if !exists {
		return map[string]interface{}{
			"status": "no_quota",
		}
	}

	return map[string]interface{}{
		"user_id":             userID,
		"total_requests":      quota.RequestCount,
		"daily_limit":         quota.DailyLimit,
		"daily_used":          quota.DailyUsed,
		"daily_remaining":     quota.DailyLimit - quota.DailyUsed,
		"bytes_consumed":      quota.BytesConsumed,
		"concurrent_requests": quota.ConcurrentRequests,
		"last_reset":          quota.LastResetTime,
		"last_daily_reset":    quota.LastDailyResetTime,
	}
}

// ResetUserQuota resets user quota
func (qm *QuotaManager) ResetUserQuota(userID string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if quota, exists := qm.quotas[userID]; exists {
		quota.RequestCount = 0
		quota.BytesConsumed = 0
		quota.LastResetTime = time.Now()
	}
}

// getLimit gets rate limit for endpoint
func (qm *QuotaManager) getLimit(endpoint string) QuotaLimit {
	if limit, exists := qm.endpoints[endpoint]; exists {
		return *limit
	}
	return qm.defaultLimit
}

// GetAllUserQuotas returns all user quotas
func (qm *QuotaManager) GetAllUserQuotas() map[string]*UserQuota {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	result := make(map[string]*UserQuota)
	for k, v := range qm.quotas {
		result[k] = v
	}
	return result
}
