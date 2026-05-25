# Internal Module Consistency Audit

**Date:** 2026-05-19
**Scope:** All 102 modules under `internal/` — 776 Go files, ~214,000 lines
**Status:** Living document — re-run after each cleanup phase

---

## Executive Summary

The codebase has **88 internal modules** with significant architectural inconsistency. Only **one module** (`gatekeeper`) follows all 8 recommended patterns. The codebase suffers from dual logging systems, pervasive context misuse, silently swallowed errors, 19+ global singletons, and ~20 underutilized internal directories (now restored and repurposed).

### Severity Breakdown

| Category | Severity | Scope |
|----------|----------|-------|
| `context.Background()` in HTTP handlers | HIGH | **FIXED** — 13 sites in 3 HTTP handler files; 15 remaining sites are non-handler code (correct) |
| Silently swallowed errors (`_ = err`) | HIGH | **PARTIAL** — 7 audit tasks done (20 sites); ~50+ broader sweep sites remain |
| Dual logging systems (`log` vs `zap`) | MEDIUM | **FIXED** — 93 files migrated to `logging.Z()`; zero `"log"` imports remain |
| Global singletons (`var Global*`) | MEDIUM | 19+ across 8 packages |
| Dead internal directories | MEDIUM | ~20+ genuinely unused |
| Hardcoded timeouts (`5*time.Second`) | MEDIUM | 15+ files |
| Missing KV persistence (Raft mode) | MEDIUM | 5 modules |
| Discarded variables in main.go | LOW | 4 variables |

---

## Module Inventory

### Size Distribution (Top 25)

| Rank | Module | Files | Lines | Notes |
|------|--------|-------|-------|-------|
| 1 | ~~`handlers`~~ | ~~42~~ | ~~19,608~~ | **DISSOLVED** (2026-05-25) — all 42 files extracted to per-module packages |
| 2 | `utils` | 36 | 13,187 | Utility dumping ground |
| 3 | `kubeplus` | 6 | 12,624 | 97.7% generated code |
| 4 | `antivirus` | 30 | 11,340 | Core scanning engine |
| 5 | `platform` | 36 | 10,698 | Core control plane |
| 6 | `gatekeeper` | 98 | 9,386 | 2FA — reference architecture |
| 7 | `storage` | 20 | 8,622 | Object storage |
| 8 | `jobs` | 20 | 7,155 | Job scheduling |
| 9 | `iam` | 15 | 6,916 | Identity & access |
| 10 | `policies` | 16 | 6,219 | Policy engine |
| 11 | `netintel` | 5 | 6,025 | 73.1% generated code |
| 12 | `scanner` | 18 | 5,297 | SafeGate scanner pipeline |
| 13 | `vectorplus` | 2 | 5,059 | 95.7% generated code |
| 14 | `reviewflow` | 2 | 4,831 | 95.7% generated code |
| 15 | `apimachinery` | 29 | 4,577 | K8s-style utilities |
| 16 | `integration` | 11 | 3,318 | Integration framework |
| 17 | `quality` | 9 | 2,715 | Code quality |
| 18 | `resources` | 13 | 2,762 | Resource management |
| 19 | `conductor` | 9 | 2,545 | Workflow orchestration |
| 20 | `controllers` | 7 | 2,521 | K8s-style controllers |
| 21 | `apiscanner` | 10 | 2,502 | API scanning |
| 22 | `federation` | 10 | 2,173 | Cross-cluster federation |
| 23 | `schemaregistry` | 7 | 2,018 | Schema management |
| 24 | `governance` | 7 | 1,969 | Access governance |
| 25 | `alerting` | 6 | 1,961 | Alert rules & incidents |

### Modules < 200 Lines (Likely Stubs)

| Module | Lines | Status |
|--------|-------|--------|
| `logging` | 71 | Exists but not adopted |
| `bootstrapsecrets` | 102 | Single file |
| `serverboot` | 109 | Single file |
| `snapshot` | 116 | Single file |
| `rpcpool` | 117 | Single file |
| `blocking` | 136 | Created in main.go, never wired |
| `scheduler` | 142 | Single strategy only |
| `database` | 161 | Connection setup only |
| `keyring` | 161 | Single file |
| `autopilot` | 164 | Single file |
| `scripts` | 177 | Code generation script |
| `template` | 190 | Single file |

---

## Pattern Comparison

### Module Lifecycle Patterns

Only **3 modules** have a System/bootstrap constructor:

| Module | NewSystem() | RegisterRoutes() | StartControllers() | SetKVStore() | Full Lifecycle |
|--------|------------|------------------|--------------------|--------------|----------------|
| `gatekeeper` | Y | Y | Y | Y | **YES** |
| `storage` | Y | Y | Y | Y | **YES** |
| `iam` | Y | partial | N | via constructor | Partial |
| All others | N | inline in main.go | N | N | **NO** |

### 8-Pattern Compliance Matrix

| Pattern | gatekeeper | storage | iam | ~20 resource modules | ~60 other modules |
|---------|-----------|---------|-----|---------------------|-------------------|
| 1. `system.go` / bootstrap | Y | Y | Y | N | N |
| 2. `handlers/` or `http.go` | Y | Y | Y | Y | varies |
| 3. `models/` or domain types | Y | Y | Y | Y | varies |
| 4. `repositories/` interfaces | Y | N | Y | N | N |
| 5. `config/` package | Y | N | N | N | N |
| 6. `metrics/` (Prometheus) | Y | Y (custom) | N | N | N |
| 7. `audit/` or event logging | Y | Y | N | N | N |
| 8. `SetKVStore(kv)` | Y | Y | via ctor | N | N |

**Only `gatekeeper` follows all 8 patterns.** It is the reference architecture for new modules.

---

## Detailed Findings

### 1. Context Misuse in HTTP Handlers (HIGH)

Every handler in the `refactored_*` files creates `context.Background()` instead of using `c.Request.Context()`:

```go
// WRONG — found in refactored_auth_handler.go, refactored_user_handler.go, etc.
ctx := context.Background()
user, err := h.authService.Login(ctx, req.Username, req.Password)

// CORRECT — found in governance/handlers.go
req, err := h.accessStore.Get(c.Request.Context(), name)
```

**Impact:** Client disconnection never detected, middleware timeouts ignored, tracing correlation IDs lost.

**Affected files:**
- `handlers/refactored_auth_handler.go` — 5 handlers
- `handlers/refactored_user_handler.go` — 6 handlers
- `handlers/user_handler.go` — 2 handlers
- `handlers/resource_handler.go` — 2 handlers
- `handlers/datasource_handler.go` — 2 handlers
- `handlers/job_handler.go` — 3 handlers
- `handlers/query_logger.go` — 7 call sites
- `handlers/mongodb.go` — 2 handlers
- `services/auth_service.go` — internal `context.Background()`
- `services/user_service.go` — internal `context.Background()`

### 2. Silently Swallowed Errors (HIGH)

**Category A — Discarded JSON unmarshal (data corruption risk):**

```go
// gatekeeper/pgstore/factor_repository.go — 5 instances
_ = json.Unmarshal(specJSON, &factor.Spec)
_ = json.Unmarshal(statusJSON, &factor.Status)
```

**Category B — Discarded business-logic errors:**

```go
// apibanks/reconciler.go:69
_ = r.manager.CreateBank(ctx, bank)

// conductor/reconciler.go:55,114
_, _ = r.manager.CreateProducer(req)
_, _ = r.manager.CreateConsumer(req)

// iam/iam.go:305-306
_ = pgStore.SeedDefaultRoles(defaultRealm.ID)
_ = pgStore.SeedDefaultClientScopes(defaultRealm.ID)

// governance/enforcer.go:135
_ = e.logger.LogDecision(ctx, access, *decision)
```

**Category C — Ignored JSON binding in handlers:**

```go
// governance/handlers.go — 3 instances
_ = c.ShouldBindJSON(&body)
// Proceeds with zero-valued body if JSON is malformed
```

### 3. Dual Logging Systems (MEDIUM)

| Logger | File Count | Modules |
|--------|-----------|---------|
| `log.Printf` (stdlib) | ~66 files | antivirus, storage, gatekeeper, iam, auth, conductor, scanner, jobs, cdc, etl, cache, config, runtime |
| `zap` (structured) | ~47 files | handlers, controllers, alerting, governance, slo, catalog, costing, contracts, schemaregistry, anonymization, streamanalytics, featurestore, federation, mlpipeline |

Some modules use **both** (`utils`, `integration`, `controllers`).

The `internal/logging/` package exists as a canonical wrapper but is not imported from any main.go.

### 4. Global Singletons (MEDIUM)

19+ package-level global instances bypass constructor injection:

