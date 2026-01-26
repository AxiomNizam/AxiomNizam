package quota

import (
	"fmt"
	"sync"
	"time"
)

// QuotaPolicy defines resource quota policies
type QuotaPolicy struct {
	ID             string
	Name           string
	Type           string
	Version        string
	Enabled        bool
	Namespace      string
	ResourceQuotas []ResourceQuota
	ScopeSelector  map[string]string
	Description    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ResourceQuota defines a quota for a specific resource
type ResourceQuota struct {
	Resource   string // "cpu", "memory", "pods", "services", "api-calls", "storage"
	Hard       string // Hard limit
	Used       string // Current usage
	Soft       string // Soft limit (for warnings)
	TimeWindow time.Duration
}

// GetID returns policy ID
func (qp *QuotaPolicy) GetID() string {
	return qp.ID
}

// GetName returns policy name
func (qp *QuotaPolicy) GetName() string {
	return qp.Name
}

// GetType returns policy type
func (qp *QuotaPolicy) GetType() string {
	return qp.Type
}

// GetVersion returns version
func (qp *QuotaPolicy) GetVersion() string {
	return qp.Version
}

// GetEnabled returns if enabled
func (qp *QuotaPolicy) GetEnabled() bool {
	return qp.Enabled
}

// Validate validates the policy
func (qp *QuotaPolicy) Validate() error {
	if qp.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}
	if qp.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}
	if len(qp.ResourceQuotas) == 0 {
		return fmt.Errorf("at least one resource quota must be defined")
	}
	return nil
}

// QuotaManager manages resource quotas
type QuotaManager struct {
	mu               sync.RWMutex
	quotaPolicies    map[string]*QuotaPolicy
	usage            map[string]ResourceUsage
	reservations     map[string]Reservation
	maxRetryAttempts int
}

// ResourceUsage tracks current usage of a resource
type ResourceUsage struct {
	Resource      string
	Namespace     string
	Current       int64
	Hard          int64
	Soft          int64
	LastUpdated   time.Time
	ExceededSince time.Time
	Exceeded      bool
}

// Reservation represents a resource reservation
type Reservation struct {
	ID        string
	Namespace string
	Resource  string
	Amount    int64
	Duration  time.Duration
	CreatedAt time.Time
	ExpiresAt time.Time
}

// NewQuotaManager creates a new quota manager
func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quotaPolicies:    make(map[string]*QuotaPolicy),
		usage:            make(map[string]ResourceUsage),
		reservations:     make(map[string]Reservation),
		maxRetryAttempts: 3,
	}
}

// RegisterQuotaPolicy registers a quota policy
func (qm *QuotaManager) RegisterQuotaPolicy(policy *QuotaPolicy) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if err := policy.Validate(); err != nil {
		return err
	}

	qm.quotaPolicies[policy.ID] = policy
	return nil
}

// AllocateResource allocates resources
func (qm *QuotaManager) AllocateResource(namespace, resource string, amount int64) (bool, error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", namespace, resource)
	usage, exists := qm.usage[key]

	if !exists {
		// Initialize usage
		usage = ResourceUsage{
			Resource:    resource,
			Namespace:   namespace,
			Current:     0,
			LastUpdated: time.Now(),
		}
	}

	// Check if allocation would exceed hard limit
	if usage.Hard > 0 && usage.Current+amount > usage.Hard {
		return false, fmt.Errorf("quota exceeded for resource %s in namespace %s", resource, namespace)
	}

	// Check soft limit
	if usage.Soft > 0 && usage.Current+amount > usage.Soft {
		if usage.ExceededSince.IsZero() {
			usage.ExceededSince = time.Now()
		}
	}

	usage.Current += amount
	usage.LastUpdated = time.Now()
	usage.Exceeded = usage.Hard > 0 && usage.Current > usage.Hard
	qm.usage[key] = usage

	return true, nil
}

// ReleaseResource releases allocated resources
func (qm *QuotaManager) ReleaseResource(namespace, resource string, amount int64) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", namespace, resource)
	usage, exists := qm.usage[key]

	if !exists {
		return fmt.Errorf("no usage record for %s in namespace %s", resource, namespace)
	}

	if usage.Current < amount {
		return fmt.Errorf("cannot release more than allocated")
	}

	usage.Current -= amount
	usage.LastUpdated = time.Now()
	usage.Exceeded = usage.Hard > 0 && usage.Current > usage.Hard
	qm.usage[key] = usage

	return nil
}

