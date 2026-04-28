# AxiomNizam — Platform Completion Plan

**Version:** 1.0  
**Date:** 2026-04-28  
**Status:** Planning — implementation not started  
**Author:** Platform Architecture Team

---

## Executive Summary

AxiomNizam is already a powerful declarative control-plane with 33 reconcilers, 88 internal modules, and 174K+ lines of Go. This document defines the roadmap to transform it from a strong data platform into a **complete, production-grade enterprise data platform** that competes with Databricks, Snowflake, and Confluent — while maintaining the unique K8s-style declarative architecture.

**Scope:** 25 new capabilities across 7 workstreams, estimated 12-16 weeks of engineering.

---

## Design Principles (Inherited)

All new modules MUST follow the established AxiomNizam patterns:

1. **Declarative resources** — Every entity is `TypeMeta + ObjectMeta + Spec + Status`
2. **Reconcile loops** — Controllers drive state, never HTTP handlers
3. **Feature-flagged** — `RECONCILER_ENABLED_<MODULE>=true` gates activation
4. **Dual-write migration** — New path writes alongside old path until validated
5. **Observe before acting** — Metrics + structured logging before production traffic
6. **etcd is truth** — All state in etcd; external systems reached only by reconcilers

**Resource template** (every new module follows this):

```go
// internal/<module>/resource.go
type <Name>Resource struct {
    resources.TypeMeta   `json:",inline"`
    resources.ObjectMeta `json:"metadata"`
    Spec   <Name>Spec           `json:"spec"`
    Status <Name>ResourceStatus `json:"status"`
}

// internal/<module>/reconciler.go
func (r *<Name>Reconciler) Reconcile(ctx context.Context, obj reconciler.Resource) reconciler.ReconcileResult {
    // Observe -> Diff -> Act -> Update Status
}
```

---

## Current State Baseline

| Metric | Value |
|--------|-------|
| Internal modules | 88 |
| Active reconciler controllers | 33 |
| Resource types | 30+ |
| etcd prefixes | 30 |
| External integrations | 13 |
| API route groups | 40+ |
| Frontend dashboards | 12 |
| Go lines | 174,000+ |

**Existing capabilities that new modules build upon:**

- Lineage tracking with column-level granularity, impact analysis, and OpenLineage support
- ETL engine with 10-step pipeline types and declarative PipelineResource
- CDC with source/sink/filter model and CDCPipelineResource
- RBAC with role/rolebinding resources
- Encryption with key rotation and field-level AES-256-GCM
- Multi-tenant isolation
- Event bus with pub/sub, DLQ, and replay
- Object storage with bucket policies and lifecycle
- Audit logging with compliance reporting

---

## Workstream Overview

| # | Workstream | Modules | Priority | Weeks |
|---|-----------|---------|----------|-------|
| WS-1 | Data Catalog and Metadata | 3 | P0 Critical | 3 |
| WS-2 | Data Quality and Contracts | 3 | P0 Critical | 3 |
| WS-3 | Schema Registry and Evolution | 2 | P0 Critical | 2 |
| WS-4 | Observability and Alerting | 4 | P1 High | 3 |
| WS-5 | Federated Query and Virtualization | 3 | P1 High | 3 |
| WS-6 | Governance and Compliance | 3 | P1 High | 2 |
| WS-7 | Advanced Analytics and ML | 4 | P2 Medium | 3 |

**Total new modules:** 22  
**Total estimated effort:** 12-16 weeks (parallelizable to 8 weeks with 2 engineers)

---

## WS-1: Data Catalog and Metadata (P0 Critical — Week 1-3)

### Why This Is Critical

Every enterprise data platform needs a unified metadata registry. Without a catalog, lineage is disconnected, governance has no inventory to govern, and users cannot discover what data exists. The catalog is the **foundation** that WS-2 (Quality), WS-5 (Federated Query), and WS-6 (Governance) all depend on.

**Competitive context:** Databricks Unity Catalog, Snowflake Information Schema, Google Data Catalog, Apache Atlas, OpenMetadata.

### Module 1.1: Catalog Core (`internal/catalog`)

**Purpose:** Unified metadata registry for all data assets across all connected databases.

**New files:**

| File | Purpose |
|------|---------|
| `internal/catalog/resource.go` | `CatalogAssetResource`, `CatalogCollectionResource` |
| `internal/catalog/reconciler.go` | `CatalogAssetReconciler` — syncs metadata from datasources |
| `internal/catalog/scanner.go` | Auto-discovery: introspects connected databases for tables/views/columns |
| `internal/catalog/indexer.go` | Full-text search index over asset metadata |
| `internal/catalog/handlers.go` | REST API: search, browse, tag, annotate |
| `internal/catalog/models.go` | Asset types, column metadata, statistics |

**Resource definitions:**

```go
const (
    CatalogAssetKind       = "CatalogAsset"
    CatalogAssetAPIVersion = "catalog.axiomnizam.io/v1"
)

type CatalogAssetSpec struct {
    AssetType      string            `json:"assetType"`      // table, view, topic, bucket, api, pipeline
    DataSourceRef  string            `json:"dataSourceRef"`  // Reference to DataSource resource
    Database       string            `json:"database"`
    Schema         string            `json:"schema"`
    Name           string            `json:"name"`
    Columns        []CatalogColumn   `json:"columns"`
    Owner          string            `json:"owner"`
    Domain         string            `json:"domain"`         // Business domain (sales, finance, etc.)
    Description    string            `json:"description"`
    Tags           []string          `json:"tags"`
    Classification DataClassification `json:"classification"`
    RefreshPolicy  RefreshPolicy     `json:"refreshPolicy"`  // How often to re-scan
}

type CatalogAssetResourceStatus struct {
    resources.ObjectStatus `json:",inline"`
    RowCount        int64      `json:"rowCount"`
    SizeBytes       int64      `json:"sizeBytes"`
    LastScannedAt   *time.Time `json:"lastScannedAt"`
    LastModifiedAt  *time.Time `json:"lastModifiedAt"`
    QualityScore    float64    `json:"qualityScore"`    // From WS-2
    PopularityScore float64    `json:"popularityScore"` // Query frequency
    ColumnCount     int        `json:"columnCount"`
    IndexCount      int        `json:"indexCount"`
}
```

**Reconciler behavior:**

1. **Observe:** Read CatalogAssetResource from etcd
2. **Diff:** Compare spec.refreshPolicy against status.lastScannedAt
3. **Act:** If stale, connect to datasource via existing `internal/datasource` registry, introspect schema, update column metadata, row counts, size
4. **Update Status:** Write back discovered metadata, quality score, popularity

**Integration points:**

- Reads from `internal/datasource` for connection details
- Feeds into `internal/lineage` for automatic node creation
- Feeds into WS-2 quality checks
- Indexed by `internal/catalog/indexer.go` for full-text search

**CLI commands:**

```
axiomnizamctl catalog list [--domain=sales] [--type=table]
axiomnizamctl catalog get <asset-name>
axiomnizamctl catalog search "customer revenue"
axiomnizamctl catalog scan <datasource-name>    # Trigger discovery
axiomnizamctl catalog describe <asset-name>     # Full metadata + lineage + quality
```

**API routes:**

```
GET    /api/v1/catalog/assets
GET    /api/v1/catalog/assets/:name
POST   /api/v1/catalog/assets                   # Register manually
DELETE /api/v1/catalog/assets/:name
GET    /api/v1/catalog/search?q=<query>
POST   /api/v1/catalog/scan/:datasource         # Trigger scan
GET    /api/v1/catalog/domains                   # List business domains
GET    /api/v1/catalog/statistics                # Platform-wide stats
```

**Environment variables:**

```
RECONCILER_ENABLED_CATALOG=true
CATALOG_SCAN_INTERVAL=1h
CATALOG_MAX_COLUMNS_PER_TABLE=1000
CATALOG_SEARCH_INDEX_PATH=/data/catalog_index
```

---

### Module 1.2: Metadata Enrichment (`internal/catalog/enrichment`)

**Purpose:** Automatic metadata enrichment via profiling, classification, and usage tracking.

**Capabilities:**

- **Column profiling:** min/max/avg/null_count/distinct_count/histogram for each column
- **PII detection:** Regex + heuristic patterns for email, phone, SSN, credit card, IP
- **Auto-classification:** Based on column names, data patterns, and PII detection
- **Usage tracking:** Query frequency per asset (integrates with existing query logger)
- **Freshness detection:** Last-modified timestamps, staleness alerts

**Reconciler:** `MetadataEnrichmentReconciler` runs on a schedule (default: every 6 hours), profiles a batch of assets, and updates their CatalogAssetResource status with enrichment data.

---

### Module 1.3: Catalog UI Dashboard (`frontend/templates/catalog-dashboard`)

**Purpose:** Visual catalog browser with search, browse-by-domain, and asset detail pages.

**New frontend files:**

```
frontend/templates/catalog-dashboard.html
frontend/templates/catalog-dashboard.js
frontend/templates/catalog-dashboard.css
```

**Features:**

- Full-text search with faceted filtering (domain, type, owner, classification)
- Asset detail page: columns, statistics, lineage graph, quality score, recent queries
- Domain tree browser (hierarchical navigation)
- Popular assets ranking
- Recently modified assets feed

---

