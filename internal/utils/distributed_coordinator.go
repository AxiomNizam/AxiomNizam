package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DistributedCoordinator manages multi-instance coordination
type DistributedCoordinator struct {
	mu              sync.RWMutex
	instanceID      string
	leaderElection  *LeaderElection
	locks           map[string]*DistributedLock
	watches         map[string][]WatchCallback
	heartbeats      map[string]time.Time
	generation      int64
}

// LeaderElection manages leader election
type LeaderElection struct {
	mu              sync.RWMutex
	leaderName      string
	candidates      map[string]*Candidate
	leaseDuration   time.Duration
	renewDeadline   time.Duration
	retryPeriod     time.Duration
	transitionTime  time.Time
}

// Candidate represents a leader candidate
type Candidate struct {
	Name              string
	LastHeartbeat     time.Time
	LeaseExpiration   time.Time
	ObservedRecord    int64
}

// DistributedLock manages distributed locks
type DistributedLock struct {
	mu              sync.RWMutex
	lockName        string
	holders         map[string]*LockHolder
	waiters         []LockWaiter
	lastModified    time.Time
}

// LockHolder represents a lock holder
type LockHolder struct {
	InstanceID      string
	AcquiredAt      time.Time
	ExpiresAt       time.Time
	RenewalCount    int
}

// LockWaiter represents a lock waiter
type LockWaiter struct {
	InstanceID      string
	WaitTime        time.Time
	Priority        int
}

// WatchCallback is called on watched resource changes
type WatchCallback func(event WatchEvent)

// WatchEvent represents a watch event
type WatchEvent struct {
	Type      string // Added, Modified, Deleted
	Resource  *ManagedResource
	Timestamp time.Time
}

// NewDistributedCoordinator creates a new coordinator
func NewDistributedCoordinator(instanceID string) *DistributedCoordinator {
	return &DistributedCoordinator{
		instanceID: instanceID,
		leaderElection: &LeaderElection{
			leaderName:    "",
			candidates:    make(map[string]*Candidate),
			leaseDuration: 30 * time.Second,
			renewDeadline: 25 * time.Second,
			retryPeriod:   5 * time.Second,
		},
		locks:      make(map[string]*DistributedLock),
		watches:    make(map[string][]WatchCallback),
		heartbeats: make(map[string]time.Time),
		generation: 1,
	}
}

// ProposeLeadership proposes this instance as a leader candidate
func (dc *DistributedCoordinator) ProposeLeadership(ctx context.Context, leaseName string) (bool, error) {
	le := dc.leaderElection
	le.mu.Lock()
	defer le.mu.Unlock()

	// Check if there's an active leader
	if le.leaderName != "" {
		if candidate, exists := le.candidates[le.leaderName]; exists {
			if time.Now().Before(candidate.LeaseExpiration) {
				return false, nil // Another leader is active
			}
		}
	}

	// Register as candidate
	candidate := &Candidate{
		Name:            dc.instanceID,
		LastHeartbeat:   time.Now(),
		LeaseExpiration: time.Now().Add(le.leaseDuration),
		ObservedRecord:  le.candidates[dc.instanceID].ObservedRecord + 1,
	}

	le.candidates[dc.instanceID] = candidate
	le.leaderName = dc.instanceID
	le.transitionTime = time.Now()

	return true, nil
}

// RenewLeadership renews the leadership lease
func (dc *DistributedCoordinator) RenewLeadership(ctx context.Context) (bool, error) {
	le := dc.leaderElection
	le.mu.Lock()
	defer le.mu.Unlock()

	if le.leaderName != dc.instanceID {
		return false, fmt.Errorf("not the current leader")
	}

	candidate := le.candidates[dc.instanceID]
	if candidate == nil {
		return false, fmt.Errorf("candidate not found")
	}

	if time.Now().After(candidate.LeaseExpiration) {
		return false, fmt.Errorf("lease expired")
	}

	candidate.LastHeartbeat = time.Now()
	candidate.LeaseExpiration = time.Now().Add(le.leaseDuration)
	candidate.RenewalCount++

	return true, nil
}

// IsLeader checks if this instance is the leader
func (dc *DistributedCoordinator) IsLeader() bool {
	le := dc.leaderElection
	le.mu.RLock()
	defer le.mu.RUnlock()

	return le.leaderName == dc.instanceID
}

// GetLeader returns the current leader name
func (dc *DistributedCoordinator) GetLeader() string {
	le := dc.leaderElection
	le.mu.RLock()
	defer le.mu.RUnlock()

	return le.leaderName
}

// AcquireLock acquires a distributed lock
func (dc *DistributedCoordinator) AcquireLock(ctx context.Context, lockName string, ttl time.Duration) (bool, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	lock, exists := dc.locks[lockName]
	if !exists {
		lock = &DistributedLock{
			lockName: lockName,
			holders:  make(map[string]*LockHolder),
			waiters:  make([]LockWaiter, 0),
		}
		dc.locks[lockName] = lock
	}

	lock.mu.Lock()
	defer lock.mu.Unlock()

	// Check if already held
	if len(lock.holders) > 0 {
		for _, holder := range lock.holders {
			if time.Now().Before(holder.ExpiresAt) {
				// Lock is still held by someone else
				lock.waiters = append(lock.waiters, LockWaiter{
					InstanceID: dc.instanceID,
					WaitTime:   time.Now(),
					Priority:   0,
				})
				return false, nil
			}
		}
	}

	// Acquire lock
	lock.holders[dc.instanceID] = &LockHolder{
		InstanceID:   dc.instanceID,
		AcquiredAt:   time.Now(),
		ExpiresAt:    time.Now().Add(ttl),
		RenewalCount: 0,
	}

	lock.lastModified = time.Now()

	return true, nil
}

