# Handler Migration to Control-Plane Architecture

**Status:** In progress (Phase 0 observability complete)
**Owner:** Platform team
**Last updated:** 2026-04-25

**Recent changes (2026-04-25):**
- Wired 8 previously unwired modules: `apibanks`, `migrations`, `blocking`, `trivy`, `heartbeat`, `autopilot`, `deployment`, `serviceregistry`
- Added per-module reconciler metrics (`internal/metrics/reconciler_metrics.go`)
- Added `GET /health/reconcilers` endpoint
- Registered all 18 reconcilers with Phase 0 metrics tracking
- New routes: `/api/v1/apibanks`, `/api/v1/trivy`, `/api/v1/deployments`, `/api/v1/service-registry`, `/api/v1/heartbeat`, `/api/v1/autopilot/evaluate`


## Target architecture

Every HTTP resource in AxiomNizam must flow through the Kubernetes-style
control plane:

```
HTTP Handler (gin)
      │ thin; bind + validate + auth
      ▼
Service  (or directly → Store for pure CRUD)
      │ business rules, cross-resource invariants
      ▼
ResourceStore[T]  ──►  Informer/Watch  ──►  WorkQueue  ──►  Controller  ──►  Reconciler
      │                                                                        │
      ▼                                                                        ▼
 etcd (authoritative)                                                External Runtime
                                                                     (SQL DB, Kafka, etc.)
```

### Non-negotiable rules

1. **Handlers never hold `*gorm.DB`, a raw Mongo client, or `sync.Mutex`
   application state.** They depend on a `Service` or a `ResourceStore[T]`.
2. **etcd is the authoritative state.** External runtimes (Postgres, Kafka,
   Firebase, Oracle, …) are reached only by reconcilers, never by handlers.
3. **Status transitions happen in a reconciler**, never in a
   `go func() { time.Sleep(); status = "Ready" }()` inside a handler.
4. **Admin write endpoints enqueue work**; the work queue drives the
   controller which drives the reconciler.

### Canonical reference implementations

| Layer | File |
|---|---|
| Handler → etcd → reconciler | [internal/handlers/resource_handler.go](../../internal/handlers/resource_handler.go) |
| Handler → service → repo | [internal/handlers/refactored_user_handler.go](../../internal/handlers/refactored_user_handler.go), [internal/handlers/refactored_auth_handler.go](../../internal/handlers/refactored_auth_handler.go) |
| etcd-backed CRUD rewrite | [internal/handlers/datasource_v1_handler.go](../../internal/handlers/datasource_v1_handler.go) |
| ResourceStore primitive | [internal/distributedstate/etcd.go](../../internal/distributedstate/etcd.go) |
| Controller / reconciler plumbing | [internal/controllers/](../../internal/controllers/), [internal/reconciler/](../../internal/reconciler/), [internal/workqueue/](../../internal/workqueue/) |

---

## Audit — as of 2026-04-18

### ✅ Compliant

| Handler | Pattern |
|---|---|
| `resource_handler.go` | etcd + reconciler (canonical) |
| `refactored_user_handler.go` | handler → service → repo |
| `refactored_auth_handler.go` | handler → service → repo |
| `datasource_v1_handler.go` | etcd-backed, supersedes `datasource_handler.go` |
| `certificate_handler.go` | etcd-backed (verify) |
| `cdc_etl_handler.go` | etcd-backed (verify) |
| `conductor` routes | via dedicated manager |

### ⚠️ Acceptable exceptions

These hold `*gorm.DB` by design — they are passthrough / proxy layers, not
stateful resources:

| Handler | Reason |
|---|---|
| `dynamic_query_handler.go` | Dynamic SQL passthrough over user-selected DB |
| `graphql_handler.go` | Dynamic GraphQL schema over DB |
| `api_metrics.go` | In-memory ring buffer for metrics (lifetime = process) |

### ❌ Non-compliant — migration required

Priority:
- **P0** = handler actively diverges from the control plane and blocks other
  work
- **P1** = legacy, replaced by a refactored variant but still wired up
- **P2** = in-memory only, needs promotion to etcd-backed

