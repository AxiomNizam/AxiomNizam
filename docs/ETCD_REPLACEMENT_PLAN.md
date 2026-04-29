# AxiomNizam — etcd Replacement Plan: Nomad-Style Embedded Storage

**Date:** 2026-04-29  
**Status:** Complete — All 7 phases implemented  
**Author:** Platform Architecture Team

---

## Executive Summary

AxiomNizam currently uses etcd as its distributed state store for all reconciled resources. This document analyzes how HashiCorp Nomad achieves the same consistency guarantees without etcd, and proposes a phased plan to replace AxiomNizam's etcd dependency with a Nomad-inspired embedded storage layer using **Raft + go-memdb + BoltDB**.

**Why consider this?**
- etcd is an external dependency that must be deployed, monitored, and maintained separately
- Nomad proves that a K8s-style control plane can work with embedded storage
- Embedded storage simplifies deployment (single binary, no external cluster)
- Reduces operational complexity for self-hosted deployments

**Key decision:** This is a major architectural change. The plan is designed to be incremental and reversible at every phase.

---

## Part 1: Architecture Comparison

### 1.1 How AxiomNizam Uses etcd Today

AxiomNizam's etcd usage is concentrated in a well-defined abstraction layer:

| Layer | Component | etcd Usage |
|-------|-----------|------------|
| **Store abstraction** | `internal/platform/store/resource_store.go` | `EtcdStore[T]` — generic CRUD + Watch for all resource types |
| **Main server** | `main.go` | Creates `clientv3.Client`, passes to stores and managers |
| **Workflow engine** | `internal/workflows/engine.go` | Persists workflow execution state |
| **VectorPlus** | `internal/vectorplus/core.go` | Persists vector index state |
| **ReviewFlow** | `internal/reviewflow/core.go` | Persists review pipeline state |
| **Object storage** | `internal/storage/store/store.go` | Persists bucket metadata |
| **Object storage access** | `internal/storage/access/access.go` | Persists policies, access keys, shares |
| **IAM** | `internal/iam/storage/storage.go` | Persists tokens, sessions |
| **JWT secret** | `main.go` | Shared demo JWT secret via etcd CAS |

**Total etcd touchpoints:** ~10 files import `clientv3`

**Key observation:** The `ResourceStore[T]` interface in `internal/platform/store/resource_store.go` is the primary abstraction. All 27+ reconcilers interact with etcd exclusively through this interface. The other ~9 files use etcd directly for legacy reasons.

### 1.2 How Nomad Stores State (Without etcd)

Nomad uses a three-layer storage architecture:

```
┌─────────────────────────────────────────────┐
│              Application Layer              │
│  (Nomad server: jobs, allocs, evals, nodes) │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│           go-memdb (In-Memory DB)           │
│  • MVCC with immutable radix trees          │
│  • Rich indexing (compound, prefix, range)  │
│  • Watch channels for change notification   │
│  • Snapshot/restore for Raft FSM            │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│         hashicorp/raft (Consensus)          │
│  • Leader election                          │
│  • Log replication across server peers      │
│  • Snapshot compaction                      │
│  • FSM interface: Apply(log) → state change │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│      BoltDB / raft-boltdb (Durable Log)     │
│  • Raft log entries persisted to disk       │
│  • Snapshots stored as files                │
│  • Single-file embedded database            │
│  • No external process needed               │
└─────────────────────────────────────────────┘
```