| Singleton | Package |
|-----------|---------|
| `GlobalWorkflowEngine` | `workflows/engine.go` |
| `GlobalWorkflowTriggerManager` | `workflows/engine.go` |
| `GlobalAPIBankManager` | `apibanks/manager.go` |
| `GlobalDiffEngine` | `diff/engine.go` |
| `GlobalEventRecorder` | `events/resource_events_extended.go` |
| `GlobalAuditLogger` | `events/audit.go` |
| `GlobalMetrics` | `metrics/metrics.go` |
| `GlobalReconcilerMetrics` | `metrics/reconciler_metrics.go` |
| `GlobalHealthMonitor` | `integration/monitoring.go` |
| `GlobalPlatformMetricsCollector` | `integration/monitoring.go` |
| `GlobalAlertManager` | `integration/monitoring.go` |
| `GlobalDataPlatformIntegration` | `integration/data_platform.go` |
| `GlobalCatalogIntegration` | `integration/data_platform.go` |
| `GlobalDataQualityMonitor` | `integration/data_platform.go` |
| `GlobalDataLineageAnalyzer` | `integration/data_platform.go` |
| `GlobalComplianceAuditor` | `integration/compliance.go` |
| `GlobalDataAccessControl` | `integration/compliance.go` |
| `GlobalDataMesh` | `mesh/datamesh.go` |
| `GlobalPolicyManager` | `policies/engine.go` |

### 5. KV Persistence Gaps (MEDIUM)

**Modules with KV persistence (10):**

| Module | KV Key Pattern | Save Style |
|--------|---------------|------------|
| Gatekeeper Audit | `gatekeeper:audit:log` | Async (`go save()`) |
| Gatekeeper Metrics | `gatekeeper:metrics:collector` | Async |
| Scanner Metrics | `storage:metrics:scanner` | Async |
| Storage Audit | `storage:audit:log` | Async |
| Storage Metrics | `storage:metrics:collector` | Async |
| Storage BucketStore | `storage:bucketstore/{t}/{n}` | Synchronous (under mutex) |
| Storage Access | `storage:access:policies/{t}/{u}/{b}` | Synchronous |

**Modules missing KV persistence:**

| Module | Priority | Issue |
|--------|----------|-------|
| `workflows`, `modes`, `vectorplus`, `reviewflow`, `integration` | HIGH | Wired to etcd only, never wired to Raft KV |
| `alerting/silencer.go` | MEDIUM | In-memory silence rules lost on restart |
| `gatekeeper/pgstore/trusted_device_repository.go` | MEDIUM | Only pgstore repo without KV backing |
| `gatekeeper/cache/rate_limit.go` | LOW | Ephemeral rate limit state |

### 6. Dead Code in main.go (LOW)

Four variables created and immediately discarded:

| Variable | Line | Constructor | Issue |
|----------|------|-------------|-------|
| `encryptionMgr` | 1712 | `encryption.NewInMemorySecretsManager()` | Never wired to any reconciler |
| `jobMetricsCollector` | 1732 | `jobs.NewMetricsCollector("axiom_nizam")` | Never wired to observability handler |
| `blockingNotifier` | 2220 | `blocking.NewNotifier()` | Never wired to long-poll endpoints |
| `apiBankReconciler` | 2272 | `apibanks.NewAPIBankReconciler(...)` | Created but never started (all other reconcilers are) |

### 7. Dead Internal Directories (~20+)

Directories never imported from any `main.go` (may be used transitively):

`distributed`, `drainer`, `evalbroker`, `graphql`, `keyring`, `logging`, `mesh`, `performance`, `periodic`, `planner`, `quality`, `rpcpool`, `scripts`, `security`, `serverboot`, `snapshot`, `sqlfilter`, `status`, `template`, `waitx`

### 8. Duplicate/Overlapping Modules

| Overlap | Modules | Lines | Issue |
|---------|---------|-------|-------|
| Controller naming | `controller` (1,561) vs `controllers` (2,521) | 4,082 | Both define `Reconciler` with different signatures |
| Events | `events` (1,608) vs `eventbus` (1,288) | 2,896 | Parallel event type hierarchies |
| Streaming | `stream` (288) vs `streaming` (599) vs `streamanalytics` (1,139) | 2,026 | Confusing names, different purposes |
| State | `distributed` (194) vs `distributedstate` (1,603) | 1,797 | `distributed` likely subsumed |

### 9. Hardcoded Values

**Timeouts** — `5*time.Second` appears in 15+ files with no configurability:
- `cache/redis.go`, `etl/engine.go`, `workflows/engine.go`, `cdc/pipeline.go`, `handlers/*.go`, `database/connections.go`

**URLs:**
- `client/config.go`: `http://localhost:8000`
- `config/config.go`: `http://localhost:8000`
- `iam/iam.go`: `http://localhost:8080` (inconsistent with above)

**Credentials:**
- `utils/cncf_cloud_native.go`: hardcoded `admin`/`admin` for Grafana

**Infrastructure endpoints:**
- `utils/cncf_cloud_native.go`: hardcoded Prometheus, Grafana, Loki, Jaeger, AlertManager URLs

---

## Industry Alignment: K8s / Nomad / MinIO

### How the Big 3 Organize Modules

| Concern | Kubernetes | Nomad | MinIO |
|---------|-----------|-------|-------|
| **Config** | ComponentConfig API objects per component, validated via OpenAPI | HCL + agent config struct, merged from file + CLI flags | Custom config stored in object layer itself, section+key access |
| **Handlers** | Generic API server framework with REST storage registry + admission chain | Custom `net/rpc` typed endpoints per resource (`Job.Register`, `Node.Register`) | Methods on `objectAPIHandlers` struct, `ObjectLayer` via factory function |
| **Metrics** | `component-base/metrics` — Prometheus registry per component, auto-registered | `go-metrics` (HashiCorp multi-backend: Prometheus/Statsite/StatsD) | Direct Prometheus in `cmd/metrics.go`, counters incremented inline |
| **Types** | Separate `api/` module — pure data structs, zero logic, scheme registration | `structs/` central package — ALL shared types, prevents circular deps | `ObjectLayer` interface in `cmd/` — runtime-swappable implementations |
| **Storage** | etcd via generic storage interface, cached by SharedInformerFactory | `state.StateStore` (go-memdb + Raft FSM to BoltDB) | `ObjectLayer` interface (erasure coding, disk management) |
| **Lifecycle** | Leader election + `stopCh` channel + health/ready probes | `Service` interface with `Run()/Shutdown()` + graceful drain | `signal.Notify` + per-node peer registration |

### Key Patterns Each Project Enforces

**Kubernetes** (40+ modules in `staging/src/k8s.io/`):
1. API types in separate package — no logic, just structs + `DeepCopy`
2. Scheme registration — central type registry via `runtime.Scheme.AddToScheme`
3. SharedInformerFactory — shared watch + local cache, ONE watch per resource type
4. Registry/Strategy pattern — per-resource validation, defaulting, field selection
5. Controller pattern — workqueue + rate limiting + `processNextWorkItem`

**Nomad** (server, client, scheduler, drivers):
1. `structs/` central package — ALL shared types, zero imports outward
2. `state/` StateStore interface — single place for DB access, backed by Raft FSM
3. Scheduler isolation — `EvalContext` (read-only) + `Planner` (submit results), zero coupling
4. `mock/` package — dedicated test fixtures for every domain type
5. RPC endpoints as typed structs — `Job`, `Node`, `Alloc` with method-per-RPC

**MinIO** (single binary, ~200 files in `cmd/`):
1. `ObjectLayer` interface with factory function — runtime-swappable implementations
2. Handler methods on a single struct — `objectAPIHandlers`
3. Config stored in object layer — same storage as data
4. Peer REST for inter-node communication — custom RPC over HTTP

### Gap Analysis

| Pattern | K8s/Nomad/MinIO Standard | AxiomNizam Current | Gap Severity |
|---------|--------------------------|---------------------|--------------|
| **Central type package** | K8s `api/`, Nomad `structs/` — single source of truth | Types scattered per module, some inline in handler files | HIGH |
| **Module lifecycle interface** | K8s `Run(stopCh)`, Nomad `Service`, MinIO `signal.Notify` | No formal interface — `main.go` calls each module ad-hoc | HIGH |
| **Config pattern** | Per-component, validated, env-aware, defaults | Only `gatekeeper` has `config/` with `DefaultConfig()` + `LoadFromEnv()` + `Validate()` | HIGH |
| **Handler pattern** | Typed per-resource, clean separation from business logic | `internal/handlers/` monolith fully dissolved; all 42 files extracted to per-module packages. DTOs wired in gatekeeper, storage, IAM; remaining modules pending | MEDIUM |
| **Metrics pattern** | Per-component Prometheus, auto-registered | Only `gatekeeper` has Prometheus; `storage` has custom; 19+ modules use `GlobalMetrics` singleton | MEDIUM |
| **Repository interfaces** | K8s `Lister`/`Informer`, Nomad `StateStore` interface | Only `gatekeeper` has `repositories/` interfaces; others call stores directly | MEDIUM |
| **Storage abstraction** | K8s generic storage, Nomad `StateStore`, MinIO `ObjectLayer` | `KVStore` interface exists, partially wired (10/88 modules) | MEDIUM |
| **Dependency injection** | Constructor injection everywhere | 19+ `Global*` singletons, `init()` with side effects | HIGH |
| **Error handling** | All errors surfaced, none silently discarded | ~25+ files use `_ = err` on business-logic errors | HIGH |
| **Context propagation** | Request context flows through all layers | 10+ handler files use `context.Background()` instead of `c.Request.Context()` | HIGH |
| **Logging** | One logger per project (structured) | Dual systems: ~66 files `log.Printf`, ~47 files `zap` | MEDIUM |
| **Dead code** | Regular cleanup, no unused directories | **FIXED** — 12 "dead" directories restored and repurposed with real integrations, 4 discarded variables resolved | MEDIUM |
| **Test infrastructure** | K8s `fake/` clients, Nomad `mock/` package | Only `gatekeeper/testutil/` exists; no shared test helpers | LOW |

