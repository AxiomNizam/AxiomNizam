// Package raft — config.go
//
// Phase 3 of the etcd replacement plan: Raft server configuration.
//
// Provides sensible defaults for single-node development and
// multi-node production deployments.  All values can be overridden
// via environment variables.
package raft

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config holds all tunables for the embedded Raft server.
type Config struct {
	// DataDir is the root directory for Raft state (logs, snapshots,
	// stable store).  Defaults to "./data/raft".
	DataDir string

	// BindAddr is the address the Raft transport listens on.
	// Defaults to "127.0.0.1:9700".
	BindAddr string

	// AdvertiseAddr is the address advertised to other Raft nodes.
	// If empty, defaults to BindAddr.  In Docker, set this to the
	// container hostname (e.g. "axiomnizam-1:9700") while binding
	// to "0.0.0.0:9700".
	AdvertiseAddr string

	// NodeID uniquely identifies this server in the Raft cluster.
	// Defaults to "node-1".
	NodeID string

	// Bootstrap indicates whether this node should bootstrap a new
	// single-node cluster.  Only the first node should set this.
	Bootstrap bool

	// Peers is a list of "host:port" addresses for other Raft nodes.
	// Empty for single-node mode.
	Peers []string

	// SnapshotInterval controls how often Raft takes automatic
	// snapshots for log compaction.  Defaults to 2 minutes.
	SnapshotInterval time.Duration

	// SnapshotThreshold is the minimum number of log entries before
	// a snapshot is triggered.  Defaults to 8192.
	SnapshotThreshold uint64

	// HeartbeatTimeout is the Raft heartbeat timeout.
	// Defaults to 1 second.
	HeartbeatTimeout time.Duration

	// ElectionTimeout is the Raft election timeout.
	// Defaults to 1 second.
	ElectionTimeout time.Duration

	// LeaderLeaseTimeout is the leader lease timeout.
	// Defaults to 500ms.
	LeaderLeaseTimeout time.Duration

	// RetainSnapshotCount is the number of snapshots to retain.
	// Defaults to 2.
	RetainSnapshotCount int
}

// DefaultConfig returns a Config populated from environment variables
// with sensible defaults for single-node development.
func DefaultConfig() *Config {
	cfg := &Config{
		DataDir:             envOrDefault("AXIOMNIZAM_RAFT_DATA_DIR", filepath.Join(".", "data", "raft")),
		BindAddr:            envOrDefault("AXIOMNIZAM_RAFT_BIND_ADDR", "127.0.0.1:9700"),
		AdvertiseAddr:       envOrDefault("AXIOMNIZAM_RAFT_ADVERTISE_ADDR", ""),
		NodeID:              envOrDefault("AXIOMNIZAM_RAFT_NODE_ID", "node-1"),
		Bootstrap:           envOrDefaultBool("AXIOMNIZAM_RAFT_BOOTSTRAP", true),
		SnapshotInterval:    2 * time.Minute,
		SnapshotThreshold:   8192,
		HeartbeatTimeout:    1 * time.Second,
		ElectionTimeout:     1 * time.Second,
		LeaderLeaseTimeout:  500 * time.Millisecond,
		RetainSnapshotCount: 2,
	}
	return cfg
}

// LogDir returns the path for Raft log storage (BoltDB).
func (c *Config) LogDir() string {
	return filepath.Join(c.DataDir, "logs")
}

// SnapshotDir returns the path for Raft snapshots.
func (c *Config) SnapshotDir() string {
	return filepath.Join(c.DataDir, "snapshots")
}

// StableDir returns the path for Raft stable store (BoltDB).
func (c *Config) StableDir() string {
	return filepath.Join(c.DataDir, "stable")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