## WS-2: Data Quality and Contracts (P0 Critical — Week 2-4)

### Why This Is Critical

Pipelines without quality checks are a liability. Data quality issues propagate downstream silently, causing bad decisions and broken dashboards. Data contracts prevent breaking changes between producers and consumers. Together they provide **trust** in the platform.

**Competitive context:** Great Expectations, Soda Core, Monte Carlo, dbt tests, Dataform assertions.

### Module 2.1: Quality Rules Engine (`internal/quality/rules`)

**Purpose:** Declarative data quality validation rules that run as reconciled resources.

**Note:** The existing `internal/quality` module has ReviewFlow scoring. This extends it with a proper rules engine.

**New files:**

| File | Purpose |
|------|---------|
| `internal/quality/rules/resource.go` | `QualityRuleResource`, `QualityCheckResource` |
| `internal/quality/rules/reconciler.go` | `QualityRuleReconciler` — executes checks on schedule |
| `internal/quality/rules/engine.go` | Rule evaluation engine (SQL, statistical, custom) |
| `internal/quality/rules/checks.go` | Built-in check types |
| `internal/quality/rules/handlers.go` | REST API for rules CRUD and check results |
| `internal/quality/rules/alerting.go` | Integration with WS-4 alerting |

**Resource definitions:**

```go
const (
    QualityRuleKind       = "QualityRule"
    QualityRuleAPIVersion = "quality.axiomnizam.io/v1"
)

type QualityRuleSpec struct {
    AssetRef       string            `json:"assetRef"`       // CatalogAsset reference
    DataSourceRef  string            `json:"dataSourceRef"`  // DataSource to query
    RuleType       QualityRuleType   `json:"ruleType"`       // freshness, volume, schema, custom_sql, statistical
    Schedule       string            `json:"schedule"`       // Cron expression
    Severity       string            `json:"severity"`       // critical, warning, info
    
    // Rule-type-specific config
    Freshness      *FreshnessRule    `json:"freshness,omitempty"`
    Volume         *VolumeRule       `json:"volume,omitempty"`
    Schema         *SchemaRule       `json:"schema,omitempty"`
    CustomSQL      *CustomSQLRule    `json:"customSQL,omitempty"`
    Statistical    *StatisticalRule  `json:"statistical,omitempty"`
    Nullness       *NullnessRule     `json:"nullness,omitempty"`
    Uniqueness     *UniquenessRule   `json:"uniqueness,omitempty"`
    Range          *RangeRule        `json:"range,omitempty"`
    Regex          *RegexRule        `json:"regex,omitempty"`
    Referential    *ReferentialRule  `json:"referential,omitempty"`
    
    // Alerting
    AlertOnFailure bool              `json:"alertOnFailure"`
    AlertChannels  []string          `json:"alertChannels"`  // slack, email, webhook
    
    // SLA
    SLA            *QualitySLA       `json:"sla,omitempty"`
}

type FreshnessRule struct {
    MaxAge         string `json:"maxAge"`         // "2h", "24h", "7d"
    TimestampColumn string `json:"timestampColumn"`
}

type VolumeRule struct {
    MinRows        int64  `json:"minRows"`
    MaxRows        int64  `json:"maxRows"`
    GrowthRate     string `json:"growthRate"`     // "5%" — max deviation from average
}

type CustomSQLRule struct {
    Query          string `json:"query"`          // Must return 0 rows to pass
    Threshold      int64  `json:"threshold"`      // Max allowed failing rows
}

type StatisticalRule struct {
    Column         string  `json:"column"`
    Metric         string  `json:"metric"`        // mean, stddev, median, p95, p99
    MinValue       float64 `json:"minValue"`
    MaxValue       float64 `json:"maxValue"`
    DeviationPct   float64 `json:"deviationPct"`  // Max % deviation from historical
}

type QualityRuleResourceStatus struct {
    resources.ObjectStatus `json:",inline"`
    LastCheckAt     *time.Time       `json:"lastCheckAt"`
    LastResult      CheckResult      `json:"lastResult"`      // pass, fail, error, skip
    ConsecutiveFails int             `json:"consecutiveFails"`
    TotalChecks     int64            `json:"totalChecks"`
    TotalPasses     int64            `json:"totalPasses"`
    TotalFailures   int64            `json:"totalFailures"`
    PassRate        float64          `json:"passRate"`
    NextCheckAt     *time.Time       `json:"nextCheckAt"`
    SLAStatus       string           `json:"slaStatus"`       // met, at_risk, breached
}
```

**Built-in check types (15):**

1. `freshness` — Data not older than X
2. `volume` — Row count within expected range
3. `schema` — Column types/names match expectation
4. `not_null` — Column has no nulls (or below threshold)
5. `unique` — Column values are unique
6. `accepted_values` — Column only contains allowed values
7. `range` — Numeric column within min/max
8. `regex` — String column matches pattern
9. `referential` — Foreign key exists in reference table
10. `custom_sql` — Arbitrary SQL returns 0 failing rows
11. `statistical` — Metric within historical bounds
12. `row_count_change` — Day-over-day change within threshold
13. `completeness` — Percentage of non-null values above threshold
14. `distribution` — Value distribution matches expected shape
15. `timeliness` — Data arrives within SLA window

**Reconciler behavior:**

1. **Observe:** Read QualityRuleResource, check schedule against lastCheckAt
2. **Diff:** If check is due (cron matches or manual trigger)
3. **Act:** Connect to datasource, execute check query, evaluate result
4. **Update Status:** Record pass/fail, update consecutive fails, trigger alert if needed

---

### Module 2.2: Data Contracts (`internal/contracts`)

**Purpose:** Schema contracts between data producers and consumers with breaking-change detection.

**New files:**

| File | Purpose |
|------|---------|
| `internal/contracts/resource.go` | `DataContractResource` |
| `internal/contracts/reconciler.go` | `DataContractReconciler` — validates contracts against actual schemas |
| `internal/contracts/validator.go` | Breaking change detection logic |
| `internal/contracts/handlers.go` | REST API |

**Resource definition:**

```go
type DataContractSpec struct {
    Producer       string              `json:"producer"`       // Team/service that owns the data
    Consumer       []string            `json:"consumers"`      // Teams that depend on it
    AssetRef       string              `json:"assetRef"`       // CatalogAsset reference
    SchemaVersion  string              `json:"schemaVersion"`  // semver
    Schema         ContractSchema      `json:"schema"`         // Expected schema
    SLA            ContractSLA         `json:"sla"`            // Freshness, availability
    Quality        ContractQuality     `json:"quality"`        // Min quality score
    Compatibility  CompatibilityMode   `json:"compatibility"`  // backward, forward, full, none
    NotifyOnBreak  []string            `json:"notifyOnBreak"`  // Channels to notify
}

type ContractSchema struct {
    Columns        []ContractColumn    `json:"columns"`
    PrimaryKey     []string            `json:"primaryKey"`
    RequiredColumns []string           `json:"requiredColumns"`
    ForbiddenChanges []string          `json:"forbiddenChanges"` // drop_column, change_type, rename
}

type CompatibilityMode string
const (
    CompatBackward CompatibilityMode = "backward"  // New schema can read old data
    CompatForward  CompatibilityMode = "forward"   // Old schema can read new data
    CompatFull     CompatibilityMode = "full"      // Both directions
    CompatNone     CompatibilityMode = "none"      // No compatibility guarantee
)
```

**Breaking change detection rules:**

| Change | Backward | Forward | Full |
|--------|----------|---------|------|
| Add optional column | OK | OK | OK |
| Add required column | BREAK | OK | BREAK |
| Remove column | OK | BREAK | BREAK |
| Rename column | BREAK | BREAK | BREAK |
| Widen type (int32 to int64) | OK | BREAK | BREAK |
| Narrow type (int64 to int32) | BREAK | OK | BREAK |
| Change nullability (nullable to required) | BREAK | OK | BREAK |

---

### Module 2.3: Quality Dashboard (`frontend/templates/quality-dashboard`)

**Purpose:** Visual quality monitoring with SLA tracking and trend analysis.

**Features:**

- Quality score heatmap across all assets
- Rule pass/fail timeline (last 30 days)
- SLA status board (met / at-risk / breached)
- Contract compliance matrix (producer vs consumer)
- Drill-down: per-asset quality history with root cause
- Alert feed with acknowledgment

---

## WS-3: Schema Registry and Evolution (P0 Critical — Week 2-3)

### Why This Is Critical

AxiomNizam has Kafka and CDC pipelines but no schema registry. Without one, producers can push breaking schema changes that silently corrupt downstream consumers. A schema registry provides **type safety for data in motion**.

**Competitive context:** Confluent Schema Registry, AWS Glue Schema Registry, Karapace.

### Module 3.1: Schema Registry (`internal/schemaregistry`)

**Purpose:** Versioned schema storage with compatibility enforcement for event streams and CDC.

**New files:**

| File | Purpose |
|------|---------|
| `internal/schemaregistry/resource.go` | `SchemaResource`, `SchemaSubjectResource` |
| `internal/schemaregistry/reconciler.go` | `SchemaReconciler` — validates compatibility on registration |
| `internal/schemaregistry/compatibility.go` | Compatibility checking engine |
| `internal/schemaregistry/serializer.go` | Avro/JSON Schema/Protobuf parsing |
| `internal/schemaregistry/handlers.go` | REST API (Confluent-compatible wire format) |
| `internal/schemaregistry/cache.go` | In-memory schema cache for fast lookups |