// ReleaseLock releases a distributed lock
func (dc *DistributedCoordinator) ReleaseLock(lockName string) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	lock, exists := dc.locks[lockName]
	if !exists {
		return fmt.Errorf("lock not found")
	}

	lock.mu.Lock()
	defer lock.mu.Unlock()

	if _, exists := lock.holders[dc.instanceID]; !exists {
		return fmt.Errorf("not lock holder")
	}

	delete(lock.holders, dc.instanceID)
	lock.lastModified = time.Now()

	// Grant lock to first waiter
	if len(lock.waiters) > 0 {
		waiter := lock.waiters[0]
		lock.waiters = lock.waiters[1:]

		lock.holders[waiter.InstanceID] = &LockHolder{
			InstanceID:   waiter.InstanceID,
			AcquiredAt:   time.Now(),
			ExpiresAt:    time.Now().Add(30 * time.Second),
			RenewalCount: 0,
		}
	}

	return nil
}

// IsLockHeld checks if a lock is held
func (dc *DistributedCoordinator) IsLockHeld(lockName string) bool {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	lock, exists := dc.locks[lockName]
	if !exists {
		return false
	}

	lock.mu.RLock()
	defer lock.mu.RUnlock()

	for _, holder := range lock.holders {
		if time.Now().Before(holder.ExpiresAt) {
			return true
		}
	}

	return false
}

// Watch watches for changes to a resource
func (dc *DistributedCoordinator) Watch(ctx context.Context, key string, callback WatchCallback) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.watches[key] = append(dc.watches[key], callback)

	return nil
}

// NotifyWatchers notifies all watchers of a change
func (dc *DistributedCoordinator) NotifyWatchers(event WatchEvent) {
	dc.mu.RLock()
	callbacks := dc.watches[event.Resource.Kind]
	dc.mu.RUnlock()

	for _, callback := range callbacks {
		go callback(event)
	}
}

// SendHeartbeat sends a heartbeat
func (dc *DistributedCoordinator) SendHeartbeat(instanceID string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.heartbeats[instanceID] = time.Now()
}

// IsInstanceAlive checks if an instance is alive
func (dc *DistributedCoordinator) IsInstanceAlive(instanceID string, timeout time.Duration) bool {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	lastHeartbeat, exists := dc.heartbeats[instanceID]
	if !exists {
		return false
	}

	return time.Since(lastHeartbeat) < timeout
}

// GetInstanceCount returns the number of active instances
func (dc *DistributedCoordinator) GetInstanceCount(timeout time.Duration) int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, lastHeartbeat := range dc.heartbeats {
		if now.Sub(lastHeartbeat) < timeout {
			count++
		}
	}

	return count
}

// SharedStateManager manages shared state across instances
type SharedStateManager struct {
	mu        sync.RWMutex
	state     map[string]interface{}
	versions  map[string]int64
	observers []StateChangeObserver
}

// StateChangeObserver observes state changes
type StateChangeObserver interface {
	OnStateChanged(key string, oldValue, newValue interface{})
}

// NewSharedStateManager creates a new shared state manager
func NewSharedStateManager() *SharedStateManager {
	return &SharedStateManager{
		state:     make(map[string]interface{}),
		versions:  make(map[string]int64),
		observers: make([]StateChangeObserver, 0),
	}
}

// Set sets a state value
func (ssm *SharedStateManager) Set(key string, value interface{}) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	oldValue := ssm.state[key]
	ssm.state[key] = value
	ssm.versions[key]++

	// Notify observers
	for _, observer := range ssm.observers {
		go observer.OnStateChanged(key, oldValue, value)
	}
}

// Get gets a state value
func (ssm *SharedStateManager) Get(key string) (interface{}, bool) {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	val, exists := ssm.state[key]
	return val, exists
}

// GetVersion gets the version of a state value
func (ssm *SharedStateManager) GetVersion(key string) int64 {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	return ssm.versions[key]
}

// RegisterObserver registers a state change observer
func (ssm *SharedStateManager) RegisterObserver(observer StateChangeObserver) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	ssm.observers = append(ssm.observers, observer)
}

// Delete deletes a state value
func (ssm *SharedStateManager) Delete(key string) {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	oldValue := ssm.state[key]
	delete(ssm.state, key)
	ssm.versions[key]++

	// Notify observers
	for _, observer := range ssm.observers {
		go observer.OnStateChanged(key, oldValue, nil)
	}
}

// CompareAndSwap atomically sets value if current matches expected
func (ssm *SharedStateManager) CompareAndSwap(key string, expected, newValue interface{}) bool {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	current, exists := ssm.state[key]
	if !exists {
		if expected == nil {
			ssm.state[key] = newValue
			ssm.versions[key]++
			return true
		}
		return false
	}

	if current == expected {
		ssm.state[key] = newValue
		ssm.versions[key]++
		return true
	}

	return false
}
