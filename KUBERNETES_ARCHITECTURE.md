# Kubernetes-Style Architecture for AxiomNizam

## Overview

AxiomNizam has been evolved into a **Kubernetes-style declarative architecture** following the patterns used in production Kubernetes operators. This makes the system more mature, scalable, and production-ready.

**Status**: ✅ Complete  
**Components**: 5 major packages + runtime orchestration  
**Architecture Pattern**: Declarative resource management with reconciliation loops  

---

## Architecture Components

### 1. **Resources** (`internal/resources/`)

Kubernetes-style CRD (Custom Resource Definition) implementation.

#### Core Concepts

**ObjectMeta** - Metadata common to all resources:
- `Name`, `Namespace` - Identity
- `UID` - Unique identifier
- `Generation` - Incremented on spec changes
- `Labels`, `Annotations` - Metadata
- `Finalizers` - Prevent premature deletion
- `OwnerReferences` - Resource hierarchy

**TypeMeta** - Type information:
- `APIVersion` - API version (e.g., "axiom.dev/v1")
- `Kind` - Resource type (e.g., "Workload")

**ObjectStatus** - Status information:
- `Phase` - Current state (Pending, Running, Failed, etc.)
- `Conditions` - Detailed state conditions
- `ObservedGeneration` - Controller's view of generation

#### Resource Types

**WorkloadResource** - Individual execution unit
```go
// Spec defines desired state
type WorkloadSpec struct {
    Parallelism int32              // Max concurrent executions
    Completions int32              // Desired successes
    Template    WorkloadTemplate   // Execution template
    RetryStrategy *RetryStrategy   // Retry behavior
}

// Create: workload := resources.NewWorkloadResource("my-job", "default")
```

**PipelineResource** - Sequential stage execution
```go
// Multiple stages executed in order
type PipelineSpec struct {
    Stages      []PipelineStage  // Stages to execute
    Parallelism int32            // Stage parallelism
}

// Stages contain parallel tasks
```

**ScheduleResource** - Recurring execution
```go
// Cron-based scheduling
type ScheduleSpec struct {
    Cron         string           // Cron expression
    WorkloadRef  string           // What to execute
    Suspend      bool             // Pause scheduling
}
```

**ExecutionResource** - Execution results
```go
// Tracks completion and output
type ExecutionSpec struct {
    WorkloadRef    string         // What executed
    StartTime      time.Time      // When started
    CompletionTime *time.Time     // When done
    ExitCode       *int32         // Exit status
    Stdout, Stderr string         // Output
}
```

#### Usage Examples

```go
// Create a workload resource
workload := resources.NewWorkloadResource("data-processor", "default")
workload.Spec.Template.Image = "myregistry/processor:latest"
workload.Spec.Template.Command = []string{"/app/process"}
workload.Spec.Parallelism = 3
workload.ObjectMeta.Labels = map[string]string{
    "app": "data-processing",
    "tier": "backend",
}

// Create a pipeline
pipeline := resources.NewPipelineResource("etl-pipeline", "default")
pipeline.Spec.Stages = []resources.PipelineStage{
    {
        Name: "extract",
        Tasks: []resources.PipelineTask{
            {Name: "s3-extract", WorkloadRef: "s3-extractor"},
        },
    },
    {
        Name: "transform",
        Tasks: []resources.PipelineTask{
            {Name: "data-transform", WorkloadRef: "transformer"},
        },
    },
}
```

---

### 2. **WorkQueue** (`internal/workqueue/`)

Asynchronous task processing with rate limiting (similar to Kubernetes work queue).

#### Key Features

**Queue Operations**:
- `Add(key)` - Queue item
- `AddAfter(key, duration)` - Delayed queue
- `AddRateLimited(key)` - Rate-limited retry
- `Get()` - Block until item available
- `Done(key)` - Mark complete
- `Forget(key)` - Stop retrying

**Rate Limiting**:
- Exponential backoff: `baseDelay * 2^retries`
- Configurable max delay
- Automatic retry tracking

