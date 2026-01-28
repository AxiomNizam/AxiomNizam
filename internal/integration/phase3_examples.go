package integration

import (
	"fmt"
)

// Phase3ExamplesUsage provides comprehensive usage examples
type Phase3ExamplesUsage struct{}

// ExampleFieldLevelEncryption demonstrates field-level encryption
func (e *Phase3ExamplesUsage) ExampleFieldLevelEncryption() string {
	example := `
// Example 1: Register Encryption Key
POST /api/v3/encryption/register-key
{
  "key_id": "prod-aes-256-v1",
  "key": "32-byte-encryption-key-value-here",
  "expires_at": "2025-12-31T23:59:59Z",
  "encrypt_type": "deterministic"
}

// Example 2: Add Encryption Policy
POST /api/v3/encryption/policy
{
  "table_name": "customers",
  "column_name": "email",
  "key_id": "prod-aes-256-v1",
  "searchable": true
}

// Example 3: Encrypt Field Value
POST /api/v3/encryption/encrypt
{
  "table_name": "customers",
  "column_name": "email",
  "value": "user@example.com"
}
Response:
{
  "encrypted_data": "a1b2c3d4e5f6...",
  "iv": "x1y2z3w4...",
  "key_id": "prod-aes-256-v1"
}

// Example 4: Decrypt Field Value
POST /api/v3/encryption/decrypt
{
  "table_name": "customers",
  "column_name": "email",
  "encrypted_data": "a1b2c3d4e5f6...",
  "iv": "x1y2z3w4...",
  "key_id": "prod-aes-256-v1"
}
Response:
{
  "decrypted_value": "user@example.com"
}

// Example 5: Rotate Encryption Key
PUT /api/v3/encryption/rotate/prod-aes-256-v1
new_key=new-32-byte-encryption-key-value-here

// Example 6: Get Encryption Metrics
GET /api/v3/encryption/metrics
Response:
{
  "metrics": {
    "total_encrypted": 1500,
    "total_decrypted": 2000,
    "total_rotations": 5,
    "encryption_errors": 2,
    "avg_encryption_time_ms": 2.5
  }
}
`
	return example
}

// ExampleDataLineageTracking demonstrates data lineage
func (e *Phase3ExamplesUsage) ExampleDataLineageTracking() string {
	example := `
// Example 1: Register Data Node
POST /api/v3/lineage/node
{
  "node_id": "tbl_customers",
  "node_name": "Customers Table",
  "node_type": "table",
  "schema": "public",
  "description": "Customer master data"
}

// Example 2: Register Another Data Node
POST /api/v3/lineage/node
{
  "node_id": "view_active_customers",
  "node_name": "Active Customers View",
  "node_type": "view",
  "schema": "public",
  "description": "Only active customers"
}

// Example 3: Create Lineage Edge
POST /api/v3/lineage/edge
{
  "source_node_id": "tbl_customers",
  "target_node_id": "view_active_customers",
  "relation_type": "reads",
  "metadata": {
    "filter_condition": "status = 'active'",
    "transformation": "filter"
  }
}

// Example 4: Get Upstream Lineage
GET /api/v3/lineage/upstream?node_id=view_active_customers
Response:
{
  "upstream_lineage": {
    "path": ["view_active_customers", "tbl_customers"],
    "edges": 1,
    "depth": 1
  }
}

// Example 5: Get Downstream Lineage
GET /api/v3/lineage/downstream?node_id=tbl_customers
Response:
  "downstream_lineage": {
    "affected_nodes": ["view_active_customers"],
    "total_affected": 1
  }
}

// Example 6: Analyze Change Impact
POST /api/v3/lineage/analyze-impact
{
  "source_node_id": "tbl_customers"
}
Response:
{
  "impact_analysis": {
    "affected_nodes": 3,
    "impact_level": "high",
    "affected_dependencies": [
      "view_active_customers",
      "rpt_customer_analysis",
      "etl_job_daily"
    ]
  }
}

// Example 7: Get Lineage Graph
GET /api/v3/lineage/graph
Response:
{
  "lineage_graph": {
    "nodes": 25,
    "edges": 40,
    "densest_area": "warehouse_schema"
  }
}

// Example 8: Get Lineage Statistics
GET /api/v3/lineage/stats
Response:
{
  "statistics": {
    "total_nodes": 25,
    "total_edges": 40,
    "total_flows": 15,
    "max_depth": 5,
    "transformation_count": 12
  }
}
`
	return example
}