**Resource definitions:**

```go
const (
    SchemaKind       = "Schema"
    SchemaAPIVersion = "schema.axiomnizam.io/v1"
)

type SchemaSpec struct {
    Subject        string            `json:"subject"`        // Topic or asset name
    SchemaType     SchemaType        `json:"schemaType"`     // avro, json, protobuf
    Schema         string            `json:"schema"`         // Schema definition (JSON string)
    References     []SchemaReference `json:"references"`     // Cross-schema references
    Compatibility  CompatibilityMode `json:"compatibility"`  // backward, forward, full, none
    Metadata       map[string]string `json:"metadata"`
    RuleSet        *SchemaRuleSet    `json:"ruleSet,omitempty"` // Validation rules
}

type SchemaResourceStatus struct {
    resources.ObjectStatus `json:",inline"`
    SchemaID       int64      `json:"schemaId"`       // Global unique ID
    Version        int        `json:"version"`        // Version within subject
    Fingerprint    string     `json:"fingerprint"`    // SHA-256 of normalized schema
    IsLatest       bool       `json:"isLatest"`
    RegisteredAt   *time.Time `json:"registeredAt"`
    CompatibleWith []int      `json:"compatibleWith"` // Compatible version numbers
}

type SchemaType string
const (
    SchemaTypeAvro     SchemaType = "AVRO"
    SchemaTypeJSON     SchemaType = "JSON"
    SchemaTypeProtobuf SchemaType = "PROTOBUF"
)
```

**Reconciler behavior:**

1. **Observe:** New SchemaResource created (version N)
2. **Diff:** Fetch previous version (N-1) from same subject
3. **Act:** Run compatibility check between N and N-1 based on subject's compatibility mode
4. **Update Status:** If compatible, assign schemaId and mark as latest. If incompatible, set condition `Compatible=False` with reason.

**Compatibility checking algorithm:**

```
For BACKWARD compatibility:
  - New schema can deserialize data written with old schema
  - Allowed: add optional field, widen type, add default
  - Forbidden: remove field, add required field without default, narrow type

For FORWARD compatibility:
  - Old schema can deserialize data written with new schema
  - Allowed: remove field, add required field
  - Forbidden: add optional field without default (old reader fails)

For FULL compatibility:
  - Both backward AND forward must pass
```

**API routes (Confluent-compatible):**

```
GET    /api/v1/schemas/subjects
GET    /api/v1/schemas/subjects/:subject/versions
GET    /api/v1/schemas/subjects/:subject/versions/:version
POST   /api/v1/schemas/subjects/:subject/versions          # Register new
POST   /api/v1/schemas/compatibility/subjects/:subject/versions/:version  # Check compat
GET    /api/v1/schemas/ids/:id
DELETE /api/v1/schemas/subjects/:subject/versions/:version
PUT    /api/v1/schemas/config/:subject                     # Set compatibility mode
GET    /api/v1/schemas/config/:subject
```

**Integration with existing modules:**

- CDC pipelines reference schema subjects for source/sink validation
- Event bus topics can require schema registration before publish
- Conductor producers validate message schema before send
- Catalog assets link to their schema versions

---

### Module 3.2: Schema Evolution Manager (`internal/schemaregistry/evolution`)

**Purpose:** Automated schema migration tooling for when breaking changes are necessary.

**Capabilities:**

- **Migration plans:** Generate step-by-step migration when breaking change is needed
- **Dual-write period:** Produce in both old and new schema during transition
- **Consumer readiness tracking:** Track which consumers have upgraded
- **Rollback support:** Revert to previous schema version if migration fails
- **Impact analysis:** Show all affected pipelines/consumers before migration

---

## WS-4: Observability and Alerting (P1 High — Week 3-5)

### Why This Is Important

AxiomNizam has metrics, structured logging, and health probes — but no **action layer**. When something goes wrong, the platform detects it but cannot notify anyone. An alerting engine closes the loop from detection to response.

**Competitive context:** PagerDuty, Grafana Alerting, Datadog Monitors, CloudWatch Alarms.

### Module 4.1: Alert Rules Engine (`internal/alerting`)

**Purpose:** Declarative alert rules that fire notifications based on metric thresholds, quality failures, or system events.

**New files:**

| File | Purpose |
|------|---------|
| `internal/alerting/resource.go` | `AlertRuleResource`, `AlertIncidentResource` |
| `internal/alerting/reconciler.go` | `AlertRuleReconciler` — evaluates rules on schedule |
| `internal/alerting/evaluator.go` | Condition evaluation engine |
| `internal/alerting/notifier.go` | Multi-channel notification dispatch |
| `internal/alerting/channels.go` | Slack, email, webhook, PagerDuty integrations |
| `internal/alerting/silencer.go` | Alert suppression and maintenance windows |
| `internal/alerting/handlers.go` | REST API |

**Resource definitions:**

```go
const (
    AlertRuleKind       = "AlertRule"
    AlertRuleAPIVersion = "alerting.axiomnizam.io/v1"
)

type AlertRuleSpec struct {
    DisplayName    string            `json:"displayName"`
    Description    string            `json:"description"`
    Severity       AlertSeverity     `json:"severity"`       // critical, warning, info
    
    // Evaluation
    EvalInterval   string            `json:"evalInterval"`   // "30s", "1m", "5m"
    Condition      AlertCondition    `json:"condition"`
    ForDuration    string            `json:"forDuration"`    // Must be true for X before firing
    
    // Notification
    Channels       []ChannelRef      `json:"channels"`       // Where to send
    Escalation     *EscalationPolicy `json:"escalation,omitempty"`
    
    // Suppression
    Silenced       bool              `json:"silenced"`
    SilenceUntil   *time.Time        `json:"silenceUntil,omitempty"`
    
    // Grouping
    Labels         map[string]string `json:"labels"`
    Annotations    map[string]string `json:"annotations"`
}

type AlertCondition struct {
    Type           ConditionType     `json:"type"`           // metric, quality, event, custom
    
    // Metric-based
    Metric         *MetricCondition  `json:"metric,omitempty"`
    
    // Quality-based (integrates with WS-2)
    Quality        *QualityCondition `json:"quality,omitempty"`
    
    // Event-based (integrates with event bus)
    Event          *EventCondition   `json:"event,omitempty"`
    
    // Reconciler health
    Reconciler     *ReconcilerCondition `json:"reconciler,omitempty"`
}

type MetricCondition struct {
    Query          string  `json:"query"`          // PromQL-style query
    Operator       string  `json:"operator"`       // gt, lt, eq, ne, gte, lte
    Threshold      float64 `json:"threshold"`
    AggregateOver  string  `json:"aggregateOver"`  // "5m", "15m", "1h"
    Aggregation    string  `json:"aggregation"`    // avg, max, min, sum, count
}

type EscalationPolicy struct {
    Levels         []EscalationLevel `json:"levels"`
}

type EscalationLevel struct {
    After          string            `json:"after"`          // "15m", "1h"
    Channels       []ChannelRef      `json:"channels"`
    RepeatInterval string            `json:"repeatInterval"` // "30m"
}
```

**Alert lifecycle (as reconciled resource):**

```
AlertRule created -> Reconciler evaluates condition on schedule
  -> Condition TRUE for forDuration -> AlertIncident created (status: firing)
    -> Notifier dispatches to channels
    -> If not acknowledged within escalation.levels[0].after -> escalate
  -> Condition FALSE -> AlertIncident resolved (status: resolved)
    -> Resolution notification sent
```

---

### Module 4.2: Notification Channels (`internal/alerting/channels`)

**Supported channels:**

| Channel | Protocol | Config |
|---------|----------|--------|
| Slack | Webhook | `webhookUrl`, `channel`, `mentionUsers` |
| Email | SMTP | `smtpHost`, `from`, `to[]`, `templateRef` |
| Webhook | HTTP POST | `url`, `headers`, `bodyTemplate` |
| PagerDuty | Events API v2 | `routingKey`, `severity` |
| Microsoft Teams | Webhook | `webhookUrl` |
| OpsGenie | API | `apiKey`, `responders` |

**Channel resource:**

```go
type NotificationChannelSpec struct {
    Type           ChannelType       `json:"type"`
    Config         map[string]string `json:"config"`
    RateLimitPerHour int             `json:"rateLimitPerHour"`
    Templates      map[string]string `json:"templates"`      // Custom message templates
}
```

---

### Module 4.3: SLO/SLA Tracking (`internal/slo`)

**Purpose:** Service Level Objectives as declarative resources with error budget tracking.

**New files:**

| File | Purpose |
|------|---------|
| `internal/slo/resource.go` | `SLOResource` |
| `internal/slo/reconciler.go` | `SLOReconciler` — calculates burn rate and budget |
| `internal/slo/calculator.go` | Error budget math |
| `internal/slo/handlers.go` | REST API |

**Resource definition:**