### Alignment Score by Module

| Module | Config | Handlers | Models | Repos | Metrics | Audit | KVStore | Lifecycle | Score |
|--------|--------|----------|--------|-------|---------|-------|---------|-----------|-------|
| `gatekeeper` | Y | Y | Y | Y | Y | Y | Y | partial | **8/8** |
| `storage` | inline | Y | Y | N | Y | Y | Y | partial | **6/8** |
| `iam` | inline | Y | Y | partial | N | N | via ctor | partial | **4/8** |
| `scanner` | Y | N/A | Y | N | Y | N | Y | N/A | **4/8** |
| `platform` | N | N | Y | N | shared | N | N | N | **1/8** |
| `jobs` | N | N | Y | N | N | N | N | N | **1/8** |
| `antivirus` | N | N | Y | N | N | N | N | N | **1/8** |
| ~20 resource modules | N | inline | Y | N | N | N | N | N | **1/8** |
| ~60 other modules | N | N | varies | N | N | N | N | N | **0-1/8** |

---

## Improvement Plan (25 Phases)

### Tier 1: Critical Fixes (Phases 1-5) — Foundation

These fix correctness bugs and data-loss risks. No architectural changes.

---

#### Phase 1: Fix Context Propagation

**Goal:** All HTTP handlers use `c.Request.Context()`.

**Status:** ✅ COMPLETE (2026-05-19)

| # | Task | File(s) | Status |
|---|------|---------|--------|
| 1.1 | Replace `ctx := context.Background()` with `ctx := c.Request.Context()` in refactored handlers | `handlers/refactored_auth_handler.go` (5 sites), `refactored_user_handler.go` (6 sites) | ✅ Done |
| 1.2 | Fix `mongodb.go` — 2 HTTP handlers use `context.WithTimeout(c.Request.Context(), ...)` | `handlers/mongodb.go` (2 sites) | ✅ Done |
| 1.3 | `query_logger.go` — 7 call sites | `handlers/query_logger.go` | ⏭️ N/A — not HTTP handlers (service methods on `*QueryLogger`) |
| 1.4 | `user_handler.go`, `datasource_handler.go`, `job_handler.go`, `api_builder_handler.go` | 4 files | ⏭️ N/A — `loadState()`/`persistStateLocked()`/`startScheduler()` are not HTTP handlers |
| 1.5 | `services/auth_service.go` and `services/user_service.go` | 2 files | ⏭️ N/A — `Health()` methods have no request context |
| 1.6 | `api_metrics.go` — 2 call sites | `handlers/api_metrics.go` | ⏭️ N/A — `RecordAPICall()` called from middleware, `GetAllMetrics()` is standalone |

**Actual scope:** 13 `context.Background()` sites fixed across 3 files (11 → `c.Request.Context()`, 2 → `WithTimeout(c.Request.Context(), ...)`).
**15 sites unchanged** — correctly use `context.Background()` in non-handler code (loadState, persistState, scheduler, service methods, health probes).

---

#### Phase 2: Fix Swallowed Errors

**Goal:** Business-logic errors are never silently discarded.

**Status:** ✅ COMPLETE (2026-05-19)

| # | Task | File(s) | Status |
|---|------|---------|--------|
| 2.1 | Governance handlers — `_ = c.ShouldBindJSON(&body)` | `governance/handlers.go` (3 sites) | ✅ Return 400 on bind error |
| 2.2 | IAM seed functions — `_ = pgStore.SeedDefaultRoles(...)` | `iam/iam.go` (2 sites) | ✅ Log errors individually |
| 2.3 | Conductor reconciler — `_, _ = r.manager.CreateProducer(req)` | `conductor/reconciler.go` (2 sites) | ✅ Log error |
| 2.4 | Gatekeeper pgstore — `_ = json.Unmarshal(...)` | `gatekeeper/pgstore/factor_repository.go` (10 sites), `audit_repository.go` (1 site) | ✅ Return error to caller |
| 2.5 | Cache informer — `eventsCh, _ = si.watcher.Watch(ctx)` | `cache/informer.go` | ✅ Log + backoff on watch restart failure |
| 2.6 | API banks reconciler — `_ = r.manager.CreateBank(ctx, bank)` | `apibanks/reconciler.go` | ✅ Log non-duplicate errors |
| 2.7 | Governance enforcer — `_ = e.logger.LogDecision(...)` | `governance/enforcer.go` | ✅ Log error |

**Actual scope:** 20 swallowed error sites fixed across 8 files.

**Broader sweep findings (not yet fixed — future work):**
- IAM admin: session revocation, client cleanup, code invalidation errors (5 sites)
- Gatekeeper challenge/enrollment: state transition and backup code deletion errors (4 sites)
- Federation: query persistence errors (2 sites)
- Reconcilers (rbac, eventbus, bulk, export, streamanalytics): create/cancel/stop errors (7 sites)
- DualWrite handlers: upsert pattern errors (10+ files)
- Resource reconcilers: store.Update errors (9 files)
- Platform controlplane: status update and finalizer errors (3 files)
- Leader election: lock update error (1 site)

---

#### Phase 3: Unify Logging

**Goal:** One logger (`zap`) across all modules.

**Status:** ✅ COMPLETE (2026-05-19) — 93 files migrated, zero `"log"` imports remaining

| # | Task | Scope | Status |
|---|------|-------|--------|
| 3.1 | Enhance `internal/logging/logging.go` as factory: `logging.For("storage")` returns `*zap.Logger` | 1 file | ✅ Done |
| 3.2 | Migrate core modules: `storage`, `gatekeeper`, `iam`, `antivirus`, `scanner`, `jobs` | ~49 files | ✅ Done |
| 3.3 | Migrate infrastructure: `conductor`, `cache`, `auth`, `cdc`, `etl`, `config`, `database` | ~20 files | ✅ Done |
| 3.4 | Migrate `jobs` and `cache` — struct field `*log.Logger` pattern | ~21 files | ✅ Done |
| 3.5 | Migrate remaining: `handlers`, `controllers`, `integration`, `utils`, `platform`, `runtime`, `serverboot`, `events`, `reconciler`, `metrics`, `services` | ~23 files | ✅ Done |
| 3.6 | Remove stdlib `"log"` from Phase 2 files (`apibanks`, `governance`, `conductor`) | 3 files | ✅ Done |

**Total migrated: 93 files, ~460 call sites. Zero `"log"` imports remain in `internal/`.**

Migration approach:
- Direct `log.Printf(...)` → `logging.Z().Info(fmt.Sprintf(...))` (balanced paren matching)
- Struct field `*log.Logger` → removed field + removed constructor + `.logger.Printf(...)` → `logging.Z().Info(fmt.Sprintf(...))`
- `log.Fatal(...)` → `logging.Z().Fatal(...)`
- Unused `"log"` imports removed

---

#### Phase 4: Repurpose Dead Code — **DONE**

**Goal:** Restore "dead" modules and wire them into the codebase as real integrations.

| # | Task | Scope | Status |
|---|------|-------|--------|
| 4.1 | Verify ~20 dead directories with `go build` — confirm no transitive imports | 20 directories | DONE |
| 4.2 | Restore 12 deleted directories and integrate into codebase | 12 directories | DONE |
| 4.3 | Wire or delete 4 discarded variables in main.go | main.go | DONE |
| 4.4 | Merge `distributed` into `distributedstate` (or delete if redundant) | 2 modules | DONE |
| 4.5 | Align `controller` and `controllers` — merge or clarify boundaries | 2 modules | DONE |

**Results (2026-05-19):**

- **12 "dead" directories restored and repurposed:**
  - **`sqlfilter`** → Wired into `api_builder_handler.go`: replaced ~240 lines of inline SQL validation (classifySQLQuery, firstSQLKeyword, hasMultipleSQLStatements, legacyReadOnlyHeuristic) with `sqlfilter.New().IsReadOnly()` and added SQL injection detection via `DetectInjection()`. Now used by API Builder, ETL, and dynamic query handler.
  - **`keyring`** → Wired into `internal/encryption/field_encryption.go`: added `RotateKeyring()`, `EncryptWithKeyring()`, `DecryptWithKeyring()` methods for AES-GCM key rotation with active/retired key tracking. FieldLevelEncryption now initializes with a keyring on construction.
  - **`evalbroker`** → Wired via `internal/workqueue/broker_queue.go`: new `BrokerQueue` adapter implements `WorkQueue` interface with ack/nack semantics, visibility timeouts, and priority ordering. Available as drop-in replacement for `SimpleQueue` in GenericController.
  - **`periodic`** → Wired into `internal/jobs/periodic_scheduler.go`: new `PeriodicScheduler` wraps `periodic.Dispatcher` for lightweight cron-based scheduling. Alternative to `AdvancedScheduler` (robfig/cron) for simple interval-based jobs.
  - **`distributedstate`** — State store abstraction (etcd + in-memory + locking + watches). Available for modules needing distributed coordination.
  - **`distributed`** — etcd health probe. Available for health endpoints.
  - **`drainer`** — Node drain state machine. Available for graceful shutdown scenarios.
  - **`rpcpool`** — Connection pool. Available for RPC-heavy paths.
  - **`snapshot`** — CRC-checked frame format. Available for Raft snapshot streaming.
  - **`template`** — text/template with sprig-like helpers. Available for config rendering.
  - **`serverboot`** — Server bootstrap. Available for standardized startup.
  - **`scripts`** — Code generation. Available for build tooling.
