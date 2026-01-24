# AxiomNizam - Architecture & Design

**Status**: ✅ Production Ready | **Version**: 1.0 | **Last Updated**: January 24, 2026

---

## 📐 System Architecture Overview

AxiomNizam is a **Kubernetes-style data control plane** with a **kubectl-like CLI**. It manages APIs, policies, workflows, and data sources using declarative YAML and event-driven architecture.

```
┌─────────────────────────────────────────┐
│         CLI (axiomnizamctl)             │
│    kubectl-style command interface      │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│       REST API (Backend Server)         │
│  /api/v1/{resources} endpoints          │
└──────────────────┬──────────────────────┘
                   │
    ┌──────────────┼──────────────┐
    │              │              │
┌───▼────┐  ┌──────▼──────┐  ┌──▼──────┐
│ Auth & │  │  Database   │  │  Cache  │
│  RBAC  │  │  Layer      │  │ (Redis) │
└────────┘  └─────────────┘  └─────────┘
    │              │              │
    └──────────────┼──────────────┘
                   │
        ┌──────────▼──────────┐
        │  PostgreSQL Stores  │
        │  - APIs             │
        │  - Policies         │
        │  - Workflows        │
        │  - DataSources      │
        │  - Jobs             │
        │  - Events           │
        └─────────────────────┘
```

---

## 🏗️ Project Structure

### Binary Organization (Go Best Practices)

```
cmd/
├── axiomnizam/              ← Backend server binary
│   └── main.go
└── axiomnizamctl/           ← CLI binary (kubectl-style)
    ├── main.go              ← Entry point
    ├── root.go              ← Root command & globals
    ├── api.go               ← API commands (create, list, get, apply, delete, describe, diff)
    ├── policy_workflow.go   ← Policy & Workflow commands
    ├── datasource_job.go    ← DataSource & Job commands
    ├── config.go            ← Config commands
    ├── completion.go        ← Shell completion generation
    └── helpers.go           ← Utility functions

internal/
├── apiserver/              ← HTTP server setup
│   └── server.go
├── auth/                   ← Authentication & authorization
│   ├── auth.go
│   ├── middleware.go
│   └── rate_limit*.go
├── cache/                  ← Caching layer (Redis)
│   ├── cache.go
│   ├── memory.go
│   ├── redis.go
│   └── middleware.go
├── config/                 ← Configuration management
│   └── config.go
├── controllers/            ← Request handlers
│   └── controller.go
├── database/               ← Database connections
│   └── connections.go
├── events/                 ← Event system
│   └── event.go
├── handlers/               ← API handlers
│   ├── admin_handler.go
│   ├── api_metrics.go
│   ├── auth_handler.go
│   ├── dynamic_query_handler.go
│   ├── notification_handler.go
│   ├── query_builder_handler.go
│   ├── query_logger_handlers.go
│   ├── transformation_handler.go
│   └── ...
├── jobs/                   ← Background job system
│   ├── job.go
│   ├── manager.go
│   ├── queue.go
│   ├── redis_queue.go
│   └── ...
├── models/                 ← Data models
│   └── models.go
├── output/                 ← CLI output formatting & error handling
│   ├── errors.go           ← Error codes & messages
│   └── formatter.go        ← Output formatters
├── policies/               ← RBAC policies
│   └── rbac.go
├── services/               ← Business logic
│   ├── auth_service.go
│   ├── user_service.go
│   ├── base.go
│   └── *_cached.go
├── utils/                  ← Utility functions
│   ├── database.go
│   ├── encryption.go
│   ├── input_validation.go
│   ├── query_builder.go
│   ├── sql_injection_protection.go
│   ├── transformer.go
│   └── ...
├── workqueue/              ← Work queue implementation
│   └── queue.go
└── client/                 ← CLI SDK client
    ├── client.go
    └── config_manager.go

main.go                     ← Entry point for backend server
go.mod                      ← Go module definition
docker-compose.yml          ← Docker Compose config
Dockerfile                  ← Backend container
README.md                   ← Main documentation
QUICKSTART.md              ← Getting started guide
```