| # | Handler | Violation | Priority | Target | Notes |
|---|---|---|---|---|---|
| 1 | `gis_train_handler.go` | ~~holds `*gorm.DB`, 11 endpoints direct to Postgres~~ | — | **DELETED 2026-04-18** | Rebuild via API Builder + DataSource |
| 2 | `gis_bdtrain_handler.go` | ~~holds `*gorm.DB`, multi-fallback queries in handler~~ | — | **DELETED 2026-04-18** | Rebuild via API Builder + DataSource |
| 3 | `gis_handler.go` | `sync.RWMutex` + in-memory slices | **P2** | `GISService` + `ResourceStore[GISLayer/Region/Marker/Dataset]` | Deprecated; API Builder depends on it |
| 4 | `gis_specialized_handler.go` | `sync.RWMutex` + in-memory map, seeded at startup | **P2** | `GISDashboard` resource + reconciler | Deprecated |
| 5 | `handlers.go` (`UserHandler`) | holds `*gorm.DB` | **P1** | replaced by `refactored_user_handler.go` | Remove registration + file |
| 6 | `user_handler.go` | `sync.RWMutex` + slices | **P1** | use `refactored_user_handler.go` | Remove registration + file |
| 7 | `auth_handler.go` | legacy | **P1** | use `refactored_auth_handler.go` | Verify callers, then remove |
| 8 | `datasource_handler.go` | `sync.RWMutex` + map | **P1** | use `datasource_v1_handler.go` | Verify no remaining mounts, then remove |
| 9 | `oracle.go` | holds `*gorm.DB` | **P0** | `OracleService` + `ResourceStore[OracleConnection]`; queries via reconciler | Or reclassify as passthrough like `dynamic_query_handler` |
| 10 | `mongodb.go` | raw Mongo driver in handler | **P0** | same pattern as Oracle |
| 11 | `firebase.go` | raw Firebase SDK in handler | **P0** | `FirebaseService` + `ResourceStore[FirebaseProject]` |
| 12 | `analytics_handler.go` | `sync.RWMutex` + map of dashboards | **P2** | `AnalyticsDashboard` resource + reconciler |
| 13 | `admin_handler.go` | `sync.RWMutex` + map | **P2** | split: promote resources to `ResourceStore`, keep ops-only endpoints thin |
| 14 | `api_builder_handler.go` | `sync.RWMutex` + map; also stores GIS/CSV state | **P0** | `APIBuilderService` + `ResourceStore[CustomAPI/CSVUpload]`; reconciler publishes generated routes |
| 15 | `query_logger*.go` | in-memory (verify) | **P2** | route through `ResourceStore[QueryLogEntry]` or a streaming sink |
| 16 | `transformation_handler.go` | check for in-memory rules | **P2** | `TransformRule` resource |
| 17 | `notification_handler.go` | check | **P2** | `NotificationChannel` resource |
| 18 | `netintel_handler.go` | check | **P2** | `NetIntelScan` resource |
| 19 | `performance_handler.go` | check | **P2** | usually acceptable (telemetry read-model) |
| 20 | `job_handler.go` | should already use JobQueue | **verify** | — |
| 21 | `query_builder_handler.go` | check | **P2** | `SavedQuery` resource |
| 22 | `cli_auth_handler.go` | check | **P1** | — |
| 23 | `login_identifier.go` | check | **P1** | — |
| 24 | `encryption_lineage_audit_workflow_handlers.go` | large multiplex file | **split** | each domain → its own service |
| 25 | `quality_rls_cdc_versioning_handlers.go` | large multiplex file | **split** | each domain → its own service |

---

## Migration recipe (per handler)

1. **Declare the resource type** under `internal/resources/<domain>/` (or
   extend an existing `apiresource`-style schema).
2. **Implement `ResourceStore[T]`** against etcd using the primitives in
   `internal/distributedstate/etcd.go`. Informer/watch come for free.
3. **Write a `Service`** under `internal/services/<domain>_service.go` that
   owns business rules, cross-resource invariants, and validation that
   cannot be expressed as struct tags.
