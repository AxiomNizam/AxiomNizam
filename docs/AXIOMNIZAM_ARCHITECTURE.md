# AxiomNizam Architecture

**Version:** 1.0
**Date:** 2026-04-27
**Status:** Living document — updated as the system evolves

---

## 1. What Is AxiomNizam?

AxiomNizam is a **declarative control-plane for data platform services** with workload orchestration primitives. It is not Kubernetes. It is not Nomad. It borrows the best patterns from both and combines them with its own innovations into a unified architecture for managing APIs, databases, pipelines, and platform services.

**One sentence:** AxiomNizam is what you get when you take K8s resource model + reconcile loops, add Nomad scheduling + deployment primitives, and apply them to a multi-database API platform instead of containers.

---

## 2. Architecture Principles

1. **Declarative over imperative** — Users declare desired state (Spec); controllers reconcile it into actual state (Status).
2. **Distributed state is the source of truth** — All platform state lives in the distributed state store (embedded Raft or external etcd, selected via `STORAGE_BACKEND`). External systems (SQL databases, Kafka, etc.) are reached only by reconcilers.
3. **Reconcile loops drive everything** — Status transitions happen in reconcilers, never in HTTP handlers.
4. **Feature-flagged migration** — Every architectural change is gated by env vars and can be rolled back instantly.
5. **Observe before acting** — Every reconciler is instrumented with metrics, structured logging, and health reporting before it touches production.

---

## 3. System Layers

```
┌─────────────────────────────────────────────────────────────────┐
│  LAYER 1: PRESENTATION                                          │
│  Frontend (Gin, port 7000) — role-based dashboards              │
│  CLI (axiomnizamctl) — kubectl-style resource management        │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  LAYER 2: API (Gin, port 8000)                                  │
│  Thin handlers → validate → write resource → return 202         │
│  Auth middleware → rate limiting → RBAC → metrics                │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  LAYER 3: DECLARATIVE STATE                                     │
│  ResourceStore[T] — typed, generic, watch-enabled persistence   │
│  Backends: RaftStore (embedded, default) or EtcdStore (external)│
│  Resources: TypeMeta + ObjectMeta + Spec + Status               │
│  40+ resource types across 22 modules                           │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  LAYER 4: CONTROL LOOP                                          │
│  GenericController[T] — watch → queue → worker → reconcile      │
│  33 controllers active, all instrumented                        │
│  InstrumentedReconciler — logging + metrics per call            │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  LAYER 5: RECONCILERS                                           │
│  Per-module: Observe → Diff → Act → Update Status               │
│  Drives imperative managers as internal implementation          │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  LAYER 6: ORCHESTRATION PRIMITIVES                              │
│  EvalBroker, PlanApplier, DeploymentController, Drainer,        │
│  HeartbeatTracker, PeriodicDispatcher, Autopilot,               │
│  ServiceRegistry, SnapshotFramer                                │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│  LAYER 7: EXTERNAL RUNTIMES                                     │
│  PostgreSQL, MySQL, MariaDB, Percona, Oracle, MongoDB           │
│  Redis/Valkey, Elasticsearch, RabbitMQ, Kafka                   │
│  Keycloak, ClamAV, OpenClaw/Ollama                              │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Patterns

### 4.1 Resource Model (from Kubernetes)

Every managed entity is a **Resource** with:

```go
type Resource struct {
    TypeMeta   // APIVersion + Kind
    ObjectMeta // Name, Namespace, UID, Generation, Labels, Annotations, Finalizers
    Spec       // Desired state (user-owned)
    Status     // Actual state (controller-owned)
}
```

Key properties:
- **Generation** increments on Spec changes
- **ObservedGeneration** records what the controller has seen
- **Conditions** provide structured status (Type, Status, Reason, Message)
- **Finalizers** prevent deletion until cleanup completes

### 4.2 Reconcile Loop (from Kubernetes)

```
Observe → Diff → Act → Update Status
```

Every reconciler implements:
```go
type Reconciler interface {
    Reconcile(ctx context.Context, obj Resource) ReconcileResult
}
```

Contract:
- Must be **idempotent** (safe to call multiple times)
- Must **update status** even on error
- Must return proper **requeue decisions**
- Must use **context deadline** for timeout

### 4.3 GenericController[T] (AxiomNizam-original)

A single reusable controller for any resource type:

```go
type GenericController[T store.Resource] struct {
    name       string
    store      store.ResourceStore[T]
    reconciler reconciler.Reconciler
    queue      *workqueue.SimpleQueue
    workers    int
    shadowMode bool
    metrics    *metrics.ReconcilerMetrics
}
```

Features:
- Watches `ResourceStore[T].Watch()` for create/update/delete events
- Enqueues resource keys into rate-limited work queue
- Worker goroutines dequeue and call `Reconcile(ctx, resource)`
- Handles `ReconcileResult.Requeue` / `RequeueAfter`
- Panic recovery — a crashing reconciler cannot kill the controller
- Initial sync on startup — lists all existing resources

### 4.4 Evaluation Broker (from Nomad)

Priority queue with explicit ack/nack semantics:

```go
type Broker struct {
    Enqueue(eval Evaluation)
    Dequeue() (Evaluation, bool)
    Ack(id string)
    Nack(id string, delay time.Duration) error
    DLQ() []Evaluation
}
```

Used for: scheduler evaluations, workflow steps, cross-region replication.

### 4.5 Plan Applier (from Nomad)

Serialized commit of scheduling decisions with optimistic concurrency:

```go
type Applier struct {
    UpsertNode(id string, total Resources)
    Snapshot() uint64
    Submit(ctx context.Context, plan Plan) (Result, error)
}
```

Plans reference a snapshot index. Stale plans get partial commits with rejected allocations.

### 4.6 Deployment Controller (from Nomad)

Canary/blue-green rollout with health gates:

```go
type Controller struct {
    UpdateAllocs(allocs []AllocState) (Decision, error)
    Promote() bool
    Fail(reason string) Decision
}
```

Phases: Pending → Running → Promoted → Succeeded/Failed.

### 4.7 Feature-Flagged Migration (AxiomNizam-original)

Three-stage migration from imperative to declarative:

```
Stage 1: RECONCILER_SHADOW_MODE=true
  → Controllers run, reconcilers execute, but don't mutate

