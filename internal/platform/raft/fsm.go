// Package raft — fsm.go
//
// Phase 2 of the etcd replacement plan: Raft Finite State Machine.
//
// The FSM is the bridge between the Raft consensus layer and the
// go-memdb state store.  When a log entry is committed by a quorum,
// the Raft library calls FSM.Apply with the log data.  The FSM
// decodes the Command and applies the corresponding mutation to
// go-memdb.
//
// The FSM also implements Snapshot and Restore for Raft's compaction
// and recovery mechanisms.
//
// Design inspired by Nomad's nomad/fsm.go — dispatch on command type,
// snapshot by iterating all tables, restore by replaying entries.
package raft

import (
	"encoding/json"
	"fmt"
	"time"
	"io"
	"sync"

	"github.com/hashicorp/go-memdb"
	hraft "github.com/hashicorp/raft"
)

// FSM implements the hashicorp/raft.FSM interface.  It holds a
// reference to the shared go-memdb instance and applies committed
// log entries as mutations.
type FSM struct {
	mu sync.RWMutex
	db *memdb.MemDB

	// tables tracks all registered table names so Snapshot can
	// iterate them.
	tables []string

	// metrics tracks FSM operation counters and timing.
	metrics *Metrics

	// logger provides structured logging for FSM operations.
	logger interface {
		Info(msg string, args ...interface{})
		Warn(msg string, args ...interface{})
		Error(msg string, args ...interface{})
	}
}

// NewFSM creates a new FSM backed by the given go-memdb instance.
// `tables` lists every table name in the schema so the FSM can
// snapshot all of them.
func NewFSM(db *memdb.MemDB, tables []string) *FSM {
	return &FSM{
		db:      db,
		tables:  tables,
		metrics: NewMetrics(),
	}
}

// SetLogger sets the structured logger for the FSM.
func (f *FSM) SetLogger(l interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}) {
	f.logger = l
}

// Metrics returns the FSM metrics for external reporting.
func (f *FSM) GetMetrics() *Metrics {
	return f.metrics
}

// Apply is called by Raft when a log entry is committed.  It decodes
// the Command and applies the mutation to go-memdb.
//
// The return value is made available via raft.ApplyFuture.Response().
// We return nil on success or an error string on failure.
func (f *FSM) Apply(log *hraft.Log) interface{} {
	start := time.Now()

	cmd, err := DecodeCommand(log.Data)
	if err != nil {
		f.metrics.RecordApplyError()
		if f.logger != nil {
			f.logger.Error("fsm: decode command failed", "error", err, "index", log.Index)
		}
		return fmt.Errorf("fsm: %w", err)
	}

	var result interface{}
	switch cmd.Type {
	case CommandCreate:
		result = f.applyCreate(cmd)
	case CommandUpdate:
		result = f.applyUpdate(cmd)
	case CommandDelete:
		result = f.applyDelete(cmd)
	default:
		f.metrics.RecordApplyError()
		if f.logger != nil {
			f.logger.Error("fsm: unknown command type", "type", cmd.Type, "index", log.Index)
		}
		return fmt.Errorf("fsm: unknown command type %d", cmd.Type)
	}

	duration := time.Since(start)
	if result != nil {
		if _, isErr := result.(error); isErr {
			f.metrics.RecordApplyError()
			if f.logger != nil {
				f.logger.Warn("fsm: apply failed", "type", cmd.Type.String(),
					"table", cmd.Table, "key", cmd.Key, "error", result, "index", log.Index)
			}
		}
	} else {
		f.metrics.RecordApply(cmd.Type, duration)
	}

	return result
}

func (f *FSM) applyCreate(cmd *Command) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	txn := f.db.Txn(true)

	// Check for existing key (CAS: create-if-absent).
	existing, err := txn.First(cmd.Table, "id", cmd.Key)
	if err != nil {
		txn.Abort()
		return fmt.Errorf("fsm: create lookup: %w", err)
	}
	if existing != nil {
		txn.Abort()
		return fmt.Errorf("fsm: create conflict: key %q already exists", cmd.Key)
	}

	entry := f.makeEntry(cmd)
	if err := txn.Insert(cmd.Table, entry); err != nil {
		txn.Abort()
		return fmt.Errorf("fsm: create insert: %w", err)
	}
	txn.Commit()
	return nil
}

func (f *FSM) applyUpdate(cmd *Command) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	txn := f.db.Txn(true)
	entry := f.makeEntry(cmd)
	if err := txn.Insert(cmd.Table, entry); err != nil {
		txn.Abort()
		return fmt.Errorf("fsm: update insert: %w", err)
	}
	txn.Commit()
	return nil
}

func (f *FSM) applyDelete(cmd *Command) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	txn := f.db.Txn(true)
	existing, err := txn.First(cmd.Table, "id", cmd.Key)
	if err != nil {
		txn.Abort()
		return fmt.Errorf("fsm: delete lookup: %w", err)
	}
	if existing == nil {
		txn.Abort()
		return nil // no-op, key doesn't exist
	}
	if err := txn.Delete(cmd.Table, existing); err != nil {
		txn.Abort()
		return fmt.Errorf("fsm: delete: %w", err)
	}
	txn.Commit()
	return nil
}