---

## 🔧 Core Components

### 1. CLI (axiomnizamctl)

**kubectl-style command-line interface** for managing AxiomNizam resources.

```bash
# Login
axiomnizamctl login

# Manage APIs
axiomnizamctl api create
axiomnizamctl api list
axiomnizamctl api get [name]
axiomnizamctl api apply -f api.yaml
axiomnizamctl api delete [name]
axiomnizamctl api describe [name]
axiomnizamctl api diff -f api.yaml

# Manage Policies
axiomnizamctl policy apply -f policy.yaml
axiomnizamctl policy list
axiomnizamctl policy describe [name]
axiomnizamctl policy diff -f policy.yaml
axiomnizamctl policy delete [name]

# Manage Workflows
axiomnizamctl workflow apply -f workflow.yaml
axiomnizamctl workflow list
axiomnizamctl workflow run [name]
axiomnizamctl workflow status [name]
axiomnizamctl workflow describe [name]
axiomnizamctl workflow diff -f workflow.yaml

# Manage DataSources
axiomnizamctl datasource create
axiomnizamctl datasource list
axiomnizamctl datasource test [name]
axiomnizamctl datasource apply -f datasource.yaml
axiomnizamctl datasource describe [name]
axiomnizamctl datasource diff -f datasource.yaml
axiomnizamctl datasource delete [name]

# Manage Jobs
axiomnizamctl job list
axiomnizamctl job get [job-id]
axiomnizamctl job status [job-id]
axiomnizamctl job describe [job-id]
axiomnizamctl job logs [job-id]
axiomnizamctl job cancel [job-id]

# Shell completion
axiomnizamctl completion bash | tee /etc/bash_completion.d/axiomnizamctl
```

**Features**:
- ✅ Login/Logout with token management
- ✅ Declarative YAML support (apply pattern)
- ✅ Describe command (shows resource + recent events)
- ✅ Diff command (previews changes before apply)
- ✅ Status command (monitors workflow/job progress)
- ✅ Shell auto-completion (bash/zsh/fish/powershell)
- ✅ Multiple output formats (table, JSON, YAML)
- ✅ Error codes with helpful suggestions

### 2. Backend API Server

**RESTful API** serving Kubernetes-style resources.

**Main Endpoints**:
```
POST   /api/v1/apis                    → Create API
GET    /api/v1/apis                    → List APIs
GET    /api/v1/apis/{name}             → Get API
PUT    /api/v1/apis/{name}             → Update API
DELETE /api/v1/apis/{name}             → Delete API
GET    /api/v1/apis/{name}/events      → Get API events

POST   /api/v1/policies                → Create Policy
GET    /api/v1/policies                → List Policies
GET    /api/v1/policies/{name}         → Get Policy
DELETE /api/v1/policies/{name}         → Delete Policy

POST   /api/v1/workflows               → Create Workflow
GET    /api/v1/workflows               → List Workflows
POST   /api/v1/workflows/{name}/run    → Run Workflow
GET    /api/v1/workflows/{name}/status → Get Status

POST   /api/v1/datasources             → Create DataSource
GET    /api/v1/datasources             → List DataSources
POST   /api/v1/datasources/{name}/test → Test Connection

GET    /api/v1/jobs                    → List Jobs
GET    /api/v1/jobs/{id}               → Get Job
GET    /api/v1/jobs/{id}/logs          → Stream Logs
POST   /api/v1/jobs/{id}/cancel        → Cancel Job
```

### 3. Authentication & Authorization (RBAC)

**Keycloak + Role-Based Access Control**

- Token-based auth with Bearer tokens
- Fine-grained RBAC policies
- Rate limiting per user
- Middleware for request validation

### 4. Database Layer

**Multi-database support**:
- PostgreSQL (primary store)
- MySQL (external data)
- MongoDB (document storage)
- Oracle (enterprise)
- Firebase (cloud)

**Connection pooling** with configurable limits for each database type.

### 5. Caching System