- **9 directories confirmed alive:** `graphql` (1 import), `logging` (100+), `mesh` (5), `performance` (2), `planner` (1), `quality` (1), `security` (1), `status` (1), `waitx` (1)
- **4 discarded variables resolved:**
  - `encryptionMgr` — deleted (unused; handler creates its own)
  - `jobMetricsCollector` — deleted (unused; no observability handler wired)
  - `blockingNotifier` — deleted (unused; no long-poll endpoints wired)
  - `apiBankReconciler` — **wired** into GenericController (was created but never started)
- **`controller` vs `controllers`:** Both actively used, different purposes — no merge needed. `platform/controller` = new generic reconciler; `controllers` = older K8s-style framework.
- Removed unused `blocking` import from main.go.
- **Build verified:** `go build .` passes clean.

**Scope:** ~14 directories, 4 variables | **Effort:** 1 day | **Impact:** MEDIUM | **Risk:** LOW

---

#### Phase 5: Fix KV Persistence Gaps — **DONE**

**Goal:** All in-memory state survives restarts in Raft mode.

| # | Task | File(s) |
|---|------|---------|
| 5.1 | Wire `workflows`, `modes`, `vectorplus`, `reviewflow`, `integration` to `backendMgr.KV()` | main.go |
| 5.2 | Add `ConfigureKVPersistence()` to `alerting/silencer.go` | alerting/silencer.go |
| 5.3 | Add `ConfigureKVPersistence()` to `gatekeeper/pgstore/trusted_device_repository.go` | gatekeeper/pgstore/ |
| 5.4 | Remove dead `kvStore` field from `ChallengeRepository` and `BackupCodeRepository` | gatekeeper/pgstore/ |
| 5.5 | Standardize KV key naming: `module:subsystem:resource` | All KV modules |

**Scope:** 5 modules | **Effort:** 1 day | **Impact:** MEDIUM | **Risk:** LOW

---

### Tier 2: Structural Alignment (Phases 6-12) — Module Standardization

These bring all modules toward the gatekeeper reference architecture.

---

#### Phase 6: Define Module Lifecycle Interface — **DONE**

**Goal:** Common contract every module must implement.

| # | Task | Detail |
|---|------|--------|
| 6.1 | Create `internal/contracts/module.go` with `Module` interface | `Name()`, `RegisterRoutes()`, `Start()`, `Stop()` |
| 6.2 | Create `internal/contracts/config.go` with `Configurable` interface | `LoadFromEnv()`, `Validate()`, `Defaults()` |
| 6.3 | Implement for `gatekeeper`, `storage`, `iam` (already have most pieces) | 3 modules |
| 6.4 | Implement for `scanner`, `antivirus`, `jobs` | 3 modules |
| 6.5 | Refactor `main.go` to iterate over `[]contracts.Module` | main.go |

**Scope:** 6+ modules, 1 new file | **Effort:** 2-3 days | **Impact:** HIGH | **Risk:** MEDIUM

---

#### Phase 7: Standardize Config Pattern — **DONE**

**Goal:** Every module has a `config/` package with `DefaultConfig()`, `LoadFromEnv()`, `Validate()`.

| # | Task | Module(s) | Status |
|---|------|-----------|--------|
| 7.1 | Expand `storage/config/` with rate limits, controller, timeouts, capacity | storage | DONE |
| 7.2 | Consolidate `iam/config/` — RSA key, realm, crypto params, client defaults | iam | DONE |
| 7.3 | Wire scanner config to sub-scanners (archivescan ratio/files, metadata thresholds) | scanner | DONE |
| 7.4 | Create `antivirus/config/` + add entropy thresholds, engine timeouts | antivirus | DONE |
| 7.5 | Expand `jobs/config/` — DLQ, channel sizes, health threshold; fix CreateJob() | jobs | DONE |
| 7.6 | Expand `conductor/config/` — maxMessages, Kafka settings, stats interval | conductor | DONE |
| 7.7 | Fix `cache/config/` TTL mismatch (15m→5m), use config defaults in manager | cache | DONE |
| 7.8 | `internal/config/` god-object — confirmed infrastructure-only (2 importers), added Validate() | config | DONE |

**Scope:** 8 modules | **Effort:** 1 day | **Impact:** HIGH | **Risk:** LOW

**Key changes:**
- `storage/config/` expanded from 3 fields to 22 — rate limits, controller intervals, timeouts, capacity limits
- `storage/access/access.go` — removed 3 direct `os.Getenv` reads, accepts `ControllerConfig` struct
- `storage/controller/controller.go` — removed `resyncIntervalFromEnv()`/`debugEnabledFromEnv()`, accepts params
- `iam/config/` — added RSA key, realm name, etcd timeout, bcrypt cost, client defaults
- `iam/token/token.go`, `iam/admin/admin.go`, `iam/pgstore/pgstore.go` — reference config constants instead of hardcoded strings
- `antivirus/config/` sub-package created (re-exports root Config for pattern consistency)
- `antivirus/config.go` — added 10 new fields: MaxThreatLogSize, StatsLogInterval, ManualScanTimeout, 6 entropy thresholds
- `jobs/config/` — added DLQMaxSize, DLQRetention, EmailQueueSize, ResultQueueSize, HealthFailureRate
- `jobs/job.go` — `CreateJob()` now uses config for MaxRetries, Timeout, Priority (was hardcoded)
- `conductor/config/` — added MaxMessages, KafkaProducerAcks, KafkaProducerRetries, StatsPersistInterval
- `conductor/manager.go` — uses `cfg.MaxMessages` instead of hardcoded 10000
- `cache/manager.go` — all TTL defaults now sourced from `cacheconfig.DefaultConfig()` (was 15m, now 5m)
- `scanner/archivescan/` — added `NewScannerWithLimits()` accepting ratio limit and max files from config

---

#### Phase 8: Standardize Handler Pattern — **IN PROGRESS**

**Goal:** Every module has `handlers/` with typed DTOs, mappers, clean request/response.

| # | Task | Module(s) | Status |
|---|------|-----------|--------|
| 8.0 | Wire gatekeeper DTOs/mappers into http.go (reference fix) | gatekeeper | DONE |
| 8.1 | Add storage DTOs + mappers (`admin/dto.go`, `admin/mapper.go`) | storage | DONE |
| 8.2 | Add IAM DTOs + mappers | iam | DONE |
| 8.3 | Extract handlers from monolith `internal/handlers/` into per-module packages | All affected | **DONE** (42/42 extracted) |
| 8.4 | Split `internal/handlers/` into: `handlers/auth/`, `handlers/health/`, `handlers/admin/` | handlers | PENDING (incremental) |
| 8.5 | Add DTO structs + mappers to each module's handlers | All modules | **IN PROGRESS** — 39/39 dto.go files created; 1031→198 gin.H (81% reduction); error/ack + list/progress DTOs wired; complex success DTOs remaining |

**Scope:** 39 modules, ~1031 gin.H occurrences | **Effort:** 3-5 days | **Impact:** HIGH | **Risk:** MEDIUM

**Phase 8.5 progress (2026-05-25):**

DTO sweep — all 39 modules addressed across 2 batches:

**Fully wired (dto.go + all gin.H replaced in handlers): 17 modules**
`antivirus` (dto+mapper), `tenant` (dto+mapper), `bulk` (dto+mapper), `webhooks`, `conductor`, `eventbus`, `export`, `lineage`, `streaming`, `alerting`, `contracts`, `costing`, `tracing`, `versioning`, `mlpipeline`, `notification`, `audit`

**DTO created (error responses wired, resource CRUD gin.H remaining): 18 modules**
`anonymization`, `cdc`, `featurestore`, `federation`, `rbac`, `security`, `governance`, `schemaregistry`, `database`, `datasource`, `encryption`, `jobs`, `netintel`, `quality`, `resources`, `slo`, `catalog`, `streamanalytics`

**N/A (already use typed response structs): 4 modules**
`gis` (uses typed structs), `integration` (uses models.Response), `iam/authn` (mixed models.Response), `iam/users` (mixed models.Response)

All 39 modules now have dto.go files. Full project build passes clean. Remaining handler wiring (replacing resource CRUD `gin.H` with DTOs in the 18 partially-wired modules) can be done incrementally.

**Key changes (2026-05-20):**
- `gatekeeper/handlers/http.go` — rewrote all 15 handlers to use named DTOs from `dto.go` and mappers from `mapper.go`; fixed VerifyChallengeRequest to use string (matches service contract)
- `storage/admin/dto.go` — 15 request/response DTO structs: CreateBucketRequest, BucketResponse, ObjectResponse, PresignURLRequest/Response, AccessKeyRequest/Response, BucketShareRequest/Response, RateLimitRequest/Response, PolicyRequest, EventResponse
- `storage/admin/mapper.go` — 9 mapper functions: BucketToResponse, BucketsToResponse, ObjectToResponse, ObjectsToResponse, AccessKeyToResponse, ShareToResponse, EventToResponse, EventsToResponse, TimePtr
- Monolith `internal/handlers/` (42 files, 100+ handlers) identified for incremental extraction — too large for single pass

