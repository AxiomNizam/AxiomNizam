# AxiomNizam

**Cloud-Native Data Platform Engine with Kubernetes-Style Control Plane**

| | |
|---|---|
| **Version** | 1.0.0 |
| **Language** | Go 1.25 |
| **Framework** | Gin HTTP + GORM ORM |
| **Architecture** | Kubernetes-style declarative control plane |
| **Go Files** | 323 files · ~82,400 lines of code |
| **Internal Packages** | 51 packages |
| **REST Endpoints** | 200+ |
| **Databases Supported** | 9 (MySQL, MariaDB, PostgreSQL, Percona, Oracle, MongoDB, Redis/Valkey, Elasticsearch, etcd) |

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [Core Components](#core-components)
  - [Control Plane](#control-plane)
  - [API Layer](#api-layer)
  - [Data Layer](#data-layer)
  - [Policy and Security](#policy-and-security)
  - [Data Integration](#data-integration)
  - [Observability](#observability)
  - [Distributed Systems](#distributed-systems)
- [CLI Tool — axiomnizamctl](#cli-tool--axiomnizamctl)
- [Frontend Dashboards](#frontend-dashboards)
- [API Reference](#api-reference)
- [Configuration](#configuration)
- [Docker Deployment](#docker-deployment)
- [Examples](#examples)
- [Dependencies](#dependencies)

---

## Overview

AxiomNizam is an enterprise-grade data platform engine built on Kubernetes-style patterns. It provides:

- **Declarative resource management** — Define desired state via YAML; controllers reconcile to reality
- **Multi-database operations** — Unified CRUD and dynamic queries across 9 database backends
- **ETL/CDC pipelines** — Extract-Transform-Load and Change Data Capture orchestration
- **Policy engine** — RBAC, admission control, data governance, rate limiting, quota enforcement
- **Event-driven architecture** — Event bus with pub/sub, topics, dead-letter queues
- **Real-time streaming** — WebSocket-based live data feeds
- **Field-level encryption** — AES-256-GCM with key rotation and audit trails
- **Network intelligence** — Anomaly detection, traffic forecasting, topology mapping
- **GIS capabilities** — Geographic layers, regions, markers, specialized dashboards
- **kubectl-like CLI** — `axiomnizamctl` with YAML apply, contexts, shell completion
- **Multi-tenancy** — Tenant isolation, quotas, member management
- **Full observability** — Distributed tracing, audit logging, compliance reporting, metrics

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                        AxiomNizam Platform                           │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │  Frontend (port 7000)                                          │  │
│  │  Dashboards: Main · Admin · GIS · Analytics · CDC/ETL · NetIntel│ │
│  └────────────────────────┬───────────────────────────────────────┘  │
│                           │ HTTP                                     │
│  ┌────────────────────────▼───────────────────────────────────────┐  │
│  │  API Server (port 8000)         CLI: axiomnizamctl             │  │
│  │  ┌──────────────────────────────────────────────────────────┐  │  │
│  │  │  Gin Router · CORS · Rate Limiting · JWT Auth Middleware │  │  │
│  │  └──────────────────────────────────────────────────────────┘  │  │
│  │                                                                │  │
│  │  ┌─────────────┐ ┌──────────────┐ ┌────────────────────────┐  │  │
│  │  │ 29 Handlers │ │ Resource API │ │ Control Plane          │  │  │
│  │  │ (CRUD, GIS, │ │ (K8s-style   │ │ Controllers +          │  │  │
│  │  │  Analytics, │ │  namespaced)  │ │ Reconcilers +          │  │  │
│  │  │  CDC/ETL,   │ │              │ │ Work Queue +           │  │  │
│  │  │  NetIntel)  │ │              │ │ Event Bus              │  │  │
│  │  └──────┬──────┘ └──────┬───────┘ └────────────┬───────────┘  │  │
│  │         │               │                      │              │  │
│  │  ┌──────▼───────────────▼──────────────────────▼───────────┐  │  │
│  │  │  Services · Repositories · In-Memory Managers           │  │  │
│  │  └──────────────────────┬──────────────────────────────────┘  │  │
│  └─────────────────────────┼──────────────────────────────────────┘  │
│                            │                                         │
│  ┌─────────────────────────▼──────────────────────────────────────┐  │
│  │  Data Layer                                                    │  │
│  │  PostgreSQL · MySQL · MariaDB · Percona · Oracle · MongoDB    │  │
│  │  Valkey/Redis (cache) · Elasticsearch (search) · etcd (state) │  │
│  └────────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │  Auth: Keycloak (port 8080) — OIDC / JWT / RBAC              │  │
│  └────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.23+ (for local development)
- 8 GB RAM minimum
- Ports available: 7000, 8000, 8080

### Option 1: Docker Compose (recommended)

```bash
cd AxiomNizam
docker-compose up -d
```

This starts: API Server (`:8000`), Frontend (`:7000`), Keycloak (`:8080`), PostgreSQL, MongoDB, Valkey, etcd.

### Option 2: Local development

```bash
# Backend
go mod download
go run main.go

# Frontend (separate terminal)
cd frontend
go mod download
go run main.go
```

### Build CLI tool

```bash
go build -o axiomnizamctl ./cmd/axiomnizamctl/
```

### Get an auth token

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/realms/axiomnizam/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=axiomnizam-backend&client_secret=YOUR_SECRET&grant_type=client_credentials" \
  | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
```

### Test

```bash
curl http://localhost:8000/health
curl http://localhost:8000/api/postgres/users -H "Authorization: Bearer $TOKEN"
```

---

## Project Structure

```
AxiomNizam/
├── main.go                          # Root API server entry point (Gin router, all routes)
├── go.mod                           # Go module definition
├── docker-compose.yml               # Full stack: API + frontend + Keycloak + 5 databases
├── Dockerfile                       # Multi-stage build for API server + CLI
├── Dockerfile.axiomnizamctl         # Standalone CLI container
├── init-postgres.sql                # Keycloak database initialization
│
├── cmd/
│   ├── axiomnizam-server/main.go    # K8s-style server with controllers, informers, event bus
│   └── axiomnizamctl/               # CLI tool (13+ source files)
│       ├── main.go                  # Entry point
│       ├── root.go                  # Root cobra command, global flags
│       ├── COMMAND_TREE.go          # Full command documentation
│       ├── commands.go              # Status, health commands
│       ├── api.go                   # API resource CRUD
│       ├── apply.go                 # YAML apply engine
│       ├── auth.go                  # login/logout/current-user
│       ├── config.go                # kubeconfig-style context management
│       ├── policy_workflow.go       # Policy & workflow commands
│       ├── datasource_job.go        # DataSource & job commands
│       ├── mesh.go                  # Data mesh commands
│       ├── apibank.go               # API bank commands
│       ├── diff.go                  # Resource diff engine
│       ├── events.go                # Event listing
│       ├── integration.go           # Integration platform commands
│       ├── cli_manager.go           # Resource manager for CLI
│       └── completion.go            # Shell auto-completion
│
├── frontend/
│   ├── main.go                      # Gin server on port 7000
│   └── templates/                   # HTML/JS/CSS dashboards
│       ├── dashboard.html/js        # Main dashboard
│       ├── admin-dashboard.html/js  # Admin panel
│       ├── gis-dashboard.*          # Geographic Information System (Leaflet.js)
│       ├── analytics-dashboard.*    # Analytics with Chart.js
│       ├── cdc-etl-dashboard.*      # CDC/ETL pipeline monitor
│       ├── netintel-dashboard.*     # Network intelligence
│       ├── system-manager.html/js   # System management
│       └── auth.js                  # JWT token handling
│
├── internal/                        # 51 packages, 300+ Go files
│   │
│   │── # ── Control Plane ──
│   ├── apiserver/                   # Resource store with watchers (K8s kube-apiserver)
│   ├── resources/                   # CRD-like resource definitions + API versions
│   ├── controllers/                 # Reconciliation controllers (K8s operators)
│   ├── reconciler/                  # Core reconciliation engine (observe → diff → act)
│   ├── workqueue/                   # Rate-limited work queue with priority levels
│   ├── runtime/                     # Controller manager with leader election
│   │
│   │── # ── Security ──
│   ├── auth/                        # JWT validation, Keycloak OIDC, middleware
│   ├── rbac/                        # Role-based access control
│   ├── security/                    # Row-level security (RLS)
│   ├── encryption/                  # AES-256-GCM field encryption, key rotation
│   ├── policies/                    # Policy engine (CEL, Rego, DSL), admission, governance
│   ├── ratelimit/                   # Rate limiting middleware, quota management
│   │
│   │── # ── Data ──
│   ├── database/                    # Multi-database connection manager (9 backends)
│   ├── models/                      # Domain models (15 model files)
│   ├── repositories/                # Data access layer (15 repository files)
│   ├── migrations/                  # GORM auto-migrations with indexes
│   ├── services/                    # Business logic (auth, user, cached variants)
│   │
│   │── # ── API ──
│   ├── handlers/                    # 29 HTTP handler files
│   ├── client/                      # API client library for CLI
│   ├── config/                      # Configuration from environment variables
│   ├── graphql/                     # Dynamic GraphQL schema from database
│   ├── docs/                        # OpenAPI spec generator
│   │
│   │── # ── Data Integration ──
│   ├── etl/                         # ETL pipeline engine (10 step types)
│   ├── cdc/                         # Change data capture with webhook subscriptions
│   ├── streaming/                   # WebSocket real-time streaming
│   ├── workflows/                   # Workflow orchestration engine
│   ├── jobs/                        # Background job queue (18 files)
│   │
│   │── # ── Observability ──
│   ├── events/                      # Domain event system (20+ event types)
│   ├── eventbus/                    # Pub/sub event broker with topics
│   ├── cache/                       # Multi-backend cache (Redis + in-memory)
│   ├── status/                      # Resource status tracking
│   ├── diff/                        # Change detection engine
│   ├── audit/                       # Audit logging, compliance reporting
│   ├── tracing/                     # Distributed tracing (OpenTelemetry-style)
│   ├── metrics/                     # Prometheus-style metrics
│   ├── performance/                 # Query performance analyzer
│   │
│   │── # ── Distributed ──
│   ├── distributed/                 # etcd cluster coordination
│   ├── distributedstate/            # Distributed state (CAS, leases, locks)
│   ├── tenant/                      # Multi-tenancy (isolation, quotas, tiers)
│   │
│   │── # ── Governance ──
│   ├── lineage/                     # Data lineage tracking, impact analysis
│   ├── quality/                     # Data quality validation, anomaly detection
│   ├── mesh/                        # Data mesh (domains, products, SLAs)
│   ├── apibanks/                    # API collection management
│   ├── netintel/                    # Network intelligence (predictions, anomalies)
│   │
│   │── # ── Utilities ──
│   ├── export/                      # Data export (CSV, JSON, Parquet, Excel)
│   ├── bulk/                        # Bulk CRUD operations with progress
│   ├── versioning/                  # Resource version history, snapshots, rollback
│   ├── webhooks/                    # Webhook delivery with retry
│   ├── output/                      # Output formatting (JSON, YAML, table, wide)
│   ├── integration/                 # Multi-phase integration framework (31 files)
│   └── utils/                       # 44 utility files (retry, crypto, validation...)
│
└── examples/                        # YAML resources + Postman collections
    ├── api.yaml                     # API resource example
    ├── policy.yaml                  # RBAC policy example
    ├── workflow.yaml                # ETL workflow example
    ├── datasource.yaml              # PostgreSQL datasource example
    ├── kubeconfig-example.yaml      # CLI configuration
    └── *.json                       # 10 Postman collections
```

---

## Core Components

### Control Plane

The Kubernetes-style control plane implements the **reconciliation pattern**: users declare desired state, controllers continuously drive actual state to match.

| Component | Package | Purpose |
|---|---|---|
| API Server | `apiserver/` | In-memory resource store with watchers, namespace support, CRUD |
| Resources | `resources/` | CRD-like definitions: `WorkloadResource`, `PipelineResource`, `ScheduleResource`, `ExecutionResource` with `ObjectMeta`, `TypeMeta`, `ObjectStatus`, finalizers, labels, annotations |
| Controllers | `controllers/` | `WorkloadReconciler`, `PipelineReconciler`, `ScheduleReconciler` |
| Reconciler | `reconciler/` | Generic reconciliation engine with `Observer`, `Differ`, `Actor` interfaces |
| Work Queue | `workqueue/` | FIFO + priority queue with exponential backoff rate limiting |
| Runtime | `runtime/` | Controller manager, lifecycle orchestration, leader election |
| Status | `status/` | Phase tracking: Pending → Validating → Acting → Succeeded/Failed |
| Diff | `diff/` | Resource diff engine with policy/workflow impact analysis |

**Reconciliation loop:**
```
User applies YAML → API Server stores resource → Controller watches for changes
  → Controller reconciles (desired vs actual) → Updates status conditions
  → Requeues with backoff if needed
```

### API Layer

**29 HTTP handlers** registered in the Gin router, organized by domain:

| Handler | Endpoints | Description |
|---|---|---|
| Health | `/health`, `/status`, `/distributed` | Liveness, readiness, cluster mode |
| Auth | `/auth/login`, `/auth/refresh`, `/auth/validate` | JWT token lifecycle |
| User CRUD | `/api/{db}/users` | Per-database CRUD (7 databases) |
| Dynamic Query | `/api/{db}/query`, `/api/{db}/schema` | SQL execution, batch, introspection |
| Resources | `/api/v1/namespaces/{ns}/{kind}/{name}` | K8s-style namespaced resource API |
| GIS | `/api/v1/gis/*` | Layers, regions, markers, datasets, specialized dashboards |
| Analytics | `/api/v1/analytics/*` | Dashboards, widgets, CSV export |
| ETL | `/api/v1/etl/*` | Pipeline CRUD, runs, connectors, observability |
| CDC | `/api/v1/cdc/*` | Pipeline management, start/pause/stop, sources/sinks |
| Network Intel | `/api/v1/netintel/*` | Parsers, logs, topology, heatmaps, predictions, anomalies |
| DataSources | `/api/v1/datasources` | DataSource CRUD + connection testing |
| Jobs | `/api/v1/jobs` | Job submission, run, cancel, logs |
| Workflows | `/api/v1/workflows` | Workflow CRUD + run |
| Policies | `/api/v1/policies` | Policy CRUD |
| Admin | `/api/admin/*` | Database/table management, API metrics |
| Notifications | `/api/notifications/*` | Discord webhook notifications |
| Transform | `/api/transform/*` | Data transformation rules, batch, preview |

Additional **13 enterprise feature services** (71 endpoints):

| Service | Endpoints | Description |
|---|---|---|
| Audit | 4 | Immutable audit trail, compliance reports |
| Tenant | 9 | Multi-tenancy, members, quotas |
| Jobs (advanced) | 7 | Submit, cancel, retry, progress, logs |
| Streaming | 6 | WebSocket streams, subscriptions |
| Bulk | 7 | Batch operations with per-item tracking |
| Versioning | 6 | Version history, snapshots, diff, rollback |
| Webhooks | 7 | Webhook CRUD, test, delivery logs |
| Event Bus | 8 | Publish, topics, subscriptions, DLQ |
| Tracing | 8 | Traces, spans, service map, metrics |
| Export | 8 | Export jobs, templates, download URL |
| Lineage | 9 | Data lineage nodes/edges, impact analysis |
| Encryption | 9 | Key management, encrypt/decrypt, policies |
| RBAC | 14 | Roles, bindings, permissions, access requests |

### Data Layer

**Multi-database support** with unified connection management:

| Database | Driver | Use Case |
|---|---|---|
| MySQL | GORM | Relational storage |
| MariaDB | GORM | Relational storage |
| PostgreSQL | GORM | Primary relational database |
| Percona | GORM | MySQL-compatible |
| Oracle | GORM | Enterprise relational |
| MongoDB | Native driver | Document storage |
| Valkey/Redis | go-redis | Caching, job queue, query logs |
| Elasticsearch | elastic-go | Full-text search |
| etcd | etcd client v3 | Distributed state, leader election |

**ORM models** (15 files in `models/`): User, AuditLog, Tenant, Job, Stream, BulkOperation, ResourceVersion, Webhook, Event/Topic, Trace/Span, ExportJob, LineageNode/Edge, EncryptionKey, Role/Permission.

**Repository pattern** (15 files in `repositories/`): Each domain has a typed repository interface with Create, GetByID, Update, Delete, FindAll + specialized queries.

**Migrations**: Auto-migration with 30+ performance indexes and foreign key relationships.

### Policy and Security

| Component | Package | Description |
|---|---|---|
| Auth | `auth/` | JWT validation with Keycloak JWKS, token middleware |
| RBAC | `rbac/` | Roles (Admin, Manager, User, Guest), hierarchical permissions |
| Row-Level Security | `security/` | Per-table/per-user policy evaluation with caching |
| Field Encryption | `encryption/` | AES-256-GCM, key versioning, rotation events |
| Policy Engine | `policies/` | CEL/Rego/DSL languages, admission control, data governance, quotas |
| Rate Limiting | `ratelimit/` | Per-token rate limiting, quota management |

**Auth flow:**
```
Client → JWT in Authorization header → Keycloak JWKS validation
  → Claims extraction → RBAC role check → Rate limit check → Handler
```

**Roles:** Admin (full access) → Manager (no delete/role mgmt) → User (own data) → Guest (read-only)

### Data Integration

| Component | Package | Description |
|---|---|---|
| ETL Engine | `etl/` | Pipeline with 10 step types: extract, transform, load, filter, map, aggregate, join, validate, enrich, deduplicate |
| CDC | `cdc/` | Change event capture (INSERT, UPDATE, DELETE), webhook subscriptions |
| Streaming | `streaming/` | WebSocket real-time data with buffering, subscriptions |
| Workflows | `workflows/` | Multi-step workflow engine with triggers (policy, resource, schedule, manual) |
| Jobs | `jobs/` | Priority queues, dead-letter, cron scheduling, Redis persistence, fairness, observability |
| GraphQL | `graphql/` | Dynamic schema generation from database tables |

### Observability

| Component | Package | Description |
|---|---|---|
| Events | `events/` | 20+ domain event types, pub/sub bus, event history |
| Event Bus | `eventbus/` | Topic-based broker with correlation IDs, DLQ |
| Audit | `audit/` | Compliance rules (GDPR, HIPAA, SOC2, PCI-DSS), risk assessment |
| Tracing | `tracing/` | OpenTelemetry-style traces, spans, service dependency maps |
| Metrics | `metrics/` | Prometheus counters, gauges, histograms |
| Performance | `performance/` | Query duration, rows scanned, cache hits, index usage |
| Lineage | `lineage/` | Data flow graphs, transformation logs, impact analysis |
| Quality | `quality/` | Data validation rules, baseline anomaly detection |

### Distributed Systems

| Component | Package | Description |
|---|---|---|
| Distributed | `distributed/` | etcd cluster health, member listing, leader detection |
| Distributed State | `distributedstate/` | Compare-and-swap, leases/TTL, watch support, distributed locks |
| Cache | `cache/` | Redis + in-memory backends, TTL, informer change notifications |
| Mesh | `mesh/` | Data mesh pattern: domains, data products, SLAs, subscriptions |

---

## CLI Tool — axiomnizamctl

A kubectl-inspired CLI for managing the platform. Built with Cobra.

### Installation

```bash
go build -o axiomnizamctl ./cmd/axiomnizamctl/

# Or as a Docker container:
docker build -f Dockerfile.axiomnizamctl -t axiomnizamctl .
```

### Core commands

```bash
# Authentication
axiomnizamctl login [server-url]
axiomnizamctl logout
axiomnizamctl current-user

# YAML apply (kubectl-style)
axiomnizamctl api apply -f api.yaml
axiomnizamctl policy apply -f policy.yaml
axiomnizamctl workflow apply -f workflow.yaml
axiomnizamctl datasource apply -f datasource.yaml

# Resource CRUD
axiomnizamctl api list
axiomnizamctl api get users-api -o yaml
axiomnizamctl api describe users-api
axiomnizamctl api delete users-api
axiomnizamctl api diff -f api.yaml

# Workflow execution
axiomnizamctl workflow run daily-etl
axiomnizamctl workflow status daily-etl

# Jobs
axiomnizamctl job list
axiomnizamctl job get <id>
axiomnizamctl job logs <id>

# Monitoring
axiomnizamctl health check
axiomnizamctl status
axiomnizamctl alerts list
axiomnizamctl metrics collect
axiomnizamctl events list

# Data platform
axiomnizamctl apibank list
axiomnizamctl mesh list
axiomnizamctl lineage trace [resource]
axiomnizamctl quality analyze
axiomnizamctl compliance check

# Config (kubeconfig-style)
axiomnizamctl config view
axiomnizamctl config use-context production
axiomnizamctl config set-cluster staging --server=https://staging:8000
```

### Global flags

| Flag | Default | Description |
|---|---|---|
| `--namespace` | `default` | Namespace scope |
| `--output` | `table` | Output format: table, json, yaml, wide |
| `--kubeconfig` | `~/.axiomnizam/config` | Config file path |
| `--context` | (current) | Override current context |
| `--verbose` | false | Enable verbose logging |
| `--dry-run` | false | Preview without applying |

### Apply workflow

```
axiomnizamctl api apply -f api.yaml
  → CLI reads & validates YAML
  → POST /api/v1/namespaces/{ns}/{kind}s
  → Server stores resource, assigns generation number
  → Informer detects change, enqueues work
  → Controller reconciles desired → actual state
  → Status updated: phase=Active, conditions=[Ready, Synced]
  → CLI polls and displays result
```

---

## Frontend Dashboards

The frontend is a Go-based web server (Gin, port 7000) serving HTML/JS/CSS:

| Route | Dashboard | Technology |
|---|---|---|
| `/` | Public Dashboard | HTML/JS |
| `/admin` | Admin Dashboard | HTML/JS |
| `/system-manager` | System Manager | HTML/JS |
| `/gis` | GIS Dashboard | Leaflet.js |
| `/analytics` | Analytics | Chart.js |
| `/cdc-etl` | CDC/ETL Monitor | HTML/JS |
| `/netintel` | Network Intelligence | HTML/JS |

---

## API Reference

### Health (no auth)

```
GET  /health       → { "status": "alive" }
GET  /status       → { version, databases, connections }
GET  /distributed  → { etcd cluster status }
```

### Authentication

```
POST /auth/login              # Get JWT token
POST /auth/refresh            # Refresh token
GET  /auth/validate           # Validate token
GET  /auth/token-status       # Token info (requires auth)
GET  /auth/admin/tokens-status  # All tokens (admin only)
```

### Database CRUD

Databases: `mysql`, `mariadb`, `postgres`, `percona`, `oracle`, `mongodb`, `firebase`

```
GET    /api/{db}/users          # List (authenticated)
GET    /api/{db}/users/:id      # Get (authenticated)
POST   /api/{db}/users          # Create (admin)
PUT    /api/{db}/users/:id      # Update (admin)
DELETE /api/{db}/users/:id      # Delete (admin)
```

### Dynamic Queries

Databases: `mysql`, `mariadb`, `postgres`, `percona`, `oracle`

```
GET  /api/{db}/query           # Execute SELECT
POST /api/{db}/query           # Execute any SQL
POST /api/{db}/query/batch     # Batch operations
GET  /api/{db}/schema          # Table schema
GET  /api/{db}/logs            # Query logs
GET  /api/{db}/stats           # Query statistics
```

### GraphQL (auth required)

```
POST /api/graphql             # Execute GraphQL query
GET  /api/graphql/schema      # Schema availability/metadata
GET  /api/graphql/playground  # GraphQL playground endpoint
```

### Kubernetes-Style Resources

```
POST   /api/v1/namespaces/{ns}/{kind}              # Create
GET    /api/v1/namespaces/{ns}/{kind}               # List
GET    /api/v1/namespaces/{ns}/{kind}/{name}        # Get
PUT    /api/v1/namespaces/{ns}/{kind}/{name}        # Update
DELETE /api/v1/namespaces/{ns}/{kind}/{name}        # Delete
GET    /api/v1/namespaces/{ns}/{kind}/{name}/status # Get status
GET    /api/v1/namespaces/{ns}/{kind}/{name}/events # Get events
```

### Data Platform

```
# ETL Pipelines
GET/POST/PUT/DELETE  /api/v1/etl/pipelines[/:id]
POST                 /api/v1/etl/pipelines/:id/run
GET                  /api/v1/etl/{runs|connectors|observability}

# CDC Pipelines
GET/POST/PUT/DELETE  /api/v1/cdc/pipelines[/:id]
POST                 /api/v1/cdc/pipelines/:id/{start|pause|stop}
GET                  /api/v1/cdc/{sources|sinks|observability}

# GIS
CRUD                 /api/v1/gis/{layers|regions|markers|datasets}[/:id]
GET                  /api/v1/gis/summary
GET                  /api/v1/gis/dashboards[/:type[/summary]]

# Analytics
GET                  /api/v1/analytics/dashboards[/:id]
PUT                  /api/v1/analytics/dashboards/:id/{widgets/:id|layout}
GET                  /api/v1/analytics/widget-types

# Network Intelligence
CRUD                 /api/v1/netintel/parsers[/:id]
GET/POST             /api/v1/netintel/logs
GET                  /api/v1/netintel/{topology|heatmap|trends|predictions|anomalies|alerts|forecasts}

# Data Transformation
POST                 /api/transform/{apply|batch|preview}
CRUD                 /api/transform/rules[/:name]
GET/POST             /api/transform/rules/{export|import}

# DataSources
CRUD                 /api/v1/datasources[/:name]
POST                 /api/v1/datasources/:name/test

# Jobs
POST   /api/v1/jobs           # Create
GET    /api/v1/jobs            # List
GET    /api/v1/jobs/:id        # Get
POST   /api/v1/jobs/:id/run   # Run
POST   /api/v1/jobs/:id/cancel  # Cancel
GET    /api/v1/jobs/:id/logs  # Logs
DELETE /api/v1/jobs/:id       # Delete
```

### Admin & System Manager (RBAC)

```
POST /api/admin/database/create
GET  /api/admin/database/list
GET  /api/admin/database/servers
POST /api/admin/database/connect
POST /api/admin/table/create
GET  /api/admin/table/list
```

### Admin-only Metrics

```
GET /api/admin/metrics/{all|count|stats}
```

### Notifications

```
POST /api/notifications/send     # Send to Discord
POST /api/notifications/health   # Health notification
POST /api/notifications/status   # Status notification
GET  /api/notifications/status   # Service status
```

---

## Configuration

### Environment Variables

```env
# Server
PORT=8000
FRONTEND_PORT=7000

# Keycloak (OIDC)
KEYCLOAK_URL=http://keycloak:8080
KEYCLOAK_REALM=axiomnizam
KEYCLOAK_CLIENT_ID=axiomnizam-backend
KEYCLOAK_CLIENT_SECRET=<secret>

# Databases
MYSQL_DSN=root:password@tcp(mysql:3306)/axiomnizam
POSTGRES_DSN=host=postgres user=postgres password=postgres dbname=axiomnizam
MONGO_URI=mongodb://mongo:27017
REDIS_ADDR=redis:6379

# Rate Limiting
MAX_CALLS_PER_TOKEN=100
TOKEN_VALIDITY_MINUTES=60

# Notifications (optional)
DISCORD_WEBHOOK_URL=https://discordapp.com/api/webhooks/...
```

### CLI Config Files

Stored at `~/.axiomnizam/`:

| File | Purpose |
|---|---|
| `config` | Contexts, clusters, users (kubeconfig-style YAML) |
| `token` | JWT token (file permissions: 0600) |

---

## Docker Deployment

### docker-compose.yml services

| Service | Port | Purpose |
|---|---|---|
| **axiomnizam** | 8000 | API server + CLI |
| **axiomnizam-frontend** | 7000 | Web dashboards |
| **keycloak** | 8080 | OAuth2/OIDC auth |
| **postgres** | 5432 | Primary database + Keycloak backend |
| **mongodb** | 27017 | Document storage |
| **valkey** | 6379 | Cache + job queue |
| **etcd** | 2379 | Distributed coordination |

Optional services (commented in compose file): MySQL, MariaDB, Percona, Oracle, Elasticsearch, Firebase Emulator.

### Commands

```bash
docker-compose up -d                  # Start all
docker-compose ps                     # Check status
docker-compose logs -f axiomnizam     # View API logs
docker-compose exec axiomnizam axiomnizamctl version  # CLI in container
```

### Dockerfiles

| File | Description |
|---|---|
| `Dockerfile` | Multi-stage build: Go 1.25 builder → Debian slim runtime. Builds both `axiomnizam` server and `axiomnizamctl` CLI. Exposes port 8000. |
| `Dockerfile.axiomnizamctl` | Standalone CLI container on Alpine. |
| `Dockerfile.ctl-test` | Test runner container for CLI integration tests. |

---

## Examples

Example YAML resources and Postman collections are in `examples/`:

| File | Description |
|---|---|
| `api.yaml` | API resource (CRUD, rate limit, cache) |
| `policy.yaml` | RBAC policy (roles, time/IP conditions) |
| `workflow.yaml` | ETL workflow (cron schedule, 3 steps) |
| `datasource.yaml` | PostgreSQL connection config |
| `kubeconfig-example.yaml` | CLI context/cluster configuration |
| `COMPLETE_API_COLLECTION.json` | Full Postman collection |
| `POSTMAN_COLLECTION.json` | Core API testing |
| `CDC_ETL_POSTMAN.json` | CDC/ETL endpoint tests |
| `DATA_TRANSFORMATION_POSTMAN.json` | Transformation tests |
| `DYNAMIC_QUERIES_POSTMAN.json` | Dynamic query tests |
| `NETWORK_INTELLIGENCE_POSTMAN_COLLECTION.json` | NetIntel tests |
| `API_METRICS_POSTMAN.json` | API metrics tests |
| `QUERY_LOGGER_POSTMAN_COLLECTION.json` | Query logging tests |
| `UTILS_POSTMAN_COLLECTION.json` | Utility endpoint tests |

---

## Dependencies

Key Go modules (`go.mod`):

| Module | Version | Purpose |
|---|---|---|
| `gin-gonic/gin` | 1.9.1 | HTTP framework |
| `gorm.io/gorm` | 1.30.0 | ORM for SQL databases |
| `gorm.io/driver/postgres` | 1.6.0 | PostgreSQL driver |
| `gorm.io/driver/mysql` | 1.6.0 | MySQL/MariaDB/Percona driver |
| `go.mongodb.org/mongo-driver` | 1.12.1 | MongoDB native driver |
| `redis/go-redis/v9` | 9.0.5 | Redis/Valkey client |
| `elastic/go-elasticsearch/v8` | 8.19.1 | Elasticsearch client |
| `go.etcd.io/etcd/client/v3` | 3.5.9 | etcd client |
| `golang-jwt/jwt/v5` | 5.0.0 | JWT tokens |
| `spf13/cobra` | 1.10.2 | CLI framework |
| `gorilla/websocket` | 1.5.3 | WebSocket support |
| `robfig/cron/v3` | 3.0.1 | Cron scheduling |
| `prometheus/client_golang` | 1.11.1 | Prometheus metrics |
| `go.opentelemetry.io/otel` | 1.28.0 | OpenTelemetry tracing |
| `go.uber.org/zap` | 1.17.0 | Structured logging |
| `golang.org/x/crypto` | 0.31.0 | Cryptographic functions |
| `google/uuid` | 1.6.0 | UUID generation |
| `graphql-go/graphql` | 0.8.1 | GraphQL engine |

---

## License

See [LICENSE](LICENSE) for details.

---

## GUI API Builder, File-to-Dashboard, Dashboard↔GIS Converter & SafeGate File Scanner

### Overview

The Admin Dashboard (`/admin`) provides powerful GUI-based features that let administrators create APIs, ingest data files, convert between dashboard types, and scan files for security threats — all without writing code.

### 1. GUI API Builder (REST + GraphQL, separate builders)

Create, test, and manage custom APIs visually from the admin interface.

- REST API Builder remains dedicated to REST endpoint definitions.
- GraphQL API Builder is a separate UI section and stores `api_type=graphql` APIs.

**Features:**
- Create APIs with name, method (GET/POST/PUT/DELETE/PATCH), path, category, and description
- Select source database and source server for each API definition
- Set authentication requirements and rate limits per API
- **Configurable response caching** — enable per-API caching with custom TTL (1–86400 seconds, default 300s)
- Define mock JSON responses for rapid prototyping
- Add query parameters with type and required/optional flags
- Create GraphQL APIs with operation name and GraphQL query body (kept separate from REST builder)
- Test APIs directly from the GUI with one click — cached responses returned instantly
- Track hit counts and status (active/draft/archived)
- Filter APIs by category and status

**Backend Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/builder/summary` | Builder summary (supports `api_type=rest|graphql`) |
| GET | `/api/v1/builder/apis` | List custom APIs (filter by `api_type`, category, status) |
| POST | `/api/v1/builder/apis` | Create a custom API (`api_type=rest|graphql`, supports `source_database`, `source_server`, `cache_enabled`, `cache_ttl`) |
| GET | `/api/v1/builder/apis/:id` | Get API details |
| PUT | `/api/v1/builder/apis/:id` | Update an API (including cache settings) |
| DELETE | `/api/v1/builder/apis/:id` | Delete an API |
| POST | `/api/v1/builder/apis/:id/test` | Test API (returns cached or mock response) |

### 2. File Upload → Auto Analytics Dashboard

Upload CSV, JSON, or Excel (.xlsx) files and automatically generate a full analytics dashboard with appropriate widgets, charts, and tables.

**Features:**
- Drag-and-drop file upload zone supporting **CSV**, **JSON**, and **Excel (.xlsx/.xls)** formats
- JSON support: array of objects or object containing a data array
- Excel support: reads first sheet, header row + data rows
- Automatic column type detection: string, number, date, geo_lat, geo_lng, geo_name
- Sample data preview table
- Auto-generates dashboard with:
  - KPI widgets (average of numeric columns)
  - Bar charts (string vs. numeric aggregation)
  - Doughnut charts (string frequency distribution)
  - Line charts (date vs. numeric trends)
  - Full data table widget
- If geo data (lat/lng) is detected, also generate a GIS map dataset with markers
- **Dashboard deletion** — delete generated dashboards from the upload history
- Upload history with file type tracking and status

**Backend Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/builder/csv/upload` | Upload file (CSV, JSON, or Excel — multipart form) |
| GET | `/api/v1/builder/csv/uploads` | List all file uploads |
| GET | `/api/v1/builder/csv/uploads/:id` | Get upload details |
| DELETE | `/api/v1/builder/csv/uploads/:id` | Delete an upload |
| POST | `/api/v1/builder/csv/uploads/:id/generate-dashboard` | Generate analytics dashboard |
| POST | `/api/v1/builder/csv/uploads/:id/generate-gis` | Generate GIS dataset (requires geo columns) |
| DELETE | `/api/v1/builder/dashboards/:id` | Delete a generated dashboard |

### 3. Dashboard ↔ GIS Converter

Convert between analytics dashboards and GIS map views bidirectionally, with automatic field mapping and confidence scoring.

**Dashboard → GIS:**
- Analyzes dashboard widgets for geographic data (lat, lng, region, city columns)
- Calculates conversion confidence score (0-100%)
- Extracts markers from table widget rows
- Creates a new GIS dataset with markers placed at detected coordinates

**GIS → Dashboard:**
- Creates analytics widgets from GIS dataset markers
- Generates KPI cards, marker distribution charts, and data tables
- Preserves all original marker metadata

**Backend Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/builder/convert/analyze` | Analyze conversion feasibility and confidence |
| POST | `/api/v1/builder/convert/dashboard-to-gis` | Convert dashboard to GIS dataset |
| POST | `/api/v1/builder/convert/gis-to-dashboard` | Convert GIS dataset to dashboard |
| GET | `/api/v1/builder/conversions` | List all conversion history |

### 4. SafeGate File Scanner

Integrated security file scanner with a 6-stage detection pipeline. Scan any uploaded file for malware, macro exploits, XSS payloads, archive bombs, and more.

**Scanner Pipeline (6 stages):**

| Scanner | Detection |
|---------|-----------|
| **Metadata Scanner** | File size limits, empty files, null bytes in text files, double-extension spoofing |
| **MIME Type Scanner** | Magic byte detection, MIME type spoofing (claimed vs detected), executable signatures (PE/ELF/Mach-O) |
| **SVG XSS Scanner** | Script tags, event handlers, javascript: URIs, data: URIs, foreignObject, external xlink, base64 injection |
| **Macro Scanner** | PDF JavaScript/auto-actions/launch/embedded/encrypted, Office VBA macros/auto-exec/shell commands |
| **Archive Bomb Scanner** | Zip bomb detection (>100:1 ratio), nesting depth limits, path traversal, executables inside archives |
| **ClamAV Antivirus** | TCP INSTREAM protocol to ClamAV daemon for full virus/malware detection |

**Features:**
- Drag-and-drop scan zone in the admin interface
- SHA256 file fingerprinting
- Severity classification: Critical, High, Medium, Low, Info
- Scan history with results tracking
- Scanner health monitoring

**Backend Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/builder/scanner/scan` | Scan a file (multipart form upload) |
| GET | `/api/v1/builder/scanner/scans` | List all scan records |
| GET | `/api/v1/builder/scanner/health` | Scanner pipeline health status |

**Docker Compose:**
ClamAV runs as a dedicated container service on port 3310:
```yaml
clamav:
  image: clamav/clamav:latest
  ports:
    - "3310:3310"
```

**Environment Variables:**
```env
SAFEGATE_CLAMAV_ADDR=clamav:3310
SAFEGATE_MAX_FILE_SIZE=104857600
```

### Admin Interface Tabs

| Tab | Description |
|-----|-------------|
| **API Builder** | Summary cards, API list with filters, create/test/delete APIs, cache configuration |
| **GraphQL API Builder** | Separate GraphQL API creation/testing with operation name, query body, and same policy/cache/rate-limit controls |
| **File → Dashboard** | Drag-drop upload (CSV/JSON/Excel), column analysis, generate dashboard or GIS, delete dashboards |
| **Dashboard ↔ GIS** | Select source, analyze confidence, convert with field mapping |
| **File Scanner** | SafeGate 6-stage security scan, drag-drop scan zone, scan history, scanner health |
| **API Testing** | Original API testing with method filters |
| **GraphQL Studio** | Execute GraphQL queries with variables and view JSON responses in-browser |
| **Control Plane** | kubectl-style resource apply/list/get/status/events + DataSource and Job operations |
| **Logs** | Real-time activity log viewer |
| **Settings** | System configuration |

### System Manager Interface Tabs

| Tab | Description |
|-----|-------------|
| **Overview** | Live system health and synthetic performance indicators |
| **Databases** | Server-aware database create/list and DB server connection workflow |
| **Users** | User CRUD with role assignment |
| **Monitoring** | Metrics and performance overview |
| **Operations** | Operational maintenance actions and operation log |
| **GraphQL Studio** | Execute GraphQL queries as system-manager/admin |
| **Control Plane** | Resource/DataSource/Job/workflow operational commands with RBAC enforcement |

System-manager (sysadmin) role is allowed to use all admin UI capabilities and admin routes exposed by the dashboard, including both REST and GraphQL API Builder workflows.
