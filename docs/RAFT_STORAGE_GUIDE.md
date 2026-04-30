# AxiomNizam — Embedded Raft Storage Guide

**Status:** Production-ready (all 7 phases complete)
**Date:** 2026-04-30

---

## Overview

AxiomNizam supports two storage backends for its control-plane state:

| Backend | Config | Deployment | Use Case |
|---------|--------|------------|----------|
| **Embedded Raft** | `STORAGE_BACKEND=raft` | Single binary, no external deps | Self-hosted, development, edge |
| **External etcd** | `STORAGE_BACKEND=etcd` | Requires etcd cluster | Existing deployments, multi-region |

The Raft backend uses HashiCorp's battle-tested libraries (same as Nomad, Consul, Vault):
- `hashicorp/raft` — Raft consensus protocol
- `hashicorp/go-memdb` — In-memory database with MVCC and rich indexing
- `hashicorp/raft-boltdb` — BoltDB-backed durable log storage

## Quick Start

### Single-Node (Development / Self-Hosted)

```bash
# Set the storage backend
export STORAGE_BACKEND=raft

# Start without etcd
docker compose up -d
```

That's it. The Raft node auto-bootstraps, elects itself leader, and stores state in `./data/raft/`.

### Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `STORAGE_BACKEND` | `etcd` | Storage backend: `raft` or `etcd` |
| `AXIOMNIZAM_RAFT_DATA_DIR` | `./data/raft` | Directory for BoltDB logs and snapshots |
| `AXIOMNIZAM_RAFT_BIND_ADDR` | `127.0.0.1:9700` | Raft RPC bind address |
| `AXIOMNIZAM_RAFT_NODE_ID` | `node-1` | Unique node identifier |
| `AXIOMNIZAM_RAFT_BOOTSTRAP` | `true` | Auto-bootstrap single-node cluster |

## Architecture

```
┌─────────────────────────────────────────────────┐
│              AxiomNizam Server                  │
│                                                 │
│  ┌──────────────────────────────────────────┐   │
│  │         ResourceStore[T] Interface       │   │
│  │  (Get, List, Create, Update, Delete,     │   │
│  │   Watch, Close)                          │   │
│  └──────────┬───────────────┬───────────────┘   │
│             │               │                   │
│  ┌──────────▼──────┐  ┌────▼──────────────┐    │
│  │   RaftStore[T]  │  │   EtcdStore[T]    │    │
│  │  (Raft backend) │  │  (etcd backend)   │    │
│  └──────────┬──────┘  └───────────────────┘    │
│             │                                   │
│  ┌──────────▼──────────────────────────────┐   │
│  │         go-memdb (In-Memory DB)         │   │
│  │  • Reads: direct from memdb (fast)      │   │
│  │  • Writes: through Raft consensus       │   │
│  └──────────┬──────────────────────────────┘   │
│             │                                   │
│  ┌──────────▼──────────────────────────────┐   │
│  │         hashicorp/raft (Consensus)      │   │
│  │  • Leader election                      │   │
│  │  • Log replication                      │   │
│  │  • Snapshot compaction                  │   │
│  └──────────┬──────────────────────────────┘   │
│             │                                   │
│  ┌──────────▼──────────────────────────────┐   │
│  │    BoltDB (Durable Log + Snapshots)     │   │
│  │  • Raft log entries on disk             │   │
│  │  • Periodic snapshots for recovery      │   │
│  └─────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
```

### How It Works

**Reads** go directly to go-memdb — no Raft round-trip, microsecond latency.

**Writes** are submitted as Raft log entries:
1. Client calls `Create()` / `Update()` / `Delete()` on `RaftStore[T]`
2. RaftStore encodes the mutation as a JSON command
3. Command is submitted to the Raft leader via `raft.Apply()`
4. Raft replicates the log entry to a quorum (1 node in single-node mode)
5. The FSM applies the mutation to go-memdb
6. RaftStore emits a watch event to subscribers

**Durability** is provided by BoltDB:
- Every Raft log entry is written to BoltDB before acknowledgment
- Periodic snapshots capture the full go-memdb state
- On restart, the FSM replays from the latest snapshot + subsequent log entries

## File Inventory

### Store Package (`internal/platform/store/`)

| File | Purpose |
|------|---------|
| `resource_store.go` | `ResourceStore[T]` interface + `EtcdStore[T]` implementation |
| `memdb_store.go` | `MemDBStore[T]` — go-memdb implementation |
| `memdb_schema.go` | go-memdb table/index schema definitions |
| `raft_store.go` | `RaftStore[T]` — reads from memdb, writes through Raft |
| `backend.go` | `BackendManager` + `NewStore[T]()` factory |
| `tables.go` | Central registry of all 40 resource table names |
| `kvstore.go` | `KVStore` interface for direct KV operations (IAM, workflows, etc.) |

### Raft Package (`internal/platform/raft/`)

