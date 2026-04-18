// Package planner implements the Nomad plan-applier pattern: a
// scheduler produces a *Plan* (a set of intended allocations against
// nodes), and the applier serialises plans through a single
// goroutine so that multiple schedulers cannot double-book the same
// node's resources.
//
// The applier receives a snapshot index with every plan.  If the
// snapshot is stale (i.e. another plan committed since the caller
// took the snapshot), the applier returns a PartialCommit result
// listing which allocations were dropped because their target node
// no longer has capacity.  The scheduler then re-evaluates and
// re-submits.
//
// This optimistic-concurrency design keeps the applier in the hot
// path only briefly — it never computes scheduling itself; it just
// enforces "my view of node capacity is still correct".
package planner

import (
	"context"
	"errors"
	"sync"
)

// Resources is a minimal shape describing what an allocation needs
// and what a node has.  Real usage will extend this with CPU /
// memory / disk / network — the applier only cares that the
// quantities can be added and compared.
type Resources struct {
	CPU    int64
	Memory int64
	Disk   int64
}

// Add returns r + other.
func (r Resources) Add(other Resources) Resources {
	return Resources{
		CPU:    r.CPU + other.CPU,
		Memory: r.Memory + other.Memory,
		Disk:   r.Disk + other.Disk,
	}
}

// Fits reports whether other fits within r (all fields >= other's).
func (r Resources) Fits(other Resources) bool {
	return r.CPU >= other.CPU && r.Memory >= other.Memory && r.Disk >= other.Disk
}

// Sub returns r - other, clamped at zero per field.
func (r Resources) Sub(other Resources) Resources {
	out := Resources{
		CPU:    r.CPU - other.CPU,
		Memory: r.Memory - other.Memory,
		Disk:   r.Disk - other.Disk,
	}
	if out.CPU < 0 {
		out.CPU = 0
	}
	if out.Memory < 0 {
		out.Memory = 0
	}
	if out.Disk < 0 {
		out.Disk = 0
	}
	return out
}

// Allocation is a proposed placement of a workload unit on a node.
type Allocation struct {
	// ID is the allocation's stable identifier.
	ID string
	// NodeID names the target node.
	NodeID string
	// Resources names the demand this allocation will impose.
	Resources Resources
	// Job identifies the parent workload for log/metric purposes.
	Job string
}

// Plan is a batch submission from a scheduler.  Allocations all
// reference the same SnapshotIndex, allowing the applier to detect
// staleness.
type Plan struct {
	// SchedulerID identifies the originating scheduler for logs.
	SchedulerID string
	// SnapshotIndex is the applier-assigned index the scheduler
	// observed.  Plans whose index < applier's current index risk
	// partial commit.
	SnapshotIndex uint64
	// Add is the list of allocations to place.
	Add []Allocation
	// Evict is the list of allocation IDs to remove.  Used for
	// preemption and rescheduling.
	Evict []string
}

// Result describes what actually committed.
type Result struct {
	// NewIndex is the applier's index after committing.
	NewIndex uint64
	// Committed are allocations that were placed.
	Committed []Allocation
	// Rejected are allocations the applier could not fit.  A non-
	// empty Rejected means the scheduler must re-evaluate and
	// resubmit — this is the signal that the plan was partial.
	Rejected []Allocation
	// Evicted are the IDs actually removed.  Evictions never fail,
	// so this always equals plan.Evict — returned for convenience.
	Evicted []string
}

// IsPartial reports whether the scheduler must re-run.
func (r Result) IsPartial() bool { return len(r.Rejected) > 0 }

// NodeState tracks the remaining capacity of one node.
type NodeState struct {
	// ID is the node identifier.
	ID string
	// Total is the node's full resource envelope.
	Total Resources
	// Used is the sum of committed-allocation resources.
	Used Resources
	// Allocations is the set of committed alloc IDs on this node —
	// used by Evict to Sub their resources.
	Allocations map[string]Resources
}

// Available returns Total - Used.
func (n *NodeState) Available() Resources { return n.Total.Sub(n.Used) }

// Applier is the serialising committer.  All plans flow through its
// single goroutine.
type Applier struct {
	mu    sync.Mutex
	nodes map[string]*NodeState
	index uint64
}

// NewApplier returns an empty applier.  Callers populate nodes via
// UpsertNode before submitting plans.
func NewApplier() *Applier {
	return &Applier{nodes: map[string]*NodeState{}, index: 1}
}

// UpsertNode registers or updates a node's total capacity.
func (a *Applier) UpsertNode(id string, total Resources) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if existing, ok := a.nodes[id]; ok {
		existing.Total = total
		return
	}
	a.nodes[id] = &NodeState{
		ID:          id,
		Total:       total,
		Allocations: map[string]Resources{},
	}
	a.index++
}

// Snapshot returns the applier's current index.  Schedulers call
// this before planning so they can submit a SnapshotIndex.
func (a *Applier) Snapshot() uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.index
}

// ErrUnknownNode is returned when a plan references a node the
// applier has never seen.  This is a hard error — not a partial
// commit — because it signals scheduler/applier disagreement that
// retry cannot fix.
var ErrUnknownNode = errors.New("planner: plan references unknown node")

// Submit commits plan.  Evictions apply first (freeing resources),
// then Adds are placed in order; additions that do not fit are
// collected into Result.Rejected.
func (a *Applier) Submit(_ context.Context, plan Plan) (Result, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Validate node references up front.
	for _, id := range plan.Evict {
		_ = id // evictions tolerate unknown IDs — may have been GC'd
	}
	for _, alloc := range plan.Add {
		if _, ok := a.nodes[alloc.NodeID]; !ok {
			return Result{}, ErrUnknownNode
		}
	}

	res := Result{Evicted: append([]string(nil), plan.Evict...)}

	// Apply evictions first.
	for _, allocID := range plan.Evict {
		for _, n := range a.nodes {
			if r, ok := n.Allocations[allocID]; ok {
				n.Used = n.Used.Sub(r)
				delete(n.Allocations, allocID)
				break
			}
		}
	}

	// Apply adds.
	for _, alloc := range plan.Add {
		n := a.nodes[alloc.NodeID]
		if !n.Available().Fits(alloc.Resources) {
			res.Rejected = append(res.Rejected, alloc)
			continue
		}
		n.Used = n.Used.Add(alloc.Resources)
		n.Allocations[alloc.ID] = alloc.Resources
		res.Committed = append(res.Committed, alloc)
	}

	a.index++
	res.NewIndex = a.index
	return res, nil
}
