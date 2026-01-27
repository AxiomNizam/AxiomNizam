package integration

import (
	"fmt"
	"sync"
	"time"

	"axiom/internal/audit"
	"axiom/internal/encryption"
	"axiom/internal/handlers"
	"axiom/internal/lineage"
	"axiom/internal/workflow"

	"github.com/gin-gonic/gin"
)

// Phase3Integration orchestrates all Phase 3 features
type Phase3Integration struct {
	encryptionMgr *encryption.FieldLevelEncryption
	lineageMgr    *lineage.DataLineageTracker
	auditMgr      *audit.AuditComplianceManager
	workflowMgr   *workflow.MultiVersionWorkflowManager
	handlers      *handlers.Phase3Handlers
	mu            sync.RWMutex
	initialized   bool
	initialized_at time.Time
}

// NewPhase3Integration creates Phase3Integration
func NewPhase3Integration() *Phase3Integration {
	return &Phase3Integration{
		encryptionMgr: encryption.NewFieldLevelEncryption(),
		lineageMgr:    lineage.NewDataLineageTracker(),
		auditMgr:      audit.NewAuditComplianceManager(),
		workflowMgr:   workflow.NewMultiVersionWorkflowManager(),
	}
}

// Initialize initializes all Phase 3 features
func (pi *Phase3Integration) Initialize() error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	if pi.initialized {
		return nil
	}

	// Create handlers
	pi.handlers = handlers.NewPhase3Handlers(
		pi.encryptionMgr,
		pi.lineageMgr,
		pi.auditMgr,
		pi.workflowMgr,
	)

	pi.initialized = true
	pi.initialized_at = time.Now()

	return nil
}

// RegisterRoutes registers all Phase 3 routes
func (pi *Phase3Integration) RegisterRoutes(router *gin.Engine) error {
	if !pi.initialized {
		if err := pi.Initialize(); err != nil {
			return err
		}
	}

	// Encryption endpoints
	encryption := router.Group("/api/v3/encryption")
	{
		encryption.POST("/register-key", pi.handlers.RegisterEncryptionKey)
		encryption.POST("/policy", pi.handlers.AddEncryptionPolicy)
		encryption.POST("/encrypt", pi.handlers.EncryptFieldValue)
		encryption.POST("/decrypt", pi.handlers.DecryptFieldValue)
		encryption.PUT("/rotate/:key_id", pi.handlers.RotateEncryptionKey)
		encryption.GET("/metrics", pi.handlers.GetEncryptionMetrics)
		encryption.GET("/status", pi.handlers.GetEncryptionStatus)
	}

	// Lineage endpoints
	lineageGroup := router.Group("/api/v3/lineage")
	{
		lineageGroup.POST("/node", pi.handlers.RegisterDataNode)
		lineageGroup.POST("/edge", pi.handlers.CreateLineageEdge)
		lineageGroup.GET("/upstream", pi.handlers.GetUpstreamLineage)
		lineageGroup.GET("/downstream", pi.handlers.GetDownstreamLineage)
		lineageGroup.POST("/analyze-impact", pi.handlers.AnalyzeImpact)
		lineageGroup.GET("/graph", pi.handlers.GetLineageGraph)
		lineageGroup.GET("/stats", pi.handlers.GetLineageStats)
	}

	// Audit & Compliance endpoints
	audit := router.Group("/api/v3/audit")
	{
		audit.POST("/log", pi.handlers.LogAuditEvent)
		audit.POST("/compliance-rule", pi.handlers.RegisterComplianceRule)
		audit.GET("/report", pi.handlers.GenerateComplianceReport)
		audit.GET("/status", pi.handlers.GetComplianceStatus)
		audit.GET("/search", pi.handlers.SearchAuditLogs)
		audit.POST("/violation", pi.handlers.RecordViolation)
	}

	// Workflow endpoints
	wfGroup := router.Group("/api/v3/workflow")
	{
		wfGroup.POST("/create", pi.handlers.CreateWorkflow)
		wfGroup.POST("/publish", pi.handlers.PublishWorkflowVersion)
		wfGroup.POST("/instance/start", pi.handlers.StartWorkflowInstance)
		wfGroup.GET("/metrics", pi.handlers.GetWorkflowMetrics)
		wfGroup.GET("/status", pi.handlers.GetWorkflowStatus)
		wfGroup.GET("/history", pi.handlers.GetInstanceHistory)
	}

	return nil
}

// GetStatus returns Phase 3 overall status
func (pi *Phase3Integration) GetStatus() map[string]interface{} {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return map[string]interface{}{
		"initialized":      pi.initialized,
		"initialized_at":   pi.initialized_at,
		"encryption":       pi.encryptionMgr.GetEncryptionStatus(),
		"lineage":          pi.lineageMgr.GetLineageStats(),
		"audit":            pi.auditMgr.GetAuditMetrics(),
		"workflow":         pi.workflowMgr.GetWorkflowStatus(),
		"total_endpoints":  25,
	}
}

// HealthCheck performs health check
func (pi *Phase3Integration) HealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"service":          "phase3-integration",
		"status":           "healthy",
		"encryption_ready": pi.encryptionMgr != nil,
		"lineage_ready":    pi.lineageMgr != nil,
		"audit_ready":      pi.auditMgr != nil,
		"workflow_ready":   pi.workflowMgr != nil,
		"timestamp":        time.Now(),
	}
}