**Two-tier caching**:
- **Memory cache** (fast, single-instance)
- **Redis cache** (distributed, multi-instance)

Automatic invalidation on resource updates.

### 6. Background Job System

**Distributed job queue** with:
- PostgreSQL or Redis backing store
- Job scheduling with cron expressions
- Workflow orchestration
- Email notifications
- Dead-letter queue for failures
- Rate limiting
- Fairness enforcement

**Job Lifecycle**:
```
Pending → Running → Success/Failed
  ↓
Dead Letter Queue (for failures)
  ↓
Retry with backoff
```

### 7. Event System

**Event-driven architecture**:
- All resource changes generate events
- Events stored in PostgreSQL
- Event log queryable via API
- Real-time notifications

---

## 🔐 Security Architecture

### Authentication Flow

```
1. User runs: axiomnizamctl login
2. CLI prompts for credentials
3. Backend validates (currently mock, integrates with Keycloak)
4. Token generated and stored in ~/.axiomnizam/config
5. All subsequent requests include token
```

### Authorization (RBAC)

```
User Request
    ↓
Auth Middleware (verify token)
    ↓
RBAC Policy Check (can user do this action?)
    ↓
Rate Limit Check (user quota)
    ↓
Handler executes
```

### Security Features

- ✅ SQL injection protection
- ✅ Input validation
- ✅ CORS protection
- ✅ Rate limiting
- ✅ Token expiration
- ✅ Encrypted passwords
- ✅ Audit logging (events)

---

## 📊 Error Handling

**Standard Error Codes** (`internal/output/errors.go`):

```go
ErrNotFound         // Resource not found
ErrUnauthorized     // Auth failed
ErrForbidden        // No permission
ErrInvalidInput     // Validation error
ErrConflict         // Resource conflict
ErrInvalidYAML      // YAML parsing error
ErrServerError      // 5xx error
ErrTimeout          // Request timeout
ErrUnavailable      // Service unavailable
```

**Error Response Format**:
```json
{
  "code": "NOT_FOUND",
  "message": "API 'users-api' not found",
  "details": "Check resource name with 'list' command",
  "timestamp": "2026-01-24T10:30:00Z"
}
```

---

## 🎯 Design Patterns Used

### 1. Kubernetes-Style Pattern
- Declarative configuration (YAML)
- Kubectl-like CLI
- Resource reconciliation
- Event-driven updates

### 2. Service Layer Pattern
- Business logic in `services/` package
- Data access in `repositories/` package
- Clear separation of concerns

### 3. Handler Pattern
- HTTP handlers in `handlers/` package
- Request parsing and validation
- Response formatting

### 4. Middleware Pattern
- Authentication middleware
- Rate limiting middleware
- Caching middleware
- Logging middleware

### 5. Cache-Aside Pattern
- Check cache first
- Fall through to DB
- Update cache on miss
- Invalidate on write

### 6. Repository Pattern
- Abstract data access
- Support multiple backends
- Easy to test (mock repositories)

---

## 🔄 Request Flow

```
CLI Request (axiomnizamctl api create)
    ↓
CLI HTTP Client (internal/client)
    ↓
Backend API Server
    ↓
Auth Middleware → Validate token & RBAC
    ↓
Handler (handlers/api_handler.go)
    ↓
Service Layer (services/api_service.go)
    ↓
Repository (repositories/api_repository.go)
    ↓
Cache Layer (check/update)
    ↓
Database
    ↓
Event System (log the change)
    ↓
Response → CLI → Format & Display
```

---

## 🚀 Performance Optimizations

### 1. Connection Pooling
- Database connections reused
- Configurable pool sizes
- Auto-cleanup of stale connections

### 2. Caching
- Frequently accessed resources cached
- Cache invalidation on updates
- Multi-tier (memory + Redis)

### 3. Query Optimization
- Indexed database queries
- Pagination for large result sets
- Select only needed fields

### 4. Rate Limiting
- User quota enforcement
- Sliding window algorithm
- Prevents abuse

### 5. Async Processing
- Long-running operations as background jobs
- Non-blocking API responses
- Job progress tracking