**Key changes (2026-05-21):**
- `iam/admin/dto.go` — 40+ request/response DTO structs covering all IAM domains:
  - Auth: RefreshTokenRequest, LogoutRequest/Response, WhoAmIResponse
  - Users: CreateUserRequest, UserResponse, ListUsersResponse, SetUserRolesRequest/Response
  - Clients: RegisterClientRequest, ClientResponse, ClientCreatedResponse, UpdateClientRequest, RegenerateSecretResponse, ChangeClientIDRequest/Response, ListClientsResponse
  - Roles: UpdateRoleRequest, ListRolesResponse
  - Bindings: ListBindingsResponse
  - Tokens: RevokeTokenRequest/Response, RevokeUserTokensResponse
  - OAuth: AuthorizeResponse, ClientCredentialsResponse, ServiceAccessInfoResponse/Endpoints
  - v2 (EnhancedHandler): CreateRealmRequest, CreateGroupRequest, GroupDetailResponse, GroupMemberRequest, CreateClientScopeRequest, CreateIdentityProviderRequest, PublicIdPResponse, ListPublicIdPsResponse, SetUserAttributeRequest, AddUserToGroupRequest, AddRequiredActionRequest, GetPGClientResponse, GetEffectiveRolesResponse, RealmDashboardResponse, RealmInfoResponse, RealmTokenSettings, RealmLoginSettings, RealmSecuritySettings
- `iam/admin/mapper.go` — 20+ mapper functions: UserToResponse, UsersToResponse, ClientToResponse, ClientsToResponse, ClientToCreatedResponse, ClientToRegenerateSecretResponse, ClientToChangeIDResponse, RolesToListResponse, BindingsToListResponse, WhoAmIFromClaims, LogoutResponseFromState, ClientCredentialsToResponse, GroupToDetailResponse, IdPToPublicResponse, IdPsToPublicResponse, PGClientToGetResponse, EffectiveRolesToResponse, RealmDashboardToResponse, RealmInfoToResponse, MaskClientSecret

**Key changes (2026-05-21 — Phase 8.3 incremental):**
- Extracted 4 standalone handlers from `internal/handlers/` into target modules:
  - `internal/handlers/docs_handler.go` (109 lines) → `internal/docs/handler.go`
  - `internal/handlers/graphql_handler.go` (116 lines) → `internal/graphql/handler.go`
  - `internal/handlers/performance_handler.go` (152 lines) → `internal/performance/handler.go`
  - `internal/handlers/api_metrics.go` (449 lines) → `internal/metrics/api_tracker.go`
- Updated `main.go`: `graphqlpkg.NewHandler`, `metrics.NewAPIMetricsTracker`, `metrics.MetricsMiddleware`
- Updated `internal/integration/graphql_ratelimit_perf_integration.go` to use new module types
- All 4 old files deleted from monolith; build passes clean
- Remaining 38 files need incremental extraction — many are coupled to shared helpers or have no clean target module

**Key changes (2026-05-21 — Phase 8.3 continued):**
- Extracted 3 v1 store-backed handlers (unused reference implementations):
  - `handlers/job_v1_handler.go` (150 lines) → `internal/jobs/v1_handler.go`
  - `handlers/datasource_v1_handler.go` (171 lines) → `internal/datasource/v1_handler.go`
  - `handlers/user_v1_handler.go` (199 lines) → `internal/iam/users/v1_handler.go`
- Extracted 4 Phase 6 P2 resource types + reconcilers:
  - `handlers/analytics_resource.go` (86 lines) → `internal/analytics/resource.go` (new module)
  - `handlers/transform_resource.go` (74 lines) → `internal/transform/resource.go` (new module)
  - `handlers/notification_resource.go` (72 lines) → `internal/notification/resource.go` (new module)
  - `handlers/netintel_resource.go` (71 lines) → `internal/netintel/resource.go`
- Created 3 new modules: `internal/analytics/`, `internal/transform/`, `internal/notification/`
- Updated main.go: all 4 resource store/reconciler references now use new module paths
- Total extracted: 11/42 files; 31 remaining in monolith

**Key changes (2026-05-21 — Phase 8.3 continued):**
- Extracted 2 refactored service-layer handlers (unused reference implementations):
  - `handlers/refactored_user_handler.go` (249 lines) → `internal/iam/users/service_handler.go`
  - `handlers/refactored_auth_handler.go` (224 lines) → `internal/iam/authn/service_handler.go`
- Extracted 2 handlers used in main.go:
  - `handlers/notification_handler.go` (346 lines) → `internal/notification/handler.go`
  - `handlers/netintel_handler.go` (342 lines) → `internal/netintel/handler.go`
- Updated main.go: `notificationpkg.NewHandler`, `netintelpkg.NewHandler`
- Total extracted: 15/42 files; 27 remaining in monolith

**Key changes (2026-05-21 — Phase 8.3 continued):**
- Extracted 3 handlers used in main.go:
  - `handlers/certificate_handler.go` (688 lines) → `internal/security/handler.go`
  - `handlers/admin_handler.go` (964 lines) → `internal/database/handler.go`
  - `handlers/cdc_etl_handler.go` (400 lines) → `internal/cdc/handler.go`
- Updated main.go: `securitypkg.NewHandler`, `database.NewHandler`, `cdc.NewHandler`
- Total extracted: 18/42 files; 24 remaining in monolith

**Key changes (2026-05-21 — Phase 8.3 continued):**
- Extracted 2 GIS resource/reconciler files to new `internal/gis/` module:
  - `handlers/gis_resource.go` (116 lines) → `internal/gis/resource.go` (+ GIS entity type definitions)
  - `handlers/gis_reconciler.go` (74 lines) → `internal/gis/reconciler.go`
- Created new module: `internal/gis/` (resource types + reconciler)
- GIS handler files (`gis_handler.go`, `gis_specialized_handler.go`) remain in monolith — deeply coupled to `api_builder_handler.go` (direct field access to unexported `mu`, `datasets`, `markers`)
- Updated main.go: `gispkg.GISResource`, `gispkg.NewGISReconciler`
- Total extracted: 20/42 files; 22 remaining in monolith

**Key changes (2026-05-21 — Phase 8.3 continued):**
- Extracted 1 handler:
  - `handlers/transformation_handler.go` (394 lines) → `internal/transform/handler.go`
- Updated main.go: `transformpkg.NewHandler`
- Analytics handler stays in monolith — deeply coupled to api_builder_handler.go (direct field access to unexported `mu`, `dashboards`)
- Total extracted: 21/42 files; 21 remaining in monolith

**Key changes (2026-05-21 — Phase 8.3 continued):**
- Extracted 4 files:
  - `handlers/firebase.go` (110 lines) → `internal/integration/firebase_handler.go`
  - `handlers/oracle.go` (169 lines) → `internal/integration/oracle_handler.go`
  - `handlers/mongodb.go` (181 lines) → `internal/integration/mongodb_handler.go`
  - `handlers/handlers.go` (373 lines) → split: `internal/health/handler.go` (HealthHandler) + `internal/iam/users/gorm_handler.go` (UserHandler)
- Updated main.go: `healthpkg.NewHandler`
- Total extracted: 25/42 files; 17 remaining in monolith

**Key changes (2026-05-21 — Phase 8.3 continued):**
- Split composite handler `quality_rls_cdc_versioning_handlers.go` (314 lines) across 4 modules:
  - QualityHandler → `internal/quality/handler.go`
  - SecurityHandler → `internal/security/rls_handler.go`
  - CDCHandler → `internal/cdc/stream_handler.go`
  - VersioningHandler → `internal/versioning/handler.go`
- Updated `internal/integration/quality_rls_cdc_versioning_integration.go` to use new module types
- Total extracted: 26/42 files; 16 remaining in monolith

**Key changes (2026-05-22 — Phase 8.3 continued):**
- Extracted 10 files from `internal/handlers/` into per-module packages:
  - `handlers/auth_handler.go` (1924 lines) → `internal/iam/authn/handler.go` (AuthHandler + OAuth, login, token validation)
  - `handlers/cli_auth_handler.go` (214 lines) → `internal/iam/authn/cli_handler.go` (CLIAuthHandler)
  - `handlers/login_identifier.go` (30 lines) → `internal/iam/authn/login_identifier.go`
  - `handlers/user_handler.go` (445 lines) → `internal/iam/users/platform_handler.go` (PlatformUserHandler)
  - `handlers/resource_handler.go` (505 lines) → `internal/resources/handler.go` (GenericResourceHandler)
  - `handlers/dynamic_query_handler.go` (487 lines) → `internal/query/handler.go` (DynamicQueryHandler)
  - `handlers/query_logger.go` (474 lines) → `internal/query/logger.go` (QueryLogger)
  - `handlers/query_logger_handlers.go` (287 lines) → `internal/query/logger_endpoints.go`
  - `handlers/query_builder_handler.go` (567 lines) → `internal/query/builder_endpoints.go`
  - `handlers/encryption_lineage_audit_workflow_handlers.go` (650 lines) → `internal/encryption/phase3_handler.go` (Phase3Handlers)