// GetMetrics returns comprehensive Phase 3 metrics
func (pi *Phase3Integration) GetMetrics() map[string]interface{} {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return map[string]interface{}{
		"encryption_metrics": pi.encryptionMgr.GetEncryptionMetrics(),
		"lineage_stats":      pi.lineageMgr.GetLineageStats(),
		"audit_metrics":      pi.auditMgr.GetAuditMetrics(),
		"workflow_status":    pi.workflowMgr.GetWorkflowStatus(),
	}
}

// SetupDefaultRules sets up default compliance rules
func (pi *Phase3Integration) SetupDefaultRules() error {
	rules := []*audit.ComplianceRule{
		{
			RuleID:      "gdpr-data-retention",
			RuleName:    "GDPR Data Retention",
			Framework:   "GDPR",
			Description: "Personal data retention period compliance",
			Severity:    "high",
			CreatedAt:   time.Now(),
		},
		{
			RuleID:      "hipaa-access-control",
			RuleName:    "HIPAA Access Control",
			Framework:   "HIPAA",
			Description: "Healthcare data access restrictions",
			Severity:    "critical",
			CreatedAt:   time.Now(),
		},
		{
			RuleID:      "soc2-audit-logging",
			RuleName:    "SOC2 Audit Logging",
			Framework:   "SOC2",
			Description: "Complete audit trail logging requirement",
			Severity:    "high",
			CreatedAt:   time.Now(),
		},
		{
			RuleID:      "pci-dss-encryption",
			RuleName:    "PCI-DSS Encryption",
			Framework:   "PCI-DSS",
			Description: "Payment data encryption requirement",
			Severity:    "critical",
			CreatedAt:   time.Now(),
		},
	}

	for _, rule := range rules {
		if err := pi.auditMgr.RegisterComplianceRule(rule); err != nil {
			return fmt.Errorf("failed to register rule %s: %w", rule.RuleID, err)
		}
	}

	return nil
}

// SetupDefaultEncryption sets up default encryption keys
func (pi *Phase3Integration) SetupDefaultEncryption() error {
	keys := []*encryption.EncryptionKey{
		{
			KeyID:       "default-aes-256",
			Key:         "32-byte-encryption-key-for-aes256", // Should be 32 bytes
			CreatedAt:   time.Now(),
			ExpiresAt:   time.Now().AddDate(1, 0, 0),
			EncryptType: "deterministic",
		},
		{
			KeyID:       "searchable-aes-256",
			Key:         "searchable-encryption-key-32-bytes",
			CreatedAt:   time.Now(),
			ExpiresAt:   time.Now().AddDate(1, 0, 0),
			EncryptType: "searchable",
		},
	}

	for _, key := range keys {
		if err := pi.encryptionMgr.RegisterKey(key); err != nil {
			return fmt.Errorf("failed to register key %s: %w", key.KeyID, err)
		}
	}

	return nil
}

// GetFeatureStatus gets status of each Phase 3 feature
func (pi *Phase3Integration) GetFeatureStatus() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"field_encryption": {
			"name":    "Field-Level Encryption",
			"status":  "active",
			"metrics": pi.encryptionMgr.GetEncryptionMetrics(),
		},
		"data_lineage": {
			"name":    "Data Lineage Tracking",
			"status":  "active",
			"stats":   pi.lineageMgr.GetLineageStats(),
		},
		"audit_compliance": {
			"name":    "Audit & Compliance Reports",
			"status":  "active",
			"metrics": pi.auditMgr.GetAuditMetrics(),
		},
		"workflow_versioning": {
			"name":   "Multi-Version Workflow Support",
			"status": "active",
			"status_info": pi.workflowMgr.GetWorkflowStatus(),
		},
	}
}

// GetAPIEndpoints returns all Phase 3 API endpoints
func (pi *Phase3Integration) GetAPIEndpoints() map[string][]string {
	return map[string][]string{
		"encryption": {
			"POST /api/v3/encryption/register-key",
			"POST /api/v3/encryption/policy",
			"POST /api/v3/encryption/encrypt",
			"POST /api/v3/encryption/decrypt",
			"PUT /api/v3/encryption/rotate/:key_id",
			"GET /api/v3/encryption/metrics",
			"GET /api/v3/encryption/status",
		},
		"lineage": {
			"POST /api/v3/lineage/node",
			"POST /api/v3/lineage/edge",
			"GET /api/v3/lineage/upstream",
			"GET /api/v3/lineage/downstream",
			"POST /api/v3/lineage/analyze-impact",
			"GET /api/v3/lineage/graph",
			"GET /api/v3/lineage/stats",
		},
		"audit": {
			"POST /api/v3/audit/log",
			"POST /api/v3/audit/compliance-rule",
			"GET /api/v3/audit/report",
			"GET /api/v3/audit/status",
			"GET /api/v3/audit/search",
			"POST /api/v3/audit/violation",
		},
		"workflow": {
			"POST /api/v3/workflow/create",
			"POST /api/v3/workflow/publish",
			"POST /api/v3/workflow/instance/start",
			"GET /api/v3/workflow/metrics",
			"GET /api/v3/workflow/status",
			"GET /api/v3/workflow/history",
		},
	}
}
