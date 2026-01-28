package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ObservationPoint records resource observations for debugging
type ObservationPoint struct {
	Timestamp  time.Time
	Phase      string // pending, running, succeeded, failed
	Message    string
	Details    map[string]interface{}
	Generation int64
	Checksum   string // fingerprint of resource state
}

// ResourceObserver tracks resource state changes (for debugging and troubleshooting)
type ResourceObserver struct {
	mu            sync.RWMutex
	observations  map[string][]ObservationPoint // resource -> observations
	maxHistoryLen int
}

// NewResourceObserver creates a new resource observer
func NewResourceObserver(maxHistory int) *ResourceObserver {
	return &ResourceObserver{
		observations:  make(map[string][]ObservationPoint),
		maxHistoryLen: maxHistory,
	}
}

// Record records an observation
func (ro *ResourceObserver) Record(resourceID string, observation ObservationPoint) {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	observation.Timestamp = time.Now()
	observations := ro.observations[resourceID]
	observations = append(observations, observation)

	// Limit history
	if len(observations) > ro.maxHistoryLen {
		observations = observations[len(observations)-ro.maxHistoryLen:]
	}

	ro.observations[resourceID] = observations
}

// GetObservations returns observation history
func (ro *ResourceObserver) GetObservations(resourceID string) []ObservationPoint {
	ro.mu.RLock()
	defer ro.mu.RUnlock()

	obs := ro.observations[resourceID]
	result := make([]ObservationPoint, len(obs))
	copy(result, obs)
	return result
}

// GetLastObservation returns the most recent observation
func (ro *ResourceObserver) GetLastObservation(resourceID string) *ObservationPoint {
	ro.mu.RLock()
	defer ro.mu.RUnlock()

	obs := ro.observations[resourceID]
	if len(obs) == 0 {
		return nil
	}
	return &obs[len(obs)-1]
}

// StateTransitionRule defines valid state transitions
type StateTransitionRule struct {
	From  string
	To    string
	Valid bool
}

// StateValidator validates resource state transitions (prevents invalid states)
type StateValidator struct {
	mu    sync.RWMutex
	rules map[string][]string // from -> valid to states
}

// NewStateValidator creates a new state validator
func NewStateValidator() *StateValidator {
	return &StateValidator{
		rules: make(map[string][]string),
	}
}

// AddTransition registers a valid state transition
func (sv *StateValidator) AddTransition(from, to string) {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	sv.rules[from] = append(sv.rules[from], to)
}

// IsValidTransition checks if state transition is allowed
func (sv *StateValidator) IsValidTransition(from, to string) bool {
	sv.mu.RLock()
	defer sv.mu.RUnlock()

	validStates := sv.rules[from]
	for _, state := range validStates {
		if state == to {
			return true
		}
	}
	return false
}

// ResourceLease represents a lease for resource coordination
type ResourceLease struct {
	Name             string
	Holder           string
	AcquireTime      time.Time
	RenewTime        time.Time
	LeaseDuration    time.Duration
	LeaseTransitions int
}

// LeaseManager manages resource leases (for distributed coordination)
type LeaseManager struct {
	mu     sync.RWMutex
	leases map[string]*ResourceLease
}

// NewLeaseManager creates a new lease manager
func NewLeaseManager() *LeaseManager {
	return &LeaseManager{
		leases: make(map[string]*ResourceLease),
	}
}

// AcquireLease tries to acquire a lease
func (lm *LeaseManager) AcquireLease(ctx context.Context, leaseName, holder string, duration time.Duration) (*ResourceLease, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lease, exists := lm.leases[leaseName]

	if exists && lease.RenewTime.Add(lease.LeaseDuration).After(time.Now()) {
		// Lease is still held
		if lease.Holder != holder {
			return nil, fmt.Errorf("lease held by %s", lease.Holder)
		}
		// Renewal by same holder
		lease.RenewTime = time.Now()
		lease.LeaseTransitions++
		return lease, nil
	}

	// Create new lease
	lease = &ResourceLease{
		Name:          leaseName,
		Holder:        holder,
		AcquireTime:   time.Now(),
		RenewTime:     time.Now(),
		LeaseDuration: duration,
	}

	lm.leases[leaseName] = lease
	return lease, nil
}

// ReleaseLease releases a lease
func (lm *LeaseManager) ReleaseLease(leaseName, holder string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lease, exists := lm.leases[leaseName]
	if !exists {
		return fmt.Errorf("lease not found")
	}

	if lease.Holder != holder {
		return fmt.Errorf("lease not held by %s", holder)
	}

	delete(lm.leases, leaseName)
	return nil
}

// GetLease returns lease information
func (lm *LeaseManager) GetLease(leaseName string) *ResourceLease {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	return lm.leases[leaseName]
}

// IsLeaseValid checks if lease is still valid
func (lm *LeaseManager) IsLeaseValid(leaseName string) bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	lease, exists := lm.leases[leaseName]
	if !exists {
		return false
	}

	return lease.RenewTime.Add(lease.LeaseDuration).After(time.Now())
}

// DependencyGraph represents dependencies between resources
type DependencyGraph struct {
	mu           sync.RWMutex
	dependencies map[string][]string // resource -> dependencies
	dependents   map[string][]string // resource -> dependents
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		dependencies: make(map[string][]string),
		dependents:   make(map[string][]string),
	}
}

// AddDependency adds a dependency relationship
func (dg *DependencyGraph) AddDependency(resource, dependsOn string) error {
	if resource == dependsOn {
		return fmt.Errorf("circular dependency: %s depends on itself", resource)
	}

	dg.mu.Lock()
	defer dg.mu.Unlock()

	// Check for circular dependency
	if dg.hasPath(dependsOn, resource) {
		return fmt.Errorf("circular dependency would be created")
	}

	dg.dependencies[resource] = append(dg.dependencies[resource], dependsOn)
	dg.dependents[dependsOn] = append(dg.dependents[dependsOn], resource)

	return nil
}

// GetDependencies returns direct dependencies
func (dg *DependencyGraph) GetDependencies(resource string) []string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	deps := dg.dependencies[resource]
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// GetDependents returns direct dependents
func (dg *DependencyGraph) GetDependents(resource string) []string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	deps := dg.dependents[resource]
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// GetTransitiveDependencies returns all transitive dependencies
func (dg *DependencyGraph) GetTransitiveDependencies(resource string) []string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	visited := make(map[string]bool)
	var traverse func(string)
	traverse = func(r string) {
		if visited[r] {
			return
		}
		visited[r] = true
		for _, dep := range dg.dependencies[r] {
			traverse(dep)
		}
	}

	traverse(resource)
	delete(visited, resource) // Don't include self

	result := make([]string, 0)
	for r := range visited {
		result = append(result, r)
	}
	return result
}

// hasPath checks if path exists from source to target
func (dg *DependencyGraph) hasPath(source, target string) bool {
	visited := make(map[string]bool)
	var traverse func(string) bool
	traverse = func(r string) bool {
		if visited[r] {
			return false
		}
		if r == target {
			return true
		}
		visited[r] = true
		for _, dep := range dg.dependents[r] {
			if traverse(dep) {
				return true
			}
		}
		return false
	}
	return traverse(source)
}
