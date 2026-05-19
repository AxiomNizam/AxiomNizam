# Internal Module Consistency Audit

**Date:** 2026-05-19
**Scope:** All 88 modules under `internal/` — 774 Go files, ~193,000 lines
**Status:** Living document — re-run after each cleanup phase

---

## Executive Summary

The codebase has **88 internal modules** with significant architectural inconsistency. Only **one module** (`gatekeeper`) follows all 8 recommended patterns. The codebase suffers from dual logging systems, pervasive context misuse, silently swallowed errors, 19+ global singletons, and ~20 dead internal directories.

### Severity Breakdown

| Category | Severity | Scope |
|----------|----------|-------|
| `context.Background()` in HTTP handlers | HIGH | 10+ handler files, ~40+ call sites |
| Silently swallowed errors (`_ = err`) | HIGH | ~25+ files |
| Dual logging systems (`log` vs `zap`) | MEDIUM | ~40 files use `log`, ~30 use `zap` |
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
| 1 | `handlers` | 42 | 19,608 | Monolith — all API handlers in one package |
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

## Improvement Plan

### Phase 1: Unify Logging

**Goal:** One logger (`zap`) across all modules.

1. Enhance `internal/logging/logging.go` as the factory: `logging.For("storage")` returns `*zap.Logger`
2. Migrate core modules first: `storage`, `gatekeeper`, `iam`, `antivirus`, `scanner`, `jobs`
3. Then infrastructure: `conductor`, `cache`, `auth`, `cdc`, `etl`, `config`
4. Remove stdlib `log` imports from all migrated files

**Scope:** ~66 files | **Effort:** 2-3 days | **Impact:** HIGH | **Risk:** LOW

### Phase 2: Fix Context Propagation

**Goal:** All HTTP handlers use `c.Request.Context()`.

1. Replace `ctx := context.Background()` with `ctx := c.Request.Context()` in all handlers
2. Where timeouts needed: `ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)`
3. Also fix `services/auth_service.go` and `services/user_service.go`

**Scope:** ~40+ call sites | **Effort:** 1 day | **Impact:** HIGH | **Risk:** LOW

### Phase 3: Fix Swallowed Errors

**Goal:** Business-logic errors are never silently discarded.

1. Governance handlers — `_ = c.ShouldBindJSON(&body)` → return 400 on error
2. IAM — `_ = pgStore.SeedDefaultRoles(...)` → log.Fatal or return error
3. Conductor reconciler — `_, _ = r.manager.CreateProducer(req)` → log error
4. Gatekeeper pgstore — `_ = json.Unmarshal(...)` → return error to caller
5. Cache informer — `eventsCh, _ = si.watcher.Watch(ctx)` → log and retry

**Scope:** ~25+ files | **Effort:** 1-2 days | **Impact:** HIGH | **Risk:** MEDIUM

### Phase 4: Clean Up Dead Code

**Goal:** Remove unused modules and discarded variables.

1. Delete ~14 genuinely unused directories (verify with `go build` first)
2. Wire or delete 4 discarded variables in main.go
3. Merge `distributed` into `distributedstate`
4. Align `controller` and `controllers` `Reconciler` interfaces

**Scope:** ~14 directories, 4 variables | **Effort:** 1 day | **Impact:** MEDIUM | **Risk:** LOW

### Phase 5: Fix KV Persistence Gaps

**Goal:** All in-memory state survives restarts in Raft mode.

1. Wire `workflows`, `modes`, `vectorplus`, `reviewflow`, `integration` to `backendMgr.KV()` in main.go
2. Add `ConfigureKVPersistence()` to `alerting/silencer.go`
3. Add `ConfigureKVPersistence()` to `gatekeeper/pgstore/trusted_device_repository.go`
4. Remove dead `kvStore` field from `ChallengeRepository` and `BackupCodeRepository`
5. Standardize KV key naming: `module:subsystem:resource` (no mixed slashes)

**Scope:** 5 modules | **Effort:** 1 day | **Impact:** MEDIUM | **Risk:** LOW

### Phase 6: Standardize Module Interface

**Goal:** All domain modules implement a common lifecycle interface.

```go
// internal/contracts/module.go
type Module interface {
    Name() string
    RegisterRoutes(rg *gin.RouterGroup)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

1. Implement for `gatekeeper`, `storage`, `iam` first
2. Retrofit the 12 backend-gated modules
3. Refactor main.go to iterate over `[]Module`

**Scope:** 15+ modules | **Effort:** 3-5 days | **Impact:** LOWER | **Risk:** HIGHER

### Phase 7: Eliminate Global Singletons

**Goal:** All state flows through constructors.

1. Replace all 19 `Global*` vars with constructor-injected instances
2. Remove `init()` functions with side effects (`workflows/engine.go`, `events/resource_events_extended.go`)

**Scope:** 19+ singletons across 8 packages | **Effort:** 2-3 days | **Impact:** LOWER | **Risk:** MEDIUM

### Phase 8: Extract Configurable Timeouts

**Goal:** No hardcoded `5*time.Second` in production code.

1. Add timeout fields to `config.Config`
2. Replace hardcoded timeouts with `cfg.DefaultTimeout`
3. Fix hardcoded URLs and credentials

**Scope:** 15+ files | **Effort:** 0.5 day | **Impact:** LOWER | **Risk:** LOW

---

## Execution Order

```
Phase 1 (Logging) ──┐
Phase 2 (Context) ──┤
Phase 3 (Errors) ───┼── Independent, any order
Phase 4 (Dead code) ┤
Phase 5 (KV gaps) ──┘
                     │
                     ▼
Phase 6 (Module interface) ── depends on cleaner codebase from 1-3
                     │
                     ▼
Phase 7 (Kill singletons) ── depends on Phase 6
                     │
Phase 8 (Timeouts) ──┘ Independent, can run anytime
```

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

*Last updated: 2026-05-19 (UTC+6)*
