# Module Enrichment Plan

> **Goal:** Bring all 56 actively-wired internal modules to the full 9-artifact enrichment standard.
>
> **Generated:** 2026-05-28 | **Branch:** `mirazv2-code-cleaningv2`

---

## Enrichment Standard (9 Artifacts)

Every production module should have:

| # | Artifact | Purpose | Reference |
|---|----------|---------|-----------|
| 1 | `models/resource.go` | K8s-style Resource with Spec/Status, TypeMeta/ObjectMeta | `internal/waitx/models/resource.go` |
| 2 | `metrics/` | Prometheus counters + MetricsCollector with Snapshot() | `internal/waitx/metrics/counters.go` |
| 3 | `audit/` | KV-persisted audit log with ConfigureKVPersistence | `internal/waitx/audit/logger.go` |
| 4 | `types.go` | Type aliases pointing to models/ (backward compat) | `internal/waitx/types.go` |
| 5 | `errors.go` | Sentinel errors + typed error structs | `internal/waitx/errors.go` |
| 6 | `dto.go` | API request/response DTOs (no gin.H) | `internal/waitx/dto.go` |
| 7 | `http.go` | HTTP handler with RegisterRoutes(group) | `internal/waitx/http.go` |
| 8 | `system.go` | NewSystem, Start, Stop, SetKVStore, Handler() | `internal/waitx/system.go` |
| 9 | `reconciler.go` | K8s-style Reconciler (Observe→Diff→Act→Update) | `internal/waitx/reconciler.go` |

---

## Current State Summary

| Tier | Criteria Met | Module Count | Description |
|------|:------------:|:------------:|-------------|
| **Tier 1** | 8-9/9 | **5** | Fully enriched (waitx, jobs, iam, storage, gatekeeper) |
| **Tier 2** | 6-7/9 | **6** | Near-complete (conductor, antivirus, audit, encryption, bulk, eventbus) |
| **Tier 3** | 5/9 | **13** | Good foundation (export, streaming, webhooks, tenant, rbac, versioning, tracing, lineage, notification, slo, contracts, apibanks, apiscanner) |
| **Tier 4** | 3-4/9 | **17** | Partial (cdc, etl, datasource, transform, netintel, alerting, governance, costing, catalog, schemaregistry, federation, mlpipeline, anonymization, streamanalytics, featurestore, quality, security) |
| **Tier 5** | 1-2/9 | **15** | Minimal/skeleton (scanner, health, graphql, query, database, performance, docs, deployment, heartbeat, serviceregistry, autopilot, trivy, migrations, ratelimit, stream) |
| **Skip** | 0/9 | **24** | Utility/infra, not enrichable (reviewflow, vectorplus, kubeplus, modes, admission, distributedstate, distributed, diff, mesh, bootstrapsecrets, keyring, evalbroker, drainer, periodic, planner, scheduler, snapshot, rpcpool, template, sqlfilter, status, blocking, serverboot, securitysiem) |

---

## Enrichment Phases

### Phase E1: Tier 4 → Tier 2 (17 modules)

These modules already have models + DTOs + handlers + reconcilers. Add **metrics/, audit/, types.go, errors.go, system.go**.

| Priority | Module | Missing | Effort |
|:--------:|--------|---------|:------:|
| P1 | **cdc** | metrics, audit, types, errors, system | 1h |
| P1 | **etl** | metrics, audit, types, errors, dto, system | 1.5h |
| P1 | **datasource** | metrics, audit, errors, system | 1h |
| P1 | **transform** | metrics, audit, types, errors, dto, system | 1.5h |
| P1 | **netintel** | metrics, audit, errors, system | 1h |
| P2 | **alerting** | metrics, audit, types, errors, system | 1h |
| P2 | **governance** | metrics, audit, errors, system | 1h |
| P2 | **costing** | metrics, audit, errors, system | 1h |
| P2 | **catalog** | metrics, audit, types, errors, system | 1h |
| P2 | **schemaregistry** | metrics, audit, types, errors, system | 1h |
| P2 | **federation** | metrics, audit, errors, system | 1h |
| P2 | **mlpipeline** | metrics, audit, types, errors, system | 1h |
| P2 | **anonymization** | metrics, audit, types, errors, system | 1h |
| P2 | **streamanalytics** | metrics, audit, types, errors, system | 1h |
| P2 | **featurestore** | metrics, audit, errors, system | 1h |
| P3 | **quality** | metrics, audit, types, errors, system | 1h |
| P3 | **security** | models, metrics, audit, types, errors, system | 1.5h |