```go
type SLOSpec struct {
    DisplayName    string        `json:"displayName"`
    Description    string        `json:"description"`
    Service        string        `json:"service"`        // What service this SLO covers
    
    // Objective
    Target         float64       `json:"target"`         // 0.999 = 99.9%
    Window         string        `json:"window"`         // "30d", "7d"
    
    // Indicator
    Indicator      SLISpec       `json:"indicator"`
    
    // Alerting integration
    BurnRateAlerts []BurnRateAlert `json:"burnRateAlerts"`
}

type SLISpec struct {
    Type           string `json:"type"`           // availability, latency, quality, freshness
    GoodQuery      string `json:"goodQuery"`      // Metric query for good events
    TotalQuery     string `json:"totalQuery"`     // Metric query for total events
    ThresholdMs    int64  `json:"thresholdMs,omitempty"` // For latency SLIs
}

type SLOResourceStatus struct {
    resources.ObjectStatus `json:",inline"`
    CurrentSLI      float64    `json:"currentSli"`      // Current measured value
    ErrorBudget     float64    `json:"errorBudget"`     // Remaining budget (0-1)
    BudgetConsumed  float64    `json:"budgetConsumed"`  // How much used (0-1)
    BurnRate        float64    `json:"burnRate"`        // Current burn rate (1.0 = normal)
    IsBreaching     bool       `json:"isBreaching"`
    TimeToExhaust   string     `json:"timeToExhaust"`   // At current burn rate
    WindowStart     *time.Time `json:"windowStart"`
    WindowEnd       *time.Time `json:"windowEnd"`
}
```

---

### Module 4.4: Cost Attribution (`internal/costing`)

**Purpose:** Per-tenant, per-query, per-pipeline resource usage tracking and chargeback.

**New files:**

| File | Purpose |
|------|---------|
| `internal/costing/resource.go` | `CostPolicyResource`, `UsageRecordResource` |
| `internal/costing/reconciler.go` | `CostReconciler` — aggregates usage into billing periods |
| `internal/costing/tracker.go` | Real-time usage tracking middleware |
| `internal/costing/handlers.go` | REST API for usage reports |

**What gets tracked:**

| Dimension | Metric | Unit |
|-----------|--------|------|
| Query execution | CPU time, rows scanned, bytes read | credits/query |
| Pipeline runs | Duration, records processed, bytes moved | credits/run |
| Storage | Bytes stored, objects count | credits/GB/month |
| API calls | Request count, bandwidth | credits/1000 requests |
| CDC streams | Events captured, lag time | credits/1M events |

**Cost policy resource:**

```go
type CostPolicySpec struct {
    TenantID       string            `json:"tenantId"`
    BillingPeriod  string            `json:"billingPeriod"`  // monthly, weekly
    Quotas         map[string]Quota  `json:"quotas"`         // Per-dimension limits
    Alerts         []CostAlert       `json:"alerts"`         // Budget alerts
    RateCard       RateCard          `json:"rateCard"`       // Price per unit
}

type Quota struct {
    Limit          float64 `json:"limit"`
    Action         string  `json:"action"`         // warn, throttle, block
    CurrentUsage   float64 `json:"currentUsage"`
}
```

---

## WS-5: Federated Query and Virtualization (P1 High — Week 4-6)

### Why This Is Important

AxiomNizam already connects to 5+ SQL databases and MongoDB. The natural next step is letting users query across them in a single request. This is the **killer feature** that few platforms offer natively — most require separate ETL to a warehouse first.

**Competitive context:** Trino/Presto, Dremio, Denodo, Starburst, Apache Drill.

### Module 5.1: Federated Query Engine (`internal/federation`)

**Purpose:** Cross-database query execution with automatic query decomposition, parallel execution, and result merging.

**New files:**

| File | Purpose |
|------|---------|
| `internal/federation/resource.go` | `FederatedQueryResource`, `VirtualTableResource` |
| `internal/federation/reconciler.go` | `VirtualTableReconciler` — maintains virtual table metadata |
| `internal/federation/planner.go` | Query planner: decomposes cross-source queries |
| `internal/federation/executor.go` | Parallel execution engine |
| `internal/federation/merger.go` | Result set merging (join, union, sort) |
| `internal/federation/optimizer.go` | Cost-based optimization (push-down predicates) |
| `internal/federation/handlers.go` | REST API |
| `internal/federation/cache.go` | Result caching layer |

**Architecture:**

```
User Query (SQL)
    |
    v
[Query Parser] -- Parse SQL into AST
    |
    v
[Catalog Lookup] -- Resolve table references to datasources (via WS-1 catalog)
    |
    v
[Query Planner] -- Decompose into per-datasource sub-queries
    |              -- Push down WHERE, LIMIT, ORDER BY where possible
    |              -- Identify cross-source JOINs
    v
[Cost Optimizer] -- Choose join order based on table statistics
    |             -- Decide: push-down vs pull-up
    v
[Parallel Executor] -- Execute sub-queries concurrently against datasources
    |                -- Stream results as they arrive
    v
[Result Merger] -- Apply cross-source JOINs, UNIONs, aggregations
    |            -- Sort, limit, project final columns
    v
[Response] -- Stream results to client (JSON or Arrow format)
```

**Resource definitions:**

```go
const (
    VirtualTableKind       = "VirtualTable"
    VirtualTableAPIVersion = "federation.axiomnizam.io/v1"
)

type VirtualTableSpec struct {
    DisplayName    string              `json:"displayName"`
    Description    string              `json:"description"`
    Sources        []VirtualSource     `json:"sources"`        // Underlying real tables
    JoinConditions []JoinCondition     `json:"joinConditions"` // How sources relate
    Columns        []VirtualColumn     `json:"columns"`        // Exposed columns
    Filters        []DefaultFilter     `json:"filters"`        // Always-applied filters
    CachePolicy    *CachePolicy        `json:"cachePolicy,omitempty"`
    RefreshSchedule string             `json:"refreshSchedule,omitempty"` // For materialized
    Materialized   bool                `json:"materialized"`   // Cache results in local store
}

type VirtualSource struct {
    Alias          string `json:"alias"`
    DataSourceRef  string `json:"dataSourceRef"`  // Reference to DataSource resource
    Database       string `json:"database"`
    Schema         string `json:"schema"`
    Table          string `json:"table"`
}

type JoinCondition struct {
    LeftAlias      string `json:"leftAlias"`
    LeftColumn     string `json:"leftColumn"`
    RightAlias     string `json:"rightAlias"`
    RightColumn    string `json:"rightColumn"`
    JoinType       string `json:"joinType"`       // inner, left, right, full
}
```

**Query API:**

```
POST /api/v1/federation/query
{
    "sql": "SELECT c.name, o.total FROM postgres.customers c JOIN mysql.orders o ON c.id = o.customer_id WHERE o.total > 1000",
    "timeout": "30s",
    "maxRows": 10000,
    "format": "json"  // json, csv, arrow
}
```

**Optimization strategies:**

1. **Predicate push-down:** WHERE clauses pushed to source databases
2. **Projection push-down:** Only SELECT needed columns from sources
3. **Limit push-down:** LIMIT pushed when no cross-source operations needed
4. **Join reordering:** Smaller table scanned first based on catalog statistics
5. **Parallel execution:** Independent sub-queries run concurrently
6. **Result streaming:** Large results streamed rather than buffered

---

### Module 5.2: Materialized Views (`internal/federation/materialized`)

**Purpose:** Pre-computed cross-source query results that refresh on schedule.

**Capabilities:**

- Define a federated query as a materialized view
- Refresh on cron schedule or on-demand
- Incremental refresh when source supports CDC
- Serve queries from local cache (PostgreSQL or in-memory)
- Automatic invalidation when source schema changes

**Resource:**

```go
type MaterializedViewSpec struct {
    Query          string        `json:"query"`          // Federated SQL
    RefreshSchedule string      `json:"refreshSchedule"` // Cron
    RefreshMode    string        `json:"refreshMode"`    // full, incremental
    StorageBackend string        `json:"storageBackend"` // postgres, memory
    TTL            string        `json:"ttl"`            // Max staleness
    Indexes        []IndexDef    `json:"indexes"`        // Indexes on materialized data
}
```

---

### Module 5.3: Query Performance Profiler (`internal/federation/profiler`)

**Purpose:** Query execution analysis with optimization recommendations.

**Capabilities:**

- Execution plan visualization (which datasources, join strategies, row estimates)
- Slow query detection and logging
- Index recommendations based on query patterns
- Query rewrite suggestions
- Historical query performance trends

**API:**

```
POST /api/v1/federation/explain
{
    "sql": "SELECT ...",
    "analyze": true  // Actually execute and measure
}

Response:
{
    "plan": {
        "type": "merge_join",
        "estimatedRows": 1500,
        "estimatedCost": 0.45,
        "children": [
            {"type": "remote_scan", "datasource": "postgres", "table": "customers", "rows": 50000},
            {"type": "remote_scan", "datasource": "mysql", "table": "orders", "rows": 120000}
        ]
    },
    "recommendations": [
        "Add index on mysql.orders.customer_id for 3x faster join",
        "Consider materializing this query (executed 47 times in last 24h)"
    ],
    "actualDuration": "234ms"
}
```

---

## WS-6: Governance and Compliance (P1 High — Week 5-6)

### Why This Is Important

Enterprise customers require GDPR, HIPAA, SOC2, and PCI-DSS compliance. AxiomNizam has RBAC, encryption, and audit — but lacks automated compliance policies, data retention enforcement, and right-to-erasure workflows. Governance turns security features into **compliance automation**.

