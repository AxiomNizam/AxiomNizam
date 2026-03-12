# AxiomNizam — Architecture Flowchart

## Platform Architecture

```mermaid
graph TB
    subgraph Clients["Clients"]
        CLI["axiomnizamctl\n(kubectl-style CLI)"]
        Browser["Web Browser"]
        Postman["Postman / HTTP Client"]
    end

    subgraph Frontend["Frontend — Port 7000"]
        FE_Server["Gin Server"]
        FE_Dash["Main Dashboard"]
        FE_Admin["Admin Dashboard"]
        FE_GIS["GIS Dashboard\n(Leaflet.js)"]
        FE_Analytics["Analytics Dashboard\n(Chart.js)"]
        FE_CDCETL["CDC/ETL Dashboard"]
        FE_NetIntel["NetIntel Dashboard"]
    end

    subgraph APIServer["API Server — Port 8000"]
        direction TB
        Router["Gin HTTP Router"]

        subgraph Middleware["Middleware Chain"]
            CORS["CORS"]
            RateLimit["Rate Limiter"]
            JWT_MW["JWT Auth"]
            Metrics_MW["Metrics Tracker"]
            RoleMW["RBAC Check"]
        end

        subgraph Handlers["29 HTTP Handlers"]
            H_Health["Health / Status"]
            H_Auth["Auth"]
            H_UserCRUD["User CRUD\n(7 databases)"]
            H_DynQuery["Dynamic Query\n(5 SQL DBs)"]
            H_Resource["K8s Resource API"]
            H_GIS["GIS"]
            H_Analytics["Analytics"]
            H_CDCETL["CDC/ETL"]
            H_NetIntel["NetIntel"]
            H_Transform["Transform"]
            H_DS["DataSource"]
            H_Job["Job"]
            H_Admin["Admin"]
            H_Notif["Notifications"]
        end

        subgraph FeatureAPIs["13 Enterprise Features"]
            F_Audit["Audit"]
            F_Tenant["Tenant"]
            F_Streaming["Streaming\n(WebSocket)"]
            F_Bulk["Bulk Ops"]
            F_Version["Versioning"]
            F_Webhook["Webhooks"]
            F_EventBus["Event Bus"]
            F_Tracing["Tracing"]
            F_Export["Export"]
            F_Lineage["Lineage"]
            F_Encryption["Encryption"]
            F_RBAC["RBAC"]
        end
    end

    subgraph ControlPlane["Kubernetes-Style Control Plane"]
        direction TB
        APIStore["API Server\n(Resource Store)"]
        Informer["Informer\n(Watch Changes)"]
        WorkQ["Work Queue\n(Priority + Rate Limit)"]
        
        subgraph Controllers["Controllers"]
            C_Workload["Workload\nReconciler"]
            C_Pipeline["Pipeline\nReconciler"]
            C_Schedule["Schedule\nReconciler"]
        end

        EventBus_Core["Event Bus\n(Pub/Sub)"]
        Runtime["Runtime\n(Controller Manager)"]
        StatusTracker["Status Tracker"]
    end

    subgraph DataIntegration["Data Integration"]
        ETL["ETL Engine\n(10 step types)"]
        CDC["CDC Engine\n(Change Capture)"]
        Workflows["Workflow Engine"]
        JobQueue["Job Queue\n(Priority, DLQ,\nCron, Fairness)"]
        GraphQL_Engine["GraphQL\n(Dynamic Schema)"]
    end

    subgraph Security["Policy and Security"]
        Keycloak["Keycloak\n(OIDC, Port 8080)"]
        JWT_Val["JWT Validator"]
        PolicyEngine["Policy Engine\n(CEL / Rego / DSL)"]
        RLS["Row-Level Security"]
        FieldEncrypt["Field Encryption\n(AES-256-GCM)"]
    end

    subgraph Observability["Observability"]
        AuditLog["Audit Logger\n(GDPR, HIPAA,\nSOC2, PCI-DSS)"]
        Tracing["Distributed Tracing\n(OpenTelemetry)"]
        PromMetrics["Prometheus Metrics"]
        QualityCheck["Data Quality"]
        LineageTracker["Data Lineage"]
        PerfAnalyzer["Performance\nAnalyzer"]
    end

    subgraph DataLayer["Data Layer"]
        subgraph SQL_DBs["SQL Databases"]
            PG["PostgreSQL"]
            MySQL["MySQL"]
            MariaDB["MariaDB"]
            Percona["Percona"]
            Oracle["Oracle"]
        end
        
        subgraph NoSQL_DBs["NoSQL / Other"]
            MongoDB["MongoDB"]
            Valkey["Valkey/Redis\n(Cache + Queue)"]
            ES["Elasticsearch"]
            Etcd["etcd\n(Distributed State)"]
        end
    end

    %% Client connections
    CLI -->|"HTTP + JWT"| Router
    Browser -->|HTTP| FE_Server
    Postman -->|"HTTP + JWT"| Router

    %% Frontend to backend
    FE_Server --> FE_Dash & FE_Admin & FE_GIS & FE_Analytics & FE_CDCETL & FE_NetIntel
    FE_Server -->|"Proxy to :8000"| Router

    %% Router through middleware to handlers
    Router --> Middleware
    Middleware --> Handlers
    Middleware --> FeatureAPIs

    %% Handlers to control plane
    H_Resource -->|"Create/Update"| APIStore
    APIStore -->|"Notify"| Informer
    Informer -->|"Enqueue"| WorkQ
    WorkQ --> Controllers
    Controllers -->|"Update"| StatusTracker
    Controllers -->|"Publish"| EventBus_Core
    Runtime -->|"Manages"| Controllers

    %% Handlers to data integration
    H_CDCETL --> ETL & CDC
    H_Job --> JobQueue
    H_DynQuery --> SQL_DBs

    %% Security
    JWT_MW --> JWT_Val
    JWT_Val --> Keycloak
    RoleMW --> PolicyEngine
    H_Auth --> Keycloak
    FieldEncrypt --> SQL_DBs

    %% Data connections
    H_UserCRUD --> SQL_DBs & MongoDB
    H_Analytics --> Valkey
    Handlers --> SQL_DBs & NoSQL_DBs
    FeatureAPIs --> SQL_DBs & NoSQL_DBs
    JobQueue --> Valkey
    ETL --> SQL_DBs
    CDC --> SQL_DBs

    %% Observability connections
    Handlers -.->|"Log"| AuditLog
    Controllers -.->|"Trace"| Tracing
    Handlers -.->|"Metrics"| PromMetrics

    %% Distributed
    Runtime -->|"Leader Election"| Etcd
```