// ExampleAuditCompliance demonstrates audit and compliance
func (e *Phase3ExamplesUsage) ExampleAuditCompliance() string {
	example := `
// Example 1: Log Audit Event
POST /api/v3/audit/log
{
  "user_id": "user@example.com",
  "action": "UPDATE",
  "resource_type": "customer",
  "resource_id": "cust_12345",
  "ip_address": "192.168.1.100",
  "changes": {
    "email": {"old": "old@example.com", "new": "new@example.com"}
  },
  "status": "success"
}

// Example 2: Register Compliance Rule
POST /api/v3/audit/compliance-rule
{
  "rule_id": "gdpr-right-to-be-forgotten",
  "rule_name": "GDPR Right to be Forgotten",
  "framework": "GDPR",
  "description": "Personal data must be deletable on request",
  "severity": "high"
}

// Example 3: Record Compliance Violation
POST /api/v3/audit/violation
{
  "rule_id": "gdpr-right-to-be-forgotten",
  "description": "User deletion request not processed within 30 days",
  "severity": "critical"
}

// Example 4: Generate Compliance Report
GET /api/v3/audit/report?framework=GDPR
Response:
{
  "compliance_report": {
    "framework": "GDPR",
    "compliance_score": 87.5,
    "findings": [
      {
        "rule_id": "gdpr-right-to-be-forgotten",
        "status": "violation",
        "severity": "critical"
      }
    ],
    "risk_assessment": "medium",
    "generated_at": "2024-01-15T10:30:00Z"
  }
}

// Example 5: Get Compliance Status
GET /api/v3/audit/status
Response:
{
  "status": {
    "gdpr_compliant": true,
    "hipaa_compliant": false,
    "soc2_compliant": true,
    "pci_dss_compliant": false,
    "total_violations": 5,
    "critical_violations": 2,
    "last_audit": "2024-01-15T10:00:00Z"
  }
}

// Example 6: Search Audit Logs
GET /api/v3/audit/search?user_id=user@example.com&resource_type=customer
Response:
{
  "audit_logs": [
    {
      "timestamp": "2024-01-15T10:30:00Z",
      "user_id": "user@example.com",
      "action": "UPDATE",
      "resource_type": "customer",
      "resource_id": "cust_12345",
      "ip_address": "192.168.1.100",
      "status": "success"
    }
  ]
}
`
	return example
}