**Competitive context:** Collibra, Alation Governance, Immuta, Privacera, Apache Ranger.

### Module 6.1: Compliance Policies (`internal/governance`)

**Purpose:** Declarative compliance policies that automatically enforce data handling rules.

**New files:**

| File | Purpose |
|------|---------|
| `internal/governance/resource.go` | `CompliancePolicyResource`, `RetentionPolicyResource`, `AccessRequestResource` |
| `internal/governance/reconciler.go` | `CompliancePolicyReconciler`, `RetentionReconciler` |
| `internal/governance/enforcer.go` | Policy enforcement engine |
| `internal/governance/erasure.go` | Right-to-erasure workflow (GDPR Article 17) |
| `internal/governance/classification.go` | Auto-classification rules |
| `internal/governance/handlers.go` | REST API |
| `internal/governance/reports.go` | Compliance report generation |

**Resource definitions:**

```go
const (
    CompliancePolicyKind       = "CompliancePolicy"
    CompliancePolicyAPIVersion = "governance.axiomnizam.io/v1"
)

type CompliancePolicySpec struct {
    Framework      string              `json:"framework"`      // gdpr, hipaa, soc2, pci_dss, custom
    DisplayName    string              `json:"displayName"`
    Description    string              `json:"description"`
    Scope          PolicyScope         `json:"scope"`          // What assets this applies to
    Rules          []ComplianceRule    `json:"rules"`
    Enforcement    EnforcementMode     `json:"enforcement"`    // audit, warn, block
    ReviewSchedule string              `json:"reviewSchedule"` // How often to re-evaluate
    Owner          string              `json:"owner"`
    Approvers      []string            `json:"approvers"`
}

type PolicyScope struct {
    Domains        []string `json:"domains"`        // Business domains
    Classifications []string `json:"classifications"` // PII, PHI, Financial
    DataSources    []string `json:"dataSources"`    // Specific datasources
    Tags           []string `json:"tags"`
    AllAssets      bool     `json:"allAssets"`      // Apply to everything
}

type ComplianceRule struct {
    ID             string          `json:"id"`
    Name           string          `json:"name"`
    Description    string          `json:"description"`
    Type           RuleType        `json:"type"`           // retention, access, encryption, masking, audit, location
    
    Retention      *RetentionRule  `json:"retention,omitempty"`
    Access         *AccessRule     `json:"access,omitempty"`
    Encryption     *EncryptionRule `json:"encryption,omitempty"`
    Masking        *MaskingRule    `json:"masking,omitempty"`
    Location       *LocationRule   `json:"location,omitempty"`
}

type RetentionRule struct {
    MaxRetentionDays int    `json:"maxRetentionDays"`
    MinRetentionDays int    `json:"minRetentionDays"` // Legal hold minimum
    Action           string `json:"action"`           // archive, delete, anonymize
    GracePeriodDays  int    `json:"gracePeriodDays"`
}

type MaskingRule struct {
    Columns        []string `json:"columns"`        // Column patterns to mask
    MaskType       string   `json:"maskType"`       // hash, redact, partial, tokenize, noise
    ExemptRoles    []string `json:"exemptRoles"`    // Roles that see unmasked data
}
```

**Compliance frameworks supported:**

| Framework | Key Rules |
|-----------|-----------|
| GDPR | Right to erasure, data minimization, consent tracking, 72h breach notification, DPO assignment |
| HIPAA | PHI encryption at rest/transit, access logging, minimum necessary, BAA tracking |
| SOC2 | Access controls, change management, availability monitoring, incident response |
| PCI-DSS | Cardholder data encryption, access restriction, audit trails, network segmentation |
| Custom | User-defined rules with any combination |

---

### Module 6.2: Data Retention Engine (`internal/governance/retention`)

**Purpose:** Automated data lifecycle management — archive, delete, or anonymize data based on retention policies.

**Reconciler behavior:**

1. **Observe:** Read RetentionPolicyResource, scan catalog for assets matching scope
2. **Diff:** For each asset, check data age against maxRetentionDays
3. **Act:** Execute retention action (archive to cold storage, delete, or anonymize)
4. **Update Status:** Record what was purged, when, compliance proof

**Retention actions:**

| Action | Behavior |
|--------|----------|
| `delete` | Hard delete rows older than retention period |
| `archive` | Move to object storage (cold tier), remove from hot database |
| `anonymize` | Replace PII columns with hashed/randomized values, keep structure |
| `aggregate` | Replace granular rows with aggregated summaries |

**Right-to-erasure workflow (GDPR):**

```
1. ErasureRequest created (subject: "user@example.com")
2. Reconciler scans ALL catalog assets for subject's data
3. For each asset: execute deletion/anonymization per policy
4. Generate compliance certificate with:
   - Assets scanned
   - Records deleted/anonymized
   - Timestamp of completion
   - Cryptographic proof (hash chain)
5. Notify requestor of completion
```

---

### Module 6.3: Access Governance (`internal/governance/access`)

**Purpose:** Data access request workflows with approval chains and time-bound access.

**Capabilities:**

- Self-service access requests (user requests access to dataset)
- Multi-level approval workflows (manager -> data owner -> compliance)
- Time-bound access grants (auto-revoke after N days)
- Access certification campaigns (periodic review of who has access to what)
- Segregation of duties enforcement
- Access audit trail with justification

**Resource:**

```go
type AccessRequestSpec struct {
    Requestor      string        `json:"requestor"`
    AssetRef       string        `json:"assetRef"`       // What they want access to
    AccessLevel    string        `json:"accessLevel"`    // read, write, admin
    Justification  string        `json:"justification"`  // Why they need it
    Duration       string        `json:"duration"`       // "30d", "90d", "permanent"
    Approvers      []string      `json:"approvers"`      // Required approvers
}

type AccessRequestResourceStatus struct {
    resources.ObjectStatus `json:",inline"`
    ApprovalStatus string      `json:"approvalStatus"` // pending, approved, denied, expired
    ApprovedBy     []string    `json:"approvedBy"`
    DeniedBy       string      `json:"deniedBy,omitempty"`
    GrantedAt      *time.Time  `json:"grantedAt,omitempty"`
    ExpiresAt      *time.Time  `json:"expiresAt,omitempty"`
    RevokedAt      *time.Time  `json:"revokedAt,omitempty"`
    RevokeReason   string      `json:"revokeReason,omitempty"`
}
```

---

## WS-7: Advanced Analytics and ML (P2 Medium — Week 6-8)

### Why This Is Important

Modern data platforms are expected to support ML workflows natively. AxiomNizam has VectorPlus for similarity search and OpenClaw for SQL assistance — extending to a full feature store and ML pipeline orchestration positions it as a complete AI-ready data platform.

**Competitive context:** Databricks MLflow, SageMaker Feature Store, Feast, Tecton, Vertex AI.

### Module 7.1: Feature Store (`internal/featurestore`)

**Purpose:** Centralized feature engineering, versioning, and serving for ML models.

**New files:**

| File | Purpose |
|------|---------|
| `internal/featurestore/resource.go` | `FeatureGroupResource`, `FeatureViewResource` |
| `internal/featurestore/reconciler.go` | `FeatureGroupReconciler` — materializes features |
| `internal/featurestore/online.go` | Online serving (low-latency point lookups) |
| `internal/featurestore/offline.go` | Offline serving (batch for training) |
| `internal/featurestore/registry.go` | Feature metadata registry |
| `internal/featurestore/handlers.go` | REST API |

**Resource definitions:**

```go
const (
    FeatureGroupKind       = "FeatureGroup"
    FeatureGroupAPIVersion = "ml.axiomnizam.io/v1"
)

type FeatureGroupSpec struct {
    DisplayName    string            `json:"displayName"`
    Description    string            `json:"description"`
    Entity         string            `json:"entity"`         // Primary entity (user, product, etc.)
    EntityKey      []string          `json:"entityKey"`      // Key columns
    Features       []FeatureSpec     `json:"features"`
    Source         FeatureSource     `json:"source"`         // Where raw data comes from
    Schedule       string            `json:"schedule"`       // Materialization schedule
    TTL            string            `json:"ttl"`            // Feature freshness requirement
    OnlineStore    *OnlineStoreConfig `json:"onlineStore,omitempty"`
    OfflineStore   *OfflineStoreConfig `json:"offlineStore,omitempty"`
    Tags           []string          `json:"tags"`
}

type FeatureSpec struct {
    Name           string `json:"name"`
    Type           string `json:"type"`           // int64, float64, string, bool, embedding
    Description    string `json:"description"`
    Transform      string `json:"transform"`      // SQL expression or function
    DefaultValue   string `json:"defaultValue,omitempty"`
    Validator      string `json:"validator,omitempty"` // Validation expression
}

type FeatureSource struct {
    Type           string `json:"type"`           // sql, stream, request
    DataSourceRef  string `json:"dataSourceRef"`
    Query          string `json:"query,omitempty"`
    StreamRef      string `json:"streamRef,omitempty"` // CDC/Kafka topic
}

type OnlineStoreConfig struct {
    Backend        string `json:"backend"`        // redis, postgres, memory
    TTL            string `json:"ttl"`
    MaxEntities    int64  `json:"maxEntities"`
}
```

**Feature serving API:**

