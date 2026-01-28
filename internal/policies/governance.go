package policies

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ResourceQuota defines resource limits (Kubernetes-style resource quotas)
type ResourceQuota struct {
	Name      string
	Namespace string
	Hard      map[string]int64 // resource -> limit
	Used      map[string]int64 // resource -> usage
	Status    string           // Active, TerminatingNamespace
	Updated   time.Time
}

// QuotaManager manages resource quotas
type QuotaManager struct {
	mu     sync.RWMutex
	quotas map[string]*ResourceQuota
}

// NewQuotaManager creates a new quota manager
func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quotas: make(map[string]*ResourceQuota),
	}
}

// CreateQuota creates a new quota
func (qm *QuotaManager) CreateQuota(quota *ResourceQuota) error {
	if quota.Name == "" {
		return fmt.Errorf("quota name required")
	}

	qm.mu.Lock()
	defer qm.mu.Unlock()

	key := fmt.Sprintf("%s/%s", quota.Namespace, quota.Name)
	if _, exists := qm.quotas[key]; exists {
		return fmt.Errorf("quota already exists")
	}

	quota.Status = "Active"
	quota.Updated = time.Now()
	if quota.Hard == nil {
		quota.Hard = make(map[string]int64)
	}
	if quota.Used == nil {
		quota.Used = make(map[string]int64)
	}

	qm.quotas[key] = quota
	return nil
}

// GetQuota returns a quota
func (qm *QuotaManager) GetQuota(namespace, name string) *ResourceQuota {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	key := fmt.Sprintf("%s/%s", namespace, name)
	return qm.quotas[key]
}

// CheckQuota checks if resource can be allocated
func (qm *QuotaManager) CheckQuota(namespace, resource string, amount int64) error {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	// Check all quotas in namespace
	for key, quota := range qm.quotas {
		if quota.Namespace != namespace {
			continue
		}

		if hard, ok := quota.Hard[resource]; ok {
			used := quota.Used[resource]
			if used+amount > hard {
				return fmt.Errorf("quota exceeded for %s in %s: %d/%d",
					resource, key, used+amount, hard)
			}
		}
	}

	return nil
}

// AllocateQuota allocates a resource quota
func (qm *QuotaManager) AllocateQuota(namespace, resource string, amount int64) error {
	if err := qm.CheckQuota(namespace, resource, amount); err != nil {
		return err
	}

	qm.mu.Lock()
	defer qm.mu.Unlock()

	for _, quota := range qm.quotas {
		if quota.Namespace == namespace {
			quota.Used[resource] += amount
			quota.Updated = time.Now()
		}
	}

	return nil
}

// ReleaseQuota releases a resource quota
func (qm *QuotaManager) ReleaseQuota(namespace, resource string, amount int64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	for _, quota := range qm.quotas {
		if quota.Namespace == namespace {
			quota.Used[resource] -= amount
			if quota.Used[resource] < 0 {
				quota.Used[resource] = 0
			}
			quota.Updated = time.Now()
		}
	}
}

// LimitRange defines per-resource limits (Kubernetes LimitRange)
type LimitRange struct {
	Name      string
	Namespace string
	Limits    []LimitRangeItem
}

// LimitRangeItem defines limits for a resource type
type LimitRangeItem struct {
	Type    string // Pod, Container, PersistentVolumeClaim
	Max     map[string]int64
	Min     map[string]int64
	Default map[string]int64
}

// LimitRangeManager validates resource limits
type LimitRangeManager struct {
	mu     sync.RWMutex
	ranges map[string]*LimitRange
}

// NewLimitRangeManager creates a new limit range manager
func NewLimitRangeManager() *LimitRangeManager {
	return &LimitRangeManager{
		ranges: make(map[string]*LimitRange),
	}
}

