package integration

const Phase2Examples = `
╔════════════════════════════════════════════════════════════════════════════════╗
║                                                                                ║
║                    PHASE 2 FEATURES USAGE EXAMPLES                           ║
║                                                                                ║
║     Data Quality | Row-Level Security | CDC | API Versioning                 ║
║                                                                                ║
╚════════════════════════════════════════════════════════════════════════════════╝


1. DATA QUALITY & VALIDATION
════════════════════════════════════════════════════════════════════════════════

Example 1: Add Validation Rules
────────────────────────────────

analyzeir := phase2.GetDataQualityAnalyzer()

// Add a rule for email validation
emailRule := &quality.ValidationRule{
  ID:       "email-format",
  FieldName: "email",
  Pattern:  "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
  Required: true,
}
analyzer.AddRule("users.email", emailRule)

// Add rule for age validation
ageRule := &quality.ValidationRule{
  ID:        "age-range",
  FieldName: "age",
  MinValue:  0,
  MaxValue:  150,
  Required:  true,
}
analyzer.AddRule("users.age", ageRule)


Example 2: Validate Record
───────────────────────────

record := map[string]interface{}{
  "email": "user@example.com",
  "age":   25,
  "name":  "John Doe",
}

errors, err := analyzer.ValidateRecord(ctx, "users", record)
if len(errors) > 0 {
  for _, e := range errors {
    fmt.Printf("Validation error in %s: %s\n", e.Field, e.Message)
  }
}


Example 3: Detect Anomalies
────────────────────────────

values := []interface{}{100, 105, 98, 102, 500} // 500 is anomaly

anomalies, err := analyzer.DetectAnomalies(ctx, "orders", "amount", values)
for _, anomaly := range anomalies {
  fmt.Printf("Anomaly detected: %v (score: %.2f)\n", anomaly.Value, anomaly.Score)
}


Example 4: Get Quality Metrics
───────────────────────────────

metrics := analyzer.GetMetrics()
fmt.Printf("Quality Score: %.1f%%\n", analyzer.GetDataQualityScore())
fmt.Printf("Passed Checks: %d\n", metrics.PassedChecks)
fmt.Printf("Failed Checks: %d\n", metrics.FailedChecks)


HTTP Endpoints:
  POST   /api/v2/quality/validate
  POST   /api/v2/quality/anomalies/:table
  GET    /api/v2/quality/metrics


2. ROW-LEVEL SECURITY (RLS)
════════════════════════════════════════════════════════════════════════════════

Example 1: Register User Context
─────────────────────────────────

rlsMgr := phase2.GetRowLevelSecurityManager()

userCtx := &security.UserContext{
  UserID:     "user123",
  RoleID:     "admin",
  TenantID:   "tenant1",
  Attributes: map[string]string{
    "department": "sales",
    "region":     "north",
  },
  Groups: []string{"admin", "managers"},
}

rlsMgr.RegisterUserContext(userCtx)


Example 2: Add RLS Policy
─────────────────────────

policy := &security.RLSPolicy{
  TableName:   "orders",
  PolicyName:  "user-owns-order",
  PolicyType:  "SELECT",
  UserID:      "user123",
  Description: "Users can only see their own orders",
  IsActive:    true,
  Attributes: map[string]string{
    "department": "sales",
  },
}

rlsMgr.AddPolicy(policy)


Example 3: Check Row Access
────────────────────────────

row := map[string]interface{}{
  "id":       1,
  "user_id":  "user123",
  "amount":   100,
  "tenant_id": "tenant1",
}

allowed, reason, err := rlsMgr.CanSelectRow(ctx, "user123", "orders", row)
if allowed {
  fmt.Println("Access granted:", reason)
} else {
  fmt.Println("Access denied:", reason)
}


Example 4: Filter Rows by RLS
──────────────────────────────

rows := []map[string]interface{}{
  {"id": 1, "user_id": "user123", "tenant_id": "tenant1"},
  {"id": 2, "user_id": "user456", "tenant_id": "tenant1"},
}

filtered, err := rlsMgr.FilterRows(ctx, "user123", "orders", rows)
// Returns only rows where user123 has access


Example 5: Get Security Stats
──────────────────────────────

stats := rlsMgr.GetSecurityStats()
fmt.Printf("Denial Rate: %.2f%%\n", stats["denial_rate"].(float64)*100)


HTTP Endpoints:
  POST   /api/v2/security/check/:table
  GET    /api/v2/security/policies/:table
  GET    /api/v2/security/stats


3. CHANGE DATA CAPTURE (CDC)
════════════════════════════════════════════════════════════════════════════════

Example 1: Capture Change Event
────────────────────────────────

cdcMgr := phase2.GetChangeDataCapture()

changeEvent := &cdc.ChangeEvent{
  TableName: "users",
  Operation: "UPDATE",
  BeforeData: map[string]interface{}{
    "id":     1,
    "name":   "John",
    "email":  "john@old.com",
  },
  AfterData: map[string]interface{}{
    "id":     1,
    "name":   "John",
    "email":  "john@new.com",
  },
}

err := cdcMgr.CaptureChange(ctx, changeEvent)


Example 2: Subscribe to Changes
────────────────────────────────

filter := &cdc.SubscriptionFilter{
  Tables:     []string{"users", "orders"},
  Operations: []string{"INSERT", "UPDATE"},
}

sub, err := cdcMgr.Subscribe(ctx, filter)

// Read changes
go func() {
  for event := range sub.Channel {
    fmt.Printf("Change detected: %s on %s\n", event.Operation, event.TableName)
  }
}()


Example 3: Create CDC Stream
─────────────────────────────

stream, err := cdcMgr.CreateStream("orders")
fmt.Printf("Stream ID: %s\n", stream.ID)

events, err := cdcMgr.GetStreamEvents(stream.ID, 100)


Example 4: Add Webhook
──────────────────────

webhook := &cdc.WebhookSubscription{
  URL: "https://api.example.com/webhooks/changes",
  TableNames: []string{"orders"},
  EventTypes: []string{"INSERT", "UPDATE"},
  Active: true,
}

cdcMgr.AddWebhook(webhook)


Example 5: Get Change History
──────────────────────────────

history := cdcMgr.GetChangeHistory("users", 50)
for _, change := range history {
  fmt.Printf("%s: %s operation on %s\n", 
    change.Timestamp, change.Operation, change.TableName)
}


Example 6: Get CDC Statistics
──────────────────────────────

stats := cdcMgr.GetCDCStats()
fmt.Printf("Total events: %d\n", stats["total_events"])
fmt.Printf("Buffer utilization: %.1f%%\n", stats["buffer_utilization"])


HTTP Endpoints:
  POST   /api/v2/cdc/capture
  GET    /api/v2/cdc/history/:table
  POST   /api/v2/cdc/stream/:table
  GET    /api/v2/cdc/subscribe
  GET    /api/v2/cdc/stats


4. API VERSIONING
════════════════════════════════════════════════════════════════════════════════

Example 1: Register API Version
────────────────────────────────

versionMgr := phase2.GetAPIVersionManager()

v2 := &versioning.APIVersion{
  Version:     "v2",
  Title:       "API Version 2",
  Description: "Enhanced version with new features",
  Endpoints:   make(map[string]*versioning.VersionedEndpoint),
  Status:      "active",
}

versionMgr.RegisterVersion(v2)


Example 2: Register Endpoint
─────────────────────────────

endpoint := &versioning.VersionedEndpoint{
  Path: "/users/:id",
  Method: "GET",
  RequestSchema: map[string]interface{}{
    "id": "string",
  },
  ResponseSchema: map[string]interface{}{
    "id": "string",
    "name": "string",
  },
}

versionMgr.RegisterEndpoint("v2", "/users/:id", "GET", endpoint)


Example 3: Deprecate Endpoint
──────────────────────────────

versionMgr.DeprecateEndpoint("v1", "/users/:id", "GET", 
  "Use v2 API instead", "/api/v2/users/:id")


Example 4: Get Deprecation Warnings
────────────────────────────────────

warnings := versionMgr.GetDeprecationWarnings("v1")
for _, warning := range warnings {
  fmt.Println("⚠️ ", warning)
}


Example 5: Create Migration Path
─────────────────────────────────

path, _ := versionMgr.CreateMigrationPath("v1", "v2")

// Add migration steps
versionMgr.AddMigrationStep("v1", "v2", &versioning.MigrationStep{
  Description: "Rename 'user_id' to 'userId'",
  Action:      "rename_field",
  Mapping: map[string]interface{}{
    "from": "user_id",
    "to":   "userId",
  },
})


Example 6: Transform Request
─────────────────────────────

oldData := map[string]interface{}{
  "user_id": 123,
  "name":    "John",
}

newData, _ := versionMgr.TransformRequest(oldData, "v1", "v2")


Example 7: Log Request
──────────────────────

versionMgr.LogRequest("client123", "v2", "/users", "GET")

usage := versionMgr.GetVersionUsage()
fmt.Printf("v2 usage: %d requests\n", usage["v2"])


Example 8: Get Version Information
───────────────────────────────────

info := versionMgr.GetVersionInfo()
fmt.Printf("Current version: %s\n", info["current_version"])


Example 9: Get Endpoint History
────────────────────────────────

history := versionMgr.GetEndpointHistory("/users/:id", "GET")
for _, entry := range history {
  fmt.Printf("Version %s: %v\n", entry["version"], entry["deprecated"])
}


HTTP Endpoints:
  GET    /api/v2/versions
  GET    /api/v2/versions/:version
  GET    /api/v2/versions/:version/warnings
  GET    /api/v2/versions/migrate/:from/:to
  GET    /api/v2/versions/usage
  POST   /api/v2/versions/transform


INTEGRATED USAGE (All Features)
════════════════════════════════════════════════════════════════════════════════

func processUserUpdate(phase2 *Phase2Features, userID string, record map[string]interface{}) {
  analyzer := phase2.GetDataQualityAnalyzer()
  rlsMgr := phase2.GetRowLevelSecurityManager()
  cdcMgr := phase2.GetChangeDataCapture()

  // 1. Validate data quality
  errors, _ := analyzer.ValidateRecord(ctx, "users", record)
  if len(errors) > 0 {
    fmt.Println("Validation failed")
    return
  }

  // 2. Check row-level security
  allowed, _, _ := rlsMgr.CanUpdateRow(ctx, userID, "users", record)
  if !allowed {
    fmt.Println("Access denied")
    return
  }

  // 3. Capture change
  event := &cdc.ChangeEvent{
    TableName: "users",
    Operation: "UPDATE",
    AfterData: record,
  }
  cdcMgr.CaptureChange(ctx, event)
  cdcMgr.PublishChange(event)

  fmt.Println("User updated successfully")
}


INITIALIZATION IN main.go
════════════════════════════════════════════════════════════════════════════════

// In your main.go
import "AxiomNizam/internal/integration"

func main() {
  db := setupDatabase()
  router := gin.Default()

  // Initialize Phase 2 features
  phase2 := integration.NewPhase2Features(db, logger)

  // Register Phase 2 routes
  phase2.RegisterRoutes(router)

  // Get status
  status := phase2.GetStatus()
  fmt.Printf("Phase 2 status: %+v\n", status)

  router.Run(":8000")
}


════════════════════════════════════════════════════════════════════════════════
`

// PrintPhase2Examples prints Phase 2 examples
func PrintPhase2Examples() {
	println(Phase2Examples)
}