**Total effort:** ~18 hours

---

### Phase E2: Tier 3 → Tier 2 (13 modules)

These have models + DTOs + reconcilers. Add **metrics/, audit/, types.go, errors.go, system.go** (same gaps as Tier 4 but fewer missing DTOs).

| Priority | Module | Missing | Effort |
|:--------:|--------|---------|:------:|
| P1 | **export** | metrics, audit, errors, system | 1h |
| P1 | **streaming** | metrics, audit, errors, system | 1h |
| P1 | **webhooks** | metrics, audit, errors, system | 1h |
| P1 | **tenant** | metrics, audit, errors, system | 1h |
| P1 | **rbac** | metrics, audit, types, errors, system | 1h |
| P1 | **versioning** | metrics, audit, errors, system | 1h |
| P1 | **tracing** | metrics, audit, errors, system | 1h |
| P1 | **lineage** | models, metrics, audit, types, errors, system | 1.5h |
| P2 | **notification** | metrics, audit, errors, system | 1h |
| P2 | **slo** | metrics, audit, errors, system | 1h |
| P2 | **contracts** | metrics, audit, types, errors, system | 1h |
| P2 | **apibanks** | metrics, audit, dto, errors, system | 1h |
| P2 | **apiscanner** | metrics, audit, types, dto, errors, system | 1.5h |

**Total effort:** ~14 hours

---

### Phase E3: Tier 2 → Tier 1 (6 modules)

These are near-complete. Add only the missing 1-3 artifacts.

| Module | Missing | Effort |
|--------|---------|:------:|
| **conductor** | audit, errors | 30min |
| **antivirus** | models | 30min |
| **audit** | metrics, types, errors, system | 1h |
| **encryption** | metrics, audit, types, errors, system | 1h |
| **bulk** | metrics, audit, types, errors, system | 1h |
| **eventbus** | metrics, audit, types, errors, system | 1h |

**Total effort:** ~5 hours

---

### Phase E4: Tier 5 → Tier 3 (15 modules)

These are minimal — mostly single-file handlers or utilities. Add **models/resource.go + dto.go + system.go** at minimum.

| Module | Current State | Target |
|--------|---------------|--------|
| **scanner** | metrics.go (file) + system.go + types.go | Add models, audit, dto, errors |
| **health** | handler.go only | Add dto, system |
| **graphql** | handler + schema + resolver | Add dto, system |
| **query** | handler + endpoints | Add dto, system |
| **database** | handler + dto | Add models, system |
| **performance** | handler + analyzer | Add models, dto, system |
| **docs** | handler + openapi | Add models, dto, system |
| **deployment** | controller only | Add models, dto, http, system |
| **heartbeat** | tracker only | Add models, dto, http, system |
| **serviceregistry** | registry only | Add models, dto, http, system |
| **autopilot** | autopilot only | Add models, dto, http, system |
| **trivy** | types + engine | Add models, dto, http, system |
| **migrations** | single file | Add models, system |
| **ratelimit** | middleware only | Add models, dto, system |
| **stream** | broker + http | Add models, system |

**Total effort:** ~15 hours

---

## Execution Order

```
Week 1: Phase E3 (Tier 2 → Tier 1) — 6 modules, 5h
        Phase E1 (Tier 4 → Tier 2) P1 modules — 5 modules, 5.5h

Week 2: Phase E1 P2 modules — 9 modules, 9h
        Phase E2 (Tier 3 → Tier 2) P1 modules — 8 modules, 8.5h

Week 3: Phase E1 P3 + Phase E2 P2 — 8 modules, 8h
        Phase E4 (Tier 5 → Tier 3) — 15 modules, 15h

Total: ~52 hours across 3 weeks
```

---

## Batch Template

Each module enrichment follows this exact pattern:

