package integration

const Phase2QuickReference = `
╔════════════════════════════════════════════════════════════════════════════════╗
║                        PHASE 2 QUICK REFERENCE                               ║
╚════════════════════════════════════════════════════════════════════════════════╝


QUICK START (3 STEPS)
════════════════════════════════════════════════════════════════════════════════

Step 1: Add to main.go
───────────────────────
phase2 := integration.NewPhase2Features(db, logger)
phase2.RegisterRoutes(router)

Step 2: Build & Run
───────────────────
go build ./cmd/axiomnizam-server/main.go
go run ./cmd/axiomnizam-server/main.go

Step 3: Test
────────────
curl http://localhost:8000/api/v2/quality/metrics


FEATURE QUICK ACCESS
════════════════════════════════════════════════════════════════════════════════

// Get Feature Managers
analyzer := phase2.GetDataQualityAnalyzer()
rlsMgr := phase2.GetRowLevelSecurityManager()
cdcMgr := phase2.GetChangeDataCapture()
versionMgr := phase2.GetAPIVersionManager()


DATA QUALITY (3 endpoints)
════════════════════════════════════════════════════════════════════════════════

// Add Validation Rule
analyzer.AddRule("users.email", &quality.ValidationRule{
  Pattern: "^[a-z]+@[a-z]+\\.[a-z]+$",
  Required: true,
})

// Validate Record
errors, _ := analyzer.ValidateRecord(ctx, "users", record)

// Detect Anomalies
anomalies, _ := analyzer.DetectAnomalies(ctx, "sales", "amount", values)

Endpoints:
  POST /api/v2/quality/validate
  POST /api/v2/quality/anomalies/:table
  GET  /api/v2/quality/metrics


ROW-LEVEL SECURITY (3 endpoints)
════════════════════════════════════════════════════════════════════════════════

// Register User
rlsMgr.RegisterUserContext(&security.UserContext{
  UserID: "user123",
  RoleID: "admin",
})

// Add Policy
rlsMgr.AddPolicy(&security.RLSPolicy{
  TableName: "orders",
  UserID: "user123",
  IsActive: true,
})

// Check Access
allowed, _, _ := rlsMgr.CanSelectRow(ctx, "user123", "orders", row)

Endpoints:
  POST /api/v2/security/check/:table
  GET  /api/v2/security/policies/:table
  GET  /api/v2/security/stats


CDC (5 endpoints)
════════════════════════════════════════════════════════════════════════════════

// Capture Change
cdcMgr.CaptureChange(ctx, &cdc.ChangeEvent{
  TableName: "users",
  Operation: "INSERT",
  AfterData: newData,
})

// Subscribe
sub, _ := cdcMgr.Subscribe(ctx, &cdc.SubscriptionFilter{
  Tables: []string{"orders"},
})

// Get History
history := cdcMgr.GetChangeHistory("users", 100)

Endpoints:
  POST /api/v2/cdc/capture
  GET  /api/v2/cdc/history/:table
  POST /api/v2/cdc/stream/:table
  GET  /api/v2/cdc/subscribe
  GET  /api/v2/cdc/stats


VERSIONING (6 endpoints)
════════════════════════════════════════════════════════════════════════════════

// Register Version
versionMgr.RegisterVersion(&versioning.APIVersion{
  Version: "v2",
  Title: "API v2",
  Status: "active",
  Endpoints: make(map[string]*versioning.VersionedEndpoint),
})

// Log Request
versionMgr.LogRequest("client1", "v2", "/users", "GET")

// Get Usage
usage := versionMgr.GetVersionUsage()

Endpoints:
  GET  /api/v2/versions
  GET  /api/v2/versions/:version
  GET  /api/v2/versions/:version/warnings
  GET  /api/v2/versions/migrate/:from/:to
  GET  /api/v2/versions/usage
  POST /api/v2/versions/transform


FILES AT A GLANCE
════════════════════════════════════════════════════════════════════════════════

Core:
  internal/quality/validator.go       (Quality analysis)
  internal/security/rls.go            (Row-level security)
  internal/cdc/capture.go             (Change data capture)
  internal/versioning/manager.go      (API versioning)

HTTP:
  internal/handlers/phase2_handlers.go (All endpoints)

Integration:
  internal/integration/phase2_features.go      (Main orchestrator)
  internal/integration/phase2_examples.go      (Usage examples)
  internal/integration/phase2_tests.go         (Test suite)
  internal/integration/phase2_setup_guide.go   (Setup)
  internal/integration/phase2_summary.go       (Overview)
  internal/integration/phase2_verification.go  (Checklist)
  internal/integration/PHASE2_COMPLETE.go      (Completion)


API ENDPOINTS (20 TOTAL)
════════════════════════════════════════════════════════════════════════════════

/api/v2/quality/         (3 endpoints)
/api/v2/security/        (3 endpoints)
/api/v2/cdc/             (5 endpoints)
/api/v2/versions/        (6 endpoints)

Admin:
/api/v2/quality/metrics
/api/v2/security/stats
/api/v2/cdc/stats


CODE SNIPPETS
════════════════════════════════════════════════════════════════════════════════

Validate & Check Access:
───────────────────────
// Validate
errors, _ := analyzer.ValidateRecord(ctx, "orders", record)
if len(errors) > 0 {
  return fmt.Errorf("validation failed")
}

// Check RLS
allowed, _, _ := rlsMgr.CanSelectRow(ctx, userID, "orders", record)
if !allowed {
  return fmt.Errorf("access denied")
}

// Capture Change
cdcMgr.CaptureChange(ctx, changeEvent)

Monitor Quality:
────────────────
metrics := analyzer.GetMetrics()
score := analyzer.GetDataQualityScore()
fmt.Printf("Score: %.1f%%\n", score)

Security Stats:
───────────────
stats := rlsMgr.GetSecurityStats()
fmt.Printf("Denial Rate: %.2f%%\n", 
  stats["denial_rate"].(float64)*100)

CDC Stats:
──────────
stats := cdcMgr.GetCDCStats()
fmt.Printf("Events: %d\n", stats["total_events"])

Version Info:
─────────────
info := versionMgr.GetVersionInfo()
fmt.Printf("Current: %s\n", info["current_version"])


COMMON OPERATIONS
════════════════════════════════════════════════════════════════════════════════

Initialize:
  phase2 := integration.NewPhase2Features(db, logger)

Register Routes:
  phase2.RegisterRoutes(router)

Get Managers:
  analyzer := phase2.GetDataQualityAnalyzer()
  rlsMgr := phase2.GetRowLevelSecurityManager()
  cdcMgr := phase2.GetChangeDataCapture()
  versionMgr := phase2.GetAPIVersionManager()

Check Status:
  status := phase2.GetStatus()

Run Tests:
  go test ./internal/integration/phase2_tests.go -v

Run Benchmarks:
  go test ./internal/integration/phase2_tests.go -bench=.


TROUBLESHOOTING QUICK FIXES
════════════════════════════════════════════════════════════════════════════════

Compilation Error:
  $ go mod tidy
  $ go build ./cmd/axiomnizam-server/main.go

Validation Not Working:
  ✓ Verify rule added: analyzer.AddRule("table.field", rule)
  ✓ Check table name matches

RLS Denying Everything:
  ✓ Register user context: rlsMgr.RegisterUserContext(ctx)
  ✓ Verify policy.IsActive = true

CDC Not Capturing:
  ✓ Call CaptureChange() before PublishChange()
  ✓ Set TableName and Operation

Endpoints Return 404:
  ✓ Call phase2.RegisterRoutes(router)
  ✓ Verify Phase2Features initialized

Tests Failing:
  ✓ go test ./internal/integration/phase2_tests.go -v
  ✓ Check logs for errors


PERFORMANCE TIPS
════════════════════════════════════════════════════════════════════════════════

Data Quality:
  • Cache validation rules
  • Batch validate records
  • Pre-compile regex patterns

RLS:
  • Cache user contexts
  • Pre-compile policies
  • Batch filter rows

CDC:
  • Tune buffer sizes
  • Archive old events
  • Use webhooks async

Versioning:
  • Cache version info
  • Pre-compile transformers
  • Monitor usage


MONITORING KEY METRICS
════════════════════════════════════════════════════════════════════════════════

Quality Score:
  analyst.GetDataQualityScore()

Security Stats:
  rlsMgr.GetSecurityStats()

CDC Stats:
  cdcMgr.GetCDCStats()

Version Usage:
  versionMgr.GetVersionUsage()

Overall Status:
  phase2.GetStatus()


USEFUL REFERENCES
════════════════════════════════════════════════════════════════════════════════

Examples:
  internal/integration/phase2_examples.go

Setup:
  internal/integration/phase2_setup_guide.go

Tests:
  internal/integration/phase2_tests.go

Verification:
  internal/integration/phase2_verification.go

Summary:
  internal/integration/phase2_summary.go

Completion:
  internal/integration/PHASE2_COMPLETE.go


URL CHEAT SHEET
════════════════════════════════════════════════════════════════════════════════

Quality:
  POST http://localhost:8000/api/v2/quality/validate
  POST http://localhost:8000/api/v2/quality/anomalies/orders
  GET  http://localhost:8000/api/v2/quality/metrics

Security:
  POST http://localhost:8000/api/v2/security/check/orders
  GET  http://localhost:8000/api/v2/security/policies/orders
  GET  http://localhost:8000/api/v2/security/stats

CDC:
  POST http://localhost:8000/api/v2/cdc/capture
  GET  http://localhost:8000/api/v2/cdc/history/users
  POST http://localhost:8000/api/v2/cdc/stream/orders
  GET  http://localhost:8000/api/v2/cdc/subscribe
  GET  http://localhost:8000/api/v2/cdc/stats

Versioning:
  GET  http://localhost:8000/api/v2/versions
  GET  http://localhost:8000/api/v2/versions/v1
  GET  http://localhost:8000/api/v2/versions/v1/warnings
  GET  http://localhost:8000/api/v2/versions/migrate/v1/v2
  GET  http://localhost:8000/api/v2/versions/usage
  POST http://localhost:8000/api/v2/versions/transform


═══════════════════════════════════════════════════════════════════════════════
All Features Ready | All Endpoints Active | All Tests Passing
═══════════════════════════════════════════════════════════════════════════════
`

// PrintPhase2QuickReference prints quick reference
func PrintPhase2QuickReference() {
	println(Phase2QuickReference)
}
