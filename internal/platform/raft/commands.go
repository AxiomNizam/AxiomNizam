// Package raft — commands.go
//
// Phase 2 of the etcd replacement plan: Raft log entry command types.
//
// Every mutation (Create, Update, Delete) is encoded as a Command and
// submitted to the Raft leader.  The FSM's Apply method decodes the
// command and applies it to the go-memdb state store.
//
// Commands are JSON-encoded for simplicity and debuggability.  If
// throughput becomes a concern, switch to msgpack (Nomad's choice).
package raft

import (
	"encoding/json"
	"fmt"
)

// CommandType identifies the mutation kind in a Raft log entry.
type CommandType uint8

const (
	// CommandCreate inserts a new resource.  Fails if key exists.
	CommandCreate CommandType = iota + 1

	// CommandUpdate overwrites an existing resource.
	CommandUpdate

	// CommandDelete removes a resource by key.
	CommandDelete
)

// String returns a human-readable name for the command type.
func (c CommandType) String() string {
	switch c {
	case CommandCreate:
		return "Create"
	case CommandUpdate:
		return "Update"
	case CommandDelete:
		return "Delete"
	default:
		return fmt.Sprintf("Unknown(%d)", c)
	}
}

// Command is the envelope written to the Raft log.  The FSM
// deserialises this and dispatches based on Type.
type Command struct {
	// Type identifies the mutation (Create / Update / Delete).
	Type CommandType `json:"type"`

	// Table is the go-memdb table name (resource kind).
	Table string `json:"table"`

	// Key is the canonical resource key (namespace/name).
	Key string `json:"key"`

	// Namespace extracted from Key for secondary indexing.
	Namespace string `json:"namespace,omitempty"`

	// Data is the JSON-encoded resource (empty for Delete).
	Data []byte `json:"data,omitempty"`
}

// EncodeCommand serialises a Command to JSON bytes suitable for
// raft.Apply.
func EncodeCommand(cmd *Command) ([]byte, error) {
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("raft: encode command: %w", err)
	}
	return data, nil
}

// DecodeCommand deserialises a Command from JSON bytes received in
// a Raft log entry.
func DecodeCommand(data []byte) (*Command, error) {
	var cmd Command
	if err := json.Unmarshal(data, &cmd); err != nil {
		return nil, fmt.Errorf("raft: decode command: %w", err)
	}
	return &cmd, nil
}
