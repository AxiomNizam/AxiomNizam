package integration

import (
	"context"
	"testing"

	"example.com/axiomnizam/internal/cdc"
	"example.com/axiomnizam/internal/quality"
	"example.com/axiomnizam/internal/security"
	"example.com/axiomnizam/internal/versioning"
	"go.uber.org/zap"
)

func TestDataQualityValidator(t *testing.T) {
	analyzer := quality.NewDataQualityAnalyzer()

	// Add validation rule
	rule := &quality.ValidationRule{
		ID:        "test-rule",
		FieldName: "email",
		Pattern:   "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
		Required:  true,
	}

	if err := analyzer.AddRule("users.email", rule); err != nil {
		t.Fatalf("Failed to add rule: %v", err)
	}

	// Validate valid record
	validRecord := map[string]interface{}{
		"email": "user@example.com",
	}

	errors, err := analyzer.ValidateRecord(context.Background(), "users", validRecord)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if len(errors) != 0 {
		t.Fatalf("Expected no errors, got %d", len(errors))
	}

	// Validate invalid record
	invalidRecord := map[string]interface{}{
		"email": "invalid-email",
	}

	errors, _ = analyzer.ValidateRecord(context.Background(), "users", invalidRecord)
	if len(errors) == 0 {
		t.Fatal("Expected validation error for invalid email")
	}

	t.Log("✓ Data quality validation tests passed")
}

func TestAnomalyDetection(t *testing.T) {
	analyzer := quality.NewDataQualityAnalyzer()

	// Normal values
	normalValues := []interface{}{100.0, 102.0, 98.0, 101.0, 99.0, 100.0}
	anomalies, _ := analyzer.DetectAnomalies(context.Background(), "sales", "amount", normalValues)

	if len(anomalies) > 0 {
		t.Fatalf("Expected no anomalies in normal data, got %d", len(anomalies))
	}

	// Values with anomaly
	anomalousValues := []interface{}{100.0, 102.0, 98.0, 500.0} // 500 is anomaly
	anomalies, _ = analyzer.DetectAnomalies(context.Background(), "sales", "amount", anomalousValues)

	if len(anomalies) == 0 {
		t.Fatal("Expected to detect anomaly")
	}

	t.Log("✓ Anomaly detection tests passed")
}

func TestRowLevelSecurity(t *testing.T) {
	rlsMgr := security.NewRowLevelSecurityManager()

	// Register user context
	userCtx := &security.UserContext{
		UserID:     "user123",
		RoleID:     "admin",
		Attributes: map[string]string{"dept": "sales"},
	}

	if err := rlsMgr.RegisterUserContext(userCtx); err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Add policy
	policy := &security.RLSPolicy{
		TableName:  "orders",
		PolicyName: "user-orders",
		PolicyType: "SELECT",
		UserID:     "user123",
		IsActive:   true,
	}

	if err := rlsMgr.AddPolicy(policy); err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}

	// Check access
	row := map[string]interface{}{
		"id":      1,
		"user_id": "user123",
	}

	allowed, _, _ := rlsMgr.CanSelectRow(context.Background(), "user123", "orders", row)

	if !allowed {
		t.Fatal("Expected access to be allowed")
	}

	// Check stats
	stats := rlsMgr.GetSecurityStats()
	if stats["total_policies"].(int) != 1 {
		t.Fatalf("Expected 1 policy, got %d", stats["total_policies"])
	}

	t.Log("✓ Row-level security tests passed")
}

func TestRowLevelSecurityDenial(t *testing.T) {
	rlsMgr := security.NewRowLevelSecurityManager()

	// Register admin user
	adminCtx := &security.UserContext{
		UserID: "admin123",
		RoleID: "admin",
	}
	rlsMgr.RegisterUserContext(adminCtx)

	// Add policy for specific role
	policy := &security.RLSPolicy{
		TableName:  "sensitive_data",
		PolicyName: "admin-only",
		PolicyType: "SELECT",
		RoleID:     "admin",
		IsActive:   true,
	}
	rlsMgr.AddPolicy(policy)

	// Try access as non-admin user
	userCtx := &security.UserContext{
		UserID: "user123",
		RoleID: "user",
	}
	rlsMgr.RegisterUserContext(userCtx)

	row := map[string]interface{}{"id": 1}
	allowed, _, _ := rlsMgr.CanSelectRow(context.Background(), "user123", "sensitive_data", row)

	if allowed {
		t.Fatal("Expected access to be denied for non-admin user")
	}

	t.Log("✓ RLS denial tests passed")
}

func TestChangeDataCapture(t *testing.T) {
	cdcMgr := cdc.NewChangeDataCapture()

	// Capture change
	event := &cdc.ChangeEvent{
		TableName: "users",
		Operation: "INSERT",
		AfterData: map[string]interface{}{
			"id":   1,
			"name": "John",
		},
	}

	if err := cdcMgr.CaptureChange(context.Background(), event); err != nil {
		t.Fatalf("Failed to capture change: %v", err)
	}

	// Get change history
	history := cdcMgr.GetChangeHistory("users", 10)

	if len(history) == 0 {
		t.Fatal("Expected to get change history")
	}

	if history[0].TableName != "users" {
		t.Fatalf("Expected table 'users', got '%s'", history[0].TableName)
	}

	t.Log("✓ CDC capture tests passed")
}