```
POST /api/v1/features/online
{
    "featureGroup": "user-features",
    "entities": [{"user_id": "123"}, {"user_id": "456"}],
    "features": ["purchase_count_30d", "avg_order_value", "churn_score"]
}

Response:
{
    "results": [
        {"user_id": "123", "purchase_count_30d": 7, "avg_order_value": 45.20, "churn_score": 0.12},
        {"user_id": "456", "purchase_count_30d": 2, "avg_order_value": 120.00, "churn_score": 0.67}
    ],
    "metadata": {"freshness": "2m", "featureGroup": "user-features", "version": 3}
}
```

---

### Module 7.2: Streaming Analytics (`internal/streamanalytics`)

**Purpose:** Real-time windowed aggregations over Kafka/event bus streams.

**New files:**

| File | Purpose |
|------|---------|
| `internal/streamanalytics/resource.go` | `StreamJobResource` |
| `internal/streamanalytics/reconciler.go` | `StreamJobReconciler` — manages streaming job lifecycle |
| `internal/streamanalytics/window.go` | Windowing functions (tumbling, sliding, session) |
| `internal/streamanalytics/aggregator.go` | Aggregation engine |
| `internal/streamanalytics/sink.go` | Output sinks (database, webhook, event bus) |
| `internal/streamanalytics/handlers.go` | REST API |

**Resource definition:**

```go
type StreamJobSpec struct {
    DisplayName    string            `json:"displayName"`
    Source         StreamSource      `json:"source"`         // Kafka topic or event bus
    Window         WindowSpec        `json:"window"`
    Aggregations   []AggregationSpec `json:"aggregations"`
    Filters        []FilterSpec      `json:"filters"`
    GroupBy        []string          `json:"groupBy"`
    Sink           StreamSink        `json:"sink"`           // Where to write results
    Parallelism    int               `json:"parallelism"`
    Watermark      string            `json:"watermark"`      // Late event tolerance
}

type WindowSpec struct {
    Type           string `json:"type"`           // tumbling, sliding, session
    Size           string `json:"size"`           // "5m", "1h"
    Slide          string `json:"slide,omitempty"` // For sliding windows
    Gap            string `json:"gap,omitempty"`   // For session windows
}

type AggregationSpec struct {
    OutputField    string `json:"outputField"`
    Function       string `json:"function"`       // count, sum, avg, min, max, p50, p95, p99, distinct_count
    InputField     string `json:"inputField"`
}
```

**Example: Real-time API metrics aggregation:**

```yaml
apiVersion: streamanalytics.axiomnizam.io/v1
kind: StreamJob
metadata:
  name: api-metrics-5min
spec:
  source:
    type: eventbus
    topic: api.requests
  window:
    type: tumbling
    size: 5m
  groupBy: ["endpoint", "method", "status_code"]
  aggregations:
    - outputField: request_count
      function: count
      inputField: "*"
    - outputField: avg_latency_ms
      function: avg
      inputField: latency_ms
    - outputField: p99_latency_ms
      function: p99
      inputField: latency_ms
    - outputField: error_rate
      function: avg
      inputField: is_error
  sink:
    type: postgres
    table: api_metrics_5min
    dataSourceRef: analytics-db
```

---

### Module 7.3: Data Anonymization (`internal/anonymization`)

**Purpose:** PII masking, synthetic data generation, and privacy-preserving transformations.

**New files:**

| File | Purpose |
|------|---------|
| `internal/anonymization/resource.go` | `AnonymizationPolicyResource` |
| `internal/anonymization/reconciler.go` | `AnonymizationReconciler` |
| `internal/anonymization/masker.go` | Masking functions |
| `internal/anonymization/synthetic.go` | Synthetic data generator |
| `internal/anonymization/handlers.go` | REST API |

**Masking techniques:**

| Technique | Example | Use Case |
|-----------|---------|----------|
| `hash` | `john@email.com` -> `a3f2b8c1...` | Consistent pseudonymization |
| `redact` | `john@email.com` -> `[REDACTED]` | Full removal |
| `partial` | `john@email.com` -> `j***@e***.com` | Partial visibility |
| `tokenize` | `john@email.com` -> `TOK_8f3a2b` | Reversible with key |
| `noise` | `salary: 75000` -> `salary: 73200` | Statistical privacy |
| `generalize` | `age: 34` -> `age: 30-40` | K-anonymity |
| `synthetic` | `John Smith` -> `Alice Johnson` | Realistic fake data |
| `shuffle` | Column values shuffled across rows | Break correlations |

**Resource:**

```go
type AnonymizationPolicySpec struct {
    Scope          PolicyScope       `json:"scope"`          // Which assets
    Rules          []AnonymRule      `json:"rules"`
    OutputMode     string            `json:"outputMode"`     // in_place, copy, view
    OutputTarget   string            `json:"outputTarget,omitempty"` // Target datasource for copy
    Schedule       string            `json:"schedule,omitempty"`     // For periodic anonymization
    PreserveStats  bool              `json:"preserveStats"`  // Maintain statistical properties
}

type AnonymRule struct {
    ColumnPattern  string `json:"columnPattern"`  // Regex or exact column name
    Classification string `json:"classification"` // PII, PHI, Financial
    Technique      string `json:"technique"`      // hash, redact, partial, etc.
    Config         map[string]string `json:"config,omitempty"`
}
```

---

### Module 7.4: ML Pipeline Orchestration (`internal/mlpipeline`)

**Purpose:** Model training, evaluation, deployment, and A/B testing as declarative resources.

**Resource definitions:**

```go
type MLPipelineSpec struct {
    DisplayName    string            `json:"displayName"`
    Steps          []MLStep          `json:"steps"`
    Schedule       string            `json:"schedule"`
    FeatureGroups  []string          `json:"featureGroups"`  // Input features
    ModelRegistry  string            `json:"modelRegistry"`  // Where to store models
    Notifications  []string          `json:"notifications"`  // Alert on completion/failure
}

type MLStep struct {
    Name           string            `json:"name"`
    Type           string            `json:"type"`           // data_prep, train, evaluate, deploy, ab_test
    Config         map[string]interface{} `json:"config"`
    DependsOn      []string          `json:"dependsOn"`
    Timeout        string            `json:"timeout"`
}

type ModelDeploymentSpec struct {
    ModelRef       string            `json:"modelRef"`       // Model artifact reference
    Version        string            `json:"version"`
    Endpoint       string            `json:"endpoint"`       // Serving endpoint path
    Replicas       int               `json:"replicas"`
    TrafficSplit   map[string]int    `json:"trafficSplit"`   // version -> percentage (A/B)
    AutoScale      *AutoScaleConfig  `json:"autoScale,omitempty"`
    Canary         *CanaryConfig     `json:"canary,omitempty"`
}
```

---

## Implementation Schedule

### Phase Dependency Graph

```
WS-1 (Catalog) ─────────────────────────────────────────────────────┐
    |                                                                 |
    ├──> WS-2 (Quality) ──> depends on catalog for asset references  |
    |        |                                                        |
    |        └──> WS-4.1 (Alerting) ──> quality failures trigger     |
    |                                     alerts                      |
    ├──> WS-5 (Federation) ──> catalog resolves table references     |
    |                                                                 |
    ├──> WS-6 (Governance) ──> catalog provides asset inventory      |
    |                                                                 |
    └──> WS-7 (ML) ──> feature store reads from catalog              |
                                                                      |
WS-3 (Schema Registry) ──> independent, integrates with CDC/EventBus |
                                                                      |
WS-4.3 (SLO) ──> independent, uses existing metrics                  |
WS-4.4 (Costing) ──> independent, uses existing query logger         |
```

### Week-by-Week Schedule

| Week | Primary | Secondary | Deliverable |
|------|---------|-----------|-------------|
| 1 | WS-1.1 Catalog Core | WS-3.1 Schema Registry | Catalog resource + reconciler + scanner; Schema resource + compatibility engine |
| 2 | WS-1.2 Metadata Enrichment | WS-3.2 Schema Evolution | Auto-profiling + PII detection; Migration planner |
| 3 | WS-2.1 Quality Rules | WS-1.3 Catalog UI | 15 built-in check types + reconciler; Catalog dashboard |
| 4 | WS-2.2 Data Contracts | WS-4.1 Alert Rules | Contract validation + breaking change detection; Alert reconciler + notifier |
| 5 | WS-5.1 Federated Query | WS-4.2 Notification Channels | Query planner + parallel executor; Slack/email/webhook channels |
| 6 | WS-5.2 Materialized Views | WS-6.1 Compliance Policies | Incremental refresh; GDPR/HIPAA/SOC2 rules |
| 7 | WS-6.2 Retention Engine | WS-4.3 SLO Tracking | Auto-purge + right-to-erasure; Error budget calculator |
| 8 | WS-7.1 Feature Store | WS-6.3 Access Governance | Online/offline serving; Approval workflows |
| 9 | WS-7.2 Streaming Analytics | WS-4.4 Cost Attribution | Windowed aggregations; Usage tracking |
| 10 | WS-7.3 Anonymization | WS-5.3 Query Profiler | Masking engine; EXPLAIN + recommendations |
| 11 | WS-7.4 ML Pipelines | WS-2.3 Quality Dashboard | Training orchestration; Quality UI |
| 12 | Integration testing | Documentation | End-to-end validation; API docs + CLI docs |