// ExampleMultiVersionWorkflow demonstrates workflow versioning
func (e *Phase3ExamplesUsage) ExampleMultiVersionWorkflow() string {
	example := `
// Example 1: Create New Workflow
POST /api/v3/workflow/create
{
  "name": "Customer Onboarding",
  "description": "Complete customer onboarding process",
  "created_by": "admin@example.com",
  "steps": [
    {
      "step_number": 1,
      "name": "Verify Email",
      "step_type": "action"
    },
    {
      "step_number": 2,
      "name": "Review Documents",
      "step_type": "approval"
    }
  ]
}
Response:
{
  "message": "workflow created",
  "workflow_id": "wf-1234567890",
  "version": "1.0.0"
}

// Example 2: Publish New Workflow Version
POST /api/v3/workflow/publish
{
  "workflow_id": "wf-1234567890",
  "change_summary": "Added KYC verification step",
  "created_by": "admin@example.com",
  "steps": [
    {
      "step_number": 1,
      "name": "Verify Email",
      "step_type": "action"
    },
    {
      "step_number": 2,
      "name": "KYC Verification",
      "step_type": "action"
    },
    {
      "step_number": 3,
      "name": "Review Documents",
      "step_type": "approval"
    }
  ]
}
Response:
{
  "message": "workflow version published",
  "version": "2.0.0"
}

// Example 3: Start Workflow Instance
POST /api/v3/workflow/instance/start
{
  "workflow_id": "wf-1234567890",
  "version": "2.0.0",
  "context_data": {
    "customer_id": "cust_12345",
    "customer_email": "user@example.com",
    "initiated_by": "admin@example.com"
  }
}
Response:
{
  "message": "workflow instance started",
  "instance_id": "inst-1234567890",
  "status": "running"
}

// Example 4: Get Workflow Metrics
GET /api/v3/workflow/metrics?workflow_id=wf-1234567890
Response:
{
  "metrics": {
    "total_instances": 150,
    "completed": 145,
    "failed": 3,
    "running": 2,
    "success_rate": 96.7,
    "avg_duration_ms": 2500,
    "total_versions": 3
  }
}

// Example 5: Get Workflow Status
GET /api/v3/workflow/status
Response:
{
  "status": {
    "total_workflows": 10,
    "total_versions": 25,
    "active_versions": 10,
    "running_instances": 15,
    "execution_logs": 2500,
    "migrations": 3
  }
}

// Example 6: Get Instance History
GET /api/v3/workflow/history?workflow_id=wf-1234567890
Response:
{
  "instance_history": [
    {
      "instance_id": "inst-1234567890",
      "workflow_version": "2.0.0",
      "status": "completed",
      "started_at": "2024-01-15T10:00:00Z",
      "completed_at": "2024-01-15T10:25:00Z"
    },
    {
      "instance_id": "inst-0987654321",
      "workflow_version": "2.0.0",
      "status": "running",
      "started_at": "2024-01-15T10:30:00Z",
      "completed_at": null
    }
  ]
}
`
	return example
}

// ExampleIntegrationFlow demonstrates integrated workflow
func (e *Phase3ExamplesUsage) ExampleIntegrationFlow() string {
	example := `
// INTEGRATED EXAMPLE: Complete Data Processing with All Phase 3 Features

// Step 1: Set up lineage for customer data flow
POST /api/v3/lineage/node
{
  "node_id": "source_customer_db",
  "node_name": "Source Customer Database",
  "node_type": "table",
  "schema": "external"
}

POST /api/v3/lineage/node
{
  "node_id": "encrypted_customer_db",
  "node_name": "Encrypted Customer Data",
  "node_type": "table",
  "schema": "public"
}

POST /api/v3/lineage/edge
{
  "source_node_id": "source_customer_db",
  "target_node_id": "encrypted_customer_db",
  "relation_type": "transforms",
  "metadata": {"operation": "encrypt"}
}

// Step 2: Register encryption policies
POST /api/v3/encryption/policy
{
  "table_name": "encrypted_customer_db",
  "column_name": "email",
  "key_id": "prod-aes-256-v1",
  "searchable": true
}

POST /api/v3/encryption/policy
{
  "table_name": "encrypted_customer_db",
  "column_name": "phone",
  "key_id": "prod-aes-256-v1",
  "searchable": true
}

// Step 3: Start customer data processing workflow
POST /api/v3/workflow/instance/start
{
  "workflow_id": "wf-customer-processing",
  "version": "1.0.0",
  "context_data": {
    "batch_id": "batch_001",
    "source_table": "source_customer_db",
    "target_table": "encrypted_customer_db"
  }
}

// Step 4: Log audit entry for data processing
POST /api/v3/audit/log
{
  "user_id": "system-process",
  "action": "BULK_UPDATE",
  "resource_type": "customer_data",
  "resource_id": "batch_001",
  "changes": {
    "encrypted_records": 1000,
    "encryption_algorithm": "AES-256-GCM"
  },
  "status": "success"
}

// Step 5: Analyze impact of changes
POST /api/v3/lineage/analyze-impact
{
  "source_node_id": "source_customer_db"
}

// Step 6: Generate compliance report
GET /api/v3/audit/report?framework=GDPR

// Step 7: Check all system metrics
GET /api/v3/encryption/metrics
GET /api/v3/lineage/stats
GET /api/v3/audit/status
GET /api/v3/workflow/status

Result: Complete audit trail of data encryption, lineage tracking, compliance validation, and workflow automation.
`
	return example
}

