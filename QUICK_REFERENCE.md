# AxiomNizam Quick Reference Guide

**Platform**: Cloud-Native Platform Engine with Kubernetes-style Control Plane  
**Architecture**: Declarative + Reconciliation Loop Pattern  
**Status**: ✅ Production Ready  

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Core Concepts](#core-concepts)
3. [API Reference](#api-reference)
4. [Component Overview](#component-overview)
5. [Common Tasks](#common-tasks)
6. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Starting the Platform

```bash
# Build
go build -o axiom-nizam .

# Run with environment
export DATABASE_URL=postgres://...
export KEYCLOAK_URL=https://...
./axiom-nizam

# Check health
curl http://localhost:8000/health
# Response: {"status":"alive"}

# Check readiness
curl http://localhost:8000/ready
# Response: {"status":"ready"}
```

### Creating Your First Workload

```bash
curl -X POST http://localhost:8000/api/v1/default/workloads \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "axiom.dev/v1",
    "kind": "Workload",
    "metadata": {
      "name": "my-task",
      "labels": {"app": "demo"}
    },
    "spec": {
      "parallelism": 1,
      "completions": 1,
      "template": {
        "image": "myimage:latest",
        "command": ["./run.sh"]
      }
    }
  }'
```

---

## Core Concepts

### Resources (CRDs)

**Definition**: Objects that represent desired state

**Types**:
- **WorkloadResource** - Single task execution
- **PipelineResource** - Multi-stage sequential workflow
- **ScheduleResource** - Recurring execution (cron)
- **ExecutionResource** - Execution history and results

```go
// Every resource has this structure
{
  "apiVersion": "axiom.dev/v1",
  "kind": "Workload",
  "metadata": {
    "name": "task-name",
    "namespace": "default",
    "labels": {"key": "value"}
  },
  "spec": {
    // Desired state (defined by user)
  },
  "status": {
    // Actual state (managed by controller)
    "phase": "Running",
    "conditions": [...]
  }
}
```

### Reconciliation Loop

**Concept**: Controllers continuously work to make actual state match desired state

**Flow**:
```
1. User creates/updates resource
   ↓
2. Controller watches for changes
   ↓
3. Controller runs reconciliation
   ↓
4. If actual ≠ desired, take action
   ↓
5. Update status
   ↓
6. Requeue if needed
```

### Work Queue

**Purpose**: Async processing with rate limiting

**Features**:
- Exponential backoff (1ms → 16s)
- Priority levels
- Automatic retries
- Dead letter queue for failures

### Event Bus

**Purpose**: Publish-subscribe event system

**Event Types**:
- user.created, user.updated, user.deleted
- job.started, job.completed, job.failed
- error.occurred

---

## API Reference

### REST Endpoints

#### Create Resource
```
POST /api/v1/{namespace}/{kind}
Content-Type: application/json

Body: {resource JSON}

Response: 201 Created
{resource with UID, timestamps}
```

#### Get Resource
```
GET /api/v1/{namespace}/{kind}/{name}

Response: 200 OK
{resource}
```

#### Update Resource
```
PUT /api/v1/{namespace}/{kind}/{name}
Content-Type: application/json

Body: {updated resource}

Response: 200 OK
{updated resource}
```

#### Delete Resource
```
DELETE /api/v1/{namespace}/{kind}/{name}

Response: 204 No Content
```

#### List Resources
```
GET /api/v1/{namespace}/{kind}
GET /api/v1/{namespace}/{kind}?labelSelector=key=value

Response: 200 OK
{
  "items": [
    {resource},
    {resource}
  ],
  "count": 2
}
```

#### Get Status
```
GET /api/v1/{namespace}/{kind}/{name}/status

Response: 200 OK
{status object}
```

#### Update Status
```
PUT /api/v1/{namespace}/{kind}/{name}/status
Content-Type: application/json

Body: {status object}

Response: 200 OK
{updated status}
```

### Health Endpoints

```
GET /health       # Liveness probe
# {"status": "alive"}

GET /ready        # Readiness probe
# {"status": "ready"}

GET /status       # Full status
# {"version": "1.0.0", "running": true, ...}
```

### Authentication

```
POST /auth/login
{
  "username": "user",
  "password": "pass"
}
# Response: {"token": "jwt..."}

GET /auth/validate
Authorization: Bearer {token}
# Response: {claims}

GET /auth/token-status
Authorization: Bearer {token}
# Response: {token info}
```

---

## Component Overview

### 1. API Server (`apiserver/`)
- REST API implementation
- Resource storage (in-memory + persistence)
- Watchers for change notifications
- CRUD operations

### 2. Resources (`resources/`)
- CRD-like definitions
- Object metadata (labels, annotations, finalizers)
- Status tracking with conditions
- Serialization

### 3. Controllers (`controllers/`)
- Reconciliation engine
- Finalizer support
- Requeue with backoff
- Periodic resyncing

### 4. Work Queue (`workqueue/`)
- FIFO + Priority queues
- Rate limiter with exponential backoff
- Worker pool
- Thread-safe operations

### 5. Services (`services/`)
- Business logic
- Authentication
- User management
- Caching support

### 6. Events (`events/`)
- Event bus (pub/sub)
- Event history
- Async delivery
- Type-based subscriptions

### 7. Jobs (`jobs/`)
- Background job processing
- Cron scheduling
- Priority queue with fairness
- Dead letter queue
- Email notifications
- Webhook callbacks

### 8. Cache (`cache/`)
- Redis backend (distributed)
- Memory backend (local)
- TTL support
- Manager for provider selection

### 9. Policies (`policies/`)
- RBAC (Role-Based Access Control)
- 4 role levels (Admin, Manager, User, Guest)
- Permission-based access

### 10. Auth (`auth/`)
- Keycloak OIDC integration
- JWT token validation
- Rate limiting
- Middleware chain

### 11. Database (`database/`)
- Multi-database support (8 backends)
- Connection pooling
- GORM abstraction for SQL
- Unified Connections struct

### 12. Runtime (`runtime/`)
- Controller orchestration
- Lifecycle management
- Health probes
- Graceful shutdown

---

## Common Tasks

### Create a Pipeline

```bash
curl -X POST http://localhost:8000/api/v1/default/pipelines \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "axiom.dev/v1",
    "kind": "Pipeline",
    "metadata": {
      "name": "data-pipeline",
      "labels": {"type": "etl"}
    },
    "spec": {
      "stages": [
        {
          "name": "extract",
          "tasks": [
            {"name": "s3-extract", "workloadRef": "extractor"}
          ]
        },
        {
          "name": "transform",
          "dependsOn": ["extract"],
          "tasks": [
            {"name": "data-transform", "workloadRef": "transformer"}
          ]
        },
        {
          "name": "load",
          "dependsOn": ["transform"],
          "tasks": [
            {"name": "db-load", "workloadRef": "loader"}
          ]
        }
      ]
    }
  }'
```

### Create a Schedule

```bash
curl -X POST http://localhost:8000/api/v1/default/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "axiom.dev/v1",
    "kind": "Schedule",
    "metadata": {
      "name": "daily-report"
    },
    "spec": {
      "cron": "0 9 * * 1-5",
      "timezone": "America/New_York",
      "workloadRef": "report-generator"
    }
  }'
```

### Monitor a Resource

```bash
# Get status
curl http://localhost:8000/api/v1/default/workloads/my-task/status

# Response:
{
  "phase": "Running",
  "conditions": [
    {
      "type": "Ready",
      "status": "True",
      "reason": "WorkloadRunning",
      "message": "Workload is executing"
    }
  ],
  "observedGeneration": 1
}
```

### Query with Labels

```bash
# Find all workloads in production
curl "http://localhost:8000/api/v1/default/workloads?labelSelector=env=prod"

# Find by multiple labels
curl "http://localhost:8000/api/v1/default/workloads?labelSelector=env=prod,tier=backend"
```

### Add a Label to Resource

```bash
# Get the resource
curl http://localhost:8000/api/v1/default/workloads/my-task > task.json

# Edit task.json to add label
# "labels": {"env": "prod"}

# Update
curl -X PUT http://localhost:8000/api/v1/default/workloads/my-task \
  -H "Content-Type: application/json" \
  -d @task.json
```

### Delete a Resource

```bash
curl -X DELETE http://localhost:8000/api/v1/default/workloads/my-task

# If resource has finalizers, deletion waits until controller cleans up
```

### Submit a Background Job

```bash
curl -X POST http://localhost:8000/jobs/submit \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "priority": 5,
    "data": {
      "to": "user@example.com",
      "subject": "Hello",
      "body": "Welcome!"
    },
    "maxRetries": 3,
    "callbackURL": "https://example.com/callback"
  }'
```

---

## Troubleshooting

### Platform Won't Start

**Problem**: `Failed to initialize runtime`

**Solutions**:
1. Check environment variables: `KEYCLOAK_URL`, `DATABASE_URL`
2. Check database connections
3. Check Redis/Valkey connection
4. View logs for details

### Resources Not Reconciling

**Problem**: Workload status stuck in "Pending"

**Solutions**:
1. Check controller logs: `grep WorkloadReconciler`
2. Verify resource doesn't have finalizers blocking it
3. Check work queue isn't overloaded
4. Verify reconciler can access the resource

### Rate Limiting Too Aggressive

**Problem**: Jobs getting backed off too quickly

**Solutions**:
1. Check `RateLimiting` config
2. Reduce `MaxCallsPerToken`
3. Increase `TokenValidityMinutes`
4. Check if system is overloaded

### Memory Growing

**Problem**: Memory usage increasing

**Solutions**:
1. Check event history limit (maxHistory)
2. Check cache TTL settings
3. Check for goroutine leaks: `pprof`
4. Clear old execution records

### High Latency

**Problem**: API calls slow

**Solutions**:
1. Check database connection pooling
2. Check cache hit rates
3. Check work queue depth
4. Profile with pprof

### Authentication Failing

**Problem**: `Keycloak initialization failed`

**Solutions**:
1. Check Keycloak is running
2. Verify `KEYCLOAK_URL` is correct
3. Check network connectivity
4. Verify realm and client ID
5. Check JWT signing key availability

### Database Connection Issues

**Problem**: `MySQL connection failed`

**Solutions**:
1. Check DSN format
2. Check host/port/credentials
3. Check database is running
4. Check firewall rules
5. Check max connections limit

---

## Configuration Reference

### Environment Variables

```bash
# Port
PORT=8000

# Databases
MYSQL_DSN=user:pass@tcp(localhost:3306)/db
POSTGRES_DSN=host=localhost user=postgres password=pass

# Keycloak
KEYCLOAK_URL=https://keycloak.example.com
KEYCLOAK_REALM=master
KEYCLOAK_CLIENT_ID=axiom

# Rate Limiting
MAX_CALLS_PER_TOKEN=100
TOKEN_VALIDITY_MINUTES=60

# Redis/Valkey
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# Discord (notifications)
DISCORD_WEBHOOK_URL=https://...
```

---

## Performance Tips

1. **Increase Cache TTL** for frequently accessed resources
2. **Adjust Work Queue** priority distribution
3. **Configure Rate Limiter** based on load
4. **Use Redis** for distributed caching
5. **Index Database** for common queries
6. **Scale Controllers** for more throughput
7. **Monitor Metrics** via health endpoints
8. **Use Jobs** for async processing
9. **Batch Operations** when possible
10. **Review Logs** for bottlenecks

---

## Security Best Practices

1. ✅ Use HTTPS in production (LB layer)
2. ✅ Enable RBAC enforcement
3. ✅ Rotate JWT signing keys
4. ✅ Use strong Keycloak configuration
5. ✅ Limit API rate per user
6. ✅ Validate all inputs
7. ✅ Use strong database passwords
8. ✅ Enable Redis AUTH
9. ✅ Restrict network access
10. ✅ Monitor security logs

---

## Architecture Diagram

```
┌─────────────────────────────────────────────┐
│      User / External Service                 │
└────────────┬────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────┐
│      REST API (port 8000)                   │
├─────────────────────────────────────────────┤
│  • POST /api/v1/{ns}/{kind}  - Create      │
│  • GET  /api/v1/{ns}/{kind}  - List        │
│  • PUT  /api/v1/{ns}/{kind}  - Update      │
│  • DELETE /api/v1/{ns}/{kind} - Delete     │
└────────────┬────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────┐
│    Resource Store (in-memory + DB)          │
├─────────────────────────────────────────────┤
│  • Workloads                                │
│  • Pipelines                                │
│  • Schedules                                │
│  • Executions                               │
└────────────┬────────────────────────────────┘
             │
      ┌──────┴──────┐
      │             │
      ▼             ▼
  ┌────────┐   ┌────────────────┐
  │Watchers│   │Controllers     │
  └────────┘   ├────────────────┤
               │Reconcilers:    │
               │• Workload      │
               │• Pipeline      │
               │• Schedule      │
               └────────┬───────┘
                        │
                        ▼
                ┌────────────────┐
                │  Work Queue    │
                ├────────────────┤
                │• Rate Limiter  │
                │• Priority Routing
                │• Retry Backoff │
                └────────┬───────┘
                         │
                ┌────────┴─────────┐
                │                  │
                ▼                  ▼
         ┌────────────┐   ┌──────────────┐
         │Job Processing  │Event Bus     │
         ├────────────┤   ├──────────────┤
         │• Queue     │   │• Pub/Sub     │
         │• Scheduler │   │• History     │
         │• DLQ       │   │• Handlers    │
         └────────────┘   └──────────────┘
                │                  │
                └────────┬─────────┘
                         │
                ┌────────┴────────┐
                │                 │
                ▼                 ▼
         ┌────────────┐   ┌──────────────┐
         │Databases   │   │Cache (Redis) │
         ├────────────┤   ├──────────────┤
         │• MySQL     │   │• Sessions    │
         │• PostgreSQL│   │• Data        │
         │• MongoDB   │   │• TTL         │
         │• Oracle    │   └──────────────┘
         │• Etcd      │
         │• Firebasе  │
         └────────────┘
```

---

## Support & Documentation

- **Architecture**: See KUBERNETES_ARCHITECTURE.md
- **Compliance**: See ARCHITECTURE_COMPLIANCE_REPORT.md
- **Analysis**: See CODEBASE_ANALYSIS.md
- **Full Docs**: See COMPLETE_DOCUMENTATION.md

---

**Last Updated**: January 24, 2026  
**Status**: ✅ Production Ready