4. **Write a `Reconciler`** under `internal/reconciler/` (or adjacent) that
   reads desired state from etcd and applies it to the external runtime
   (SQL DB, Kafka, Firebase, etc.). Only the reconciler talks to the runtime.
5. **Refactor the handler** to depend on the service (or the store for pure
   CRUD). Handler methods must be ≤ ~30 lines: bind → validate → call →
   render.
6. **Switch registration** in `main.go` to the new handler. Keep the old
   one until all clients migrate, then delete the file and this row.
7. **Update this document** (move the row to ✅ Compliant, note the date).

## Checklist for new handlers (no exceptions)

- [ ] No `*gorm.DB`, `*mongo.Client`, `*firebase.App`, etc. in handler
      struct.
- [ ] No `sync.Mutex` / `sync.RWMutex` protecting domain state (only
      short-lived caches).
- [ ] Depends on a `Service` or `ResourceStore[T]` interface.
- [ ] Status / progress fields written only by a reconciler.
- [ ] Unit tests use a fake store, not a real DB.
- [ ] Registered with the existing auth + rate-limit middleware chain.

## Open questions

- Should `dynamic_query_handler` and `graphql_handler` be formalised as
  "passthrough handlers" with an explicit marker interface so audits don't
  keep flagging them?
- Do we want a single `GISDashboard` kind with a `type` discriminator, or
  separate kinds (`AgricultureDashboard`, `MedicalDashboard`, …)?

---

## P2 Resource-ification Status (2026-04-25)

The following 11 modules now have `resource.go` + `reconciler.go` files
implementing `reconciler.Reconciler` with proper TypeMeta/ObjectMeta/Spec/Status
and ObservedGeneration tracking. All 17 reconcilers are initialized with
`EtcdStore[T]` in `main.go`.

| Module | Resource Types | Reconcilers | EtcdStore | Controller Loop |
|---|---|---|---|---|
| bulk | `BulkOperationResource` | `BulkOperationReconciler` | ✅ | ⚠️ initialized, not looping |
| eventbus | `TopicResource`, `SubscriptionResource` | `TopicReconciler`, `SubscriptionReconciler` | ✅ | ⚠️ initialized, not looping |
| export | `ExportJobResource` | `ExportJobReconciler` | ✅ | ⚠️ initialized, not looping |
| streaming | `StreamResource` | `StreamReconciler` | ✅ | ⚠️ initialized, not looping |
| rbac | `RoleResource`, `RoleBindingResource` | `RoleReconciler`, `RoleBindingReconciler` | ✅ | ⚠️ initialized, not looping |
| versioning | `VersionPolicyResource` | `VersionPolicyReconciler` | ✅ | ⚠️ initialized, not looping |
| tracing | `TracingConfigResource` | `TracingConfigReconciler` | ✅ | ⚠️ initialized, not looping |
| lineage | `LineageNodeResource` | `LineageNodeReconciler` | ✅ | ⚠️ initialized, not looping |
| audit | `AuditPolicyResource` | `AuditPolicyReconciler` | ✅ | ⚠️ initialized, not looping |
| encryption | `EncryptionKeyResource`, `EncryptionPolicyResource` | `EncryptionKeyReconciler`, `EncryptionPolicyReconciler` | ✅ | ⚠️ initialized, not looping |
| conductor | `ProducerResource`, `ConsumerResource` | `ProducerReconciler`, `ConsumerReconciler` | ✅ | ⚠️ initialized, not looping |
| webhooks | `WebhookResource` | `WebhookReconciler` | ✅ | ⚠️ initialized, not looping |
| tenant | `TenantV1Resource` | `TenantReconciler` | ✅ | ⚠️ initialized, not looping |

**Next step:** Wire these to `GenericController[T]` in shadow mode per
[MIGRATION_PLAN.md](./MIGRATION_PLAN.md) Phase 1.

Additionally, audit (`/api/v1/audit/*`) and encryption (`/api/v1/encryption/*`)
routes are now mounted in `main.go` with proper auth middleware.

