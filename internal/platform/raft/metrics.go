// Package raft — metrics.go
//
// Raft subsystem metrics following Nomad's pattern of measuring every
// FSM operation and server lifecycle event.
package raft

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks operational counters for the Raft subsystem.
// Thread-safe for concurrent access from FSM Apply, snapshot, and
// server goroutines.
type Metrics struct {
	// FSM operation counters.
	ApplyCount   atomic.Int64
	CreateCount  atomic.Int64
	UpdateCount  atomic.Int64
	DeleteCount  atomic.Int64
	ApplyErrors  atomic.Int64

	// Snapshot counters.
	SnapshotCount  atomic.Int64
	RestoreCount   atomic.Int64
	SnapshotErrors atomic.Int64

	// Timing (last operation duration in microseconds).
	LastApplyMicros    atomic.Int64
	LastSnapshotMicros atomic.Int64

	// Leadership.
	LeaderChanges atomic.Int64
	IsLeader      atomic.Bool

	// Server state.
	StartedAt time.Time

	mu            sync.RWMutex
	lastApplyTime time.Time
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		StartedAt: time.Now(),
	}
}

// RecordApply records a successful FSM apply operation.
func (m *Metrics) RecordApply(cmdType CommandType, duration time.Duration) {
	m.ApplyCount.Add(1)
	m.LastApplyMicros.Store(duration.Microseconds())

	switch cmdType {
	case CommandCreate:
		m.CreateCount.Add(1)
	case CommandUpdate:
		m.UpdateCount.Add(1)
	case CommandDelete:
		m.DeleteCount.Add(1)
	}

	m.mu.Lock()
	m.lastApplyTime = time.Now()
	m.mu.Unlock()
}

// RecordApplyError records a failed FSM apply operation.
func (m *Metrics) RecordApplyError() {
	m.ApplyErrors.Add(1)
}

// RecordSnapshot records a snapshot operation.
func (m *Metrics) RecordSnapshot(duration time.Duration) {
	m.SnapshotCount.Add(1)
	m.LastSnapshotMicros.Store(duration.Microseconds())
}

// RecordSnapshotError records a failed snapshot.
func (m *Metrics) RecordSnapshotError() {
	m.SnapshotErrors.Add(1)
}

// RecordLeaderChange records a leadership transition.
func (m *Metrics) RecordLeaderChange(isLeader bool) {
	m.LeaderChanges.Add(1)
	m.IsLeader.Store(isLeader)
}

// LastApplyTime returns the time of the last successful apply.
func (m *Metrics) LastApplyTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastApplyTime
}

// Snapshot returns a point-in-time copy of all metrics for reporting.
func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		ApplyCount:         m.ApplyCount.Load(),
		CreateCount:        m.CreateCount.Load(),
		UpdateCount:        m.UpdateCount.Load(),
		DeleteCount:        m.DeleteCount.Load(),
		ApplyErrors:        m.ApplyErrors.Load(),
		SnapshotCount:      m.SnapshotCount.Load(),
		RestoreCount:       m.RestoreCount.Load(),
		SnapshotErrors:     m.SnapshotErrors.Load(),
		LastApplyMicros:    m.LastApplyMicros.Load(),
		LastSnapshotMicros: m.LastSnapshotMicros.Load(),
		LeaderChanges:      m.LeaderChanges.Load(),
		IsLeader:           m.IsLeader.Load(),
		Uptime:             time.Since(m.StartedAt),
		LastApplyTime:      m.LastApplyTime(),
	}
}

// MetricsSnapshot is a serialisable point-in-time view of Raft metrics.
type MetricsSnapshot struct {
	ApplyCount         int64         `json:"apply_count"`
	CreateCount        int64         `json:"create_count"`
	UpdateCount        int64         `json:"update_count"`
	DeleteCount        int64         `json:"delete_count"`
	ApplyErrors        int64         `json:"apply_errors"`
	SnapshotCount      int64         `json:"snapshot_count"`
	RestoreCount       int64         `json:"restore_count"`
	SnapshotErrors     int64         `json:"snapshot_errors"`
	LastApplyMicros    int64         `json:"last_apply_micros"`
	LastSnapshotMicros int64         `json:"last_snapshot_micros"`
	LeaderChanges      int64         `json:"leader_changes"`
	IsLeader           bool          `json:"is_leader"`
	Uptime             time.Duration `json:"uptime"`
	LastApplyTime      time.Time     `json:"last_apply_time"`
}