func TestCDCSubscription(t *testing.T) {
	cdcMgr := cdc.NewChangeDataCapture()

	// Create subscription
	filter := &cdc.SubscriptionFilter{
		Tables:     []string{"users"},
		Operations: []string{"INSERT", "UPDATE"},
	}

	sub, err := cdcMgr.Subscribe(context.Background(), filter)
	if err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	if sub == nil || sub.Channel == nil {
		t.Fatal("Expected valid subscription")
	}

	// Cleanup
	cdcMgr.Unsubscribe(sub)

	t.Log("✓ CDC subscription tests passed")
}

func TestCDCStream(t *testing.T) {
	cdcMgr := cdc.NewChangeDataCapture()

	// Create stream
	stream, err := cdcMgr.CreateStream("orders")
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	if stream.TableName != "orders" {
		t.Fatalf("Expected table 'orders', got '%s'", stream.TableName)
	}

	if stream.Status != "active" {
		t.Fatalf("Expected status 'active', got '%s'", stream.Status)
	}

	t.Log("✓ CDC stream tests passed")
}

func TestAPIVersioning(t *testing.T) {
	versionMgr := versioning.NewAPIVersionManager("v1")

	// Register version
	v2 := &versioning.APIVersion{
		Version:     "v2",
		Title:       "API v2",
		Description: "Enhanced API",
		Endpoints:   make(map[string]*versioning.VersionedEndpoint),
		Status:      "active",
	}

	if err := versionMgr.RegisterVersion(v2); err != nil {
		t.Fatalf("Failed to register version: %v", err)
	}

	// Get version
	retrieved, err := versionMgr.GetVersion("v2")
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if retrieved.Version != "v2" {
		t.Fatalf("Expected version 'v2', got '%s'", retrieved.Version)
	}

	t.Log("✓ API versioning tests passed")
}

func TestDeprecationWarnings(t *testing.T) {
	versionMgr := versioning.NewAPIVersionManager("v1")

	v1 := &versioning.APIVersion{
		Version:   "v1",
		Title:     "API v1",
		Endpoints: make(map[string]*versioning.VersionedEndpoint),
		Status:    "active",
	}
	versionMgr.RegisterVersion(v1)

	// Deprecate version
	if err := versionMgr.DeprecateVersion("v1", "Use v2 instead"); err != nil {
		t.Fatalf("Failed to deprecate version: %v", err)
	}

	// Get warnings
	warnings := versionMgr.GetDeprecationWarnings("v1")

	if len(warnings) == 0 {
		t.Fatal("Expected deprecation warnings")
	}

	t.Log("✓ Deprecation warning tests passed")
}

func TestAPIVersionUsage(t *testing.T) {
	versionMgr := versioning.NewAPIVersionManager("v1")

	v1 := &versioning.APIVersion{
		Version:   "v1",
		Title:     "API v1",
		Endpoints: make(map[string]*versioning.VersionedEndpoint),
		Status:    "active",
	}
	versionMgr.RegisterVersion(v1)

	// Log requests
	versionMgr.LogRequest("client1", "v1", "/users", "GET")
	versionMgr.LogRequest("client2", "v1", "/users", "GET")

	usage := versionMgr.GetVersionUsage()

	if usage["v1"] != 2 {
		t.Fatalf("Expected usage of 2, got %d", usage["v1"])
	}

	t.Log("✓ API version usage tests passed")
}

// Benchmarks

func BenchmarkValidation(b *testing.B) {
	analyzer := quality.NewDataQualityAnalyzer()

	rule := &quality.ValidationRule{
		ID:        "bench-rule",
		FieldName: "email",
		Pattern:   "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
	}
	analyzer.AddRule("users.email", rule)

	record := map[string]interface{}{
		"email": "user@example.com",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		analyzer.ValidateRecord(context.Background(), "users", record)
	}

	b.ReportAllocs()
}

func BenchmarkRLSCheck(b *testing.B) {
	rlsMgr := security.NewRowLevelSecurityManager()

	userCtx := &security.UserContext{
		UserID: "user123",
		RoleID: "admin",
	}
	rlsMgr.RegisterUserContext(userCtx)

	policy := &security.RLSPolicy{
		TableName:  "orders",
		PolicyName: "test",
		UserID:     "user123",
		IsActive:   true,
	}
	rlsMgr.AddPolicy(policy)

	row := map[string]interface{}{"id": 1}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rlsMgr.CanSelectRow(context.Background(), "user123", "orders", row)
	}

	b.ReportAllocs()
}

func BenchmarkCDCCapture(b *testing.B) {
	cdcMgr := cdc.NewChangeDataCapture()

	event := &cdc.ChangeEvent{
		TableName: "users",
		Operation: "INSERT",
		AfterData: map[string]interface{}{"id": 1},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cdcMgr.CaptureChange(context.Background(), event)
	}

	b.ReportAllocs()
}

func TestPhase2Integration(t *testing.T) {
	logger, _ := zap.NewProduction()

	// Create Phase 2 features (requires DB mock, so simplified test)
	analyzer := quality.NewDataQualityAnalyzer()
	rlsMgr := security.NewRowLevelSecurityManager()
	cdcMgr := cdc.NewChangeDataCapture()
	versionMgr := versioning.NewAPIVersionManager("v1")

	// Verify all components initialized
	if analyzer == nil || rlsMgr == nil || cdcMgr == nil || versionMgr == nil {
		t.Fatal("Phase 2 components not initialized")
	}

	logger.Info("All Phase 2 components operational")

	t.Log("✓ Phase 2 integration test passed")
}
