// Package raft — server.go
//
// Phase 3 of the etcd replacement plan: Raft server initialization,
// peer management, and lifecycle.
//
// The Server wraps hashicorp/raft with:
//   - BoltDB-backed log and stable stores (durable across restarts)
//   - File-based snapshot store
//   - TCP transport for multi-node clusters
//   - Single-node bootstrap for development
//
// Design inspired by Nomad's nomad/server.go.
package raft

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-memdb"
	hraft "github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

// Server manages the embedded Raft node and its associated go-memdb
// state store.
type Server struct {
	config *Config
	raft   *hraft.Raft
	fsm    *FSM
	db     *memdb.MemDB

	transport hraft.Transport
	logStore  hraft.LogStore
	stable    hraft.StableStore
	snapshots hraft.SnapshotStore

	logger hclog.Logger
}

// NewServer creates and starts an embedded Raft server.
//
// `db` is the shared go-memdb instance that the FSM will mutate.
// `tables` lists every table name in the go-memdb schema so the FSM
// can snapshot all of them.
func NewServer(cfg *Config, db *memdb.MemDB, tables []string) (*Server, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "raft",
		Level:  hclog.Info,
		Output: os.Stderr,
	})

	s := &Server{
		config: cfg,
		db:     db,
		fsm:    NewFSM(db, tables),
		logger: logger,
	}

	if err := s.setupRaft(); err != nil {
		return nil, fmt.Errorf("raft server: %w", err)
	}

	return s, nil
}

func (s *Server) setupRaft() error {
	// Ensure data directories exist.
	for _, dir := range []string{s.config.LogDir(), s.config.StableDir(), s.config.SnapshotDir()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", dir, err)
		}
	}

	// Raft configuration.
	raftCfg := hraft.DefaultConfig()
	raftCfg.LocalID = hraft.ServerID(s.config.NodeID)
	raftCfg.HeartbeatTimeout = s.config.HeartbeatTimeout
	raftCfg.ElectionTimeout = s.config.ElectionTimeout
	raftCfg.LeaderLeaseTimeout = s.config.LeaderLeaseTimeout
	raftCfg.SnapshotInterval = s.config.SnapshotInterval
	raftCfg.SnapshotThreshold = s.config.SnapshotThreshold
	raftCfg.Logger = s.logger

	// TCP transport.
	addr, err := net.ResolveTCPAddr("tcp", s.config.BindAddr)
	if err != nil {
		return fmt.Errorf("resolve bind addr: %w", err)
	}
	transport, err := hraft.NewTCPTransport(s.config.BindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return fmt.Errorf("tcp transport: %w", err)
	}
	s.transport = transport

	// BoltDB log store.
	logStorePath := filepath.Join(s.config.LogDir(), "raft-log.bolt")
	logStore, err := raftboltdb.NewBoltStore(logStorePath)
	if err != nil {
		return fmt.Errorf("log store: %w", err)
	}
	s.logStore = logStore

	// BoltDB stable store (for Raft metadata: current term, voted-for).
	stablePath := filepath.Join(s.config.StableDir(), "raft-stable.bolt")
	stableStore, err := raftboltdb.NewBoltStore(stablePath)
	if err != nil {
		return fmt.Errorf("stable store: %w", err)
	}
	s.stable = stableStore

	// File snapshot store.
	snapshots, err := hraft.NewFileSnapshotStore(s.config.SnapshotDir(), s.config.RetainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("snapshot store: %w", err)
	}
	s.snapshots = snapshots

	// Create the Raft instance.
	r, err := hraft.NewRaft(raftCfg, s.fsm, s.logStore, s.stable, s.snapshots, s.transport)
	if err != nil {
		return fmt.Errorf("new raft: %w", err)
	}
	s.raft = r

	// Bootstrap if configured (single-node or first node).
	if s.config.Bootstrap {
		cfg := hraft.Configuration{
			Servers: []hraft.Server{
				{
					ID:      hraft.ServerID(s.config.NodeID),
					Address: hraft.ServerAddress(s.config.BindAddr),
				},
			},
		}
		f := s.raft.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			// ErrCantBootstrap is expected if already bootstrapped.
			if err != hraft.ErrCantBootstrap {
				s.logger.Warn("bootstrap cluster", "error", err)
			}
		}
	}

	return nil
}

// Apply submits a pre-encoded command to the Raft leader.  Blocks
// until the log entry is committed or the timeout expires.
func (s *Server) Apply(data []byte, timeout time.Duration) error {
	f := s.raft.Apply(data, timeout)
	if err := f.Error(); err != nil {
		return fmt.Errorf("raft apply: %w", err)
	}
	// Check the FSM response for application-level errors.
	if resp := f.Response(); resp != nil {
		if err, ok := resp.(error); ok {
			return err
		}
	}
	return nil
}

// IsLeader returns true if this node is the current Raft leader.
func (s *Server) IsLeader() bool {
	return s.raft.State() == hraft.Leader
}

// LeaderAddr returns the address of the current leader, or empty
// string if unknown.
func (s *Server) LeaderAddr() string {
	addr, _ := s.raft.LeaderWithID()
	return string(addr)
}

// State returns the current Raft state (Follower, Candidate, Leader,
// Shutdown).
func (s *Server) State() hraft.RaftState {
	return s.raft.State()
}

// DB returns the underlying go-memdb instance for direct reads.
func (s *Server) DB() *memdb.MemDB {
	return s.db
}

// Shutdown gracefully stops the Raft server.
func (s *Server) Shutdown() error {
	f := s.raft.Shutdown()
	return f.Error()
}

// AddPeer adds a new voting server to the Raft cluster.
func (s *Server) AddPeer(id, addr string) error {
	f := s.raft.AddVoter(hraft.ServerID(id), hraft.ServerAddress(addr), 0, 10*time.Second)
	return f.Error()
}

// RemovePeer removes a server from the Raft cluster.
func (s *Server) RemovePeer(id string) error {
	f := s.raft.RemoveServer(hraft.ServerID(id), 0, 10*time.Second)
	return f.Error()
}