// AddLimitRange adds a limit range
func (lrm *LimitRangeManager) AddLimitRange(lr *LimitRange) error {
	if lr.Name == "" {
		return fmt.Errorf("limit range name required")
	}

	lrm.mu.Lock()
	defer lrm.mu.Unlock()

	key := fmt.Sprintf("%s/%s", lr.Namespace, lr.Name)
	lrm.ranges[key] = lr
	return nil
}

// ValidateLimits validates resource against limit ranges
func (lrm *LimitRangeManager) ValidateLimits(namespace, resourceType string, values map[string]int64) error {
	lrm.mu.RLock()
	ranges := lrm.ranges
	lrm.mu.RUnlock()

	for _, limitRange := range ranges {
		if limitRange.Namespace != namespace {
			continue
		}

		for _, item := range limitRange.Limits {
			if item.Type != resourceType {
				continue
			}

			for resource, value := range values {
				if max, ok := item.Max[resource]; ok && value > max {
					return fmt.Errorf("exceeds max %s limit: %d > %d",
						resource, value, max)
				}
				if min, ok := item.Min[resource]; ok && value < min {
					return fmt.Errorf("below min %s limit: %d < %d",
						resource, value, min)
				}
			}
		}
	}

	return nil
}

// NetworkPolicy defines network access (Kubernetes NetworkPolicy)
type NetworkPolicy struct {
	Name        string
	Namespace   string
	PodSelector map[string]string
	PolicyTypes []string // Ingress, Egress
	Ingress     []NetworkPolicyRule
	Egress      []NetworkPolicyRule
}

// NetworkPolicyRule defines network rules
type NetworkPolicyRule struct {
	From  []NetworkPolicyPeer
	Ports []NetworkPolicyPort
}

// NetworkPolicyPeer defines source/destination
type NetworkPolicyPeer struct {
	PodSelector       map[string]string
	NamespaceSelector map[string]string
}

// NetworkPolicyPort defines port rules
type NetworkPolicyPort struct {
	Protocol string
	Port     int
}

// NetworkPolicyManager enforces network policies
type NetworkPolicyManager struct {
	mu       sync.RWMutex
	policies map[string]*NetworkPolicy
}

// NewNetworkPolicyManager creates a new network policy manager
func NewNetworkPolicyManager() *NetworkPolicyManager {
	return &NetworkPolicyManager{
		policies: make(map[string]*NetworkPolicy),
	}
}

// AddNetworkPolicy adds a network policy
func (npm *NetworkPolicyManager) AddNetworkPolicy(policy *NetworkPolicy) error {
	if policy.Name == "" {
		return fmt.Errorf("network policy name required")
	}

	npm.mu.Lock()
	defer npm.mu.Unlock()

	key := fmt.Sprintf("%s/%s", policy.Namespace, policy.Name)
	npm.policies[key] = policy
	return nil
}

// IsConnectionAllowed checks if connection is allowed
func (npm *NetworkPolicyManager) IsConnectionAllowed(
	namespace string,
	sourcePod, destPod map[string]string,
	port int, protocol string) bool {

	npm.mu.RLock()
	defer npm.mu.RUnlock()

	// Find policies for destination pod
	for _, policy := range npm.policies {
		if policy.Namespace != namespace {
			continue
		}

		// Check if policy applies to dest pod
		if !matchesSelector(destPod, policy.PodSelector) {
			continue
		}

		// Check ingress rules
		if len(policy.Ingress) == 0 {
			continue // No ingress rules = allow all
		}

		allowed := false
		for _, rule := range policy.Ingress {
			if matchesRule(rule, sourcePod, port, protocol) {
				allowed = true
				break
			}
		}

		if !allowed {
			return false
		}
	}

	return true
}

// SecurityPolicy defines security context
type SecurityPolicy struct {
	Name     string
	Kind     string
	Rules    []SecurityPolicyRule
	Priority int
}

// SecurityPolicyRule defines security rules
type SecurityPolicyRule struct {
	Users         []string
	Verbs         []string
	Resources     []string
	ResourceNames []string
	Restrictions  map[string]interface{}
}