Stage 2: DUAL_WRITE_<MODULE>=true
  → Handlers write to store AND call managers

Stage 3: RECONCILER_AUTHORITATIVE_<MODULE>=true
  → Handlers write to store only, reconciler drives manager
```

Rollback at any stage: flip the flag to false (instant, no deploy).

---

## 5. Module Inventory (100 modules)

### 5.1 Control-Plane Infrastructure

| Module | Purpose | Pattern |
|---|---|---|
| `reconciler` | Reconciler interface + StandardReconciler + InstrumentedReconciler | K8s |
| `workqueue` | Rate-limited work queues with exponential backoff | K8s |
| `informer` | SharedInformer with event handlers and cache sync | K8s |
| `cache` | ThreadSafeStore, InformerFactory, Redis/memory cache | K8s |
| `controller` | Controller builder, event handlers, leader election | K8s |
| `controllers` | Concrete controller implementations (workload, pipeline, schedule) | K8s |
| `runtime` | ControllerManager — multi-controller lifecycle | K8s |
| `apimachinery` | Conditions, finalizers, owners, labels, selectors, patches | K8s |
| `resources` | Base resource types (TypeMeta, ObjectMeta, ObjectStatus) | K8s |
| `platform/store` | ResourceStore[T] + EtcdStore, MemDBStore, RaftStore, KVStore, BackendManager | AxiomNizam |
| `platform/raft` | Embedded Raft server, FSM, BoltDB persistence | AxiomNizam |
| `platform/controller` | GenericController[T] — reusable watch+queue+worker | AxiomNizam |
| `platform/featureflags` | Per-module migration flags + storage backend selection | AxiomNizam |
| `platform/dualwrite` | Async best-effort store write helper | AxiomNizam |

### 5.2 Orchestration Primitives

| Module | Purpose | Pattern |
|---|---|---|
| `evalbroker` | Priority queue with ack/nack + visibility timeout + DLQ | Nomad |
| `planner` | Plan applier with optimistic concurrency | Nomad |
| `deployment` | Canary/blue-green deployment controller | Nomad |
| `drainer` | Batched node eviction with deadline | Nomad |
| `heartbeat` | TTL-based liveness tracker | Nomad |
| `periodic` | Min-heap cron dispatcher | Nomad |
| `autopilot` | Server health classification + voter promotion | Consul |
| `serviceregistry` | Service discovery with TTL health checks | Consul |
| `snapshot` | Length-prefixed CRC-checked frame streaming | Nomad |
| `scheduler` | Bin-pack scoring for workload placement | Nomad |

### 5.3 Platform Services (Reconciled)

| Module | Resource Type | Reconciler | Controller |
|---|---|---|---|
| `bulk` | BulkOperationResource | BulkOperationReconciler | GenericController |
| `eventbus` | TopicResource, SubscriptionResource | TopicReconciler, SubscriptionReconciler | GenericController |
| `export` | ExportJobResource | ExportJobReconciler | GenericController |
| `streaming` | StreamResource | StreamReconciler | GenericController |
| `rbac` | RoleResource, RoleBindingResource | RoleReconciler, RoleBindingReconciler | GenericController |
| `versioning` | VersionPolicyResource | VersionPolicyReconciler | GenericController |
| `tracing` | TracingConfigResource | TracingConfigReconciler | GenericController |
| `lineage` | LineageNodeResource | LineageNodeReconciler | GenericController |
| `audit` | AuditPolicyResource | AuditPolicyReconciler | GenericController |
| `encryption` | EncryptionKeyResource, EncryptionPolicyResource | EncryptionKeyReconciler, EncryptionPolicyReconciler | GenericController |
| `conductor` | ProducerResource, ConsumerResource | ProducerReconciler, ConsumerReconciler | GenericController |
| `webhooks` | WebhookResource | WebhookReconciler | GenericController |
| `tenant` | TenantV1Resource | TenantReconciler | GenericController |
| `jobs` | JobResource | JobController | GenericController |
| `etl` | PipelineResource | PipelineController | GenericController |
| `cdc` | CDCPipelineResource | CDCPipelineController | GenericController |
| `policies` | PolicyResource | PolicyReconciler | GenericController |
| `datasource` | DataSourceV1Resource | DataSourceReconciler | GenericController |
| `iam/users` | UserResource | UserReconciler | GenericController |
| `apiscanner` | APIScanResource | APIScanReconciler | GenericController |
| `apibanks` | APIBankResource | APIBankReconciler | GenericController |
| `storage` | BucketResource | BucketController | Dedicated |

### 5.4 API & Data Layer

| Module | Purpose |
|---|---|
| `handlers` | HTTP handlers for all API endpoints (36 files) |
| `apiserver` | Generic resource API server with extension routes |
| `graphql` | Dynamic GraphQL schema generation and execution |
| `database` | Multi-database connection management (5 SQL + MongoDB) |
| `sqlfilter` | SQL query validation, injection detection, complexity analysis |
| `scanner` | SafeGate file scanning pipeline (MIME, SVG, macro, ClamAV) |

### 5.5 Identity & Access

| Module | Purpose |
|---|---|
| `iam` | OIDC/OAuth identity provider (admin, authn, authz, token, users) |
| `auth` | JWT validation, JWKS refresh, rate limiting, role middleware |
| `security` | Row-level security policies |

### 5.6 Observability

| Module | Purpose |
|---|---|
| `metrics` | Prometheus metrics + ReconcilerMetrics + KeySpaceMonitor |
| `tracing` | Distributed tracing with OpenTelemetry |
| `logging` | Structured logging |
| `health` | Liveness/readiness probes |

### 5.7 Extension Modules

| Module | Purpose |
|---|---|
| `kubeplus` | Admission control, CRD validation, scheduler heuristics |
| `netintel` | Network intelligence with mode detectors |
| `vectorplus` | Vector similarity search and indexing |
| `reviewflow` | Staged review pipeline with quality checks |
| `integration` | Cross-module integration layers (Phase 1/2/3) |

---

## 6. Data Flow

### 6.1 Write Path (Reconciler-Authoritative)

```
User → POST /api/v1/bulk/operations
         │
         ▼
  BulkHandler.SubmitBulkOperation()
         │
         ├── if isAuthoritative():
         │     │
         │     ▼
         │   buildOperationResource(req)
         │     │
         │     ▼
         │   store.Create(ctx, BulkOperationResource)
         │     │
         │     ▼
         │   return 202 Accepted {"status": "Pending"}
         │
         │   [async — GenericController picks up the resource]
         │     │
         │     ▼
         │   ResourceStore.Watch() → WatchEvent{Type: Added}
         │     │
         │     ▼
         │   SimpleQueue.Add(key)
         │     │
         │     ▼
         │   Worker.processItem(key)
         │     │
         │     ▼
         │   InstrumentedReconciler.Reconcile(ctx, resource)
         │     │
         │     ├── log: "reconcile module=bulk key=default/op-123 ..."
         │     ├── metrics: RecordReconcile(duration, success, requeue)
         │     │
         │     ▼
         │   BulkOperationReconciler.Reconcile()
         │     │
         │     ├── Observe: read Spec from resource
         │     ├── Act: call manager.SubmitOperation()
         │     ├── Update Status: OperationStatus=Running, Progress=0
         │     ├── Requeue: true (poll progress)
         │     │
         │     ▼
         │   store.Update(ctx, resource) — status updated
         │
         └── else (fallback — old path):
               │
               ▼
             manager.SubmitOperation(op)
               │
               ▼
             dualWriteOperation(created) — best-effort store write
               │
               ▼
             return 200 OK
