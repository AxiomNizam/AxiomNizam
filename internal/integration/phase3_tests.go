package integration

import (
	"testing"
	"time"

	"axiom/internal/audit"
	"axiom/internal/encryption"
	"axiom/internal/lineage"
	"axiom/internal/workflow"
)

// TestEncryptionRegistration tests key registration
func TestEncryptionRegistration(t *testing.T) {
	mgr := encryption.NewFieldLevelEncryption()

	key := &encryption.EncryptionKey{
		KeyID:     "test-key-1",
		Key:       "32-byte-encryption-key-value-12345",
		CreatedAt: time.Now(),
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
		KeyID: "test-key-1",
		Key:   "32-byte-encryption-key-value-12345",
	}
	mgr.RegisterKey(key)

	policy := &encryption.FieldEncryptionPolicy{
		TableName:  "test_table",
		ColumnName: "email",
		KeyID:      "test-key-1",
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
		KeyID: "test-key-1",
		Key:   "32-byte-encryption-key-value-12345",
	}
	mgr.RegisterKey(key)

	testData := []byte("test@example.com")
	encrypted, err := mgr.EncryptField("users", "email", testData)
	if err != nil {
		t.Errorf("Encryption failed: %v", err)
		return
	}

	decrypted, err := mgr.DecryptField(encrypted)
	if err != nil {
		t.Errorf("Decryption failed: %v", err)
		return
	}

	if string(decrypted) != string(testData) {
		t.Errorf("Decrypted data mismatch: got %s, want %s", string(decrypted), string(testData))
	}
}

// TestKeyRotation tests key rotation
func TestKeyRotation(t *testing.T) {
	mgr := encryption.NewFieldLevelEncryption()

	key := &encryption.EncryptionKey{
		KeyID: "test-key-1",
		Key:   "32-byte-encryption-key-value-12345",
	}
	mgr.RegisterKey(key)

	newKey := "new-32-byte-encryption-key-value-123"
	err := mgr.RotateKey("test-key-1", newKey)
	if err != nil {
		t.Errorf("Key rotation failed: %v", err)
	}
}

// TestLineageNodeRegistration tests data node registration
func TestLineageNodeRegistration(t *testing.T) {
	mgr := lineage.NewDataLineageTracker()

	node := &lineage.DataLineageNode{
		NodeID:   "tbl_test",
		NodeName: "Test Table",
		NodeType: "table",
		Schema:   "public",
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
		NodeID:   "tbl_source",
		NodeName: "Source Table",
		NodeType: "table",
	}
	mgr.RegisterDataNode(node1)

	node2 := &lineage.DataLineageNode{
		NodeID:   "tbl_target",
		NodeName: "Target Table",
		NodeType: "table",
	}
	mgr.RegisterDataNode(node2)

	edge := &lineage.DataLineageEdge{
		SourceNodeID: "tbl_source",
		TargetNodeID: "tbl_target",
		RelationType: "reads",
		CreatedAt:    time.Now(),
	}

	err := mgr.CreateLineageEdge(edge)
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
			NodeID:   "node_" + string(rune(i)),
			NodeName: "Node " + string(rune(i)),
			NodeType: "table",
		}
		mgr.RegisterDataNode(node)
	}

	// Create edges
	edges := []*lineage.DataLineageEdge{
		{
			SourceNodeID: "node_1",
			TargetNodeID: "node_2",
			RelationType: "reads",
			CreatedAt:    time.Now(),
		},
		{
			SourceNodeID: "node_2",
			TargetNodeID: "node_3",
			RelationType: "writes",
			CreatedAt:    time.Now(),
		},
	}

	for _, edge := range edges {
		mgr.CreateLineageEdge(edge)
	}

	_, err := mgr.GetUpstreamLineage("node_3")
	if err != nil {
		t.Errorf("Failed to get upstream lineage: %v", err)
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
		IPAddress:    "192.168.1.1",
		Timestamp:    time.Now(),
		Status:       "success",
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
		RuleID:      "test-rule-1",
		RuleName:    "Test Rule",
		Framework:   "GDPR",
		Description: "Test compliance rule",
		Severity:    "high",
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
		RuleID:    "test-rule-1",
		RuleName:  "Test Rule",
		Framework: "GDPR",
	}
	mgr.RegisterComplianceRule(rule)

	violation := &audit.ComplianceViolation{
		RuleID:     "test-rule-1",
		Description: "Test violation",
		Severity:   "high",
		DetectedAt: time.Now(),
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
		RuleID:      "gdpr-rule-1",
		RuleName:    "GDPR Rule",
		Framework:   "GDPR",
		Description: "Test GDPR rule",
	}
	mgr.RegisterComplianceRule(rule)

	report, err := mgr.GenerateComplianceReport("GDPR")
	if err != nil {
		t.Errorf("Failed to generate compliance report: %v", err)
		return
	}

	if report.Framework != "GDPR" {
		t.Errorf("Report framework mismatch: got %s, want GDPR", report.Framework)
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
		KeyID: "bench-key",
		Key:   "32-byte-encryption-key-value-12345",
	}
	mgr.RegisterKey(key)

	testData := []byte("test@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.EncryptField("users", "email", testData)
	}
}

// BenchmarkLineageEdgeCreation benchmarks lineage operations
func BenchmarkLineageEdgeCreation(b *testing.B) {
	mgr := lineage.NewDataLineageTracker()

	for i := 0; i < 100; i++ {
		node := &lineage.DataLineageNode{
			NodeID:   "node_" + string(rune(i)),
			NodeName: "Node " + string(rune(i)),
			NodeType: "table",
		}
		mgr.RegisterDataNode(node)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		edge := &lineage.DataLineageEdge{
			SourceNodeID: "node_1",
			TargetNodeID: "node_2",
			RelationType: "reads",
			CreatedAt:    time.Now(),
		}
		mgr.CreateLineageEdge(edge)
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
			Status:       "success",
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
