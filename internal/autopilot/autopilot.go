// Package autopilot implements Consul/Nomad-style cluster
// autopilot: a background health checker that classifies servers
// as healthy / stale / failing, promotes non-voting members when
// the voting set is short a member, and offers a CleanupDeadServers
// capability that kicks out peers that have been gone for too long.
//
// The package exposes primitives — it does not itself speak Raft.
// Callers provide a Transport that returns each peer's LastContact
// and current role; autopilot decides what to do and returns an
// action plan.
package autopilot

import (
	"context"
	"sync"
	"time"
)

// Role identifies a peer's role in the consensus cluster.
type Role int

const (
	// RoleVoter is a full voting member.
	RoleVoter Role = iota
	// RoleNonVoter is a member that has caught up but is still on
	// probation.
	RoleNonVoter
	// RoleLearner is a joining member that has not yet caught up.
	RoleLearner
)

// Server describes one member's observable state.
type Server struct {
	// ID is the stable server identifier.
	ID string
	// Address is the RPC endpoint.
	Address string
	// Role names the server's current consensus role.
	Role Role
	// LastContact is the time since the server was last reachable.
	LastContact time.Duration
	// LastIndex is the highest log index the server has applied.
	LastIndex uint64
	// Healthy is the transport's own up/down verdict.
	Healthy bool
}

// Config tunes autopilot.
type Config struct {
	// MaxTrailingLogs is the largest acceptable lag in log index
	// before a peer is considered stale.
	MaxTrailingLogs uint64
	// LastContactThreshold marks peers silent longer than this as
	// failing.  Canonical default is 200ms for a LAN cluster.
	LastContactThreshold time.Duration
	// DeadServerCleanup enables automatic removal of dead servers.
	// Defaults to true — matches Nomad default.
	DeadServerCleanup bool
	// MinQuorum prevents autopilot from reducing the voter count
	// below this floor.  Essential in small clusters where losing
	// one more voter would break availability.
	MinQuorum int
}

// Action is what autopilot wants done.
type Action int

const (
	// ActionNone signals a healthy peer that requires no change.
	ActionNone Action = iota
	// ActionPromote changes a RoleNonVoter to RoleVoter.
	ActionPromote
	// ActionRemove excludes the server from the configuration.
	ActionRemove
)

// Decision pairs a server with the action autopilot recommends.
type Decision struct {
	// Server is the subject.
	Server Server
	// Action is the recommended transition.
	Action Action
	// Reason is a human-readable description suitable for logs.
	Reason string
}

// Autopilot is stateless between evaluations — the state lives on
// the caller's side.  Repeated Evaluate calls are the way to drive
// the state machine forward.
type Autopilot struct {
	cfg Config
	mu  sync.Mutex
}

// New returns an autopilot with defaulted config.
func New(cfg Config) *Autopilot {
	if cfg.LastContactThreshold == 0 {
		cfg.LastContactThreshold = 200 * time.Millisecond
	}
	if cfg.MaxTrailingLogs == 0 {
		cfg.MaxTrailingLogs = 250
	}
	if cfg.MinQuorum == 0 {
		cfg.MinQuorum = 3
	}
	return &Autopilot{cfg: cfg}
}

// Evaluate inspects peers and returns the decisions autopilot
// recommends.  The leader index is used to compute trailing lag.
func (a *Autopilot) Evaluate(_ context.Context, peers []Server, leaderIndex uint64) []Decision {
	a.mu.Lock()
	cfg := a.cfg
	a.mu.Unlock()

	voters := 0
	for _, p := range peers {
		if p.Role == RoleVoter && p.Healthy {
			voters++
		}
	}

	var out []Decision
	for _, p := range peers {
		// Treat the leader (LastContact==0, Healthy, Voter) as stable.
		trailing := uint64(0)
		if leaderIndex > p.LastIndex {
			trailing = leaderIndex - p.LastIndex
		}

		// Promote non-voters that have caught up.
		if p.Role == RoleNonVoter && p.Healthy &&
			p.LastContact <= cfg.LastContactThreshold &&
			trailing <= cfg.MaxTrailingLogs {
			out = append(out, Decision{
				Server: p, Action: ActionPromote,
				Reason: "non-voter caught up to within trailing/contact thresholds",
			})
			continue
		}

		// Remove dead voters if enabled and quorum would survive.
		if cfg.DeadServerCleanup && !p.Healthy {
			if p.Role == RoleVoter && voters-1 < cfg.MinQuorum {
				// Refuse to kick — would break quorum.
				out = append(out, Decision{
					Server: p, Action: ActionNone,
					Reason: "dead server retained to preserve quorum",
				})
				continue
			}
			out = append(out, Decision{
				Server: p, Action: ActionRemove,
				Reason: "dead server cleanup",
			})
			if p.Role == RoleVoter {
				voters--
			}
			continue
		}
		out = append(out, Decision{Server: p, Action: ActionNone})
	}
	return out
}