// SecurityPolicyManager enforces security policies
type SecurityPolicyManager struct {
	mu       sync.RWMutex
	policies []*SecurityPolicy
}

// NewSecurityPolicyManager creates a new security policy manager
func NewSecurityPolicyManager() *SecurityPolicyManager {
	return &SecurityPolicyManager{
		policies: make([]*SecurityPolicy, 0),
	}
}

// AddSecurityPolicy adds a security policy
func (spm *SecurityPolicyManager) AddSecurityPolicy(policy *SecurityPolicy) error {
	if policy.Name == "" {
		return fmt.Errorf("security policy name required")
	}

	spm.mu.Lock()
	defer spm.mu.Unlock()

	spm.policies = append(spm.policies, policy)
	return nil
}

// CanPerformAction checks if action is allowed
func (spm *SecurityPolicyManager) CanPerformAction(user string, verb, resource string) bool {
	spm.mu.RLock()
	defer spm.mu.RUnlock()

	for _, policy := range spm.policies {
		for _, rule := range policy.Rules {
			userMatch := len(rule.Users) == 0
			for _, u := range rule.Users {
				if u == user || u == "*" {
					userMatch = true
					break
				}
			}

			verbMatch := len(rule.Verbs) == 0
			for _, v := range rule.Verbs {
				if v == verb || v == "*" {
					verbMatch = true
					break
				}
			}

			resourceMatch := len(rule.Resources) == 0
			for _, r := range rule.Resources {
				if r == resource || r == "*" {
					resourceMatch = true
					break
				}
			}

			if userMatch && verbMatch && resourceMatch {
				return true
			}
		}
	}

	return false
}

// RateLimitPolicy defines rate limiting
type RateLimitPolicy struct {
	Name      string
	Kind      string // User, IP, ResourceType
	Requests  int64
	Window    time.Duration
	BurstSize int64
}

// RateLimiter enforces rate limits
type RateLimiter struct {
	mu       sync.RWMutex
	counters map[string]*int64 // key -> request count
	policies map[string]*RateLimitPolicy
	windows  map[string]time.Time // key -> window start
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		counters: make(map[string]*int64),
		policies: make(map[string]*RateLimitPolicy),
		windows:  make(map[string]time.Time),
	}
}

// AddPolicy adds a rate limit policy
func (rl *RateLimiter) AddPolicy(policy *RateLimitPolicy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name required")
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.policies[policy.Name] = policy
	return nil
}

// AllowRequest checks if request is allowed under rate limit
func (rl *RateLimiter) AllowRequest(policyName, key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	policy, ok := rl.policies[policyName]
	if !ok {
		return true // No policy = allow
	}

	now := time.Now()
	fullKey := fmt.Sprintf("%s:%s", policyName, key)

	// Check window
	if windowStart, ok := rl.windows[fullKey]; ok {
		if now.Sub(windowStart) > policy.Window {
			// New window
			count := int64(1)
			rl.counters[fullKey] = &count
			rl.windows[fullKey] = now
			return true
		}
	} else {
		// First request
		count := int64(1)
		rl.counters[fullKey] = &count
		rl.windows[fullKey] = now
		return true
	}

	// Check limit
	counter := rl.counters[fullKey]
	current := atomic.AddInt64(counter, 1)
	return current <= policy.Requests
}

// Helper functions
func matchesSelector(pod, selector map[string]string) bool {
	for key, value := range selector {
		if pod[key] != value {
			return false
		}
	}
	return true
}

func matchesRule(rule NetworkPolicyRule, sourcePod map[string]string, port int, protocol string) bool {
	for _, from := range rule.From {
		if matchesSelector(sourcePod, from.PodSelector) {
			for _, p := range rule.Ports {
				if (p.Port == port || p.Port == 0) && (p.Protocol == protocol || p.Protocol == "") {
					return true
				}
			}
		}
	}
	return false
}