---

## New File Inventory (Complete)

### By Module

| Module | New Files | Estimated Lines |
|--------|-----------|-----------------|
| `internal/catalog/` | 6 files | 2,500 |
| `internal/catalog/enrichment/` | 3 files | 1,200 |
| `internal/quality/rules/` | 6 files | 2,800 |
| `internal/contracts/` | 4 files | 1,500 |
| `internal/schemaregistry/` | 6 files | 2,200 |
| `internal/schemaregistry/evolution/` | 3 files | 1,000 |
| `internal/alerting/` | 7 files | 3,000 |
| `internal/slo/` | 4 files | 1,500 |
| `internal/costing/` | 4 files | 1,800 |
| `internal/federation/` | 8 files | 4,000 |
| `internal/federation/materialized/` | 3 files | 1,200 |
| `internal/federation/profiler/` | 3 files | 1,000 |
| `internal/governance/` | 7 files | 3,000 |
| `internal/governance/retention/` | 3 files | 1,200 |
| `internal/governance/access/` | 3 files | 1,200 |
| `internal/featurestore/` | 6 files | 2,500 |
| `internal/streamanalytics/` | 6 files | 2,500 |
| `internal/anonymization/` | 5 files | 2,000 |
| `internal/mlpipeline/` | 5 files | 2,000 |
| Frontend dashboards | 9 files | 4,500 |
| **TOTAL** | **101 files** | **~41,600 lines** |

### New etcd Prefixes

```
/axiomnizam/catalogassets/
/axiomnizam/catalogcollections/
/axiomnizam/qualityrules/
/axiomnizam/qualitychecks/
/axiomnizam/datacontracts/
/axiomnizam/schemas/
/axiomnizam/schemasubjects/
/axiomnizam/alertrules/
/axiomnizam/alertincidents/
/axiomnizam/notificationchannels/
/axiomnizam/slos/
/axiomnizam/costpolicies/
/axiomnizam/usagerecords/
/axiomnizam/virtualtables/
/axiomnizam/materializedviews/
/axiomnizam/compliancepolicies/
/axiomnizam/retentionpolicies/
/axiomnizam/accessrequests/
/axiomnizam/featuregroups/
/axiomnizam/featureviews/
/axiomnizam/streamjobs/
/axiomnizam/anonymizationpolicies/
/axiomnizam/mlpipelines/
/axiomnizam/modeldeployments/
```

### New GenericControllers (wired in main.go)

```go
// WS-1: Catalog
genericctrl.NewGenericController("catalog-asset", catalogAssetStore, catalogAssetReconciler, 2)
genericctrl.NewGenericController("catalog-enrichment", catalogAssetStore, enrichmentReconciler, 1)

// WS-2: Quality
genericctrl.NewGenericController("quality-rule", qualityRuleStore, qualityRuleReconciler, 2)
genericctrl.NewGenericController("data-contract", contractStore, contractReconciler, 1)

// WS-3: Schema Registry
genericctrl.NewGenericController("schema", schemaStore, schemaReconciler, 2)

// WS-4: Alerting
genericctrl.NewGenericController("alert-rule", alertRuleStore, alertRuleReconciler, 2)
genericctrl.NewGenericController("alert-incident", alertIncidentStore, alertIncidentReconciler, 1)
genericctrl.NewGenericController("slo", sloStore, sloReconciler, 1)
genericctrl.NewGenericController("cost-policy", costPolicyStore, costReconciler, 1)

// WS-5: Federation
genericctrl.NewGenericController("virtual-table", virtualTableStore, virtualTableReconciler, 1)
genericctrl.NewGenericController("materialized-view", matViewStore, matViewReconciler, 1)

// WS-6: Governance
genericctrl.NewGenericController("compliance-policy", compliancePolicyStore, complianceReconciler, 1)
genericctrl.NewGenericController("retention-policy", retentionStore, retentionReconciler, 1)
genericctrl.NewGenericController("access-request", accessRequestStore, accessRequestReconciler, 1)

// WS-7: ML
genericctrl.NewGenericController("feature-group", featureGroupStore, featureGroupReconciler, 2)
genericctrl.NewGenericController("stream-job", streamJobStore, streamJobReconciler, 2)
genericctrl.NewGenericController("anonymization-policy", anonPolicyStore, anonReconciler, 1)
genericctrl.NewGenericController("ml-pipeline", mlPipelineStore, mlPipelineReconciler, 2)
genericctrl.NewGenericController("model-deployment", modelDeployStore, modelDeployReconciler, 1)
```

**New controller count:** 20
**Total controllers after completion:** 53 (33 existing + 20 new)

---

## New CLI Commands

```
axiomnizamctl
|
+-- catalog
|   +-- list [--domain] [--type] [--owner]
|   +-- get <name>
|   +-- search <query>
|   +-- scan <datasource>
|   +-- describe <name>
|   +-- apply -f catalog-asset.yaml
|
+-- quality
|   +-- rule list
|   +-- rule get <name>
|   +-- rule apply -f rule.yaml
|   +-- rule run <name>              # Manual trigger
|   +-- check list [--asset] [--status]
|   +-- check get <name>
|   +-- score <asset-name>           # Show quality score
|
+-- contract
|   +-- list
|   +-- get <name>
|   +-- apply -f contract.yaml
|   +-- validate <name>              # Check current compliance
|   +-- diff <name>                  # Show schema drift
|
+-- schema
|   +-- list [--subject]
|   +-- get <subject> [--version]
|   +-- register -f schema.json --subject <name>
|   +-- compatibility <subject> -f new-schema.json
|   +-- delete <subject> --version <n>
|
+-- alert
|   +-- rule list
|   +-- rule get <name>
|   +-- rule apply -f alert-rule.yaml
|   +-- rule silence <name> --duration 2h
|   +-- incident list [--status firing|resolved]
|   +-- incident ack <name>
|   +-- incident resolve <name>
|   +-- channel list
|   +-- channel apply -f channel.yaml
|   +-- channel test <name>          # Send test notification
|
+-- slo
|   +-- list
|   +-- get <name>
|   +-- apply -f slo.yaml
|   +-- status                       # Show all SLO statuses
|   +-- budget <name>                # Show error budget
|
+-- federation
|   +-- query "SELECT ..."           # Execute federated query
|   +-- explain "SELECT ..."         # Show execution plan
|   +-- virtual-table list
|   +-- virtual-table apply -f vt.yaml
|   +-- materialized-view list
|   +-- materialized-view refresh <name>
|
+-- governance
|   +-- policy list
|   +-- policy apply -f policy.yaml
|   +-- policy audit <name>          # Run compliance audit
|   +-- retention list
|   +-- retention apply -f retention.yaml
|   +-- erasure request --subject "user@email.com"
|   +-- erasure status <request-id>
|   +-- access request --asset <name> --level read --justification "..."
|   +-- access list [--status pending|approved]
|   +-- access approve <request-id>
|   +-- access deny <request-id> --reason "..."
|
+-- feature
|   +-- group list
|   +-- group get <name>
|   +-- group apply -f feature-group.yaml
|   +-- serve --group <name> --entity '{"user_id":"123"}'
|
+-- cost
|   +-- usage [--tenant] [--period]
|   +-- quota list
|   +-- report [--format csv|json]
```

---

## New API Routes Summary

| Workstream | Prefix | Route Count |
|-----------|--------|-------------|
| WS-1 Catalog | `/api/v1/catalog/*` | 8 |
| WS-2 Quality | `/api/v1/quality/*` | 10 |
| WS-2 Contracts | `/api/v1/contracts/*` | 6 |
| WS-3 Schema | `/api/v1/schemas/*` | 9 |
| WS-4 Alerting | `/api/v1/alerts/*` | 12 |
| WS-4 SLO | `/api/v1/slos/*` | 6 |
| WS-4 Costing | `/api/v1/costs/*` | 5 |
| WS-5 Federation | `/api/v1/federation/*` | 8 |
| WS-6 Governance | `/api/v1/governance/*` | 12 |
| WS-7 Features | `/api/v1/features/*` | 8 |
| WS-7 Streaming | `/api/v1/stream-analytics/*` | 6 |
| WS-7 Anonymization | `/api/v1/anonymization/*` | 5 |
| WS-7 ML | `/api/v1/ml/*` | 8 |
| **TOTAL** | | **103 new routes** |

---

## New Frontend Dashboards

| Dashboard | File | Features |
|-----------|------|----------|
| Catalog Browser | `catalog-dashboard.html/js/css` | Search, browse, asset detail, domain tree |
| Quality Monitor | `quality-dashboard.html/js/css` | Heatmap, timeline, SLA board, drill-down |
| Alert Center | `alert-dashboard.html/js/css` | Active incidents, history, silence management |
| SLO Dashboard | `slo-dashboard.html/js/css` | Error budgets, burn rates, compliance |
| Federation Console | `federation-dashboard.html/js/css` | Query editor, explain visualizer, virtual tables |
| Governance Center | `governance-dashboard.html/js/css` | Compliance status, retention, access requests |
| Feature Store | `featurestore-dashboard.html/js/css` | Feature groups, serving metrics, freshness |
| Cost Center | `cost-dashboard.html/js/css` | Usage charts, quotas, tenant breakdown |
| ML Operations | `mlops-dashboard.html/js/css` | Pipelines, models, A/B tests, metrics |

