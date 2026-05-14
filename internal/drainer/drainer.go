// Package drainer implements the Nomad node-drain state machine:
// when an operator marks a node for drain (for patching, scaling,
// or decommissioning), the drainer evicts that node's allocations
// in controlled batches so that downstream reschedulers are not
// overwhelmed.
//
// The drainer is intentionally decoupled from the scheduler: it
// emits EvictionRequests over a channel; a separate consumer (the
// planner, typically) handles actual placement.
package drainer

import (
	"context"
	"errors"
	"sync"
	"time"
)

// DrainMode controls termination behaviour.
type DrainMode int

const (
	// DrainModeDefault evicts workloads gradually, respecting the
	// deadline and batch size.  Allocations past the deadline are
	// force-evicted.
	DrainModeDefault DrainMode = iota
	// DrainModeForce skips the grace period and evicts everything
	// immediately — used for emergencies.
	DrainModeForce
)

// Spec parameterises a drain.
type Spec struct {
	// NodeID names the node being drained.
	NodeID string
	// Mode selects the drain strategy.
	Mode DrainMode
	// Deadline is the absolute time after which remaining allocs
	// are force-evicted.  Zero disables the deadline.
	Deadline time.Time
	// IgnoreSystemJobs, when true, exempts system-tier allocations
	// (they will survive the drain — used when drain is
	// preparation for a config reload, not a decommission).
	IgnoreSystemJobs bool
	// BatchSize caps concurrent evictions.  Zero defaults to 1.
	BatchSize int
	// BatchDelay spaces evictions out to avoid thundering-herd
	// reschedules.  Zero defaults to 500ms.
	BatchDelay time.Duration
}

// EvictionRequest is emitted for each allocation the drainer wants
// terminated.  A consumer is responsible for actually stopping the
// workload and confirming via the Complete method.
type EvictionRequest struct {
	// AllocID is the allocation to terminate.
	AllocID string
	// NodeID is the node it runs on.
	NodeID string
	// SystemJob is true if it belongs to a system-tier job; the
	// consumer may choose to honour IgnoreSystemJobs itself.
	SystemJob bool
	// Force indicates the deadline has passed — the consumer should
	// terminate without respecting any graceful-shutdown period.
	Force bool
}

// Allocation is the minimal shape the drainer needs to know about.
type Allocation struct {
	// ID is the alloc identifier.
	ID string
	// NodeID is the allocation's current home.
	NodeID string
	// System flags system-tier allocs.
	System bool
}

// Drainer owns the state machine.
type Drainer struct {
	mu        sync.Mutex
	active    map[string]*drain
	evictions chan EvictionRequest
}

// drain tracks one in-progress drain.
type drain struct {
	spec    Spec
	pending []Allocation
	ctx     context.Context
	cancel  context.CancelFunc
}

// New returns an idle drainer.
func New() *Drainer {
	return &Drainer{
		active:    map[string]*drain{},
		evictions: make(chan EvictionRequest, 64),
	}
}

// Evictions returns the receive channel for eviction requests.
// Consumers should range over it until the drainer is closed.
func (d *Drainer) Evictions() <-chan EvictionRequest { return d.evictions }

// ErrAlreadyDraining is returned if a drain is already active for
// the node.  Callers must Cancel first if they want to replace it.
var ErrAlreadyDraining = errors.New("drainer: node already draining")

// Begin starts a drain for the given node with the provided alloc
// list.  The drainer runs in a goroutine and emits evictions to the
// Evictions channel.
func (d *Drainer) Begin(spec Spec, allocs []Allocation) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.active[spec.NodeID]; ok {
		return ErrAlreadyDraining
	}
	if spec.BatchSize == 0 {
		spec.BatchSize = 1
	}
	if spec.BatchDelay == 0 {
		spec.BatchDelay = 500 * time.Millisecond
	}
	// Copy relevant allocs (filter out system jobs if requested).
	var pending []Allocation
	for _, a := range allocs {
		if a.NodeID != spec.NodeID {
			continue
		}
		if a.System && spec.IgnoreSystemJobs {
			continue
		}
		pending = append(pending, a)
	}
	ctx, cancel := context.WithCancel(context.Background())
	dr := &drain{spec: spec, pending: pending, ctx: ctx, cancel: cancel}
	d.active[spec.NodeID] = dr
	go d.run(dr)
	return nil
}

// Cancel halts a drain.  In-flight evictions already emitted are
// not recalled — the consumer may have acted on them already.
func (d *Drainer) Cancel(nodeID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if dr, ok := d.active[nodeID]; ok {
		dr.cancel()
		delete(d.active, nodeID)
	}
}

// run is the per-drain goroutine.
func (d *Drainer) run(dr *drain) {
	defer func() {
		d.mu.Lock()
		delete(d.active, dr.spec.NodeID)
		d.mu.Unlock()
	}()

	forced := dr.spec.Mode == DrainModeForce
	ticker := time.NewTicker(dr.spec.BatchDelay)
	defer ticker.Stop()

	pending := dr.pending
	for len(pending) > 0 {
		// Check deadline.
		if !dr.spec.Deadline.IsZero() && time.Now().After(dr.spec.Deadline) {
			forced = true
		}

		batchSize := dr.spec.BatchSize
		if batchSize > len(pending) {
			batchSize = len(pending)
		}
		batch := pending[:batchSize]
		pending = pending[batchSize:]

		for _, a := range batch {
			req := EvictionRequest{
				AllocID:   a.ID,
				NodeID:    a.NodeID,
				SystemJob: a.System,
				Force:     forced,
			}
			select {
			case d.evictions <- req:
			case <-dr.ctx.Done():
				return
			}
		}
		if len(pending) == 0 {
			return
		}

		// Honour batch delay unless forced.
		if !forced {
			select {
			case <-ticker.C:
			case <-dr.ctx.Done():
				return
			}
		}
	}
}