// ReserveResource reserves resources for future use
func (qm *QuotaManager) ReserveResource(namespace, resource string, amount int64, duration time.Duration) (string, error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	// Check quota first
	key := fmt.Sprintf("%s:%s", namespace, resource)
	usage, exists := qm.usage[key]

	if !exists {
		usage = ResourceUsage{
			Resource:  resource,
			Namespace: namespace,
			Current:   0,
		}
	}

	if usage.Hard > 0 && usage.Current+amount > usage.Hard {
		return "", fmt.Errorf("insufficient quota for reservation")
	}

	reservationID := fmt.Sprintf("res-%d", time.Now().UnixNano())
	reservation := Reservation{
		ID:        reservationID,
		Namespace: namespace,
		Resource:  resource,
		Amount:    amount,
		Duration:  duration,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}

	qm.reservations[reservationID] = reservation
	return reservationID, nil
}

// GetUsage gets current usage
func (qm *QuotaManager) GetUsage(namespace, resource string) (ResourceUsage, error) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", namespace, resource)
	usage, exists := qm.usage[key]

	if !exists {
		return ResourceUsage{}, fmt.Errorf("no usage found for %s in %s", resource, namespace)
	}

	return usage, nil
}

// GetQuotaPolicies returns all quota policies for a namespace
func (qm *QuotaManager) GetQuotaPolicies(namespace string) []*QuotaPolicy {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	var policies []*QuotaPolicy
	for _, policy := range qm.quotaPolicies {
		if policy.Namespace == namespace || policy.Namespace == "" {
			policies = append(policies, policy)
		}
	}

	return policies
}

// CleanupExpiredReservations cleans up expired reservations
func (qm *QuotaManager) CleanupExpiredReservations() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	now := time.Now()
	for id, reservation := range qm.reservations {
		if now.After(reservation.ExpiresAt) {
			delete(qm.reservations, id)
		}
	}
}

// NamespaceQuotaScope defines quota scope for a namespace
type NamespaceQuotaScope struct {
	Namespace      string
	SoftQuotas     map[string]int64
	HardQuotas     map[string]int64
	TimeWindow     time.Duration
	AlertThreshold float64 // e.g., 0.8 for 80%
}

// NewNamespaceQuotaScope creates a new namespace quota scope
func NewNamespaceQuotaScope(namespace string) *NamespaceQuotaScope {
	return &NamespaceQuotaScope{
		Namespace:      namespace,
		SoftQuotas:     make(map[string]int64),
		HardQuotas:     make(map[string]int64),
		TimeWindow:     24 * time.Hour,
		AlertThreshold: 0.8,
	}
}

// SetSoftQuota sets a soft quota
func (nqs *NamespaceQuotaScope) SetSoftQuota(resource string, limit int64) {
	nqs.SoftQuotas[resource] = limit
}

// SetHardQuota sets a hard quota
func (nqs *NamespaceQuotaScope) SetHardQuota(resource string, limit int64) {
	nqs.HardQuotas[resource] = limit
}

// PriorityBasedQuota assigns different quotas based on priority
type PriorityBasedQuota struct {
	Priority       int // Higher number = higher priority
	ResourceLimits map[string]int64
	Description    string
}

// QuotaScheduler schedules quota assignments
type QuotaScheduler struct {
	mu        sync.RWMutex
	schedules map[string]QuotaSchedule
}

// QuotaSchedule defines a scheduled quota change
type QuotaSchedule struct {
	ID          string
	Namespace   string
	Resource    string
	NewQuota    int64
	StartTime   time.Time
	EndTime     time.Time
	Recurring   string // "daily", "weekly", "monthly"
	Description string
}

// NewQuotaScheduler creates a new quota scheduler
func NewQuotaScheduler() *QuotaScheduler {
	return &QuotaScheduler{
		schedules: make(map[string]QuotaSchedule),
	}
}

// ScheduleQuotaChange schedules a quota change
func (qs *QuotaScheduler) ScheduleQuotaChange(schedule QuotaSchedule) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.schedules[schedule.ID] = schedule
}

// GetActiveSchedules gets currently active schedules
func (qs *QuotaScheduler) GetActiveSchedules() []QuotaSchedule {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	now := time.Now()
	var active []QuotaSchedule

	for _, schedule := range qs.schedules {
		if now.After(schedule.StartTime) && now.Before(schedule.EndTime) {
			active = append(active, schedule)
		}
	}

	return active
}
