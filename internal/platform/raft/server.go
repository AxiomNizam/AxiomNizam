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

	logger  hclog.Logger
	metrics *Metrics
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
		config:  cfg,
		db:      db,
		fsm:     NewFSM(db, tables),
		logger:  logger,
		metrics: NewMetrics(),
	}

	// Wire logger into FSM for structured logging.
	s.fsm.SetLogger(logger)
	s.fsm.metrics = s.metrics

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
	// Determine the advertise address.  When running in Docker,
	// BindAddr is typically "0.0.0.0:9700" (not advertisable), so
	// we use AdvertiseAddr (e.g. "axiomnizam-1:9700") instead.
	advertiseAddr := s.config.BindAddr
	if s.config.AdvertiseAddr != "" {
		advertiseAddr = s.config.AdvertiseAddr
	}
	addr, err := net.ResolveTCPAddr("tcp", advertiseAddr)
	if err != nil {
		return fmt.Errorf("resolve advertise addr: %w", err)
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
		// Use advertise address in cluster config so peers can reach us.
		bootstrapAddr := s.config.BindAddr
		if s.config.AdvertiseAddr != "" {
			bootstrapAddr = s.config.AdvertiseAddr
		}
		cfg := hraft.Configuration{
			Servers: []hraft.Server{
				{
					ID:      hraft.ServerID(s.config.NodeID),
					Address: hraft.ServerAddress(bootstrapAddr),
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
//
// Returns ErrNotLeader if this node is not the leader.  Callers
// should retry on the leader node.
func (s *Server) Apply(data []byte, timeout time.Duration) error {
	if s.raft.State() != hraft.Leader {
		return ErrNotLeader
	}

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

// ErrNotLeader is returned when a write is attempted on a non-leader node.
var ErrNotLeader = fmt.Errorf("node is not the leader")

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

// Shutdown gracefully stops the Raft server and closes all stores.
func (s *Server) Shutdown() error {
	s.logger.Info("shutting down raft server")

	f := s.raft.Shutdown()
	if err := f.Error(); err != nil {
		s.logger.Error("raft shutdown error", "error", err)
	}

	// Close BoltDB stores to release file locks.
	if closer, ok := s.logStore.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			s.logger.Error("log store close error", "error", err)
		}
	}
	if closer, ok := s.stable.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			s.logger.Error("stable store close error", "error", err)
		}
	}

	return f.Error()
}

// GetMetrics returns the Raft subsystem metrics.
func (s *Server) GetMetrics() *Metrics {
	return s.metrics
}

// QuickStatus builds a stats map from non-blocking atomic reads ONLY.
// Unlike Stats(), this NEVER touches the Raft main loop and will
// always return instantly, even during elections or heavy replication.
//
// Missing vs Stats(): latest_configuration, num_peers (these require
// GetConfiguration which blocks on the main loop).
func (s *Server) QuickStatus() map[string]string {
	toString := func(v uint64) string {
		return fmt.Sprintf("%d", v)
	}

	m := map[string]string{
		"state":          s.raft.State().String(),
		"term":           toString(s.raft.CurrentTerm()),
		"last_log_index": toString(s.raft.LastIndex()),
		"commit_index":   toString(s.raft.CommitIndex()),
		"applied_index":  toString(s.raft.AppliedIndex()),
	}

	// LastContact
	if s.raft.State() == hraft.Leader {
		m["last_contact"] = "0"
	} else {
		lc := s.raft.LastContact()
		if lc.IsZero() {
			m["last_contact"] = "never"
		} else {
			m["last_contact"] = fmt.Sprintf("%v", time.Since(lc))
		}
	}

	return m
}

// Stats returns Raft internal stats (term, commit index, etc.).
// WARNING: This calls GetConfiguration() internally which BLOCKS on the
// Raft main loop.  Use QuickStatus() for non-blocking health checks.
func (s *Server) Stats() map[string]string {
	return s.raft.Stats()
}

// LastContact returns the time since this node last heard from the
// leader.  Useful for health checks.
func (s *Server) LastContact() time.Time {
	return s.raft.LastContact()
}

// LeaderWithID returns the address and ID of the current leader.
func (s *Server) LeaderWithID() (string, string) {
	addr, id := s.raft.LeaderWithID()
	return string(addr), string(id)
}

// GetConfiguration returns the current Raft cluster configuration
// (list of servers, their roles, and suffrage).
func (s *Server) GetConfiguration() ([]PeerInfo, error) {
	future := s.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return nil, err
	}
	var peers []PeerInfo
	for _, srv := range future.Configuration().Servers {
		peers = append(peers, PeerInfo{
			ID:       string(srv.ID),
			Address:  string(srv.Address),
			Suffrage: srv.Suffrage.String(),
		})
	}
	return peers, nil
}

// PeerInfo describes a server in the Raft cluster.
type PeerInfo struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Suffrage string `json:"suffrage"`
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

// TriggerSnapshot forces an on-demand Raft snapshot. Blocks until complete.
func (s *Server) TriggerSnapshot() error {
	f := s.raft.Snapshot()
	return f.Error()
}

// DataDir returns the root data directory for Raft state.
func (s *Server) DataDir() string { return s.config.DataDir }

// SnapshotDir returns the snapshot directory path.
func (s *Server) SnapshotDir() string { return s.config.SnapshotDir() }

// LogDir returns the log store directory path.
func (s *Server) LogDir() string { return s.config.LogDir() }

// StableDir returns the stable store directory path.
func (s *Server) StableDir() string { return s.config.StableDir() }