```

### 6.2 Read Path

```
User → GET /api/v1/bulk/operations/:id
         │
         ▼
  BulkHandler.GetOperation()
         │
         ▼
  manager.GetOperation(id) — reads from in-memory/store state
         │
         ▼
  return 200 OK {operation}
```

---

## 7. Deployment Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  docker-compose                                             │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │
│  │ axiomnizam  │  │  frontend   │  │  keycloak   │       │
│  │  :8000      │  │  :7000      │  │  :8080      │       │
│  │  :8001(rt)  │  │             │  │             │       │
│  └──────┬──────┘  └─────────────┘  └─────────────┘       │
│         │                                                   │
│  ┌──────▼──────┐  ┌─────────────┐  ┌─────────────┐       │
│  │   Raft/etcd │  │  postgres   │  │   valkey    │       │
│  │  :9700/:2379│  │  :5432      │  │  :6379      │       │
│  └─────────────┘  └─────────────┘  └─────────────┘       │
│                                                             │
│  ┌─────────────┐  (optional: openclaw profile)             │
│  │   clamav    │  ┌─────────────┐  ┌─────────────┐       │
│  │  :3310      │  │  openclaw   │  │   ollama    │       │
│  └─────────────┘  │  :18789     │  │  :11434     │       │
│                    └─────────────┘  └─────────────┘       │
└─────────────────────────────────────────────────────────────┘
```