---

## 🧪 Testing Strategy

### Unit Tests
- Service layer tests
- Handler tests
- Utility function tests

### Integration Tests
- Full request/response cycle
- Database interactions
- Cache behavior

### E2E Tests
- CLI commands
- Full workflows
- Multi-step processes

---

## 📈 Monitoring & Observability

### Metrics Tracked
- Request latency
- Error rates
- Cache hit ratio
- Database connection count
- Job success rate
- Job duration
- API throughput

### Logging
- Structured logging (JSON)
- Log levels (DEBUG, INFO, WARN, ERROR)
- Request tracing
- Event audit logs

### Healthchecks
- `/health` endpoint
- Database connectivity
- Cache availability
- Service dependencies

---

## 🔧 Configuration

**Config file location**: `~/.axiomnizam/config`

```yaml
server:
  address: localhost:8000
  env: development

auth:
  token: <user-token>
  expires_at: <timestamp>

database:
  postgres:
    host: localhost
    port: 5432
    database: axiomnizam
  mysql:
    host: localhost
    port: 3306

cache:
  type: redis           # memory or redis
  redis_url: localhost:6379

output_format: table    # table, json, yaml
```

---

## 🛠️ Go Module Structure

**Module**: `axiom-nizam`

**Key Dependencies**:
```
github.com/spf13/cobra          → CLI framework
gopkg.in/yaml.v3                → YAML parsing
github.com/redis/go-redis       → Redis client
github.com/lib/pq               → PostgreSQL driver
github.com/go-sql-driver/mysql  → MySQL driver
github.com/mongodb/mongo-go-driver → MongoDB client
```

---

## 🔄 Deployment Architecture

### Single Node (Development)
```
Local Machine
├── Backend Server (port 8000)
├── PostgreSQL (port 5432)
├── Redis (port 6379)
└── CLI (local binary)
```

### Docker Compose (Testing)
```
docker-compose up
├── Backend Container
├── Keycloak Container
├── PostgreSQL Container
├── Redis Container
├── MySQL Container
└── Frontend Container
```

### Kubernetes (Production)
```
Deployment: axiomnizam-backend
├── Replicas: 3+
├── CPU: 500m per pod
├── Memory: 512Mi per pod
│
Service: axiomnizam-api (LoadBalancer)
├── Port: 8000
│
StatefulSet: axiomnizam-postgres
├── Persistent Volume: 50Gi
│
StatefulSet: axiomnizam-redis
├── Persistent Volume: 10Gi
```

---

## 🏆 Code Quality Standards

- ✅ Go conventions (cmd/, internal/)
- ✅ Interface-based design
- ✅ Error handling (explicit error returns)
- ✅ Logging and observability
- ✅ Documentation (package-level comments)
- ✅ Testing (unit, integration, E2E)
- ✅ Security (input validation, SQL injection protection)

---

## 📝 Development Guidelines

### Adding a New Resource Type

1. **Create model** in `internal/models/models.go`
2. **Create handler** in `internal/handlers/{resource}_handler.go`
3. **Create service** in `internal/services/{resource}_service.go`
4. **Create repository** in `internal/repositories/{resource}_repository.go`
5. **Create CLI commands** in `cmd/axiomnizamctl/{resource}.go`
6. **Add routes** in `cmd/axiomnizam/main.go`
7. **Add tests** for all layers
8. **Update documentation**

### Adding a New Database Type

1. Add connection logic to `internal/database/connections.go`
2. Add driver import in `main.go`
3. Create handler for dynamic queries
4. Test with sample data
5. Document connection string format

---

## ✅ Compliance Checklist

- ✅ Correct binary organization (cmd/, internal/)
- ✅ Interface-based services
- ✅ Middleware pattern for cross-cutting concerns
- ✅ Error handling with specific error types
- ✅ Configuration management
- ✅ Logging and observability
- ✅ Security best practices
- ✅ Database abstraction
- ✅ Caching implementation
- ✅ Background job system
- ✅ RBAC implementation
- ✅ Event-driven architecture
