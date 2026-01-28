package integration

import (
	"testing"
	"time"

	"example.com/axiomnizam/internal/audit"
	"example.com/axiomnizam/internal/encryption"
	"example.com/axiomnizam/internal/lineage"
	"example.com/axiomnizam/internal/workflow"
)

// TestEncryptionRegistration tests key registration
func TestEncryptionRegistration(t *testing.T) {
	mgr := encryption.NewFieldLevelEncryption()

	key := &encryption.EncryptionKey{
		ID:          "test-key-1",
		KeyMaterial: "32-byte-encryption-key-value-12345",
		Algorithm:   "AES-256",
		KeyLength:   256,
		CreatedAt:   time.Now(),
	}

	err := mgr.RegisterKey(key)
	if err != nil {
		t.Errorf("Failed to register key: %v", err)
	}
}

// TestEncryptionPolicy tests policy registration
func TestEncryptionPolicy(t *testing.T) {
	mgr := encryption.NewFieldLevelEncryption()

	key := &encryption.EncryptionKey{
		ID:          "test-key-1",
		KeyMaterial: "32-byte-encryption-key-value-12345",
	}
	mgr.RegisterKey(key)

	policy := &encryption.FieldEncryptionPolicy{
		Name:         "email_policy",
		ResourceType: "users",
		KeyID:        "test-key-1",
	}

	err := mgr.AddEncryptionPolicy(policy)
	if err != nil {
		t.Errorf("Failed to add encryption policy: %v", err)
	}
}

// TestEncryptDecrypt tests encryption and decryption
func TestEncryptDecrypt(t *testing.T) {
	mgr := encryption.NewFieldLevelEncryption()

	key := &encryption.EncryptionKey{
		ID:          "test-key-1",
		KeyMaterial: "32-byte-encryption-key-value-12345",
	}
	mgr.RegisterKey(key)

	testData := "test@example.com"
	encrypted, err := mgr.EncryptField("users.email", testData, "test-key-1")
	if err != nil {
		t.Errorf("Encryption failed: %v", err)
		return
	}

	decrypted, err := mgr.DecryptField(encrypted)
	if err != nil {
		t.Errorf("Decryption failed: %v", err)
		return
	}

	decryptedStr, ok := decrypted.(string)
	if !ok {
		t.Errorf("Decryption returned non-string type")
		return
	}

	if decryptedStr != testData {
		t.Errorf("Decrypted data mismatch: got %s, want %s", decryptedStr, testData)
	}
}

// TestKeyRotation tests key rotation
func TestKeyRotation(t *testing.T) {
	mgr := encryption.NewFieldLevelEncryption()

	key := &encryption.EncryptionKey{
		ID:          "test-key-1",
		KeyMaterial: "32-byte-encryption-key-value-12345",
	}
	mgr.RegisterKey(key)

	newKey := &encryption.EncryptionKey{
		ID:          "test-key-2",
		KeyMaterial: "new-32-byte-encryption-key-value-123",
	}
	_, err := mgr.RotateKey("test-key-1", newKey)
	if err != nil {
		t.Errorf("Key rotation failed: %v", err)
	}
}

// TestLineageNodeRegistration tests data node registration
func TestLineageNodeRegistration(t *testing.T) {
	mgr := lineage.NewDataLineageTracker()

	node := &lineage.DataLineageNode{
		ID:     "tbl_test",
		Name:   "Test Table",
		Type:   "table",
		Schema: "public",
	}

	err := mgr.RegisterDataNode(node)
	if err != nil {
		t.Errorf("Failed to register node: %v", err)
	}
}