---

## 8. Security Architecture

| Layer | Mechanism |
|---|---|
| Authentication | Keycloak OIDC/JWT with JWKS refresh |
| Authorization | Role-based middleware (admin, system-manager, user) |
| Rate limiting | Per-token call limits with configurable windows |
| SQL safety | `sqlfilter` module — injection detection, read-only enforcement, dialect validation |
| File scanning | SafeGate pipeline (MIME, SVG, macro, archive, ClamAV) |
| Encryption | AES-256-GCM field-level encryption with key rotation |
| Audit | Tamper-evident hash chain with compliance reporting |
| Secrets | Bootstrap secrets in Raft KV/PostgreSQL with env var override |

---

## 9. Observability

| Endpoint | Purpose |
|---|---|
| `GET /health` | Basic liveness |
| `GET /status` | All connection statuses |
| `GET /distributed` | Storage backend health (Raft leader status or etcd cluster) |
| `GET /health/reconcilers` | Per-module reconciler status + key-space metrics |

Metrics tracked per reconciler:
- TotalReconciles, TotalSuccesses, TotalErrors, TotalRequeues
- AvgDurationMs, LastDurationMs, ConsecutiveErrors
- Running, ShadowMode, Initialized

---

## 10. Key Numbers

| Metric | Value |
|---|---|
| Internal modules | 100 |
| Go files | 619 |
| Go lines | ~207,000 |
| Reconciler controllers active | 33 |
| GenericControllers | 29 |
| Resource types | 53+ |
| Store prefixes monitored | 30+ |
| Feature flags | 27+ |
| External integrations | 12 (5 SQL + MongoDB + Redis + ES + RabbitMQ + Kafka + Keycloak + ClamAV) |
| API route groups | 40+ |
| Frontend dashboards | 12 |

---

## 11. What Makes This Architecture Unique

1. **GenericController[T]** — One controller for all resource types via Go generics. K8s needs a separate controller per type.

2. **InstrumentedReconciler** — Automatic structured logging + per-module metrics for every reconcile call. Built-in, not bolted on.

3. **Feature-flagged migration** — Shadow → dual-write → authoritative, per module, with instant rollback. No other platform has this.

4. **ResourceStore[T] with Watch** — Simplified generic store combining K8s watch semantics with Go generics. Pluggable backends (Raft, etcd). No CRD registration needed.

5. **Multi-backend reconciliation** — Single reconciler pattern drives 5 SQL databases + MongoDB + Kafka + RabbitMQ. K8s reconcilers typically drive one API.

6. **Hybrid K8s + Nomad** — Declarative resources with reconcile loops (K8s) + evaluation-based scheduling with plan applier (Nomad) + canary deployments with health gates (Nomad) + TTL heartbeats with batched draining (Nomad).

7. **SQL filter engine** — Built-in injection detection (11 patterns), complexity analysis, dialect-aware validation, query normalization. Not a WAF — it's part of the control plane.

8. **Platform service reconciliation** — Bulk ops, event bus, exports, streaming, RBAC, versioning, lineage, tracing, encryption, conductor — all as declarative resources. No other platform does this for internal services.

---

## 12. Related Documents

- [MIGRATION_PLAN.md](./architecture/MIGRATION_PLAN.md) — Production migration plan (Phases 0-6)
- [ARCHITECTURE.md](../ARCHITECTURE.md) — Runtime architecture flowchart
- [SECURITY_README.md](../SECURITY_README.md) — Security posture and findings
- [SECURITY_AUDIT.md](./SECURITY_AUDIT.md) — Code-level security audit (38 findings)
- [README.md](../README.md) — Project overview and quick start