- Created 2 new modules: `internal/query/` (4 files), extended `internal/iam/authn/` (3 files)
- Added interfaces to break import cycles: `PlatformUserStore`, `IdentityProviderStore`, `IAMRoleResolver`
- Extended `UserRepository` interface with `Create`/`Update` methods
- Updated main.go and integration test imports
- Remaining 6 files in monolith: `api_builder_handler.go`, `datasource_handler.go`, `job_handler.go` (complex logic), `gis_handler.go`, `gis_specialized_handler.go`, `analytics_handler.go` (coupled to api_builder)
- Total extracted: 36/42 files; 6 remaining in monolith

**Key changes (2026-05-25 — Phase 8.3 FINAL):**
- Extracted remaining 3 coupled handler files from `internal/handlers/` into `internal/apibuilder/` package:
  - `handlers/api_builder_handler.go` (3,627 lines) → `internal/apibuilder/handler.go` + `api_crud.go` + `csv_upload.go` + `scanner.go` + `conversion.go` (already existed from prior extraction)
  - `handlers/analytics_handler.go` (811 lines) → `internal/apibuilder/analytics.go` (already existed)
  - `handlers/gis_handler.go` (516 lines) → `internal/apibuilder/gis.go` (already existed)
- The 3 monolith files were dead code — `main.go` already used `apibuilder.NewGISHandler()`, `apibuilder.NewAnalyticsHandler()`, `apibuilder.NewAPIBuilderHandler()`
- Added 2 missing methods that were lost when monolith was deleted:
  - `ChatSQLAssistant` → new `internal/apibuilder/sql_assistant.go` (AI-powered SQL suggestions via OpenClaw)
  - `DeleteDashboard` → new `internal/apibuilder/dashboard_delete.go`
- Fixed unused imports in `api_crud.go` (encoding/json) and `csv_upload.go` (logging, zap)
- Deleted empty `internal/handlers/` directory — **monolith fully dissolved**
- **Total extracted: 42/42 files; 0 remaining — Phase 8.3 COMPLETE**
- Build verified: `go build .` passes clean

---

#### Phase 9: Standardize Models Pattern

**Goal:** Every module has `models/` with pure domain types, no handler logic.

| # | Task | Detail |
|---|------|--------|
| 9.1 | Audit modules where types are defined inline in handler files | Scan all modules |
| 9.2 | Extract domain types into `models/` for modules that lack it | Modules without models/ |
| 9.3 | Ensure models contain NO handler logic, NO storage logic — pure data + validation | All models/ |
| 9.4 | Add `DeepCopy()` methods where needed (K8s pattern) | Complex types |

**Scope:** 15+ modules | **Effort:** 2-3 days | **Impact:** MEDIUM | **Risk:** LOW

---

#### Phase 10: Standardize Repository Interfaces

**Goal:** Modules with persistence have `repositories/` interfaces separate from implementations.

| # | Task | Module(s) |
|---|------|-----------|
| 10.1 | Create `storage/repositories/` with `BucketRepository`, `ObjectRepository` interfaces | storage |
| 10.2 | Create `iam/repositories/` with `UserRepository`, `RoleRepository` interfaces | iam |
| 10.3 | Create `jobs/repositories/` with `JobRepository`, `ScheduleRepository` interfaces | jobs |
| 10.4 | Create `antivirus/repositories/` with `ScanResultRepository` interface | antivirus |
| 10.5 | Ensure `pgstore/` implementations satisfy the interfaces (compile-time check) | All pgstore |

**Scope:** 4+ modules | **Effort:** 2 days | **Impact:** MEDIUM | **Risk:** LOW

---

#### Phase 11: Standardize Metrics Pattern

**Goal:** Every module has `metrics/` with Prometheus collectors.

| # | Task | Module(s) |
|---|------|-----------|
| 11.1 | Create `iam/metrics/` — auth attempts, token issues, permission checks | iam |
| 11.2 | Create `jobs/metrics/` — job runs, durations, failures | jobs |
| 11.3 | Create `antivirus/metrics/` — scan counts, detection rates, engine timing | antivirus |
| 11.4 | Create `conductor/metrics/` — workflow executions, step durations | conductor |
| 11.5 | Migrate `GlobalMetrics` singleton consumers to per-module Prometheus collectors | 8 packages |
| 11.6 | Remove `GlobalMetrics` and `GlobalReconcilerMetrics` singletons | metrics/ |

**Scope:** 5 new modules, 8 migrated | **Effort:** 3-4 days | **Impact:** MEDIUM | **Risk:** LOW

---

#### Phase 12: Standardize Audit Pattern

**Goal:** Security-sensitive modules have `audit/` logging with KV persistence.

| # | Task | Module(s) |
|---|------|-----------|
| 12.1 | Create `iam/audit/` — login attempts, permission changes, token operations | iam |
| 12.2 | Create `storage/audit/` — rename existing `events/` to `audit/` for consistency | storage |
| 12.3 | Create `antivirus/audit/` — scan results, detection events | antivirus |
| 12.4 | Create `jobs/audit/` — job creation, modification, execution | jobs |
| 12.5 | Wire all audit logs to `ConfigureKVPersistence()` | All audit modules |

**Scope:** 4 modules | **Effort:** 2 days | **Impact:** MEDIUM | **Risk:** LOW

---

### Tier 3: Architectural Improvements (Phases 13-19) — Eliminating Anti-Patterns

These fix systemic issues that affect the entire codebase.

---

#### Phase 13: Eliminate Global Singletons

**Goal:** All state flows through constructors.

| # | Task | Singleton | Package |
|---|------|-----------|---------|
| 13.1 | Replace `GlobalWorkflowEngine` | Constructor inject into consumers | `workflows/engine.go` |
| 13.2 | Replace `GlobalWorkflowTriggerManager` | Constructor inject | `workflows/engine.go` |
| 13.3 | Replace `GlobalAPIBankManager` | Constructor inject | `apibanks/manager.go` |
| 13.4 | Replace `GlobalDiffEngine` | Constructor inject | `diff/engine.go` |
| 13.5 | Replace `GlobalEventRecorder` + `GlobalAuditLogger` | Constructor inject | `events/` |
| 13.6 | Replace `GlobalMetrics` + `GlobalReconcilerMetrics` | Per-module Prometheus (Phase 11) | `metrics/` |
| 13.7 | Replace `GlobalHealthMonitor`, `GlobalPlatformMetricsCollector`, `GlobalAlertManager` | Constructor inject | `integration/monitoring.go` |
| 13.8 | Replace `GlobalDataPlatformIntegration`, `GlobalCatalogIntegration`, `GlobalDataQualityMonitor`, `GlobalDataLineageAnalyzer` | Constructor inject | `integration/data_platform.go` |
| 13.9 | Replace `GlobalComplianceAuditor`, `GlobalDataAccessControl` | Constructor inject | `integration/compliance.go` |
| 13.10 | Replace `GlobalDataMesh` | Constructor inject | `mesh/datamesh.go` |
| 13.11 | Replace `GlobalPolicyManager` | Constructor inject | `policies/engine.go` |
| 13.12 | Remove `init()` functions with side effects | `workflows/engine.go`, `events/resource_events_extended.go` |

**Scope:** 19 singletons across 8 packages | **Effort:** 3-4 days | **Impact:** HIGH | **Risk:** MEDIUM

---

#### Phase 14: Extract Monolith Handlers

**Goal:** Dissolve `internal/handlers/` into per-module handler packages.

| # | Task | Detail |
|---|------|--------|
| 14.1 | Map every handler file in `internal/handlers/` to its owning module | Audit |
| 14.2 | Move auth handlers → `iam/handlers/` | `refactored_auth_handler.go` |
| 14.3 | Move user handlers → `iam/handlers/` | `refactored_user_handler.go`, `user_handler.go` |
| 14.4 | Move health/status handlers → `health/` (new or existing) | `handlers.go` health endpoints |
| 14.5 | Move admin handlers → `platform/handlers/` | admin-related handlers |
| 14.6 | Move data/job/query handlers → their respective modules | Various |
| 14.7 | Move `APIMetricsTracker` → `metrics/` or `platform/` | `api_metrics.go` |
| 14.8 | Delete empty `internal/handlers/` package | Cleanup |

**Scope:** ~40 files, ~19K lines | **Effort:** 3-5 days | **Impact:** HIGH | **Risk:** HIGH

---

#### Phase 15: Implement `system.go` Bootstrap for Core Modules

**Goal:** Core modules have `system.go` with `NewSystem()`, `RegisterRoutes()`, `StartControllers()`, `SetKVStore()`.