// Snapshot returns an FSMSnapshot that captures the current state of
// all tables.  Called by Raft for log compaction.
func (f *FSM) Snapshot() (hraft.FSMSnapshot, error) {
	start := time.Now()

	f.mu.RLock()
	defer f.mu.RUnlock()

	var entries []snapshotEntry

	txn := f.db.Txn(false)
	defer txn.Abort()

	for _, table := range f.tables {
		it, err := txn.Get(table, "id")
		if err != nil {
			f.metrics.RecordSnapshotError()
			return nil, fmt.Errorf("fsm: snapshot table %q: %w", table, err)
		}
		for raw := it.Next(); raw != nil; raw = it.Next() {
			se := snapshotEntry{Table: table}

			// Handle both fsmEntry and kvFSMEntry types.
			switch e := raw.(type) {
			case *fsmEntry:
				se.Key = e.Key
				se.Namespace = e.Namespace
				se.Data = e.Data
			case *kvFSMEntry:
				se.Key = e.Key
				se.Data = e.Data
			default:
				// Fallback: JSON round-trip for unknown types.
				b, _ := json.Marshal(raw)
				var m map[string]json.RawMessage
				_ = json.Unmarshal(b, &m)
				if k, ok := m["Key"]; ok {
					_ = json.Unmarshal(k, &se.Key)
				}
				if d, ok := m["Data"]; ok {
					se.Data = d
				}
			}

			entries = append(entries, se)
		}
	}

	f.metrics.RecordSnapshot(time.Since(start))
	return &FSMSnapshot{entries: entries}, nil
}

// Restore replaces the entire FSM state from a snapshot.  Called on
// startup when recovering from a snapshot or when a follower receives
// a snapshot from the leader.
func (f *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var entries []snapshotEntry
	if err := json.NewDecoder(rc).Decode(&entries); err != nil {
		return fmt.Errorf("fsm: restore decode: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Rebuild: write all entries in a single transaction.
	txn := f.db.Txn(true)
	for _, se := range entries {
		var entry interface{}
		if se.Table == "_kv" {
			entry = &kvFSMEntry{
				Key:  se.Key,
				Data: se.Data,
			}
		} else {
			entry = &fsmEntry{
				Key:       se.Key,
				Namespace: se.Namespace,
				Data:      se.Data,
			}
		}
		if err := txn.Insert(se.Table, entry); err != nil {
			txn.Abort()
			return fmt.Errorf("fsm: restore insert %q: %w", se.Key, err)
		}
	}
	txn.Commit()

	f.metrics.RestoreCount.Add(1)
	if f.logger != nil {
		f.logger.Info("fsm: state restored", "entries", len(entries))
	}
	return nil
}

// fsmEntry is the envelope stored in go-memdb by the FSM.  It mirrors
// the memdbEntry in the store package but lives here to avoid a
// circular dependency.  When RaftStore wraps both, it will use the
// same underlying go-memdb instance.
type fsmEntry struct {
	Key       string
	Namespace string
	Data      []byte
}

// kvFSMEntry is the envelope for the _kv table.  It has the same
// fields as kvEntry in the store package so go-memdb type assertions
// work correctly when the store reads entries inserted by the FSM.
type kvFSMEntry struct {
	Key       string
	Namespace string    // unused for KV, but matches kvEntry schema
	Data      []byte
	ExpiresAt time.Time // zero means no expiry
}

// makeEntry creates the correct entry type for the given table.
// The _kv table uses kvFSMEntry; all others use fsmEntry.
func (f *FSM) makeEntry(cmd *Command) interface{} {
	if cmd.Table == "_kv" {
		return &kvFSMEntry{
			Key:  cmd.Key,
			Data: cmd.Data,
		}
	}
	return &fsmEntry{
		Key:       cmd.Key,
		Namespace: cmd.Namespace,
		Data:      cmd.Data,
	}
}

// snapshotEntry is the serialisation format for a single entry in a
// Raft snapshot.
type snapshotEntry struct {
	Table     string `json:"table"`
	Key       string `json:"key"`
	Namespace string `json:"namespace,omitempty"`
	Data      json.RawMessage `json:"data"`
}

// FSMSnapshot implements raft.FSMSnapshot.
type FSMSnapshot struct {
	entries []snapshotEntry
}

// Persist writes the snapshot to the given sink.
func (s *FSMSnapshot) Persist(sink hraft.SnapshotSink) error {
	data, err := json.Marshal(s.entries)
	if err != nil {
		sink.Cancel()
		return fmt.Errorf("fsm: snapshot persist encode: %w", err)
	}
	if _, err := sink.Write(data); err != nil {
		sink.Cancel()
		return fmt.Errorf("fsm: snapshot persist write: %w", err)
	}
	return sink.Close()
}

// Release is a no-op — we don't hold external resources.
func (s *FSMSnapshot) Release() {}