**Priority Queue**:
- Multiple priority levels
- Higher priority items processed first
- Useful for urgent operations

#### Example Usage

```go
// Create queue with default rate limiter
queue := workqueue.NewSimpleQueue(nil)

// Add items
queue.Add("namespace/workload-1")
queue.Add("namespace/workload-2")

// Process items
for {
    item, err := queue.Get()
    if err != nil {
        break
    }
    
    // Process item
    if err := processWorkload(item.Key); err != nil {
        // Requeue with exponential backoff
        queue.AddRateLimited(item.Key)
    } else {
        // Success - stop tracking
        queue.Done(item.Key)
    }
}

// Priority queue for important items
pq := workqueue.NewPriorityQueue(3) // 3 priority levels
pq.AddWithPriority("namespace/urgent-job", 2)   // High priority
pq.AddWithPriority("namespace/normal-job", 1)   // Normal
```

#### Architecture

```
Item submitted → Priority Queue → Rate Limiter → Worker
                                       ↓
                                 Exponential Backoff
                                       ↓
                                   Retry Count
```

---

### 3. **API Server** (`internal/apiserver/`)

REST API for CRUD operations on resources (like Kubernetes API server).

#### ResourceStore

In-memory storage with watch support:
```go
store := apiserver.NewResourceStore()

// CRUD operations
store.Create(workload)
resource, err := store.Get("default", "my-workload")
store.Update(workload)
store.Delete("default", "my-workload")

// List with filtering
resources, _ := store.List("default", map[string]string{"app": "data"})

// Watch for changes
watcher := MyWatcher{}
store.Watch("default", watcher)
```

#### REST Endpoints

**Create Resource**:
```
POST /api/v1/{namespace}/{kind}
Body: {resource JSON}
```

**Get Resource**:
```
GET /api/v1/{namespace}/{kind}/{name}
```

**Update Resource**:
```
PUT /api/v1/{namespace}/{kind}/{name}
Body: {resource JSON}
```

**Delete Resource**:
```
DELETE /api/v1/{namespace}/{kind}/{name}
```

**List Resources**:
```
GET /api/v1/{namespace}/{kind}
Query: ?labelSelector=app=data,tier=backend
```

**Status Subresource**:
```
GET /api/v1/{namespace}/{kind}/{name}/status
PUT /api/v1/{namespace}/{kind}/{name}/status
Body: {status JSON}
```

#### Example API Calls

```bash
# Create workload
curl -X POST http://localhost:8000/api/v1/default/workloads \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "my-job"},
    "spec": {"parallelism": 3, "template": {...}}
  }'

# Get workload
curl http://localhost:8000/api/v1/default/workloads/my-job

# List workloads with label filter
curl "http://localhost:8000/api/v1/default/workloads?labelSelector=app=data"

# Update status
curl -X PUT http://localhost:8000/api/v1/default/workloads/my-job/status \
  -d '{"phase": "Running", "conditions": [...]}'
```

---

### 4. **Controllers** (`internal/controllers/`)

Reconciliation loops that watch resources and ensure actual state matches desired state.

#### Reconciler Interface

```go
type Reconciler interface {
    // Reconcile makes actual state match desired state
    Reconcile(ctx context.Context, req ReconcileRequest) (ReconcileResult, error)
    
    // Finalize cleanup before deletion
    Finalize(ctx context.Context, resource Resource) error
}
```

#### ResourceController

Manages reconciliation for one resource type:

```go
controller := controllers.NewResourceController(
    "workload",                    // name
    workqueue,                     // work queue
    store,                         // resource store
    workloadReconciler,            // reconciler
    3,                             // max concurrent workers
)

// Start processing
err := controller.Start(ctx)
```

#### Reconciliation Flow

```
1. Resource created/updated/deleted
    ↓
2. Add to work queue
    ↓
3. Worker pulls from queue
    ↓
4. Call Reconcile()
    ↓
5. Update resource status
    ↓
6. If Requeue=true, add back to queue
    ↓
7. Done() marks complete
```