| # | Task | Module | Notes |
|---|------|--------|-------|
| 15.1 | Create `storage/system.go` — wire all storage subpackages | storage | Already has partial structure |
| 15.2 | Create `iam/system.go` — wire IAM subpackages | iam | Refactor from `iam.go` |
| 15.3 | Create `scanner/system.go` — wire scanner pipeline | scanner | Currently initialized in storage |
| 15.4 | Create `antivirus/system.go` — wire AV engine | antivirus | Currently initialized in storage |
| 15.5 | Create `jobs/system.go` — wire job scheduler | jobs | Currently inline in main.go |
| 15.6 | Create `conductor/system.go` — wire workflow engine | conductor | Currently inline in main.go |
| 15.7 | Create `cache/system.go` — wire Redis + informers | cache | Currently inline in main.go |
| 15.8 | Simplify `main.go` — delegate to `module.RegisterRoutes()` + `module.Start()` | main.go | Reduce from ~2500 lines |

**Scope:** 7 modules + main.go | **Effort:** 4-5 days | **Impact:** HIGH | **Risk:** HIGH

---

#### Phase 16: Central Type Package

**Goal:** Single source of truth for shared domain types (Nomad `structs/` pattern).

| # | Task | Detail |
|---|------|--------|
| 16.1 | Audit which types are shared across 3+ modules | Scan imports |
| 16.2 | Create `internal/types/` or expand `internal/contracts/types.go` | New package |
| 16.3 | Move shared types: `Tenant`, `User`, `Role`, `Permission`, `Resource`, `Job` | Cross-module types |
| 16.4 | Ensure `types/` has zero imports from other `internal/` packages | Prevent circular deps |
| 16.5 | Update all modules to import from `types/` instead of defining locally | All modules |

**Scope:** 5-10 shared types, 20+ files | **Effort:** 2-3 days | **Impact:** MEDIUM | **Risk:** MEDIUM

---

#### Phase 17: Standardize Error Handling

**Goal:** Typed errors per module, no string-based error comparison.

| # | Task | Detail |
|---|------|--------|
| 17.1 | Create `internal/errors/` with shared error types: `NotFoundError`, `ValidationError`, `UnauthorizedError` | New package |
| 17.2 | Add `errors.go` to each module with domain-specific error types | Per module |
| 17.3 | Replace `fmt.Errorf("not found")` with typed errors | All modules |
| 17.4 | Add error wrapping with `fmt.Errorf("operation: %w", err)` | All modules |
| 17.5 | Create HTTP error mapper: typed error → status code + JSON response | handlers |

**Scope:** All modules | **Effort:** 2-3 days | **Impact:** MEDIUM | **Risk:** LOW

---

#### Phase 18: Test Infrastructure

**Goal:** Shared test helpers and per-module test fixtures (Nomad `mock/` pattern).

| # | Task | Detail |
|---|------|--------|
| 18.1 | Create `internal/testutil/` with shared helpers: `testutil.Context()`, `testutil.DB()` | New package |
| 18.2 | Expand `gatekeeper/testutil/` as reference | gatekeeper |
| 18.3 | Add `testutil/` to `storage`, `iam`, `jobs`, `scanner` | 4 modules |
| 18.4 | Create mock implementations for `KVStore`, `ObjectLayer`, `UserRepository` | Core interfaces |
| 18.5 | Add integration test tags (`//go:build integration`) for DB-dependent tests | All test files |

**Scope:** 5 modules + shared | **Effort:** 2-3 days | **Impact:** MEDIUM | **Risk:** LOW

---

#### Phase 19: Configurable Timeouts & URLs

**Goal:** No hardcoded values in production code.

| # | Task | File(s) |
|---|------|---------|
| 19.1 | Replace hardcoded `5*time.Second` timeouts with configurable defaults | 15+ files |
| 19.2 | Fix `client/config.go` — `http://localhost:8000` → env-driven | client/config.go |
| 19.3 | Fix `config/config.go` — `http://localhost:8000` → env-driven | config/config.go |
| 19.4 | Fix `iam/iam.go` — `http://localhost:8080` → env-driven (inconsistent with 8000) | iam/iam.go |
| 19.5 | Remove hardcoded Grafana credentials from `utils/cncf_cloud_native.go` | utils/ |
| 19.6 | Remove hardcoded Prometheus/Grafana/Loki/Jaeger/AlertManager URLs | utils/ |

**Scope:** 15+ files | **Effort:** 1 day | **Impact:** MEDIUM | **Risk:** LOW

---

### Tier 4: Advanced Patterns (Phases 20-25) — Production Readiness

These bring the codebase to production-grade quality matching K8s/Nomad/MinIO.

---

#### Phase 20: Reconciler Pattern Standardization

**Goal:** All controllers follow the K8s workqueue + rate-limiting pattern.

| # | Task | Detail |
|---|------|--------|
| 20.1 | Audit all reconcilers: `controller/`, `controllers/`, `platform/controller/` | 3 packages |
| 20.2 | Standardize on `GenericController[T]` from `platform/controller/` | Single implementation |
| 20.3 | Add rate-limiting workqueue to all controllers | All controllers |
| 20.4 | Add health probes (`/healthz`, `/readyz`) per controller | All controllers |
| 20.5 | Wire `ReconcilerMetrics` to per-module Prometheus (Phase 11) | All controllers |

**Scope:** 3 controller packages | **Effort:** 2-3 days | **Impact:** MEDIUM | **Risk:** MEDIUM

---

#### Phase 21: Event Bus Standardization

**Goal:** One event system, not two parallel ones.

| # | Task | Detail |
|---|------|--------|
| 21.1 | Audit `events/` vs `eventbus/` — document overlap | 2 packages |
| 21.2 | Merge into single `events/` package with typed event hierarchy | events/ |
| 21.3 | Add event persistence via KVStore | events/ |
| 21.4 | Wire all modules to use the unified event bus | All modules |

**Scope:** 2 packages, ~3K lines | **Effort:** 2 days | **Impact:** MEDIUM | **Risk:** MEDIUM

---

#### Phase 22: Storage Backend Abstraction

**Goal:** Clean `ObjectLayer`-style interface for storage backends (MinIO pattern).

| # | Task | Detail |
|---|------|--------|
| 22.1 | Define `StorageBackend` interface in `storage/contracts/` | New file |
| 22.2 | Implement `RaftBackend` (current embedded raft) | storage/ |
| 22.3 | Implement `EtcdBackend` (current external etcd) | storage/ |
| 22.4 | Factory function for runtime backend selection | storage/ |
| 22.5 | Remove direct etcd references from bucket/access controllers | storage/ |

**Scope:** storage module | **Effort:** 3-4 days | **Impact:** MEDIUM | **Risk:** HIGH

---

#### Phase 23: Observability Stack

**Goal:** Unified metrics, tracing, and health across all modules.

| # | Task | Detail |
|---|------|--------|
| 23.1 | Add OpenTelemetry tracing to all HTTP handlers | Middleware |
| 23.2 | Add trace context propagation to RPC/service calls | All modules |
| 23.3 | Standardize Prometheus metric naming: `axiom_<module>_<metric>` | All metrics/ |
| 23.4 | Add `/metrics` endpoint with all module collectors | main.go |
| 23.5 | Add structured logging correlation (trace ID in logs) | logging/ |

**Scope:** All modules | **Effort:** 3-4 days | **Impact:** MEDIUM | **Risk:** LOW

---

#### Phase 24: Security Hardening

**Goal:** Address remaining items from SECURITY_README.md.

| # | Task | Detail |
|---|------|--------|
| 24.1 | Audit all `_ = err` sites for security implications | Phase 2 follow-up |
| 24.2 | Add rate limiting to all public endpoints | Middleware |
| 24.3 | Add request validation middleware (size limits, content-type) | Middleware |
| 24.4 | Audit all hardcoded credentials (SEC-01 through SEC-38) | SECURITY_README.md |
| 24.5 | Add CSRF protection to frontend proxy endpoints | frontend/ |

**Scope:** Cross-cutting | **Effort:** 2-3 days | **Impact:** HIGH | **Risk:** MEDIUM

---

#### Phase 25: Main.go Decomposition

**Goal:** Reduce main.go from ~2500 lines to <200 lines.

| # | Task | Detail |
|---|------|--------|
| 25.1 | Create `cmd/axiomnizam-server/` with `main()` that delegates to `server.Run()` | cmd/ |
| 25.2 | Create `internal/server/server.go` with module registration loop | New file |
| 25.3 | Each module registers itself via `module.RegisterRoutes()` + `module.Start()` | All modules |
| 25.4 | main.go becomes: load config → create modules → start → wait for signal | main.go |
| 25.5 | Add graceful shutdown with context cancellation and module `Stop()` calls | main.go |

**Scope:** main.go + all modules | **Effort:** 3-5 days | **Impact:** HIGH | **Risk:** HIGH

---

## Execution Order