// TestLineageEdgeCreation tests lineage edge creation
func TestLineageEdgeCreation(t *testing.T) {
	mgr := lineage.NewDataLineageTracker()

	node1 := &lineage.DataLineageNode{
		ID:   "tbl_source",
		Name: "Source Table",
		Type: "table",
	}
	mgr.RegisterDataNode(node1)

	node2 := &lineage.DataLineageNode{
		ID:   "tbl_target",
		Name: "Target Table",
		Type: "table",
	}
	mgr.RegisterDataNode(node2)

	// Create lineage edge
	_, err := mgr.CreateLineageEdge("tbl_source", "tbl_target", "reads")
	if err != nil {
		t.Errorf("Failed to create edge: %v", err)
	}
}

// TestLineageTrace tests data flow tracing
func TestLineageTrace(t *testing.T) {
	mgr := lineage.NewDataLineageTracker()

	// Create nodes
	for i := 1; i <= 3; i++ {
		node := &lineage.DataLineageNode{
			ID:   "node_" + string(rune(i)),
			Name: "Node " + string(rune(i)),
			Type: "table",
		}
		mgr.RegisterDataNode(node)
	}

	// Create edges
	edges := []*lineage.DataLineageEdge{
		{
			SourceID:     "node_1",
			TargetID:     "node_2",
			RelationType: "reads",
			CreatedAt:    time.Now(),
		},
		{
			SourceID:     "node_2",
			TargetID:     "node_3",
			RelationType: "writes",
			CreatedAt:    time.Now(),
		},
	}

	for i, edge := range edges {
		_, err := mgr.CreateLineageEdge(edge.SourceID, edge.TargetID, edge.RelationType)
		if err != nil {
			t.Errorf("Failed to create edge %d: %v", i, err)
		}
	}

	upstream := mgr.GetUpstreamLineage("node_3")
	if len(upstream) == 0 {
		t.Errorf("Expected upstream nodes, got none")
	}
}

// TestAuditLogging tests audit event logging
func TestAuditLogging(t *testing.T) {
	mgr := audit.NewAuditComplianceManager()

	auditLog := &audit.AuditLog{
		UserID:       "user123",
		Action:       "UPDATE",
		ResourceType: "customer",
		ResourceID:   "cust_123",
		SourceIP:     "192.168.1.1",
		Timestamp:    time.Now(),
		Result:       "SUCCESS",
	}

	err := mgr.LogAuditEvent(auditLog)
	if err != nil {
		t.Errorf("Failed to log audit event: %v", err)
	}
}

