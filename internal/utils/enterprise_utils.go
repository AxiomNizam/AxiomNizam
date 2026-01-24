package utils

import (
	"fmt"
	"sync"
	"time"
)

// EnterpriseCache provides enterprise-grade caching
type EnterpriseCache struct {
	mu   sync.RWMutex
	data map[string]CacheItem
	ttl  time.Duration
}

// CacheItem holds cached data with metadata
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
	Created   time.Time
	Accessed  time.Time
	Hits      int64
}

// NewEnterpriseCache creates a new enterprise cache
func NewEnterpriseCache(defaultTTL time.Duration) *EnterpriseCache {
	return &EnterpriseCache{
		data: make(map[string]CacheItem),
		ttl:  defaultTTL,
	}
}

// Set sets a cache item
func (ec *EnterpriseCache) Set(key string, value interface{}) {
	ec.SetWithTTL(key, value, ec.ttl)
}

// SetWithTTL sets a cache item with custom TTL
func (ec *EnterpriseCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.data[key] = CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		Created:   time.Now(),
		Accessed:  time.Now(),
		Hits:      0,
	}
}

// Get retrieves a cache item
func (ec *EnterpriseCache) Get(key string) (interface{}, bool) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	item, exists := ec.data[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		delete(ec.data, key)
		return nil, false
	}

	item.Hits++
	item.Accessed = time.Now()
	ec.data[key] = item

	return item.Value, true
}

// Delete deletes a cache item
func (ec *EnterpriseCache) Delete(key string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	delete(ec.data, key)
}

// Clear clears all cache items
func (ec *EnterpriseCache) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.data = make(map[string]CacheItem)
}

// Stats returns cache statistics
func (ec *EnterpriseCache) Stats() map[string]interface{} {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	totalHits := int64(0)
	for _, item := range ec.data {
		totalHits += item.Hits
	}

	return map[string]interface{}{
		"items":      len(ec.data),
		"total_hits": totalHits,
	}
}

// TokenManager manages API tokens
type TokenManager struct {
	mu     sync.RWMutex
	tokens map[string]TokenInfo
}

// TokenInfo holds token information
type TokenInfo struct {
	Token     string
	ExpiresAt time.Time
	Scopes    []string
	UserID    string
	Created   time.Time
	LastUsed  time.Time
	Active    bool
}

// NewTokenManager creates a new token manager
func NewTokenManager() *TokenManager {
	return &TokenManager{
		tokens: make(map[string]TokenInfo),
	}
}

// CreateToken creates a new token
func (tm *TokenManager) CreateToken(userID string, scopes []string, duration time.Duration) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	helper := NewCryptographicHelper()
	token, err := helper.GenerateSecureToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	tm.tokens[token] = TokenInfo{
		Token:     token,
		ExpiresAt: time.Now().Add(duration),
		Scopes:    scopes,
		UserID:    userID,
		Created:   time.Now(),
		LastUsed:  time.Now(),
		Active:    true,
	}

	return token, nil
}

// ValidateToken validates a token
func (tm *TokenManager) ValidateToken(token string) (bool, *TokenInfo) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	info, exists := tm.tokens[token]
	if !exists {
		return false, nil
	}

	if !info.Active || time.Now().After(info.ExpiresAt) {
		return false, nil
	}

	return true, &info
}

// RevokeToken revokes a token
func (tm *TokenManager) RevokeToken(token string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	info, exists := tm.tokens[token]
	if !exists {
		return fmt.Errorf("token not found")
	}

	info.Active = false
	tm.tokens[token] = info
	return nil
}

// QuotaManager manages resource quotas
type QuotaManager struct {
	mu     sync.RWMutex
	quotas map[string]QuotaInfo
}

// QuotaInfo holds quota information
type QuotaInfo struct {
	ResourceID     string
	Limit          int64
	Used           int64
	ResetTime      time.Time
	AlertThreshold float64
}

// NewQuotaManager creates a new quota manager
func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quotas: make(map[string]QuotaInfo),
	}
}

// SetQuota sets a resource quota
func (qm *QuotaManager) SetQuota(resourceID string, limit int64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.quotas[resourceID] = QuotaInfo{
		ResourceID:     resourceID,
		Limit:          limit,
		Used:           0,
		ResetTime:      time.Now().Add(24 * time.Hour),
		AlertThreshold: 0.8,
	}
}

// IncrementUsage increments resource usage
func (qm *QuotaManager) IncrementUsage(resourceID string, amount int64) (bool, error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quota, exists := qm.quotas[resourceID]
	if !exists {
		return false, fmt.Errorf("quota not found for resource: %s", resourceID)
	}

	if quota.Used+amount > quota.Limit {
		return false, fmt.Errorf("quota exceeded for resource: %s", resourceID)
	}

	quota.Used += amount
	qm.quotas[resourceID] = quota
	return true, nil
}

// GetUsage returns current usage for a resource
func (qm *QuotaManager) GetUsage(resourceID string) (int64, int64, error) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quota, exists := qm.quotas[resourceID]
	if !exists {
		return 0, 0, fmt.Errorf("quota not found for resource: %s", resourceID)
	}

	return quota.Used, quota.Limit, nil
}

// ComplianceChecker checks compliance with policies
type ComplianceChecker struct {
	rules map[string]ComplianceRule
}

// ComplianceRule defines a compliance rule
type ComplianceRule struct {
	ID          string
	Name        string
	Description string
	CheckFunc   func() bool
	Severity    string // critical, high, medium, low
	Category    string // security, performance, reliability
}

// NewComplianceChecker creates a new compliance checker
func NewComplianceChecker() *ComplianceChecker {
	return &ComplianceChecker{
		rules: make(map[string]ComplianceRule),
	}
}