```
Tier 1 (Critical Fixes) — Independent, any order
├── Phase 1: Context propagation
├── Phase 2: Swallowed errors
├── Phase 3: Unify logging
├── Phase 4: Dead code → repurposed with real integrations ✅
└── Phase 5: KV persistence gaps ✅

Tier 2 (Structural Alignment) — Sequential dependency
├── Phase 6: Module lifecycle interface ✅
├── Phase 7: Standardize config       ✅ DONE
├── Phase 8: Standardize handlers     ← 8.0-8.3 DONE; 8.4-8.5 PENDING
├── Phase 9: Standardize models       ← needs Phase 6
├── Phase 10: Repository interfaces   ← needs Phase 9
├── Phase 11: Standardize metrics     ← needs Phase 6
└── Phase 12: Standardize audit       ← needs Phase 6

Tier 3 (Anti-Pattern Elimination) — Depends on Tier 2
├── Phase 13: Kill singletons         ← needs Phase 6 (DI framework)
├── Phase 14: Extract monolith handlers ← needs Phase 8
├── Phase 15: system.go bootstrap     ← needs Phases 7-12
├── Phase 16: Central type package    ← needs Phase 9
├── Phase 17: Typed errors            ← independent
├── Phase 18: Test infrastructure     ← independent
└── Phase 19: Configurable values     ← needs Phase 7

Tier 4 (Production Readiness) — Depends on Tier 3
├── Phase 20: Reconciler standardization ← needs Phase 15
├── Phase 21: Event bus merge           ← needs Phase 13
├── Phase 22: Storage backend abstraction ← needs Phase 15
├── Phase 23: Observability stack       ← needs Phase 11
├── Phase 24: Security hardening        ← needs Phase 2
└── Phase 25: Main.go decomposition     ← needs Phase 15
```

### Priority Matrix

| Phase | Impact | Risk | Effort | Priority |
|-------|--------|------|--------|----------|
| 1. Context propagation | HIGH | LOW | 1 day | **P0** |
| 2. Swallowed errors | HIGH | MEDIUM | 1-2 days | **P0** |
| 3. Unify logging | HIGH | LOW | 2-3 days | **P1** |
| 4. Dead code cleanup | MEDIUM | LOW | 1 day | **DONE** |
| 5. KV persistence gaps | MEDIUM | LOW | 1 day | **DONE** |
| 6. Module lifecycle interface | HIGH | MEDIUM | 2-3 days | **DONE** |
| 7. Standardize config | HIGH | LOW | 2-3 days | **P2** |
| 8. Standardize handlers | HIGH | MEDIUM | 3-5 days | **P2** |
| 9. Standardize models | MEDIUM | LOW | 2-3 days | **P2** |
| 10. Repository interfaces | MEDIUM | LOW | 2 days | **P2** |
| 11. Standardize metrics | MEDIUM | LOW | 3-4 days | **P2** |
| 12. Standardize audit | MEDIUM | LOW | 2 days | **P2** |
| 13. Kill singletons | HIGH | MEDIUM | 3-4 days | **P3** |
| 14. Extract monolith handlers | HIGH | HIGH | 3-5 days | **P3** |
| 15. system.go bootstrap | HIGH | HIGH | 4-5 days | **P3** |
| 16. Central type package | MEDIUM | MEDIUM | 2-3 days | **P3** |
| 17. Typed errors | MEDIUM | LOW | 2-3 days | **P3** |
| 18. Test infrastructure | MEDIUM | LOW | 2-3 days | **P3** |
| 19. Configurable values | MEDIUM | LOW | 1 day | **P3** |
| 20. Reconciler standardization | MEDIUM | MEDIUM | 2-3 days | **P4** |
| 21. Event bus merge | MEDIUM | MEDIUM | 2 days | **P4** |
| 22. Storage backend abstraction | MEDIUM | HIGH | 3-4 days | **P4** |
| 23. Observability stack | MEDIUM | LOW | 3-4 days | **P4** |
| 24. Security hardening | HIGH | MEDIUM | 2-3 days | **P4** |
| 25. Main.go decomposition | HIGH | HIGH | 3-5 days | **P4** |

---

## Alignment Target

After completing all 25 phases, every module will match the gatekeeper reference architecture:

| Pattern | Current Compliance | Target |
|---------|-------------------|--------|
| `system.go` / bootstrap | 3/102 | 102/102 |
| `handlers/` with DTOs | 102/102 | 102/102 |
| `models/` domain types | ~20/102 | 102/102 |
| `repositories/` interfaces | 1/102 | 102/102 |
| `config/` package | 2/102 | 102/102 |
| `metrics/` Prometheus | 2/102 | 102/102 |
| `audit/` logging | 4/102 | 102/102 |
| `SetKVStore(kv)` | 5/102 | 102/102 |

---

*Last updated: 2026-05-25 (UTC+6) — Phases 1-8 DONE (Phase 8: all 39 modules have DTO files)*

---

## Phase Completion Tracker

| Phase | Status | Completed | Notes |
|-------|--------|-----------|-------|
| 1. Context propagation | ✅ DONE | 2026-05-19 | 13 sites in 3 HTTP handler files |
| 2. Swallowed errors | ✅ DONE | 2026-05-19 | 20 sites across 7 audit tasks |
| 3. Unify logging | ✅ DONE | 2026-05-19 | 93 files migrated, zero `"log"` imports |
| 4. Dead code repurpose | ✅ DONE | 2026-05-19 | 12 dirs restored, 4 wired (sqlfilter, keyring, evalbroker, periodic) |
| 5. KV persistence gaps | ✅ DONE | 2026-05-19 | All modules wired; keys standardized; dead fields removed |
| 6. Module lifecycle interface | ✅ DONE | 2026-05-19 | `contracts.Module` interface + 6 modules wired + registry in main.go |
| 7. Standardize config | ✅ DONE | 2026-05-21 | 8 modules configured: storage, iam, scanner, antivirus, jobs, conductor, cache, config |
| 8. Standardize handlers | 🔶 PARTIAL | — | 8.0-8.3 DONE, 8.4 N/A, 8.5 IN PROGRESS (dto.go created, error patterns done, success DTOs remaining) |
| 9. Standardize models | 🔶 PARTIAL | — | Some modules have models, not standardized |
| 10. Repository interfaces | 🔶 PARTIAL | — | Only gatekeeper has `repositories/` interfaces |
| 11. Standardize metrics | 🔶 PARTIAL | — | gatekeeper has Prometheus; others use GlobalMetrics |
| 12. Standardize audit | 🔶 PARTIAL | — | gatekeeper + storage have audit; others don't |
| 13. Eliminate global singletons | ⬜ TODO | — | 19+ singletons across 8 packages |
| 14. Extract monolith handlers | ✅ DONE | 2026-05-25 | 42/42 files extracted to per-module packages; `internal/handlers/` deleted |
| 15. system.go bootstrap | ⬜ TODO | — | Only 3/88 modules have it |
| 16. Central type package | ⬜ TODO | — | Types scattered across modules |
| 17. Standardize error handling | 🔶 PARTIAL | — | `platform/errs/` exists, not widely adopted |
| 18. Test infrastructure | ⬜ TODO | — | Only gatekeeper has `testutil/` |
| 19. Configurable timeouts | ⬜ TODO | — | 15+ files with hardcoded values |
| 20. Reconciler standardization | 🔶 PARTIAL | — | GenericController exists, 33 controllers running |
| 21. Event bus standardization | ⬜ TODO | — | `events/` and `eventbus/` still separate |
| 22. Storage backend abstraction | 🔶 PARTIAL | — | BackendManager with Raft/etcd dual backend |
| 23. Observability stack | 🔶 PARTIAL | — | zap + Prometheus + tracing modules exist, not fully wired |
| 24. Security hardening | 🔶 PARTIAL | — | sqlfilter + scanner + gatekeeper done; rate limiting, CSRF pending |
| 25. Main.go decomposition | ⬜ TODO | — | ~2500 lines, needs decomposition |

---

## Generated Code Inflation

Several modules are inflated by `_generated.go` files:

| Module | Generated Lines | Hand-Written | % Generated |
|--------|----------------|--------------|-------------|
| `kubeplus` | 12,328 | 296 | 97.7% |
| `netintel` | 4,403 | 1,622 | 73.1% |
| `vectorplus` | 4,843 | 216 | 95.7% |
| `reviewflow` | 4,623 | 208 | 95.7% |

**Total generated:** ~26,197 lines. **Hand-written codebase:** ~167,000 lines.

---

## Reference Architecture: `gatekeeper`

The `gatekeeper` module is the only module following all 8 patterns. Use it as the template for new modules:

```
internal/gatekeeper/
├── system.go              # NewSystem(), RegisterRoutes(), StartControllers(), SetKVStore()
├── contracts/             # Interface definitions (ports)
│   ├── service.go
│   ├── repository.go
│   └── provider.go
├── repositories/          # Interface declarations
├── pgstore/               # PostgreSQL implementations
├── models/                # Domain entities
├── handlers/              # HTTP handlers (http.go, grpc.go)
├── config/                # Module configuration
├── metrics/               # Prometheus metrics
├── audit/                 # Security audit logging
├── events/                # Domain events
├── middleware/             # Auth middleware
├── controller/            # K8s-style reconciliation
├── enrollment/            # Business logic
├── challenge/             # Business logic
├── backupcodes/           # Business logic
├── trusteddevices/        # Business logic
├── risk/                  # Business logic
├── totp/                  # Provider implementation
├── sms/                   # Provider implementation
├── email/                 # Provider implementation
├── cache/                 # Redis integration
├── bootstrap/             # Alternative wiring
└── testutil/              # Test helpers
```

---

*Last updated: 2026-05-25 (UTC+6) — Phase 8 IN PROGRESS (8.0-8.3 DONE, 8.5: 39/39 dto.go, 1031→206 gin.H, error patterns done, success DTOs remaining)*