**Key Nomad source files (from [github.com/hashicorp/nomad](https://github.com/hashicorp/nomad)):**

| File | Purpose | Relevance to AxiomNizam |
|------|---------|------------------------|
| `nomad/state/state_store.go` | go-memdb schema + CRUD operations | Replaces `EtcdStore[T]` |
| `nomad/fsm.go` | Raft FSM — applies log entries to state store | New component needed |
| `nomad/server.go` | Raft setup, leader election, peer management | New component needed |
| `nomad/state/schema.go` | Table/index definitions for go-memdb | Maps to AxiomNizam resource types |
| `nomad/snapshot.go` | Snapshot creation/restoration | New component needed |

**Key Go libraries used by Nomad:**

| Library | Purpose | License |
|---------|---------|---------|
| `github.com/hashicorp/go-memdb` | In-memory DB with MVCC, indexing, watches | MPL-2.0 |
| `github.com/hashicorp/raft` | Raft consensus implementation | MPL-2.0 |
| `github.com/hashicorp/raft-boltdb/v2` | BoltDB backend for Raft log storage | MPL-2.0 |
| `go.etcd.io/bbolt` | BoltDB fork (maintained by etcd team) | MIT |

### 1.3 Feature Comparison

| Feature | etcd (AxiomNizam today) | Nomad-style (Raft + go-memdb) |
|---------|------------------------|-------------------------------|
| Consistency | Linearizable (Raft-based) | Linearizable (Raft-based) |
| Deployment | External cluster (3-5 nodes) | Embedded in server binary |
| Watch/notify | etcd Watch API | go-memdb Watch channels |
| Transactions | etcd Txn (CAS) | go-memdb write transactions |
| Indexing | Key prefix only | Rich: compound, prefix, range, unique |
| Query capability | Get/List by prefix | Full table scans, index lookups, range queries |
| Snapshot/backup | etcd snapshot | Raft snapshot (automatic) |
| Max data size | Recommended <8GB | Limited by RAM (go-memdb is in-memory) |
| Persistence | etcd WAL + snapshots | BoltDB WAL + Raft snapshots |
| Operational overhead | High (separate cluster) | Low (embedded, single binary) |
| Multi-node | etcd cluster | Raft peer set (3-5 server nodes) |

---

## Part 2: Migration Strategy

### 2.1 Approach: Interface-First, Incremental Swap

The migration leverages AxiomNizam's existing `ResourceStore[T]` interface. Since all reconcilers already use this interface, we can swap the implementation without changing any reconciler code.

```
Phase 1: Build MemDBStore[T] implementing ResourceStore[T]
Phase 2: Build Raft FSM that applies mutations to MemDBStore
Phase 3: Build RaftStore[T] that wraps MemDBStore + Raft
Phase 4: Wire RaftStore into main.go behind a feature flag
Phase 5: Migrate direct etcd users (storage, IAM, workflows)
Phase 6: Remove etcd dependency
```

### 2.2 The Interface We Must Satisfy

From `internal/platform/store/resource_store.go`:

```go
type ResourceStore[T Resource] interface {
    Get(ctx context.Context, key string) (T, error)
    List(ctx context.Context, namespace string) ([]T, error)
    Create(ctx context.Context, obj T) error
    Update(ctx context.Context, obj T) error
    Delete(ctx context.Context, key string) error
    Watch(ctx context.Context) (<-chan WatchEvent[T], error)
    Close() error
}
```

This is the only interface that matters. Any implementation that satisfies it can replace etcd.

---

## Part 3: Detailed Implementation Plan

### Phase 1: go-memdb State Store (Week 1-2)

**Goal:** Build an in-memory store that satisfies `ResourceStore[T]` using go-memdb.

**New files:**

| File | Purpose |
|------|---------|
| `internal/platform/store/memdb_store.go` | `MemDBStore[T]` — go-memdb implementation of `ResourceStore[T]` |
| `internal/platform/store/memdb_schema.go` | Table/index schema definitions for all resource types |

**What to take from Nomad:**
- Schema pattern from `nomad/state/schema.go` — how to define tables with compound indexes
- Transaction pattern from `nomad/state/state_store.go` — how to wrap go-memdb transactions
- Watch pattern from `nomad/state/state_store.go` — how to use go-memdb's watch channels

**Key design decisions:**
- One go-memdb table per resource Kind (e.g., `"catalog_assets"`, `"quality_rules"`)
- Primary index: resource key (namespace/name)
- Secondary indexes: by kind, by namespace, by label selectors
- Watch channels map to go-memdb's built-in watch sets

**Estimated effort:** 3-4 days

### Phase 2: Raft FSM (Week 2-3)

**Goal:** Build a Raft Finite State Machine that applies log entries to the go-memdb store.

**New files:**

| File | Purpose |
|------|---------|
| `internal/platform/raft/fsm.go` | FSM implementation: Apply, Snapshot, Restore |
| `internal/platform/raft/commands.go` | Log entry command types (Create, Update, Delete) |
| `internal/platform/raft/snapshot.go` | Snapshot serialization/deserialization |

**What to take from Nomad:**
- FSM structure from `nomad/fsm.go` — Apply switch on command type, snapshot/restore
- Command encoding from `nomad/structs/structs.go` — msgpack encoding of log entries
- Snapshot pattern from `nomad/snapshot.go` — iterate all tables, serialize to snapshot

**Key design decisions:**
- Commands are msgpack-encoded structs with a type tag
- Apply() dispatches to the appropriate go-memdb mutation
- Snapshots serialize the entire go-memdb state to a binary blob
- Restore() rebuilds go-memdb from a snapshot

**Estimated effort:** 3-4 days

### Phase 3: Raft Server Setup (Week 3-4)

**Goal:** Initialize and manage a Raft cluster embedded in the AxiomNizam server.

**New files:**

| File | Purpose |
|------|---------|
| `internal/platform/raft/server.go` | Raft server initialization, peer management |
| `internal/platform/raft/config.go` | Configuration (data dir, peers, timeouts) |
| `internal/platform/raft/transport.go` | TCP transport for Raft RPC |

**What to take from Nomad:**
- Server setup from `nomad/server.go` — Raft initialization, bootstrap, peer join
- Transport from `nomad/server.go` — TCP transport with TLS support
- Configuration from `nomad/config.go` — sensible defaults for timeouts

**Key design decisions:**
- Single-node mode for development (no peers needed)
- 3-node mode for production (standard Raft quorum)
- Data directory: `AXIOMNIZAM_DATA_DIR/raft/` (BoltDB files)
- Bootstrap: first node auto-bootstraps, others join via peer address

**Estimated effort:** 3-4 days

### Phase 4: RaftStore[T] — The Unified Store (Week 4-5)

**Goal:** Build `RaftStore[T]` that wraps MemDBStore + Raft into a single `ResourceStore[T]` implementation.

**New files:**

| File | Purpose |
|------|---------|
| `internal/platform/store/raft_store.go` | `RaftStore[T]` — reads from go-memdb, writes through Raft |

**Behavior:**
- `Get()` / `List()` — read directly from go-memdb (fast, no Raft)
- `Create()` / `Update()` / `Delete()` — submit as Raft log entry, wait for commit
- `Watch()` — use go-memdb watch channels (triggered when FSM applies log)

**Key design decisions:**
- Reads are local (no Raft round-trip) — consistent because FSM is always up-to-date on leader
- Writes go through Raft for consensus — linearizable
- Stale reads on followers are acceptable for status/list endpoints (configurable)

**Estimated effort:** 2-3 days

### Phase 5: Feature-Flagged Integration (Week 5-6)

**Goal:** Wire `RaftStore` into `main.go` behind a feature flag, running alongside etcd.

**Changes to `main.go`:**

```go
var resourceStore store.ResourceStore[T]

if os.Getenv("STORAGE_BACKEND") == "raft" {
    // Use embedded Raft + go-memdb
    raftServer := raft.NewServer(raftConfig)
    resourceStore = store.NewRaftStore[T](raftServer)
} else {
    // Use etcd (current default)
    resourceStore = store.NewEtcdStore[T](etcdClient, prefix, factory)
}
```

**Dual-run validation:**
- Run both backends simultaneously
- Compare results on every read
- Log discrepancies
- Gradually shift traffic to Raft backend

**Estimated effort:** 2-3 days

### Phase 6: Migrate Direct etcd Users (Week 6-8)

**Goal:** Migrate the ~9 files that use etcd directly (not through ResourceStore).

| File | Current Usage | Migration Path |
|------|--------------|----------------|
| `internal/workflows/engine.go` | Persists workflow state | Move to ResourceStore or Raft KV |
| `internal/vectorplus/core.go` | Persists vector index | Move to ResourceStore |
| `internal/reviewflow/core.go` | Persists review pipeline | Move to ResourceStore |
| `internal/storage/store/store.go` | Bucket metadata | Move to ResourceStore |
| `internal/storage/access/access.go` | Policies, access keys | Move to ResourceStore |
| `internal/iam/storage/storage.go` | Tokens, sessions | Move to ResourceStore or Raft KV |
| `main.go` | JWT secret CAS | Move to Raft KV |

**Estimated effort:** 5-7 days

### Phase 7: Remove etcd Dependency (Week 8-9)

**Goal:** Remove etcd from docker-compose, go.mod, and all imports.

**Steps:**
1. Remove `etcd` service from `docker-compose.yml`
2. Remove `go.etcd.io/etcd/client/v3` from `go.mod`
3. Remove `EtcdStore[T]` from `internal/platform/store/resource_store.go` (or keep as optional backend)
4. Update documentation and deployment guides
5. Update `.env.example` to remove `ETCD_ENDPOINT`

**Estimated effort:** 1-2 days

---

## Part 4: Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Data loss during migration | Critical | Dual-run validation in Phase 5; snapshot before cutover |
| go-memdb RAM usage with 50+ resource types | Medium | Profile memory; go-memdb is efficient (immutable radix trees) |
| Raft complexity (leader election, split brain) | Medium | Use hashicorp/raft which is battle-tested in Nomad, Consul, Vault |
| Single-node mode loses durability | Medium | BoltDB provides crash recovery; recommend 3-node for production |
| Watch semantics differ between etcd and go-memdb | Medium | go-memdb watches are per-table, not per-key; adapter needed |
| Performance regression on writes (Raft round-trip) | Low | Raft writes are ~1-5ms; etcd writes are similar |
| Breaking change for existing deployments | High | Feature flag allows gradual rollout; etcd remains default initially |

---

## Part 5: What NOT to Change

These aspects of AxiomNizam's architecture remain unchanged:

- **ResourceStore[T] interface** — the contract stays the same
- **Reconciler pattern** — observe/diff/act/status loop is storage-agnostic
- **Resource format** — TypeMeta + ObjectMeta + Spec + Status
- **REST API handlers** — they call store methods, don't care about backend
- **CLI (axiomnizamctl)** — talks to REST API, not storage directly
- **Frontend dashboards** — talk to REST API

---

## Part 6: Timeline Summary

| Phase | Duration | Deliverable | Risk | Status |
|-------|----------|-------------|------|--------|
| 1. go-memdb store | 2 weeks | `MemDBStore[T]` passing all store tests | Low | ✅ Done |
| 2. Raft FSM | 1 week | FSM with Apply/Snapshot/Restore | Medium | ✅ Done |
| 3. Raft server | 1 week | Embedded Raft with single-node bootstrap | Medium | ✅ Done |
| 4. RaftStore | 1 week | Unified store: reads from memdb, writes through Raft | Low | ✅ Done |
| 5. Feature-flagged integration | 1 week | `STORAGE_BACKEND=raft` flag in main.go | Low | ✅ Done |
| 6. Migrate direct users | 2 weeks | All 9 direct etcd files migrated | Medium | ✅ Done |
| 7. Remove etcd | 1 week | etcd removed from docker-compose and go.mod | Low | ✅ Done |
| **Total** | **~9 weeks** | **etcd-free AxiomNizam** | |

---

## Part 7: Alternative Approaches Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| **Nomad-style (Raft + go-memdb)** | Battle-tested, embedded, no external deps | Complex to implement, RAM-bound | ✅ Recommended |
| **SQLite + Litestream** | Simple, single-file, replication via Litestream | No built-in consensus, weaker consistency | ❌ Insufficient for control plane |
| **BadgerDB** | Embedded KV, LSM-tree, good performance | No built-in consensus, no watch API | ❌ Missing features |
| **CockroachDB** | Distributed SQL, strong consistency | External dependency (heavier than etcd) | ❌ Defeats purpose |
| **Keep etcd, improve abstraction** | No migration risk, proven | Keeps external dependency | ⚠️ Fallback option |

---

## Part 8: Decision Criteria

Proceed with this migration if:

1. ✅ Self-hosted deployment simplicity is a priority
2. ✅ Reducing operational dependencies is valued
3. ✅ The team has capacity for 9 weeks of infrastructure work
4. ✅ The `ResourceStore[T]` interface is stable (it is)

Keep etcd if:

1. The team prefers operational familiarity with etcd
2. Multi-region federation is needed (etcd has better cross-region support)
3. The 9-week investment is not justified by deployment simplification

---

*Document maintained by Platform Architecture Team. Review before implementation begins.*