// AddRule adds a compliance rule
func (cc *ComplianceChecker) AddRule(rule ComplianceRule) {
	cc.rules[rule.ID] = rule
}

// RunChecks runs all compliance checks
func (cc *ComplianceChecker) RunChecks() []ComplianceResult {
	var results []ComplianceResult

	for _, rule := range cc.rules {
		passed := rule.CheckFunc()
		results = append(results, ComplianceResult{
			RuleID:   rule.ID,
			RuleName: rule.Name,
			Passed:   passed,
			Severity: rule.Severity,
			Category: rule.Category,
		})
	}

	return results
}

// ComplianceResult holds a compliance check result
type ComplianceResult struct {
	RuleID   string
	RuleName string
	Passed   bool
	Severity string
	Category string
}

// AuditTrail records system events for audit purposes
type AuditTrail struct {
	mu      sync.RWMutex
	events  []AuditEvent
	maxSize int
}

// AuditEvent represents an auditable event
type AuditEvent struct {
	Timestamp    time.Time
	EventType    string
	UserID       string
	Action       string
	ResourceType string
	ResourceID   string
	Status       string // success, failure
	Details      map[string]interface{}
}

// NewAuditTrail creates a new audit trail
func NewAuditTrail(maxSize int) *AuditTrail {
	return &AuditTrail{
		events:  make([]AuditEvent, 0),
		maxSize: maxSize,
	}
}

// LogEvent logs an audit event
func (at *AuditTrail) LogEvent(event AuditEvent) {
	at.mu.Lock()
	defer at.mu.Unlock()

	event.Timestamp = time.Now()
	at.events = append(at.events, event)

	// Keep only the most recent events
	if len(at.events) > at.maxSize {
		at.events = at.events[len(at.events)-at.maxSize:]
	}
}

// GetEvents returns all audit events
func (at *AuditTrail) GetEvents() []AuditEvent {
	at.mu.RLock()
	defer at.mu.RUnlock()

	// Return a copy to avoid external modifications
	events := make([]AuditEvent, len(at.events))
	copy(events, at.events)
	return events
}

// ConfigManager manages application configuration
type ConfigManager struct {
	mu     sync.RWMutex
	config map[string]interface{}
}

// NewConfigManager creates a new config manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		config: make(map[string]interface{}),
	}
}

// Set sets a configuration value
func (cm *ConfigManager) Set(key string, value interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config[key] = value
}

// Get gets a configuration value
func (cm *ConfigManager) Get(key string) (interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	value, exists := cm.config[key]
	return value, exists
}

// GetString gets a string configuration value
func (cm *ConfigManager) GetString(key string) string {
	value, exists := cm.Get(key)
	if !exists {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// GetInt gets an integer configuration value
func (cm *ConfigManager) GetInt(key string) int {
	value, exists := cm.Get(key)
	if !exists {
		return 0
	}
	if num, ok := value.(int); ok {
		return num
	}
	return 0
}

// GetBool gets a boolean configuration value
func (cm *ConfigManager) GetBool(key string) bool {
	value, exists := cm.Get(key)
	if !exists {
		return false
	}
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

// GetAll returns all configuration
func (cm *ConfigManager) GetAll() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy
	config := make(map[string]interface{})
	for k, v := range cm.config {
		config[k] = v
	}
	return config
}

// AcquireConnection acquires a connection from the pool
type RequestIDGenerator struct {
	mu      sync.Mutex
	counter int64
	prefix  string
}

// NewRequestIDGenerator creates a new request ID generator
func NewRequestIDGenerator(prefix string) *RequestIDGenerator {
	return &RequestIDGenerator{
		counter: 0,
		prefix:  prefix,
	}
}

// GenerateID generates a new request ID
func (rig *RequestIDGenerator) GenerateID() string {
	rig.mu.Lock()
	defer rig.mu.Unlock()

	rig.counter++
	return fmt.Sprintf("%s-%d-%d", rig.prefix, time.Now().UnixNano(), rig.counter)
}

// FeatureFlagManager manages feature flags
type FeatureFlagManager struct {
	mu    sync.RWMutex
	flags map[string]FeatureFlag
}

// FeatureFlag defines a feature flag
type FeatureFlag struct {
	Name        string
	Enabled     bool
	Percentage  int // 0-100 for gradual rollout
	UserGroups  []string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewFeatureFlagManager creates a new feature flag manager
func NewFeatureFlagManager() *FeatureFlagManager {
	return &FeatureFlagManager{
		flags: make(map[string]FeatureFlag),
	}
}

// CreateFlag creates a new feature flag
func (ffm *FeatureFlagManager) CreateFlag(name string, enabled bool) {
	ffm.mu.Lock()
	defer ffm.mu.Unlock()

	ffm.flags[name] = FeatureFlag{
		Name:       name,
		Enabled:    enabled,
		Percentage: 100,
		UserGroups: make([]string, 0),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// IsEnabled checks if a feature flag is enabled
func (ffm *FeatureFlagManager) IsEnabled(name string) bool {
	ffm.mu.RLock()
	defer ffm.mu.RUnlock()

	flag, exists := ffm.flags[name]
	return exists && flag.Enabled
}

// SetEnabled enables or disables a feature flag
func (ffm *FeatureFlagManager) SetEnabled(name string, enabled bool) error {
	ffm.mu.Lock()
	defer ffm.mu.Unlock()

	flag, exists := ffm.flags[name]
	if !exists {
		return fmt.Errorf("feature flag not found: %s", name)
	}

	flag.Enabled = enabled
	flag.UpdatedAt = time.Now()
	ffm.flags[name] = flag
	return nil
}