```
internal/<module>/
├── models/
│   └── resource.go      # Resource struct + Spec/Status + GetGeneration()
├── metrics/
│   └── counters.go      # promauto counters + MetricsCollector + Snapshot()
├── audit/
│   └── logger.go        # KV-persisted audit log + ConfigureKVPersistence()
├── types.go             # Type aliases: type X = models.X
├── errors.go            # Sentinel errors + typed error structs
├── dto.go               # Request/Response DTOs (no gin.H)
├── http.go              # Handler + RegisterRoutes(group)
├── system.go            # NewSystem + Start + Stop + SetKVStore + Handler()
└── reconciler.go        # Reconcile(ctx, Resource) ReconcileResult
```

---

## Metrics Naming Convention

All per-module Prometheus metrics use `axiom_<module>` namespace:

| Module | Namespace |
|--------|-----------|
| cdc | `axiom_cdc` |
| etl | `axiom_etl` |
| datasource | `axiom_datasource` |
| transform | `axiom_transform` |
| netintel | `axiom_netintel` |
| alerting | `axiom_alerting` |
| governance | `axiom_governance` |
| costing | `axiom_costing` |
| catalog | `axiom_catalog` |
| schemaregistry | `axiom_schemaregistry` |
| federation | `axiom_federation` |
| mlpipeline | `axiom_mlpipeline` |
| anonymization | `axiom_anonymization` |
| streamanalytics | `axiom_streamanalytics` |
| featurestore | `axiom_featurestore` |
| ... | `axiom_<module>` |

---

## Audit Log KV Key Convention

All per-module audit logs use `module:audit:log` pattern:

| Module | KV Key |
|--------|--------|
| cdc | `cdc:audit:log` |
| etl | `etl:audit:log` |
| datasource | `datasource:audit:log` |
| ... | `<module>:audit:log` |

---

## Wiring Convention

Each enriched module gets wired in `main.go` following this pattern:

```go
// ModuleName — brief description
modSystem := mod.NewSystem()
modSystem.SetKVStore(backendMgr.KV())
_ = modSystem.Start(ctx)
modSystem.Handler().RegisterRoutes(router.Group("/api/v1", authMiddleware))
log.Println("✅ ModuleName started")
```

For modules with reconcilers:

```go
// ModuleName reconciler
modStore := platformstore.NewStore[*mod.Resource](backendMgr, "module-name", func() *mod.Resource { return &mod.Resource{} })
modReconciler := reconcilerpkg.NewInstrumented("module-name",
    mod.NewReconciler(), reconcilerMetrics)
reconcilerMetrics.Register("module-name")
go genericctrl.NewGenericController("module-name", modStore, modReconciler, 1, shadowMode, reconcilerMetrics).Start(ctx)
log.Println("  ✅ ModuleName controller started")
```

---

## Tracking

| Phase | Modules | Status | Date |
|-------|:-------:|:------:|------|
| E1 | 17 (Tier 4 → Tier 2) | ⬜ TODO | — |
| E2 | 13 (Tier 3 → Tier 2) | ⬜ TODO | — |
| E3 | 6 (Tier 2 → Tier 1) | ⬜ TODO | — |
| E4 | 15 (Tier 5 → Tier 3) | ⬜ TODO | — |
| **Total** | **51 modules** | — | ~52h |

---

## Excluded from Enrichment

These 24 modules are utility/infrastructure code with no API surface, no reconciler, and no user-facing routes. They do not need enrichment:

- **reviewflow** — approval workflow algorithm (used inline)
- **vectorplus** — vector search algorithm (used inline)
- **kubeplus/** — K8s extensions (admission, crd, scheduler) — used inline
- **netintel/modes** — network intel algorithm modes
- **admission** — admission webhook logic (used inline)
- **distributedstate** — distributed state primitives
- **distributed** — distributed systems utilities
- **diff** — diff engine
- **mesh** — service mesh primitives
- **bootstrapsecrets** — bootstrap secret management
- **keyring** — key management
- **evalbroker** — evaluation broker
- **drainer** — drain logic
- **periodic** — periodic task runner
- **planner** — planning algorithm
- **scheduler** — scheduling algorithm
- **snapshot** — snapshot logic
- **rpcpool** — RPC connection pool
- **template** — template engine
- **sqlfilter** — SQL injection filter
- **status** — status utilities
- **blocking** — blocking notification
- **serverboot** — server bootstrapping
- **securitysiem** — SIEM types only (no runtime)

---

*Last updated: 2026-05-28 (UTC+6)*