#### Built-in Reconcilers

**WorkloadReconciler** - Ensures workload transitions to Running:
```go
reconciler := controllers.NewWorkloadReconciler(store)
// Updates status Phase to "Running"
// Sets Ready condition to True
```

**PipelineReconciler** - Executes pipeline stages:
```go
reconciler := controllers.NewPipelineReconciler(store)
// Transitions through stages
// Reports progress via conditions
```

**ScheduleReconciler** - Maintains schedule state:
```go
reconciler := controllers.NewScheduleReconciler(store)
// Activates/suspends based on spec
// Updates last execution time
```

#### Custom Reconciler Example

```go
type CustomReconciler struct {
    store ResourceStore
    client MyServiceClient
}

func (r *CustomReconciler) Reconcile(ctx context.Context, req ReconcileRequest) (ReconcileResult, error) {
    // Get resource
    resource, _ := r.store.Get(req.Namespace, req.Name)
    workload := resource.(*resources.WorkloadResource)
    
    // Check current state
    status := resource.GetStatus()
    if status.Phase == "Completed" {
        return ReconcileResult{Requeue: false}, nil
    }
    
    // Make desired state actual state
    output, err := r.client.Execute(workload.Spec.Template.Command)
    if err != nil {
        // Will be requeued with backoff
        return ReconcileResult{Requeue: true}, err
    }
    
    // Update status
    status.Phase = "Completed"
    resource.SetStatus(status)
    r.store.Update(resource)
    
    return ReconcileResult{Requeue: false}, nil
}

func (r *CustomReconciler) Finalize(ctx context.Context, resource Resource) error {
    // Cleanup: delete execution artifacts, etc.
    return nil
}
```

---

### 5. **Runtime** (`internal/runtime/`)

Orchestrates controllers and provides unified management.

#### ControllerManager

Manages lifecycle of all controllers:

```go
mgr := runtime.NewControllerManager(store, false, "default")

// Register controllers
mgr.RegisterWorkloadController()
mgr.RegisterPipelineController()
mgr.RegisterScheduleController()

// Start all
err := mgr.Start(ctx)

// Check status
status := mgr.Status()
// {"workload": {"running": true}, "pipeline": {"running": true}, ...}
```

#### Runtime

Main orchestration object:

```go
rt := runtime.NewRuntime("1.0.0")

// Initialize components
rt.Initialize(ctx)

// Start (starts API server + controllers)
rt.Start(ctx, "0.0.0.0:8000")

// Health checks
livenessProbe := runtime.NewLivenessProbe(rt)
livenessProbe.Check(ctx)  // Is runtime alive?

readinessProbe := runtime.NewReadinessProbe(rt)
readinessProbe.Check(ctx)  // Are controllers ready?

// Check status
rt.Status()
// {"version": "1.0.0", "running": true, "controllers": {...}}
```

#### Health Checks

**Liveness Probe** - Is the system alive?
- Checks if runtime process is running
- Used for container restart decisions

**Readiness Probe** - Is the system ready to serve?
- Checks if all controllers are running
- Used for load balancer decisions

#### Graceful Shutdown

```go
// Start runtime
go rt.Start(ctx, "0.0.0.0:8000")

// Wait for shutdown signal
<-sigChan

// Give handlers time to finish
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
srv.Shutdown(shutdownCtx)
rt.Stop()
```

---