**Total new dashboards:** 9
**Total dashboards after completion:** 21 (12 existing + 9 new)

---

## Environment Variables (New)

```bash
# WS-1: Catalog
RECONCILER_ENABLED_CATALOG=true
CATALOG_SCAN_INTERVAL=1h
CATALOG_MAX_COLUMNS_PER_TABLE=1000
CATALOG_SEARCH_INDEX_PATH=/data/catalog_index
CATALOG_PII_DETECTION_ENABLED=true

# WS-2: Quality
RECONCILER_ENABLED_QUALITY_RULES=true
QUALITY_CHECK_TIMEOUT=60s
QUALITY_MAX_CONCURRENT_CHECKS=10
QUALITY_HISTORY_RETENTION_DAYS=90

# WS-3: Schema Registry
RECONCILER_ENABLED_SCHEMA=true
SCHEMA_COMPATIBILITY_DEFAULT=backward
SCHEMA_MAX_VERSIONS_PER_SUBJECT=100

# WS-4: Alerting
RECONCILER_ENABLED_ALERTING=true
ALERT_EVAL_INTERVAL=30s
ALERT_NOTIFICATION_RETRY_COUNT=3
ALERT_DEDUP_INTERVAL=5m
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
SLACK_WEBHOOK_URL=
PAGERDUTY_ROUTING_KEY=

# WS-4: SLO
RECONCILER_ENABLED_SLO=true
SLO_EVAL_INTERVAL=1m

# WS-4: Costing
RECONCILER_ENABLED_COSTING=true
COST_AGGREGATION_INTERVAL=5m

# WS-5: Federation
RECONCILER_ENABLED_FEDERATION=true
FEDERATION_QUERY_TIMEOUT=30s
FEDERATION_MAX_ROWS=100000
FEDERATION_CACHE_TTL=5m
FEDERATION_MAX_CONCURRENT_QUERIES=20

# WS-6: Governance
RECONCILER_ENABLED_GOVERNANCE=true
RETENTION_CHECK_INTERVAL=24h
ERASURE_REQUEST_TIMEOUT=72h
ACCESS_REQUEST_AUTO_EXPIRE_DAYS=90

# WS-7: Feature Store
RECONCILER_ENABLED_FEATURESTORE=true
FEATURE_ONLINE_STORE_BACKEND=redis
FEATURE_OFFLINE_STORE_BACKEND=postgres
FEATURE_MATERIALIZATION_INTERVAL=1h

# WS-7: Streaming Analytics
RECONCILER_ENABLED_STREAM_ANALYTICS=true
STREAM_ANALYTICS_CHECKPOINT_INTERVAL=30s

# WS-7: Anonymization
RECONCILER_ENABLED_ANONYMIZATION=true
ANONYMIZATION_BATCH_SIZE=10000

# WS-7: ML Pipelines
RECONCILER_ENABLED_ML_PIPELINE=true
ML_MODEL_STORAGE_PATH=/data/models
ML_SERVING_PORT=8002
```

---

## Docker Compose Additions

```yaml
# Add to docker-compose.yml services:

  # Redis/Valkey for feature store online serving and caching
  valkey:
    image: valkey/valkey:latest
    container_name: valkey
    ports:
      - "6379:6379"
    networks:
      - app-network
    volumes:
      - valkey-data:/data
    command: valkey-server --appendonly yes --dir /data
    healthcheck:
      test: ["CMD", "valkey-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  # MinIO for model artifact storage (optional, can use internal storage)
  # minio:
  #   image: minio/minio:latest
  #   container_name: minio
  #   ports:
  #     - "9000:9000"
  #     - "9001:9001"
  #   environment:
  #     MINIO_ROOT_USER: minioadmin
  #     MINIO_ROOT_PASSWORD: minioadmin
  #   command: server /data --console-address ":9001"
  #   networks:
  #     - app-network
  #   volumes:
  #     - minio-data:/data
```

---

## Testing Strategy

### Per-Module Test Requirements

Each new module requires:

1. **Unit tests** for reconciler logic (mock store + mock manager)
2. **Integration tests** for API handlers (httptest + real etcd)
3. **Reconciler contract tests** (idempotency, requeue behavior, error handling)

### Test file naming:

```
internal/<module>/reconciler_test.go
internal/<module>/handlers_test.go
internal/<module>/<specific>_test.go
```

### Key integration test scenarios:

| Scenario | Modules Involved |
|----------|-----------------|
| Catalog scan discovers table, quality rule auto-created | catalog + quality |
| Schema registration fails compatibility, alert fires | schema + alerting |
| CDC pipeline schema drift detected, contract violation | cdc + contracts + alerting |
| Retention policy purges old data, audit trail created | governance + audit |
| Feature materialization uses federated query | featurestore + federation |
| Cost quota exceeded, tenant throttled | costing + ratelimit |
| Erasure request cascades through all assets | governance + catalog + anonymization |

---

## Success Metrics

### Platform Completeness Score

| Category | Current | After WS-1-3 | After WS-4-6 | After WS-7 |
|----------|---------|--------------|--------------|-------------|
| Data Discovery | 30% | 90% | 95% | 95% |
| Data Quality | 20% | 85% | 90% | 90% |
| Data Governance | 40% | 50% | 90% | 95% |
| Observability | 60% | 65% | 95% | 95% |
| Query Capability | 70% | 70% | 90% | 95% |
| ML/AI Readiness | 15% | 15% | 20% | 80% |
| **Overall** | **39%** | **63%** | **80%** | **92%** |

### Key Performance Indicators

| KPI | Target |
|-----|--------|
| Catalog scan latency (1000 tables) | < 30s |
| Quality check execution (single rule) | < 5s |
| Schema compatibility check | < 100ms |
| Alert evaluation cycle | < 1s |
| Federated query (2 sources, 10K rows) | < 3s |
| Feature online serving (p99) | < 10ms |
| Streaming analytics window flush | < 500ms |

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| etcd key-space explosion (24 new prefixes) | High | Implement compaction, TTL on transient resources (checks, incidents), monitor with existing keyspace alerting |
| Federation query performance | Medium | Result caching, materialized views, query timeout enforcement |
| Schema registry backward compat complexity | Medium | Start with JSON Schema only, add Avro/Protobuf incrementally |
| Governance erasure cascading failures | High | Dry-run mode first, per-asset timeout, partial completion tracking |
| Feature store online serving latency | Medium | Redis/Valkey backend, connection pooling, pre-warming |
| Alert storm (many rules firing simultaneously) | Medium | Deduplication, grouping, rate limiting per channel |

---

## Rollback Strategy

Every new module follows the established AxiomNizam rollback pattern:

1. **Feature flag OFF** = module completely disabled, zero impact
2. **No schema migrations** = etcd prefixes are additive, removing them is safe
3. **No existing code modified** = new modules are purely additive
4. **Independent deployment** = each workstream can ship independently

```bash
# Disable any module instantly:
export RECONCILER_ENABLED_CATALOG=false
export RECONCILER_ENABLED_QUALITY_RULES=false
export RECONCILER_ENABLED_SCHEMA=false
# ... restart server
```

---

## Post-Completion: Platform Comparison

After all 7 workstreams are complete, AxiomNizam will cover:

| Capability | Databricks | Snowflake | AxiomNizam |
|-----------|-----------|-----------|------------|
| Multi-database support | Limited | Single | 7 databases |
| Declarative control plane | No | No | Yes (K8s-style) |
| Data catalog | Unity Catalog | Information Schema | Full catalog with enrichment |
| Data quality | Lakehouse monitoring | Data quality | 15 rule types + contracts |
| Schema registry | No (use Confluent) | Schema evolution | Built-in with compatibility |
| Federated query | No | No | Cross-database SQL |
| Data governance | Unity governance | Governance | GDPR/HIPAA/SOC2/PCI-DSS |
| Feature store | Feature Engineering | No | Online + offline serving |
| Streaming analytics | Structured Streaming | Streams | Windowed aggregations |
| ML pipelines | MLflow | Snowpark ML | Training + deployment + A/B |
| Alerting | No (use external) | Alerts | Multi-channel with escalation |
| SLO tracking | No | No | Error budgets + burn rate |
| Cost attribution | Cost management | Credit usage | Per-tenant/query/pipeline |
| Object storage | No | Stages | Native S3-compatible |
| Event bus | No | No | Pub/sub + DLQ + replay |
| CDC pipelines | Delta Live Tables | Streams | Source/sink/filter model |
| ETL pipelines | Workflows | Tasks | 10-step engine |
| API management | No | No | Full lifecycle |
| CLI tooling | dbx CLI | SnowSQL | kubectl-style (axiomnizamctl) |
| Self-hosted | No | No | Yes (docker-compose) |

---

## Conclusion

This plan adds 22 new modules (101 files, ~41,600 lines) to AxiomNizam, bringing the total to 110 internal modules, 53 reconciler controllers, and 54+ etcd prefixes. The platform will cover 92% of enterprise data platform requirements while maintaining its unique declarative architecture advantage.

**Start with WS-1 (Catalog)** — it is the foundation everything else builds on.

---

*Document maintained by Platform Architecture Team. Update as implementation progresses.*