// TestComplianceRuleRegistration tests rule registration
func TestComplianceRuleRegistration(t *testing.T) {
	mgr := audit.NewAuditComplianceManager()

	rule := &audit.ComplianceRule{
		ID:          "test-rule-1",
		Framework:   "GDPR",
		Requirement: "Test Rule",
		Description: "Test compliance rule",
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	err := mgr.RegisterComplianceRule(rule)
	if err != nil {
		t.Errorf("Failed to register compliance rule: %v", err)
	}
}

// TestViolationRecording tests violation recording
func TestViolationRecording(t *testing.T) {
	mgr := audit.NewAuditComplianceManager()

	rule := &audit.ComplianceRule{
		ID:        "test-rule-1",
		Framework: "GDPR",
	}
	mgr.RegisterComplianceRule(rule)

	violation := &audit.ComplianceViolation{
		RuleID:           "test-rule-1",
		Description:      "Test violation",
		Severity:         "high",
		Timestamp:        time.Now(),
		AffectedResource: "resource_1",
	}

	err := mgr.RecordViolation(violation)
	if err != nil {
		t.Errorf("Failed to record violation: %v", err)
	}
}

// TestComplianceReport tests report generation
func TestComplianceReport(t *testing.T) {
	mgr := audit.NewAuditComplianceManager()

	rule := &audit.ComplianceRule{
		ID:          "gdpr-rule-1",
		Framework:   "GDPR",
		Requirement: "GDPR Compliance",
		Description: "Test GDPR rule",
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	mgr.RegisterComplianceRule(rule)

	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	endDate := now

	report, err := mgr.GenerateComplianceReport("GDPR", startDate, endDate)
	if err != nil {
		t.Errorf("Failed to generate compliance report: %v", err)
		return
	}

	if report.ReportType != "GDPR" {
		t.Errorf("Report type mismatch: got %s, want GDPR", report.ReportType)
	}
}

// TestWorkflowCreation tests workflow creation
func TestWorkflowCreation(t *testing.T) {
	mgr := workflow.NewMultiVersionWorkflowManager()

	wfDef := &workflow.WorkflowDefinition{
		Name:        "Test Workflow",
		Description: "Test workflow definition",
		CreatedBy:   "admin",
		Status:      "draft",
	}

	version, err := mgr.CreateWorkflow(wfDef)
	if err != nil {
		t.Errorf("Failed to create workflow: %v", err)
		return
	}

	if version.Version != "1.0.0" {
		t.Errorf("Version mismatch: got %s, want 1.0.0", version.Version)
	}
}

// TestWorkflowInstance tests workflow instance creation
func TestWorkflowInstance(t *testing.T) {
	mgr := workflow.NewMultiVersionWorkflowManager()

	wfDef := &workflow.WorkflowDefinition{
		Name:      "Test Workflow",
		CreatedBy: "admin",
		Status:    "published",
	}
	mgr.CreateWorkflow(wfDef)

	instance, err := mgr.StartWorkflowInstance(wfDef.ID, "1.0.0", map[string]interface{}{
		"test_key": "test_value",
	})
	if err != nil {
		t.Errorf("Failed to start workflow instance: %v", err)
		return
	}

	if instance.Status != "running" {
		t.Errorf("Instance status mismatch: got %s, want running", instance.Status)
	}
}

// TestWorkflowVersioning tests workflow versioning
func TestWorkflowVersioning(t *testing.T) {
	mgr := workflow.NewMultiVersionWorkflowManager()

	wfDef := &workflow.WorkflowDefinition{
		Name:      "Test Workflow",
		CreatedBy: "admin",
	}
	mgr.CreateWorkflow(wfDef)

	newDef := &workflow.WorkflowDefinition{
		Name:      "Test Workflow Updated",
		CreatedBy: "admin",
	}
	mgr.PublishWorkflowVersion(wfDef.ID, newDef)

	versions := mgr.GetWorkflowVersions(wfDef.ID)
	if len(versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(versions))
	}
}

// BenchmarkEncryption benchmarks encryption
func BenchmarkEncryption(b *testing.B) {
	mgr := encryption.NewFieldLevelEncryption()
	key := &encryption.EncryptionKey{
		ID:        "bench-key",
		Algorithm: "AES-256",
	}
	mgr.RegisterKey(key)

	testData := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.EncryptField("users|email", testData, "bench-key")
	}
}

// BenchmarkLineageEdgeCreation benchmarks lineage operations
func BenchmarkLineageEdgeCreation(b *testing.B) {
	mgr := lineage.NewDataLineageTracker()

	for i := 0; i < 100; i++ {
		node := &lineage.DataLineageNode{
			ID:   "node_" + string(rune(i)),
			Name: "Node " + string(rune(i)),
			Type: "table",
		}
		mgr.RegisterDataNode(node)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.CreateLineageEdge("node_1", "node_2", "reads")
	}
}

// BenchmarkAuditLogging benchmarks audit logging
func BenchmarkAuditLogging(b *testing.B) {
	mgr := audit.NewAuditComplianceManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auditLog := &audit.AuditLog{
			UserID:       "user123",
			Action:       "UPDATE",
			ResourceType: "customer",
			ResourceID:   "cust_123",
			Timestamp:    time.Now(),
		}
		mgr.LogAuditEvent(auditLog)
	}
}

// BenchmarkWorkflowCreation benchmarks workflow creation
func BenchmarkWorkflowCreation(b *testing.B) {
	mgr := workflow.NewMultiVersionWorkflowManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wfDef := &workflow.WorkflowDefinition{
			Name:      "Test Workflow",
			CreatedBy: "admin",
		}
		mgr.CreateWorkflow(wfDef)
	}
}
