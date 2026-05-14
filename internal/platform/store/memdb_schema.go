// Package store — memdb_schema.go
//
// Phase 1 of the etcd replacement plan: go-memdb table/index schema
// definitions for the AxiomNizam embedded storage layer.
//
// Each resource Kind maps to one go-memdb table.  Rather than defining
// a static table per Kind at compile time (which would require editing
// this file for every new resource), we use a single generic table
// name derived from the store's prefix and build the schema
// dynamically at store-creation time.
//
// Indexes:
//   - "id"        (primary)  — resource key  (namespace/name)
//   - "namespace" (secondary) — namespace prefix for List filtering
package store

import (
	"github.com/hashicorp/go-memdb"
)

// memdbTableSchema returns a go-memdb TableSchema for a single
// resource Kind identified by `tableName`.
//
// The table stores *memdbEntry values keyed by the resource's
// canonical key (namespace/name).
func memdbTableSchema(tableName string) *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: tableName,
		Indexes: map[string]*memdb.IndexSchema{
			// Primary index: unique resource key (e.g. "default/my-asset")
			"id": {
				Name:    "id",
				Unique:  true,
				Indexer: &memdb.StringFieldIndex{Field: "Key"},
			},
			// Secondary index: namespace for List(namespace) queries
			"namespace": {
				Name:         "namespace",
				Unique:       false,
				AllowMissing: true,
				Indexer:      &memdb.StringFieldIndex{Field: "Namespace"},
			},
		},
	}
}

// memdbEntry is the envelope stored in go-memdb.  We keep the raw
// JSON bytes alongside the extracted key and namespace so the indexes
// work without deserialising on every lookup.
type memdbEntry struct {
	// Key is the canonical resource key (namespace/name or just name).
	Key string

	// Namespace extracted from the resource for secondary indexing.
	Namespace string

	// Data is the JSON-encoded resource.  We store bytes rather than
	// the typed value because go-memdb is not generic — it stores
	// interface{} and we need to avoid type-assertion pitfalls across
	// different T instantiations.
	Data []byte
}

// newMemDBSchema builds a *memdb.DBSchema containing a single table
// for the given tableName.  Each MemDBStore instance creates its own
// go-memdb database with one table.
func newMemDBSchema(tableName string) *memdb.DBSchema {
	return &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			tableName: memdbTableSchema(tableName),
		},
	}
}

// NewMultiTableSchema builds a *memdb.DBSchema containing one table
// per entry in tableNames.  Used by the Raft server which needs a
// single shared go-memdb instance for all resource kinds.
//
// The special table name "_kv" gets the KV table schema (for direct
// key-value operations used by workflows, vectorplus, etc.).
func NewMultiTableSchema(tableNames []string) *memdb.DBSchema {
	tables := make(map[string]*memdb.TableSchema, len(tableNames))
	for _, name := range tableNames {
		if name == "_kv" {
			tables[name] = KVTableSchema()
		} else {
			tables[name] = memdbTableSchema(name)
		}
	}
	return &memdb.DBSchema{Tables: tables}
}
