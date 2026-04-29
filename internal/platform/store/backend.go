// Package store — backend.go
//
// Phase 5 of the etcd replacement plan: storage backend abstraction.
//
// This file provides the BackendManager which initialises either the
// etcd or Raft storage backend based on the STORAGE_BACKEND env var,
// and a generic NewStore helper that creates the correct ResourceStore
// implementation for each resource Kind.
//
// Usage in main.go:
//
//	bm, err := store.NewBackendManager()
//	defer bm.Close()
//	bulkStore := store.NewStore[*bulk.BulkOperationResource](bm, "bulkoperations", factory)
package store

import (
	"fmt"
	"log"
	"os"
	"strings"

	axraft "example.com/axiomnizam/internal/platform/raft"

	"github.com/hashicorp/go-memdb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Backend identifies the active storage backend.
type Backend string

const (
	BackendEtcd Backend = "etcd"
	BackendRaft Backend = "raft"
)

// BackendManager holds the shared infrastructure for whichever
// storage backend is active.  Create one at startup, then use
// NewStore to create per-Kind stores.
type BackendManager struct {
	// Active backend type.
	Backend Backend

	// Etcd client — non-nil when Backend == BackendEtcd.
	EtcdClient *clientv3.Client

	// Raft server — non-nil when Backend == BackendRaft.
	RaftServer *axraft.Server

	// tables tracks all registered table names (Raft mode only).
	tables []string

	// kvStore is the lazily-initialised KVStore for direct KV ops.
	kvStore KVStore
}

// NewBackendManager creates a BackendManager based on the
// STORAGE_BACKEND environment variable.
//
// For "raft": initialises the embedded Raft server with a shared
// go-memdb instance containing all resource tables.
//
// For "etcd" (default): expects etcdClient to be provided later via
// SetEtcdClient.
func NewBackendManager(tableNames []string) (*BackendManager, error) {
	backend := Backend(strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_BACKEND"))))
	if backend == "" {
		backend = BackendEtcd
	}

	bm := &BackendManager{
		Backend: backend,
		tables:  tableNames,
	}

	if backend == BackendRaft {
		if err := bm.initRaft(tableNames); err != nil {
			return nil, fmt.Errorf("backend manager: %w", err)
		}
	}

	return bm, nil
}

// SetEtcdClient sets the etcd client for etcd-backend mode.  Called
// after database.InitConnections in main.go.
func (bm *BackendManager) SetEtcdClient(client *clientv3.Client) {
	bm.EtcdClient = client
}

// IsRaft returns true if the active backend is Raft.
func (bm *BackendManager) IsRaft() bool {
	return bm.Backend == BackendRaft
}

// IsEtcd returns true if the active backend is etcd.
func (bm *BackendManager) IsEtcd() bool {
	return bm.Backend == BackendEtcd
}

// Close shuts down the active backend.
func (bm *BackendManager) Close() error {
	if bm.kvStore != nil {
		if closer, ok := bm.kvStore.(interface{ Close() }); ok {
			closer.Close()
		}
	}
	if bm.RaftServer != nil {
		return bm.RaftServer.Shutdown()
	}
	return nil
}

// KV returns the KVStore for direct key-value operations.
// Lazily initialised on first call.
func (bm *BackendManager) KV() KVStore {
	if bm.kvStore != nil {
		return bm.kvStore
	}
	bm.kvStore = NewKVStore(bm)
	return bm.kvStore
}

func (bm *BackendManager) initRaft(tableNames []string) error {
	log.Println("📦 Initializing embedded Raft storage backend...")

	// Build a single go-memdb schema with all resource tables.
	schema := NewMultiTableSchema(tableNames)
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return fmt.Errorf("create memdb: %w", err)
	}

	// Start the Raft server.
	cfg := axraft.DefaultConfig()
	server, err := axraft.NewServer(cfg, db, tableNames)
	if err != nil {
		return fmt.Errorf("start raft server: %w", err)
	}

	bm.RaftServer = server
	log.Printf("  ✅ Raft server started (node=%s, addr=%s, leader=%v)",
		cfg.NodeID, cfg.BindAddr, server.IsLeader())
	return nil
}

// NewStore creates a ResourceStore[T] using the active backend.
//
// For etcd: creates an EtcdStore with the given prefix.
// For raft: creates a RaftStore backed by the shared Raft server.
//
// The `tableName` is used as the go-memdb table name (raft) or
// converted to an etcd prefix (etcd).  The `factory` allocates a
// new zero-value T for deserialisation.
func NewStore[T Resource](bm *BackendManager, tableName string, factory func() T) ResourceStore[T] {
	switch bm.Backend {
	case BackendRaft:
		return NewRaftStore[T](bm.RaftServer, tableName, factory, nil)
	default:
		// etcd: convert table name to etcd prefix.
		prefix := "/axiomnizam/" + tableName + "/"
		return NewEtcdStore[T](bm.EtcdClient, prefix, factory)
	}
}