## Request Flow

```mermaid
sequenceDiagram
    participant C as Client (CLI/Browser)
    participant R as Gin Router
    participant MW as Middleware
    participant H as Handler
    participant S as Resource Store
    participant I as Informer
    participant WQ as Work Queue
    participant CR as Controller
    participant DB as Database

    C->>R: HTTP Request + JWT Token
    R->>MW: CORS → Rate Limit → JWT Auth → Metrics
    MW->>MW: Validate token via Keycloak JWKS
    MW->>MW: Check RBAC role/permissions
    MW->>H: Route to handler

    alt Database CRUD
        H->>DB: GORM query (SQL) or Native (Mongo)
        DB-->>H: Result
        H-->>C: JSON Response
    end

    alt K8s Resource API
        H->>S: Store resource (desired state)
        S->>I: Notify watcher
        I->>WQ: Enqueue reconcile request
        WQ->>CR: Dequeue with rate limiting
        CR->>CR: Reconcile (desired vs actual)
        CR->>S: Update status + conditions
        S-->>C: Resource with status
    end

    alt ETL/CDC Pipeline
        H->>DB: Create pipeline record
        H->>WQ: Enqueue pipeline run
        WQ->>CR: Execute pipeline steps
        CR->>DB: Extract → Transform → Load
        CR-->>C: Run status
    end
```

## Control Plane Reconciliation Loop

```mermaid
graph LR
    A["User applies YAML"] --> B["API Server stores resource"]
    B --> C["Informer detects change"]
    C --> D["Work Queue enqueues"]
    D --> E["Controller dequeues"]
    E --> F{"Desired == Actual?"}
    F -->|No| G["Execute changes"]
    G --> H["Update status"]
    H --> I{"Success?"}
    I -->|Yes| J["Status: Succeeded\nCondition: Ready"]
    I -->|No| K["Requeue with\nexponential backoff"]
    K --> D
    F -->|Yes| J
```

## Data Flow Architecture

```mermaid
graph LR
    subgraph Sources["Data Sources"]
        S1["PostgreSQL"]
        S2["MySQL"]
        S3["MongoDB"]
        S4["External APIs"]
        S5["CSV/Files"]
    end

    subgraph Processing["Processing"]
        ETL["ETL Engine"]
        CDC["CDC Engine"]
        Transform["Transformer"]
        Quality["Quality Validator"]
    end

    subgraph Storage["Storage & Cache"]
        DB["Primary Databases"]
        Cache["Valkey/Redis Cache"]
        ES["Elasticsearch Index"]
    end

    subgraph Output["Output"]
        API["REST API"]
        WS["WebSocket Streams"]
        Export["Export (CSV/JSON/Parquet)"]
        Webhook["Webhooks"]
        GQL["GraphQL"]
    end

    Sources --> ETL
    Sources --> CDC
    ETL --> Transform --> Quality --> DB
    CDC --> DB
    DB --> Cache
    DB --> ES
    DB --> API & WS & Export & Webhook & GQL
    Cache --> API
```

## Authentication Flow

```mermaid
sequenceDiagram
    participant U as User/CLI
    participant API as API Server
    participant KC as Keycloak
    participant RL as Rate Limiter

    U->>KC: POST /token (credentials)
    KC-->>U: JWT Access Token + Refresh Token

    U->>API: Request + Authorization: Bearer <JWT>
    API->>API: Parse JWT, extract claims
    API->>KC: Validate via JWKS endpoint
    KC-->>API: Public key for verification
    API->>API: Verify signature + expiry
    API->>RL: Check rate limit for token
    RL-->>API: Allowed (or 429 Too Many Requests)
    API->>API: Check RBAC role (admin/user/guest)
    API-->>U: Response (or 401/403)
```