// ExampleCodeSnippets provides Go code examples
func (e *Phase3ExamplesUsage) ExampleCodeSnippets() string {
	example := `
// Example: Using Phase 3 Managers Directly in Go

package main

import (
	"axiom/internal/encryption"
	"axiom/internal/lineage"
	"axiom/internal/audit"
	"axiom/internal/workflow"
)

func main() {
	// Initialize managers
	encMgr := encryption.NewFieldLevelEncryption()
	linMgr := lineage.NewDataLineageTracker()
	audMgr := audit.NewAuditComplianceManager()
	wfMgr := workflow.NewMultiVersionWorkflowManager()

	// Encryption example
	key := &encryption.EncryptionKey{
		KeyID: "app-key-1",
		Key:   "32-byte-key-value-here-pad-to-32",
	}
	encMgr.RegisterKey(key)

	encField, _ := encMgr.EncryptField("users", "email", []byte("user@example.com"))
	println("Encrypted:", encField.KeyID)

	// Lineage example
	node := &lineage.DataLineageNode{
		NodeID:   "tbl_users",
		NodeName: "Users Table",
		NodeType: "table",
	}
	linMgr.RegisterDataNode(node)

	// Audit example
	auditLog := &audit.AuditLog{
		UserID:       "admin",
		Action:       "CREATE",
		ResourceType: "user",
		Status:       "success",
	}
	audMgr.LogAuditEvent(auditLog)

	// Workflow example
	wfDef := &workflow.WorkflowDefinition{
		Name:      "Data Processing",
		CreatedBy: "admin",
	}
	wfMgr.CreateWorkflow(wfDef)

	// Get metrics
	encMetrics := encMgr.GetEncryptionMetrics()
	linStats := linMgr.GetLineageStats()
	audMetrics := audMgr.GetAuditMetrics()
	wfStatus := wfMgr.GetWorkflowStatus()

	println("Encryption metrics:", encMetrics)
	println("Lineage stats:", linStats)
	println("Audit metrics:", audMetrics)
	println("Workflow status:", wfStatus)
}
`
	return example
}

// PrintAllExamples prints all examples
func (e *Phase3ExamplesUsage) PrintAllExamples() {
	fmt.Println("=== PHASE 3: FIELD-LEVEL ENCRYPTION EXAMPLES ===")
	fmt.Println(e.ExampleFieldLevelEncryption())
	fmt.Println("\n=== PHASE 3: DATA LINEAGE TRACKING EXAMPLES ===")
	fmt.Println(e.ExampleDataLineageTracking())
	fmt.Println("\n=== PHASE 3: AUDIT & COMPLIANCE EXAMPLES ===")
	fmt.Println(e.ExampleAuditCompliance())
	fmt.Println("\n=== PHASE 3: MULTI-VERSION WORKFLOW EXAMPLES ===")
	fmt.Println(e.ExampleMultiVersionWorkflow())
	fmt.Println("\n=== PHASE 3: INTEGRATED WORKFLOW EXAMPLE ===")
	fmt.Println(e.ExampleIntegrationFlow())
	fmt.Println("\n=== PHASE 3: GO CODE EXAMPLES ===")
	fmt.Println(e.ExampleCodeSnippets())
}