| File | Purpose |
|------|---------|
| `commands.go` | Raft log entry command types (Create, Update, Delete) |
| `fsm.go` | Raft FSM — Apply, Snapshot, Restore |
| `config.go` | Server configuration with env var defaults |
| `server.go` | Raft server lifecycle, peer management, BoltDB setup |

### IAM Adapter (`internal/iam/storage/`)

| File | Purpose |
|------|---------|
| `kvadapter.go` | `iamBackend` interface + `kvStoreBackend` adapter for IAM repositories |

## What Still Uses etcd Directly

These modules have their own internal persistence that falls back gracefully when etcd is unavailable:

| Module | File | Behavior in Raft Mode |
|--------|------|----------------------|
| Workflows | `internal/workflows/engine.go` | In-memory only (state not persisted across restarts) |
| VectorPlus | `internal/vectorplus/core.go` | In-memory only |
| ReviewFlow | `internal/reviewflow/core.go` | In-memory only |
| Storage (buckets) | `internal/storage/store/store.go` | In-memory only |
| Storage (access) | `internal/storage/access/access.go` | In-memory only |
| NetIntel modes | `internal/netintel/modes/core.go` | In-memory only |

These modules work correctly in Raft mode — they just don't persist their internal state across server restarts. To add persistence, each module can be migrated to use `backendMgr.KV()` (the `KVStore` interface is already available).

## What's Fully Migrated

| Component | Stores | Backend |
|-----------|--------|---------|
| 30 reconciler controllers | `NewStore[T]()` factory | Raft or etcd via BackendManager |
| IAM (clients, roles, sessions, tokens) | `NewKV*Repository()` | Raft KVStore or etcd |
| JWT secret bootstrap | `KVStore.CAS()` | Raft KVStore or etcd |
| Platform managers (bulk, eventbus, etc.) | `platformStateStore` | In-memory (graceful nil-etcd) |

## Production Deployment Considerations

### Single-Node

Suitable for development, self-hosted, and edge deployments.

```bash
STORAGE_BACKEND=raft
AXIOMNIZAM_RAFT_DATA_DIR=/var/lib/axiomnizam/raft
AXIOMNIZAM_RAFT_BOOTSTRAP=true
```

**Durability:** BoltDB provides crash recovery. Data survives process restarts.

**Availability:** Single point of failure. If the node goes down, the service is unavailable until it restarts.

### Multi-Node (3-Node Cluster)

For production deployments requiring high availability. Requires 3 nodes minimum for Raft quorum.

```bash
# Node 1
AXIOMNIZAM_RAFT_NODE_ID=node-1
AXIOMNIZAM_RAFT_BIND_ADDR=10.0.0.1:9700
AXIOMNIZAM_RAFT_BOOTSTRAP=true

# Node 2
AXIOMNIZAM_RAFT_NODE_ID=node-2
AXIOMNIZAM_RAFT_BIND_ADDR=10.0.0.2:9700
AXIOMNIZAM_RAFT_BOOTSTRAP=false

# Node 3
AXIOMNIZAM_RAFT_NODE_ID=node-3
AXIOMNIZAM_RAFT_BIND_ADDR=10.0.0.3:9700
AXIOMNIZAM_RAFT_BOOTSTRAP=false
```

**Note:** Multi-node peer joining is implemented (`Server.AddPeer()`) but the HTTP API for dynamic peer management is not yet exposed. For now, peers must be added programmatically.

### Migration from etcd to Raft

1. Stop the server
2. Set `STORAGE_BACKEND=raft` in `.env`
3. Start the server — it will bootstrap with empty state
4. Re-create resources via the API or `axiomnizamctl apply`

**Note:** There is no automatic data migration from etcd to Raft. Resources must be re-created. For zero-downtime migration, run both backends in parallel (dual-write) and gradually shift traffic.

### Rollback to etcd

1. Stop the server
2. Set `STORAGE_BACKEND=etcd` (or remove the variable)
3. Ensure etcd is running and accessible
4. Start the server — it will use etcd as before

## Remaining Work (Optional Enhancements)

These are not required for the Raft backend to function but would improve the experience:

| Enhancement | Priority | Description |
|-------------|----------|-------------|
| Per-module KVStore migration | Medium | Migrate workflows, vectorplus, reviewflow, storage to use `backendMgr.KV()` for persistence across restarts |
| Multi-node peer management API | Medium | HTTP endpoints for adding/removing Raft peers dynamically |
| etcd-to-Raft data migration tool | Low | CLI tool to export etcd state and import into Raft |
| Raft cluster health endpoint | Low | `/health/raft` endpoint showing leader status, term, peers |
| Snapshot backup/restore CLI | Low | `axiomnizamctl raft snapshot` and `axiomnizamctl raft restore` |
| Remove etcd dependency entirely | Low | Remove `go.etcd.io/etcd/client/v3` from go.mod (requires migrating all ~30 files that still import it) |

---

*This document is maintained alongside the [ETCD Replacement Plan](ETCD_REPLACEMENT_PLAN.md).*
