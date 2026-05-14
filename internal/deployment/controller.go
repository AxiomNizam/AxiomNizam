// Package deployment implements a canary / blue-green rollout
// controller in the shape of Nomad's deployment-watcher.
//
// The controller owns one Deployment per version transition for a
// given workload.  It tracks the progress of each allocation in
// the new version, promotes canaries once their health gate passes,
// and rolls back to the previous version if the health gate fails
// within the deadline.
//
// The controller never shuts allocations down directly — it emits
// Decisions that a higher layer (the planner, the drainer) executes.
// Keeping it decision-only makes unit-testing trivial.
package deployment

import (
	"errors"
	"sync"
	"time"
)

// Phase is the deployment lifecycle state.
type Phase string

const (
	// PhasePending means allocations have not yet started.
	PhasePending Phase = "pending"
	// PhaseRunning means canary allocations are in progress.
	PhaseRunning Phase = "running"
	// PhasePromoted means canaries passed and the rollout is
	// completing by replacing the remaining old allocations.
	PhasePromoted Phase = "promoted"
	// PhaseSucceeded is the terminal good state.
	PhaseSucceeded Phase = "succeeded"
	// PhaseFailed is the terminal bad state — rollback triggered.
	PhaseFailed Phase = "failed"
)

// Strategy selects the rollout shape.
type Strategy string

const (
	// StrategyCanary: place Canary new allocations, wait for health,
	// promote, then batch-replace the remainder.
	StrategyCanary Strategy = "canary"
	// StrategyBlueGreen: stand up the full new version in parallel,
	// swap traffic, then tear down the old version.
	StrategyBlueGreen Strategy = "blue-green"
)

// Spec parameterises a deployment.
type Spec struct {
	// JobID names the workload.
	JobID string
	// Version is the new version identifier.
	Version string
	// Strategy picks canary or blue-green behaviour.
	Strategy Strategy
	// Canary is the number of canaries to float in the canary case;
	// ignored for blue-green.
	Canary int
	// TotalAllocations is the target replica count after promotion.
	TotalAllocations int
	// HealthyDeadline bounds the time to reach the "all healthy"
	// state before rollback.  Zero disables the deadline.
	HealthyDeadline time.Duration
	// AutoPromote allows the controller to promote without manual
	// intervention once the canary health gate passes.
	AutoPromote bool
	// MinHealthyTime is how long an alloc must report healthy before
	// it counts toward promotion.
	MinHealthyTime time.Duration
}

// AllocState is the per-allocation snapshot.
type AllocState struct {
	// ID is the alloc ID.
	ID string
	// Canary is true if this is one of the canary allocs.
	Canary bool
	// Healthy is true once probes say so.
	Healthy bool
	// HealthySince records when Healthy last transitioned to true.
	HealthySince time.Time
}

// Deployment is the controller's state for one rollout.
type Deployment struct {
	Spec      Spec
	Phase     Phase
	StartedAt time.Time
	Allocs    map[string]*AllocState
}

// Decision is what the controller recommends doing next.
type Decision struct {
	// PromoteCanaries, when true, tells the planner to deploy the
	// remaining replicas up to TotalAllocations.
	PromoteCanaries bool
	// Rollback, when true, tells the planner to revert to the
	// previous version and evict all new-version allocs.
	Rollback bool
	// Complete, when true, signals terminal success.
	Complete bool
	// Reason is human-readable context for logs.
	Reason string
}

// Controller owns one Deployment.  Real callers maintain a map
// keyed by JobID; we keep the controller itself single-deployment
// for simplicity.
type Controller struct {
	mu  sync.Mutex
	d   *Deployment
	now func() time.Time // overridable for tests
}

// NewController starts a deployment in PhasePending.
func NewController(spec Spec) *Controller {
	return &Controller{
		d: &Deployment{
			Spec:      spec,
			Phase:     PhasePending,
			StartedAt: time.Now(),
			Allocs:    map[string]*AllocState{},
		},
		now: time.Now,
	}
}

// State returns the current deployment; callers must not mutate it.
func (c *Controller) State() *Deployment {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.d
}

// ErrTerminal is returned by UpdateAllocs when the deployment has
// already completed (succeeded or failed) — further updates are
// ignored.
var ErrTerminal = errors.New("deployment: already terminal")

// UpdateAllocs replaces the alloc state set with the provided map
// and runs the decision engine.
func (c *Controller) UpdateAllocs(allocs []AllocState) (Decision, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.d.Phase == PhaseSucceeded || c.d.Phase == PhaseFailed {
		return Decision{}, ErrTerminal
	}

	// Overlay updates.
	for _, a := range allocs {
		existing, ok := c.d.Allocs[a.ID]
		if !ok {
			cp := a
			c.d.Allocs[a.ID] = &cp
			continue
		}
		if a.Healthy && !existing.Healthy {
			existing.HealthySince = c.now()
		}
		existing.Healthy = a.Healthy
		existing.Canary = a.Canary
	}

	if c.d.Phase == PhasePending {
		c.d.Phase = PhaseRunning
	}
	return c.decide(), nil
}

// decide runs the core state machine under lock.
func (c *Controller) decide() Decision {
	now := c.now()
	spec := c.d.Spec

	// Check for rollout deadline.
	if spec.HealthyDeadline > 0 && now.Sub(c.d.StartedAt) > spec.HealthyDeadline {
		if c.d.Phase != PhasePromoted {
			c.d.Phase = PhaseFailed
			return Decision{Rollback: true, Reason: "healthy deadline elapsed"}
		}
	}

	// Count healthy allocations, respecting MinHealthyTime.
	healthy, canaryHealthy := 0, 0
	for _, a := range c.d.Allocs {
		if !a.Healthy {
			continue
		}
		if spec.MinHealthyTime > 0 && now.Sub(a.HealthySince) < spec.MinHealthyTime {
			continue
		}
		healthy++
		if a.Canary {
			canaryHealthy++
		}
	}

	// Canary promotion gate.
	if c.d.Phase == PhaseRunning && spec.Strategy == StrategyCanary &&
		canaryHealthy >= spec.Canary && spec.Canary > 0 {
		if !spec.AutoPromote {
			return Decision{Reason: "canary healthy — awaiting manual promotion"}
		}
		c.d.Phase = PhasePromoted
		return Decision{PromoteCanaries: true, Reason: "canary healthy — auto-promoting"}
	}

	// Blue-green and post-promote completion gate.
	if (c.d.Phase == PhasePromoted || spec.Strategy == StrategyBlueGreen) &&
		healthy >= spec.TotalAllocations && spec.TotalAllocations > 0 {
		c.d.Phase = PhaseSucceeded
		return Decision{Complete: true, Reason: "all allocations healthy"}
	}
	return Decision{}
}

// Promote is the manual-promotion entry point, used when AutoPromote
// is false.  Safe to call even if the canary gate has not been met;
// returns false and changes nothing in that case.
func (c *Controller) Promote() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.d.Phase != PhaseRunning {
		return false
	}
	// Verify canary health.
	healthy := 0
	for _, a := range c.d.Allocs {
		if a.Canary && a.Healthy {
			healthy++
		}
	}
	if healthy < c.d.Spec.Canary {
		return false
	}
	c.d.Phase = PhasePromoted
	return true
}

// Fail forces the deployment into PhaseFailed, returning a rollback
// decision.  Used by operators overriding a stuck rollout.
func (c *Controller) Fail(reason string) Decision {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.d.Phase = PhaseFailed
	return Decision{Rollback: true, Reason: reason}
}