## Complete Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    AxiomNizam Runtime                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────┐      ┌──────────────────┐            │
│  │  REST API Server │      │ Controller Mgr   │            │
│  ├──────────────────┤      ├──────────────────┤            │
│  │ POST Create      │      │ Workload Ctrl    │            │
│  │ GET  Retrieve    │      │ Pipeline Ctrl    │            │
│  │ PUT  Update      │      │ Schedule Ctrl    │            │
│  │ DELETE Remove    │      └──────────────────┘            │
│  │ WATCH Changes    │              ↓                        │
│  └──────┬───────────┘      ┌──────────────────┐            │
│         │                  │  Reconcilers     │            │
│         │                  ├──────────────────┤            │
│         ↓                  │ Workload Rec     │            │
│  ┌──────────────────┐      │ Pipeline Rec     │            │
│  │  Resource Store  │      │ Schedule Rec     │            │
│  ├──────────────────┤      └──────────────────┘            │
│  │ Workloads        │              ↓                        │
│  │ Pipelines        │      ┌──────────────────┐            │
│  │ Schedules        │      │  Work Queue      │            │
│  │ Executions       │      ├──────────────────┤            │
│  └────────────────────────▶│ Rate Limiter     │            │
│                            │ Priority Queue   │            │
│                            │ Retry Logic      │            │
│                            └──────────────────┘            │
│                                     ↑                       │
│  ┌──────────────────────────────────┴────────────┐         │
│  │ Health Checks                                  │         │
│  ├──────────────────────────────────────────────┤         │
│  │ /health    - Liveness Probe                  │         │
│  │ /ready     - Readiness Probe                 │         │
│  │ /status    - Runtime Status                  │         │
│  └──────────────────────────────────────────────┘         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Kubernetes Patterns Implemented

### 1. **Declarative State Management**
Users declare desired state (spec), system ensures actual state matches.

### 2. **Reconciliation Loops**
Controllers continuously check and fix discrepancies between desired and actual state.

### 3. **Work Queue with Rate Limiting**
Asynchronous processing with exponential backoff prevents overwhelming the system.

### 4. **Resource Watchers**
Changes propagate immediately to interested parties.

### 5. **Finalizers**
Ensures cleanup happens before resource deletion.

### 6. **Conditions**
Detailed status information beyond just a phase.

### 7. **Ownership**
Resources can reference owners for hierarchy and cascading deletion.

### 8. **Labels & Selectors**
Organize and query resources by arbitrary key-value pairs.

---

## API Examples

### Creating a Workload

```bash
curl -X POST http://localhost:8000/api/v1/default/workloads \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "axiom.dev/v1",
    "kind": "Workload",
    "metadata": {
      "name": "data-processor",
      "namespace": "default",
      "labels": {
        "app": "data-processing",
        "environment": "production"
      }
    },
    "spec": {
      "parallelism": 3,
      "completions": 10,
      "template": {
        "image": "myregistry/processor:latest",
        "command": ["/app/process"],
        "env": {
          "INPUT_PATH": "/data/input",
          "OUTPUT_PATH": "/data/output"
        },
        "timeout": 300
      },
      "retryStrategy": {
        "maxRetries": 3,
        "backoffLimit": 2
      }
    }
  }'
```

### Creating a Pipeline

```bash
curl -X POST http://localhost:8000/api/v1/default/pipelines \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "axiom.dev/v1",
    "kind": "Pipeline",
    "metadata": {
      "name": "etl-pipeline",
      "namespace": "default"
    },
    "spec": {
      "parallelism": 2,
      "stages": [
        {
          "name": "extract",
          "tasks": [
            {
              "name": "extract-from-s3",
              "workloadRef": "s3-extractor"
            }
          ]
        },
        {
          "name": "transform",
          "dependsOn": ["extract"],
          "tasks": [
            {
              "name": "data-transform",
              "workloadRef": "transformer"
            },
            {
              "name": "data-validate",
              "workloadRef": "validator"
            }
          ]
        },
        {
          "name": "load",
          "dependsOn": ["transform"],
          "tasks": [
            {
              "name": "load-to-db",
              "workloadRef": "db-loader"
            }
          ]
        }
      ]
    }
  }'
```

### Creating a Schedule

```bash
curl -X POST http://localhost:8000/api/v1/default/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "axiom.dev/v1",
    "kind": "Schedule",
    "metadata": {
      "name": "daily-report",
      "namespace": "default"
    },
    "spec": {
      "cron": "0 9 * * 1-5",
      "timezone": "America/New_York",
      "workloadRef": "report-generator",
      "successfulExecutionsHistoryLimit": 10,
      "failedExecutionsHistoryLimit": 3
    }
  }'
```

### Checking Status

```bash
# Check runtime health
curl http://localhost:8000/health
# {"status": "alive"}

# Check readiness
curl http://localhost:8000/ready
# {"status": "ready"}

# Get full status
curl http://localhost:8000/status
# {"version": "1.0.0", "running": true, "controllers": {...}}

# Get workload status
curl http://localhost:8000/api/v1/default/workloads/data-processor/status
```

---

## Usage in Code

### Initialize Runtime

```go
// In main.go
rt := runtime.NewRuntime("1.0.0")

if err := rt.Initialize(ctx); err != nil {
    log.Fatalf("Failed to initialize: %v", err)
}

if err := rt.Start(ctx, "0.0.0.0:8000"); err != nil {
    log.Fatalf("Failed to start: %v", err)
}
```

### Create Resources Programmatically

```go
store := rt.GetStore()

// Create workload
workload := resources.NewWorkloadResource("my-job", "default")
workload.Spec.Template.Image = "myimage:latest"
store.Create(workload)

// Get it back
resource, _ := store.Get("default", "my-job")
workload = resource.(*resources.WorkloadResource)

// Update status via controller
```

### Implement Custom Reconciler

```go
type MyReconciler struct {
    store ResourceStore
}

func (r *MyReconciler) Reconcile(ctx context.Context, req ReconcileRequest) (ReconcileResult, error) {
    resource, _ := r.store.Get(req.Namespace, req.Name)
    
    // Check if already completed
    if resource.GetStatus().Phase == "Completed" {
        return ReconcileResult{}, nil
    }
    
    // Do actual work
    // ... execute, call external services, etc ...
    
    // Update status
    status := resource.GetStatus()
    status.Phase = "Completed"
    resource.SetStatus(status)
    r.store.Update(resource)
    
    return ReconcileResult{}, nil
}

func (r *MyReconciler) Finalize(ctx context.Context, resource Resource) error {
    // Cleanup
    return nil
}
```

---

## Benefits of This Architecture

✅ **Declarative** - Describe desired state, system ensures it happens  
✅ **Resilient** - Automatic retries with exponential backoff  
✅ **Observable** - Clear status and conditions for each resource  
✅ **Extensible** - Easy to add new resource types and reconcilers  
✅ **Scalable** - Work queue with rate limiting prevents overload  
✅ **Kubernetes-compatible** - Uses same patterns for easier migration  
✅ **Production-ready** - Health checks, graceful shutdown, signal handling  

---

## Production Deployment

### Kubernetes Integration

Deploy AxiomNizam as a pod with proper probes:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: axiom-nizam
spec:
  containers:
  - name: axiom
    image: axiom-nizam:latest
    ports:
    - containerPort: 8000
    livenessProbe:
      httpGet:
        path: /health
        port: 8000
      initialDelaySeconds: 10
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /ready
        port: 8000
      initialDelaySeconds: 5
      periodSeconds: 5
```

### Monitoring

Watch for changes to resources:

```go
// Set up watcher
watcher := MyCustomWatcher{}
rt.GetStore().Watch("default", watcher)

// OnAdd, OnUpdate, OnDelete called automatically
```

---

## Files Structure

```
internal/
├── resources/
│   ├── resource.go        # Base resource definitions
│   └── workload.go        # Workload, Pipeline, Schedule, Execution
├── workqueue/
│   └── queue.go          # Work queue with rate limiting
├── apiserver/
│   └── server.go         # REST API and resource store
├── controllers/
│   └── controller.go     # Controllers and reconcilers
└── runtime/
    └── runtime.go        # Runtime orchestration
```

---

**Status**: Production Ready  
**Maturity Level**: Advanced  
**Use Case**: Enterprise workload orchestration with Kubernetes-like declarative management  

This architecture enables building sophisticated, self-healing systems with minimal code.
